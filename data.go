package goserver

import (
	"context"

	"google.golang.org/grpc"

	kspb "github.com/brotherlogic/keystore/proto"
	google_protobuf "github.com/golang/protobuf/ptypes/any"
)

type byteStore interface {
	load(ctx context.Context, key string) (*google_protobuf.Any, error)
	save(ctx context.Context, key string, data *google_protobuf.Any) error
}

type keystore struct {
	dial func(ctx context.Context, server string) (*grpc.ClientConn, error)
}

func (k *keystore) load(ctx context.Context, key string) (*google_protobuf.Any, error) {
	conn, err := k.dial(ctx, "keystore")
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	store := kspb.NewKeyStoreServiceClient(conn)
	resp, err := store.Read(ctx, &kspb.ReadRequest{Key: key})
	if err != nil {
		return nil, err
	}

	return resp.GetPayload(), err
}

func (k *keystore) save(ctx context.Context, key string, value *google_protobuf.Any) error {
	conn, err := k.dial(ctx, "keystore")
	if err != nil {
		return err
	}
	defer conn.Close()

	store := kspb.NewKeyStoreServiceClient(conn)
	_, err = store.Save(ctx, &kspb.SaveRequest{Key: key, Value: value})
	return err
}
