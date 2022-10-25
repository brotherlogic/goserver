package goserver

import (
	"context"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	dspb "github.com/brotherlogic/datastore/proto"
	dpb "github.com/brotherlogic/dstore/proto"
	kspb "github.com/brotherlogic/keystore/proto"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

// NewMemoryStore build a memory store for testing
func (s *GoServer) NewMemoryStore() translatedStore {
	return &mts{store: &memstore{mem: make(map[string]*anypb.Any)}}
}

// NewFailMemoryStore fails
func (s *GoServer) NewFailMemoryStore() translatedStore {
	return &fts{}
}

type translatedStore interface {
	Load(ctx context.Context, key string, message proto.Message) error
	Save(ctx context.Context, key string, message proto.Message) error
}

type fts struct{}

func (fts *fts) Load(ctx context.Context, key string, message proto.Message) error {
	return fmt.Errorf("Built to fail")
}

func (fts *fts) Save(ctx context.Context, key string, message proto.Message) error {
	return fmt.Errorf("Built to fail")
}

type mts struct {
	store byteStore
}

func (mts *mts) Load(ctx context.Context, key string, message proto.Message) error {
	data, err := mts.store.load(ctx, key)
	if err != nil {
		return err
	}

	return proto.Unmarshal(data.GetValue(), message)
}

func (mts *mts) Save(ctx context.Context, key string, message proto.Message) error {
	bytes, err := proto.Marshal(message)
	if err != nil {
		return err
	}
	return mts.store.save(ctx, key, &anypb.Any{Value: bytes})
}

type byteStore interface {
	load(ctx context.Context, key string) (*anypb.Any, error)
	save(ctx context.Context, key string, data *anypb.Any) error
}

type memstore struct {
	mem map[string]*anypb.Any
}

func (m *memstore) load(ctx context.Context, key string) (*anypb.Any, error) {
	if val, ok := m.mem[key]; ok {
		return val, nil
	}

	return nil, status.Errorf(codes.InvalidArgument, "Not found")
}

func (m *memstore) save(ctx context.Context, key string, data *anypb.Any) error {
	m.mem[key] = data
	return nil
}

type keystore struct {
	dial func(ctx context.Context, server string) (*grpc.ClientConn, error)
}

func (k *keystore) load(ctx context.Context, key string) (*anypb.Any, error) {
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

func (k *keystore) save(ctx context.Context, key string, value *anypb.Any) error {
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

func (d *datastore) load(ctx context.Context, key string) (*anypb.Any, error) {
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

func (d *datastore) save(ctx context.Context, key string, value *anypb.Any) error {
	conn, err := d.dial(ctx, "datastore")
	if err != nil {
		return err
	}
	defer conn.Close()

	store := dspb.NewDatastoreServiceClient(conn)
	_, err = store.Write(ctx, &dspb.WriteRequest{Key: key, Value: value})
	return err
}

type dstore struct {
	dial func(ctx context.Context, server string) (*grpc.ClientConn, error)
}

func (d *dstore) load(ctx context.Context, key string) (*anypb.Any, error) {
	conn, err := d.dial(ctx, "dstore")
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	store := dpb.NewDStoreServiceClient(conn)
	resp, err := store.Read(ctx, &dpb.ReadRequest{Key: key})
	if err != nil {
		return nil, err
	}

	if resp.GetConsensus() < 0.5 {
		return nil, fmt.Errorf("Unable to get read consensus: %v", resp.GetConsensus())
	}

	return resp.GetValue(), err
}

func (d *dstore) save(ctx context.Context, key string, value *anypb.Any) error {
	conn, err := d.dial(ctx, "dstore")
	if err != nil {
		return err
	}
	defer conn.Close()

	store := dpb.NewDStoreServiceClient(conn)
	r, err := store.Write(ctx, &dpb.WriteRequest{Key: key, Value: value})

	if r.GetConsensus() < 0.5 {
		return fmt.Errorf("Unable to get write consensus: %v", r.GetConsensus())
	}
	return err
}
