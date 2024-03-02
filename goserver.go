package goserver

import (
	"context"
	"crypto/md5"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"

	githubridgeclient "github.com/brotherlogic/githubridge/client"
	keystoreclient "github.com/brotherlogic/keystore/client"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/grpc/status"

	pb "github.com/brotherlogic/discovery/proto"
	pbg "github.com/brotherlogic/goserver/proto"
	utils "github.com/brotherlogic/goserver/utils"
	pbd "github.com/brotherlogic/monitor/proto"

	pid "github.com/struCoder/pidusage"
)

// Registerable Allows the system to register itself
type Registerable interface {
	DoRegister(server *grpc.Server)
	ReportHealth() bool
	GetState() []*pbg.State
	Shutdown(ctx context.Context) error
}

type baseRegistrable struct{ Registerable }

type sFunc struct {
	fun     func(ctx context.Context) error
	lFun    func(ctx context.Context) (time.Time, error)
	d       time.Duration
	nm      bool
	key     string
	noTrace bool
	source  string
	runOnce bool
}

const (
	registerFreq = time.Second
)

var (
	runningBinaryTimestamp = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "running_binary_timestamp",
		Help: "The timestamp of the running binary",
	})
	badDial = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "bad_dial",
		Help: "Calls into bad dials",
	}, []string{"call"})
	elections = promauto.NewCounter(prometheus.CounterOpts{
		Name: "elections",
		Help: "Calls into bad dials",
	})
	bits = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "running_bits",
		Help: "Calls into bad dials",
	})
)

// GoServer The basic server construct
type GoServer struct {
	Port                    int32
	Registry                *pb.RegistryEntry
	dialler                 dialler
	monitorBuilder          monitorBuilder
	clientBuilder           clientBuilder
	heartbeatChan           chan int
	heartbeatCount          int
	heartbeatTime           time.Duration
	Register                Registerable
	SkipLog                 bool
	servingFuncs            []sFunc
	KSclient                keystoreclient.Keystoreclient
	suicideTime             time.Duration
	Killme                  bool
	hearts                  int
	BadHearts               int
	failMaster              int
	milestoneMutex          *sync.Mutex
	failLogs                int
	failMessage             string
	startup                 time.Time
	config                  *pbg.ServerConfig
	cpuMutex                *sync.Mutex
	AlertsFired             int
	Sudo                    bool
	alertError              string
	RunningFile             string
	badHeartMessage         string
	traceCount              int
	traceFails              int
	traceFailMessage        string
	moteCount               int
	lastMoteTime            time.Duration
	lastMoteFail            string
	badPorts                int64
	regTime                 time.Duration
	MemCap                  int
	traces                  []*rpcStats
	RPCTracing              bool
	LameDuck                bool
	SendTrace               bool
	masterRequests          int64
	masterRequestFails      int64
	masterRequestFailReason string
	runTimes                map[string]time.Time
	runTimesMutex           *sync.Mutex
	marks                   int64
	incoming                int64
	outgoing                int64
	registerAttempts        int64
	alertWait               time.Time
	logWait                 time.Time
	AlertsSkipped           int64
	logsSkipped             int64
	RunInV2                 bool
	noRegister              bool
	latestMem               int
	activeRPCs              map[string]int
	activeRPCsMutex         *sync.Mutex
	SkipIssue               bool
	masterMutex             *sync.Mutex
	masterv                 int64
	mastervfail             int64
	serverr                 int
	serverrmax              int
	clientr                 int
	DiskLog                 bool
	dlogHandle              *os.File
	NoBody                  bool
	preppedDLog             bool
	SkipElect               bool
	Store                   translatedStore
	FlagUseDataStore        bool
	FlagUseDStore           bool
	NeedsLead               bool
	CurrentLead             string
	LeadState               pbg.LeadState
	LeadFails               int
	lastLogCheck            time.Time
	NoProm                  bool
	Bits                    int
	runner                  string
	registerTime            time.Time
	serverName              string
	IssueCount              int
	ElideRequests           bool
	ghbclient               githubridgeclient.GithubridgeClient
}

