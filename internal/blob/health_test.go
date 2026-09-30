package blob

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

// TestProbeWriteAcceptsWritableRoot pins the write probe's happy path:
// a writable root passes and leaves no probe file behind.
func TestProbeWriteAcceptsWritableRoot(t *testing.T) {
	root := t.TempDir()
	if err := ProbeWrite(root); err != nil {
		t.Fatalf("ProbeWrite: %v", err)
	}
	entries, err := os.ReadDir(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Errorf("probe left %d files behind", len(entries))
	}
}

// TestProbeWriteFailsOnReadOnlyRoot pins the failure mode the probe
// exists for: a root that cannot accept writes must fail, not pass.
func TestProbeWriteFailsOnReadOnlyRoot(t *testing.T) {
	parent := t.TempDir()
	root := filepath.Join(parent, "ro")
	if err := os.Mkdir(root, 0o500); err != nil {
		t.Fatal(err)
	}
	if err := ProbeWrite(root); err == nil {
		t.Fatal("read-only root must fail the write probe")
	}
}

// TestStoreHealthCheckDelegatesToProbe pins the container surface:
// Store.HealthCheck answers with the same write probe over the store's
// own root.
func TestStoreHealthCheckDelegatesToProbe(t *testing.T) {
	store, err := New(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if err := store.HealthCheck(context.Background()); err != nil {
		t.Fatalf("HealthCheck: %v", err)
	}
}
