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
	"os/exec"
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
	}, []string{"method", "status"})
	serverPeak = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "rpc_server_peak_requests",
		Help: "The number of server requests",
	})

	startupRoutines = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "startup_routines",
		Help: "The number of running routines at startup",
	})

	startupTime = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "startup_time",
		Help: "The number of running routines at startup",
	})

	clientRequests = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "rpc_client_requests",
		Help: "The number of server requests",
	}, []string{"method", "status"})
	openClients = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "rpc_open_clients",
		Help: "The number of server requests",
	}, []string{"method"})
	repeatRequests = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "rpc_repeat_requests",
		Help: "The number of server requests",
	}, []string{"method"})
	lockingRequests = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "rpc_lock_requests",
		Help: "The number of server locking requests",
	}, []string{"lock"})

	serverLatency = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "rpc_server_latency",
		Help:    "The latency of server requests",
		Buckets: []float64{5, 10, 25, 50, 100, 250, 500, 1000, 2000, 4000, 8000, 16000, 32000, 64000, 128000, 256000, 1024000},
	}, []string{"method"})
	clientLatency = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "rpc_client_latency",
		Help:    "The latency of client requests",
		Buckets: []float64{.005 * 1000, .01 * 1000, .025 * 1000, .05 * 1000, .1 * 1000, .25 * 1000, .5 * 1000, 1 * 1000, 2.5 * 1000, 5 * 1000, 10 * 1000, 100 * 1000, 1000 * 1000},
	}, []string{"method"})
	repeatLatency = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "rpc_repeat_latency",
		Help:    "The latency of repeat requests",
		Buckets: []float64{.005 * 1000, .01 * 1000, .025 * 1000, .05 * 1000, .1 * 1000, .25 * 1000, .5 * 1000, 1 * 1000, 2.5 * 1000, 5 * 1000, 10 * 1000, 100 * 1000, 1000 * 1000},
	}, []string{"method"})
)

