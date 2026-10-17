package api

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/larsartmann/webphone/internal/domain"
	"github.com/larsartmann/webphone/internal/readmodels"
	"github.com/larsartmann/webphone/internal/session"
)

// contextSessionGate mirrors the wiring's session gate for tests that
// drive handlers directly: the session rides the request context, the
// 401 wording stays session.SignInFirst (one home, in the session
// package).
type contextSessionGate struct{}

func (contextSessionGate) RequireSession(w http.ResponseWriter, r *http.Request) (session.Session, bool) {
	sess, ok := session.From(r.Context())
	if !ok {
		http.Error(w, session.SignInFirst, http.StatusUnauthorized)
		return session.Session{}, false
	}
	return sess, true
}

func (contextSessionGate) NotifyContactsChanged(domain.Extension) {}

// Fuzz target (plan T18, 2026-09-20): a hostile client bashing the
// contacts JSON API must never produce a 5xx — every malformed body is a
// 400 (decode) or 422 (bad number), and a valid body is 204. The handler
// is driven DIRECTLY over a session-carrying request context: fuzz
// workers forbid the HTTP harness's t.Helper, and this keeps each
// iteration cheap (one throwaway read-model engine, no server boot).
func FuzzContactsAPISave(f *testing.F) {
	f.Add([]byte(`{"name":"Ada","number":"+441632960961"}`))
	f.Add([]byte(`{"name":"","number":""}`))
	f.Add([]byte(`{}`))
	f.Add([]byte(`not json`))
	f.Add([]byte(`{"name":"control-chars","number":"a\x01b"}`))
	f.Add([]byte(`{"name":"overflow","number":"+9999999999999999999999"}`))

	f.Fuzz(func(t *testing.T, body []byte) {
		engine, err := readmodels.Open(t.TempDir())
		if err != nil {
			t.Fatal(err)
		}
		h := New(Deps{Server: contextSessionGate{}, Contacts: engine.Contacts()})

		sess := session.Session{Extension: domain.MustParseExtension("1001")}
		req := httptest.NewRequest("POST", "/api/contacts",
			io.NopCloser(strings.NewReader(string(body))))
		req = req.WithContext(session.With(t.Context(), sess))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		h.SaveContact(rec, req)

		switch code := rec.Code; code {
		case 400, 422, 204:
			// honest outcomes
		default:
			t.Fatalf("hostile body answered %d (want 400/422/204)", code)
		}
	})
}
