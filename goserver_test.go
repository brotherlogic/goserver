package goserver

import (
	"errors"
	"fmt"
	"log"
	"sync"
	"testing"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	pb "github.com/brotherlogic/discovery/proto"
	pbg "github.com/brotherlogic/goserver/proto"
	pbd "github.com/brotherlogic/monitor/proto"
)

type basicGetter struct{}

func (hostGetter basicGetter) Hostname() (string, error) {
	return "basic", nil
}

type failingGetter struct{}

func (hostGetter failingGetter) Hostname() (string, error) {
	return "", errors.New("Built to Fail")
}

type passingDialler struct{}

func (dialler passingDialler) Dial(host string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	return nil, nil
}

type failingDialler struct{}

func (dialler failingDialler) Dial(host string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	return nil, errors.New("Built to fail")
}

type passingDiscoveryServiceClient struct {
	upreregister     int32
	returnWeakMaster bool
	returnMaster     bool
}

func (DiscoveryServiceClient passingDiscoveryServiceClient) RegisterService(ctx context.Context, in *pb.RegisterRequest, opts ...grpc.CallOption) (*pb.RegisterResponse, error) {
	return &pb.RegisterResponse{Service: &pb.RegistryEntry{Port: 35 + DiscoveryServiceClient.upreregister, Identifier: in.GetService().Identifier, WeakMaster: DiscoveryServiceClient.returnWeakMaster, Master: DiscoveryServiceClient.returnMaster}}, nil
}

func (DiscoveryServiceClient passingDiscoveryServiceClient) Discover(ctx context.Context, in *pb.DiscoverRequest, opts ...grpc.CallOption) (*pb.DiscoverResponse, error) {
	return &pb.DiscoverResponse{Service: &pb.RegistryEntry{Ip: "10.10.10.10", Port: 23}}, nil
}

func (DiscoveryServiceClient passingDiscoveryServiceClient) ListAllServices(ctx context.Context, in *pb.ListRequest, opts ...grpc.CallOption) (*pb.ListResponse, error) {
	return &pb.ListResponse{}, nil
}

func (DiscoveryServiceClient passingDiscoveryServiceClient) State(ctx context.Context, in *pb.StateRequest, opts ...grpc.CallOption) (*pb.StateResponse, error) {
	return &pb.StateResponse{}, nil
}

type failPassDiscoveryServiceClient struct {
	fails int
}

func (DiscoveryServiceClient failPassDiscoveryServiceClient) RegisterService(ctx context.Context, in *pb.RegisterRequest, opts ...grpc.CallOption) (*pb.RegisterResponse, error) {
	return &pb.RegisterResponse{Service: &pb.RegistryEntry{Port: 35, Identifier: in.GetService().Identifier}}, nil
}

func (DiscoveryServiceClient *failPassDiscoveryServiceClient) Discover(ctx context.Context, in *pb.DiscoverRequest, opts ...grpc.CallOption) (*pb.DiscoverResponse, error) {
	if DiscoveryServiceClient.fails > 0 {
		DiscoveryServiceClient.fails--
		return nil, grpc.Errorf(codes.Unavailable, "Made up failure %v", 23)
	}
	return &pb.DiscoverResponse{Service: &pb.RegistryEntry{Ip: "10.10.10.10", Port: 23}}, nil
}

func (DiscoveryServiceClient failPassDiscoveryServiceClient) ListAllServices(ctx context.Context, in *pb.ListRequest, opts ...grpc.CallOption) (*pb.ListResponse, error) {
	return &pb.ListResponse{}, nil
}

func (DiscoveryServiceClient failPassDiscoveryServiceClient) State(ctx context.Context, in *pb.StateRequest, opts ...grpc.CallOption) (*pb.StateResponse, error) {
	return &pb.StateResponse{}, nil
}

type failingDiscoveryServiceClient struct{}

