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
