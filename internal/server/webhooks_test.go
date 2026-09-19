package server

import (
	"bytes"
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

func TestMessageStatusWebhookUpdatesTranscript(t *testing.T) {
	server := newTestServer(t)

	// Seed an outbound SMS through the loopback gateway; its receipt
	// carries the provider_ref the status callback must quote.
	c := clientFor(t, server)
	payload, _ := json.Marshal(map[string]string{"extension": "1001", "password": "pw"})
	c.do(http.MethodPost, "/api/session", payload, "application/json")
	form, contentType := multipartBody(t, map[string]string{"to": "+441632960961", "body": "receipt test"}, nil)
	if resp, body := c.do(http.MethodPost, "/messages/send", form, contentType); resp.StatusCode != http.StatusOK {
		t.Fatalf("sms send: %d %s", resp.StatusCode, body)
	}
	owner, err := domain.ParseExtension("1001")
	if err != nil {
		t.Fatal(err)
	}
	threads, err := server.messages.ListThreads(context.Background(), owner)
	if err != nil || len(threads) != 1 {
		t.Fatalf("seeded threads: %d (err %v)", len(threads), err)
	}
	msgs, err := server.messages.ListMessages(context.Background(), owner, threads[0].Thread.ID, 10)
	if err != nil || len(msgs) != 1 || msgs[0].ProviderRef == "" {
		t.Fatalf("seeded messages: %d (err %v)", len(msgs), err)
	}
	_, listBody := c.do(http.MethodGet, "/partials/messages", nil, "")
	match := regexp.MustCompile(`href="(/messages/[^"]+)"`).FindSubmatch(listBody)
	if match == nil {
		t.Fatal("no thread link in list")
	}

	// The provider's delivered verdict flips the badge the transcript shows.
	statusBody, _ := json.Marshal(map[string]string{
		"provider_ref": msgs[0].ProviderRef, "status": "delivered",
	})
	req, _ := http.NewRequest(http.MethodPost, server.URL+"/hooks/message/status", bytes.NewReader(statusBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test-secret")
	resp, err := server.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusAccepted {
		t.Fatalf("message status hook: %d (want 202)", resp.StatusCode)
	}

	_, body := c.do(http.MethodGet, "/partials"+string(match[1]), nil, "")
	if !strings.Contains(string(body), "delivered") {
		t.Fatalf("transcript missing delivered badge: %.300s", body)
	}

	// Unknown refs, empty refs, and non-verdict statuses are rejected.
	for name, want := range map[string]int{
		"unknown ref":   http.StatusNotFound,
		"empty ref":     http.StatusBadRequest,
		"sent status":   http.StatusBadRequest,
		"queued status": http.StatusBadRequest,
	} {
		ref := "nope"
		status := "delivered"
		switch name {
		case "empty ref":
			ref = ""
		case "sent status":
			status = "sent"
		case "queued status":
			status = "queued"
		}
		payload, _ := json.Marshal(map[string]string{"provider_ref": ref, "status": status})
		req, _ = http.NewRequest(http.MethodPost, server.URL+"/hooks/message/status", bytes.NewReader(payload))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer test-secret")
		if resp, _ := server.Client().Do(req); resp.StatusCode != want {
			t.Fatalf("%s: %d (want %d)", name, resp.StatusCode, want)
		}
	}
}

func TestFaxStatusWebhookUpdatesJob(t *testing.T) {
	server := newTestServer(t)

	// Seed an outbound fax through the loopback gateway; its receipt
	// carries the provider_ref the status callback must quote.
	c := clientFor(t, server)
	payload, _ := json.Marshal(map[string]string{"extension": "1001", "password": "pw"})
	c.do(http.MethodPost, "/api/session", payload, "application/json")
	pdf := []byte("%PDF-1.4 status\n%%EOF\n")
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

	// The provider's failure verdict flips the job and explains why.
	statusBody, _ := json.Marshal(map[string]string{
		"provider_ref": jobs[0].ProviderRef, "status": "failed", "error": "remote hung up",
	})
	req, _ := http.NewRequest(http.MethodPost, server.URL+"/hooks/fax/status", bytes.NewReader(statusBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test-secret")
	resp, err := server.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusAccepted {
		t.Fatalf("fax status hook: %d (want 202)", resp.StatusCode)
	}

	_, body := c.do(http.MethodGet, "/partials/fax", nil, "")
	if !strings.Contains(string(body), "failed") || !strings.Contains(string(body), "remote hung up") {
		t.Fatalf("fax panel missing failure: %.300s", body)
	}

	// Unknown refs, empty refs, and invalid statuses are rejected.
	badRef, _ := json.Marshal(map[string]string{"provider_ref": "nope", "status": "failed"})
	req, _ = http.NewRequest(http.MethodPost, server.URL+"/hooks/fax/status", bytes.NewReader(badRef))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test-secret")
	if resp, _ := server.Client().Do(req); resp.StatusCode != http.StatusNotFound {
		t.Fatalf("unknown provider_ref: %d (want 404)", resp.StatusCode)
	}
	emptyRef, _ := json.Marshal(map[string]string{"provider_ref": "", "status": "failed"})
	req, _ = http.NewRequest(http.MethodPost, server.URL+"/hooks/fax/status", bytes.NewReader(emptyRef))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test-secret")
	if resp, _ := server.Client().Do(req); resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("empty provider_ref: %d (want 400, parity with message hook)", resp.StatusCode)
	}
	badStatus, _ := json.Marshal(map[string]string{"provider_ref": jobs[0].ProviderRef, "status": "queued"})
	req, _ = http.NewRequest(http.MethodPost, server.URL+"/hooks/fax/status", bytes.NewReader(badStatus))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test-secret")
	if resp, _ := server.Client().Do(req); resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("invalid status: %d (want 400)", resp.StatusCode)
	}
}

func TestWebhooksSecretAndInbound(t *testing.T) {
	c := newClient(t)
	server := newTestServer(t)
	c.base = server.URL

	inbound := map[string]any{"owner": "1001", "from": "+4915112345678", "body": "hooked"}
	payload, _ := json.Marshal(inbound)

	req, _ := http.NewRequest(http.MethodPost, server.URL+"/hooks/message", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	if badResp, _ := c.http.Do(req); badResp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("webhook without secret: %d (want 401)", badResp.StatusCode)
	}

	req, _ = http.NewRequest(http.MethodPost, server.URL+"/hooks/message", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer wrong")
	if badResp, _ := c.http.Do(req); badResp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("webhook with wrong secret: %d", badResp.StatusCode)
	}

	// Right secret stores the message; reading it back needs a session.
	req, _ = http.NewRequest(http.MethodPost, server.URL+"/hooks/message", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test-secret")
	if okResp, _ := c.http.Do(req); okResp.StatusCode != http.StatusAccepted {
		t.Fatalf("webhook with secret: status %d", okResp.StatusCode)
	}
}
