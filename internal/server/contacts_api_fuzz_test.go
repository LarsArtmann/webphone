// Fuzz target (plan T18, 2026-09-20): a hostile client bashing the
// contacts JSON API must never produce a 5xx — every malformed body is a
// 400 (decode) or 422 (bad number), and a valid body is 204. Fuzzing the
// full handler through the real router catches decode-layer panics and
// framework surprises, not just the happy path.
package server

import (
	"net/http"
	"testing"
)

func FuzzContactsAPISave(f *testing.F) {
	c := clientFor(f, newTestServerWithConfig(f, "", nil))
	c.login("1001", "pw")

	f.Add([]byte(`{"name":"Ada","number":"+441632960961"}`))
	f.Add([]byte(`{"name":"","number":""}`))
	f.Add([]byte(`{}`))
	f.Add([]byte(`not json`))
	f.Add([]byte(`{"name":"` + "\x00\x01\x7f" + `","number":"abc"}`))
	f.Add([]byte(`{"name":"overflow","number":"` + string(make([]byte, 0)) + `99999999999999999999"}`))

	f.Fuzz(func(t *testing.T, body []byte) {
		resp, _ := c.do(http.MethodPost, "/api/contacts", body, "application/json")
		switch resp.StatusCode {
		case http.StatusBadRequest, http.StatusUnprocessableEntity, http.StatusNoContent:
			// honest outcomes
		default:
			t.Fatalf("hostile body answered %d (want 400/422/204)", resp.StatusCode)
		}
	})
}
