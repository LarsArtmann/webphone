// Self-send guard pins for the fax lane (SUPERB send-failure train C,
// 2026-09-24): a fax addressed to the extension's own DID (config
// identities) is refused locally — the provider is never consulted —
// while the job is persisted as failed (evidence-preserving) and the
// error classifies as a provider Rejection so the 422 send-failure arm
// renders it.
package fax_test

import (
	"context"
	"strings"
	"testing"

	"github.com/larsartmann/go-error-family"
	errorfamilytest "github.com/larsartmann/go-error-family/errorfamilytest"
	"github.com/larsartmann/webphone/internal/blob"
	"github.com/larsartmann/webphone/internal/domain"
	"github.com/larsartmann/webphone/internal/fax"
	"github.com/larsartmann/webphone/internal/gateway"
	"github.com/larsartmann/webphone/internal/store"
)

// countingFaxGateway proves the guard short-circuits before the provider.
type countingFaxGateway struct{ calls int }

func (g *countingFaxGateway) SendFax(_ context.Context, _ gateway.OutboundFax) (gateway.Receipt, error) {
	g.calls++
	return gateway.Receipt{ProviderRef: "counted"}, nil
}

func newSelfSendFaxService(t *testing.T, identities map[string]string) (*fax.Service, *countingFaxGateway) {
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
	gw := &countingFaxGateway{}
	return fax.New(store.NewFaxes(db), blobs, gw, nil, identities), gw
}

var testPDF = []byte("%PDF-1.4\n%test\ntrailer<<>>\n%%EOF\n")

func TestFaxToOwnNumberIsRefusedLocally(t *testing.T) {
	svc, gw := newSelfSendFaxService(t, map[string]string{"100": "+15550001111"})

	job, err := svc.Send(context.Background(),
		domain.MustParseExtension("100"), domain.MustParsePhone("+15550001111"), "doc.pdf", testPDF)

	errorfamilytest.AssertFamily(t, err, errorfamily.Rejection)
	if gw.calls != 0 {
		t.Fatalf("provider was consulted %d times; the guard must short-circuit", gw.calls)
	}
	if job.Status != domain.FaxFailed {
		t.Fatalf("job status = %q, want failed (evidence-preserving)", job.Status)
	}
	if !strings.Contains(job.Error, "own number") {
		t.Fatalf("job error lost the reason: %q", job.Error)
	}
}

// TestFaxWithoutIdentitiesSkipsTheGuard pins the guard-off default: no
// identities map leaves the gateway in charge.
func TestFaxWithoutIdentitiesSkipsTheGuard(t *testing.T) {
	svc, gw := newSelfSendFaxService(t, nil)

	job, err := svc.Send(context.Background(),
		domain.MustParseExtension("100"), domain.MustParsePhone("+15550001111"), "doc.pdf", testPDF)

	if err != nil {
		t.Fatalf("guard fired without identities: %v", err)
	}
	if gw.calls != 1 || job.Status != domain.FaxSending {
		t.Fatalf("fax must reach the gateway: calls=%d status=%q", gw.calls, job.Status)
	}
}
