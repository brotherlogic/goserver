package goserver

import (
	"errors"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"log"
	"testing"
	"time"

	pb "github.com/brotherlogic/discovery/proto"
	pbd "github.com/brotherlogic/monitor/proto"
)

type passingDialler struct{}

func (dialler passingDialler) Dial(host string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	return &grpc.ClientConn{}, nil
}

type failingDialler struct{}

func (dialler failingDialler) Dial(host string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	return &grpc.ClientConn{}, errors.New("Built to fail")
}

type passingDiscoveryServiceClient struct{}

func (DiscoveryServiceClient passingDiscoveryServiceClient) RegisterService(ctx context.Context, in *pb.RegistryEntry, opts ...grpc.CallOption) (*pb.RegistryEntry, error) {
	return &pb.RegistryEntry{Port: 35}, nil
}

func (DiscoveryServiceClient passingDiscoveryServiceClient) Discover(ctx context.Context, in *pb.RegistryEntry, opts ...grpc.CallOption) (*pb.RegistryEntry, error) {
	return &pb.RegistryEntry{}, nil
}

func (DiscoveryServiceClient passingDiscoveryServiceClient) ListAllServices(ctx context.Context, in *pb.Empty, opts ...grpc.CallOption) (*pb.ServiceList, error) {
	return &pb.ServiceList{}, nil
}

type failingDiscoveryServiceClient struct{}

func (DiscoveryServiceClient failingDiscoveryServiceClient) RegisterService(ctx context.Context, in *pb.RegistryEntry, opts ...grpc.CallOption) (*pb.RegistryEntry, error) {
	return &pb.RegistryEntry{}, errors.New("Built to fail")
}

func (DiscoveryServiceClient failingDiscoveryServiceClient) Discover(ctx context.Context, in *pb.RegistryEntry, opts ...grpc.CallOption) (*pb.RegistryEntry, error) {
	return &pb.RegistryEntry{}, nil
}

func (DiscoveryServiceClient failingDiscoveryServiceClient) ListAllServices(ctx context.Context, in *pb.Empty, opts ...grpc.CallOption) (*pb.ServiceList, error) {
	return &pb.ServiceList{}, nil
}

type passingBuilder struct{}

func (clientBuilder passingBuilder) NewDiscoveryServiceClient(conn *grpc.ClientConn) pb.DiscoveryServiceClient {
	return passingDiscoveryServiceClient{}
}

type failingBuilder struct{}

func (clientBuilder failingBuilder) NewDiscoveryServiceClient(conn *grpc.ClientConn) pb.DiscoveryServiceClient {
	return failingDiscoveryServiceClient{}
}

type passingMonitorServiceClient struct{}

func (MonitorServiceClient passingMonitorServiceClient) ReceiveHeartbeat(ctx context.Context, in *pb.RegistryEntry, opts ...grpc.CallOption) (*pbd.Empty, error) {
	return &pbd.Empty{}, nil
}

type passingMonitorBuilder struct{}

func (monitorBuilder passingMonitorBuilder) NewMonitorServiceClient(conn *grpc.ClientConn) pbd.MonitorServiceClient {
	return passingMonitorServiceClient{}
}

func TestFailToDial(t *testing.T) {
	server := GoServer{}
	madeupport := server.registerServer("madeup", "madeup", false, failingDialler{}, passingBuilder{})

	if madeupport > 0 {
		t.Errorf("Dial failure did not lead to bad port")
	}
}

func TestFailToRegister(t *testing.T) {
	server := GoServer{}
	madeupport := server.registerServer("madeup", "madeup", false, passingDialler{}, failingBuilder{})

	if madeupport > 0 {
		t.Errorf("Dial failure did not lead to bad port")
	}

}

func TestRegisterServer(t *testing.T) {
	server := GoServer{}
	madeupport := server.registerServer("madeup", "madeup", false, passingDialler{}, passingBuilder{})

	if madeupport != 35 {
		t.Errorf("Port number is wrong: %v", madeupport)
	}
}

func TestGetIP(t *testing.T) {
	ip := getLocalIP()
	if ip == "" || ip == "127.0.0.1" {
		t.Errorf("Get IP is returning the wrong address: %v", ip)
	}
}

type TestServer struct {
	GoServer
}

func InitTestServer() TestServer {
	s := TestServer{}
	s.PrepServer()
	s.monitorBuilder = passingMonitorBuilder{}
	s.dialler = passingDialler{}
	s.heartbeatTime = time.Millisecond
	return s
}

func (s *TestServer) Serve() {
	log.Printf("Serving!")
	go s.heartbeat()
}

func TestHeartbeat(t *testing.T) {
	server := InitTestServer()
	server.Serve()

	//Wait 10 seconds
	time.Sleep(10 * time.Millisecond)

	server.teardown()
	if server.heartbeatCount < 9 {
		t.Errorf("Did not deliver heartbeats")
	}
}

func TestRegister(t *testing.T) {
	server := GoServer{}
	server.Register(&grpc.Server{})
}
