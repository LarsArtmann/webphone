package server

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/larsartmann/webphone/internal/textagent"
)

// stubReviewAPI is the review-API stand-in the handlers' tests ride on:
// scripted status + payload per route, a call log the assertions read.
type stubReviewAPI struct {
	server  *httptest.Server
	calls   atomic.Int32
	status  atomic.Int32
	payload string
}

func newStubReviewAPI(t *testing.T) *stubReviewAPI {
	t.Helper()
	stub := &stubReviewAPI{payload: `{"drafts":[{"id":"d-7","contact":"441632960961",
		"trigger":"inbound","text":"Guten Morgen","en_gloss":"Good morning",
		"language":"de","confidence":93,"status":"pending",
		"created_at":"2026-10-12T18:00:00Z"}],"frozen":[],"suppressed":[],"mode":"drafts"}`}
	stub.status.Store(http.StatusOK)
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		stub.calls.Add(1)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(int(stub.status.Load()))
		switch {
		case strings.HasSuffix(r.URL.Path, "/approve"):
			_, _ = w.Write([]byte(`{"id":"d-7","status":"sent","provider_ref":"stub-1"}`))
		case strings.HasSuffix(r.URL.Path, "/discard"):
			_, _ = w.Write([]byte(`{"ok":true}`))
		default:
			_, _ = w.Write([]byte(stub.payload))
		}
	})
	stub.server = httptest.NewServer(mux)
	t.Cleanup(stub.server.Close)
	return stub
}

