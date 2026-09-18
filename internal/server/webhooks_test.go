package server

import (
	"context"
	"encoding/json/v2"
	"net/http"
	"regexp"
	"strings"
	"testing"

	"github.com/larsartmann/webphone/internal/domain"
)

// signIn opens a tab session for extension 1001 on a fresh test server.
func signIn(t *testing.T, server *testServer) *client {
	t.Helper()
	c := clientFor(t, server)
	payload, err := json.Marshal(map[string]string{"extension": "1001", "password": "pw"})
	if err != nil {
		t.Fatal(err)
	}
	resp, body := c.do(http.MethodPost, "/api/session", payload, "application/json")
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("session create: %d %s", resp.StatusCode, body)
	}
	return c
}

func unreadBadge(t *testing.T, c *client) string {
	t.Helper()
	resp, body := c.do(http.MethodGet, "/", nil, "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("shell render: %d", resp.StatusCode)
	}
	match := regexp.MustCompile(`class="wp-nav-badge"[^>]*>(\d+)<`).FindSubmatch(body)
	if match == nil {
		return ""
	}
	return string(match[1])
}

func deliverInbound(t *testing.T, server *testServer, from, bodyText string) {
	t.Helper()
	payload, err := json.Marshal(map[string]any{
		"owner": "1001", "from": from, "body": bodyText,
	})
	if err != nil {
		t.Fatal(err)
	}
	req, err := http.NewRequest(http.MethodPost, server.URL+"/hooks/message", strings.NewReader(string(payload)))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Authorization", "Bearer test-secret")
	req.Header.Set("Content-Type", "application/json")
	resp, err := server.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusAccepted {
		t.Fatalf("inbound webhook: %d", resp.StatusCode)
	}
}

func postHook(t *testing.T, server *testServer, path, payload string) (int, string) {
	t.Helper()
	req, err := http.NewRequest(http.MethodPost, server.URL+path, strings.NewReader(payload))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Authorization", "Bearer test-secret")
	req.Header.Set("Content-Type", "application/json")
	resp, err := server.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()
	return resp.StatusCode, string(readAll(t, resp))
}

func TestUnreadBadgeCachesAndInvalidates(t *testing.T) {
	server := newTestServer(t)
	c := signIn(t, server)

	if badge := unreadBadge(t, c); badge != "" {
		t.Fatalf("fresh session must have no badge, got %q", badge)
	}

	// The webhook lands AFTER the shell render cached 0: only the
	// invalidation (not a re-render) can surface it.
	deliverInbound(t, server, "+441632960961", "ping")
	if badge := unreadBadge(t, c); badge != "1" {
		t.Fatalf("badge after inbound: got %q, want 1", badge)
	}

	// Opening the thread marks it read; the badge must drop without
	// waiting for the TTL.
	threadRow := regexp.MustCompile(`hx-get="/partials/messages/([^"]+)"`)
	resp, body := c.do(http.MethodGet, "/partials/messages", nil, "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("threads partial: %d", resp.StatusCode)
	}
	match := threadRow.FindSubmatch(body)
	if match == nil {
		t.Fatal("no thread row rendered")
	}
	resp, _ = c.do(http.MethodGet, "/partials/messages/"+string(match[1]), nil, "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("thread partial: %d", resp.StatusCode)
	}
	if badge := unreadBadge(t, c); badge != "" {
		t.Fatalf("badge after mark-read: got %q, want none", badge)
	}
}

func TestFaxStatusWebhookParsesFlexiblePageCounts(t *testing.T) {
	cases := []struct {
		name      string
		payload   string
		wantPages string
	}{
		{name: "numeric pages", payload: `"pages":3`, wantPages: "3"},
		{name: "string pages", payload: `"pages":"4"`, wantPages: "4"},
		{name: "page_count alias", payload: `"page_count":5`, wantPages: "5"},
		{name: "num_pages alias as string", payload: `"num_pages":"6"`, wantPages: "6"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			server := newTestServer(t)
			c := signIn(t, server)

			// Seed an outbound fax job; the loopback receipt's provider
			// ref is what the status webhook must quote.
			pdf := []byte("%PDF-1.4 pages\n%%EOF\n")
			form, contentType := multipartBody(t,
				map[string]string{"to": "+441632960961"},
				map[string]struct {
					Name    string
					Content []byte
				}{"document": {Name: "doc.pdf", Content: pdf}})
			if resp, body := c.do(http.MethodPost, "/fax/send", form, contentType); resp.StatusCode != http.StatusOK {
				t.Fatalf("fax send: %d %s", resp.StatusCode, body)
			}
			owner, err := domain.ParseExtension("1001")
			if err != nil {
				t.Fatal(err)
			}
			jobs, err := server.faxes.List(context.Background(), owner, 10)
			if err != nil || len(jobs) != 1 {
				t.Fatalf("seeded fax jobs: %d (err %v)", len(jobs), err)
			}

			verdict := `{"provider_ref":"` + jobs[0].ProviderRef + `","status":"transmitted",` + tc.payload + `}`
			if got, body := postHook(t, server, "/hooks/fax/status", verdict); got != http.StatusAccepted {
				t.Fatalf("fax status hook: %d %s", got, body)
			}

			_, panel := c.do(http.MethodGet, "/partials/fax", nil, "")
			want := tc.wantPages + " page(s)"
			if !strings.Contains(string(panel), want) {
				t.Fatalf("fax panel missing %q: %.300s", want, panel)
			}

			// A garbage page count must be a clean 400, not a panic.
			broken := `{"provider_ref":"` + jobs[0].ProviderRef + `","status":"transmitted","pages":"many"}`
			if got, _ := postHook(t, server, "/hooks/fax/status", broken); got != http.StatusBadRequest {
				t.Fatalf("garbage page count: %d (want 400)", got)
			}
		})
	}
}

func TestInboundFaxWebhookAcceptsFlexiblePageCounts(t *testing.T) {
	cases := []struct{ name, pagesField string }{
		{name: "numeric", pagesField: `"pages":2`},
		{name: "string", pagesField: `"pages":"3"`},
		{name: "alias", pagesField: `"page_count":"2"`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			server := newTestServer(t)
			payload := `{"owner":"1001","from":"+441632960961",` + tc.pagesField + `,` +
				`"pdf_base64":"JVBERi0xLjQKJSVFPRo="}`
			if got, body := postHook(t, server, "/hooks/fax", payload); got != http.StatusAccepted {
				t.Fatalf("inbound fax hook: %d %s", got, body)
			}
			c := signIn(t, server)
			_, panel := c.do(http.MethodGet, "/partials/fax", nil, "")
			if !strings.Contains(string(panel), "page(s)") {
				t.Fatalf("fax panel missing page count: %.300s", panel)
			}
		})
	}
}
