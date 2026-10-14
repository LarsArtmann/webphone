// Parity pins for the delivery machine, derived from the OBSERVED
// semantics of messaging.Service.Send and DeliveryReceipt (2026-10-14,
// declarative/DDD train A1.4) BEFORE any rewiring: every row encodes
// today's behavior as data. Rewiring must keep this file byte-stable.
package domain_test

import (
	"testing"

	"github.com/larsartmann/webphone/internal/domain"
)

// everyState enumerates the machine's from-states exhaustively — the
// verdict rows below assert legality from EACH of them, which is the
// out-of-order overwrite rule stated as data.
func everyState() []domain.OutboundStatus {
	return []domain.OutboundStatus{
		domain.StatusQueued,
		domain.StatusSent,
		domain.StatusDelivered,
		domain.StatusFailed,
	}
}

func TestDeliveryAdvanceSendAcceptedFromQueued(t *testing.T) {
	next, ok := domain.DeliveryAdvance(
		domain.Delivery{Status: domain.StatusQueued, FailureKind: domain.FailureTransient, FailureDetail: "stale"},
		domain.SendAccepted("ref-1"),
	)
	if !ok {
		t.Fatal("send accepted from queued must be a legal move")
	}
	if next.Status != domain.StatusSent {
		t.Fatalf("status = %q, want sent", next.Status)
	}
	if next.ProviderRef != "ref-1" {
		t.Fatalf("provider ref = %q, want ref-1", next.ProviderRef)
	}
	if next.FailureKind != domain.FailureNone || next.FailureDetail != "" {
		t.Fatalf("send accepted must clear the failure story, got %q/%q", next.FailureKind, next.FailureDetail)
	}
}

func TestDeliveryAdvanceSendRefusedRecordsVocabularyVerbatim(t *testing.T) {
	for _, tc := range []struct {
		kind domain.FailureKind
	}{
		{domain.FailureRejected},
		{domain.FailureTransient},
	} {
		next, ok := domain.DeliveryAdvance(
			domain.Delivery{Status: domain.StatusQueued},
			domain.SendRefused(tc.kind, "boom"),
		)
		if !ok {
			t.Fatalf("send refused (%s) from queued must be a legal move", tc.kind)
		}
		if next.Status != domain.StatusFailed {
			t.Fatalf("status = %q, want failed", next.Status)
		}
		if next.FailureKind != tc.kind || next.FailureDetail != "boom" {
			t.Fatalf("failure story = %q/%q, want %q/boom", next.FailureKind, next.FailureDetail, tc.kind)
		}
		if next.ProviderRef != "" {
			t.Fatalf("a refused send carries no provider ref, got %q", next.ProviderRef)
		}
	}
}

// TestDeliveryAdvanceVerdictsAreLegalFromEveryState pins today's
// out-of-order semantics: a verdict overwrites whatever state it finds
// (delivered clears the failure story; failed records the provider
// kind), duplicates are legal same-state moves, and no from-state is
// refused. Tightening this is owner call A1.7.
func TestDeliveryAdvanceVerdictsAreLegalFromEveryState(t *testing.T) {
	for _, from := range everyState() {
		start := domain.Delivery{Status: from, FailureKind: domain.FailureTransient, FailureDetail: "old"}
		if from == domain.StatusFailed {
			start.ProviderRef = "ref-1"
		}

		next, ok := domain.DeliveryAdvance(start, domain.ProviderVerdict(domain.VerdictDelivered, ""))
		if !ok {
			t.Fatalf("delivered verdict from %s must be a legal move (overwrite as today)", from)
		}
		if next.Status != domain.StatusDelivered {
			t.Fatalf("delivered verdict from %s: status = %q, want delivered", from, next.Status)
		}
		if next.FailureKind != domain.FailureNone || next.FailureDetail != "" {
			t.Fatalf("delivered verdict from %s must clear the failure story, got %q/%q", from, next.FailureKind, next.FailureDetail)
		}

		next, ok = domain.DeliveryAdvance(start, domain.ProviderVerdict(domain.VerdictFailed, "carrier refused"))
		if !ok {
			t.Fatalf("failed verdict from %s must be a legal move (overwrite as today)", from)
		}
		if next.Status != domain.StatusFailed {
			t.Fatalf("failed verdict from %s: status = %q, want failed", from, next.Status)
		}
		if next.FailureKind != domain.FailureProvider || next.FailureDetail != "carrier refused" {
			t.Fatalf("failed verdict from %s: failure story = %q/%q, want provider/carrier refused", from, next.FailureKind, next.FailureDetail)
		}
	}
}

