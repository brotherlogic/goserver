package goserver

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
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
	pbd "github.com/brotherlogic/monitor/monitorproto"
)

// Registerable Allows the system to register itself
type Registerable interface {
	DoRegister(server *grpc.Server)
	ReportHealth() bool
	Mote(master bool) error
	GetState() []*pbg.State
}

type baseRegistrable struct{ Registerable }

type sFunc struct {
	fun func()
	d   time.Duration
}

const (
	registerFreq = time.Second
)

// GoServer The basic server construct
type GoServer struct {
	Servername     string
	Port           int32
	Registry       *pb.RegistryEntry
	dialler        dialler
	monitorBuilder monitorBuilder
	clientBuilder  clientBuilder
	heartbeatChan  chan int
	heartbeatCount int
	heartbeatTime  time.Duration
	Register       Registerable
	SkipLog        bool
	servingFuncs   []sFunc
	KSclient       keystoreclient.Keystoreclient
	suicideTime    time.Duration
	Killme         bool
	hearts         int
	badHearts      int
	failMaster     int
	milestoneMutex *sync.Mutex
	milestones     map[string][]*pbd.Milestone
}

// PrepServer builds out the server for use.
func (s *GoServer) PrepServer() {
	s.heartbeatChan = make(chan int)
	s.heartbeatTime = registerFreq
	s.monitorBuilder = mainMonitorBuilder{}
	s.dialler = grpcDialler{}
	s.clientBuilder = mainBuilder{}
	s.suicideTime = time.Minute
	s.Killme = true
	s.hearts = 0
	s.badHearts = 0
	s.failMaster = 0
	s.milestoneMutex = &sync.Mutex{}
	s.milestones = make(map[string][]*pbd.Milestone)

	//Turn off grpc logging
	grpclog.SetLogger(log.New(ioutil.Discard, "", -1))
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
		conn, err := d.Dial(utils.RegistryIP+":"+strconv.Itoa(utils.RegistryPort), grpc.WithInsecure())
		if err == nil {
			c := b.NewDiscoveryServiceClient(conn)
			ctx, cancel := context.WithTimeout(context.Background(), registerFreq)
			defer cancel()
			s.hearts++
			r, err := c.RegisterService(ctx, &pb.RegisterRequest{Service: s.Registry}, grpc.FailFast(false))
			if err == nil {
				s.Registry = r.GetService()
			} else {
				s.badHearts++
				s.Log(fmt.Sprintf("ERROR ON REG: %v", err))
			}
			e, ok := status.FromError(err)
			if ok && (e.Code() != codes.DeadlineExceeded && e.Code() != codes.OK) {
				s.Log(fmt.Sprintf("DOWNGRADE: %v", err))
				s.Registry.Master = false
				s.failMaster++
			}
		}
		s.close(conn)
	}
}

//Log a simple string message
func (s *GoServer) Log(message string) {
	go func() {
		if !s.SkipLog {
			monitorIP, monitorPort := s.GetIP("monitor")
			if monitorPort > 0 {
				conn, err := s.dialler.Dial(monitorIP+":"+strconv.Itoa(int(monitorPort)), grpc.WithInsecure())
				if err == nil {
					monitor := s.monitorBuilder.NewMonitorServiceClient(conn)
					messageLog := &pbd.MessageLog{Message: message, Entry: s.Registry}
					ctx, cancel := context.WithTimeout(context.Background(), time.Second)
					defer cancel()
					monitor.WriteMessageLog(ctx, messageLog, grpc.FailFast(false))
					s.close(conn)
				}
			}
		}
	}()
}

//LogFunction logs a function call
func (s *GoServer) LogFunction(f string, t time.Time) {
	go func() {
		if !s.SkipLog {
			monitorIP, monitorPort := s.GetIP("monitor")
			if monitorPort > 0 {
				conn, err := s.dialler.Dial(monitorIP+":"+strconv.Itoa(int(monitorPort)), grpc.WithInsecure())
				if err == nil {
					monitor := s.monitorBuilder.NewMonitorServiceClient(conn)
					functionCall := &pbd.FunctionCall{Binary: s.Servername, Name: f, Time: int32(time.Now().Sub(t).Nanoseconds() / 1000000)}
					s.milestoneMutex.Lock()
					if val, ok := s.milestones[f]; ok {
						functionCall.Milestones = val
						delete(s.milestones, f)
					}
					s.milestoneMutex.Unlock()
					ctx, cancel := context.WithTimeout(context.Background(), time.Second)
					defer cancel()
					monitor.WriteFunctionCall(ctx, functionCall, grpc.FailFast(false))
					s.close(conn)
				}
			}
		}
	}()
}

// LogMilestone logs out a milestone
func (s *GoServer) LogMilestone(f string, m string, t time.Time) {
	s.milestoneMutex.Lock()
	if _, ok := s.milestones[f]; !ok {
		s.milestones[f] = make([]*pbd.Milestone, 0)
	}

	s.milestones[f] = append(s.milestones[f], &pbd.Milestone{Name: m, Time: int32(time.Now().Sub(t).Nanoseconds() / 1000000)})
	s.milestoneMutex.Unlock()
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

// RegisterServer Registers a server with the system and gets the port number it should use
func (s *GoServer) registerServer(IP string, servername string, external bool, dialler dialler, builder clientBuilder, getter hostGetter) int32 {
	conn, err := dialler.Dial(utils.RegistryIP+":"+strconv.Itoa(utils.RegistryPort), grpc.WithInsecure())
	if err != nil {
		return -1
	}

	registry := builder.NewDiscoveryServiceClient(conn)
	hostname, err := getter.Hostname()
	if err != nil {
		hostname = "Server-" + IP
	}
	entry := pb.RegistryEntry{Ip: IP, Name: servername, ExternalPort: external, Identifier: hostname}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r, err := registry.RegisterService(ctx, &pb.RegisterRequest{Service: &entry}, grpc.FailFast(false))
	if err != nil {
		s.close(conn)
		return -1
	}
	s.Registry = r.GetService()
	s.close(conn)

	return r.GetService().Port
}
