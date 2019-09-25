package goserver

import (
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"runtime"
	"runtime/pprof"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/brotherlogic/goserver/utils"
	"github.com/golang/protobuf/proto"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"

	pbbs "github.com/brotherlogic/buildserver/proto"
	pb "github.com/brotherlogic/discovery/proto"
	pbgh "github.com/brotherlogic/githubcard/proto"
	pbgbs "github.com/brotherlogic/gobuildslave/proto"
	pbl "github.com/brotherlogic/goserver/proto"
	pbks "github.com/brotherlogic/keystore/proto"
	pbd "github.com/brotherlogic/monitor/proto"
	pbt "github.com/brotherlogic/tracer/proto"

	ps "github.com/mitchellh/go-ps"

	// This enables pprof
	_ "net/http/pprof"

	//Needed to pull in gzip encoding init
	_ "google.golang.org/grpc/encoding/gzip"
)

type rpcStats struct {
	source    string
	rpcName   string
	count     int64
	errors    int64
	lastError string
	timeIn    time.Duration
	memChange int64
	origin    string
	latencies []time.Duration
}

func (s *GoServer) trace(c context.Context, name string) context.Context {
	go func() {
		s.sendTrace(c, name, time.Now())
	}()

	// Add in the context
	md, _ := metadata.FromIncomingContext(c)
	return metadata.NewOutgoingContext(c, md)
}

func (s *GoServer) mark(c context.Context, t time.Duration) {
	go func() {
		s.sendMark(c, t)
	}()
}

func (s *GoServer) validateMaster() error {
	if s.Registry.Version == pb.RegistryEntry_V2 {
		entry, err := utils.ResolveV2(s.Registry.Name)
		if err != nil {
			return err
		}

		if entry.Identifier != s.Registry.Identifier {
			return fmt.Errorf("We are no longer master")
		}
	} else {
		ip, _, err := utils.Resolve(s.Registry.Name, s.Registry.Name)
		if err != nil {
			return err
		}

		if s.Registry.Ip != ip {
			return fmt.Errorf("We are no longer master, %v is", ip)
		}
	}

	return nil
}

func (s *GoServer) sendTrace(c context.Context, name string, t time.Time) error {
	md, found := metadata.FromIncomingContext(c)
	if found {
		if _, ok := md["trace-id"]; ok {
			idt := md["trace-id"][0]

			if idt != "" {
				id := idt
				if strings.HasPrefix(id, "test") {
					return errors.New("Test trace")
				}

				conn, err := s.DialMaster("tracer")
				if err == nil {
					defer conn.Close()
					ctx, cancel := context.WithTimeout(context.Background(), time.Second)
					defer cancel()
					client := pbt.NewTracerServiceClient(conn)

					// Adjust the time if necessary
					if t.IsZero() {
						t = time.Now()
					}

					m := &pbt.Event{Server: s.Registry.Identifier, Binary: s.Registry.Name, Id: id, Call: name, Timestamp: t.UnixNano()}
					_, err := client.Record(ctx, &pbt.RecordRequest{Event: m})
					return err
				}
			}

			return fmt.Errorf("Unable to trace - maybe because of %v", md)
		}
	}
	s.incoming++
	return fmt.Errorf("Unable to trace - context: %v", c)
}

func (s *GoServer) sendMark(c context.Context, t time.Duration) error {
	md, found := metadata.FromIncomingContext(c)
	if found {
		if _, ok := md["trace-id"]; ok {
			idt := md["trace-id"][0]

			if idt != "" {
				id := idt
				if strings.HasPrefix(id, "test") {
					return errors.New("Test trace")
				}

				conn, err := s.DialMaster("tracer")
				if err == nil {
					defer conn.Close()
					ctx, cancel := context.WithTimeout(context.Background(), time.Second)
					defer cancel()
					client := pbt.NewTracerServiceClient(conn)

					_, err := client.Mark(ctx, &pbt.MarkRequest{LongRunningId: id, RunningTimeInMs: t.Nanoseconds() / 1000000, Origin: s.Registry.Name})
					return err
				}
			}

			return fmt.Errorf("Unable to trace - maybe because of %v", md)
		}
	}
	return fmt.Errorf("Unable to trace - context: %v", c)
}

