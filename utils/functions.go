package utils

import (
	"fmt"
	"math"
	"math/rand"
	"os"
	"reflect"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"

	pbdi "github.com/brotherlogic/discovery/proto"
)

// BuildContext builds a context object for use
func BuildContext(label, origin string) (context.Context, context.CancelFunc) {
	con, can := generateContext(origin, time.Hour)
	return con, can
}

// ManualContext builds a context object for use
func ManualContext(label string, t time.Duration) (context.Context, context.CancelFunc) {
	con, can := generateContext(label, t)
	return con, can
}

func generateContext(origin string, t time.Duration) (context.Context, context.CancelFunc) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	hostname, err := os.Hostname()
	if err != nil {
		hostname = fmt.Sprintf("BadHostGet-%v", err)
	}
	tracev := fmt.Sprintf("%v-%v-%v-%v", origin, time.Now().Unix(), r.Int63(), hostname)
	mContext := metadata.AppendToOutgoingContext(context.Background(), "trace-id", tracev)
	return context.WithTimeout(mContext, t)
}

func RefreshContext(ctx context.Context, key string, t time.Duration) (context.Context, context.CancelFunc) {
	key, err := GetContextKey(ctx)
	if err != nil {
		return generateContext("context-failed", t)
	}
	mContext := metadata.AppendToOutgoingContext(context.Background(), "trace-id", key)
	return context.WithTimeout(mContext, t)
}

func GetContextKey(ctx context.Context) (string, error) {
	md, found := metadata.FromIncomingContext(ctx)
	if found {
		if _, ok := md["trace-id"]; ok {
			idt := md["trace-id"][0]

			if idt != "" {
				return idt, nil
			}
		}
	}

	md, found = metadata.FromOutgoingContext(ctx)
	if found {
		if _, ok := md["trace-id"]; ok {
			idt := md["trace-id"][0]

			if idt != "" {
				return idt, nil
			}
		}
	}

	return "", status.Errorf(codes.NotFound, "Could not extract trace-id from incoming or outgoing")
}

// FuzzyMatch experimental fuzzy match
func FuzzyMatch(matcher, matchee proto.Message) error {
	in := reflect.ValueOf(matcher)
	out := reflect.ValueOf(matchee)

	return matchStruct(in.Elem(), out.Elem())
}

func matchStruct(in, out reflect.Value) error {
	if in.Kind() != reflect.Struct || out.Kind() != reflect.Struct {
		return fmt.Errorf("Bad field: (%v) %v and (%v) %v", in.Kind(), in, out.Kind(), out)
	}
	if in.NumField() != out.NumField() {
		return fmt.Errorf("Field numbers do not match : %v and %v", in.NumField(), out.NumField())
	}
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
	case reflect.Array:
		if in.Len() != out.Len() {
			return fmt.Errorf("Mismatch in array length")
		}
		// Match empty arrays
		if in.Len() == 0 {
			return nil
		}
		return fmt.Errorf("Can't handle arrays yet %v vs %v", in, out)
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
	conn, err := grpc.Dial("192.168.86.49:50055", grpc.WithInsecure())
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
		return &pbdi.RegistryEntry{Ip: LocalIP, Port: int32(RegistryPort)}, nil
	}

	conn, err := grpc.Dial(LocalDiscover, grpc.WithInsecure())
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

	return nil, status.Errorf(codes.NotFound, "Found %v services but no master", len(val.GetServices()))
}

// ResolveV3 resolves out a server
func ResolveV3(name string) ([]*pbdi.RegistryEntry, error) {
	if name == "discover" {
		return []*pbdi.RegistryEntry{&pbdi.RegistryEntry{Ip: LocalIP, Port: int32(RegistryPort)}}, nil
	}

	conn, err := grpc.Dial(LocalDiscover, grpc.WithInsecure())
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

	services := make([]*pbdi.RegistryEntry, 0)
	// Get master
	for _, service := range val.GetServices() {
		services = append(services, service)
	}

	return services, nil
}