// TestDeliveryAdvanceSendEventsFromSettledStatesAreInert pins the
// machine's guard: send events only move from queued (Send is
// synchronous), so a late send answer against a settled state is a
// programming error the machine refuses rather than fabricates.
func TestDeliveryAdvanceSendEventsFromSettledStatesAreInert(t *testing.T) {
	settled := []domain.OutboundStatus{domain.StatusSent, domain.StatusDelivered, domain.StatusFailed}
	for _, from := range settled {
		start := domain.Delivery{Status: from, ProviderRef: "ref-1"}

		if next, ok := domain.DeliveryAdvance(start, domain.SendAccepted("ref-2")); ok {
			t.Fatalf("send accepted from %s must be inert, got %+v", from, next)
		}
		if next, ok := domain.DeliveryAdvance(start, domain.SendRefused(domain.FailureTransient, "late")); ok {
			t.Fatalf("send refused from %s must be inert, got %+v", from, next)
		}
	}
}

func TestVerdictOf(t *testing.T) {
	for _, tc := range []struct {
		status domain.OutboundStatus
		want   domain.Verdict
		ok     bool
	}{
		{domain.StatusDelivered, domain.VerdictDelivered, true},
		{domain.StatusFailed, domain.VerdictFailed, true},
		{domain.StatusQueued, "", false},
		{domain.StatusSent, "", false},
		{domain.OutboundStatus(""), "", false},
		{domain.OutboundStatus("queued "), "", false},
	} {
		got, ok := domain.VerdictOf(tc.status)
		if ok != tc.ok || got != tc.want {
			t.Fatalf("VerdictOf(%q) = %q,%v; want %q,%v", tc.status, got, ok, tc.want, tc.ok)
		}
	}
}

func TestVerdictStatusMapping(t *testing.T) {
	if domain.VerdictDelivered.Status() != domain.StatusDelivered ||
		domain.VerdictFailed.Status() != domain.StatusFailed {
		t.Fatal("verdict strings must equal the statuses they produce (webhook vocabulary)")
	}
}

func TestFailureKindIsRetryable(t *testing.T) {
	for _, tc := range []struct {
		kind domain.FailureKind
		want bool
	}{
		{domain.FailureTransient, true},
		{domain.FailureNone, false},
		{domain.FailureRejected, false},
		{domain.FailureProvider, false},
	} {
		if got := tc.kind.IsRetryable(); got != tc.want {
			t.Fatalf("IsRetryable(%q) = %v, want %v", tc.kind, got, tc.want)
		}
	}
}

// TestMessageDeliveryRoundTrip pins the Message<->Delivery bridge:
// projecting and re-applying must be value-preserving.
func TestMessageDeliveryRoundTrip(t *testing.T) {
	msg := domain.Message{
		Status:        domain.StatusFailed,
		ProviderRef:   "ref-1",
		FailureKind:   domain.FailureTransient,
		FailureDetail: "timeout",
	}
	roundTrip := msg.WithDelivery(msg.DeliveryOf())
	if roundTrip.DeliveryOf() != msg.DeliveryOf() {
		t.Fatalf("delivery round-trip changed the delivery state: %+v", roundTrip.DeliveryOf())
	}
}