func (DiscoveryServiceClient failingDiscoveryServiceClient) RegisterService(ctx context.Context, in *pb.RegisterRequest, opts ...grpc.CallOption) (*pb.RegisterResponse, error) {
	return &pb.RegisterResponse{}, grpc.Errorf(codes.Internal, "Built to Fail")
}

func (DiscoveryServiceClient failingDiscoveryServiceClient) Discover(ctx context.Context, in *pb.DiscoverRequest, opts ...grpc.CallOption) (*pb.DiscoverResponse, error) {
	return &pb.DiscoverResponse{}, errors.New("Built to fail")
}

func (DiscoveryServiceClient failingDiscoveryServiceClient) ListAllServices(ctx context.Context, in *pb.ListRequest, opts ...grpc.CallOption) (*pb.ListResponse, error) {
	return &pb.ListResponse{}, nil
}
func (DiscoveryServiceClient failingDiscoveryServiceClient) State(ctx context.Context, in *pb.StateRequest, opts ...grpc.CallOption) (*pb.StateResponse, error) {
	return &pb.StateResponse{}, nil
}

type passingBuilder struct {
	returnMaster     bool
	returnWeakMaster bool
}

func (clientBuilder passingBuilder) NewDiscoveryServiceClient(conn *grpc.ClientConn) pb.DiscoveryServiceClient {
	return passingDiscoveryServiceClient{returnMaster: clientBuilder.returnMaster, returnWeakMaster: clientBuilder.returnWeakMaster}
}

type passingFailBuilder struct{}

func (clientBuilder passingFailBuilder) NewDiscoveryServiceClient(conn *grpc.ClientConn) pb.DiscoveryServiceClient {
	return &failPassDiscoveryServiceClient{fails: 1}
}

type failingBuilder struct{}

func (clientBuilder failingBuilder) NewDiscoveryServiceClient(conn *grpc.ClientConn) pb.DiscoveryServiceClient {
	return failingDiscoveryServiceClient{}
}

type passingMonitorServiceClient struct {
	failLog bool
}

func (MonitorServiceClient passingMonitorServiceClient) ReadMessageLogs(ctx context.Context, in *pb.RegistryEntry, opts ...grpc.CallOption) (*pbd.MessageLogReadResponse, error) {
	return &pbd.MessageLogReadResponse{}, nil
}

func (MonitorServiceClient passingMonitorServiceClient) WriteMessageLog(ctx context.Context, in *pbd.MessageLog, opts ...grpc.CallOption) (*pbd.LogWriteResponse, error) {
	if MonitorServiceClient.failLog {
		return &pbd.LogWriteResponse{}, grpc.Errorf(codes.Internal, "Built to fail")
	}
	return &pbd.LogWriteResponse{}, nil
}

type passingMonitorBuilder struct {
	failLog bool
}

func (monitorBuilder passingMonitorBuilder) NewMonitorServiceClient(conn *grpc.ClientConn) pbd.MonitorServiceClient {
	return passingMonitorServiceClient{failLog: monitorBuilder.failLog}
}

func run(ctx context.Context) error {
	log.Printf("Run")
	return nil
}

func TestRegister(t *testing.T) {
	server := GoServer{}
	server.SkipLog = true
	server.PrepServer()
	server.RegisterRepeatingTask(run, "test_task", time.Second*5)
}

func TestNoRegister(t *testing.T) {
	server := GoServer{}
	server.SkipLog = true
	server.PrepServerNoRegister(int32(50055))
}

func TestCPUGet(t *testing.T) {
	server := GoServer{}
	server.cpuMutex = &sync.Mutex{}
	cpu, mem := server.getCPUUsage()

	if cpu <= 0 || mem <= 0 {
		t.Errorf("Bad pull from cpu or mem: %v and %v", cpu, mem)
	}
}

func TestToggleSudo(t *testing.T) {
	server := GoServer{}
	server.RunSudo()
	if !server.Sudo {
		t.Errorf("Sudo not enabled")
	}
}

