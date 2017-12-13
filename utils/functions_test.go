package utils

import (
	"testing"

	pb "github.com/brotherlogic/goserver/proto"
)

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
