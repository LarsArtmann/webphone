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
	if got := bootJoin(nil, a); got != a {
		t.Fatalf("single survivor: want the bare error, got %v", got)
	}
	joined := bootJoin(a, nil, b)
	want := "alpha\nbeta"
	if joined.Error() != want {
		t.Fatalf("message = %q, want %q", joined.Error(), want)
	}
	var aggregate *bootJoinError
	if !errors.As(joined, &aggregate) || len(aggregate.errs) != 2 {
		t.Fatalf("Unwrap() []error traversal broken: %v", joined)
	}
}
