package utils

import (
	"context"
	"time"

	"google.golang.org/grpc"

	pbdi "github.com/brotherlogic/discovery/proto"
)

// Resolve resolves out a server
func Resolve(name string) (string, int32, error) {
	conn, err := grpc.Dial(Discover, grpc.WithInsecure())
	if err != nil {
		return "", -1, err
	}
	defer conn.Close()

	registry := pbdi.NewDiscoveryServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	val, err := registry.Discover(ctx, &pbdi.RegistryEntry{Name: name})
	if err != nil {
		return "", -1, err
	}
	return val.GetIp(), val.GetPort(), err
}
