package goserver

import (
	"context"
	"fmt"

	"google.golang.org/protobuf/types/known/anypb"

	dspb "github.com/brotherlogic/dstore/proto"
)

func (s *GoServer) LoadData(ctx context.Context, key string, consensus float32) ([]byte, error) {
	conn, err := s.FDialServer(ctx, "dstore")
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	client := dspb.NewDStoreServiceClient(conn)
	res, err := client.Read(ctx, &dspb.ReadRequest{Key: key})
	if err != nil {
		return nil, err
	}

	if res.GetConsensus() < consensus {
		return nil, fmt.Errorf("could not get read consensus (%v)", res.GetConsensus())
	}

	return res.GetValue().GetValue(), nil
}

func (s *GoServer) SaveData(ctx context.Context, data []byte, key string, consensus float32) error {
	conn, err := s.FDialServer(ctx, "dstore")
	if err != nil {
		return err
	}
	defer conn.Close()

	client := dspb.NewDStoreServiceClient(conn)
	res, err := client.Write(ctx, &dspb.WriteRequest{Key: key, Value: &anypb.Any{Value: data}})
	if err != nil {
		return err
	}

	if res.GetConsensus() < consensus {
		return fmt.Errorf("could not get write consensus (%v)", res.GetConsensus())
	}

	return nil
}
