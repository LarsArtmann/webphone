// Receipt.Resolution contract pins (2026-09-30 data-model review, filed
// twice): the receipt — not the caller's knowledge of the gateway TYPE —
// says whether the acceptance is already the verdict. Loopback resolves
// immediate; webhook providers defer to the status webhook; and the zero
// value stays deferred so a hand-built receipt can never over-promise.
package gateway

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/larsartmann/webphone/internal/domain"
)

func TestLoopbackReceiptsResolveImmediate(t *testing.T) {
	l := &Loopback{prefix: "test"}
	ctx := context.Background()

	msg, err := l.SendMessage(ctx, OutboundMessage{})
	if err != nil || msg.Resolution != ResolutionImmediate {
		t.Fatalf("SendMessage: resolution = %v, err = %v (want ResolutionImmediate, nil)", msg.Resolution, err)
	}

	fax, err := l.SendFax(ctx, OutboundFax{})
	if err != nil || fax.Resolution != ResolutionImmediate {
		t.Fatalf("SendFax: resolution = %v, err = %v (want ResolutionImmediate, nil)", fax.Resolution, err)
	}
}

func TestWebhookReceiptsResolveDeferred(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"provider_ref":"ref-1"}`))
	}))
	defer server.Close()

	gw := &Webhook{provider{cfg: configGateway(server.URL, "secret"), client: server.Client()}}
	receipt, err := gw.SendMessage(context.Background(), OutboundMessage{
		Owner: domain.MustParseExtension("1001"),
		To:    domain.MustParsePhone("+441632960961"),
		Body:  "resolution probe",
	})
	if err != nil {
		t.Fatalf("SendMessage: %v", err)
	}
	if receipt.Resolution != ResolutionDeferred {
		t.Fatalf("webhook receipt resolution = %v, want ResolutionDeferred (the verdict owes the status webhook)", receipt.Resolution)
	}
	if receipt.ProviderRef != "ref-1" {
		t.Fatalf("provider ref = %q, want %q", receipt.ProviderRef, "ref-1")
	}
}

func TestZeroReceiptStaysDeferred(t *testing.T) {
	var receipt Receipt
	if receipt.Resolution != ResolutionDeferred {
		t.Fatalf("zero-value Receipt resolves %v; the safe default must be ResolutionDeferred", receipt.Resolution)
	}
}
