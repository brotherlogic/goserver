package utils

import (
	"context"
	"reflect"
	"time"

	"github.com/golang/protobuf/proto"
	"google.golang.org/grpc"

	pbdi "github.com/brotherlogic/discovery/proto"
)

//FuzzyMatch experimental fuzzy match
func FuzzyMatch(matcher, matchee proto.Message) bool {
	in := reflect.ValueOf(matcher)
	out := reflect.ValueOf(matchee)

	for i := 0; i < in.Elem().NumField(); i++ {
		switch in.Elem().Field(i).Kind() {
		case reflect.Int32, reflect.Int64, reflect.Uint32, reflect.Uint64:
			if in.Elem().Field(i).Int() != 0 && in.Elem().Field(i).Int() != out.Elem().Field(i).Int() {
				return false
			}
		case reflect.Bool:
			return false
		case reflect.String:
			if in.Elem().Field(i).String() != "" && in.Elem().Field(i).String() != out.Elem().Field(i).String() {
				return false
			}
		default:
			return false
		}
	}
	return true
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
