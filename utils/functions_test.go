package utils

import (
	"log"
	"testing"

	pbgd "github.com/brotherlogic/godiscogs"
	pb "github.com/brotherlogic/goserver/proto"
	pbrc "github.com/brotherlogic/recordcollection/proto"
)

func TestEmbed(t *testing.T) {
	a := &pb.EmbeddedTest{Blah: &pb.Alive{Name: "blah"}}
	if !FuzzyMatch(a, a) {
		t.Errorf("Failure to match on embedded: %v", a)
	}
}

func TestFuzzyMatchDetailed(t *testing.T) {
	a := &pbrc.Record{Release: &pbgd.Release{Id: 1234, FolderId: 1, Title: "Bonkers"}}
	b := &pbrc.Record{Release: &pbgd.Release{FolderId: 1}}

	if !FuzzyMatch(b, a) {
		t.Errorf("Failed to match on detailed: %v != %v", b, a)
	}
}

func TestFuzzyMatchDetailedEmpty(t *testing.T) {
	a := &pbrc.Record{Release: &pbgd.Release{Id: 1234, FolderId: 1, Title: "Bonkers"}, Metadata: &pbrc.ReleaseMetadata{}}
	b := &pbrc.Record{Metadata: &pbrc.ReleaseMetadata{}}

	if !FuzzyMatch(b, a) {
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
		if actual != tt.match {
			t.Errorf("Failure in match %v vs %v", tt.b, a)
		}
	}
}

func TestGetContext(t *testing.T) {
	ctx, cancel := BuildContext("TestGetContext", "testing", pb.ContextType_REGULAR)
	defer cancel()

	v := ctx.Value("trace-id").(string)
	if len(v) == 0 {
		t.Errorf("No value set: %v", v)
	}

	log.Printf("Value = %v", v)
}
