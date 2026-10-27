package paperless

import (
	"errors"
	"testing"
)

// TestArchiveErrorKeepsCause pins the envelope contract: byte-identical
// message, cause reachable through Unwrap.
func TestArchiveErrorKeepsCause(t *testing.T) {
	cause := errors.New("task timeout")
	wrapped := archiveErrf("paperless: task: %w", cause)

	if wrapped.Error() != "paperless: task: task timeout" {
		t.Fatalf("message = %q", wrapped.Error())
	}
	if !errors.Is(wrapped, cause) {
		t.Fatal("cause not reachable through Unwrap")
	}
}
