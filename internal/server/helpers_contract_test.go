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
	"github.com/larsartmann/webphone/internal/store"
)

// Micro-tests for the one-home helpers the 2026-09-22 dedup trains
// landed caller-pinned only (the avatarFor lesson: an untested first cut
// shipped a real bug). Each test pins the helper's OWN contract; the
// caller-level behavior stays pinned where it already was.

func applyToRecorder(h *handlers, kind, ref string, apply func(context.Context, string) error) *httptest.ResponseRecorder {
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
	h := &handlers{hooksIdem: newIdemStore(hookIdempotencyTTL)}

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

// TestRecordCallIdemContract pins the call-journal key recorder: an
// empty key (client sent none) records nothing — the legacy
// never-dedupe shape — and a non-empty key is recorded. The handler's
// use of these semantics (record only on 204 outcomes, 502 stays
// retryable, extension-namespaced keys) is pinned end-to-end by
// TestAPICallLoggingContract in crm_test.go.
func TestRecordCallIdemContract(t *testing.T) {
	h := &handlers{callsIdem: newIdemStore(callsIdempotencyTTL)}
	h.recordCallIdem("")
	if h.callsIdem.seen("") {
		t.Fatal("an empty key must never be recorded: keyless reports never dedupe")
	}
	h.recordCallIdem("1001:uuid-1")
	if !h.callsIdem.seen("1001:uuid-1") {
		t.Fatal("a recorded key must be seen")
	}
	if h.callsIdem.seen("1001:uuid-2") {
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
// JSON API's store-failure lane: a broken store surfaces as the shared
// 500 (save) and a plain 404 (delete never leaks store detail).
func TestContactsAPIStoreFailureRoutesThroughContactSaveFailed(t *testing.T) {
	db, err := store.Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	server := newTestServerWithPhoneAPI(t, "", func(d *Deps) {
		d.Contacts = store.NewContacts(db)
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

// TestContactsAPIListCapRefusesNewNumbers pins the atomic cap at the
// API surface: at ContactsMaxPerExtension the store refuses a NEW
// number with the actionable 422, while renaming an EXISTING number
// still succeeds (the upsert branch is not capped — one statement, one
// invariant).
func TestContactsAPIListCapRefusesNewNumbers(t *testing.T) {
	db, err := store.Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	contacts := store.NewContacts(db)
	owner := domain.MustParseExtension("1001")
	ctx := context.Background()
	for i := range store.ContactsMaxPerExtension {
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
