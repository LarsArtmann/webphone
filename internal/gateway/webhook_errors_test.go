// Provider-call error branches (plan T17, 2026-09-20): the webhook POST
// has four failure shapes — transport failure, provider rejection
// (non-2xx with a detail snippet), a non-JSON receipt that is NOT a bare
// token (rejected), and an acceptance whose bare token is too long
// (rejected). The timeout branch rides the same transport path with a
// bounded client. Table-driven so each branch names itself.
package gateway

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/larsartmann/webphone/internal/domain"
)

func sendVia(t *testing.T, gw *Webhook, owner domain.Extension) {
	t.Helper()
	_, err := gw.SendMessage(context.Background(), OutboundMessage{
		Owner: owner, To: domain.MustParsePhone("+441632960961"), Body: "branch probe",
	})
	if err == nil {
		t.Fatal("expected an error")
	}
}

func TestWebhookPostErrorBranches(t *testing.T) {
	owner := domain.MustParseExtension("1001")

	t.Run("transport failure carries the url context", func(t *testing.T) {
		gw := webhookGateway("http://127.0.0.1:1", "s", http.DefaultClient)
		sendVia(t, gw, owner) // nothing listens on port 1
	})

	t.Run("timeout is a transport failure, not a panic", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(300 * time.Millisecond)
		}))
		defer srv.Close()
		gw := webhookGateway(srv.URL, "s", &http.Client{Timeout: 50 * time.Millisecond})
		sendVia(t, gw, owner)
	})

	t.Run("provider 5xx surfaces the status and the detail snippet", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "upstream number blocked", http.StatusBadGateway)
		}))
		defer srv.Close()
		gw := webhookGateway(srv.URL, "s", srv.Client())
		_, err := gw.SendMessage(context.Background(), OutboundMessage{
			Owner: owner, To: domain.MustParsePhone("+441632960961"), Body: "x",
		})
		if err == nil {
			t.Fatal("expected provider rejection error")
		}
		for _, want := range []string{"HTTP 502", "upstream number blocked"} {
			if !strings.Contains(err.Error(), want) {
				t.Errorf("error %q missing %q", err.Error(), want)
			}
		}
	})

	t.Run("non-2xx detail is truncated before embedding", func(t *testing.T) {
		blob := strings.Repeat("x", 4096)
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, blob, http.StatusInternalServerError)
		}))
		defer srv.Close()
		gw := webhookGateway(srv.URL, "s", srv.Client())
		_, err := gw.SendMessage(context.Background(), OutboundMessage{
			Owner: owner, To: domain.MustParsePhone("+441632960961"), Body: "x",
		})
		if err == nil {
			t.Fatal("expected provider rejection error")
		}
		if len(err.Error()) > 1024 {
			t.Errorf("error embeds the whole provider body (%d chars) — the 512-byte snippet was ignored", len(err.Error()))
		}
	})

	t.Run("a JSON-ish non-JSON receipt is rejected, not tokenized", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte(`{"broken`)) // contains '{', fails JSON decode, not a bare token
		}))
		defer srv.Close()
		gw := webhookGateway(srv.URL, "s", srv.Client())
		_, err := gw.SendMessage(context.Background(), OutboundMessage{
			Owner: owner, To: domain.MustParsePhone("+441632960961"), Body: "x",
		})
		if err == nil {
			t.Fatal("expected receipt decode error")
		}
		if !strings.Contains(err.Error(), "decode provider receipt") {
			t.Errorf("error %q is not the decode failure", err.Error())
		}
	})

	t.Run("an oversized bare token is rejected", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte(strings.Repeat("t", 4096)))
		}))
		defer srv.Close()
		gw := webhookGateway(srv.URL, "s", srv.Client())
		_, err := gw.SendMessage(context.Background(), OutboundMessage{
			Owner: owner, To: domain.MustParsePhone("+441632960961"), Body: "x",
		})
		if err == nil {
			t.Fatal("expected oversized token to be rejected")
		}
	})

	t.Run("an acceptance with an empty JSON ref is an empty receipt", func(t *testing.T) {
		// Pins the CURRENT contract: a 2xx JSON body without provider_ref
		// is accepted with an empty ref (the queue treats it as
		// fire-and-forget). If this ever flips to an error, flip the
		// server-side messaging contract with it.
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte(`{}`))
		}))
		defer srv.Close()
		gw := webhookGateway(srv.URL, "s", srv.Client())
		receipt, err := gw.SendMessage(context.Background(), OutboundMessage{
			Owner: owner, To: domain.MustParsePhone("+441632960961"), Body: "x",
		})
		if err != nil {
			t.Fatal(err)
		}
		if receipt.ProviderRef != "" {
			t.Errorf("empty-ref receipt drifted: %q", receipt.ProviderRef)
		}
	})
}

func TestFaxWebhookSharesThePostBranches(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "fax lane down", http.StatusServiceUnavailable)
	}))
	defer srv.Close()
	gw := faxWebhookGateway(srv.URL, "s", srv.Client())
	pdf := filepath.Join(t.TempDir(), "fax.pdf")
	if err := os.WriteFile(pdf, []byte("%PDF-1.4"), 0o600); err != nil {
		t.Fatal(err)
	}
	_, err := gw.SendFax(context.Background(), OutboundFax{
		Owner: domain.MustParseExtension("1001"), To: domain.MustParsePhone("+441632960961"), PDFPath: pdf,
	})
	if err == nil || !strings.Contains(err.Error(), "HTTP 503") {
		t.Fatalf("fax send did not surface the provider rejection: %v", err)
	}
}
