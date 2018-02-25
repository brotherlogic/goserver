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

func main() {

	var host = flag.String("host", "", "Host")
	var port = flag.String("port", "", "Port")
	var name = flag.String("name", "", "Name")
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
			fmt.Printf("%v and %v (%v) with %v\n", s.GetKey(), time.Unix(s.GetTimeValue(), 0), s.GetValue(), s.GetText())
		}
	} else {
		ip, port, _ := utils.Resolve(*name)
		fmt.Printf("DIAL: %v\n", ip)
		conn, err := grpc.Dial(ip+":"+strconv.Itoa(int(port)), grpc.WithInsecure())
		if err != nil {
			log.Fatalf("Unable to reach server %v:%v -> %v", ip, port, err)
		}
		defer conn.Close()

		check := pb.NewGoserverServiceClient(conn)
		health, err := check.IsAlive(context.Background(), &pb.Alive{})
		fmt.Printf("%v and %v\n", health, err)

		state, _ := check.State(context.Background(), &pb.Empty{})
		for _, s := range state.GetStates() {
			fmt.Printf("%v and %v (%v) with %v\n", s.GetKey(), time.Unix(s.GetTimeValue(), 0), s.GetValue(), s.GetText())
		}
	}
}
