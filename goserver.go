package goserver

import (
	"context"
	"crypto/md5"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strconv"
	"sync"
	"time"

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
	Mote(ctx context.Context, master bool) error
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
)

// GoServer The basic server construct
type GoServer struct {
	Servername              string
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
}

func (s *GoServer) getCPUUsage() (float64, float64) {
	s.cpuMutex.Lock()
	defer s.cpuMutex.Unlock()
	v, _ := pid.GetStat(os.Getpid())
	return v.CPU, v.Memory
}

//RunSudo runs as sudo
func (s *GoServer) RunSudo() {
	s.Sudo = true
}

// PrepServer builds out the server for use.
func (s *GoServer) PrepServer() {
	s.prepareServer(false)
}

// PrepServerNoRegister builds out a server that doesn't register
func (s *GoServer) PrepServerNoRegister(port int32) {
	s.Port = port
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

	if s.FlagUseDataStore {
		s.Store = &mts{store: &datastore{s.FDialServer}}
	} else {
		s.Store = &mts{store: &keystore{s.FDialServer}}
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
		//We can't be master if we're lame ducking
		if s.LameDuck {
			s.Registry.Master = false
		}

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
				if !s.Registry.Master && r.GetService().WeakMaster && !r.GetService().Master {
					s.masterRequests++
					r.GetService().Master = false
				}
				s.Registry = r.GetService()
			} else {
				s.BadHearts++
			}
			e, ok := status.FromError(err)
			if ok && (e.Code() != codes.DeadlineExceeded && e.Code() != codes.OK && e.Code() != codes.Unavailable) {
				s.badHeartMessage = fmt.Sprintf("%v", err)
				s.Registry.Master = false
				s.failMaster++
			}
		}
	}
}

//GetIP gets an IP address from the discovery server
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

func getLocalIP() string {
	ifaces, _ := net.Interfaces()

	var ip net.IP
	for _, i := range ifaces {
		addrs, _ := i.Addrs()

		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
				if ipnet.IP.To4() != nil {
					ip = ipnet.IP
				}
			}
		}
	}

	return ip.String()
}

//Dial a local server
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

func (s *GoServer) Log(message string) {
	s.CtxLog(context.Background(), message)
}

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
