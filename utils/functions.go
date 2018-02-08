package utils

import (
	"context"
	"fmt"
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

	return matchStruct(in.Elem(), out.Elem())
}

func matchStruct(in, out reflect.Value) bool {
	for i := 0; i < in.NumField(); i++ {
		if !(doMatch(in.Field(i), out.Field(i))) {
			return false
		}
	}
	return true
}

func doMatch(in, out reflect.Value) bool {
	switch in.Kind() {
	case reflect.Int32, reflect.Int64, reflect.Uint32, reflect.Uint64:
		if in.Int() != 0 && in.Int() != out.Int() {
			return false
		}
	case reflect.Bool:
		if !in.Bool() || out.Bool() {
			return true
		}
		return false
	case reflect.String:
		if in.String() != "" && in.String() != out.String() {
			return false
		}
	case reflect.Ptr:
		if in.IsNil() {
			return true
		}
		return doMatch(in.Elem(), out.Elem())
	case reflect.Struct:
		return matchStruct(in, out)
	case reflect.Slice:
		// We ignore slices for now
		return true
	default:
		fmt.Printf("Error in parsing fuzzy match: %v -> %v\n", in.Kind(), out)
		return false
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
	val, err := registry.Discover(ctx, &pbdi.DiscoverRequest{Request: &pbdi.RegistryEntry{Name: name}})
	if err != nil {
		return "", -1, err
	}
	return val.GetService().GetIp(), val.GetService().GetPort(), err
}
