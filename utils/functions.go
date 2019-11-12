package utils

import (
	"fmt"
	"math"
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
	con, can := generateContext(origin, time.Hour)
	return con, can
}

// ManualContext builds a context object for use
func ManualContext(label, origin string, t time.Duration) (context.Context, context.CancelFunc) {
	con, can := generateContext(origin, t)
	return con, can
}

func generateContext(origin string, t time.Duration) (context.Context, context.CancelFunc) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	tracev := fmt.Sprintf("%v-%v-%v", origin, time.Now().Unix(), r.Int63())
	mContext := metadata.AppendToOutgoingContext(context.Background(), "trace-id", tracev)
	return context.WithTimeout(mContext, t)
}

//FuzzyMatch experimental fuzzy match
func FuzzyMatch(matcher, matchee proto.Message) error {
	in := reflect.ValueOf(matcher)
	out := reflect.ValueOf(matchee)

	return matchStruct(in.Elem(), out.Elem())
}

func matchStruct(in, out reflect.Value) error {
	for i := 0; i < in.NumField(); i++ {
		err := doMatch(in.Field(i), out.Field(i))
		if err != nil {
			return fmt.Errorf("Bad match in field %v: %v", i, err)
		}
	}
	return nil
}

func doMatch(in, out reflect.Value) error {
	switch in.Kind() {
	case reflect.Int32, reflect.Int64, reflect.Uint32, reflect.Uint64:
		if in.Int() != 0 && in.Int() != out.Int() {
			return fmt.Errorf("Mismatch in ints %v vs %v", in.Int(), out.Int())
		}
	case reflect.Float32, reflect.Float64:
		if in.Float() != 0 && in.Float() != out.Float() && !math.IsNaN(in.Float()) && !math.IsNaN(out.Float()) {
			return fmt.Errorf("Mismatch in floats %v vs %v", in.Float(), out.Float())
		}
	case reflect.Bool:
		if !in.Bool() || out.Bool() {
			return nil
		}
		return fmt.Errorf("Mimstatch in bools")
	case reflect.String:
		if in.String() != "" && in.String() != out.String() {
			return fmt.Errorf("Mismatch in strings %v vs %v", in.String(), out.String())
		}
	case reflect.Ptr:
		if in.IsNil() {
			return nil
		}
		return doMatch(in.Elem(), out.Elem())
	case reflect.Struct:
		return matchStruct(in, out)
	case reflect.Slice:
		// We ignore slices for now
		return nil
	default:
		return fmt.Errorf("Error in parsing fuzzy match: %v -> %v\n", in.Kind(), out)
	}

	return nil
}

// Resolve resolves out a server
func Resolve(name, origin string) (string, int32, error) {
	if name == "discover" {
		return RegistryIP, int32(RegistryPort), nil
	}
	conn, err := grpc.Dial(Discover, grpc.WithInsecure())
	if err != nil {
		return "", -1, err
	}
	defer conn.Close()

	registry := pbdi.NewDiscoveryServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	val, err := registry.Discover(ctx, &pbdi.DiscoverRequest{Caller: origin, Request: &pbdi.RegistryEntry{Name: name}})
	if err != nil {
		return "", -1, err
	}
	return val.GetService().GetIp(), val.GetService().GetPort(), err
}

// Resolve resolves out a server
func ResolveV2(name string) (*pbdi.RegistryEntry, error) {
	if name == "discover" {
		return &pbdi.RegistryEntry{Ip: RegistryIP, Port: int32(RegistryPort)}, nil
	}

	conn, err := grpc.Dial(Discover, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	registry := pbdi.NewDiscoveryServiceV2Client(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	val, err := registry.Get(ctx, &pbdi.GetRequest{Job: name})
	if err != nil {
		return nil, err
	}

	if len(val.GetServices()) != 1 {
		return nil, fmt.Errorf("Found %v services", len(val.GetServices()))
	}

	return val.GetServices()[0], err
}

// GetMaster resolves out a server
func GetMaster(name, caller string) (*pbdi.RegistryEntry, error) {
	conn, err := grpc.Dial(Discover, grpc.WithInsecure())
	if err != nil {
		return &pbdi.RegistryEntry{}, err
	}
	defer conn.Close()

	registry := pbdi.NewDiscoveryServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	val, err := registry.Discover(ctx, &pbdi.DiscoverRequest{Caller: caller, Request: &pbdi.RegistryEntry{Name: name}})
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
