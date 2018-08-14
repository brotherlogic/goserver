package goserver

import (
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/brotherlogic/goserver/utils"
	"github.com/golang/protobuf/proto"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/brotherlogic/discovery/proto"
	pbl "github.com/brotherlogic/goserver/proto"
	pbks "github.com/brotherlogic/keystore/proto"
	pbd "github.com/brotherlogic/monitor/monitorproto"

	//Needed to pull in gzip encoding init
	_ "google.golang.org/grpc/encoding/gzip"
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
func (s *GoServer) RegisterServer(servername string, external bool) bool {
	s.Servername = servername
	s.Port = s.getRegisteredServerPort(getLocalIP(), s.Servername, external)
	return s.Port > 0
}

func (s *GoServer) close(conn *grpc.ClientConn) {
	if conn != nil {
		conn.Close()
	}
}

// RegisterServingTask registers tasks to run when serving
func (s *GoServer) RegisterServingTask(task func(ctx context.Context)) {
	s.servingFuncs = append(s.servingFuncs, sFunc{fun: task, d: 0})
}

// RegisterRepeatingTask registers a repeating task with a given frequency
func (s *GoServer) RegisterRepeatingTask(task func(ctx context.Context), freq time.Duration) {
	s.servingFuncs = append(s.servingFuncs, sFunc{fun: task, d: freq})
}

// RegisterRepeatingTaskNonMaster registers a repeating task with a given frequency
func (s *GoServer) RegisterRepeatingTaskNonMaster(task func(ctx context.Context), freq time.Duration) {
	s.servingFuncs = append(s.servingFuncs, sFunc{fun: task, d: freq, nm: true})
}

// IsAlive Reports liveness of the server
func (s *GoServer) IsAlive(ctx context.Context, in *pbl.Alive) (*pbl.Alive, error) {
	if s.Register.ReportHealth() {
		return &pbl.Alive{Name: s.Registry.GetName()}, nil
	}
	return nil, errors.New("Server reports unhealthy")
}

//State gets the state of the server.
func (s *GoServer) State(ctx context.Context, in *pbl.Empty) (*pbl.ServerState, error) {
	states := s.Register.GetState()
	states = append(states, &pbl.State{Key: "hearts", Value: int64(s.hearts)})
	states = append(states, &pbl.State{Key: "bad_hearts", Value: int64(s.badHearts)})
	states = append(states, &pbl.State{Key: "fail_master", Value: int64(s.failMaster)})
	states = append(states, &pbl.State{Key: "fail_log", Value: int64(s.failLogs)})
	states = append(states, &pbl.State{Key: "fail_message", Text: s.failMessage})
	return &pbl.ServerState{States: states}, nil
}

// Mote promotes or demotes a server into production
func (s *GoServer) Mote(ctx context.Context, in *pbl.MoteRequest) (*pbl.Empty, error) {
	err := s.Register.Mote(ctx, in.Master)

	// If we were able to mote then we should inform discovery
	if err == nil {
		s.Registry.Master = in.Master
		s.reregister(s.dialler, s.clientBuilder)
	}

	return &pbl.Empty{}, err
}

func (s *GoServer) getRegisteredServerPort(IP string, servername string, external bool) int32 {
	return s.registerServer(IP, servername, external, grpcDialler{}, mainBuilder{}, osHostGetter{})
}

//Save a protobuf
func (s *GoServer) Save(ctx context.Context, key string, p proto.Message) error {
	return s.KSclient.Save(ctx, key, p)
}

func (s *GoServer) run(t sFunc) {
	time.Sleep(time.Minute)
	if t.d == 0 {
		ctx, cancel := utils.BuildContext(fmt.Sprintf("%v-NoD-Repeat", s.Registry.Name), s.Registry.Name, pbl.ContextType_INFINITE)
		defer cancel()
		t.fun(ctx)
	} else {
		for true {
			if s.Registry.GetMaster() || t.nm {
				ctx, cancel := utils.BuildContext(fmt.Sprintf("%v-Repeat-%v", s.Registry.Name, t.d), s.Registry.Name, pbl.ContextType_LONG)
				defer cancel()
				t.fun(ctx)
			}
			time.Sleep(t.d)
		}
	}
}

//Read a protobuf
func (s *GoServer) Read(ctx context.Context, key string, typ proto.Message) (proto.Message, *pbks.ReadResponse, error) {
	return s.KSclient.Read(ctx, key, typ)
}

//GetServers gets an IP address from the discovery server
func (s *GoServer) GetServers(servername string) ([]*pb.RegistryEntry, error) {
	conn, err := s.dialler.Dial(utils.RegistryIP+":"+strconv.Itoa(utils.RegistryPort), grpc.WithInsecure())
	if err == nil {
		registry := s.clientBuilder.NewDiscoveryServiceClient(conn)
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		r, err := registry.ListAllServices(ctx, &pb.ListRequest{}, grpc.FailFast(false))
		e, ok := status.FromError(err)
		if ok && e.Code() == codes.Unavailable {
			r, err = registry.ListAllServices(ctx, &pb.ListRequest{}, grpc.FailFast(false))
		}

		if err == nil {
			s.close(conn)
			arr := make([]*pb.RegistryEntry, 0)
			for _, s := range r.GetServices().GetServices() {
				if s.GetName() == servername {
					arr = append(arr, s)
				}
			}
			return arr, nil
		}
	}
	s.close(conn)
	return nil, fmt.Errorf("Unable to establish connection")
}

// Serve Runs the server
func (s *GoServer) Serve() error {
	lis, err := net.Listen("tcp", ":"+strconv.Itoa(int(s.Port)))
	if err != nil {
		return err
	}
	server := grpc.NewServer(
		grpc.RPCCompressor(grpc.NewGZIPCompressor()),
		grpc.RPCDecompressor(grpc.NewGZIPDecompressor()),
		grpc.MaxRecvMsgSize(1024*1024*1024),
		grpc.MaxSendMsgSize(1024*1024*1024),
	)
	s.Register.DoRegister(server)
	pbl.RegisterGoserverServiceServer(server, s)
	s.setupHeartbeats()

	go s.suicideWatch()

	// Background all the serving funcs
	for _, f := range s.servingFuncs {
		go s.run(f)
	}

	server.Serve(lis)
	return nil
}
