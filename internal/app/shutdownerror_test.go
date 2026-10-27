package app

import (
	"errors"
	"testing"
)

// TestShutdownErrorFormat pins the aggregate against the errors.Join
// product it replaces: one line per cause, single-cause passthrough,
// and Unwrap() []error traversal for the boot report's classifier.
func TestShutdownErrorFormat(t *testing.T) {
	a := errors.New("alpha")
	b := errors.New("beta")

	single := &ShutdownError{errs: []error{a}}
	if single.Error() != "alpha" {
		t.Fatalf("single cause: %q", single.Error())
	}
	joined := &ShutdownError{errs: []error{a, b}}
	if joined.Error() != "alpha\nbeta" {
		t.Fatalf("message = %q, want %q", joined.Error(), "alpha\nbeta")
	}
	if halves := joined.Unwrap(); len(halves) != 2 || halves[0] != a || halves[1] != b {
		t.Fatalf("Unwrap() []error traversal broken: %v", halves)
	}
}
