package goserver

import (
	"context"
	"testing"

	pb "github.com/brotherlogic/goserver/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

func TestBasicStore(t *testing.T) {
	mts := &mts{store: &memstore{mem: make(map[string]*anypb.Any)}}

	test := &pb.Alive{Name: "testing"}

	err := mts.Save(context.Background(), "test", test)
	if err != nil {
		t.Fatalf("Unable to save: %v", err)
	}

	t2 := &pb.Alive{}
	err = mts.Load(context.Background(), "test", t2)
	if err != nil {
		t.Fatalf("Unable to load: %v", err)
	}

	if t2.GetName() != "testing" {
		t.Errorf("Translation has failed: %v", t2)
	}
}
