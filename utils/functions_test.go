package utils

import "testing"

import pb "github.com/brotherlogic/goserver/proto"

func TestMatcherPass(t *testing.T) {
	p1 := &pb.State{Key: "Testing"}
	p2 := &pb.State{Key: "Testing", Value: 12345}

	if !FuzzyMatch(p1, p2) {
		t.Errorf("%v and %v are not matching :-(", p1, p2)
	}
}

func TestMatcherFail(t *testing.T) {
	p1 := &pb.State{Key: "Besting"}
	p2 := &pb.State{Key: "Testing", Value: 12345}

	if FuzzyMatch(p1, p2) {
		t.Errorf("%v and %v are matching :-(", p1, p2)
	}
}