func TestFailToDial(t *testing.T) {
	server := GoServer{SkipIssue: true}
	madeupport, _ := server.registerServer("madeup", "madeup", false, false, false, failingDialler{}, passingBuilder{}, basicGetter{})

	if madeupport > 0 {
		t.Errorf("Dial failure did not lead to bad port")
	}
}

func TestFailToRegister(t *testing.T) {
	server := GoServer{SkipIssue: true}
	madeupport, _ := server.registerServer("madeup", "madeup", false, false, false, passingDialler{}, failingBuilder{}, basicGetter{})

	if madeupport > 0 {
		t.Errorf("Dial failure did not lead to bad port")
	}
}

func TestFailToGet(t *testing.T) {
	server := GoServer{SkipIssue: true}
	server.registerServer("madeup", "madeup", false, false, false, passingDialler{}, passingBuilder{}, failingGetter{})

	if server.Registry.Identifier != "Server-madeup" {
		t.Errorf("Server has not registered correctly: %v", server.Registry)
	}
}

func TestStraightDial(t *testing.T) {
	server := GoServer{SkipIssue: true}
	_, err := server.Dial("madeup", passingDialler{}, passingBuilder{})
	if err != nil {
		t.Errorf("Dial has failed: %v", err)
	}
}

func TestFailedDialler(t *testing.T) {
	server := GoServer{SkipIssue: true}
	_, err := server.Dial("madeup", failingDialler{}, passingBuilder{})
	if err == nil {
		t.Errorf("Dial has failed: %v", err)
	}
}

func TestBadRegistry(t *testing.T) {
	server := GoServer{SkipIssue: true}
	_, err := server.Dial("madeup", passingDialler{}, failingBuilder{})
	if err == nil {
		t.Errorf("Dial has failed: %v", err)
	}
}

func TestGetIPSuccess(t *testing.T) {
	server := GoServer{SkipIssue: true}
	server.clientBuilder = passingBuilder{}
	server.dialler = passingDialler{}
	_, port := server.GetIP("madeup")
	if port < 0 {
		t.Errorf("Get IP has failed")
	}
}

func TestGetIPOneFail(t *testing.T) {
	server := GoServer{SkipIssue: true}
	server.clientBuilder = passingFailBuilder{}
	server.dialler = passingDialler{}
	_, port := server.GetIP("madeup")
	if port < 0 {
		t.Errorf("Get IP has failed")
	}
}

func TestGetIPFail(t *testing.T) {
	server := GoServer{SkipIssue: true}
	server.clientBuilder = failingBuilder{}
	server.dialler = passingDialler{}
	_, port := server.GetIP("madeup")
	if port >= 0 {
		t.Errorf("Failing builder has not failed")
	}
}

func TestBadRegister(t *testing.T) {
	server := GoServer{SkipIssue: true}
	server.reregister(failingDialler{}, passingBuilder{})
}

func TestLameDuckRegister(t *testing.T) {
	server := GoServer{SkipIssue: true}
	server.SkipLog = true
	server.PrepServer()

	server.Registry = &pb.RegistryEntry{Port: 23, Master: true}
	server.LameDuck = true
	server.reregister(passingDialler{}, passingBuilder{})

	if server.Registry.Master {
		t.Errorf("Server is still master")
	}
}

func TestBadReregister(t *testing.T) {
	server := GoServer{SkipIssue: true}
	server.Registry = &pb.RegistryEntry{Port: 23}
	server.reregister(passingDialler{}, passingBuilder{})

	if server.badPorts != 1 {
		t.Errorf("Port failed: %v", server.Registry)
	}
}
func TestRegisterServer(t *testing.T) {
	server := GoServer{SkipIssue: true}
	madeupport, _ := server.registerServer("madeup", "madeup", false, false, false, passingDialler{}, passingBuilder{}, basicGetter{})

	if madeupport != 35 {
		t.Errorf("Port number is wrong: %v", madeupport)
	}

	server.reregister(passingDialler{}, passingBuilder{})

	if server.Registry.GetPort() != 35 {
		t.Errorf("Not stored registry info: %v", server.Registry)
	}
}