func (s *GoServer) ChooseLead(ctx context.Context, req *pbl.ChooseLeadRequest) (*pbl.ChooseLeadResponse, error) {
	s.Log(fmt.Sprintf("ChooseLead %v and %v -> %v", req.GetServer(), s.Registry.Identifier, strings.Compare(req.GetServer(), s.Registry.Identifier)))
	if strings.Compare(req.GetServer(), s.Registry.Identifier) > 0 {
		return &pbl.ChooseLeadResponse{Chosen: s.Registry.Identifier}, nil
	}

	return &pbl.ChooseLeadResponse{Chosen: req.GetServer()}, nil
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
	conn, err := s.FDial(fmt.Sprintf("%v:%v", entry.GetIdentifier(), entry.GetPort()))
	if err != nil {
		return err
	}
	defer conn.Close()
	client := pbl.NewGoserverServiceClient(conn)
	_, err = client.IsAlive(ctx, &pbl.Alive{})
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

				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()

				conn, err := s.FDialServer(ctx, "tracer")
				if err == nil {
					defer conn.Close()
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
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()

				conn, err := s.FDialServer(ctx, "tracer")
				if err == nil {
					defer conn.Close()
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
	s.RaiseIssue("Do Dial", fmt.Sprintf("Has called Do Dial -> %v", entry))
	return s.BaseDial(entry.Ip + ":" + strconv.Itoa(int(entry.Port)))
}

// BaseDial dials a connection
func (s *GoServer) BaseDial(c string) (*grpc.ClientConn, error) {
	badDial.With(prometheus.Labels{"call": "basedial-" + c}).Inc()
	s.RaiseIssue("BadBaseDial", fmt.Sprintf("%v has called BaseDial -> %v", s.Registry, c))
	return grpc.Dial(c, grpc.WithInsecure(), s.withClientUnaryInterceptor())
}

// NewBaseDial dials a connection
func (s *GoServer) NewBaseDial(c string) (*grpc.ClientConn, error) {
	badDial.With(prometheus.Labels{"call": "newbasedial-" + c}).Inc()
	s.RaiseIssue("BadNewBaseDial", fmt.Sprintf("%v has called NewBaseDial", s.Registry))
	s.Log(fmt.Sprintf("%v is calling %v via NewBaseDial", s.Registry.Name, c))
	return grpc.Dial("discovery:///"+c,
		grpc.WithInsecure(),
		s.withClientUnaryInterceptor())
}

// DialServer dials a given server
func (s *GoServer) DialServer(server, host string) (*grpc.ClientConn, error) {
	s.RaiseIssue("Dial Server", "Has called DialServer")
	entries, err := utils.ResolveAll(server)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.Identifier == host {
			return s.DoDial(entry)
		}
	}

	return nil, status.Errorf(codes.NotFound, "Unable to locate server called %v", server)
}

//DialLocal dials through the local discover
func (s *GoServer) DialLocal(server string) (*grpc.ClientConn, error) {
	s.RaiseIssue("Dial Local", "has called Dial Local")
	entry, err := utils.ResolveV2(server)
	if err != nil {
		return nil, err
	}
	return s.DoDial(entry)
}

// DialMaster dials the master server
func (s *GoServer) DialMaster(server string) (*grpc.ClientConn, error) {
	s.RaiseIssue("Dial Master", "Has called dial master")
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
	rSize := proto.Size(req.(proto.Message))
	if rSize > 100 {
		s.DLog(ctx, fmt.Sprintf("C: %v <- %v bytes", method, rSize))
	} else {
		s.DLog(ctx, fmt.Sprintf("C: %v <- %v bytes", method, req))
	}

	s.clientr++
	if s.clientr > 50 {
		s.Log(fmt.Sprintf("%v is running %v client requests -> %v (e.g. %v) [%+v]", s.Registry, s.clientr, method, req, ctx))
	}
	defer func() {
		s.clientr--
	}()

	s.activeRPCsMutex.Lock()
	s.activeRPCs[method]++
	s.activeRPCsMutex.Unlock()

	var tracer *rpcStats
	if s.RPCTracing {
		tracer = s.getTrace(method, "client")
	}

	// Calls the handler
	t := time.Now()

	var err error
	s.outgoing++
	openClients.With(prometheus.Labels{"method": method}).Inc()
	tracev, err := utils.GetContextKey(ctx)
	if err == nil {
		ctx = metadata.AppendToOutgoingContext(ctx, "trace-id", tracev)
	}
	err = invoker(ctx, method, req, reply, cc, opts...)
	openClients.With(prometheus.Labels{"method": method}).Dec()
	s.outgoing--

	clientRequests.With(prometheus.Labels{"status": status.Convert(err).Code().String(), "method": method}).Inc()
	clientLatency.With(prometheus.Labels{"method": method}).Observe(float64(time.Now().Sub(t).Nanoseconds() / 1000000))

	if s.RPCTracing {
		s.recordTrace(ctx, tracer, method, time.Now().Sub(t), err, req, false)
	}

	s.activeRPCsMutex.Lock()
	s.activeRPCs[method]--
	s.activeRPCsMutex.Unlock()

	s.DLog(ctx, fmt.Sprintf("C: %v -> %v", method, err))
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

	if err != nil {
		code := status.Convert(err)
		if code.Code() == codes.Unknown || code.Code() == codes.Internal {
			tracer.errors++
			tracer.lastError = fmt.Sprintf("%v", err)

			if float64(tracer.errors)/float64(tracer.count) > 0.8 && tracer.count > 10 {
				key, _ := utils.GetContextKey(ctx)
				s.RaiseIssue(fmt.Sprintf("Error for %v", name), fmt.Sprintf("%v; %v [%v]: %v calls %v errors (%v)", key, s.Registry.Identifier, s.RunningFile, tracer.count, tracer.errors, err))
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
			s.RaiseIssue("Over Active Service", fmt.Sprintf("rpc_%v%v is busy -> %v QPS / %v QTie", tracer.source, tracer.rpcName, qps, tracer.timeIn/time.Now().Sub(s.startup)))
		}
	}
}

func (s *GoServer) serverInterceptor(ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler) (interface{}, error) {

	if s.NoBody {
		s.DLog(ctx, fmt.Sprintf("S: %v <- bytes %v", info.FullMethod, proto.Size(req.(proto.Message))))
	} else {
		s.DLog(ctx, fmt.Sprintf("S: %v <- %v", info.FullMethod, req))
	}

	s.serverr++
	if s.serverr > s.serverrmax {
		s.serverrmax = s.serverr
		serverPeak.Set(float64(s.serverrmax))
	}
	defer func() {
		s.serverr--
	}()

	s.activeRPCsMutex.Lock()
	s.activeRPCs[info.FullMethod]++
	s.activeRPCsMutex.Unlock()

	var tracer *rpcStats
	if s.RPCTracing {
		tracer = s.getTrace(info.FullMethod, "server")
	}

	// Calls the handler
	if s.SendTrace {
		ctx = s.trace(ctx, info.FullMethod)
	}
	t := time.Now()
	h, err := s.runHandle(ctx, handler, req, tracer, info.FullMethod)

	if info.FullMethod != "/goserver.goserverService/IsAlive" || err != nil {
		serverRequests.With(prometheus.Labels{"status": status.Convert(err).Code().String(), "method": info.FullMethod}).Inc()
		serverLatency.With(prometheus.Labels{"method": info.FullMethod}).Observe(float64(time.Now().Sub(t).Milliseconds()))
	}

	if time.Now().Sub(t) > time.Hour {
		s.RaiseIssue("Slow Request", fmt.Sprintf("%v on %v/%v took %v (%v)", info.FullMethod, s.Registry.GetName(), s.Registry.GetIdentifier(), time.Now().Sub(t), req))
	}

	if err == nil && h != nil {
		if proto.Size(h.(proto.Message)) > 1024*1024 {
			s.RaiseIssue("Large Response", fmt.Sprintf("%v has produced a large response from %v (%vMb) -> %v", info.FullMethod, req, proto.Size(h.(proto.Message))/(1024*1024), ctx))
		}
	}
	s.activeRPCsMutex.Lock()
	s.activeRPCs[info.FullMethod]--
	s.activeRPCsMutex.Unlock()

	s.DLog(ctx, fmt.Sprintf("S: %v -> %v", info.FullMethod, err))

	return h, err
}

func (s *GoServer) runHandle(ctx context.Context, handler grpc.UnaryHandler, req interface{}, tracer *rpcStats, name string) (resp interface{}, err error) {
	ti := time.Now()

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

func (s *GoServer) checkMem(mem float64) {
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
			s.Log(fmt.Sprintf("Running GC with memory %v (took %v) %v -> %v", mem, time.Since(t), mem, mem2))
			mem = mem2
		}

		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		if m.HeapAlloc > 15000000 && time.Since(s.startup) > time.Minute*5 {
			var m runtime.MemStats
			runtime.ReadMemStats(&m)
			s.RaiseIssue("Memory Pressure", fmt.Sprintf("Memory is high: %v bytes but hang on: %v (%v/%v)", mem, m.HeapAlloc, s.Registry.Identifier, s.Registry.Name))
		}

		s.checkMem(mem)

		//commit suicide if we're detached from the parent and we're not sudoing
		if s.Killme {
			ctx, cancel := utils.ManualContext("goserver-suicide-"+s.Registry.Identifier, time.Minute)
			if s.Sudo {
				p, err := ps.FindProcess(os.Getppid())
				s.DLog(ctx, fmt.Sprintf("SUDO PARENT: %v and %v", p, err))
				if err == nil && p.PPid() == 1 {
					os.Exit(1)
				} else {
					if p.Executable() != "sudo" {
						s.CtxLog(ctx, fmt.Sprintf("Not exiting: %v, %+v", err, p))
					}
				}
			} else {
				s.DLog(ctx, fmt.Sprintf("NOSUDO PARENT: %v and %v", os.Getppid(), s.Killme))
				if os.Getppid() == 1 && s.Killme {
					os.Exit(1)
				}
			}
			cancel()
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

// RegisterServerIgnore registers this server with ignore master set.
func (s *GoServer) RegisterServer(servername string, external bool) error {
	s.serverName = servername

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
	start := s.registerAttempts
	for err != nil && s.registerAttempts-start < 10 {
		s.registerAttempts++
		port, err = s.getRegisteredServerPort(getLocalIP(), s.serverName, external, false)
		s.Port = port
	}
	return err
}

// RegisterServerV2 registers this server under the v2 protocol
func (s *GoServer) RegisterServerV2(external bool) error {
	// Short circuit if we don't need to register
	if s.noRegister {
		IP := getLocalIP()
		hostname, err := osHostGetter{}.Hostname()
		if err != nil {
			hostname = "Server-" + IP
		}
		entry := &pb.RegistryEntry{Ip: IP, Name: s.serverName, ExternalPort: false, Identifier: hostname, Port: s.Port, Version: pb.RegistryEntry_V2}
		s.Registry = entry

		return nil
	}

	err := fmt.Errorf("First fail")
	port := int32(0)
	start := s.registerAttempts
	for err != nil && s.registerAttempts-start < 10 {
		s.registerAttempts++
		port, err = s.getRegisteredServerPort(getLocalIP(), s.serverName, external, true)
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
	s.RaiseIssue("Locking task", "Is trying to register a locking task")
	s.servingFuncs = append(s.servingFuncs, sFunc{lFun: task, key: key, source: "locking"})
}

// RegisterServingTask registers tasks to run when serving
func (s *GoServer) RegisterServingTask(task func(ctx context.Context) error, key string) {
	s.RaiseIssue("Serving Task", "Is trying to register a serving task")
	s.servingFuncs = append(s.servingFuncs, sFunc{fun: task, d: 0, key: key, source: "repeat"})
}

// RegisterRepeatingTask registers a repeating task with a given frequency
func (s *GoServer) RegisterRepeatingTask(task func(ctx context.Context) error, key string, freq time.Duration) {
	s.RaiseIssue("Repeating Task", "Is trying to register a repeating task")
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
	s.RaiseIssue("Repeating Task No Trace", "Is trying to register a no trace repeating task")
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
	s.RaiseIssue("Repeat Task Non Master", "registered non master repeating task")
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
	if s.Registry == nil {
		return nil, fmt.Errorf("You haven't registered yet")
	}
	err := s.RegisterServerV2(s.Registry.ExternalPort)
	return &pbl.ReregisterResponse{}, err
}

// Shutdown brings the server down
func (s *GoServer) Shutdown(ctx context.Context, in *pbl.ShutdownRequest) (*pbl.ShutdownResponse, error) {
	s.LameDuck = true
	defer func() {
		go func() {
			time.Sleep(time.Second * 5)
			s.DLog(ctx, "Doing shutdown now")
			os.Exit(1)
		}()
	}()

	err := s.Register.Shutdown(ctx)
	if err != nil {
		return nil, err
	}

	conn, err := s.FDialServer(ctx, "discovery")
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	registry := pb.NewDiscoveryServiceV2Client(conn)

	_, err = registry.Unregister(ctx, &pb.UnregisterRequest{Reason: "requested-shutdown", Service: s.Registry})
	if err != nil {
		return nil, err
	}

	return &pbl.ShutdownResponse{}, nil
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

// Acquires a distributed lock for an hour
func (s *GoServer) acquireLock(lockName string) (time.Time, bool, error) {
	ctx, cancel := utils.ManualContext(lockName, time.Minute)
	defer cancel()

	conn, err := s.FDialServer(ctx, "versionserver")
	if err != nil {
		return time.Now(), false, err
	}
	defer conn.Close()

	client := pbv.NewVersionServerClient(conn)
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
	ctx, cancel := utils.ManualContext(lockName, time.Minute)
	defer cancel()

	conn, err := s.FDialServer(ctx, "versionserver")
	if err != nil {
		return err
	}
	defer conn.Close()

	client := pbv.NewVersionServerClient(conn)
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
		lockingRequests.With(prometheus.Labels{"lock": lockName}).Inc()
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
		s.Log(fmt.Sprintf("Sleeping for %v (%v) from %v", ti.Sub(time.Now())+time.Second*30, err, t.key))
		time.Sleep(ti.Sub(time.Now()) + time.Second*30)
	}
}

func (s *GoServer) runFuncInternal(t sFunc) {
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
		repeatLatency.With(prometheus.Labels{"method": name}).Observe(float64(time.Now().Sub(t1).Nanoseconds() / 1000000))
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

func (s *GoServer) run(t sFunc) {
	time.Sleep(time.Minute)

	if t.source == "locking" {
		s.runLockingTask(t)
	}

	if t.d == 0 && !t.runOnce {
		t.fun(context.Background())
	} else {
		for true {
			s.runFuncInternal(t)
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
	s.RaiseIssue("Trying to Get Servers", "Is trying to get servers")
	conn, err := s.dialler.Dial(utils.RegistryIP+":"+strconv.Itoa(utils.RegistryPort), grpc.WithInsecure())
	if err == nil {
		defer conn.Close()
		registry := s.clientBuilder.NewDiscoveryServiceClient(conn)
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		r, err := registry.ListAllServices(ctx, &pb.ListRequest{})
		e, ok := status.FromError(err)
		if ok && e.Code() == codes.Unavailable {
			r, err = registry.ListAllServices(ctx, &pb.ListRequest{})
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
	time.Sleep(time.Second * 2)
	s.Log(fmt.Sprintf("Starting (new) %v on port %v startup (%v)", s.RunningFile, s.Registry.Port, time.Since(s.registerTime)))
	startupTime.Set(float64(time.Since(s.registerTime).Milliseconds()))

	lis, err := net.Listen("tcp", ":"+strconv.Itoa(int(s.Port)))
	if err != nil {
		deets, err2 := exec.Command("sudo", "lsof", "-i", fmt.Sprintf(":%v", s.Port)).Output()
		s.Log(fmt.Sprintf("Unable to start: %v (%v), %v", err, string(deets), err2))

		// Silent exit
		if strings.Contains(fmt.Sprintf("%v", err2), "address already in use") {
			return nil
		}

		return fmt.Errorf("Bad startup (%v) -> %v, %v", string(deets), err, err2)
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
	go s.pickLead()

	// Enable profiling
	go http.ListenAndServe(fmt.Sprintf(":%v", s.Port+1), nil)

	// Enable prometheus
	if !s.NoProm && s.Registry.Identifier != "rdisplay" {
		http.Handle("/metrics", promhttp.Handler())
		go func() {
			http.ListenAndServe(fmt.Sprintf(":%v", s.Port+2), nil)
		}()
	}

	// Background all the serving funcs
	for _, f := range s.servingFuncs {
		go s.run(f)
	}

	s.startup = time.Now()

	// Count the number of active routines here
	startupRoutines.Set(float64(runtime.NumGoroutine()))

	server.Serve(lis)
	return nil
}

func init() {
	resolver.Register(&utils.DiscoveryServerResolverBuilder{})
}

func (s *GoServer) ImmediateIssue(ctx context.Context, title, body string, print bool) (*pbgh.Issue, error) {
	if s.SkipIssue {
		return &pbgh.Issue{Number: 12}, nil
	}
	conn, err := s.FDialServer(ctx, "githubcard")
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	client := pbgh.NewGithubClient(conn)
	return client.AddIssue(ctx, &pbgh.Issue{Service: s.serverName, Title: title, Body: body, Sticky: false, PrintImmediately: print})
}

func (s *GoServer) DeleteIssue(ctx context.Context, number int32) error {
	if s.SkipIssue {
		return nil
	}
	conn, err := s.FDialServer(ctx, "githubcard")
	if err != nil {
		return err
	}
	defer conn.Close()

	client := pbgh.NewGithubClient(conn)
	_, err = client.DeleteIssue(ctx, &pbgh.DeleteRequest{Issue: &pbgh.Issue{Service: s.serverName, Number: number}})
	return err
}

//RaiseIssue raises an issue
func (s *GoServer) RaiseIssue(title, body string) {
	s.IssueCount++
	if s.SkipIssue {
		log.Printf("Raising Issue %v -> %v", title, body)
	}
	if time.Now().Before(s.alertWait) {
		s.AlertsSkipped++
	}

	s.alertWait = time.Now().Add(time.Minute * 10)
	s.AlertsFired++

	go func() {
		if !s.SkipIssue && len(body) != 0 {
			ctx, cancel := utils.ManualContext(fmt.Sprintf("%v-%v", s.Registry.GetName(), "issue"), time.Minute)
			defer cancel()
			s.ImmediateIssue(ctx, title, body, false)
		} else {
			s.alertError = "Skip log enabled"
		}
	}()
}

var issueBounces = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: "server_bounces",
	Help: "The number of server requests",
}, []string{"error"})

//BounceIssue raises an issue for a different source
func (s *GoServer) BounceIssue(title, body string, job string) {
	s.AlertsFired++
	go func() {
		if !s.SkipLog {
			ctx, cancel := utils.ManualContext(fmt.Sprintf("%v-%v", s.Registry.GetName(), "issue"), time.Minute)
			defer cancel()
			conn, err := s.FDialServer(ctx, "githubcard")
			if err == nil {
				defer conn.Close()
				client := pbgh.NewGithubClient(conn)
				ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
				defer cancel()

				_, err := client.AddIssue(ctx, &pbgh.Issue{Service: job, Title: title, Body: body})
				issueBounces.With(prometheus.Labels{"error": fmt.Sprintf("%v", err)}).Inc()
				if err != nil {
					s.alertError = fmt.Sprintf("Failure to add issue: %v", err)
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
	conn, err := s.FDialServer(ctx, "buildserver")
	if err != nil {
		s.Log(fmt.Sprintf("Unable to dial buildserver: %v", err))
	}
	if err == nil {
		defer conn.Close()
		client := pbbs.NewBuildServiceClient(conn)

		// Build out the current state
		infoString := ""

		_, err := client.ReportCrash(ctx, &pbbs.CrashRequest{
			Version: s.RunningFile,
			Origin:  s.Registry.Name,
			Job: &pbgbs.Job{
				Name: s.Registry.Name,
			},
			Crash: &pbbs.Crash{
				ErrorMessage: crashText + "\n" + infoString,
				CrashType:    ctype}})
		s.Log(fmt.Sprintf("Reported crash: %v", err))
	}
}

//PLog a simple string message with priority
func (s *GoServer) PLog(ictx context.Context, message string, level pbd.LogLevel) {
	if s.SkipLog {
		log.Printf("LOG %v ->%v with %v", message, s.activeRPCsMutex, s.SkipLog)
	}
	go func() {
		if !s.SkipLog && s.Registry != nil {
			ctx, cancel := utils.ManualContext(fmt.Sprintf("%v-%v", s.Registry.GetName(), "logging"), time.Second)
			defer cancel()
			conn, err := s.FDialSpecificServer(ctx, "logging", s.Registry.Identifier)
			if err == nil {
				defer conn.Close()
				logger := lpb.NewLoggingServiceClient(conn)

				_, err := logger.Log(ctx, &lpb.LogRequest{Log: &lpb.Log{Timestamp: time.Now().UnixNano(), Origin: s.Registry.GetName(), Server: s.Registry.GetIdentifier(), Log: message, Ttl: int32((time.Hour * 24).Seconds())}})
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
func (s *GoServer) registerServer(IP string, servername string, external bool, v2 bool, dialler dialler, builder clientBuilder, getter hostGetter) (int32, error) {
	if !v2 {
		s.RaiseIssue("Trying to register as v1", "Is trying to register")
	}
	if v2 {
		// Now we can prep the dlog
		if !s.preppedDLog && s.DiskLog {
			s.prepDLog(servername)
		} else {
			s.Log(fmt.Sprintf("Not setting up disk logging %v and %v", s.preppedDLog, s.DiskLog))
		}

		conn, err := dialler.Dial(utils.LocalDiscover, grpc.WithInsecure())
		defer s.close(conn)
		registry := pb.NewDiscoveryServiceV2Client(conn)
		hostname, err := getter.Hostname()
		if err != nil {
			hostname = "Server-" + IP
		}
		entry := &pb.RegistryEntry{Ip: IP, Name: servername, ExternalPort: external, Identifier: hostname, TimeToClean: 5000, Version: pb.RegistryEntry_V2}
		ctx, cancel := utils.ManualContext("register-"+servername, time.Second*30)
		defer cancel()
		t := time.Now()
		s.DLog(ctx, fmt.Sprintf("REGISTER: %v", entry))
		r, err := registry.RegisterV2(ctx, &pb.RegisterRequest{Service: entry})
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
	entry := pb.RegistryEntry{Ip: IP, Name: servername, ExternalPort: external, Identifier: hostname, TimeToClean: 5000}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	t := time.Now()
	r, err := registry.RegisterService(ctx, &pb.RegisterRequest{Service: &entry})
	s.regTime = time.Now().Sub(t)
	if err != nil {
		return -1, err
	}
	s.Registry = r.GetService()

	return r.GetService().Port, nil
}
