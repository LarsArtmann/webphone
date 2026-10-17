package hooks

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/larsartmann/webphone/internal/store"
)

func applyToRecorder(h *Handlers, kind, ref string, apply func(context.Context, string) error) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/hooks/"+kind+"/status", nil)
	h.applyStatusWebhook(w, r, kind, ref, apply)
	return w
}

// TestApplyStatusWebhookContract pins the shared verdict-hook tail: the
// 400 texts are byte-stable (a provider may log them), a replay answers
// 202 without re-applying, an unknown ref 404s with the kind named, a
// generic failure stays retryable (NOT recorded), and only a successful
// apply consumes the idempotency key.
func TestApplyStatusWebhookContract(t *testing.T) {
	h := &Handlers{idem: newIdemStore(hookIdempotencyTTL)}

	t.Run("empty ref is a 400 with the stable text", func(t *testing.T) {
		w := applyToRecorder(h, "message", "", func(context.Context, string) error {
			t.Fatal("apply must not run for an empty ref")
			return nil
		})
		if w.Code != http.StatusBadRequest || w.Body.String() != "provider_ref is required\n" {
			t.Fatalf("empty ref: %d %q", w.Code, w.Body.String())
		}
	})

	t.Run("success applies once and the replay is inert", func(t *testing.T) {
		calls := 0
		apply := func(context.Context, string) error { calls++; return nil }
		if w := applyToRecorder(h, "fax", "gw-1", apply); w.Code != http.StatusAccepted {
			t.Fatalf("first apply: %d", w.Code)
		}
		if w := applyToRecorder(h, "fax", "gw-1", apply); w.Code != http.StatusAccepted {
			t.Fatalf("replay: %d", w.Code)
		}
		if calls != 1 {
			t.Fatalf("apply ran %d times (want exactly 1: only successes are deduped)", calls)
		}
	})

	t.Run("unknown ref is a 404 that names the lane", func(t *testing.T) {
		w := applyToRecorder(h, "message", "gw-404", func(context.Context, string) error {
			return store.ErrNotFound
		})
		if w.Code != http.StatusNotFound {
			t.Fatalf("unknown ref: %d", w.Code)
		}
		if got := w.Body.String(); !strings.HasPrefix(got, "could not update message: ") {
			t.Fatalf("404 body %q must name the message lane", got)
		}
	})

	t.Run("generic failure stays retryable", func(t *testing.T) {
		boom := errors.New("disk on fire")
		calls := 0
		apply := func(context.Context, string) error { calls++; return boom }
		w := applyToRecorder(h, "fax", "gw-err", apply)
		if w.Code != http.StatusInternalServerError {
			t.Fatalf("failed apply: %d", w.Code)
		}
		if w.Body.Len() == 0 {
			t.Fatal("500 body must carry the redacted SafeDetail text, not be empty")
		}
		// The failed attempt must NOT have consumed the key: the retry
		// reaches apply again and can still succeed.
		retry := func(context.Context, string) error { calls++; return nil }
		if w := applyToRecorder(h, "fax", "gw-err", retry); w.Code != http.StatusAccepted {
			t.Fatalf("retry after failure: %d", w.Code)
		}
		if calls != 2 {
			t.Fatalf("apply ran %d times (want 2: failures stay retryable)", calls)
		}
	})
}
