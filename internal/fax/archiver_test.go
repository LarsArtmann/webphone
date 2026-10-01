// The optional downstream archive seam (fax → Paperless, 2026-10-01):
// Receive offers every persisted inbound fax to the archiver on a
// detached goroutine; a nil archiver keeps the pre-integration behavior.
package fax_test

import (
	"context"
	"path/filepath"
	"testing"
	"testing/synctest"
	"time"

	"github.com/larsartmann/webphone/internal/blob"
	"github.com/larsartmann/webphone/internal/domain"
	"github.com/larsartmann/webphone/internal/fax"
	"github.com/larsartmann/webphone/internal/gateway"
	"github.com/larsartmann/webphone/internal/store"
)

// recordingArchiver captures one archive offer.
type recordingArchiver struct {
	offered chan fax.ArchiveOffer
}

func newRecordingArchiver() *recordingArchiver {
	return &recordingArchiver{offered: make(chan fax.ArchiveOffer, 1)}
}

func (a *recordingArchiver) ArchiveFax(_ context.Context, job domain.FaxJob, pdf []byte) error {
	a.offered <- fax.ArchiveOffer{Job: job, PDF: pdf}
	return nil
}

func newArchiveFaxService(t *testing.T, archiver fax.Archiver) (*fax.Service, *store.Faxes, *blob.Store) {
	t.Helper()
	db, err := store.Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	blobs, err := blob.New(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	faxes := store.NewFaxes(db)
	svc := fax.New(faxes, blobs, gateway.NewFaxGateway(
		gateway.Config{Mode: gateway.Loopback}, nil,
	), nil, nil, archiver)
	return svc, faxes, blobs
}

// TestReceiveArchivesInbound pins the happy path: the archived payload is
// the spooled PDF re-read from the blob store, carrying the persisted
// job (id, remote, direction, received stamp) the archive titles by.
func TestReceiveArchivesInbound(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		archiver := newRecordingArchiver()
		svc, faxes, _ := newArchiveFaxService(t, archiver)

		received := time.Date(2026, 10, 1, 12, 30, 0, 0, time.UTC)
		job, err := svc.Receive(context.Background(), domain.InboundFax{
			Owner:    domain.MustParseExtension("100"),
			From:     domain.MustParsePhone("+4930123456"),
			Pages:    2,
			PDFBytes: testPDF,
			Received: received,
		})
		if err != nil {
			t.Fatal(err)
		}

		var offer fax.ArchiveOffer
		select {
		case offer = <-archiver.offered:
		case <-time.After(10 * time.Second):
			t.Fatal("archiver was never offered the fax")
		}

		if offer.Job.ID != job.ID {
			t.Errorf("archived job id = %s, want %s", offer.Job.ID, job.ID)
		}
		if offer.Job.Remote.String() != "+4930123456" || offer.Job.Direction != domain.FaxInbound {
			t.Errorf("archived job lost remote/direction: %+v", offer.Job)
		}
		if !offer.Job.CreatedAt.Equal(received) {
			t.Errorf("archived job created = %v, want %v", offer.Job.CreatedAt, received)
		}
		if string(offer.PDF) != string(testPDF) {
			t.Errorf("archived bytes differ from the spooled pdf (%d vs %d)", len(offer.PDF), len(testPDF))
		}
		if filepath.Base(offer.Job.DocumentPath) == "" {
			t.Error("archived job carries no document path")
		}

		// The offer rides AFTER persistence: the job is already in the store.
		if _, err := faxes.ByID(context.Background(), job.ID); err != nil {
			t.Errorf("archived fax is not persisted: %v", err)
		}
	})
}

// TestReceiveWithoutArchiverPinsNilOff pins the nil-off default: no
// archiver, no goroutine, unchanged flow.
func TestReceiveWithoutArchiverPinsNilOff(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		svc, faxes, _ := newArchiveFaxService(t, nil)

		job, err := svc.Receive(context.Background(), domain.InboundFax{
			Owner:    domain.MustParseExtension("100"),
			From:     domain.MustParsePhone("+4930123456"),
			PDFBytes: testPDF,
		})
		if err != nil {
			t.Fatal(err)
		}
		if job.Status != domain.FaxReceived {
			t.Errorf("status = %q, want received", job.Status)
		}
		if _, err := faxes.ByID(context.Background(), job.ID); err != nil {
			t.Errorf("fax not persisted: %v", err)
		}
		time.Sleep(time.Second) // synctest: would surface a stray archive goroutine panic
	})
}
