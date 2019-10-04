package utils

import (
	"math"
	"testing"
	"time"

	pbgd "github.com/brotherlogic/godiscogs"
	pb "github.com/brotherlogic/goserver/proto"
	pbrc "github.com/brotherlogic/recordcollection/proto"
	"google.golang.org/grpc/metadata"
)

func BenchmarkBuildContext(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, cancel := generateContext("blah", time.Minute)
		cancel()
	}
}

func TestEmbed(t *testing.T) {
	a := &pb.EmbeddedTest{Blah: &pb.Alive{Name: "blah"}}
	if FuzzyMatch(a, a) != nil {
		t.Errorf("Failure to match on embedded: %v", FuzzyMatch(a, a))
	}
}

func TestFuzzyMatchDetailed(t *testing.T) {
	a := &pbrc.Record{Release: &pbgd.Release{Id: 1234, FolderId: 1, Title: "Bonkers"}}
	b := &pbrc.Record{Release: &pbgd.Release{FolderId: 1}}

	if FuzzyMatch(b, a) != nil {
		t.Errorf("Failed to match on detailed: %v != %v -> %v", b, a, FuzzyMatch(b, a))
	}
}

func TestFuzzyMatchWithNaN(t *testing.T) {
	a := &pbrc.Record{Metadata: &pbrc.ReleaseMetadata{OverallScore: float32(math.NaN())}}
	b := &pbrc.Record{Metadata: &pbrc.ReleaseMetadata{OverallScore: float32(math.NaN())}}

	if FuzzyMatch(b, a) != nil {
		t.Errorf("Failed to match on detailed: %v != %v -> %v", b, a, FuzzyMatch(b, a))
	}
}

func TestFuzzyMatchDetailedEmpty(t *testing.T) {
	a := &pbrc.Record{Release: &pbgd.Release{Id: 1234, FolderId: 1, Title: "Bonkers"}, Metadata: &pbrc.ReleaseMetadata{}}
	b := &pbrc.Record{Metadata: &pbrc.ReleaseMetadata{}}

	if FuzzyMatch(b, a) != nil {
		t.Errorf("Failed to match on detailed: %v != %v", b, a)
	}
}

func TestFuzzyMatch(t *testing.T) {
	a := &pb.State{Key: "blah", Value: 1233}
	var testData = []struct {
		b     *pb.State
		match bool
	}{
		{&pb.State{Key: "blah"}, true},
		{&pb.State{Value: 1233}, true},
		{&pb.State{Value: 1233, Key: "blah"}, true},
		{&pb.State{Key: "donkey"}, false},
		{&pb.State{Value: 1234}, false},
		{&pb.State{Key: "blah", Value: 1234}, false},
		{&pb.State{Key: "donkey", Value: 1233}, false},
	}

	for _, tt := range testData {
		actual := FuzzyMatch(tt.b, a)
		if (tt.match && actual != nil) || (!tt.match && actual == nil) {
			t.Errorf("Failure in match %v vs %v", tt.b, a)
		}
	}
}

func TestGetContext(t *testing.T) {
	ctx, cancel := BuildContext("TestGetContext", "testing")
	defer cancel()

	md, found := metadata.FromOutgoingContext(ctx)
	if !found {
		t.Fatalf("No context: %v", md)
	}
}
