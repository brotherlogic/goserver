package main

import (
	"flag"
	"log"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	pb "github.com/brotherlogic/goserver/proto"
)

func main() {

	var host = flag.String("host", "", "Host")
	var port = flag.String("port", "", "Port")
	flag.Parse()

	log.Printf("Connection to %v:%v", *host, *port)
	conn, err := grpc.Dial(*host+":"+*port, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Unable to reach server %v:%v -> %v", *host, *port, err)
	}
	defer conn.Close()

	check := pb.NewGoserverServiceClient(conn)
	health, err := check.IsAlive(context.Background(), &pb.Alive{})
	log.Printf("%v and %v", health, err)

	state, _ := check.State(context.Background(), &pb.Empty{})
	for _, s := range state.GetStates() {
		log.Printf("%v and %v (%v)", s.GetKey(), time.Unix(s.GetTimeValue(), 0), s.GetValue())
	}
}
