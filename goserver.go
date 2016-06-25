package goserver

import (
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"log"
	"net"
	"strconv"

	pb "github.com/brotherlogic/discovery/proto"
	pbd "github.com/brotherlogic/monitor/proto"
)

const (
	registryIP   = "10.0.1.17"
	registryPort = 50055
)

// GoServer The basic server construct
type GoServer struct {
	servername string
	port       int32
	registry   pb.RegistryEntry
}

// Register Registers grpc endpoints
func (s *GoServer) Register(server *grpc.Server) {
	// To be extended by other classes
}

type dialler interface {
	Dial(host string, opts ...grpc.DialOption) (*grpc.ClientConn, error)
}

type clientBuilder interface {
	NewDiscoveryServiceClient(conn *grpc.ClientConn) pb.DiscoveryServiceClient
}

func (s *GoServer) sendHeartbeat(monitorIP string, monitorPort int, dialler dialler, builder clientBuilder) {
	conn, _ := dialler.Dial(monitorIP+":"+strconv.Itoa(registryPort), grpc.WithInsecure())
	defer conn.Close()
	monitor := pbd.NewMonitorServiceClient(conn)
	monitor.ReceiveHeartbeat(context.Background(), &s.registry)
}

func getLocalIP() string {
	ifaces, _ := net.Interfaces()

	var ip net.IP
	for _, i := range ifaces {
		addrs, _ := i.Addrs()

		for _, addr := range addrs {
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			}
		}
	}

	return ip.String()
}

// RegisterServer Registers a server with the system and gets the port number it should use
func (s *GoServer) registerServer(IP string, servername string, external bool, dialler dialler, builder clientBuilder) int32 {
	conn, err := dialler.Dial(registryIP+":"+strconv.Itoa(registryPort), grpc.WithInsecure())
	if err != nil {
		log.Printf("Could not connect: %v", err)
		return -1
	}

	defer conn.Close()
	registry := builder.NewDiscoveryServiceClient(conn)

	entry := pb.RegistryEntry{Ip: IP, Name: servername, ExternalPort: external}
	r, err := registry.RegisterService(context.Background(), &entry)
	if err != nil {
		log.Printf("Could not register service: %v", err)
		return -1
	}

	return r.Port
}
