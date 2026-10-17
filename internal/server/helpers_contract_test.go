package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/larsartmann/webphone/internal/domain"
	"github.com/larsartmann/webphone/internal/readmodels"
	"github.com/larsartmann/webphone/internal/server/hooks"
)

// Micro-tests for the one-home helpers the 2026-09-22 dedup trains
// landed caller-pinned only (the avatarFor lesson: an untested first cut
// shipped a real bug). Each test pins the helper's OWN contract; the
// caller-level behavior stays pinned where it already was. The
// applyStatusWebhook contract moved with the carve into
// internal/server/hooks (apply_status_test.go).

// TestRecordCallIdemContract pins the call-journal key recorder: an
// empty key (client sent none) records nothing — the legacy
// never-dedupe shape — and a non-empty key is recorded. The handler's
// use of these semantics (record only on 204 outcomes, 502 stays
// retryable, extension-namespaced keys) is pinned end-to-end by
// TestAPICallLoggingContract in crm_test.go.
func TestRecordCallIdemContract(t *testing.T) {
	h := &handlers{callsIdem: hooks.NewGate(nil, hooks.CallIdempotencyTTL)}
	h.recordCallIdem("")
	if h.callsIdem.Seen("") {
		t.Fatal("an empty key must never be recorded: keyless reports never dedupe")
	}
	h.recordCallIdem("1001:uuid-1")
	if !h.callsIdem.Seen("1001:uuid-1") {
		t.Fatal("a recorded key must be seen")
	}
	if h.callsIdem.Seen("1001:uuid-2") {
		t.Fatal("a different key must not be seen")
	}
}

// TestContactSaveFailedText pins the one-home failure surface: both
// contact-write homes answer this exact 500, so the island's toast and
// the tab's banner can never diverge in words.
func TestContactSaveFailedText(t *testing.T) {
	w := httptest.NewRecorder()
	contactSaveFailed(w)
	if w.Code != http.StatusInternalServerError || w.Body.String() != "could not save the contact\n" {
		t.Fatalf("contactSaveFailed: %d %q", w.Code, w.Body.String())
	}
}

// TestInternalErrorRedactsDetail pins the SafeDetail consistency
// contract (audit finding #3): a 500 body carries the op prefix and the
// family default message, NEVER the raw internal error text — while the
// panel variant (safeDetail) returns the same redacted copy for
// degraded page renders.
func TestInternalErrorRedactsDetail(t *testing.T) {
	h := &handlers{}
	internal := errors.New("sql: database is closed (file /var/lib/webphone/x.db)")
	r := httptest.NewRequest(http.MethodGet, "/partials/messages/t-1", nil)

	w := httptest.NewRecorder()
	h.internalError(w, r, "load conversation", internal)
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("internalError: %d (want 500)", w.Code)
	}
	// An unclassified error defaults to the transient family (the
	// classify-for-user precedent) — the user gets a retryable story,
	// never the raw store text.
	if got := w.Body.String(); got != "load conversation: A temporary error occurred. Please try again in a few moments.\n" {
		t.Fatalf("internalError body: %q", got)
	}
	if strings.Contains(w.Body.String(), "database is closed") {
		t.Fatal("internalError leaked the internal detail")
	}

	if got := h.safeDetail(r, "render tab", internal); got != "A temporary error occurred. Please try again in a few moments." {
		t.Fatalf("safeDetail: %q (want the family default, not the raw text)", got)
	}
}

// TestContactsAPIStoreFailureRoutesThroughContactSaveFailed pins the
// JSON API's store-failure lane: a broken read-model store (engine shut
// down) surfaces as the shared 500 (save) and a plain 404 (delete never
// leaks store detail).
func TestContactsAPIStoreFailureRoutesThroughContactSaveFailed(t *testing.T) {
	engine, err := readmodels.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	contacts := engine.Contacts()
	if err := engine.Shutdown(); err != nil {
		t.Fatal(err)
	}
	server := newTestServerWithPhoneAPI(t, "", func(d *Deps) {
		d.Contacts = contacts
	})
	c := signIn(t, server)

	resp, body := c.do(http.MethodPost, "/api/contacts",
		[]byte(`{"name":"Alice","number":"+441632960961"}`), "application/json")
	if resp.StatusCode != http.StatusInternalServerError || string(body) != "could not save the contact\n" {
		t.Fatalf("save over a broken store: %d %q (want the shared 500 text)", resp.StatusCode, body)
	}
	resp, _ = c.do(http.MethodDelete, "/api/contacts?id=whatever", nil, "")
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("delete over a broken store: %d (want a plain 404)", resp.StatusCode)
	}
}