// DoDial dials a server
func (s *GoServer) DoDial(entry *pb.RegistryEntry) (*grpc.ClientConn, error) {
	return grpc.Dial(entry.Ip+":"+strconv.Itoa(int(entry.Port)), grpc.WithInsecure(), s.withClientUnaryInterceptor(), grpc.WithMaxMsgSize(1024*1024*1024))
}

// DialServer dials a given server
func (s *GoServer) DialServer(server, host string) (*grpc.ClientConn, error) {
	entries, err := utils.ResolveAll(server)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.Identifier == host {
			return s.DoDial(entry)
		}
	}

	return nil, fmt.Errorf("Unable to locate server called %v", server)
}

// DialMaster dials the master server
func (s *GoServer) DialMaster(server string) (*grpc.ClientConn, error) {
	if s.Registry.Version == pb.RegistryEntry_V2 {
		entry, err := utils.ResolveV2(server)
		if err != nil {
			return nil, err
		}
		return s.DoDial(entry)
	}

	ip, port, err := utils.Resolve(server, s.Registry.Name)
	if err != nil {
		return nil, err
	}

	return s.DoDial(&pb.RegistryEntry{Ip: ip, Port: port})
}

func (s *GoServer) withClientUnaryInterceptor() grpc.DialOption {
	return grpc.WithUnaryInterceptor(s.clientInterceptor)
}

func (s *GoServer) clientInterceptor(ctx context.Context,
	method string,
	req interface{},
	reply interface{},
	cc *grpc.ClientConn,
	invoker grpc.UnaryInvoker,
	opts ...grpc.CallOption,
) error {

	s.activeRPCsMutex.Lock()
	s.activeRPCs[method]++
	s.activeRPCsMutex.Unlock()

	var tracer *rpcStats
	if s.RPCTracing {
		for _, trace := range s.traces {
			if trace.rpcName == method && trace.source == "client" {
				tracer = trace
			}
		}

		if tracer == nil {
			tracer = &rpcStats{rpcName: method, count: 0, latencies: make([]time.Duration, 100), source: "client"}
			s.traces = append(s.traces, tracer)
		}
	}

	// Calls the handler
	t := time.Now()

	var err error
	if s.LameDuck {
		err = fmt.Errorf("Server is lameducking")
	} else {
		s.outgoing++
		err = invoker(ctx, method, req, reply, cc, opts...)
		s.outgoing--
	}

	if s.RPCTracing {
		tracer.latencies[tracer.count%100] = time.Now().Sub(t)
		tracer.count++
		tracer.timeIn += time.Now().Sub(t)
		if err != nil {
			tracer.errors++
			tracer.lastError = fmt.Sprintf("%v", err)
		}
	}

	s.activeRPCsMutex.Lock()
	s.activeRPCs[method]--
	s.activeRPCsMutex.Unlock()
	return err
}

