package textagent

import (
	"context"
	"encoding/json/v2"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/larsartmann/go-error-family"
)

// stubAgent is a minimal review-API stand-in: it records every request
// (method, path, bearer, decoded JSON body) and answers from the
// scripted handler.
type stubAgent struct {
	server *httptest.Server
	got    chan stubRequest
}

type stubRequest struct {
	method string
	path   string
	bearer string
	body   map[string]any
}

func newStubAgent(t *testing.T, handle func(stubRequest) (int, string)) *stubAgent {
	t.Helper()
	stub := &stubAgent{got: make(chan stubRequest, 8)}
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		req := stubRequest{method: r.Method, path: r.URL.Path, bearer: r.Header.Get("Authorization")}
		if raw, readErr := io.ReadAll(r.Body); readErr == nil && len(raw) > 0 {
			_ = json.Unmarshal(raw, &req.body)
		}
		stub.got <- req
		status, body := handle(req)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	})
	stub.server = httptest.NewServer(mux)
	t.Cleanup(stub.server.Close)
	return stub
}

func (s *stubAgent) lastRequest(t *testing.T) stubRequest {
	t.Helper()
	select {
	case req := <-s.got:
		return req
	default:
		t.Fatal("stub agent saw no request")
		return stubRequest{}
	}
}

