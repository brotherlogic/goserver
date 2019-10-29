package main

import (
	"flag"
	"fmt"
	"log"
	"strconv"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	pb "github.com/brotherlogic/goserver/proto"
	"github.com/brotherlogic/goserver/utils"
	"github.com/golang/protobuf/proto"
	_ "google.golang.org/grpc/encoding/gzip"
)

func buildState(s *pb.State, uptime int64) string {
	if len(s.Text) > 0 {
		return fmt.Sprintf("%v", s.Text)
	}

	if s.Value > 0 {
		return fmt.Sprintf("%v (%2.2f)", s.Value, float64(s.Value)/float64(uptime))
	}

	if s.Fraction > 0 {
		return fmt.Sprintf("%2.2f", s.Fraction)
	}

	if s.TimeValue > 0 {
		return fmt.Sprintf("%v", time.Unix(s.TimeValue, 0))
	}

	if s.TimeDuration > 0 {
		return fmt.Sprintf("%v", time.Duration(s.TimeDuration).Round(time.Millisecond))
	}

	return ""
}

func getUptime(states []*pb.State) int64 {
	for _, state := range states {
		if state.GetKey() == "uptime" {
			return int64(time.Duration(state.TimeDuration).Seconds())
		}
	}
	return 1
}

func main() {

	var host = flag.String("host", "", "Host")
	var port = flag.String("port", "", "Port")
	var name = flag.String("name", "", "Name")
	var server = flag.String("server", "", "Server")
	var all = flag.String("all", "", "All")
	var action = flag.String("action", "", "Action")
	flag.Parse()

	if len(*host) > 0 {
		conn, err := grpc.Dial(*host+":"+*port, grpc.WithInsecure())
		if err != nil {
			log.Fatalf("Unable to reach server %v:%v -> %v", *host, *port, err)
		}
		defer conn.Close()

		check := pb.NewGoserverServiceClient(conn)

		state, err := check.State(context.Background(), &pb.Empty{})
		if err != nil {
			log.Fatalf("Unable to read state: %v", err)
		}

		uptime := getUptime(state.GetStates())
		for _, st := range state.GetStates() {
			state := buildState(st, uptime)
			if len(state) > 0 {
				fmt.Printf("%v: %v -> %v\n", *host, st.GetKey(), state)
			}

		}
	} else if len(*name) > 0 {
		ip, port, err := utils.Resolve(*name, "goserver-cliname")
		conn, err := grpc.Dial(ip+":"+strconv.Itoa(int(port)), grpc.WithInsecure())
		if err != nil {
			log.Fatalf("Unable to reach server %v:%v -> %v", ip, port, err)
		}
		defer conn.Close()

		check := pb.NewGoserverServiceClient(conn)

		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()
		state, _ := check.State(ctx, &pb.Empty{})
		fmt.Printf("DIAL: %v - %v [%v]\n", ip, port, proto.Size(state))
		uptime := getUptime(state.GetStates())
		for _, st := range state.GetStates() {
			state := buildState(st, uptime)
			if len(state) > 0 {
				fmt.Printf("%v: %v -> %v\n", ip, st.GetKey(), state)
			}
		}
	} else {
		servers, err := utils.ResolveAll(*all)
		if err != nil {
			log.Fatalf("Error: %v", err)
		}
		for _, s := range servers {
			if !s.IgnoresMaster && (*server == "" || *server == s.Identifier) {

				fmt.Printf("SERVER: %v\n", s)
				conn, err := grpc.Dial(s.Ip+":"+strconv.Itoa(int(s.Port)), grpc.WithInsecure())
				if err != nil {
					log.Fatalf("Unable to reach server %v", s)
				}
				defer conn.Close()

				check := pb.NewGoserverServiceClient(conn)

				if len(*action) > 0 {
					val, err := check.Shutdown(context.Background(), &pb.ShutdownRequest{})
					fmt.Printf("Shutdown: %v, %v", val, err)
					return
				}

				state, err := check.State(context.Background(), &pb.Empty{})
				if err != nil {
					fmt.Printf("Error reading state: %v\n", err)
				}
				uptime := getUptime(state.GetStates())
				for _, st := range state.GetStates() {
					state := buildState(st, uptime)
					if len(state) > 0 {
						prechar := ""
						if s.Master {
							prechar = "*"
						}
						fmt.Printf("%v%v (%v): %v -> %v\n", prechar, s.Identifier, s.Name, st.GetKey(), state)
					}

				}
			}
		}
	}
}
