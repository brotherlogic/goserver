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
	_ "google.golang.org/grpc/encoding/gzip"
)

func buildState(s *pb.State) string {
	if len(s.Text) > 0 {
		return fmt.Sprintf("%v", s.Text)
	}

	if s.Value > 0 {
		return fmt.Sprintf("%v", s.Value)
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

func main() {

	var host = flag.String("host", "", "Host")
	var port = flag.String("port", "", "Port")
	var name = flag.String("name", "", "Name")
	var server = flag.String("server", "", "Server")
	var all = flag.String("all", "", "All")
	flag.Parse()

	if len(*host) > 0 {
		conn, err := grpc.Dial(*host+":"+*port, grpc.WithInsecure())
		if err != nil {
			log.Fatalf("Unable to reach server %v:%v -> %v", *host, *port, err)
		}
		defer conn.Close()

		check := pb.NewGoserverServiceClient(conn)
		health, err := check.IsAlive(context.Background(), &pb.Alive{})
		fmt.Printf("%v and %v\n", health, err)

		state, _ := check.State(context.Background(), &pb.Empty{})
		for _, s := range state.GetStates() {
			fmt.Printf("%v and %v (%v, %v) with %v\n", s.GetKey(), time.Unix(s.GetTimeValue(), 0), s.GetValue(), s.GetFraction(), s.GetText())
		}
	} else if len(*name) > 0 {
		ip, port, err := utils.Resolve(*name)
		fmt.Printf("DIAL: %v\n", ip)
		conn, err := grpc.Dial(ip+":"+strconv.Itoa(int(port)), grpc.WithInsecure())
		if err != nil {
			log.Fatalf("Unable to reach server %v:%v -> %v", ip, port, err)
		}
		defer conn.Close()

		check := pb.NewGoserverServiceClient(conn)
		health, err := check.IsAlive(context.Background(), &pb.Alive{})
		fmt.Printf("%v and %v\n", health, err)

		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()
		state, _ := check.State(ctx, &pb.Empty{})
		for _, st := range state.GetStates() {
			state := buildState(st)
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
				health, err := check.IsAlive(context.Background(), &pb.Alive{})
				fmt.Printf("%v and %v\n", health, err)

				state, _ := check.State(context.Background(), &pb.Empty{})
				for _, st := range state.GetStates() {
					state := buildState(st)
					if len(state) > 0 {

						fmt.Printf("%v (%v): %v -> %v\n", s.Identifier, s.Name, st.GetKey(), state)
					}

				}
			}
		}
	}
}
