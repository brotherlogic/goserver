package goserver

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"runtime"
	"runtime/debug"
	"runtime/pprof"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/brotherlogic/goserver/utils"
	"github.com/golang/protobuf/proto"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/resolver"
	"google.golang.org/grpc/status"

	pbbs "github.com/brotherlogic/buildserver/proto"
	pb "github.com/brotherlogic/discovery/proto"
	pbgh "github.com/brotherlogic/githubcard/proto"
	pbgbs "github.com/brotherlogic/gobuildslave/proto"
	pbl "github.com/brotherlogic/goserver/proto"
	pbks "github.com/brotherlogic/keystore/proto"
	lpb "github.com/brotherlogic/logging/proto"
	pbd "github.com/brotherlogic/monitor/proto"
	pbt "github.com/brotherlogic/tracer/proto"
	pbv "github.com/brotherlogic/versionserver/proto"

	ps "github.com/mitchellh/go-ps"

	// This enables pprof
	_ "net/http/pprof"
)

type rpcStats struct {
	source      string
	rpcName     string
	count       int64
	errors      int64
	nferrors    int64
	lastNFError string
	lastError   string
	timeIn      time.Duration
	memChange   int64
	origin      string
	latencies   []time.Duration
}

var (
	serverRequests = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "rpc_server_requests",
		Help: "The number of server requests",
	}, []string{"method"})
	clientRequests = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "rpc_client_requests",
		Help: "The number of server requests",
	}, []string{"method"})
	repeatRequests = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "rpc_repeat_requests",
		Help: "The number of server requests",
	}, []string{"method"})
	repeatLatency = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name: "rpc_repeat_latency",
		Help: "The latency of repeat requests",
	}, []string{"method"})
)

func init() {
	resolver.Register(&utils.DiscoveryServerResolverBuilder{})
}
func (s *GoServer) trace(c context.Context, name string) context.Context {
	go func() {
		s.sendTrace(c, name, time.Now())
	}()

	// Add in the context
	md, _ := metadata.FromIncomingContext(c)
	return metadata.NewOutgoingContext(c, md)
}

func (s *GoServer) mark(c context.Context, t time.Duration, m string) {
	go func() {
		s.sendMark(c, t, m)
	}()
}

func (s *GoServer) alive(ctx context.Context, entry *pb.RegistryEntry) error {
	conn, err := s.DoDial(entry)
	if err != nil {
		return err
	}
	defer conn.Close()
	client := pbl.NewGoserverServiceClient(conn)
	_, err = client.IsAlive(ctx, &pbl.Alive{})
	return err
}

