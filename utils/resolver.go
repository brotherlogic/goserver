package utils

import (
	"fmt"

	"google.golang.org/grpc/resolver"
)

const (
	discoveryScheme = "discovery"
)

type DiscoveryServerResolverBuilder struct{}

func (*DiscoveryServerResolverBuilder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	r := &discoveryResolver{
		cc: cc,
	}
	err := r.fillServer(target.Endpoint)
	return r, err
}
func (*DiscoveryServerResolverBuilder) Scheme() string { return discoveryScheme }

type DiscoveryClientResolverBuilder struct{}

func (*DiscoveryClientResolverBuilder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	r := &discoveryResolver{
		cc: cc,
	}
	err := r.fillClient(target.Endpoint)
	return r, err
}
func (*DiscoveryClientResolverBuilder) Scheme() string { return discoveryScheme }

type discoveryResolver struct {
	cc resolver.ClientConn
}

func (r *discoveryResolver) fillServer(endpoint string) error {
	entries, err := ResolveV3(endpoint)
	if err != nil {
		return err
	}

	addrs := make([]resolver.Address, len(entries))
	for i, ent := range entries {
		addrs[i] = resolver.Address{Addr: fmt.Sprintf("%v:%v", ent.Ip, ent.Port)}
	}
	r.cc.UpdateState(resolver.State{Addresses: addrs})
	return nil
}

func (r *discoveryResolver) fillClient(endpoint string) error {
	entries, err := ResolveV3Client(endpoint)
	if err != nil {
		return err
	}

	addrs := make([]resolver.Address, len(entries))
	for i, ent := range entries {
		addrs[i] = resolver.Address{Addr: fmt.Sprintf("%v:%v", ent.Ip, ent.Port)}
	}
	if len(addrs) == 0 {
		return fmt.Errorf("Unable to resolve %v", endpoint)
	}
	r.cc.UpdateState(resolver.State{Addresses: addrs})
	return nil
}

func (*discoveryResolver) ResolveNow(o resolver.ResolveNowOptions) {}
func (*discoveryResolver) Close()                                  {}
