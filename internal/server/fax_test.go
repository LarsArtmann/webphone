package server

import (
	"bytes"
	"encoding/base64"
	"encoding/json/v2"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"

	"github.com/larsartmann/webphone/internal/config"
)

func TestFaxSendAndDocument(t *testing.T) {
	c := newClient(t)
	c.login("1001", "pw")

	pdf := []byte("%PDF-1.4\n%test\ntrailer<<>>\n%%EOF\n")
	form, contentType := multipartBody(t,
		map[string]string{"to": "+441632960961"},
		map[string]struct {
			Name    string
			Content []byte
		}{"document": {Name: "doc.pdf", Content: pdf}})
	resp, body := c.do(http.MethodPost, "/fax/send", form, contentType)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("fax send: %d %s", resp.StatusCode, body)
	}
	if !strings.Contains(string(body), "transmitted") {
		t.Fatal("loopback fax must resolve to transmitted")
	}

	match := regexp.MustCompile(`href="(/fax/[^/]+/document)"`).FindSubmatch(body)
	if match == nil {
		t.Fatal("document link missing")
	}
	resp, body = c.do(http.MethodGet, string(match[1]), nil, "")
	if resp.StatusCode != http.StatusOK || !bytes.HasPrefix(body, []byte("%PDF")) {
		t.Fatalf("fax document: %d", resp.StatusCode)
	}

	// Non-PDF is rejected.
	form, contentType = multipartBody(t,
		map[string]string{"to": "+441632960961"},
		map[string]struct {
			Name    string
			Content []byte
		}{"document": {Name: "x.txt", Content: []byte("not a pdf")}})
	resp, _ = c.do(http.MethodPost, "/fax/send", form, contentType)
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("non-pdf fax: %d (want 422)", resp.StatusCode)
	}
}

// TestFaxProviderRefusalAnswers422 pins the fax lane's refusal arm
// (send-failure train E): a provider that ANSWERS with a 4xx reason
// surfaces that reason at 422 — input feedback, not a system fault —
// and the job is saved as failed with the provider's detail.
func TestFaxProviderRefusalAnswers422(t *testing.T) {
	rejecting := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "217022: not a valid fax destination", http.StatusUnprocessableEntity)
	}))
	t.Cleanup(rejecting.Close)
	server := newTestServerWithConfig(t, "", func(cfg *config.Config) {
		cfg.Gateway = config.Gateway{
			Mode:          config.GatewayWebhook,
			WebhookURL:    rejecting.URL,
			WebhookSecret: "test-secret",
		}
	})
	c := clientFor(t, server)
	c.login("1001", "pw")

	pdf := []byte("%PDF-1.4\n%test\ntrailer<<>>\n%%EOF\n")
	form, contentType := multipartBody(t,
		map[string]string{"to": "+441632960977"},
		map[string]struct {
			Name    string
			Content []byte
		}{"document": {Name: "doc.pdf", Content: pdf}})
	resp, body := c.do(http.MethodPost, "/fax/send", form, contentType)
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("refused fax: %d %s (want 422, the refusal arm)", resp.StatusCode, body)
	}
	if !strings.Contains(string(body), "217022: not a valid fax destination") {
		t.Fatalf("the provider's refusal detail must render: %.300s", body)
	}
	if !strings.Contains(string(body), "failed") {
		t.Fatal("the refused job must show as failed in the re-rendered panel")
	}
}

// TestFaxToOwnNumberIsRefusedLocally pins the fax lane's self-send guard
// (send-failure train C): with the owner's DID known (config identities),
// a fax to it is refused locally — 422 with the reason, no provider
// roundtrip — and the job is saved as failed (evidence-preserving).
func TestFaxToOwnNumberIsRefusedLocally(t *testing.T) {
	server := newTestServerWithConfig(t, "", func(cfg *config.Config) {
		cfg.Identities = map[string]string{"1001": "+17287289311"}
	})
	c := clientFor(t, server)
	c.login("1001", "pw")

	form, contentType := multipartBody(t,
		map[string]string{"to": "+17287289311"},
		map[string]struct {
			Name    string
			Content []byte
		}{"document": {Name: "doc.pdf", Content: []byte("%PDF-1.4\n%test\ntrailer<<>>\n%%EOF\n")}})
	resp, body := c.do(http.MethodPost, "/fax/send", form, contentType)
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("self fax: %d %s (want 422, refused locally)", resp.StatusCode, body)
	}
	if !strings.Contains(string(body), "own number") {
		t.Fatalf("422 body lost the refusal reason: %.300s", body)
	}
	if !strings.Contains(string(body), "failed") {
		t.Fatal("the refused job must show as failed in the re-rendered panel")
	}
}

