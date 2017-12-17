package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	pb "github.com/brotherlogic/goserver/proto"
_ "google.golang.org/grpc/encoding/gzip"
)

func main() {

	var host = flag.String("host", "", "Host")
	var port = flag.String("port", "", "Port")
	flag.Parse()

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
}