func (s *GoServer) pickLead() {
	ctx, cancel := utils.ManualContext("server-picklead", time.Minute)
	defer cancel()
	if !s.NeedsLead {
		return
	}

	for !s.LameDuck {
		s.CtxLog(ctx, fmt.Sprintf("ElectionState %v -> %v", s.LeadState, s.CurrentLead))
		switch s.LeadState {
		case pbg.LeadState_ELECTING:
			s.runSimpleElection()
		case pbg.LeadState_FOLLOWER:
			if !s.verifyFollower() {
				s.LeadFails++
			} else {
				s.LeadFails = 0
			}

			if s.LeadFails > 10 {
				s.LeadState = pbg.LeadState_ELECTING
			}
		}

		time.Sleep(time.Second * 30)
	}
}

func (s *GoServer) verifyFollower() bool {
	ctx, cancel := utils.ManualContext("verifying", time.Second*30)
	defer cancel()

	conn, err := s.FDialSpecificServer(ctx, s.Registry.Name, s.CurrentLead)
	if err != nil {
		s.CtxLog(ctx, fmt.Sprintf("Unable to dial leader: %v -> %v", s.CurrentLead, err))
		return false
	}
	defer conn.Close()

	gsclient := pbg.NewGoserverServiceClient(conn)
	_, err = gsclient.IsAlive(ctx, &pbg.Alive{})
	if err != nil {
		s.CtxLog(ctx, fmt.Sprintf("Leader is not alive: %v -> %v", s.CurrentLead, err))
		return false
	}

	return true
}

func (s *GoServer) runSimpleElection() {
	elections.Inc()
	ctx, cancel := utils.ManualContext("electing", time.Second*30)
	defer cancel()

	friends, err := s.FFind(ctx, s.Registry.Name)
	if err != nil {
		s.CtxLog(ctx, fmt.Sprintf("Unable to list friends for election: %v", err))
		return
	}

	for _, friend := range friends {
		s.CtxLog(ctx, fmt.Sprintf("ChooseLead comparison %v and %v", friend, s.Registry.Identifier))
		if !strings.HasPrefix(friend, s.Registry.Identifier) {
			conn, err := s.FDial(friend)
			if err != nil {
				return
			}
			defer conn.Close()

			gsclient := pbg.NewGoserverServiceClient(conn)
			win, err := gsclient.ChooseLead(ctx, &pbg.ChooseLeadRequest{Server: s.Registry.Identifier})
			if err != nil {
				return
			}

			if win.GetChosen() != s.Registry.Identifier {
				s.LeadState = pbg.LeadState_FOLLOWER
				s.CurrentLead = win.GetChosen()
				return
			}
		}
	}

	s.LeadState = pbg.LeadState_LEADER
}

func (s *GoServer) getCPUUsage() (float64, float64) {
	s.cpuMutex.Lock()
	defer s.cpuMutex.Unlock()
	v, _ := pid.GetStat(os.Getpid())
	return v.CPU, v.Memory
}

// RunSudo runs as sudo
func (s *GoServer) RunSudo() {
	s.Sudo = true
}

// PrepServer builds out the server for use.
func (s *GoServer) PrepServer(name string) {
	s.registerTime = time.Now()

	s.serverName = name
	s.prepareServer(false)

	//tp, err := tracerProvider(name)
	//if err != nil {
	//		s.RaiseIssue("Unable to trace", fmt.Sprintf("Error is %v", err))
	//	}
	//otel.SetTracerProvider(tp)
	//otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

}

// PrepServerNoRegister builds out a server that doesn't register
func (s *GoServer) PrepServerNoRegister(name string, port int32) {
	s.Port = port
	s.registerTime = time.Now()
	s.serverName = name
	s.prepareServer(true)
}

