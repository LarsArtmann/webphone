// Resolution-branch pins for the fax lane (SUPERB declarative train
// B1.5, 2026-10-14): drive branches on the Receipt's Resolution — the
// contract — never on the gateway's concrete type (09-30 review P5: a
// second synchronous gateway silently took the wrong branch under the
// old type-assert). A gateway that is NOT *gateway.Loopback yet resolves
// immediate must still settle the job synchronously; a deferred receipt
// must leave it sending until the status webhook lands.
package fax_test

import (
	"context"
	"testing"

	"github.com/larsartmann/webphone/internal/blob"
	"github.com/larsartmann/webphone/internal/domain"
	"github.com/larsartmann/webphone/internal/fax"
	"github.com/larsartmann/webphone/internal/gateway"
)

// resolvingFaxGateway is that second synchronous gateway: a foreign type
// whose receipt declares its own resolution. Any type-sniffing in drive
// would misroute it.
type resolvingFaxGateway struct {
	resolution gateway.Resolution
	calls      int
}

func (g *resolvingFaxGateway) SendFax(_ context.Context, _ gateway.OutboundFax) (gateway.Receipt, error) {
	g.calls++
	return gateway.Receipt{ProviderRef: "resolved", Resolution: g.resolution}, nil
}

func newResolutionFaxService(t *testing.T, gw *resolvingFaxGateway) *fax.Service {
	t.Helper()
	blobs, err := blob.New(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	return fax.New(newTestFaxes(t), blobs, gw, nil, nil, nil)
}

func TestDriveBranchesOnReceiptResolution(t *testing.T) {
	t.Run("immediate receipt from a non-loopback gateway settles the job", func(t *testing.T) {
		gw := &resolvingFaxGateway{resolution: gateway.ResolutionImmediate}
		svc := newResolutionFaxService(t, gw)

		job, err := svc.Send(context.Background(),
			domain.MustParseExtension("100"), domain.MustParsePhone("+15550001111"), "doc.pdf", testPDF)

		if err != nil {
			t.Fatalf("send: %v", err)
		}
		if gw.calls != 1 {
			t.Fatalf("provider calls = %d, want 1", gw.calls)
		}
		if job.Status != domain.FaxTransmitted {
			t.Fatalf("immediate receipt must settle synchronously: status = %q, want transmitted", job.Status)
		}
		if job.ProviderRef != "resolved" {
			t.Fatalf("provider ref lost on the settled job: %q", job.ProviderRef)
		}
	})

	t.Run("deferred receipt keeps the job sending", func(t *testing.T) {
		gw := &resolvingFaxGateway{resolution: gateway.ResolutionDeferred}
		svc := newResolutionFaxService(t, gw)

		job, err := svc.Send(context.Background(),
			domain.MustParseExtension("100"), domain.MustParsePhone("+15550001111"), "doc.pdf", testPDF)

		if err != nil {
			t.Fatalf("send: %v", err)
		}
		if job.Status != domain.FaxSending {
			t.Fatalf("deferred receipt must not settle: status = %q, want sending (the webhook owes the verdict)", job.Status)
		}
		if job.ProviderRef != "resolved" {
			t.Fatalf("acceptance ref lost: %q", job.ProviderRef)
		}
	})
}

// TestVerdictSemantics pins the verdict contract end-to-end (train C2,
// re-adjudicated against the pinned panel behavior): a same-status
// verdict is a legal REFRESHING move (a later hook may carry the real
// page count), a different late verdict overwrites, and only an event
// against an inbound "received" fact is inert. Replay dedup itself is
// the hooks seam's job (ref-keyed hooksIdem) — see
// TestWebhookStatusIdempotent at the server layer.
func TestVerdictSemantics(t *testing.T) {
	gw := &resolvingFaxGateway{resolution: gateway.ResolutionImmediate}
	svc := newResolutionFaxService(t, gw)
	ctx := context.Background()
	owner := domain.MustParseExtension("100")

	job, err := svc.Send(ctx, owner, domain.MustParsePhone("+15550001111"), "doc.pdf", testPDF)
	if err != nil || job.Status != domain.FaxTransmitted {
		t.Fatalf("setup: err=%v status=%q", err, job.Status)
	}

	refreshed, err := svc.UpdateProviderStatus(ctx, job.ProviderRef, domain.FaxTransmitted, 7, "")
	if err != nil {
		t.Fatalf("same-status verdict: %v", err)
	}
	if refreshed.Status != domain.FaxTransmitted {
		t.Fatalf("refresh keeps the settled status: %q", refreshed.Status)
	}
	row, err := svc.Get(ctx, owner, job.ID)
	if err != nil {
		t.Fatal(err)
	}
	if row.Pages != 7 {
		t.Fatalf("same-status verdict must refresh the payload: pages = %d, want 7", row.Pages)
	}

	late, err := svc.UpdateProviderStatus(ctx, job.ProviderRef, domain.FaxFailed, 2, "late failure")
	if err != nil {
		t.Fatalf("late verdict: %v", err)
	}
	if late.Status != domain.FaxFailed {
		t.Fatalf("late failed verdict must overwrite: %q", late.Status)
	}
	row, err = svc.Get(ctx, owner, job.ID)
	if err != nil {
		t.Fatal(err)
	}
	if row.Status != domain.FaxFailed || row.Pages != 2 || row.Error != "late failure" {
		t.Fatalf("late verdict must land: %+v", row)
	}

	// The remaining inert class: an inbound fact has no lifecycle — a
	// stray verdict for its ref changes nothing.
	inbound, err := svc.Receive(ctx, domain.InboundFax{
		Owner: owner, From: domain.MustParsePhone("+15550002222"),
		Pages: 3, PDFBytes: testPDF, ProviderRef: "inbound-ref",
	})
	if err != nil || inbound.Status != domain.FaxReceived {
		t.Fatalf("inbound setup: err=%v status=%q", err, inbound.Status)
	}
	stray, err := svc.UpdateProviderStatus(ctx, "inbound-ref", domain.FaxTransmitted, 9, "")
	if err != nil {
		t.Fatalf("stray verdict against an inbound fact: %v", err)
	}
	if stray.Status != domain.FaxReceived {
		t.Fatalf("inbound fact must stay received: %q", stray.Status)
	}
	row, err = svc.Get(ctx, owner, stray.ID)
	if err != nil {
		t.Fatal(err)
	}
	if row.Status != domain.FaxReceived || row.Pages != 3 {
		t.Fatalf("stray verdict must not write: %+v", row)
	}
}
