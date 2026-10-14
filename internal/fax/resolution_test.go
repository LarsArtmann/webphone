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
	"github.com/larsartmann/webphone/internal/store"
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
	db, err := store.Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	blobs, err := blob.New(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	return fax.New(store.NewFaxes(db), blobs, gw, nil, nil, nil)
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
