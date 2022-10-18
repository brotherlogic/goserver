package goserver

import (
	"context"
	"fmt"
	"math/rand"

	"github.com/brotherlogic/goserver/utils"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"

	dpb "github.com/brotherlogic/discovery/proto"
)

// FDialServer dial a specific job
func (s *GoServer) FDialServer(ctx context.Context, servername string) (*grpc.ClientConn, error) {
	if servername == "discovery" {
		return s.FDial(fmt.Sprintf("%v:%v", utils.LocalIP, utils.RegistryPort))
	}

	conn, err := s.FDial(utils.LocalDiscover)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	registry := dpb.NewDiscoveryServiceV2Client(conn)
	val, err := registry.Get(ctx, &dpb.GetRequest{Job: servername})
	if err != nil {
		return nil, err
	}

	// Pick a server at random
	servernum := rand.Intn(len(val.GetServices()))
	return s.FDial(fmt.Sprintf("%v:%v", val.GetServices()[servernum].GetIp(), val.GetServices()[servernum].GetPort()))
}

// FFind finds all servers
func (s *GoServer) FFind(ctx context.Context, servername string) ([]string, error) {
	if servername == "discovery" {
		return []string{}, fmt.Errorf("Cannot multi dial discovery")
	}

	conn, err := s.FDial(utils.LocalDiscover)
	if err != nil {
		return []string{}, err
	}
	defer conn.Close()

	registry := dpb.NewDiscoveryServiceV2Client(conn)
	val, err := registry.Get(ctx, &dpb.GetRequest{Job: servername})
	if err != nil {
		return []string{}, err
	}

	// Pick a server at random
	ret := []string{}
	for _, entry := range val.GetServices() {
		ret = append(ret, fmt.Sprintf("%v:%v", entry.Identifier, entry.Port))
	}
	return ret, nil
}

// FDialSpecificServer dial a specific job on a specific host
func (s *GoServer) FDialSpecificServer(ctx context.Context, servername string, host string) (*grpc.ClientConn, error) {
	if servername == "discovery" {
		return s.FDial(fmt.Sprintf("%v:%v", utils.LocalIP, utils.RegistryPort))
	}

	conn, err := s.FDial(utils.LocalDiscover)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	registry := dpb.NewDiscoveryServiceV2Client(conn)
	val, err := registry.Get(ctx, &dpb.GetRequest{Job: servername, Server: host})
	if err != nil {
		return nil, err
	}

	if len(val.GetServices()) == 0 {
		return nil, fmt.Errorf("Cannot locate server: %v,%v", servername, host)
	}

	// Pick a server at random
	return s.FDial(fmt.Sprintf("%v:%v", val.GetServices()[0].GetIp(), val.GetServices()[0].GetPort()))
}

// FFindSpecificServer dial a specific job on a specific host
func (s *GoServer) FFindSpecificServer(ctx context.Context, servername string, host string) (*dpb.RegistryEntry, error) {
	if servername == "discovery" {
		return &dpb.RegistryEntry{Ip: utils.LocalIP, Port: utils.RegistryPort}, nil
	}

	conn, err := s.FDial(utils.LocalDiscover)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	registry := dpb.NewDiscoveryServiceV2Client(conn)
	val, err := registry.Get(ctx, &dpb.GetRequest{Job: servername, Server: host})
	if err != nil {
		return nil, err
	}

	if len(val.GetServices()) == 0 {
		return nil, fmt.Errorf("Cannot locate server: %v,%v", servername, host)
	}

	// Pick a server at random
	return val.GetServices()[0], nil
}

// FDial fundamental dial
func (s *GoServer) FDial(host string) (*grpc.ClientConn, error) {
	return grpc.Dial(host, grpc.WithInsecure(),
		s.withClientUnaryInterceptor(),
		grpc.WithUnaryInterceptor(otelgrpc.UnaryClientInterceptor()))
}

// FPDial fundamental dial
func (s *GoServer) FPDial(host string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	opts = append(opts, grpc.WithInsecure())
	opts = append(opts, grpc.WithChainUnaryInterceptor(
		otelgrpc.UnaryClientInterceptor(),
		s.clientInterceptor,
	))
	return grpc.Dial(host, opts...)
}
