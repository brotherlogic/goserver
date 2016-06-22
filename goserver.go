package goserver

import (
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"log"
	"strconv"

	pb "github.com/brotherlogic/discovery/proto"
)

const (
	registryIP   = "10.0.1.17"
	registryPort = 50052
)

// GoServer The basic server construct
type GoServer struct {
}

type dialler interface {
	Dial(host string, opts ...grpc.DialOption) (*grpc.ClientConn, error)
}

type clientBuilder interface {
	NewDiscoveryServiceClient(conn *grpc.ClientConn) pb.DiscoveryServiceClient
}

// RegisterServer Registers a server with the system and gets the port number it should use
func (s *GoServer) registerServer(IP string, servername string, dialler dialler, builder clientBuilder) int32 {
	conn, err := dialler.Dial(registryIP+":"+strconv.Itoa(registryPort), grpc.WithInsecure())
	if err != nil {
		log.Printf("Could not connect: %v", err)
		return -1
	}

	defer conn.Close()
	registry := builder.NewDiscoveryServiceClient(conn)

	entry := pb.RegistryEntry{Ip: IP, Name: servername}
	r, err := registry.RegisterService(context.Background(), &entry)
	if err != nil {
		log.Printf("Could not register service: %v", err)
		return -1
	}

	return r.Port
}
