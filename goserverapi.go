package goserver

import (
	"google.golang.org/grpc"
	"log"
	"net"
	"strconv"

	pb "github.com/brotherlogic/discovery/proto"
)

type grpcDialler struct{}

func (dialler grpcDialler) Dial(host string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	return grpc.Dial(host, opts...)
}

type mainBuilder struct{}

func (clientBuilder mainBuilder) NewDiscoveryServiceClient(conn *grpc.ClientConn) pb.DiscoveryServiceClient {
	return pb.NewDiscoveryServiceClient(conn)
}

// RegisterServer registers this server
func (s *GoServer) RegisterServer(external bool) {
	s.port = s.getRegisteredServerPort(getLocalIP(), s.servername, external)
}

func (s *GoServer) getRegisteredServerPort(IP string, servername string, external bool) int32 {
	return s.registerServer(IP, servername, external, grpcDialler{}, mainBuilder{})
}

// Serve Runs the server
func (s *GoServer) Serve(port int) {
	log.Printf("%v is serving!", s)
	lis, _ := net.Listen("tcp", ":"+strconv.Itoa(int(s.port)))
	server := grpc.NewServer()
	s.Register(server)
	server.Serve(lis)
}
