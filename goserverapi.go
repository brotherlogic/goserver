package goserver

import (
	"google.golang.org/grpc"
	"log"
	"net"
	"strconv"

	pb "github.com/brotherlogic/discovery/proto"
	pbd "github.com/brotherlogic/monitor/proto"
)

type grpcDialler struct{}

func (dialler grpcDialler) Dial(host string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	return grpc.Dial(host, opts...)
}

type mainMonitorBuilder struct{}

func (monitorBuilder mainMonitorBuilder) NewMonitorServiceClient(conn *grpc.ClientConn) pbd.MonitorServiceClient {
	return pbd.NewMonitorServiceClient(conn)
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
func (s *GoServer) Serve(servername string) {
	s.PrepServer(servername)
	s.monitorBuilder = mainMonitorBuilder{}
	s.dialler = grpcDialler{}

	log.Printf("%v is serving!", s)
	lis, _ := net.Listen("tcp", ":"+strconv.Itoa(int(s.port)))
	server := grpc.NewServer()
	s.Register(server)
	server.Serve(lis)
}