func TestInboundFaxWebhookStoresDocument(t *testing.T) {
	server := newTestServer(t)
	pdf := []byte("%PDF-1.4 hook\n%%EOF\n")
	payload, _ := json.Marshal(map[string]any{
		"owner": "1001", "from": "+491700000000", "pages": 2,
		"pdf_base64": base64.StdEncoding.EncodeToString(pdf),
	})
	req, _ := http.NewRequest(http.MethodPost, server.URL+"/hooks/fax", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test-secret")
	resp, err := server.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusAccepted {
		t.Fatalf("fax webhook: %d", resp.StatusCode)
	}
}

// TestFaxResendFailedJobCreatesNewJob pins the resend lane (M14 C6): a
// FAILED outbound row offers a resend button; resending submits the
// stored document as a NEW job (the original row stays as evidence),
// and the retry rides the same gateway verdicts as a fresh send.
func TestFaxResendFailedJobCreatesNewJob(t *testing.T) {
	accept := false
	gateway := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if accept {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"provider_ref":"retry-1"}`))
			return
		}
		http.Error(w, "217022: not a valid fax destination", http.StatusUnprocessableEntity)
	}))
	t.Cleanup(gateway.Close)
	server := newTestServerWithConfig(t, "", func(cfg *config.Config) {
		cfg.Gateway = config.Gateway{
			Mode:          config.GatewayWebhook,
			WebhookURL:    gateway.URL,
			WebhookSecret: "test-secret",
		}
	})
	c := clientFor(t, server)
	c.login("1001", "pw")

	// Phase 1: the provider refuses — the job lands as failed.
	pdf := []byte("%PDF-1.4\n%test\ntrailer<<>>\n%%EOF\n")
	form, contentType := multipartBody(t,
		map[string]string{"to": "+441632960977"},
		map[string]struct {
			Name    string
			Content []byte
		}{"document": {Name: "doc.pdf", Content: pdf}})
	resp, body := c.do(http.MethodPost, "/fax/send", form, contentType)
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("initial send: %d %s (want the refused arm)", resp.StatusCode, body)
	}
	match := regexp.MustCompile(`hx-post="(/fax/[^/]+/resend)"`).FindSubmatch(body)
	if match == nil {
		t.Fatal("a failed outbound row must offer the resend button")
	}
	if !strings.Contains(string(body), "wp-steps-failed") {
		t.Fatal("a failed outbound row must render the failed timeline state")
	}

	// Phase 2: the provider accepts — the resend becomes a new sending job.
	accept = true
	resp, body = c.do(http.MethodPost, string(match[1]), nil, "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("resend: %d %s", resp.StatusCode, body)
	}
	page := string(body)
	if !strings.Contains(page, "sending") {
		t.Fatal("the retried job must show as sending (deferred webhook verdict)")
	}
	if !strings.Contains(page, "failed") {
		t.Fatal("the original failed row must survive the resend as evidence")
	}
	if got := strings.Count(page, "wp-fax-row"); got != 2 {
		t.Fatalf("resend must add a row, not mutate the failed one: %d rows", got)
	}
}

// TestFaxResendRefusesNonFailedRows pins the state guards: delivered,
// in-flight, and inbound rows never offer re-submission — only failed
// outbound ones do (else resending would duplicate or fabricate
// traffic). Wrong-owner ids are plain 404s.
func TestFaxResendRefusesNonFailedRows(t *testing.T) {
	server := newTestServer(t)
	c := clientFor(t, server)
	c.login("1001", "pw")

	pdf := []byte("%PDF-1.4\n%test\ntrailer<<>>\n%%EOF\n")
	form, contentType := multipartBody(t,
		map[string]string{"to": "+441632960961"},
		map[string]struct {
			Name    string
			Content []byte
		}{"document": {Name: "doc.pdf", Content: pdf}})
	resp, body := c.do(http.MethodPost, "/fax/send", form, contentType)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("loopback send: %d %s", resp.StatusCode, body)
	}
	if strings.Contains(string(body), "/resend") {
		t.Fatal("a transmitted row must not offer resend")
	}
	if !strings.Contains(string(body), "wp-steps-done") {
		t.Fatal("a transmitted row must render the done timeline state")
	}
	outboundID := regexp.MustCompile(`href="(/fax/([^/]+)/document)`).FindSubmatch(body)
	if outboundID == nil {
		t.Fatal("no fax row to test the guard with")
	}

	// Delivered outbound: 422 with the state reason.
	resp, body = c.do(http.MethodPost, "/fax/"+string(outboundID[2])+"/resend", nil, "")
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("resend of a transmitted fax: %d (want 422)", resp.StatusCode)
	}
	if !strings.Contains(string(body), "only failed outbound faxes can be resent") {
		t.Fatalf("422 body must carry the state reason: %.300s", body)
	}

	// Inbound (delivered by the webhook below): also 422.
	payload, _ := json.Marshal(map[string]any{
		"owner": "1001", "from": "+491700000000", "pages": 1,
		"pdf_base64": base64.StdEncoding.EncodeToString(pdf),
	})
	req, _ := http.NewRequest(http.MethodPost, server.URL+"/hooks/fax", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test-secret")
	if hookResp, err := server.Client().Do(req); err != nil || hookResp.StatusCode != http.StatusAccepted {
		t.Fatalf("inbound webhook: %v", err)
	}
	_, body = c.do(http.MethodGet, "/partials/fax", nil, "")
	inboundID := regexp.MustCompile(`wp-fax-in[^>]*.*?href="(/fax/([^/]+)/document)`).FindSubmatch(body)
	if inboundID == nil {
		t.Fatalf("no inbound row to test the guard with: %.300s", body)
	}
	resp, _ = c.do(http.MethodPost, "/fax/"+string(inboundID[2])+"/resend", nil, "")
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("resend of an inbound fax: %d (want 422)", resp.StatusCode)
	}

	// Another extension's job: owner-scoped 404, no information leak.
	other := clientFor(t, server)
	other.login("1002", "pw")
	resp, _ = other.do(http.MethodPost, "/fax/"+string(outboundID[2])+"/resend", nil, "")
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("resend of a foreign job: %d (want 404)", resp.StatusCode)
	}
}