func TestRegisterDemoteServer(t *testing.T) {
	server := GoServer{SkipLog: true, SkipIssue: true}
	madeupport, _ := server.registerServer("madeup", "madeup", false, false, false, passingDialler{}, passingBuilder{}, basicGetter{})

	if madeupport != 35 {
		t.Errorf("Port number is wrong: %v", madeupport)
	}

	//Re-register as Master
	server.Registry.Master = true
	server.reregister(passingDialler{}, passingBuilder{})

	//Re-register and fail heartbeatTime
	server.reregister(passingDialler{}, failingBuilder{})

	if server.Registry.Master {
		t.Errorf("Registry has not demoted: %v", server.Registry)
	}
}

func TestLog(t *testing.T) {
	server := InitTestServer()
	server.Log("MadeUpLog")
}

func TestGetIP(t *testing.T) {
	ip := getLocalIP()
	if ip == "" || ip == "127.0.0.1" {
		t.Errorf("Get IP is returning the wrong address: %v", ip)
	}
}

type TestServer struct {
	*GoServer
	failMote bool
}

func (s TestServer) DoRegister(server *grpc.Server) {
	//Do Nothing
}

func (s TestServer) ReportHealth() bool {
	return true
}

func (s TestServer) Mote(ctx context.Context, master bool) error {
	if s.failMote {
		return fmt.Errorf("Built to fail")
	}
	return nil
}

func (s TestServer) Shutdown(ctx context.Context) error {
	return nil
}

func (s TestServer) GetState() []*pbg.State {
	return []*pbg.State{}
}

func InitTestServer() TestServer {
	s := TestServer{&GoServer{}, false}
	s.Register = s
	s.SkipLog = true
	s.PrepServer()
	s.monitorBuilder = passingMonitorBuilder{}
	s.dialler = passingDialler{}
	s.heartbeatTime = time.Millisecond
	s.clientBuilder = passingBuilder{}
	s.Registry = &pb.RegistryEntry{Name: "testserver"}
	log.Printf("Set heartbeat time")
	s.SkipLog = true
	return s
}

func InitTestServerWithOptions(failMote bool) TestServer {
	s := TestServer{&GoServer{}, failMote}
	s.Register = s
	s.SkipLog = true
	s.PrepServer()
	s.monitorBuilder = passingMonitorBuilder{}
	s.dialler = passingDialler{}
	s.heartbeatTime = time.Millisecond
	s.clientBuilder = passingBuilder{}
	s.Registry = &pb.RegistryEntry{Name: "testserver"}
	log.Printf("Set heartbeat time")
	return s
}

func TestHeartbeat(t *testing.T) {
	server := InitTestServer()
	server.SkipLog = true
	go server.Serve()
	log.Printf("Done Serving")

	time.Sleep(20 * time.Millisecond)
	log.Printf("Tearing Down")
	server.teardown()
}

func TestMasterRegister(t *testing.T) {
	server := InitTestServer()
	server.PrepServer()
	server.Registry = &pb.RegistryEntry{Port: 23, Master: false}

	server.reregister(passingDialler{}, passingBuilder{returnWeakMaster: true})

	if server.Registry.Master {
		t.Errorf("Master has been recorded")
	}
}

func TestMasterRegisterFailMote(t *testing.T) {
	server := InitTestServerWithOptions(true)
	server.PrepServer()
	server.Registry = &pb.RegistryEntry{Port: 23, Master: false}

	server.reregister(passingDialler{}, passingBuilder{returnWeakMaster: true})

	if server.Registry.Master {
		t.Errorf("Master has been recorded")
	}
}

func TestDoLog(t *testing.T) {
	server := InitTestServer()
	server.Log("hello")
	server.Log("hello")
	if server.logsSkipped != 0 {
		t.Errorf("Missed double log skip")
	}

}
