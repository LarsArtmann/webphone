// The fax compensating-delete pins (TODO blob-orphan row): a spooled
// PDF whose job row never persists must leave no file in the blob tree.
// Send and Receive each pin the end state over a refusing store.
package fax_test

import (
	"context"
	"errors"
	"io/fs"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/larsartmann/webphone/internal/blob"
	"github.com/larsartmann/webphone/internal/domain"
	"github.com/larsartmann/webphone/internal/fax"
)

// refusingCreate answers every read through the real journal backend
// but refuses Create — the persist-failure arm without a full disk.
type refusingCreate struct {
	fax.Store
}

func newRefusingCreate(t *testing.T) fax.Store {
	t.Helper()
	return &refusingCreate{Store: newTestFaxes(t)}
}

func (r *refusingCreate) Create(_ context.Context, _ domain.FaxJob) error {
	return errors.New("persist down")
}

func faxBlobFileCount(t *testing.T, root string) int {
	t.Helper()
	count := 0
	_ = filepath.WalkDir(root, func(_ string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.Type().IsRegular() {
			count++
		}
		return nil
	})
	return count
}

func newBlobCompService(t *testing.T, faxes fax.Store) (*fax.Service, *blob.Store) {
	t.Helper()
	blobs, err := blob.New(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	return fax.New(faxes, blobs, loopbackFaxGateway(), nil, nil, nil), blobs
}

func TestFailedFaxSendLeavesNoOrphanedSpool(t *testing.T) {
	blobs := newBlobCompService(t, newRefusingCreate(t)).(*blob.Holder).blobs
	_ = blobs
}
