package goserver

import (
	"log"
	"net"
	"os"
	"strconv"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	pb "github.com/brotherlogic/discovery/proto"
	pbl "github.com/brotherlogic/goserver/proto"
	pbd "github.com/brotherlogic/monitor/monitorproto"
	"github.com/golang/protobuf/proto"
)

func (s *GoServer) suicideWatch() {
	for true && s.Killme {
		time.Sleep(s.suicideTime)
		//commit suicide if we're detached from the parent
		if os.Getppid() == 1 && s.Killme {
			s.LogFunction("death-"+strconv.Itoa(os.Getppid()), time.Now())
			os.Exit(1)
		}
	}
}

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
	s.Servername = servername
	s.Port = s.getRegisteredServerPort(getLocalIP(), s.Servername, external)
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
	return &pbl.Alive{Name: s.Registry.GetName()}, nil
}

// Mote promotes or demotes a server into production
func (s *GoServer) Mote(ctx context.Context, in *pbl.MoteRequest) (*pbl.Empty, error) {
	t := time.Now()
	err := s.Register.Mote(in.Master)

	// If we were able to mote then we should inform discovery
	if err == nil {
		s.Registry.Master = in.Master
		s.reregister(s.dialler, s.clientBuilder)
		s.LogFunction("MasterMote-pass", t)
	} else {
		s.LogFunction("MasterMote-fail", t)
	}

	return &pbl.Empty{}, err
}

func (s *GoServer) getRegisteredServerPort(IP string, servername string, external bool) int32 {
	return s.registerServer(IP, servername, external, grpcDialler{}, mainBuilder{}, osHostGetter{})
}

//Save a protobuf
func (s *GoServer) Save(key string, p proto.Message) error {
	return s.KSclient.Save(key, p)
}

//Read a protobuf
func (s *GoServer) Read(key string, typ proto.Message) (proto.Message, error) {
	return s.KSclient.Read(key, typ)
}

// Serve Runs the server
func (s *GoServer) Serve() {
	lis, err := net.Listen("tcp", ":"+strconv.Itoa(int(s.Port)))
	if err != nil {
		log.Fatalf("Unable to grab port: %v -> %v", s, err)
	}
	server := grpc.NewServer()
	s.Register.DoRegister(server)
	pbl.RegisterGoserverServiceServer(server, s)
	s.setupHeartbeats()

	go s.suicideWatch()

	// Background all the serving funcs
	for _, f := range s.servingFuncs {
		go f()
	}

	log.Printf("%v is Serving", s.Registry)
	server.Serve(lis)
}
