package goserver

import (
	"log"
	"net"
	"os"
	"strconv"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	pb "github.com/brotherlogic/discovery/proto"
	pbl "github.com/brotherlogic/goserver/proto"
	pbd "github.com/brotherlogic/monitor/monitorproto"
	"github.com/golang/protobuf/proto"
)

type osHostGetter struct{}

func (hostGetter osHostGetter) Hostname() (string, error) {
	return os.Hostname()
}

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

func (s *GoServer) close(conn *grpc.ClientConn) {
	if conn != nil {
		conn.Close()
	}
}

// RegisterServingTask registers tasks to run when serving
func (s *GoServer) RegisterServingTask(task func()) {
	s.servingFuncs = append(s.servingFuncs, task)
}

// IsAlive Reports liveness of the server
func (s *GoServer) IsAlive(ctx context.Context, in *pbl.Alive) (*pbl.Alive, error) {
	return &pbl.Alive{}, nil
}

func (s *GoServer) getRegisteredServerPort(IP string, servername string, external bool) int32 {
	log.Printf("HERE with %v and %v", IP, servername)
	return s.registerServer(IP, servername, external, grpcDialler{}, mainBuilder{}, osHostGetter{})
}

//Save a protobuf
func (s *GoServer) Save(key string, p proto.Message) error {
	return s.ksclient.Save(key, p)
}

//Read a protobuf
func (s *GoServer) Read(key string, typ proto.Message) (proto.Message, error) {
	return s.ksclient.Load(key, typ)
}

// Serve Runs the server
func (s *GoServer) Serve() {
	log.Printf("%v is serving!", s)
	lis, err := net.Listen("tcp", ":"+strconv.Itoa(int(s.port)))
	if err != nil {
		log.Fatalf("Unable to grab port: %v -> %v", s, err)
	}
	server := grpc.NewServer()
	s.Register.DoRegister(server)
	pbl.RegisterGoserverServiceServer(server, s)
	s.setupHeartbeats(s.dialler, s.clientBuilder)

	// Background all the serving funcs
	for _, f := range s.servingFuncs {
		go f()
	}

	server.Serve(lis)
}
