package utils

import (
	"context"
	"time"

	"github.com/golang/protobuf/proto"
	"google.golang.org/grpc"

	pbdi "github.com/brotherlogic/discovery/proto"
)

//FuzzyMatch is a loose matcher for two protobufs
func FuzzyMatch(matcher, matchee proto.Message) bool {

	matcher1 := proto.Clone(matcher)
	matcher2 := proto.Clone(matcher)
	matchee1 := proto.Clone(matchee)
	matchee2 := proto.Clone(matchee)

	proto.Merge(matcher1, matchee1)
	proto.Merge(matchee2, matcher2)

	return proto.Equal(matcher1, matchee2)
}

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
