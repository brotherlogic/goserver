package utils

import (
	"fmt"
	"math/rand"
	"reflect"
	"time"

	"github.com/golang/protobuf/proto"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	pbdi "github.com/brotherlogic/discovery/proto"
)

// BuildContext builds a context object for use
func BuildContext(label, origin string) (context.Context, context.CancelFunc) {
	con, can := generateContext(origin)
	return con, can
}

func generateContext(origin string) (context.Context, context.CancelFunc) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	tracev := fmt.Sprintf("%v-%v-%v", origin, time.Now().Unix(), r.Int63())
	mContext := metadata.AppendToOutgoingContext(context.Background(), "trace-id", tracev)
	return context.WithTimeout(mContext, time.Hour)
}

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
	case reflect.Float32, reflect.Float64:
		if in.Float() != 0 && in.Float() != out.Float() {
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

// GetMaster resolves out a server
func GetMaster(name string) (*pbdi.RegistryEntry, error) {
	conn, err := grpc.Dial(Discover, grpc.WithInsecure())
	if err != nil {
		return &pbdi.RegistryEntry{}, err
	}
	defer conn.Close()

	registry := pbdi.NewDiscoveryServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	val, err := registry.Discover(ctx, &pbdi.DiscoverRequest{Request: &pbdi.RegistryEntry{Name: name}})
	return val.GetService(), err
}

// ResolveAll gets all servers
func ResolveAll(name string) ([]*pbdi.RegistryEntry, error) {
	entries := make([]*pbdi.RegistryEntry, 0)
	conn, err := grpc.Dial(Discover, grpc.WithInsecure())
	if err != nil {
		return entries, err
	}
	defer conn.Close()

	registry := pbdi.NewDiscoveryServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	val, err := registry.ListAllServices(ctx, &pbdi.ListRequest{})
	if err != nil {
		return entries, err
	}
	for _, entry := range val.GetServices().Services {
		if len(name) == 0 || entry.Name == name {
			entries = append(entries, entry)
		}
	}

	return entries, nil
}