func (s *GoServer) prepareServer(register bool) {
	s.activeRPCs = make(map[string]int)
	s.activeRPCsMutex = &sync.Mutex{}
	s.heartbeatChan = make(chan int)
	s.heartbeatTime = registerFreq
	s.monitorBuilder = mainMonitorBuilder{}
	s.dialler = grpcDialler{}
	s.clientBuilder = mainBuilder{}
	s.suicideTime = time.Minute
	s.Killme = true
	s.hearts = 0
	s.BadHearts = 0
	s.failMaster = 0
	s.failLogs = 0
	s.failMessage = ""
	s.config = &pbg.ServerConfig{}
	s.cpuMutex = &sync.Mutex{}
	s.AlertsFired = 0
	s.alertError = ""
	s.badHeartMessage = ""
	s.traceCount = 0
	s.traceFails = 0
	s.traceFailMessage = ""
	s.moteCount = 0
	s.MemCap = 100000000
	s.traces = []*rpcStats{}
	s.alertWait = time.Now()
	s.noRegister = register
	s.DiskLog = true

	//Turn off grpc logging
	grpclog.SetLogger(log.New(ioutil.Discard, "", -1))

	// Set the file details
	ex, _ := os.Executable()
	data, _ := ioutil.ReadFile(ex)
	info, _ := os.Stat(ex)
	s.runner = ex
	s.RunningFile = fmt.Sprintf("%x", md5.Sum(data))
	runningBinaryTimestamp.Set(float64(info.ModTime().Unix()))

	// Enable RPC tracing
	s.RPCTracing = true

	// Don't send out traces
	s.SendTrace = false

	s.runTimes = make(map[string]time.Time)
	s.runTimesMutex = &sync.Mutex{}

	// Build out the ksclient
	s.KSclient = *keystoreclient.GetClient(s.FDialServer)

	s.masterMutex = &sync.Mutex{}

	if s.FlagUseDStore {
		s.Store = &mts{store: &dstore{s.FDialServer}}
	} else if s.FlagUseDataStore {
		s.Store = &mts{store: &datastore{s.FDialServer}}
	} else {
		s.Store = &mts{store: &keystore{s.FDialServer}}
	}

	s.lastLogCheck = time.Now()

	output, err := exec.Command("cat", "/proc/version").CombinedOutput()
	if err != nil {
		log.Fatalf("Cannot run job: %v", err)
	}

	if strings.Contains(string(output), "aarch64") {
		s.Bits = 64
	} else {
		s.Bits = 32
	}
	bits.Set(float64(s.Bits))

	//Prep the dlog since we're not on the register path
	if !s.preppedDLog {
		s.prepDLog(s.serverName)
	}
}

func (s *GoServer) teardown() {
	s.heartbeatChan <- 0
}

func (s *GoServer) heartbeat() {
	s.RaiseIssue("Heartbeat call", fmt.Sprintf("Called heartbeat"))
	running := true
	for running {
		s.reregister(s.dialler, s.clientBuilder)
		select {
		case <-s.heartbeatChan:
			running = false
		default:
			time.Sleep(s.heartbeatTime)
		}
	}
}

func (s *GoServer) reregister(d dialler, b clientBuilder) {
	if s.Registry != nil {
		conn, err := s.FDial(utils.RegistryIP + ":" + strconv.Itoa(utils.RegistryPort))
		if err == nil {
			defer s.close(conn)
			c := b.NewDiscoveryServiceClient(conn)
			ctx, cancel := context.WithTimeout(context.Background(), registerFreq)
			defer cancel()
			s.hearts++
			r, err := c.RegisterService(ctx, &pb.RegisterRequest{Service: s.Registry})
			if err == nil {
				if s.Registry.Port > 0 && s.Registry.Port != r.GetService().Port {
					s.badPorts++
				}
				s.Registry = r.GetService()
			} else {
				s.BadHearts++
			}
			e, ok := status.FromError(err)
			if ok && (e.Code() != codes.DeadlineExceeded && e.Code() != codes.OK && e.Code() != codes.Unavailable) {
				s.badHeartMessage = fmt.Sprintf("%v", err)
				s.failMaster++
			}
		}
	}
}

