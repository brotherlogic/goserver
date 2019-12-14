package utils

import (
	"errors"
	"fmt"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/resolver"

	pbd "github.com/brotherlogic/discovery/proto"
)

// PickFirstBalancerName is the name of the pick_first balancer.
const PickFirstBalancerName = "my_pick_first"

func newPickfirstBuilder() balancer.Builder {
	return &pickfirstBuilder{}
}

type pickfirstBuilder struct{}

func (*pickfirstBuilder) Build(cc balancer.ClientConn, opt balancer.BuildOptions) balancer.Balancer {
	return &pickfirstBalancer{cc: cc}
}

func (*pickfirstBuilder) Name() string {
	return PickFirstBalancerName
}

type pickfirstBalancer struct {
	picker picker
	cc     balancer.ClientConn
}

var _ balancer.V2Balancer = &pickfirstBalancer{} // Assert we implement v2

func (b *pickfirstBalancer) HandleResolvedAddrs(addrs []resolver.Address, err error) {
	if err != nil {
		return
	}
	b.UpdateClientConnState(balancer.ClientConnState{ResolverState: resolver.State{Addresses: addrs}}) // Ignore error
}

func (b *pickfirstBalancer) ResolverError(err error) {

}

func (b *pickfirstBalancer) HandleSubConnStateChange(sc balancer.SubConn, s connectivity.State) {
	b.UpdateSubConnState(sc, balancer.SubConnState{ConnectivityState: s})
}

func (b *pickfirstBalancer) UpdateClientConnState(cs balancer.ClientConnState) error {
	if len(cs.ResolverState.Addresses) == 0 {
		b.ResolverError(errors.New("produced zero addresses"))
		return balancer.ErrBadResolverState
	}
	for _, address := range cs.ResolverState.Addresses {
		sc, err := (b.cc.NewSubConn([]resolver.Address{address}, balancer.NewSubConnOptions{}))
		if err == nil {
			b.picker.Add(sc, address)
		}
	}
	/*if b.sc == nil {
		var err error
		b.sc, err = b.cc.NewSubConn(cs.ResolverState.Addresses, balancer.NewSubConnOptions{})
		if err != nil {
			if grpclog.V(2) {
				grpclog.Errorf("pickfirstBalancer: failed to NewSubConn: %v", err)
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

func (b *pickfirstBalancer) UpdateSubConnState(sc balancer.SubConn, s balancer.SubConnState) {
	/*if grpclog.V(2) {
		grpclog.Infof("pickfirstBalancer: HandleSubConnStateChange: %p, %v", sc, s)
	}
	if b.sc != sc {
		if grpclog.V(2) {
			grpclog.Infof("pickfirstBalancer: ignored state change because sc is not recognized")
		}
		return
	}*/

	if s.ConnectivityState == connectivity.Shutdown {
		return
	}

	switch s.ConnectivityState {
	case connectivity.Ready, connectivity.Idle:
		done := b.picker.Ready(sc)
		if done {
			b.cc.UpdateState(balancer.State{ConnectivityState: s.ConnectivityState, Picker: &b.picker})
		}
	case connectivity.Connecting:
		// Pass
	case connectivity.TransientFailure:
		done := b.picker.Failure(sc)
		if done {
			b.cc.UpdateState(balancer.State{ConnectivityState: s.ConnectivityState, Picker: &b.picker})
		}
	}
}

func (b *pickfirstBalancer) Close() {

}

type picker struct {
	sc       []balancer.SubConn
	resolved []bool
	ready    []bool
	address  []resolver.Address
	lastPick int
}

func build(scs []balancer.SubConn) *picker {
	p := &picker{sc: scs, resolved: make([]bool, len(scs)), ready: make([]bool, len(scs))}
	return p
}

func (p *picker) Pick(info balancer.PickInfo) (balancer.PickResult, error) {
	for i, res := range p.sc {
		if i > p.lastPick {
			if p.ready[i] {
				p.lastPick = i
				return balancer.PickResult{SubConn: res}, nil
			}
		}
	}
	return balancer.PickResult{}, fmt.Errorf("No servers available")
}

func (p *picker) Add(sc balancer.SubConn, addr resolver.Address) {
	p.sc = append(p.sc, sc)
	p.resolved = append(p.resolved, false)
	p.ready = append(p.ready, false)
	p.address = append(p.address, addr)
	sc.Connect()
}

func (p *picker) allReady() bool {
	for _, r := range p.resolved {
		if !r {
			return false
		}
	}
	return true
}

func (p *picker) Ready(sc balancer.SubConn) bool {
	for i, isc := range p.sc {
		if sc == isc {
			p.resolved[i] = true
			p.ready[i] = true
		}
	}
	return p.allReady()
}

func (p *picker) Failure(sc balancer.SubConn) bool {
	for i, isc := range p.sc {
		if sc == isc {
			p.resolved[i] = true
			unregister(p.address[i].Addr)
		}
	}
	return p.allReady()
}

func unregister(addr string) {
	conn, err := grpc.Dial("192.168.86.249:50055", grpc.WithInsecure())
	if err != nil {
		return
	}
	client := pbd.NewDiscoveryServiceV2Client(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	client.Unregister(ctx, &pbd.UnregisterRequest{Address: addr})
}

func init() {
	balancer.Register(newPickfirstBuilder())
}
