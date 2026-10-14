package domain

import "testing"

// FaxAdvance parity pins (SUPERB declarative train C2, 2026-10-14):
// the table replaces the implicit transitions of fax.drive and
// UpdateProviderStatus; these cases pin the observed semantics BEFORE
// the service rewires onto it.
func TestFaxAdvanceLegalMoves(t *testing.T) {
	cases := []struct {
		name  string
		from  FaxStatus
		event FaxEvent
		want  FaxStatus
	}{
		{"accept from queued", FaxQueued, FaxAccepted(), FaxSending},
		{"submit failure from queued", FaxQueued, FaxSubmitFailed(), FaxFailed},
		{"immediate settle from sending", FaxSending, FaxSettled(), FaxTransmitted},
		{"verdict transmitted from sending", FaxSending, FaxVerdict(FaxTransmitted), FaxTransmitted},
		{"verdict failed from sending", FaxSending, FaxVerdict(FaxFailed), FaxFailed},
		{"late transmitted after failed", FaxFailed, FaxVerdict(FaxTransmitted), FaxTransmitted},
		{"late failed after transmitted", FaxTransmitted, FaxVerdict(FaxFailed), FaxFailed},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := FaxAdvance(tc.from, tc.event)
			if !ok || got != tc.want {
				t.Fatalf("FaxAdvance(%q, %v) = %q, %v; want %q, true", tc.from, tc.event.kind, got, ok, tc.want)
			}
		})
	}
}

// The inert set: submit events off queued, settle off sending, and
// everything from received (an inbound fact has no lifecycle). Note
// what is NOT here: same-status verdicts are legal refreshing moves —
// replay dedup is the hooks seam's job (ref-keyed hooksIdem), pinned
// by TestWebhookStatusIdempotent at the server layer.
func TestFaxAdvanceInertMoves(t *testing.T) {
	cases := []struct {
		name  string
		from  FaxStatus
		event FaxEvent
	}{
		{"accept from sending", FaxSending, FaxAccepted()},
		{"settle straight from queued", FaxQueued, FaxSettled()},
		{"settle from transmitted", FaxTransmitted, FaxSettled()},
		{"submit failure from sending", FaxSending, FaxSubmitFailed()},
		{"verdict against an inbound fact", FaxReceived, FaxVerdict(FaxTransmitted)},
		{"accept against an inbound fact", FaxReceived, FaxAccepted()},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := FaxAdvance(tc.from, tc.event)
			if ok || got != tc.from {
				t.Fatalf("FaxAdvance(%q, %v) = %q, %v; want %q, false", tc.from, tc.event.kind, got, ok, tc.from)
			}
		})
	}
}
