// Blob family pins (family-adoption train, 2026-09-30): every blob
// failure classifies Infrastructure (our storage, not retryable by the
// user) except the traversal refusal, which is a Rejection of the
// input. Families must survive the service-level wraps the inbound
// paths add, so classifyForUser keeps answering 502 generic for blob
// failures and never blames the user.
package blob_test

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/larsartmann/go-error-family"
	errorfamilytest "github.com/larsartmann/go-error-family/errorfamilytest"
	"github.com/larsartmann/webphone/internal/blob"
)

func TestBlobFailuresAreInfrastructure(t *testing.T) {
	root := filepath.Join(t.TempDir(), "blobs")

	// A root whose parent is a FILE makes every mkdir/write/open fail
	// deterministically without mocking the filesystem.
	blocker := filepath.Join(t.TempDir(), "blocker")
	if err := os.WriteFile(blocker, nil, 0o640); err != nil {
		t.Fatal(err)
	}

	_, err := blob.New(filepath.Join(blocker, "nested", "root"))
	if err == nil {
		t.Fatal("New under a file: want error")
	}
	errorfamilytest.AssertFamily(t, err, errorfamily.Infrastructure)
	errorfamilytest.AssertCode(t, err, "blob.root")
	errorfamilytest.AssertRetryable(t, err, false)

	s, err := blob.New(root)
	if err != nil {
		t.Fatal(err)
	}

	if err := os.Chmod(root, 0o500); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(root, 0o700) })

	if _, err := s.Save("attachments", ".pdf", nil); err == nil {
		t.Fatal("Save under a read-only root: want error")
	} else {
		errorfamilytest.AssertFamily(t, err, errorfamily.Infrastructure)
	}

	if err := os.Chmod(root, 0o700); err != nil {
		t.Fatal(err)
	}

	if _, err := s.Open(filepath.Join("..", "escape")); err == nil {
		t.Fatal("Open traversal: want refusal")
	} else {
		errorfamilytest.AssertFamily(t, err, errorfamily.Rejection)
		errorfamilytest.AssertCode(t, err, "blob.escape")
	}

	if _, err := s.Open("missing.pdf"); err == nil {
		t.Fatal("Open missing: want error")
	} else {
		errorfamilytest.AssertFamily(t, err, errorfamily.Infrastructure)
		errorfamilytest.AssertCode(t, err, "blob.open")
	}
}

func TestBlobFamilySurvivesServiceWraps(t *testing.T) {
	s, err := blob.New(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	_, openErr := s.Open("missing.pdf")
	if openErr == nil {
		t.Fatal("want open error")
	}
	for _, wrap := range []error{
		fmt.Errorf("store inbound attachment: %w", openErr),
		fmt.Errorf("spool inbound pdf: %w", openErr),
		errors.Join(openErr, errors.New("sweep note")),
	} {
		errorfamilytest.AssertFamily(t, wrap, errorfamily.Infrastructure)
	}
}
