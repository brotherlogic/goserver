package goserver

import (
	"context"
	"fmt"

	"google.golang.org/grpc"

	dspb "github.com/brotherlogic/datastore/proto"
	kspb "github.com/brotherlogic/keystore/proto"
	"github.com/golang/protobuf/proto"
	google_protobuf "github.com/golang/protobuf/ptypes/any"
)

type translatedStore interface {
	load(ctx context.Context, key string, message proto.Message) error
	save(ctx context.Context, key string, message proto.Message) error
}

type mts struct {
	store byteStore
}

func (mts *mts) load(ctx context.Context, key string, message proto.Message) error {
	data, err := mts.store.load(ctx, key)
	if err != nil {
		return err
	}

	return proto.Unmarshal(data.GetValue(), message)
}

func (mts *mts) save(ctx context.Context, key string, message proto.Message) error {
	bytes, err := proto.Marshal(message)
	if err != nil {
		return err
	}
	return mts.store.save(ctx, key, &google_protobuf.Any{Value: bytes})
}

type byteStore interface {
	load(ctx context.Context, key string) (*google_protobuf.Any, error)
	save(ctx context.Context, key string, data *google_protobuf.Any) error
}

type memstore struct {
	mem map[string]*google_protobuf.Any
}

func (m *memstore) load(ctx context.Context, key string) (*google_protobuf.Any, error) {
	if val, ok := m.mem[key]; ok {
		return val, nil
	}

	return nil, fmt.Errorf("Not found")
}

func (m *memstore) save(ctx context.Context, key string, data *google_protobuf.Any) error {
	m.mem[key] = data
	return nil
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

	store := dspb.NewDatastoreServiceClient(conn)
	_, err = store.Write(ctx, &dspb.WriteRequest{Key: key, Value: value})
	return err
}

type datastore struct {
	dial func(ctx context.Context, server string) (*grpc.ClientConn, error)
}

func (d *datastore) load(ctx context.Context, key string) (*google_protobuf.Any, error) {
	conn, err := d.dial(ctx, "datastore")
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	store := dspb.NewDatastoreServiceClient(conn)
	resp, err := store.Read(ctx, &dspb.ReadRequest{Key: key})
	if err != nil {
		return nil, err
	}

	return resp.GetValue(), err
}

func (d *datastore) save(ctx context.Context, key string, value *google_protobuf.Any) error {
	conn, err := d.dial(ctx, "datastore")
	if err != nil {
		return err
	}
	defer conn.Close()

	store := kspb.NewKeyStoreServiceClient(conn)
	_, err = store.Save(ctx, &kspb.SaveRequest{Key: key, Value: value})
	return err
}