// TestContactsAPIListCapRefusesNewNumbers pins the cap at the API
// surface: at ContactsMaxPerExtension the read-model store refuses a
// NEW number with the actionable 422, while renaming an EXISTING
// number still succeeds (the upsert branch is not capped).
func TestContactsAPIListCapRefusesNewNumbers(t *testing.T) {
	contacts := newTestReadmodels(t)
	owner := domain.MustParseExtension("1001")
	ctx := context.Background()
	for i := range readmodels.ContactsMaxPerExtension {
		phone, phoneErr := domain.ParsePhone(fmt.Sprintf("+441632960%03d", i))
		if phoneErr != nil {
			t.Fatal(phoneErr)
		}
		if err := contacts.Save(ctx, domain.Contact{
			ID: domain.GenerateContactID(), Owner: owner,
			Name: "Cap", Phone: phone, CreatedAt: time.Now(),
		}); err != nil {
			t.Fatalf("seed contact %d: %v", i, err)
		}
	}
	server := newTestServerWithPhoneAPI(t, "", func(d *Deps) { d.Contacts = contacts })
	c := signIn(t, server)

	resp, body := c.do(http.MethodPost, "/api/contacts",
		[]byte(`{"name":"Over","number":"+441632969999"}`), "application/json")
	if resp.StatusCode != http.StatusUnprocessableEntity || !strings.Contains(string(body), "contact list is full") {
		t.Fatalf("new number at cap: %d %q (want the actionable 422)", resp.StatusCode, body)
	}
	resp, _ = c.do(http.MethodPost, "/api/contacts",
		[]byte(`{"name":"Renamed","number":"+441632960000"}`), "application/json")
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("rename of an existing number at cap: %d (want 204 — the upsert is not capped)", resp.StatusCode)
	}
}

// TestCRMNumbersSkipsBlanks pins the page-collection helper: blank
// caller IDs (withheld numbers, CDRs without a dial target) never reach
// the resolver, and the extractor keeps it generic over row types.
func TestCRMNumbersSkipsBlanks(t *testing.T) {
	type row struct{ Number string }
	got := crmNumbers([]row{{"+441632960961"}, {""}, {"2000"}, {""}}, func(r row) string { return r.Number })
	if len(got) != 2 || got[0] != "+441632960961" || got[1] != "2000" {
		t.Fatalf("crmNumbers: %v (want the two non-blank numbers in order)", got)
	}
	if got := crmNumbers([]row{}, func(r row) string { return r.Number }); len(got) != 0 {
		t.Fatalf("crmNumbers on no rows: %v", got)
	}
}

// apiContactSaved: BOTH JSON mutations publish the payload-less
// "contacts" nudge (the island's dropdown re-fetches on it) before the
// TestContactsAPIMutationsNudgeThen204 pins the epilogue contract of
// apiContactSaved: BOTH JSON mutations publish the payload-less
// 204 lands — by the time the response is received the nudge is already
// out — and a REFUSED mutation (422) must not nudge at all.
func TestContactsAPIMutationsNudgeThen204(t *testing.T) {
	server := newTestServer(t)
	events := subscribeEvents(t, server)
	c := signIn(t, server)

	resp, body := c.do(http.MethodPost, "/api/contacts",
		[]byte(`{"name":"Alice","number":"+441632960961"}`), "application/json")
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("save: %d %s", resp.StatusCode, body)
	}
	if ev := expectEvent(t, events, "contacts"); ev.Data != "" {
		t.Fatalf("contacts nudge must be payload-less, got data %q", ev.Data)
	}

	id := listContacts(t, c).Personal[0].ID
	resp, _ = c.do(http.MethodDelete, "/api/contacts?id="+id, nil, "")
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("delete: %d", resp.StatusCode)
	}
	expectEvent(t, events, "contacts")

	resp, _ = c.do(http.MethodPost, "/api/contacts",
		[]byte(`{"name":"Bad","number":"???"}`), "application/json")
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("undialable number: %d (want 422)", resp.StatusCode)
	}
	assertNoEvent(t, events, "contacts")
}
