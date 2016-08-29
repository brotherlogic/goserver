package goserver

import (
	"log"
	"net"
	"strconv"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	pb "github.com/brotherlogic/discovery/proto"
	pbd "github.com/brotherlogic/monitor/monitorproto"
)

const (
	registryIP   = "10.0.1.17"
	registryPort = 50055
)

// Registerable Allows the system to register itself
type Registerable interface {
	DoRegister(server *grpc.Server)
}

type baseRegistrable struct{ Registerable }

// GoServer The basic server construct
type GoServer struct {
	servername     string
	port           int32
	registry       pb.RegistryEntry
	monitor        pb.RegistryEntry
	dialler        dialler
	monitorBuilder monitorBuilder
	clientBuilder  clientBuilder
	heartbeatChan  chan int
	heartbeatCount int
	heartbeatTime  time.Duration
	Register       Registerable
	SkipLog        bool
}

// PrepServer builds out the server for use.
func (s *GoServer) PrepServer() {
	s.heartbeatChan = make(chan int)
	s.heartbeatTime = time.Minute * 1
	s.monitorBuilder = mainMonitorBuilder{}
	s.dialler = grpcDialler{}
	s.clientBuilder = mainBuilder{}
}

func (s *GoServer) teardown() {
	s.heartbeatChan <- 0
}

func (s *GoServer) heartbeat() {
	running := true
	for running {
		s.sendHeartbeat(s.monitor.Ip, int(s.monitor.Port), s.dialler, s.monitorBuilder)
		select {
		case <-s.heartbeatChan:
			running = false
		default:
			log.Printf("Sleeping for %v", s.heartbeatTime)
			time.Sleep(s.heartbeatTime)
		}
	}
}

//Log a simple string message
func (s *GoServer) Log(message string) {
	if !s.SkipLog {
		conn, _ := s.dialler.Dial(s.monitor.Ip+":"+strconv.Itoa(int(s.monitor.Port)), grpc.WithInsecure())
		monitor := s.monitorBuilder.NewMonitorServiceClient(conn)
		messageLog := &pbd.MessageLog{Message: message, Entry: &s.registry}
		monitor.WriteMessageLog(context.Background(), messageLog)
		s.close(conn)
	}
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

func (s *GoServer) sendHeartbeat(monitorIP string, monitorPort int, dialler dialler, builder monitorBuilder) {
	conn, _ := dialler.Dial(monitorIP+":"+strconv.Itoa(monitorPort), grpc.WithInsecure())
	monitor := builder.NewMonitorServiceClient(conn)
	monitor.ReceiveHeartbeat(context.Background(), &s.registry)
	log.Printf("BEAT")
	s.heartbeatCount++
	s.close(conn)
}

func getLocalIP() string {
	ifaces, _ := net.Interfaces()

	var ip net.IP
	for _, i := range ifaces {
		log.Printf("HERE 1 = %v", i)
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

func (s *GoServer) setupHeartbeats(dialler dialler, builder clientBuilder) {
	log.Printf("Setting up heartbeats")
	conn, _ := dialler.Dial(registryIP+":"+strconv.Itoa(registryPort), grpc.WithInsecure())

	registry := builder.NewDiscoveryServiceClient(conn)
	log.Printf("REGISTRY IS HERE %v", registry)
	entry := pb.RegistryEntry{Name: "monitor"}
	r, err := registry.Discover(context.Background(), &entry)

	if err == nil {
		s.monitor = *r
		log.Printf("Running heartbeats")
		go s.heartbeat()
	}
	log.Printf("Heartbeats beating")
	s.close(conn)
}

// RegisterServer Registers a server with the system and gets the port number it should use
func (s *GoServer) registerServer(IP string, servername string, external bool, dialler dialler, builder clientBuilder) int32 {
	conn, err := dialler.Dial(registryIP+":"+strconv.Itoa(registryPort), grpc.WithInsecure())
	if err != nil {
		log.Printf("Could not connect: %v", err)
		return -1
	}

	registry := builder.NewDiscoveryServiceClient(conn)
	entry := pb.RegistryEntry{Ip: IP, Name: servername, ExternalPort: external}
	r, err := registry.RegisterService(context.Background(), &entry)
	if err != nil {
		log.Printf("Could not register this service: %v", err)
		return -1
	}
	s.registry = *r
	s.close(conn)
	return r.Port
}
