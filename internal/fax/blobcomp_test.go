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
	"github.com/larsartmann/webphone/internal/gateway"
)

// refusingCreate answers every read through the real journal backend
// but refuses Create — the persist-failure arm without a full disk.
type refusingCreate struct {
	fax.Store
}

func (r *refusingCreate) Create(_ context.Context, _ domain.FaxJob) error {
	return errors.New("persist down")
}

// stubFaxGateway never fires: a compensated Send fails at the store,
// before the gateway is reached.
type stubFaxGateway struct{}

func (stubFaxGateway) SendFax(_ context.Context, _ gateway.OutboundFax) (gateway.Receipt, error) {
	return gateway.Receipt{ProviderRef: "unreached"}, nil
}

// faxBlobFileCount walks the store root and counts regular files — zero
// after a compensated failure is the whole assertion.
func faxBlobFileCount(t *testing.T, root string) int {
	t.Helper()
	count := 0
	err := filepath.WalkDir(root, func(_ string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.Type().IsRegular() {
			count++
		}
		return nil
	})
	if err != nil && !strings.Contains(err.Error(), "no such file") {
		t.Fatalf("walk blob root: %v", err)
	}
	return count
}

func newBlobCompService(t *testing.T) (*fax.Service, *blob.Store) {
	t.Helper()
	blobs, err := blob.New(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	store := &refusingCreate{Store: newTestFaxes(t)}
	return fax.New(store, blobs, stubFaxGateway{}, nil, nil, nil), blobs
}

func TestFailedFaxSendLeavesNoOrphanedSpool(t *testing.T) {
	svc, blobs := newBlobCompService(t)

	_, err := svc.Send(context.Background(), domain.MustParseExtension("100"),
		domain.MustParsePhone("+491713"), "doc.pdf", testPDF)
	if err == nil || !strings.Contains(err.Error(), "persist down") {
		t.Fatalf("Send verdict = %v, want the persist failure", err)
	}
	if n := faxBlobFileCount(t, blobs.Root()); n != 0 {
		t.Fatalf("%d spool files survived a failed persist — orphans retention can never find", n)
	}
}

func TestFailedFaxReceiveLeavesNoOrphanedSpool(t *testing.T) {
	svc, blobs := newBlobCompService(t)

	_, err := svc.Receive(context.Background(), domain.InboundFax{
		Owner:       domain.MustParseExtension("100"),
		From:        domain.MustParsePhone("+491714"),
		ProviderRef: "inbound-ref-1",
		Pages:       1,
		PDFBytes:    testPDF,
		Received:    time.Now(),
	})
	if err == nil || !strings.Contains(err.Error(), "persist down") {
		t.Fatalf("Receive verdict = %v, want the persist failure", err)
	}
	if n := faxBlobFileCount(t, blobs.Root()); n != 0 {
		t.Fatalf("%d spool files survived a failed inbound persist — orphans retention can never find", n)
	}
}
