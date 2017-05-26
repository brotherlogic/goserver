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
	registryIP   = "192.168.86.34"
	registryPort = 50055
)

// Registerable Allows the system to register itself
type Registerable interface {
	DoRegister(server *grpc.Server)
	ReportHealth() bool
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
	servingFuncs   []func()
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
		s.reregister(s.dialler, s.clientBuilder)
		select {
		case <-s.heartbeatChan:
			running = false
		default:
			log.Printf("Sleeping for %v", s.heartbeatTime)
			time.Sleep(s.heartbeatTime)
		}
	}
}

func (s *GoServer) reregister(d dialler, b clientBuilder) {
	conn, err := d.Dial(registryIP+":"+strconv.Itoa(registryPort), grpc.WithInsecure())
	defer conn.Close()
	log.Printf("Re-registering %v:%v -> %v", registryIP, registryPort, s.registry)
	if err == nil {
		c := b.NewDiscoveryServiceClient(conn)
		c.RegisterService(context.Background(), &s.registry)
	} else {
		log.Printf("Dialling discovery failed: %v", err)
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

//Dial a local server
func (s *GoServer) Dial(server string, dialler dialler, builder clientBuilder) (*grpc.ClientConn, error) {
	conn, err := dialler.Dial(registryIP+":"+strconv.Itoa(registryPort), grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	registry := builder.NewDiscoveryServiceClient(conn)
	entry := pb.RegistryEntry{Name: server}
	r, err := registry.Discover(context.Background(), &entry)
	if err != nil {
		return nil, err
	}

	return dialler.Dial(r.Ip+":"+strconv.Itoa(int(r.Port)), grpc.WithInsecure())
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
func (s *GoServer) registerServer(IP string, servername string, external bool, dialler dialler, builder clientBuilder, getter hostGetter) int32 {
	conn, err := dialler.Dial(registryIP+":"+strconv.Itoa(registryPort), grpc.WithInsecure())
	if err != nil {
		log.Printf("Could not connect: %v", err)
		return -1
	}

	registry := builder.NewDiscoveryServiceClient(conn)
	hostname, err := getter.Hostname()
	if err != nil {
		hostname = "Server-" + IP
	}
	entry := pb.RegistryEntry{Ip: IP, Name: servername, ExternalPort: external, Identifier: hostname}
	r, err := registry.RegisterService(context.Background(), &entry)
	if err != nil {
		log.Printf("Could not register this service: %v", err)
		return -1
	}
	s.registry = *r
	s.close(conn)

	log.Printf("Registered %v on port %v", servername, r.Port)
	return r.Port
}