func (s *GoServer) validateMaster(ctx context.Context) error {
	//Master ignores don't validate
	if s.Registry.IgnoresMaster {
		return nil
	}

	if s.Registry.Version == pb.RegistryEntry_V2 {
		//Acquire a validation lock
		s.masterMutex.Lock()
		defer s.masterMutex.Unlock()

		entry, err := utils.ResolveV2(s.Registry.Name)
		if err == nil {
			err = s.alive(ctx, entry)
		}
		if err != nil {
			//Let's master elect if we can't find a master
			if code := status.Convert(err); code.Code() == codes.NotFound || code.Code() == codes.Unavailable {
				err := s.masterElect(ctx)
				return err
			}

			return err
		}

		if entry.Identifier != s.Registry.Identifier {
			return status.Errorf(codes.FailedPrecondition, "We are no longer master but %v is", entry)
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

func (s *GoServer) masterElect(ctx context.Context) error {
	if s.Registry.Version == pb.RegistryEntry_V1 {
		return fmt.Errorf("V1 does not perform master election")
	}

	conn, err := s.DoDial(s.Registry)
	if err != nil {
		return err
	}
	defer conn.Close()
	client := pbl.NewGoserverServiceClient(conn)
	_, err = client.Mote(ctx, &pbl.MoteRequest{Master: true})
	if err != nil {
		return err
	}

	conn2, err := s.DialLocal("discover")
	if err != nil {
		return err
	}
	defer conn2.Close()

	client2 := pb.NewDiscoveryServiceV2Client(conn2)
	resp, err := client2.MasterElect(ctx, &pb.MasterRequest{Service: s.Registry, MasterElect: true})
	s.Log(fmt.Sprintf("Master elect response: %v, %v", resp, err))
	if err == nil {
		s.Registry = resp.GetService()
	}
	return err
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

func (s *GoServer) sendMark(c context.Context, t time.Duration, message string) error {
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

					_, err := client.Mark(ctx, &pbt.MarkRequest{LongRunningId: id, RunningTimeInMs: t.Nanoseconds() / 1000000, Origin: s.Registry.Name, RequestMessage: message})
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
	return s.BaseDial(entry.Ip + ":" + strconv.Itoa(int(entry.Port)))
}

// BaseDial dials a connection
func (s *GoServer) BaseDial(c string) (*grpc.ClientConn, error) {
	return grpc.Dial(c, grpc.WithInsecure(), s.withClientUnaryInterceptor(), grpc.WithMaxMsgSize(1024*1024*1024))
}

// NewBaseDial dials a connection
func (s *GoServer) NewBaseDial(c string) (*grpc.ClientConn, error) {
	return grpc.Dial("discovery:///"+c,
		grpc.WithInsecure(),
		s.withClientUnaryInterceptor(),
		grpc.WithBalancerName("my_pick_first"))
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

//DialLocal dials through the local discover
func (s *GoServer) DialLocal(server string) (*grpc.ClientConn, error) {
	entry, err := utils.ResolveV2(server)
	if err != nil {
		return nil, err
	}
	return s.DoDial(entry)
}

// DialMaster dials the master server
func (s *GoServer) DialMaster(server string) (*grpc.ClientConn, error) {
	if s.Registry == nil || s.Registry.Version == pb.RegistryEntry_V2 {
		entries, err := utils.ResolveV3(server)
		if err != nil {
			return nil, err
		}
		if len(entries) == 0 {
			return nil, fmt.Errorf("Could not find any servers for %v", server)
		}
		return s.DoDial(entries[0])
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

	clientRequests.With(prometheus.Labels{"method": method}).Inc()

	var tracer *rpcStats
	if s.RPCTracing {
		tracer = s.getTrace(method, "client")
	}

	// Calls the handler
	t := time.Now()

	var err error
	s.outgoing++
	err = invoker(ctx, method, req, reply, cc, opts...)
	retries := 1
	for retries < 3 && err != nil {
		err = invoker(ctx, method, req, reply, cc, opts...)
		retries++
	}
	s.outgoing--

	if s.RPCTracing {
		s.recordTrace(ctx, tracer, method, time.Now().Sub(t), err, req, false)
	}

	s.activeRPCsMutex.Lock()
	s.activeRPCs[method]--
	s.activeRPCsMutex.Unlock()
	return err
}

func (s *GoServer) getTrace(name, source string) *rpcStats {
	for _, trace := range s.traces {
		if trace != nil {
			if trace.rpcName == name && source == source {
				return trace
			}
		}
	}

	tracer := &rpcStats{rpcName: name, count: 0, latencies: make([]time.Duration, 100), source: source}
	s.traces = append(s.traces, tracer)
	return tracer
}

func (s *GoServer) recordTrace(ctx context.Context, tracer *rpcStats, name string, timeTaken time.Duration, err error, req interface{}, mark bool) {
	tracer.latencies[tracer.count%100] = timeTaken
	tracer.count++
	tracer.timeIn += timeTaken
	tracer.origin = fmt.Sprintf("%+v", ctx)

	if err != nil {
		code := status.Convert(err)
		if code.Code() == codes.Unknown || code.Code() == codes.Internal {
			tracer.errors++
			tracer.lastError = fmt.Sprintf("%v", err)

			if float64(tracer.errors)/float64(tracer.count) > 0.8 && tracer.count > 10 {
				s.RaiseIssue(ctx, fmt.Sprintf("Error for %v", name), fmt.Sprintf("%v [%v]: %v calls %v errors (%v)", s.Registry.Identifier, s.RunningFile, tracer.count, tracer.errors, err), false)
			}
		} else {
			tracer.nferrors++
			tracer.lastNFError = fmt.Sprintf("%v", err)
		}
	}

	// Raise an issue on a long call
	if timeTaken > time.Second*5 && mark {
		s.marks++
		s.mark(ctx, timeTaken, fmt.Sprintf("%v/%v: %v -> %v", s.Registry.Name, s.Registry.Identifier, name, req))
	}

	if tracer.count > 100 {
		seconds := time.Now().Sub(s.startup).Nanoseconds() / 1000000000
		qps := float64(tracer.count) / float64(seconds)
		if tracer.timeIn/time.Now().Sub(s.startup) > time.Second {
			peer, found := peer.FromContext(ctx)

			if found {
				s.Log(fmt.Sprintf("High: (%+v), %v", peer.Addr, ctx))
			}
			s.RaiseIssue(ctx, "Over Active Service", fmt.Sprintf("rpc_%v%v is busy -> %v QPS / %v QTie", tracer.source, tracer.rpcName, qps, tracer.timeIn/time.Now().Sub(s.startup)), false)
		}
	}
}

func (s *GoServer) serverInterceptor(ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler) (interface{}, error) {

	s.activeRPCsMutex.Lock()
	s.activeRPCs[info.FullMethod]++
	s.activeRPCsMutex.Unlock()

	serverRequests.With(prometheus.Labels{"method": info.FullMethod}).Inc()

	var tracer *rpcStats
	if s.RPCTracing {
		tracer = s.getTrace(info.FullMethod, "server")
	}

	// Calls the handler
	if s.SendTrace {
		ctx = s.trace(ctx, info.FullMethod)
	}
	h, err := s.runHandle(ctx, handler, req, tracer, info.FullMethod)

	if err == nil && h != nil {
		if proto.Size(h.(proto.Message)) > 1024*1024 {
			s.RaiseIssue(ctx, "Large Response", fmt.Sprintf("%v has produced a large response from %v (%vMb) -> %v", info.FullMethod, req, proto.Size(h.(proto.Message))/(1024*1024), ctx), false)
		}
	}
	s.activeRPCsMutex.Lock()
	s.activeRPCs[info.FullMethod]--
	s.activeRPCsMutex.Unlock()
	return h, err
}

func (s *GoServer) runHandle(ctx context.Context, handler grpc.UnaryHandler, req interface{}, tracer *rpcStats, name string) (resp interface{}, err error) {
	ti := time.Now()

	// Immediate return without trace if we're not master and we expect to be so
	// Or if we expect to be able to master, try electing to be master
	if !s.Registry.IgnoresMaster && !s.Registry.Master && !strings.HasPrefix(name, "/goserver") {
		if s.Registry.Version == pb.RegistryEntry_V1 {
			err = fmt.Errorf("Cannot handle %v - we are not master", name)
		} else if s.Registry.Version == pb.RegistryEntry_V2 {
			s.masterv++
			err = s.validateMaster(ctx)
			if err != nil {
				s.mastervfail++
			}
		}
		if err != nil {
			return nil, err
		}
	}

	defer func() {
		if s.RPCTracing {
			if r := recover(); r != nil {
				err = fmt.Errorf("%v", r)
				s.Log(fmt.Sprintf("Crashed: %v", string(debug.Stack())))
				s.SendCrash(ctx, fmt.Sprintf("%v", string(debug.Stack())), pbbs.Crash_PANIC)
				s.recordTrace(ctx, tracer, name, time.Now().Sub(ti), err, "", true)
			} else {
				s.recordTrace(ctx, tracer, name, time.Now().Sub(ti), err, "", true)
			}
		}
	}()
	resp, err = handler(ctx, req)
	return
}

// HTTPGet gets an http resource
func (s *GoServer) HTTPGet(ctx context.Context, url string, useragent string) (string, error) {

	var tracer *rpcStats
	if s.RPCTracing {
		tracer = s.getTrace("http_get", "client")
	}

	t := time.Now()

	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", useragent)
	var response http.Response
	body := []byte{}
	if err == nil {
		response, err := client.Do(req)
		if err == nil {
			defer response.Body.Close()
			body, _ = ioutil.ReadAll(response.Body)
		}
	}

	if response.StatusCode != 200 && response.StatusCode != 201 && response.StatusCode != 204 && response.StatusCode != 0 {
		err = fmt.Errorf("Error reading url: %v and %v", response.StatusCode, string(body))
	}

	if s.RPCTracing {
		s.recordTrace(ctx, tracer, "http_get", time.Now().Sub(t), err, url, false)
	}

	return string(body), err
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
	return s.RegisterServerIgnore(servername, external, false)
}

// RegisterServerIgnore registers this server with ignore master set.
func (s *GoServer) RegisterServerIgnore(servername string, external bool, ignore bool) error {
	s.Servername = servername

	// Short circuit if we don't need to register
	if s.noRegister {
		IP := getLocalIP()
		hostname, err := osHostGetter{}.Hostname()
		if err != nil {
			hostname = "Server-" + IP
		}
		entry := &pb.RegistryEntry{Ip: IP, Name: servername, ExternalPort: false, Identifier: hostname, Port: s.Port, IgnoresMaster: ignore}
		s.Registry = entry

		return nil
	}

	err := fmt.Errorf("First fail")
	port := int32(0)
	start := s.registerAttempts
	for err != nil && s.registerAttempts-start < 10 {
		s.registerAttempts++
		port, err = s.getRegisteredServerPort(getLocalIP(), s.Servername, external, false, ignore)
		s.Port = port
	}
	return err
}

// RegisterServerV2 registers this server under the v2 protocol
func (s *GoServer) RegisterServerV2(servername string, external bool, ignore bool) error {
	s.Servername = servername

	// Short circuit if we don't need to register
	if s.noRegister {
		IP := getLocalIP()
		hostname, err := osHostGetter{}.Hostname()
		if err != nil {
			hostname = "Server-" + IP
		}
		entry := &pb.RegistryEntry{Ip: IP, Name: servername, ExternalPort: false, Identifier: hostname, Port: s.Port, Version: pb.RegistryEntry_V2, IgnoresMaster: ignore}
		s.Registry = entry

		return nil
	}

	err := fmt.Errorf("First fail")
	port := int32(0)
	start := s.registerAttempts
	for err != nil && s.registerAttempts-start < 10 {
		s.registerAttempts++
		port, err = s.getRegisteredServerPort(getLocalIP(), s.Servername, external, true, ignore)
		s.Port = port
	}
	return err
}

func (s *GoServer) close(conn *grpc.ClientConn) {
	if conn != nil {
		conn.Close()
	}
}

// RegisterLockingTask registers a locking task to run
func (s *GoServer) RegisterLockingTask(task func(ctx context.Context) (time.Time, error), key string) {
	s.servingFuncs = append(s.servingFuncs, sFunc{lFun: task, key: key, source: "locking"})
}

// RegisterServingTask registers tasks to run when serving
func (s *GoServer) RegisterServingTask(task func(ctx context.Context) error, key string) {
	s.servingFuncs = append(s.servingFuncs, sFunc{fun: task, d: 0, key: key, source: "repeat"})
}

// RegisterRepeatingTask registers a repeating task with a given frequency
func (s *GoServer) RegisterRepeatingTask(task func(ctx context.Context) error, key string, freq time.Duration) {
	if freq < time.Second*5 {
		log.Fatalf("%v is too short for %v", freq, key)
	}
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
	if freq < time.Second*5 {
		log.Fatalf("%v is too short for %v", freq, key)
	}

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
	nilTraces := int64(0)
	for _, t := range s.traces {
		if t == nil {
			nilTraces++
		}
	}
	states = append(states, &pbl.State{Key: "masterv", Value: s.masterv})
	states = append(states, &pbl.State{Key: "masterv_fail", Value: s.mastervfail})
	states = append(states, &pbl.State{Key: "reg", Text: fmt.Sprintf("%v", s.Registry)})
	states = append(states, &pbl.State{Key: "nil_traces", Value: nilTraces})
	states = append(states, &pbl.State{Key: "alert_wait", TimeValue: s.alertWait.Unix()})
	states = append(states, &pbl.State{Key: "active_rpcs", Text: fmt.Sprintf("%v", s.activeRPCs)})
	states = append(states, &pbl.State{Key: "memory", Text: fmt.Sprintf("%v/%v", s.latestMem, s.MemCap)})
	states = append(states, &pbl.State{Key: "register_attempts", Value: s.registerAttempts})
	states = append(states, &pbl.State{Key: "incoming_counts", Value: s.incoming})
	states = append(states, &pbl.State{Key: "outgoing_counts", Value: s.outgoing})
	states = append(states, &pbl.State{Key: "marks_sent", Value: s.marks})
	states = append(states, &pbl.State{Key: "running_binary", Text: s.RunningFile})
	states = append(states, &pbl.State{Key: "running_binary_date", Text: s.RunningFileDate})
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
		states = append(states, &pbl.State{Key: "rpc_" + trace.source + trace.rpcName + "_nferrors", Value: trace.nferrors})
		states = append(states, &pbl.State{Key: "rpc_" + trace.source + trace.rpcName + "_lastnferror", Text: trace.lastNFError})
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

// Reregister this server
func (s *GoServer) Reregister(ctx context.Context, in *pbl.ReregisterRequest) (*pbl.ReregisterResponse, error) {
	err := s.RegisterServerV2(s.Registry.Name, s.Registry.ExternalPort, s.Registry.IgnoresMaster)
	return &pbl.ReregisterResponse{}, err
}

// Shutdown brings the server down
func (s *GoServer) Shutdown(ctx context.Context, in *pbl.ShutdownRequest) (*pbl.ShutdownResponse, error) {
	s.LameDuck = true
	fmt.Printf("Shutting down %v, %v", in, ctx)
	go func() {
		time.Sleep(time.Minute)
		err := s.Register.Shutdown(ctx)
		if err != nil {
			s.Log(fmt.Sprintf("Shutdown cancelled: %v", err))
			return
		}

		// Unregister us from discovery
		conn, err := s.DialLocal("discover")
		if err != nil {
			s.Log(fmt.Sprintf("Unable to shutdown: %v", err))
			return
		}
		defer conn.Close()

		registry := pb.NewDiscoveryServiceV2Client(conn)
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		_, err = registry.Unregister(ctx, &pb.UnregisterRequest{Service: s.Registry})
		if err == nil {
			os.Exit(1)
		}

		s.Log(fmt.Sprintf("Cannot shutdown: %v", err))
	}()
	return &pbl.ShutdownResponse{}, nil
}

// Mote promotes or demotes a server into production
func (s *GoServer) Mote(ctx context.Context, in *pbl.MoteRequest) (*pbl.Empty, error) {
	st := time.Now()
	s.moteCount++

	// We can't mote to master if we're lame ducking
	err := s.Register.Mote(ctx, in.Master && !s.LameDuck)

	// If we were able to mote then we should inform discovery if we're running in V1
	if err == nil && s.Registry.Version == pb.RegistryEntry_V1 {
		s.Registry.Master = in.Master
		s.reregister(s.dialler, s.clientBuilder)
	}

	if err == nil {
		s.Registry.Master = in.Master
	}

	s.lastMoteTime = time.Now().Sub(st)
	s.lastMoteFail = fmt.Sprintf("%v", err)
	return &pbl.Empty{}, err
}

func (s *GoServer) getRegisteredServerPort(IP string, servername string, external bool, v2 bool, im bool) (int32, error) {
	return s.registerServer(IP, servername, external, v2, im, grpcDialler{}, mainBuilder{}, osHostGetter{})
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

// Acquires a distributed lock for an hour
func (s *GoServer) acquireLock(lockName string) (time.Time, bool, error) {
	conn, err := s.DialMaster("versionserver")
	if err != nil {
		return time.Now(), false, err
	}
	defer conn.Close()

	client := pbv.NewVersionServerClient(conn)
	ctx, cancel := utils.BuildContext(lockName, lockName)
	defer cancel()
	resp, err := client.SetIfLessThan(ctx, &pbv.SetIfLessThanRequest{
		TriggerValue: time.Now().Unix(),
		Set: &pbv.Version{
			Key:    lockName,
			Value:  time.Now().Add(time.Hour).Unix(),
			Setter: s.Registry.Name,
		},
	})

	if err != nil {
		return time.Now(), false, err
	}

	return time.Unix(resp.Response.Value, 0), resp.Success, nil
}

func (s *GoServer) runLockTask(lockName string, t sFunc) (time.Time, error) {
	var tracer *rpcStats
	repeatRequests.With(prometheus.Labels{"method": "/" + t.key}).Inc()
	if s.RPCTracing {
		tracer = s.getTrace("/"+t.key, t.source)
	}

	ctx, cancel := utils.BuildContext(lockName, lockName)
	defer cancel()

	ti := time.Now()
	rt, err := t.lFun(ctx)
	if s.RPCTracing {
		s.recordTrace(ctx, tracer, "/"+t.key, time.Now().Sub(ti), err, "", false)
	}

	return rt, err
}

// Acquires a distributed lock for an hour
func (s *GoServer) setLock(lockName string, ti time.Time) error {
	conn, err := s.DialMaster("versionserver")
	if err != nil {
		return err
	}
	defer conn.Close()

	client := pbv.NewVersionServerClient(conn)
	ctx, cancel := utils.BuildContext(lockName, lockName)
	defer cancel()
	_, err = client.SetVersion(ctx, &pbv.SetVersionRequest{
		Set: &pbv.Version{
			Key:    lockName,
			Value:  ti.Unix(),
			Setter: s.Registry.Name,
		},
	})
	return err
}

func (s *GoServer) runLockingTask(t sFunc) {
	lockName := s.Registry.Name + "-" + t.key
	for true {
		//Read the lock
		ti, success, err := s.acquireLock(lockName)
		if success {
			// Run the task
			ti, err = s.runLockTask(lockName, t)
			if err == nil {
				// Set the lock
				err = s.setLock(lockName, ti)
			} else {
				s.Log(fmt.Sprintf("Failed to run task: %v, %v", t.key, err))
			}
		}

		// Wait until we can possibly acquire the lock
		s.Log(fmt.Sprintf("Sleeping for %v (%v) from %v", ti.Sub(time.Now()), err, t.key))
		time.Sleep(ti.Sub(time.Now()))
	}
}

func (s *GoServer) run(t sFunc) {
	time.Sleep(time.Minute)

	if t.source == "locking" {
		s.runLockingTask(t)
	}

	if t.d == 0 && !t.runOnce {
		t.fun(context.Background())
	} else {
		for true {
			name := fmt.Sprintf("%v-Repeat-(%v)-%v", s.Registry.Name, t.key, t.d)
			var ctx context.Context
			var cancel context.CancelFunc
			if t.noTrace {
				ctx, cancel = context.WithTimeout(context.Background(), time.Hour)
			} else {
				ctx, cancel = utils.BuildContext(name, name)
			}
			defer cancel()

			var err error
			if !t.nm {
				s.masterv++
				err = s.validateMaster(ctx)
				if err != nil {
					s.mastervfail++
				}
			}
			if err == nil {
				s.activeRPCsMutex.Lock()
				s.activeRPCs[name]++
				s.activeRPCsMutex.Unlock()

				s.runTimesMutex.Lock()
				s.runTimes[t.key] = time.Now()
				s.runTimesMutex.Unlock()

				var tracer *rpcStats
				if s.RPCTracing {
					tracer = s.getTrace("/"+t.key, t.source)
				}

				repeatRequests.With(prometheus.Labels{"method": name}).Inc()
				t1 := time.Now()
				s.runFunc(ctx, tracer, t)
				repeatLatency.With(prometheus.Labels{"method": name}).Observe(float64(time.Now().Sub(t1).Nanoseconds()))
				s.activeRPCsMutex.Lock()
				s.activeRPCs[name]--
				s.activeRPCsMutex.Unlock()

			} else if err != nil {
				var tracer *rpcStats
				if s.RPCTracing {
					tracer = s.getTrace("/"+t.key, t.source)
				}

				s.recordTrace(ctx, tracer, "/"+t.key, 0, err, "", false)
			}
			time.Sleep(t.d)
			if t.runOnce {
				return
			}
		}
	}
}

func (s *GoServer) runFunc(ctx context.Context, tracer *rpcStats, t sFunc) {
	ti := time.Now()
	var err error

	defer func() {
		if s.RPCTracing {
			if r := recover(); r != nil {
				err = fmt.Errorf("%v", r)
				s.SendCrash(ctx, fmt.Sprintf("%v", string(debug.Stack())), pbbs.Crash_PANIC)
				s.recordTrace(ctx, tracer, "/"+t.key, time.Now().Sub(ti), err, "", false)
			} else {
				s.recordTrace(ctx, tracer, "/"+t.key, time.Now().Sub(ti), err, "", false)
			}
		}

	}()
	err = t.fun(ctx)
}

//Read a protobuf
func (s *GoServer) Read(ctx context.Context, key string, typ proto.Message) (proto.Message, *pbks.ReadResponse, error) {
	return s.KSclient.Read(ctx, key, typ)
}

//GetServers gets an IP address from the discovery server
func (s *GoServer) GetServers(servername string) ([]*pb.RegistryEntry, error) {
	conn, err := s.dialler.Dial(utils.RegistryIP+":"+strconv.Itoa(utils.RegistryPort), grpc.WithInsecure())
	if err == nil {
		defer conn.Close()
		registry := s.clientBuilder.NewDiscoveryServiceClient(conn)
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		r, err := registry.ListAllServices(ctx, &pb.ListRequest{}, grpc.FailFast(false))
		e, ok := status.FromError(err)
		if ok && e.Code() == codes.Unavailable {
			r, err = registry.ListAllServices(ctx, &pb.ListRequest{}, grpc.FailFast(false))
		}

		if err == nil {
			arr := make([]*pb.RegistryEntry, 0)
			for _, s := range r.GetServices().GetServices() {
				if s.GetName() == servername {
					arr = append(arr, s)
				}
			}
			return arr, nil
		}
	}
	return nil, fmt.Errorf("Unable to establish connection")
}

// Serve Runs the server
func (s *GoServer) Serve(opt ...grpc.ServerOption) error {
	s.Log(fmt.Sprintf("Starting %v on port %v", s.RunningFile, s.Registry.Port))

	lis, err := net.Listen("tcp", ":"+strconv.Itoa(int(s.Port)))
	if err != nil {
		return err
	}
	fullOpts := append(opt,
		grpc.UnaryInterceptor(s.serverInterceptor),
	)
	server := grpc.NewServer(fullOpts...)
	s.Register.DoRegister(server)
	pbl.RegisterGoserverServiceServer(server, s)

	if !s.noRegister {
		if s.Registry.Version == pb.RegistryEntry_V1 {
			s.setupHeartbeats()
		}
	}
	go s.suicideWatch()

	// Enable profiling
	go http.ListenAndServe(fmt.Sprintf(":%v", s.Port+1), nil)

	// Enable prometheus
	http.Handle("/metrics", promhttp.Handler())
	go func() {
		http.ListenAndServe(fmt.Sprintf(":%v", s.Port+2), nil)
	}()

	// Background all the serving funcs
	for _, f := range s.servingFuncs {
		go s.run(f)
	}

	s.startup = time.Now()

	server.Serve(lis)
	return nil
}

func init() {
	resolver.Register(&utils.DiscoveryServerResolverBuilder{})
}

//RaiseIssue raises an issue
func (s *GoServer) RaiseIssue(ctx context.Context, title, body string, sticky bool) {
	if time.Now().Before(s.alertWait) {
		s.AlertsSkipped++
		return
	}

	s.alertWait = time.Now().Add(time.Minute * 10)
	s.AlertsFired++

	go func() {
		if !s.SkipIssue || len(body) == 0 {
			conn, err := s.NewBaseDial("githubcard")
			if err == nil {
				defer conn.Close()
				client := pbgh.NewGithubClient(conn)
				ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
				defer cancel()

				_, err := client.AddIssue(ctx, &pbgh.Issue{Service: s.Servername, Title: title, Body: body, Sticky: sticky}, grpc.FailFast(false))
				s.alertWait = time.Now().Add(time.Minute * 10)
				s.alertError = fmt.Sprintf("Cannot locate githubcard")

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
		if !s.SkipLog && s.Registry != nil {
			conn, err := s.NewBaseDial("logging")
			if err == nil {
				defer conn.Close()
				logger := lpb.NewLoggingServiceClient(conn)

				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()
				_, err := logger.Log(ctx, &lpb.LogRequest{Log: &lpb.Log{Timestamp: time.Now().Unix(), Origin: s.Registry.GetName(), Server: s.Registry.GetIdentifier(), Log: message, Ttl: int32((time.Hour * 24).Seconds())}})
				e, ok := status.FromError(err)
				if ok && err != nil && e.Code() != codes.DeadlineExceeded {
					s.failLogs++
					s.failMessage = fmt.Sprintf("%v", message)
				}
			}
		}
	}()
}

// RegisterServer Registers a server with the system and gets the port number it should use
func (s *GoServer) registerServer(IP string, servername string, external bool, v2 bool, im bool, dialler dialler, builder clientBuilder, getter hostGetter) (int32, error) {
	if v2 {
		conn, err := dialler.Dial(utils.LocalDiscover, grpc.WithInsecure())
		defer s.close(conn)
		registry := pb.NewDiscoveryServiceV2Client(conn)
		hostname, err := getter.Hostname()
		if err != nil {
			hostname = "Server-" + IP
		}
		entry := pb.RegistryEntry{Ip: IP, Name: servername, ExternalPort: external, Identifier: hostname, TimeToClean: 5000, Version: pb.RegistryEntry_V2, IgnoresMaster: im}
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()
		t := time.Now()
		r, err := registry.RegisterV2(ctx, &pb.RegisterRequest{Service: &entry}, grpc.FailFast(false))
		s.regTime = time.Now().Sub(t)
		if err != nil {
			return -1, err
		}
		s.Registry = r.GetService()

		return r.GetService().Port, nil
	}

	conn, err := dialler.Dial(utils.RegistryIP+":"+strconv.Itoa(utils.RegistryPort), grpc.WithInsecure())
	defer s.close(conn)
	if err != nil {
		return -1, err
	}

	registry := builder.NewDiscoveryServiceClient(conn)
	hostname, err := getter.Hostname()
	if err != nil {
		hostname = "Server-" + IP
	}
	entry := pb.RegistryEntry{Ip: IP, Name: servername, ExternalPort: external, Identifier: hostname, TimeToClean: 5000, IgnoresMaster: im}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	t := time.Now()
	r, err := registry.RegisterService(ctx, &pb.RegisterRequest{Service: &entry}, grpc.FailFast(false))
	s.regTime = time.Now().Sub(t)
	if err != nil {
		return -1, err
	}
	s.Registry = r.GetService()

	return r.GetService().Port, nil
}