// secretFile writes a one-line secret the way the deployment pipeline
// does and returns its path.
func secretFile(t *testing.T, value string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "textagent_secret")
	if err := os.WriteFile(path, []byte(value+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func mustClient(t *testing.T, url string) *Client {
	t.Helper()
	client, err := New(url, secretFile(t, "review-secret"))
	if err != nil {
		t.Fatal(err)
	}
	return client
}

func TestNewRejectsMissingAndEmptySecret(t *testing.T) {
	if _, err := New("http://127.0.0.1:8071", filepath.Join(t.TempDir(), "absent")); err == nil {
		t.Error("a missing secret file must fail New")
	}
	if _, err := New("http://127.0.0.1:8071", secretFile(t, "   ")); err == nil {
		t.Error("an empty secret file must fail New")
	}
	if client := mustClient(t, "http://127.0.0.1:8071/"); client.base != "http://127.0.0.1:8071" {
		t.Errorf("trailing slash must be trimmed, got %q", client.base)
	}
}

const reviewsPayload = `{"drafts":[{"id":"d-1","contact":"4915112345678","trigger":"inbound",
"text":"Guten Morgen!","en_gloss":"Good morning!","language":"de","confidence":93,
"status":"pending","created_at":"2026-10-12T18:00:00Z"}],
"frozen":[{"contact":"491599999999","reason":"kill-switch"}],
"suppressed":[{"contact":"491588888888","since":"2026-10-01","keyword":"STOP"}],"mode":"drafts"}`

func TestReviewsRidesBearerAndDecodes(t *testing.T) {
	stub := newStubAgent(t, func(stubRequest) (int, string) { return http.StatusOK, reviewsPayload })
	client := mustClient(t, stub.server.URL)
	reviews, err := client.Reviews(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	req := stub.lastRequest(t)
	if req.bearer != "Bearer review-secret" {
		t.Errorf("bearer missing, got %q", req.bearer)
	}
	if len(reviews.Drafts) != 1 || reviews.Drafts[0].Text != "Guten Morgen!" || reviews.Drafts[0].Confidence != 93 {
		t.Errorf("draft decode drift: %+v", reviews.Drafts)
	}
	if len(reviews.Frozen) != 1 || len(reviews.Suppressed) != 1 || reviews.Mode != "drafts" {
		t.Errorf("review payload decode drift: %+v", reviews)
	}
}

func TestApproveSendsEmptyObjectVerbatimAndTextAsEdit(t *testing.T) {
	stub := newStubAgent(t, func(stubRequest) (int, string) {
		return http.StatusOK, `{"id":"d-1","status":"sent","provider_ref":"stub-1"}`
	})
	client := mustClient(t, stub.server.URL)
	if _, err := client.Approve(context.Background(), "d-1", ""); err != nil {
		t.Fatal(err)
	}
	if req := stub.lastRequest(t); len(req.body) != 0 {
		t.Errorf("verbatim approve must send an empty object, got %v", req.body)
	}
	if _, err := client.Approve(context.Background(), "d-1", "Edited"); err != nil {
		t.Fatal(err)
	}
	if req := stub.lastRequest(t); req.body["text"] != "Edited" {
		t.Errorf("edited approve must carry the text override, got %v", req.body)
	}
}

func TestSetModeAndResetMode(t *testing.T) {
	const modePayload = `{"mode":"full","mode_source":"runtime","default_mode":"drafts",
"modes":["off","drafts","after-hours","full"]}`
	stub := newStubAgent(t, func(stubRequest) (int, string) { return http.StatusOK, modePayload })
	client := mustClient(t, stub.server.URL)
	state, err := client.SetMode(context.Background(), "full")
	if err != nil {
		t.Fatal(err)
	}
	req := stub.lastRequest(t)
	if req.method != http.MethodPost || req.path != "/mode" {
		t.Errorf("mode switch must POST /mode, got %s %s", req.method, req.path)
	}
	if req.body["mode"] != "full" {
		t.Errorf("mode switch must carry the chosen mode, got %v", req.body)
	}
	if state.Mode != "full" || state.ModeSource != "runtime" || state.DefaultMode != "drafts" || len(state.Modes) != 4 {
		t.Errorf("mode state decode drift: %+v", state)
	}
	if _, err := client.ResetMode(context.Background()); err != nil {
		t.Fatal(err)
	}
	if req := stub.lastRequest(t); req.body["reset"] != true {
		t.Errorf("reset must send {\"reset\": true}, got %v", req.body)
	}
}

func TestDraftForCarriesToAndOptionalHint(t *testing.T) {
	stub := newStubAgent(t, func(stubRequest) (int, string) {
		return http.StatusOK, `{"id":"d-2","text":"Entwurf"}`
	})
	client := mustClient(t, stub.server.URL)
	if _, err := client.DraftFor(context.Background(), "+4915112345678", "hint text"); err != nil {
		t.Fatal(err)
	}
	req := stub.lastRequest(t)
	if req.path != "/draft" || req.body["to"] != "+4915112345678" || req.body["hint"] != "hint text" {
		t.Errorf("draft payload drift: %+v", req)
	}
}

func TestHTTPErrorCarriesStatusMessageAndFamily(t *testing.T) {
	stub := newStubAgent(t, func(stubRequest) (int, string) {
		return http.StatusServiceUnavailable, `{"error":"gemini key is a placeholder"}`
	})
	client := mustClient(t, stub.server.URL)
	_, err := client.Reviews(context.Background())
	if err == nil {
		t.Fatal("a 503 answer must fail")
	}
	agentErr, ok := errors.AsType[*Error](err)
	if !ok {
		t.Fatalf("error must be *textagent.Error, got %T", err)
	}
	if agentErr.Status != http.StatusServiceUnavailable || agentErr.Message != "gemini key is a placeholder" {
		t.Errorf("error drift: %+v", agentErr)
	}
	if agentErr.ErrorFamily() != errorfamily.Transient || agentErr.ErrorCode() != "textagent.http" {
		t.Errorf("family/code drift: %s / %s", agentErr.ErrorFamily(), agentErr.ErrorCode())
	}
	if !strings.Contains(err.Error(), "503") {
		t.Errorf("message must name the status: %q", err.Error())
	}
}

func TestUnreachableAgentIsTransient(t *testing.T) {
	dead := httptest.NewServer(http.NotFoundHandler())
	dead.Close() // port closed: every dial fails
	client := mustClient(t, dead.URL)
	_, err := client.Reviews(context.Background())
	if err == nil {
		t.Fatal("an unreachable agent must fail")
	}
	if errorfamily.Classify(err) != errorfamily.Transient {
		t.Errorf("transport failure must classify Transient, got %s (%v)", errorfamily.Classify(err), err)
	}
}

func TestDraftIDGuard(t *testing.T) {
	client := mustClient(t, "http://127.0.0.1:1")
	for _, id := range []string{"", "../reviews", "a/b", "a%2Fb"} {
		if _, err := client.Approve(context.Background(), id, ""); err == nil {
			t.Errorf("id %q must be rejected", id)
		}
		if err := client.Discard(context.Background(), id); err == nil {
			t.Errorf("id %q must be rejected", id)
		}
	}
}