// GetIP gets an IP address from the discovery server
func (s *GoServer) GetIP(servername string) (string, int) {
	badDial.With(prometheus.Labels{"call": "getip-" + servername}).Inc()
	conn, err := s.dialler.Dial(utils.RegistryIP+":"+strconv.Itoa(utils.RegistryPort), grpc.WithInsecure())
	if err == nil {
		defer s.close(conn)
		registry := s.clientBuilder.NewDiscoveryServiceClient(conn)
		entry := &pb.RegistryEntry{Name: servername}
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		s.RaiseIssue("GetIp Call", fmt.Sprintf("Called GetIP"))
		r, err := registry.Discover(ctx, &pb.DiscoverRequest{Request: entry})
		e, ok := status.FromError(err)
		if ok && e.Code() == codes.Unavailable {
			r, err = registry.Discover(ctx, &pb.DiscoverRequest{Request: entry})
		}

		if err == nil {
			return r.GetService().Ip, int(r.GetService().Port)
		}
	}
	return "", -1
}

type hostGetter interface {
	Hostname() (string, error)
}

type monitorBuilder interface {
	NewMonitorServiceClient(conn *grpc.ClientConn) pbd.MonitorServiceClient
}

type dialler interface {
	Dial(host string, opts ...grpc.DialOption) (*grpc.ClientConn, error)
}

type clientBuilder interface {
	NewDiscoveryServiceClient(conn *grpc.ClientConn) pb.DiscoveryServiceClient
}

func (s *GoServer) getLocalIP(ctx context.Context) string {
	ifaces, _ := net.Interfaces()

	var ip net.IP
	for _, i := range ifaces {
		addrs, _ := i.Addrs()

		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
				if ipnet.IP.To4() != nil && strings.Contains(ipnet.IP.String(), "192.168") {
					ip = ipnet.IP
				}
			}
		}
	}

	if ip == nil || ip.String() == "<nil>" {
		s.CtxLog(ctx, fmt.Sprintf("Unable to get IP: %v", ifaces))
	}

	return ip.String()
}

// Dial a local server
func (s *GoServer) Dial(server string, dialler dialler, builder clientBuilder) (*grpc.ClientConn, error) {
	badDial.With(prometheus.Labels{"call": "dial-" + server}).Inc()
	conn, err := dialler.Dial(utils.RegistryIP+":"+strconv.Itoa(utils.RegistryPort), grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	defer s.close(conn)
	registry := builder.NewDiscoveryServiceClient(conn)
	entry := pb.RegistryEntry{Name: server}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	s.RaiseIssue("Called Dial", fmt.Sprintf("Called Dial"))
	r, err := registry.Discover(ctx, &pb.DiscoverRequest{Request: &entry})
	if err != nil {
		return nil, err
	}

	return dialler.Dial(r.GetService().Ip+":"+strconv.Itoa(int(r.GetService().Port)), grpc.WithInsecure())
}

func (s *GoServer) setupHeartbeats() {
	go s.heartbeat()
}

var (
	skipLog = promauto.NewCounter(prometheus.CounterOpts{
		Name: "skiplog",
		Help: "Calls into bad dials",
	})
)

func (s *GoServer) CtxLog(ctx context.Context, message string) {
	if s.SkipLog {
		log.Printf(message)
	}

	if !s.SkipLog {
		s.DLog(ctx, message)
	}

	if time.Now().Before(s.logWait) {
		skipLog.Inc()
		return
	}
	s.logWait = time.Now().Add(time.Second)
	s.PLog(ctx, message, pbd.LogLevel_DISCARD)
}
