package goserver

import (
	"crypto/md5"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/brotherlogic/goserver/utils"
	"github.com/golang/protobuf/proto"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	pb "github.com/brotherlogic/discovery/proto"
	pbgh "github.com/brotherlogic/githubcard/proto"
	pbl "github.com/brotherlogic/goserver/proto"
	pbks "github.com/brotherlogic/keystore/proto"
	pbd "github.com/brotherlogic/monitor/monitorproto"
	pbt "github.com/brotherlogic/tracer/proto"

	ps "github.com/mitchellh/go-ps"

	//Needed to pull in gzip encoding init
	_ "google.golang.org/grpc/encoding/gzip"
)

func (s *GoServer) suicideWatch() {
	for true && s.Killme {
		time.Sleep(s.suicideTime)

		//commit suicide if we're detached from the parent and we're not sudoing
		if s.Killme {
			if s.Sudo {
				p, err := ps.FindProcess(os.Getppid())
				if err == nil && p.PPid() == 1 {
					os.Exit(1)
				}
			} else {
				if os.Getppid() == 1 && s.Killme {
					os.Exit(1)
				}
			}
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
func (s *GoServer) RegisterRepeatingTask(task func(ctx context.Context), key string, freq time.Duration) {
	s.servingFuncs = append(s.servingFuncs, sFunc{fun: task, d: freq, key: key})
	found := false
	for _, c := range s.config.Periods {
		if c.Key == key {
			found = true
		}
	}

	if !found {
		s.config.Periods = append(s.config.Periods, &pbl.TaskPeriod{Key: key, Period: int64(freq)})
	}
}

// RegisterRepeatingTaskNonMaster registers a repeating task with a given frequency
func (s *GoServer) RegisterRepeatingTaskNonMaster(task func(ctx context.Context), key string, freq time.Duration) {
	s.servingFuncs = append(s.servingFuncs, sFunc{fun: task, d: freq, nm: true, key: key})
	found := false
	for _, c := range s.config.Periods {
		if c.Key == key {
			found = true
		}
	}

	if !found {
		s.config.Periods = append(s.config.Periods, &pbl.TaskPeriod{Key: key, Period: int64(freq)})
	}

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
	states = append(states, &pbl.State{Key: "running_binary", Text: s.runningFile})
	states = append(states, &pbl.State{Key: "hearts", Value: int64(s.hearts)})
	states = append(states, &pbl.State{Key: "bad_hearts", Value: int64(s.badHearts)})
	states = append(states, &pbl.State{Key: "fail_master", Value: int64(s.failMaster)})
	states = append(states, &pbl.State{Key: "fail_log", Value: int64(s.failLogs)})
	states = append(states, &pbl.State{Key: "fail_message", Text: s.failMessage})
	states = append(states, &pbl.State{Key: "startup_time", TimeValue: s.startup.Unix()})
	s.cpuMutex.Lock()
	states = append(states, &pbl.State{Key: "cpu", Fraction: s.getCPUUsage()})
	s.cpuMutex.Unlock()
	states = append(states, &pbl.State{Key: "periods", Value: int64(len(s.config.Periods))})
	states = append(states, &pbl.State{Key: "alerts_sent", Value: int64(s.AlertsFired)})
	states = append(states, &pbl.State{Key: "alerts_error", Text: s.alertError})
	if s.Sudo {
		p, err := ps.FindProcess(os.Getppid())
		if err == nil {
			states = append(states, &pbl.State{Key: "parent", Value: int64(p.PPid())})
		} else {
			states = append(states, &pbl.State{Key: "parent_error", Text: fmt.Sprintf("%v", err)})
		}
	}
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
		name := fmt.Sprintf("%v-NoD-Repeat", s.Registry.Name)
		ctx, cancel := utils.BuildContext(name, s.Registry.Name, pbl.ContextType_INFINITE)
		defer cancel()
		t.fun(ctx)
		utils.SendTrace(ctx, name, time.Now(), pbt.Milestone_END, s.Registry.Name)
	} else {
		for true {
			if s.Registry.GetMaster() || t.nm {
				name := fmt.Sprintf("%v-Repeat-(%v)-%v", s.Registry.Name, t.key, t.d)
				ctx, cancel := utils.BuildContext(name, s.Registry.Name, pbl.ContextType_LONG)
				defer cancel()
				t.fun(ctx)
				utils.SendTrace(ctx, name, time.Now(), pbt.Milestone_END, s.Registry.Name)
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
	// Set the file details
	ex, _ := os.Executable()
	data, _ := ioutil.ReadFile(ex)
	s.runningFile = fmt.Sprintf("%x", md5.Sum(data))

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

	s.startup = time.Now()

	server.Serve(lis)
	return nil
}

//RaiseIssue raises an issue
func (s *GoServer) RaiseIssue(ctx context.Context, title, body string, sticky bool) {
	s.AlertsFired++
	go func() {
		if !s.SkipLog {
			ip, port, _ := utils.Resolve("githubcard")
			if port > 0 {
				conn, err := grpc.Dial(ip+":"+strconv.Itoa(int(port)), grpc.WithInsecure())
				if err == nil {
					defer conn.Close()
					client := pbgh.NewGithubClient(conn)
					ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
					defer cancel()

					_, err := client.AddIssue(ctx, &pbgh.Issue{Service: s.Servername, Title: title, Body: body, Sticky: sticky}, grpc.FailFast(false))
					if err != nil {
						s.alertError = fmt.Sprintf("Failure to add issue: %v", err)
					}
				}
			} else {
				s.alertError = fmt.Sprintf("Cannot locate githubcard")
			}
		} else {
			s.alertError = "Skip log enabled"
		}
	}()
}

//BounceIssue raises an issue for a different source
func (s *GoServer) BounceIssue(ctx context.Context, title, body string, job string) {
	s.AlertsFired++
	go func() {
		if !s.SkipLog {
			ip, port, _ := utils.Resolve("githubcard")
			if port > 0 {
				conn, err := grpc.Dial(ip+":"+strconv.Itoa(int(port)), grpc.WithInsecure())
				if err == nil {
					defer conn.Close()
					client := pbgh.NewGithubClient(conn)
					ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
					defer cancel()

					_, err := client.AddIssue(ctx, &pbgh.Issue{Service: job, Title: title, Body: body}, grpc.FailFast(false))
					if err != nil {
						s.alertError = fmt.Sprintf("Failure to add issue: %v", err)
					}
				}
			} else {
				s.alertError = fmt.Sprintf("Cannot locate githubcard")
			}
		} else {
			s.alertError = "Skip log enabled"
		}
	}()
}

//LogTrace logs out a trace
func (s *GoServer) LogTrace(c context.Context, l string, t time.Time, ty pbt.Milestone_MilestoneType) context.Context {
	go func() {
		if !s.SkipLog {
			utils.SendTrace(c, l, t, ty, s.Registry.Name)
		}
	}()

	// Add in the context
	md, _ := metadata.FromIncomingContext(c)
	return metadata.NewOutgoingContext(c, md)
}
