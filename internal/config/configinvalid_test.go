package config

import (
	"errors"
	"testing"
)

// TestConfigInvalidErrorPrefix pins the boot-greppable prefix: every
// rejected-config verdict renders "config invalid: <cause>" and keeps
// the cause reachable through Unwrap.
func TestConfigInvalidErrorPrefix(t *testing.T) {
	cause := errors.New("addr is empty")
	verdict := &ConfigInvalidError{err: cause}

	if verdict.Error() != "config invalid: addr is empty" {
		t.Fatalf("message = %q", verdict.Error())
	}
	if !errors.Is(verdict, cause) {
		t.Fatal("cause not reachable through Unwrap")
	}
}
