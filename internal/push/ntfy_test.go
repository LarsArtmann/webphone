package push_test

import (
	"context"
	"encoding/json/v2"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/larsartmann/go-error-family"
	errorfamilytest "github.com/larsartmann/go-error-family/errorfamilytest"
	"github.com/larsartmann/webphone/internal/push"
)

func TestNtfySendPublishesRingJSON(t *testing.T) {
	var gotPath string
	var gotAuth string
	var gotBody map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotAuth = r.Header.Get("Authorization")
		raw, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("read body: %v", err)
		}
		if err := json.Unmarshal(raw, &gotBody); err != nil {
			t.Errorf("decode body: %v", err)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	sender, err := push.NewNtfySender(server.URL, "secret-token")
	if err != nil {
		t.Fatal(err)
	}
	ring := push.Ring{Event: push.EventRing, From: "+15550000001", CallID: "call-1"}
	if err := sender.Send(context.Background(), "wp-topic", ring); err != nil {
		t.Fatalf("send: %v", err)
	}

	if gotPath != "/" {
		t.Fatalf("path %q, want /", gotPath)
	}
	if gotAuth != "Bearer secret-token" {
		t.Fatalf("auth %q, want Bearer secret-token", gotAuth)
	}
	if gotBody["topic"] != "wp-topic" {
		t.Fatalf("topic %v, want wp-topic", gotBody["topic"])
	}
	if gotBody["priority"].(float64) != 5 {
		t.Fatalf("priority %v, want 5", gotBody["priority"])
	}
	message, _ := gotBody["message"].(string)
	if !strings.Contains(message, `"event":"ring"`) || !strings.Contains(message, `"call_id":"call-1"`) {
		t.Fatalf("message %q does not carry the ring JSON", message)
	}

	// ring_end rides the default priority.
	if err := sender.Send(context.Background(), "wp-topic", push.Ring{Event: push.EventRingEnd, CallID: "call-1"}); err != nil {
		t.Fatalf("send end: %v", err)
	}
	if gotBody["priority"].(float64) != 3 {
		t.Fatalf("ring_end priority %v, want 3", gotBody["priority"])
	}
}

func TestNtfySendClassifiesFailures(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()
	sender, err := push.NewNtfySender(server.URL, "bad")
	if err != nil {
		t.Fatal(err)
	}
	err = sender.Send(context.Background(), "wp-topic", push.Ring{Event: push.EventRing})
	if !errors.Is(err, push.ErrUnauthorized) {
		t.Fatalf("401 error = %v, want ErrUnauthorized", err)
	}
	errorfamilytest.AssertFamily(t, err, errorfamily.Rejection)
	errorfamilytest.AssertCode(t, err, "push.ntfy_send")
}

func TestNtfySenderDisabledWhenUnconfigured(t *testing.T) {
	sender, err := push.NewNtfySender("", "")
	if err != nil {
		t.Fatal(err)
	}
	if sender.Enabled() {
		t.Fatal("empty url must disable the sender")
	}
	if err := sender.Send(context.Background(), "t", push.Ring{}); !errors.Is(err, push.ErrDisabled) {
		t.Fatalf("error = %v, want ErrDisabled", err)
	}
}

func TestValidNtfyTopic(t *testing.T) {
	for _, topic := range []string{"wp-8f2a", "webphone-1234567890abcdef", "a"} {
		if !push.ValidNtfyTopic(topic) {
			t.Fatalf("topic %q should be valid", topic)
		}
	}
	for _, topic := range []string{"", "has space", "has/slash", strings.Repeat("x", 65)} {
		if push.ValidNtfyTopic(topic) {
			t.Fatalf("topic %q should be invalid", topic)
		}
	}
}

// TestNtfySendDraftPublishesDraftJSON pins the draft-ready lane's wire:
// the fact rides the message field as compact JSON, priority stays at
// the default (good news is not urgent), the title names the news for
// the stock app's shade.
func TestNtfySendDraftPublishesDraftJSON(t *testing.T) {
	var gotBody map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("read body: %v", err)
		}
		if err := json.Unmarshal(raw, &gotBody); err != nil {
			t.Errorf("decode body: %v", err)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	sender, err := push.NewNtfySender(server.URL, "")
	if err != nil {
		t.Fatal(err)
	}
	draft := push.Draft{Event: push.EventDraftReady, Contact: "+15550000002", DraftID: "d-9"}
	if err := sender.SendDraft(context.Background(), "wp-topic", draft); err != nil {
		t.Fatalf("send draft: %v", err)
	}

	if gotBody["title"] != "Draft ready" {
		t.Fatalf("title %v, want Draft ready", gotBody["title"])
	}
	if gotBody["priority"].(float64) != 3 {
		t.Fatalf("priority %v, want 3 (default)", gotBody["priority"])
	}
	message, ok := gotBody["message"].(string)
	if !ok {
		t.Fatalf("message %v is not a string", gotBody["message"])
	}
	var fact map[string]any
	if err := json.Unmarshal([]byte(message), &fact); err != nil {
		t.Fatalf("message is not JSON: %v (%s)", err, message)
	}
	if fact["event"] != "draft_ready" || fact["contact"] != "+15550000002" || fact["draft_id"] != "d-9" {
		t.Fatalf("draft fact mismatch: %v", fact)
	}
}
