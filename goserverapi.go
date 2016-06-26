package goserver

import (
	"log"
	"net"
	"strconv"

	"google.golang.org/grpc"

	pb "github.com/brotherlogic/discovery/proto"
	pbd "github.com/brotherlogic/monitor/monitorproto"
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
func (s *GoServer) RegisterServer(servername string, external bool) {
	s.servername = servername
	s.port = s.getRegisteredServerPort(getLocalIP(), s.servername, external)
}

func (s *GoServer) getRegisteredServerPort(IP string, servername string, external bool) int32 {
	log.Printf("HERE with %v and %v", IP, servername)
	return s.registerServer(IP, servername, external, grpcDialler{}, mainBuilder{})
}

// Serve Runs the server
func (s *GoServer) Serve() {
	log.Printf("%v is serving!", s)
	lis, _ := net.Listen("tcp", ":"+strconv.Itoa(int(s.port)))
	server := grpc.NewServer()
	s.Register.DoRegister(server)
	s.setupHeartbeats(s.dialler, s.clientBuilder)
	server.Serve(lis)
}
