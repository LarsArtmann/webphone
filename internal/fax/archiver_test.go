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
	"github.com/larsartmann/webphone/internal/config"
	"github.com/larsartmann/webphone/internal/domain"
	"github.com/larsartmann/webphone/internal/fax"
	"github.com/larsartmann/webphone/internal/gateway"
	"github.com/larsartmann/webphone/internal/store"
)

// archivedFax is what the fake records per offer.
type archivedFax struct {
	job domain.FaxJob
	pdf []byte
}

// recordingArchiver captures one archive offer.
type recordingArchiver struct {
	offered chan archivedFax
}

func newRecordingArchiver() *recordingArchiver {
	return &recordingArchiver{offered: make(chan archivedFax, 1)}
}

func (a *recordingArchiver) ArchiveFax(_ context.Context, job domain.FaxJob, pdf []byte) error {
	a.offered <- archivedFax{job: job, pdf: pdf}
	return nil
}

func newArchiveFaxService(t *testing.T, archiver fax.Archiver) (*fax.Service, *store.Faxes) {
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
		config.Gateway{Mode: config.GatewayLoopback}, gateway.DefaultClient(),
	), nil, nil, archiver)
	return svc, faxes
}

// TestReceiveArchivesInbound pins the happy path: the archived payload is
// the spooled PDF re-read from the blob store, carrying the persisted
// job (id, remote, direction, received stamp) the archive titles by.
func TestReceiveArchivesInbound(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		archiver := newRecordingArchiver()
		svc, faxes := newArchiveFaxService(t, archiver)

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

		var got archivedFax
		select {
		case got = <-archiver.offered:
		case <-time.After(10 * time.Second):
			t.Fatal("archiver was never offered the fax")
		}

		if got.job.ID != job.ID {
			t.Errorf("archived job id = %s, want %s", got.job.ID, job.ID)
		}
		if got.job.Remote.String() != "+4930123456" || got.job.Direction != domain.FaxInbound {
			t.Errorf("archived job lost remote/direction: %+v", got.job)
		}
		if !got.job.CreatedAt.Equal(received) {
			t.Errorf("archived job created = %v, want %v", got.job.CreatedAt, received)
		}
		if string(got.pdf) != string(testPDF) {
			t.Errorf("archived bytes differ from the spooled pdf (%d vs %d)", len(got.pdf), len(testPDF))
		}
		if filepath.Base(got.job.DocumentPath) == "" {
			t.Error("archived job carries no document path")
		}

		// The offer rides AFTER persistence: the job is already in the store.
		if _, err := faxes.Get(context.Background(), job.Owner, job.ID); err != nil {
			t.Errorf("archived fax is not persisted: %v", err)
		}
	})
}

// TestReceiveWithoutArchiverPinsNilOff pins the nil-off default: no
// archiver, no goroutine, unchanged flow.
func TestReceiveWithoutArchiverPinsNilOff(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		svc, faxes := newArchiveFaxService(t, nil)

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
		if _, err := faxes.Get(context.Background(), job.Owner, job.ID); err != nil {
			t.Errorf("fax not persisted: %v", err)
		}
		time.Sleep(time.Second) // synctest: would surface a stray archive goroutine panic
	})
}
