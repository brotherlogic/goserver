package goserver

import (
	"context"
	"fmt"
	"math/rand"

	"github.com/brotherlogic/goserver/utils"
	"google.golang.org/grpc"

	dpb "github.com/brotherlogic/discovery/proto"
)

// FDialServer dial a specific job
func (s *GoServer) FDialServer(ctx context.Context, servername string) (*grpc.ClientConn, error) {
	if servername == "discover" {
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

// FDial fundamental dial
func (s *GoServer) FDial(host string) (*grpc.ClientConn, error) {
	return grpc.Dial(host, grpc.WithInsecure(), s.withClientUnaryInterceptor())
}