func (s *GoServer) serverInterceptor(ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler) (interface{}, error) {

	s.activeRPCsMutex.Lock()
	s.activeRPCs[info.FullMethod]++
	s.activeRPCsMutex.Unlock()

	var tracer *rpcStats
	if s.RPCTracing {
		for _, trace := range s.traces {
			if trace.rpcName == info.FullMethod && trace.source == "server" {
				tracer = trace
			}
		}

		if tracer == nil {
			tracer = &rpcStats{rpcName: info.FullMethod, count: 0, latencies: make([]time.Duration, 100), source: "server"}
			s.traces = append(s.traces, tracer)
		}
	}

	// Calls the handler
	if s.SendTrace {
		ctx = s.trace(ctx, info.FullMethod)
	}
	t := time.Now()
	h, err := handler(ctx, req)

	if s.RPCTracing {
		tracer.latencies[tracer.count%100] = time.Now().Sub(t)
		tracer.count++
		tracer.timeIn += time.Now().Sub(t)

		if err != nil {
			tracer.errors++
			tracer.lastError = fmt.Sprintf("%v", err)
		}

		// Raise an issue on a long call
		if time.Now().Sub(t) > time.Second*5 {
			s.Log(fmt.Sprintf("Long Call %v, %v -> %v", info.FullMethod, time.Now(), t))
			s.marks++
			s.mark(ctx, time.Now().Sub(t))
		}

		if tracer.count > 100 {
			seconds := time.Now().Sub(s.startup).Nanoseconds() / 1000000000
			qps := float64(tracer.count) / float64(seconds)
			if qps > float64(10) {
				peer, found := peer.FromContext(ctx)

				if found {
					s.Log(fmt.Sprintf("High: (%+v), %v", peer.Addr, ctx))
				}
				s.RaiseIssue(ctx, "Over Active Service", fmt.Sprintf("rpc_%v%v is busy -> %v QPS", tracer.source, tracer.rpcName, qps), false)
			}
		}

	}

	if err == nil {
		if !strings.HasSuffix(info.FullMethod, "State") {
			if proto.Size(h.(proto.Message)) > 1024*1024 {
				s.RaiseIssue(ctx, "Large Response", fmt.Sprintf("%v has produced a large response from %v (%vMb)", info.FullMethod, req, proto.Size(h.(proto.Message))/(1024*1024)), false)
			}
		}
	}
	s.activeRPCsMutex.Lock()
	s.activeRPCs[info.FullMethod]--
	s.activeRPCsMutex.Unlock()
	return h, err
}