// ResolveV3 resolves out a server
func ResolveV3Client(name string) ([]*pbdi.RegistryEntry, error) {
	if name == "discover" {
		return []*pbdi.RegistryEntry{&pbdi.RegistryEntry{Ip: LocalIP, Port: int32(RegistryPort)}}, nil
	}

	conn, err := grpc.Dial("192.168.86.49:50055", grpc.WithInsecure())
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

	services := make([]*pbdi.RegistryEntry, 0)
	// Get master
	for _, service := range val.GetServices() {
		services = append(services, service)
	}

	return services, nil
}

// GetMaster resolves out a server
func GetMaster(name, caller string) (*pbdi.RegistryEntry, error) {
	conn, err := grpc.Dial("192.168.86.49:50055", grpc.WithInsecure())
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
	conn, err := grpc.Dial(fmt.Sprintf("127.0.0.1:%v", RegistryPort), grpc.WithInsecure())
	if err != nil {
		return entries, err
	}
	defer conn.Close()

	registry := pbdi.NewDiscoveryServiceV2Client(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	val, err := registry.Get(ctx, &pbdi.GetRequest{Job: name})
	if err != nil {
		return entries, err
	}
	for _, entry := range val.GetServices() {
		if len(name) == 0 || entry.Name == name {
			entries = append(entries, entry)
		}
	}

	return entries, nil
}

// ResolveAll gets all servers
func BaseResolveAll(name string) ([]*pbdi.RegistryEntry, error) {
	entries := make([]*pbdi.RegistryEntry, 0)
	conn, err := grpc.Dial(Discover, grpc.WithInsecure())
	if err != nil {
		return entries, err
	}
	defer conn.Close()

	registry := pbdi.NewDiscoveryServiceV2Client(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	val, err := registry.Get(ctx, &pbdi.GetRequest{Job: name})
	if err != nil {
		return entries, err
	}
	for _, entry := range val.GetServices() {
		if len(name) == 0 || entry.Name == name {
			entries = append(entries, entry)
		}
	}

	return entries, nil
}

// LFFind finds all servers
func LFFind(ctx context.Context, servername string) ([]string, error) {
	if servername == "discover" {
		return []string{}, fmt.Errorf("Cannot multi dial discovery")
	}

	conn, err := LFDial(Discover)
	if err != nil {
		return []string{}, err
	}
	defer conn.Close()

	registry := pbdi.NewDiscoveryServiceV2Client(conn)
	val, err := registry.Get(ctx, &pbdi.GetRequest{Job: servername})
	if err != nil {
		return []string{}, err
	}

	// Pick a server at random
	ret := []string{}
	for _, entry := range val.GetServices() {
		ret = append(ret, fmt.Sprintf("%v:%v", entry.Identifier, entry.Port))
	}
	return ret, nil
}

// LFDial fundamental dial
func LFDial(host string) (*grpc.ClientConn, error) {
	return grpc.Dial(host,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(otelgrpc.UnaryClientInterceptor()))
}

// LFDialServer dial a specific job
func LFDialServer(ctx context.Context, servername string) (*grpc.ClientConn, error) {
	conn, err := LFDial(Discover)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	registry := pbdi.NewDiscoveryServiceV2Client(conn)
	val, err := registry.Get(ctx, &pbdi.GetRequest{Job: servername})
	if err != nil {
		return nil, err
	}

	// Pick a server at random
	rand.Seed(time.Now().UnixNano())
	servernum := rand.Intn(len(val.GetServices()))
	return LFDial(fmt.Sprintf("%v:%v", val.GetServices()[servernum].GetIp(), val.GetServices()[servernum].GetPort()))
}

// LFDialServer dial a specific job
func LFDialSpecificServer(ctx context.Context, servername, host string) (*grpc.ClientConn, error) {
	conn, err := LFDial(Discover)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	registry := pbdi.NewDiscoveryServiceV2Client(conn)
	val, err := registry.Get(ctx, &pbdi.GetRequest{Job: servername})
	if err != nil {
		return nil, err
	}

	// Pick a server at random
	rand.Seed(time.Now().UnixNano())
	for _, server := range val.GetServices() {
		if server.GetIdentifier() == host {
			return LFDial(fmt.Sprintf("%v:%v", server.GetIp(), server.GetPort()))
		}
	}

	return nil, fmt.Errorf("Unable to locate %v on %v", servername, host)
}
