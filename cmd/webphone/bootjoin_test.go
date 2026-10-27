package main

import (
	"errors"
	"testing"
)

// TestBootJoinPinsJoinSemantics pins the aggregate against the
// errors.Join product it replaces: nils dropped, all-nil answers nil,
// and the message format is one line per surviving cause.
func TestBootJoinPinsJoinSemantics(t *testing.T) {
	a := errors.New("alpha")
	b := errors.New("beta")

	if bootJoin(nil, nil) != nil {
		t.Fatal("all-nil input: want nil")
	}
	if got := bootJoin(nil, a); got == nil || got.Error() != "alpha" {
		t.Fatalf("single survivor: want Join-parity wrapper, got %v", got)
	}
	joined := bootJoin(a, nil, b)
	want := "alpha\nbeta"
	if joined.Error() != want {
		t.Fatalf("message = %q, want %q", joined.Error(), want)
	}
	aggregate, ok := errors.AsType[*bootJoinError](joined)
	if !ok || len(aggregate.errs) != 2 {
		t.Fatalf("Unwrap() []error traversal broken: %v", joined)
	}
}
