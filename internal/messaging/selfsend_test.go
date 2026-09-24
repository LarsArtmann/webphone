// Self-send guard pins (SUPERB send-failure train C, 2026-09-24): a send
// addressed to the extension's own DID (config identities) is refused
// LOCALLY — the provider is never consulted — while the row is still
// persisted as failed (evidence-preserving) and the error classifies as a
// provider Rejection so the 422 send-failure arm renders it.
package messaging_test

import (
	"context"
	"strings"
	"testing"

	"github.com/larsartmann/go-error-family"
	errorfamilytest "github.com/larsartmann/go-error-family/errorfamilytest"
	"github.com/larsartmann/webphone/internal/blob"
	"github.com/larsartmann/webphone/internal/domain"
	"github.com/larsartmann/webphone/internal/gateway"
	"github.com/larsartmann/webphone/internal/messaging"
	"github.com/larsartmann/webphone/internal/store"
)

// countingGateway proves the guard short-circuits before the provider.
type countingGateway struct{ calls int }

func (g *countingGateway) SendMessage(_ context.Context, _ gateway.OutboundMessage) (gateway.Receipt, error) {
	g.calls++
	return gateway.Receipt{ProviderRef: "counted"}, nil
}

func newSelfSendService(t *testing.T, identities map[string]string) (*messaging.Service, *countingGateway) {
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
	gw := &countingGateway{}
	return messaging.New(store.NewMessages(db), blobs, gw, nil, identities), gw
}

func TestSendToOwnNumberIsRefusedLocally(t *testing.T) {
	svc, gw := newSelfSendService(t, map[string]string{"100": "+15550001111"})

	msg, err := svc.Send(context.Background(),
		domain.MustParseExtension("100"), domain.MustParsePhone("+15550001111"), "to myself", nil)

	errorfamilytest.AssertFamily(t, err, errorfamily.Rejection)
	if gw.calls != 0 {
		t.Fatalf("provider was consulted %d times; the guard must short-circuit", gw.calls)
	}
	if msg.Status != domain.StatusFailed {
		t.Fatalf("row status = %q, want failed (evidence-preserving)", msg.Status)
	}
	if msg.FailureKind != "rejected" {
		t.Fatalf("failure kind = %q, want rejected (no retry offer)", msg.FailureKind)
	}
	if !strings.Contains(msg.FailureDetail, "own number") {
		t.Fatalf("failure detail lost the reason: %q", msg.FailureDetail)
	}
}

// TestSendToOwnNumberMatchesSanitizedDID pins the parity the guard relies
// on: the DID from config and the destination sanitize to the same
// dialable string even when their formatting differs.
func TestSendToOwnNumberMatchesSanitizedDID(t *testing.T) {
	svc, gw := newSelfSendService(t, map[string]string{"100": "+1 (555) 000-1111"})

	_, err := svc.Send(context.Background(),
		domain.MustParseExtension("100"), domain.MustParsePhone("+15550001111"), "still myself", nil)

	errorfamilytest.AssertFamily(t, err, errorfamily.Rejection)
	if gw.calls != 0 {
		t.Fatalf("formatted DID must still match after sanitize; provider calls = %d", gw.calls)
	}
}

// TestSendWithoutIdentitiesSkipsTheGuard pins the guard-off default: no
// identities map (or no DID for this owner) leaves the gateway in charge —
// loopback dev and unconfigured installs keep working.
func TestSendWithoutIdentitiesSkipsTheGuard(t *testing.T) {
	for name, identities := range map[string]map[string]string{
		"nil map":     nil,
		"other owner": {"200": "+15550002222"},
	} {
		t.Run(name, func(t *testing.T) {
			svc, gw := newSelfSendService(t, identities)

			msg, err := svc.Send(context.Background(),
				domain.MustParseExtension("100"), domain.MustParsePhone("+15550001111"), "free to try", nil)

			if err != nil {
				t.Fatalf("guard fired without an identity for this owner: %v", err)
			}
			if gw.calls != 1 || msg.Status != domain.StatusSent {
				t.Fatalf("send must reach the gateway: calls=%d status=%q", gw.calls, msg.Status)
			}
		})
	}
}
