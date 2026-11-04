// Blob store roundtrip pins (atomic-write train, 2026-11-04): Save must
// hand back bytes that read back identically, land with the 0o640 mode,
// and leave no atomic-write staging residue (a `.tmp` corpse beside a
// blob would be swept by nothing and confuse the retention pass).
package blob_test

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/larsartmann/webphone/internal/blob"
)

func TestSaveLoadRoundtrip(t *testing.T) {
	s, err := blob.New(filepath.Join(t.TempDir(), "blobs"))
	if err != nil {
		t.Fatal(err)
	}

	want := []byte("%PDF-1.7 fax-bytes-attachment-bytes")
	rel, err := s.Save("faxes", ".pdf", want)
	if err != nil {
		t.Fatal(err)
	}

	got, err := func() ([]byte, error) {
		r, err := s.Open(rel)
		if err != nil {
			return nil, err
		}
		defer func() { _ = r.Close() }()
		return io.ReadAll(r)
	}()
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(want) {
		t.Fatalf("roundtrip mismatch: got %q want %q", got, want)
	}

	abs := s.Abs(rel)
	info, err := os.Stat(abs)
	if err != nil {
		t.Fatal(err)
	}
	if perm := info.Mode().Perm(); perm != 0o640 {
		t.Fatalf("saved blob mode: got %o want 640", perm)
	}

	entries, err := os.ReadDir(filepath.Dir(abs))
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".tmp") {
			t.Fatalf("staging residue left beside blob: %s", e.Name())
		}
	}
}
