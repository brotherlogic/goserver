package utils

import (
	"errors"
	"fmt"
	"math/rand"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/golang/protobuf/proto"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	pbdi "github.com/brotherlogic/discovery/proto"
	pb "github.com/brotherlogic/goserver/proto"
	pbt "github.com/brotherlogic/tracer/proto"
)

//SendTrace sends out a tracer trace
func SendTrace(c context.Context, l string, t time.Time, ty pbt.Milestone_MilestoneType, o string) error {
	md, found := metadata.FromIncomingContext(c)
	if found {
		idt := md["trace-id"][0]

		if idt != "" {
			id := idt
			if strings.HasPrefix(id, "test") {
				return errors.New("Test trace")
			}
			traceIP, tracePort, _ := Resolve("tracer")
			if tracePort > 0 {
				conn, err := grpc.Dial(traceIP+":"+strconv.Itoa(int(tracePort)), grpc.WithInsecure())
				defer conn.Close()
				if err == nil {
					ctx, cancel := context.WithTimeout(context.Background(), time.Second)
					defer cancel()
					client := pbt.NewTracerServiceClient(conn)

					m := &pbt.Milestone{Label: l, Timestamp: time.Now().UnixNano(), Origin: o, Type: ty}
					p := &pbt.ContextProperties{Id: id}
					if ty == pbt.Milestone_START {
						p.Label = l
					}
					_, err := client.Record(ctx, &pbt.RecordRequest{Milestone: m, Properties: p})
					return err
				}
			}
		}

		return fmt.Errorf("Unable to trace - maybe because of %v", md)
	}
	return fmt.Errorf("Unable to trace - context: %v", c)
}

// BuildContext builds a context object for use
func BuildContext(origin string, t pb.ContextType) (context.Context, context.CancelFunc) {
	con, can := generateContext(origin, t)
	SendTrace(con, "Generate", time.Now(), pbt.Milestone_START, origin)
	return con, can
}

func generateContext(origin string, t pb.ContextType) (context.Context, context.CancelFunc) {
	tracev := fmt.Sprintf("%v-%v-%v", origin, time.Now().Unix(), rand.Int63())
	baseContext := context.WithValue(context.Background(), "trace-id", tracev)
	mContext := metadata.NewOutgoingContext(baseContext, metadata.Pairs("trace-id", tracev))
	if t == pb.ContextType_REGULAR {
		return context.WithTimeout(mContext, time.Second)
	}

	if t == pb.ContextType_LONG {
		return context.WithTimeout(mContext, time.Hour)
	}

	return mContext, func() {}
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
