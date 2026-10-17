package hooks

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestWebhookFailRedactsInternalDetail pins the redaction contract: a
// provider-facing 5xx must not echo internal error text (store paths, SQL,
// filesystem state). The full detail belongs in the server log only; the
// body carries the library's redacted SafeDetail text.
func TestWebhookFailRedactsInternalDetail(t *testing.T) {
	rec := httptest.NewRecorder()
	secret := errors.New("secret internal state: attachments table locked")
	webhookFail(rec, "test", secret)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("webhookFail status %d, want 500", rec.Code)
	}
	if body := rec.Body.String(); strings.Contains(body, "secret internal state") {
		t.Errorf("internal detail leaked to the webhook body: %q", body)
	}
	if body := rec.Body.String(); strings.TrimSpace(body) == "" {
		t.Error("redacted body must not be empty")
	}
}
