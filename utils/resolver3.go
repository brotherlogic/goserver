package utils

import (
	"errors"
	"fmt"

	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/resolver"
)

// PickFirstBalancerName is the name of the pick_first balancer.
const PickonlyfirstBalancerName = "pick_only_first"

func newPickonlyfirstBuilder() balancer.Builder {
	return &pickonlyfirstBuilder{}
}

type pickonlyfirstBuilder struct{}

func (*pickonlyfirstBuilder) Build(cc balancer.ClientConn, opt balancer.BuildOptions) balancer.Balancer {
	return &pickonlyfirstBalancer{cc: cc, onlypicker: onlypicker{lastPick: -1}}
}

func (*pickonlyfirstBuilder) Name() string {
	return PickonlyfirstBalancerName
}

type pickonlyfirstBalancer struct {
	onlypicker onlypicker
	cc         balancer.ClientConn
}

var _ balancer.V2Balancer = &pickonlyfirstBalancer{} // Assert we implement v2

func (b *pickonlyfirstBalancer) HandleResolvedAddrs(addrs []resolver.Address, err error) {
	if err != nil {
		return
	}
	b.UpdateClientConnState(balancer.ClientConnState{ResolverState: resolver.State{Addresses: addrs}}) // Ignore error
}

func (b *pickonlyfirstBalancer) ResolverError(err error) {

}

func (b *pickonlyfirstBalancer) HandleSubConnStateChange(sc balancer.SubConn, s connectivity.State) {
	b.UpdateSubConnState(sc, balancer.SubConnState{ConnectivityState: s})
}

func (b *pickonlyfirstBalancer) UpdateClientConnState(cs balancer.ClientConnState) error {
	if len(cs.ResolverState.Addresses) == 0 {
		b.ResolverError(errors.New("produced zero addresses"))
		return balancer.ErrBadResolverState
	}
	for _, address := range cs.ResolverState.Addresses {
		sc, err := (b.cc.NewSubConn([]resolver.Address{address}, balancer.NewSubConnOptions{}))
		if err == nil {
			b.onlypicker.Add(sc, address)
		}
	}
	/*if b.sc == nil {
		var err error
		b.sc, err = b.cc.NewSubConn(cs.ResolverState.Addresses, balancer.NewSubConnOptions{})
		if err != nil {
			if grpclog.V(2) {
				grpclog.Errorf("pickonlyfirstBalancer: failed to NewSubConn: %v", err)
			}
			b.state = connectivity.TransientFailure
			b.cc.UpdateState(balancer.State{ConnectivityState: connectivity.TransientFailure,
				Picker: &picker{err: status.Errorf(codes.Unavailable, "error creating connection: %v", err)}},
			)
			return balancer.ErrBadResolverState
		}
		b.state = connectivity.Idle
		b.cc.UpdateState(balancer.State{ConnectivityState: connectivity.Idle, Picker: &picker{result: balancer.PickResult{SubConn: b.sc}}})
		b.sc.Connect()
	} else {
		b.sc.UpdateAddresses(cs.ResolverState.Addresses)
		b.sc.Connect()
	}*/
	return nil
}

func (b *pickonlyfirstBalancer) UpdateSubConnState(sc balancer.SubConn, s balancer.SubConnState) {
	/*if grpclog.V(2) {
		grpclog.Infof("pickonlyfirstBalancer: HandleSubConnStateChange: %p, %v", sc, s)
	}
	if b.sc != sc {
		if grpclog.V(2) {
			grpclog.Infof("pickonlyfirstBalancer: ignored state change because sc is not recognized")
		}
		return
	}*/

	if s.ConnectivityState == connectivity.Shutdown {
		return
	}

	switch s.ConnectivityState {
	case connectivity.Ready, connectivity.Idle:
		done := b.onlypicker.Ready(sc)
		if done {
			b.cc.UpdateState(balancer.State{ConnectivityState: s.ConnectivityState, Picker: &b.onlypicker})
		}
	case connectivity.Connecting:
		// Pass
	case connectivity.TransientFailure:
		done := b.onlypicker.Failure(sc)
		if done {
			b.cc.UpdateState(balancer.State{ConnectivityState: s.ConnectivityState, Picker: &b.onlypicker})
		}
	}
}

func (b *pickonlyfirstBalancer) Close() {

}

type onlypicker struct {
	sc       []balancer.SubConn
	resolved []bool
	ready    []bool
	address  []resolver.Address
	lastPick int
}

func buildonly(scs []balancer.SubConn) *onlypicker {
	p := &onlypicker{sc: scs, resolved: make([]bool, len(scs)), ready: make([]bool, len(scs))}
	return p
}

func (p *onlypicker) Pick(info balancer.PickInfo) (balancer.PickResult, error) {
	for i, res := range p.sc {
		if p.ready[i] {
			p.lastPick = i
			return balancer.PickResult{SubConn: res}, nil
		}
	}
	return balancer.PickResult{}, fmt.Errorf("No servers available")
}

func (p *onlypicker) Add(sc balancer.SubConn, addr resolver.Address) {
	p.sc = append(p.sc, sc)
	p.resolved = append(p.resolved, false)
	p.ready = append(p.ready, false)
	p.address = append(p.address, addr)
	sc.Connect()
}

func (p *onlypicker) allReady() bool {
	for _, r := range p.resolved {
		if !r {
			return false
		}
	}
	return true
}

func (p *onlypicker) Ready(sc balancer.SubConn) bool {
	for i, isc := range p.sc {
		if sc == isc {
			p.resolved[i] = true
			p.ready[i] = true
		}
	}
	return p.allReady()
}

func (p *onlypicker) Failure(sc balancer.SubConn) bool {
	for i, isc := range p.sc {
		if sc == isc {
			p.resolved[i] = true
			unregister(p.address[i].Addr)
		}
	}
	return p.allReady()
}

func init() {
	balancer.Register(newPickonlyfirstBuilder())
}
