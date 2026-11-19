// The retry OWNS its document (2026-11-19 full-code-review M1): Resend
// copies the original's blob instead of sharing the path, so retention
// expiring the original can never unlink the file under a still-live
// retry. The old shared-path shape made the sweep able to destroy a
// resolvable row's document mid-retry — a data-integrity break in the
// blob store, the only storage truth.
package fax_test

import (
	"context"
	"io"
	"testing"

	"github.com/larsartmann/webphone/internal/blob"
	"github.com/larsartmann/webphone/internal/domain"
	"github.com/larsartmann/webphone/internal/fax"
	"github.com/larsartmann/webphone/internal/gateway"
)

func TestResendOwnsItsBlobCopy(t *testing.T) {
	blobs, err := blob.New(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	gw := &resolvingFaxGateway{resolution: gateway.ResolutionImmediate}
	svc := fax.New(newTestFaxes(t), blobs, gw, nil, nil, nil)
	ctx := context.Background()
	owner := domain.MustParseExtension("100")

	job, err := svc.Send(ctx, owner, domain.MustParsePhone("+15550001111"), "doc.pdf", testPDF)
	if err != nil {
		t.Fatalf("send: %v", err)
	}
	if _, err := svc.UpdateProviderStatus(ctx, job.ProviderRef, domain.FaxFailed, 1, "refused"); err != nil {
		t.Fatalf("fail the job: %v", err)
	}

	retry, err := svc.Resend(ctx, owner, job.ID)
	if err != nil {
		t.Fatalf("resend: %v", err)
	}
	if retry.DocumentPath == job.DocumentPath {
		t.Fatal("the retry shares the original's DocumentPath — retention would unlink the shared file under the live retry")
	}

	for _, rel := range []domain.RelPath{job.DocumentPath, retry.DocumentPath} {
		doc, err := blobs.Open(rel)
		if err != nil {
			t.Fatalf("document %s unreadable after resend: %v", rel, err)
		}
		_, readErr := io.ReadAll(doc)
		_ = doc.Close() //nolint:erraudit // read-complete: nothing left to do on close failure
		if readErr != nil {
			t.Fatalf("document %s unreadable content: %v", rel, readErr)
		}
	}
}