func (s *GoServer) suicideWatch() {
	for true {
		time.Sleep(s.suicideTime)

		// Commit suicide if our memory usage is high
		_, mem := s.getCPUUsage()
		s.latestMem = int(mem)

		// GC is we're 90% of the memory cap
		if mem > float64(s.MemCap)*0.9 {
			t := time.Now()
			runtime.GC()
			_, mem2 := s.getCPUUsage()
			s.latestMem = int(mem2)
			s.Log(fmt.Sprintf("Running GC with memory %v (took %v) %v -> %v", mem, time.Now().Sub(t), mem, mem2))
			mem = mem2

		}

		if mem > float64(s.MemCap) {
			s.Log(fmt.Sprintf("Memory exceeded, killing ourselves"))
			ctx, cancel := utils.BuildContext("goserver-crash", s.Registry.Name)
			defer cancel()
			s.SendCrash(ctx, "Memory is Too high", pbbs.Crash_MEMORY)
			memProfile, err := os.Create("/home/simon/" + s.Registry.Name + "-heap.prof")
			if err != nil {
				log.Fatal(err)
			}
			if err = pprof.WriteHeapProfile(memProfile); err != nil {
				log.Fatal(err)
			}
			memProfile.Close()
			cpuProfile, err := os.Create("/home/simon/" + s.Registry.Name + "-cpu.prof")
			if err != nil {
				log.Fatal(err)
			}
			if err := pprof.StartCPUProfile(cpuProfile); err != nil {
				log.Fatal("could not start CPU profile: ", err)
			}
			defer pprof.StopCPUProfile()
			time.Sleep(time.Second * 5)
			os.Exit(1)
		}

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
func (s *GoServer) RegisterServer(servername string, external bool) error {
	s.Servername = servername

	// Short circuit if we don't need to register
	if s.noRegister {
		IP := getLocalIP()
		hostname, err := osHostGetter{}.Hostname()
		if err != nil {
			hostname = "Server-" + IP
		}
		entry := &pb.RegistryEntry{Ip: IP, Name: servername, ExternalPort: false, Identifier: hostname, Port: s.Port}
		s.Registry = entry

		return nil
	}

	err := fmt.Errorf("First fail")
	port := int32(0)
	for err != nil {
		s.registerAttempts++
		port, err = s.getRegisteredServerPort(getLocalIP(), s.Servername, external, false)
		s.Port = port
	}
	return err
}

// RegisterServerV2 registers this server under the v2 protocol
func (s *GoServer) RegisterServerV2(servername string, external bool) error {
	s.Servername = servername

	// Short circuit if we don't need to register
	if s.noRegister {
		IP := getLocalIP()
		hostname, err := osHostGetter{}.Hostname()
		if err != nil {
			hostname = "Server-" + IP
		}
		entry := &pb.RegistryEntry{Ip: IP, Name: servername, ExternalPort: false, Identifier: hostname, Port: s.Port, Version: pb.RegistryEntry_V2}
		s.Registry = entry

		return nil
	}

	err := fmt.Errorf("First fail")
	port := int32(0)
	for err != nil {
		s.registerAttempts++
		port, err = s.getRegisteredServerPort(getLocalIP(), s.Servername, external, true)
		s.Port = port
	}
	return err
}

func (s *GoServer) close(conn *grpc.ClientConn) {
	if conn != nil {
		conn.Close()
	}
}

// RegisterServingTask registers tasks to run when serving
func (s *GoServer) RegisterServingTask(task func(ctx context.Context) error, key string) {
	s.servingFuncs = append(s.servingFuncs, sFunc{fun: task, d: 0, key: key, source: "repeat"})
}

// RegisterRepeatingTask registers a repeating task with a given frequency
func (s *GoServer) RegisterRepeatingTask(task func(ctx context.Context) error, key string, freq time.Duration) {
	s.servingFuncs = append(s.servingFuncs, sFunc{fun: task, d: freq, key: key, source: "repeat"})
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

// RegisterRepeatingTaskNoTrace registers a repeating task with a given frequency
func (s *GoServer) RegisterRepeatingTaskNoTrace(task func(ctx context.Context) error, key string, freq time.Duration) {
	s.servingFuncs = append(s.servingFuncs, sFunc{fun: task, d: freq, key: key, noTrace: true, source: "repeat"})
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
func (s *GoServer) RegisterRepeatingTaskNonMaster(task func(ctx context.Context) error, key string, freq time.Duration) {
	s.servingFuncs = append(s.servingFuncs, sFunc{fun: task, d: freq, nm: true, key: key, source: "repeat"})
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
	s.activeRPCsMutex.Lock()
	defer s.activeRPCsMutex.Unlock()
	states = append(states, &pbl.State{Key: "alert_wait", TimeValue: s.alertWait.Unix()})
	states = append(states, &pbl.State{Key: "active_rpcs", Text: fmt.Sprintf("%v", s.activeRPCs)})
	states = append(states, &pbl.State{Key: "memory", Text: fmt.Sprintf("%v/%v", s.latestMem, s.MemCap)})
	states = append(states, &pbl.State{Key: "register_attempts", Value: s.registerAttempts})
	states = append(states, &pbl.State{Key: "incoming_counts", Value: s.incoming})
	states = append(states, &pbl.State{Key: "outgoing_counts", Value: s.outgoing})
	states = append(states, &pbl.State{Key: "marks_sent", Value: s.marks})
	states = append(states, &pbl.State{Key: "running_binary", Text: s.RunningFile})
	states = append(states, &pbl.State{Key: "hearts", Value: int64(s.hearts)})
	states = append(states, &pbl.State{Key: "bad_hearts", Value: int64(s.BadHearts)})
	states = append(states, &pbl.State{Key: "bad_heart_message", Text: s.badHeartMessage})
	states = append(states, &pbl.State{Key: "master_requests", Value: int64(s.masterRequests)})
	states = append(states, &pbl.State{Key: "master_requests_fails", Value: int64(s.masterRequestFails)})
	states = append(states, &pbl.State{Key: "master_requests_fail_reason", Text: s.masterRequestFailReason})
	states = append(states, &pbl.State{Key: "fail_master", Value: int64(s.failMaster)})
	states = append(states, &pbl.State{Key: "fail_log", Value: int64(s.failLogs)})
	states = append(states, &pbl.State{Key: "fail_message", Text: s.failMessage})
	states = append(states, &pbl.State{Key: "startup_time", TimeValue: s.startup.Unix()})
	states = append(states, &pbl.State{Key: "uptime", TimeDuration: time.Now().Sub(s.startup).Nanoseconds()})
	states = append(states, &pbl.State{Key: "periods", Value: int64(len(s.config.Periods))})
	states = append(states, &pbl.State{Key: "alerts_sent", Value: int64(s.AlertsFired)})
	states = append(states, &pbl.State{Key: "alerts_error", Text: s.alertError})
	states = append(states, &pbl.State{Key: "alerts_skipped", Value: s.AlertsSkipped})
	states = append(states, &pbl.State{Key: "mote_count", Value: int64(s.moteCount)})
	states = append(states, &pbl.State{Key: "last_mote_time", Text: fmt.Sprintf("%v", s.lastMoteTime)})
	states = append(states, &pbl.State{Key: "last_mote_fail", Text: s.lastMoteFail})

	s.runTimesMutex.Lock()
	for key, ti := range s.runTimes {
		states = append(states, &pbl.State{Key: "last_run_" + key, TimeValue: ti.Unix()})
	}
	s.runTimesMutex.Unlock()

	if s.Sudo {
		p, err := ps.FindProcess(os.Getppid())
		if err == nil {
			states = append(states, &pbl.State{Key: "parent", Value: int64(p.PPid())})
		} else {
			states = append(states, &pbl.State{Key: "parent_error", Text: fmt.Sprintf("%v", err)})
		}
	}

	states = append(states, &pbl.State{Key: "bad_ports", Value: s.badPorts})
	states = append(states, &pbl.State{Key: "reg_time", TimeDuration: s.regTime.Nanoseconds()})

	for _, trace := range s.traces {

		states = append(states, &pbl.State{Key: "rpc_" + trace.source + trace.rpcName + "_count", Value: trace.count})
		states = append(states, &pbl.State{Key: "rpc_" + trace.source + trace.rpcName + "_errors", Value: trace.errors})
		states = append(states, &pbl.State{Key: "rpc_" + trace.source + trace.rpcName + "_lasterror", Text: trace.lastError})
		states = append(states, &pbl.State{Key: "rpc_" + trace.source + trace.rpcName + "_mem", Value: trace.memChange})
		if trace.count > 0 {
			states = append(states, &pbl.State{Key: "rpc_" + trace.source + trace.rpcName + "_avgTime", TimeDuration: trace.timeIn.Nanoseconds() / trace.count})
		}

		arrCopy := []time.Duration{}
		ind := 0
		for i, v := range trace.latencies {
			if v > 0 {
				arrCopy = append(arrCopy, v)
				ind = i
			}
		}

		sort.SliceStable(arrCopy, func(i, j int) bool {
			return arrCopy[i] < arrCopy[j]
		})
		if ind > 0 {
			states = append(states, &pbl.State{Key: "rpc_" + trace.source + trace.rpcName + "_maxTime", TimeDuration: arrCopy[len(arrCopy)-1].Nanoseconds()})
			states = append(states, &pbl.State{Key: "rpc_" + trace.source + trace.rpcName + "_minTime", TimeDuration: arrCopy[0].Nanoseconds()})
		}
	}

	return &pbl.ServerState{States: states}, nil
}

// Shutdown brings the server down
func (s *GoServer) Shutdown(ctx context.Context, in *pbl.ShutdownRequest) (*pbl.ShutdownResponse, error) {
	s.LameDuck = true
	go func() {
		time.Sleep(time.Minute)
		err := s.Register.Shutdown(ctx)
		if err != nil {
			s.Log(fmt.Sprintf("Shutdown cancelled: %v", err))
		} else {
			os.Exit(1)
		}
	}()
	return &pbl.ShutdownResponse{}, nil
}

// Mote promotes or demotes a server into production
func (s *GoServer) Mote(ctx context.Context, in *pbl.MoteRequest) (*pbl.Empty, error) {
	st := time.Now()
	s.moteCount++

	// We can't mote to master if we're lame ducking
	err := s.Register.Mote(ctx, in.Master && !s.LameDuck)

	// If we were able to mote then we should inform discovery
	if err == nil {
		s.Registry.Master = in.Master
		s.reregister(s.dialler, s.clientBuilder)
	}

	s.lastMoteTime = time.Now().Sub(st)
	s.lastMoteFail = fmt.Sprintf("%v", err)
	return &pbl.Empty{}, err
}

func (s *GoServer) getRegisteredServerPort(IP string, servername string, external bool, v2 bool) (int32, error) {
	return s.registerServer(IP, servername, external, v2, grpcDialler{}, mainBuilder{}, osHostGetter{})
}

//Save a protobuf
func (s *GoServer) Save(ctx context.Context, key string, p proto.Message) error {
	return s.KSclient.Save(ctx, key, p)
}

//RunBackgroundTask with tracing and tracking
func (s *GoServer) RunBackgroundTask(task func(ctx context.Context) error, name string) {
	go s.run(sFunc{
		fun:     task,
		key:     name,
		source:  "background",
		runOnce: true,
		nm:      true,
	})
}

func (s *GoServer) run(t sFunc) {
	time.Sleep(time.Minute)
	if t.d == 0 && !t.runOnce {
		t.fun(context.Background())
	} else {
		for true {
			err := s.validateMaster()
			if err == nil || t.nm {
				name := fmt.Sprintf("%v-Repeat-(%v)-%v", s.Registry.Name, t.key, t.d)
				var ctx context.Context
				var cancel context.CancelFunc
				if t.noTrace {
					ctx, cancel = context.WithTimeout(context.Background(), time.Hour)
				} else {
					ctx, cancel = utils.BuildContext(name, name)
				}
				defer cancel()
				s.runTimesMutex.Lock()
				s.runTimes[t.key] = time.Now()
				s.runTimesMutex.Unlock()

				var tracer *rpcStats
				if s.RPCTracing {
					for _, trace := range s.traces {
						if trace.rpcName == "/"+t.key && trace.source == t.source {
							tracer = trace
						}
					}

					if tracer == nil {
						tracer = &rpcStats{rpcName: "/" + t.key, count: 0, latencies: make([]time.Duration, 100), source: t.source}
						s.traces = append(s.traces, tracer)
					}
				}

				ti := time.Now()
				err := t.fun(ctx)
				if s.RPCTracing {
					tracer.latencies[tracer.count%100] = time.Now().Sub(ti)
					tracer.count++
					tracer.timeIn += time.Now().Sub(ti)
					if err != nil {
						tracer.errors++
						tracer.lastError = fmt.Sprintf("%v", err)

						if float64(tracer.errors)/float64(tracer.count) > 0.5 && tracer.count > 10 {
							s.RaiseIssue(ctx, "Failing Task", fmt.Sprintf("%v is failing at %v [%v]", tracer.rpcName, float64(tracer.errors)/float64(tracer.count), tracer.lastError), false)
						}
					}
				}
			}
			time.Sleep(t.d)
			if t.runOnce {
				return
			}
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
	s.Log(fmt.Sprintf("Starting %v on port %v", s.RunningFile, s.Registry.Port))

	lis, err := net.Listen("tcp", ":"+strconv.Itoa(int(s.Port)))
	if err != nil {
		return err
	}
	server := grpc.NewServer(
		grpc.RPCCompressor(grpc.NewGZIPCompressor()),
		grpc.RPCDecompressor(grpc.NewGZIPDecompressor()),
		grpc.MaxRecvMsgSize(1024*1024*1024),
		grpc.MaxSendMsgSize(1024*1024*1024),
		grpc.UnaryInterceptor(s.serverInterceptor),
	)
	s.Register.DoRegister(server)
	pbl.RegisterGoserverServiceServer(server, s)

	if !s.noRegister {
		if s.Registry.Version == pb.RegistryEntry_V1 {
			s.setupHeartbeats()
		}
		go s.suicideWatch()
	}

	// Enable profiling
	go http.ListenAndServe(fmt.Sprintf(":%v", s.Port+1), nil)

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

	if time.Now().Before(s.alertWait) {
		s.AlertsSkipped++
		return
	}

	go func() {
		if !s.SkipLog || len(body) == 0 {
			ip, port, _ := utils.Resolve("githubcard", s.Registry.Name+"-ri")
			if port > 0 {
				conn, err := grpc.Dial(ip+":"+strconv.Itoa(int(port)), grpc.WithInsecure())
				if err == nil {
					defer conn.Close()
					client := pbgh.NewGithubClient(conn)
					ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
					defer cancel()

					_, err := client.AddIssue(ctx, &pbgh.Issue{Service: s.Servername, Title: title, Body: body, Sticky: sticky}, grpc.FailFast(false))
					if err != nil {
						st := status.Convert(err)
						if st.Code() == codes.ResourceExhausted {
							s.alertWait = time.Now().Add(time.Minute * 10)
						} else {
							s.alertError = fmt.Sprintf("Failure to add issue: %v", err)
						}
					}
				}
			} else {
				s.alertWait = time.Now().Add(time.Minute * 10)
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
			ip, port, _ := utils.Resolve("githubcard", s.Registry.Name+"-bi")
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

//SendCrash reports a crash
func (s *GoServer) SendCrash(ctx context.Context, crashText string, ctype pbbs.Crash_CrashType) {
	conn, err := s.DialMaster("buildserver")
	if err == nil {
		defer conn.Close()
		client := pbbs.NewBuildServiceClient(conn)

		// Build out the current state
		info, _ := s.State(ctx, &pbl.Empty{})
		infoString := ""
		for _, str := range info.GetStates() {
			infoString += fmt.Sprintf("%v = %v\n", str.Key, str)
		}

		client.ReportCrash(ctx, &pbbs.CrashRequest{
			Version: s.RunningFile,
			Origin:  s.Registry.Name,
			Job: &pbgbs.Job{
				Name: s.Registry.Name,
			},
			Crash: &pbbs.Crash{
				ErrorMessage: crashText + "\n" + infoString,
				CrashType:    ctype}})
	}
}

//PLog a simple string message with priority
func (s *GoServer) PLog(message string, level pbd.LogLevel) {
	go func() {
		if !s.SkipLog {
			conn, err := s.DialMaster("monitor")
			if err == nil {
				defer conn.Close()
				monitor := s.monitorBuilder.NewMonitorServiceClient(conn)
				messageLog := &pbd.MessageLog{Message: message, Entry: s.Registry, Level: level}
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()
				_, err := monitor.WriteMessageLog(ctx, messageLog, grpc.FailFast(false))
				e, ok := status.FromError(err)
				if ok && err != nil && e.Code() != codes.DeadlineExceeded {
					s.failLogs++
					s.failMessage = fmt.Sprintf("%v", message)
				}
				s.close(conn)
			}
		}
	}()
}

// RegisterServer Registers a server with the system and gets the port number it should use
func (s *GoServer) registerServer(IP string, servername string, external bool, v2 bool, dialler dialler, builder clientBuilder, getter hostGetter) (int32, error) {
	conn, err := dialler.Dial(utils.RegistryIP+":"+strconv.Itoa(utils.RegistryPort), grpc.WithInsecure())
	if err != nil {
		return -1, err
	}

	if v2 {
		registry := pb.NewDiscoveryServiceV2Client(conn)
		hostname, err := getter.Hostname()
		if err != nil {
			hostname = "Server-" + IP
		}
		entry := pb.RegistryEntry{Ip: IP, Name: servername, ExternalPort: external, Identifier: hostname, TimeToClean: 5000}
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()
		t := time.Now()
		r, err := registry.RegisterV2(ctx, &pb.RegisterRequest{Service: &entry}, grpc.FailFast(false))
		s.regTime = time.Now().Sub(t)
		if err != nil {
			s.close(conn)
			return -1, err
		}
		s.Registry = r.GetService()
		s.close(conn)

		return r.GetService().Port, nil

	}

	registry := builder.NewDiscoveryServiceClient(conn)
	hostname, err := getter.Hostname()
	if err != nil {
		hostname = "Server-" + IP
	}
	entry := pb.RegistryEntry{Ip: IP, Name: servername, ExternalPort: external, Identifier: hostname, TimeToClean: 5000}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	t := time.Now()
	r, err := registry.RegisterService(ctx, &pb.RegisterRequest{Service: &entry}, grpc.FailFast(false))
	s.regTime = time.Now().Sub(t)
	if err != nil {
		s.close(conn)
		return -1, err
	}
	s.Registry = r.GetService()
	s.close(conn)

	return r.GetService().Port, nil
}