// agentClientFor wires a real seam client onto the given review-API URL.
func agentClientFor(t *testing.T, url string) *textagent.Client {
	t.Helper()
	secret := filepath.Join(t.TempDir(), "secret")
	if err := os.WriteFile(secret, []byte("review-secret\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	client, err := textagent.New(url, secret)
	if err != nil {
		t.Fatal(err)
	}
	return client
}

func clientWithAgent(t *testing.T, stub *stubReviewAPI) *client {
	t.Helper()
	server := newTestServerWithPhoneAPI(t, "", func(deps *Deps) {
		deps.TextAgent = agentClientFor(t, stub.server.URL)
	})
	return clientFor(t, server)
}

// TestTextAgentSeamOffIsInvisible: a nil client must leave NO surface —
// the routes answer 404 and the Messages tab renders neither the
// drafts panel nor the AI compose button.
func TestTextAgentSeamOffIsInvisible(t *testing.T) {
	c := newClient(t)
	c.login("1001", "pw")
	for path := range map[string]struct{} {
		"/partials/messages/drafts": {},
	} {
		if resp, _ := c.do(http.MethodGet, path, nil, ""); resp.StatusCode != http.StatusNotFound {
			t.Errorf("GET %s must 404 with the seam off, got %d", path, resp.StatusCode)
		}
	}
	form, contentType := multipartBody(t, map[string]string{"to": "+441632960961", "body": "x"}, nil)
	for path := range map[string]struct{} {
		"/messages/drafts/d-7/approve": {},
		"/messages/drafts/d-7/discard": {},
		"/messages/ai-draft":           {},
	} {
		if resp, _ := c.do(http.MethodPost, path, form, contentType); resp.StatusCode != http.StatusNotFound {
			t.Errorf("POST %s must 404 with the seam off, got %d", path, resp.StatusCode)
		}
	}
	_, body := c.do(http.MethodGet, "/partials/messages", nil, "")
	page := string(body)
	if strings.Contains(page, "wp-drafts") || strings.Contains(page, "wp-ai-draft") {
		t.Error("the seam-off Messages tab must render no drafts panel and no AI button")
	}
}

// TestDraftsPanelRidesTheList: the panel mounts inside the Messages
// tab, polls its own partial, and renders the pending draft honestly.
func TestDraftsPanelRidesTheList(t *testing.T) {
	stub := newStubReviewAPI(t)
	c := clientWithAgent(t, stub)
	c.login("1001", "pw")

	_, body := c.do(http.MethodGet, "/partials/messages", nil, "")
	page := string(body)
	for _, marker := range []string{`id="wp-drafts"`, "Guten Morgen", `id="wp-draft-7"`, `id="wp-draft-count"`, "93%", "Good morning", `class="wp-ai-draft"`, `hx-post="/messages/drafts/d-7/approve"`} {
		if !strings.Contains(page, marker) {
			t.Errorf("messages tab missing %q", marker)
		}
	}

	resp, panel := c.do(http.MethodGet, "/partials/messages/drafts", nil, "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("drafts partial: %d %s", resp.StatusCode, panel)
	}
	if !strings.Contains(string(panel), `id="wp-drafts"`) {
		t.Error("drafts partial must render the section for the outerHTML poll swap")
	}
	if stub.calls.Load() < 2 {
		t.Errorf("each surface must fetch its own review list, saw %d calls", stub.calls.Load())
	}
}

// TestApproveRoundTripRefreshesThePanel: the approve action rides the
// stub's gateway verdict and answers with the refreshed (now empty)
// drafts panel.
func TestApproveRoundTripRefreshesThePanel(t *testing.T) {
	stub := newStubReviewAPI(t)
	c := clientWithAgent(t, stub)
	c.login("1001", "pw")

	resp, panel := c.do(http.MethodPost, "/messages/drafts/d-7/approve", nil, "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("approve: %d %s", resp.StatusCode, panel)
	}
	if !strings.Contains(string(panel), `id="wp-drafts"`) {
		t.Error("approve must answer the refreshed drafts panel")
	}
	if stub.calls.Load() < 2 {
		t.Errorf("approve must hit the review API (action + refresh), saw %d calls", stub.calls.Load())
	}
}

// TestDiscardRoundTrip: same contract for the discard button.
func TestDiscardRoundTrip(t *testing.T) {
	stub := newStubReviewAPI(t)
	c := clientWithAgent(t, stub)
	c.login("1001", "pw")

	resp, panel := c.do(http.MethodPost, "/messages/drafts/d-7/discard", nil, "")
	if resp.StatusCode != http.StatusOK || !strings.Contains(string(panel), `id="wp-drafts"`) {
		t.Fatalf("discard: %d %s", resp.StatusCode, panel)
	}
}

// TestAIDraftPrefillsTheComposer: the on-demand draft lands in the
// new-conversation composer; the agent sees the number plus the typed
// hint.
func TestAIDraftPrefillsTheComposer(t *testing.T) {
	stub := newStubReviewAPI(t)
	c := clientWithAgent(t, stub)
	c.login("1001", "pw")

	form, contentType := multipartBody(t, map[string]string{"to": "+441632960961", "body": "ask about Tuesday"}, nil)
	resp, panel := c.do(http.MethodPost, "/messages/ai-draft", form, contentType)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("ai-draft: %d %s", resp.StatusCode, panel)
	}
	if !strings.Contains(string(panel), "Guten Morgen") {
		t.Error("the drafted text must prefill the composer")
	}
}

// TestAIDraftThreadPathPrefillsTheReplyBox: a thread-scoped ask re-opens
// the conversation with the draft in the reply composer.
func TestAIDraftThreadPathPrefillsTheReplyBox(t *testing.T) {
	stub := newStubReviewAPI(t)
	c := clientWithAgent(t, stub)
	c.login("1001", "pw")

	send, sendType := multipartBody(t, map[string]string{"to": "+441632960961", "body": "hello"}, nil)
	if resp, body := c.do(http.MethodPost, "/messages/send", send, sendType); resp.StatusCode != http.StatusOK {
		t.Fatalf("send: %d %s", resp.StatusCode, body)
	}
	_, list := c.do(http.MethodGet, "/partials/messages", nil, "")
	match := regexp.MustCompile(`href="(/messages/[^"]+)"`).FindSubmatch(list)
	if match == nil {
		t.Fatal("no thread link in list")
	}
	threadPath := string(match[1])

	form, contentType := multipartBody(t, map[string]string{"to": "+441632960961", "body": ""}, nil)
	resp, view := c.do(http.MethodPost, "/messages/ai-draft?thread="+strings.TrimPrefix(threadPath, "/messages/"), form, contentType)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("ai-draft thread path: %d %s", resp.StatusCode, view)
	}
	page := string(view)
	if !strings.Contains(page, "Guten Morgen") || !strings.Contains(page, `id="thread-transcript"`) {
		t.Error("the thread view must re-render with the draft in the reply composer")
	}
}

// TestAgentOutageDegradesHonestly: a 503 review API renders the
// unavailable line instead of the list, and a failed approve rides the
// panel-error contract (status + banner) instead of a silent miss.
func TestAgentOutageDegradesHonestly(t *testing.T) {
	stub := newStubReviewAPI(t)
	stub.status.Store(http.StatusServiceUnavailable)
	c := clientWithAgent(t, stub)
	c.login("1001", "pw")

	_, body := c.do(http.MethodGet, "/partials/messages", nil, "")
	if !strings.Contains(string(body), "AI drafts are unavailable") {
		t.Error("an unreachable agent must render the unavailable line, not an empty panel")
	}
	resp, panel := c.do(http.MethodPost, "/messages/drafts/d-7/approve", nil, "")
	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("approve against a fail-closed agent must carry the agent status, got %d", resp.StatusCode)
	}
	if !strings.Contains(string(panel), `class="wp-error"`) {
		t.Error("the failed action must land the visible error banner")
	}
}
