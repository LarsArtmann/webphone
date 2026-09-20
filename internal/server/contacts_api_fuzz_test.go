// Fuzz target (plan T18, 2026-09-20): a hostile client bashing the
// contacts JSON API must never produce a 5xx — every malformed body is a
// 400 (decode) or 422 (bad number), and a valid body is 204. The handler
// is driven DIRECTLY over a session-carrying request context: fuzz
// workers forbid the HTTP harness's t.Helper, and this keeps each
// iteration cheap (one :memory: store, no server boot).
package server

import (
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/larsartmann/webphone/internal/domain"
	"github.com/larsartmann/webphone/internal/session"
	"github.com/larsartmann/webphone/internal/store"
)

func FuzzContactsAPISave(f *testing.F) {
	f.Add([]byte(`{"name":"Ada","number":"+441632960961"}`))
	f.Add([]byte(`{"name":"","number":""}`))
	f.Add([]byte(`{}`))
	f.Add([]byte(`not json`))
	f.Add([]byte(`{"name":"control-chars","number":"a\x01b"}`))
	f.Add([]byte(`{"name":"overflow","number":"+9999999999999999999999"}`))

	f.Fuzz(func(t *testing.T, body []byte) {
		db, err := store.Open(":memory:")
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = db.Close() })
		h := &handlers{deps: Deps{Contacts: store.NewContacts(db)}}

		sess := session.Session{Extension: domain.MustParseExtension("1001")}
		req := httptest.NewRequest("POST", "/api/contacts",
			io.NopCloser(strings.NewReader(string(body))))
		req = req.WithContext(session.With(t.Context(), sess))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		h.apiSaveContact(rec, req)

		switch code := rec.Code; code {
		case 400, 422, 204:
			// honest outcomes
		default:
			t.Fatalf("hostile body answered %d (want 400/422/204)", code)
		}
	})
}
