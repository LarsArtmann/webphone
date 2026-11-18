package app

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json/v2"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
	"time"
)

// httptestNew serves handler over a throwaway server, returning its
// base URL (the app tests' one-line boot).
func httptestNew(t testing.TB, handler http.Handler) string {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	return server.URL
}

// sseContext is the SSE reader's cancelable context.
func sseContext() (context.Context, func()) {
	return context.WithCancel(context.Background())
}

// The full composition-root cycle over one live app: the front-door
// login, a composer send through the real handler chain, an inbound
// webhook, and the SSE stream carrying the resulting nudge — the whole
// product loop WITHOUT the server-package harness, so a wiring that
// only the container could break (a missing Deps leg, a hub that never
// reaches the mux) fails here instead of in production.
func TestAppBootCycleLoginSendHookSSE(t *testing.T) {
	application := startApp(t, false)
	server := httptestNew(t, application.Handler)
	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatal(err)
	}
	httpClient := &http.Client{Jar: jar, Timeout: 5 * time.Second}
	var currentToken string
	token := func() string {
		t.Helper()
		resp, err := httpClient.Get(server + "/login")
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = resp.Body.Close() }()
		body, _ := io.ReadAll(resp.Body)
		match := regexp.MustCompile(`name="csrf-token" content="([^"]+)"`).FindSubmatch(body)
		if match == nil {
			t.Fatal("no csrf token on the login page")
		}
		return string(match[1])
	}
	post := func(path string, body []byte, contentType string) (*http.Response, []byte) {
		t.Helper()
		req, err := http.NewRequest(http.MethodPost, server+path, bytes.NewReader(body))
		if err != nil {
			t.Fatal(err)
		}
		if contentType != "" {
			req.Header.Set("Content-Type", contentType)
		}
		req.Header.Set("X-CSRF-Token", currentToken)
		resp, err := httpClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = resp.Body.Close() }()
		respBody, _ := io.ReadAll(resp.Body)
		return resp, respBody
	}

	// 1. Login through the front door, adopt the rotated token.
	currentToken = token()
	payload, _ := json.Marshal(map[string]string{"extension": "1001", "password": "pw"})
	resp, body := post("/api/session", payload, "application/json")
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("login: %d %s", resp.StatusCode, body)
	}
	adopt, err := http.NewRequest(http.MethodGet, server+"/api/csrf", nil)
	if err != nil {
		t.Fatal(err)
	}
	adoptResp, err := httpClient.Do(adopt)
	if err != nil {
		t.Fatal(err)
	}
	adoptBody, _ := io.ReadAll(adoptResp.Body)
	_ = adoptResp.Body.Close()
	var rotated struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(adoptBody, &rotated); err != nil || rotated.Token == "" {
		t.Fatalf("csrf adoption: %s %v", adoptBody, err)
	}
	currentToken = rotated.Token

	// 2. Open the SSE stream BEFORE the traffic — and wait for the
	// hub's `connected` greeting before sending anything: a subscription
	// that is still handshaking when the notifier publishes would MISS
	// the nudge (SSE events are not replayed), which the full-suite load
	// made a real race rather than a theoretical one.
	events := make(chan string, 16)
	sseCtx, stopSSE := sseContext()
	defer stopSSE()
	go func() {
		req, err := http.NewRequestWithContext(sseCtx, http.MethodGet, server+"/events", nil)
		if err != nil {
			return
		}
		resp, err := httpClient.Do(req)
		if err != nil {
			return
		}
		defer func() { _ = resp.Body.Close() }()
		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			if line := scanner.Text(); strings.HasPrefix(line, "event:") {
				events <- strings.TrimSpace(strings.TrimPrefix(line, "event:"))
			}
		}
	}()
	waitEvent := func(name string) {
		t.Helper()
		deadline := time.After(5 * time.Second)
		for {
			select {
			case got := <-events:
				if got == name {
					return
				}
			case <-deadline:
				t.Fatalf("SSE event %q never arrived on the live stream", name)
			}
		}
	}
	waitEvent("connected")

	// 3. A composer send through the real multipart lane.
	var form bytes.Buffer
	writer := multipart.NewWriter(&form)
	if err := writer.WriteField("to", "+441632960501"); err != nil {
		t.Fatal(err)
	}
	if err := writer.WriteField("body", "boot cycle send"); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	resp, body = post("/messages/send", form.Bytes(), writer.FormDataContentType())
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("send: %d %s", resp.StatusCode, body)
	}
	waitEvent("threads")

	// 4. The inbound webhook lands and the same stream nudges the panel.
	hookPayload, _ := json.Marshal(map[string]string{"owner": "1001", "from": "+441632960502", "body": "boot cycle inbound"})
	req, err := http.NewRequest(http.MethodPost, server+"/hooks/message", bytes.NewReader(hookPayload))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test-secret")
	hookResp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	_ = hookResp.Body.Close()
	if hookResp.StatusCode != http.StatusAccepted {
		t.Fatalf("inbound hook: %d", hookResp.StatusCode)
	}
	waitEvent("threads")
	waitEvent("thread")
}
