package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHistorySearchFiltersEntries(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// The login credential probe (voicemail summary) is answered
		// unconditionally: this stub exists to test history search, not
		// directory auth.
		if strings.HasPrefix(r.URL.RequestURI(), "/phone-api/voicemail/") && strings.HasSuffix(r.URL.RequestURI(), "/summary") {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"new":0,"old":0}`))
			return
		}
		if !strings.HasPrefix(r.URL.RequestURI(), "/phone-api/history") {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"entries":[
			{"context":"public","caller_id_number":"+441632960961","caller_id_name":"Alice","destination_number":"1001","start":"2026-09-18 09:00","billsec":30},
			{"context":"from-internal","caller_id_number":"1001","caller_id_name":"","destination_number":"+493012345678","start":"2026-09-18 10:00","billsec":12},
			{"context":"public","caller_id_number":"+491700000000","caller_id_name":"Bob","destination_number":"1001","start":"2026-09-18 11:00","billsec":0}
		]}`))
	}))
	t.Cleanup(upstream.Close)

	server := newTestServerWithPhoneAPI(t, upstream.URL)
	c := signIn(t, server)

	// Unfiltered: everything, capped view. (Rows render numbers, not
	// caller names.)
	_, body := c.do(http.MethodGet, "/partials/history", nil, "")
	page := string(body)
	for _, want := range []string{"+441632960961", "+493012345678", "+491700000000"} {
		if !strings.Contains(page, want) {
			t.Errorf("unfiltered history missing %q", want)
		}
	}

	// Substring over numbers and names.
	_, body = c.do(http.MethodGet, "/partials/history?q=4917", nil, "")
	page = string(body)
	if !strings.Contains(page, "+491700000000") || strings.Contains(page, "+441632960961") {
		t.Errorf("q=4917 must keep the Bob record, drop Alice: %.400s", page)
	}
	_, body = c.do(http.MethodGet, "/partials/history?q=alice", nil, "")
	page = string(body)
	if !strings.Contains(page, "+441632960961") || strings.Contains(page, "+491700000000") {
		t.Errorf("q=alice must match case-insensitively by name: %.400s", page)
	}

	// Direction filter.
	_, body = c.do(http.MethodGet, "/partials/history?dir=in", nil, "")
	page = string(body)
	if strings.Contains(page, "+493012345678") {
		t.Errorf("dir=in must drop the outbound record: %.400s", page)
	}
	_, body = c.do(http.MethodGet, "/partials/history?dir=out", nil, "")
	page = string(body)
	if !strings.Contains(page, "+493012345678") || strings.Contains(page, "+441632960961") {
		t.Errorf("dir=out must keep only the outbound record: %.400s", page)
	}

	// The form echoes the active filter back.
	_, body = c.do(http.MethodGet, "/partials/history?q=alice&dir=in", nil, "")
	page = string(body)
	if !strings.Contains(page, `value="alice"`) {
		t.Errorf("filter form must echo the query: %.400s", page)
	}
}
