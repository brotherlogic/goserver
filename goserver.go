package goserver

import (
	"crypto/md5"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/brotherlogic/keystore/client"
	"golang.org/x/net/context"
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
	AlertsSkipped           int64
	RunInV2                 bool
	noRegister              bool
	latestMem               int
	activeRPCs              map[string]int
	activeRPCsMutex         *sync.Mutex
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

	//Turn off grpc logging
	grpclog.SetLogger(log.New(ioutil.Discard, "", -1))

	// Set the file details
	ex, _ := os.Executable()
	data, _ := ioutil.ReadFile(ex)
	s.RunningFile = fmt.Sprintf("%x", md5.Sum(data))

	// Enable RPC tracing
	s.RPCTracing = true
	s.SendTrace = true

	s.runTimes = make(map[string]time.Time)
	s.runTimesMutex = &sync.Mutex{}

	// Build out the ksclient
	s.KSclient = *keystoreclient.GetClient(s.DialMaster)
}

func (s *GoServer) teardown() {
	s.heartbeatChan <- 0
}

func (s *GoServer) heartbeat() {
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

		conn, err := d.Dial(utils.RegistryIP+":"+strconv.Itoa(utils.RegistryPort), grpc.WithInsecure())
		if err == nil {
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
		s.close(conn)
	}
}

//Log a simple string message
func (s *GoServer) Log(message string) {
	s.PLog(message, pbd.LogLevel_DISCARD)
}

//GetIP gets an IP address from the discovery server
func (s *GoServer) GetIP(servername string) (string, int) {
	conn, err := s.dialler.Dial(utils.RegistryIP+":"+strconv.Itoa(utils.RegistryPort), grpc.WithInsecure())
	if err == nil {
		registry := s.clientBuilder.NewDiscoveryServiceClient(conn)
		entry := &pb.RegistryEntry{Name: servername}
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		r, err := registry.Discover(ctx, &pb.DiscoverRequest{Request: entry}, grpc.FailFast(false))
		e, ok := status.FromError(err)
		if ok && e.Code() == codes.Unavailable {
			r, err = registry.Discover(ctx, &pb.DiscoverRequest{Request: entry}, grpc.FailFast(false))
		}

		if err == nil {
			s.close(conn)
			return r.GetService().Ip, int(r.GetService().Port)
		}
	}
	s.close(conn)
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
	conn, err := dialler.Dial(utils.RegistryIP+":"+strconv.Itoa(utils.RegistryPort), grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	registry := builder.NewDiscoveryServiceClient(conn)
	entry := pb.RegistryEntry{Name: server}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r, err := registry.Discover(ctx, &pb.DiscoverRequest{Request: &entry}, grpc.FailFast(false))
	if err != nil {
		s.close(conn)
		return nil, err
	}
	s.close(conn)

	return dialler.Dial(r.GetService().Ip+":"+strconv.Itoa(int(r.GetService().Port)), grpc.WithInsecure())
}

func (s *GoServer) setupHeartbeats() {
	go s.heartbeat()
}
