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

// TestHistoryRowsOfferCallBack pins the dial affordance: every CDR row
// with a dialable number carries a data-dial button (the caller's CID
// number for inbound legs, the dialled destination for outbound ones),
// and rows without a number render no dead button.
func TestHistoryRowsOfferCallBack(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.RequestURI(), "/phone-api/voicemail/") {
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
			{"context":"public","caller_id_number":"","caller_id_name":"Withheld","destination_number":"1001","start":"2026-09-18 11:00","billsec":5}
		]}`))
	}))
	t.Cleanup(upstream.Close)

	server := newTestServerWithPhoneAPI(t, upstream.URL)
	c := signIn(t, server)

	_, body := c.do(http.MethodGet, "/partials/history", nil, "")
	page := string(body)
	for _, want := range []string{
		`data-dial="+441632960961"`, // inbound: call the caller back
		`data-dial="+493012345678"`, // outbound: redial the destination
	} {
		if !strings.Contains(page, want) {
			t.Errorf("history row missing %s: %.400s", want, page)
		}
	}
	if got := strings.Count(page, "data-dial="); got != 2 {
		t.Errorf("rows without a CID number must render no dial button: %d data-dial attributes", got)
	}
	// T20d/T20e: dialable rows also offer a prefilled SMS compose and
	// the save-as-contact star (with the resolved display name riding
	// along for the island's save).
	for _, want := range []string{
		`data-sms="+441632960961"`,
		`data-sms="+493012345678"`,
		`data-save-contact="+441632960961"`,
		`data-save-contact="+493012345678"`,
	} {
		if !strings.Contains(page, want) {
			t.Errorf("history row missing %s: %.400s", want, page)
		}
	}
	if !strings.Contains(page, `data-name="Alice"`) {
		t.Errorf("the CRM-resolved name should ride the save-as-contact gesture: %.400s", page)
	}
}

// TestHistoryOutcomeFilterAndDayGroups pins the D8 outcome filter, the
// D9 day grouping, the D10 missed styling, and the E2 URL-addressable
// filter state.
func TestHistoryOutcomeFilterAndDayGroups(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.RequestURI(), "/phone-api/voicemail/") {
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
			{"context":"public","caller_id_number":"+491700000000","caller_id_name":"Bob","destination_number":"1001","start":"2026-09-19 11:00","billsec":0},
			{"context":"from-internal","caller_id_number":"1001","caller_id_name":"","destination_number":"+493012345678","start":"2026-09-19 12:00","billsec":12}
		]}`))
	}))
	t.Cleanup(upstream.Close)

	server := newTestServerWithPhoneAPI(t, upstream.URL)
	c := signIn(t, server)

	// Missed: only the inbound zero-billsec record survives.
	_, body := c.do(http.MethodGet, "/partials/history?outcome=missed", nil, "")
	page := string(body)
	if !strings.Contains(page, "+491700000000") || strings.Contains(page, "+441632960961") || strings.Contains(page, "+493012345678") {
		t.Errorf("outcome=missed must keep only Bob's missed record: %.400s", page)
	}
	if !strings.Contains(page, `value="missed" selected`) {
		t.Errorf("the outcome select must echo the filter: %.400s", page)
	}

	// Answered drops both zero-billsec shapes (missed inbound AND any
	// unanswered leg).
	_, body = c.do(http.MethodGet, "/partials/history?outcome=answered", nil, "")
	page = string(body)
	if !strings.Contains(page, "+441632960961") || !strings.Contains(page, "+493012345678") || strings.Contains(page, "+491700000000") {
		t.Errorf("outcome=answered must keep only the two answered records: %.400s", page)
	}

	// Unfiltered: day groups render one head per distinct day, in order,
	// with the group count; the missed row carries the danger styling.
	_, body = c.do(http.MethodGet, "/partials/history", nil, "")
	page = string(body)
	if got := strings.Count(page, "wp-day-head"); got != 2 {
		t.Errorf("two distinct days must render two day heads, got %d: %.600s", got, page)
	}
	head1 := strings.Index(page, "wp-day-head")
	head2 := strings.Index(page[head1+1:], "wp-day-head") + head1 + 1
	if head1 >= head2 {
		t.Errorf("day heads must render in record order")
	}
	if !strings.Contains(page, "wp-row-missed") || !strings.Contains(page, "✖") {
		t.Errorf("the missed record must carry the missed row styling + glyph: %.600s", page)
	}

	// E2: the filter form pushes its URL — filters are addressable.
	if !strings.Contains(page, `hx-push-url="true"`) {
		t.Errorf("the filter form must push its request URL: %.400s", page)
	}
}

// TestHistoryEmptyStateIsFilterAware (M17 J8): an active outcome filter
// with zero matches must not claim nothing was ever recorded — the
// unfiltered copy would be a lie under a filter.
func TestHistoryEmptyStateIsFilterAware(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.RequestURI(), "/phone-api/voicemail/") {
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
			{"context":"public","caller_id_number":"+441632960961","caller_id_name":"Alice","destination_number":"1001","start":"2026-09-18 09:00","billsec":30}
		]}`))
	}))
	t.Cleanup(upstream.Close)

	server := newTestServerWithPhoneAPI(t, upstream.URL)
	c := signIn(t, server)

	_, body := c.do(http.MethodGet, "/partials/history?outcome=missed", nil, "")
	page := string(body)
	if !strings.Contains(page, "No calls match this filter.") || strings.Contains(page, "No calls recorded yet.") {
		t.Errorf("a filter with no matches must say so, not claim an empty history: %.400s", page)
	}
}

// TestHistoryDispositionFilter pins the T3 verdict filter: the select
// narrows the transcript section to calls whose note slot carries the
// confirmed verdict, the CDR section stays empty under an active
// disposition filter (CDR rows carry no call uuid to join the verdict
// with), the filter echoes back into the form, and garbage falls back
// to unfiltered.
func TestHistoryDispositionFilter(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.RequestURI(), "/phone-api/voicemail/") {
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
			{"context":"public","caller_id_number":"+449162000123","caller_id_name":"Cara","destination_number":"1001","start":"2026-09-18 09:00","billsec":30}
		]}`))
	}))
	t.Cleanup(upstream.Close)

	provider, _ := fakeProvider(t, http.StatusOK, "x")
	server := newTestServerWithPhoneAPI(t, upstream.URL, func(d *Deps) { d.ASR = asrClientFor(t, provider.URL) })
	c := signIn(t, server)

	if resp, body := postTranscript(t, c, "call-1", "in", "+4989123456", "first call"); resp.StatusCode != http.StatusNoContent {
		t.Fatalf("transcript 1: %d %s", resp.StatusCode, body)
	}
	if resp, body := postTranscript(t, c, "call-2", "in", "+4930123456", "second call"); resp.StatusCode != http.StatusNoContent {
		t.Fatalf("transcript 2: %d %s", resp.StatusCode, body)
	}
	if resp, body := confirmDisposition(t, c, "call-1", map[string]any{"disposition": "answered", "number": "+4989123456", "direction": "in"}); resp.StatusCode != http.StatusNoContent {
		t.Fatalf("confirm: %d %s", resp.StatusCode, body)
	}

	// answered: call-1 keeps its badge, the unconfirmed call-2 drops,
	// and no CDR row leaks through the verdict filter.
	_, body := c.do(http.MethodGet, "/partials/history?disposition=answered", nil, "")
	page := string(body)
	if !strings.Contains(page, "wp-disposition-chip") || !strings.Contains(page, ">Answered</span>") {
		t.Errorf("the confirmed call must render its badge: %.600s", page)
	}
	if strings.Contains(page, "+4930123456") {
		t.Errorf("the unconfirmed call must drop under disposition=answered: %.600s", page)
	}
	if strings.Contains(page, "+449162000123") {
		t.Errorf("no CDR row may render under an active disposition filter: %.600s", page)
	}
	if !strings.Contains(page, `value="answered" selected`) {
		t.Errorf("the disposition select must echo the filter: %.400s", page)
	}

	// missed: no call carries that verdict — the disposition-specific
	// empty state speaks, no transcript section renders.
	_, body = c.do(http.MethodGet, "/partials/history?disposition=missed", nil, "")
	page = string(body)
	if !strings.Contains(page, "No calls carry this verdict yet.") {
		t.Errorf("a verdict with no calls must render its own empty state: %.600s", page)
	}
	if strings.Contains(page, "wp-transcript-call") {
		t.Errorf("no transcript may render under a verdict no call carries: %.600s", page)
	}

	// Unfiltered: both transcripts render, but the badge appears exactly
	// once (call-2 was never confirmed — no glyph for it).
	_, body = c.do(http.MethodGet, "/partials/history", nil, "")
	page = string(body)
	if got := strings.Count(page, "wp-disposition-chip"); got != 1 {
		t.Errorf("only the confirmed call may carry a badge, got %d: %.600s", got, page)
	}
	if !strings.Contains(page, "+4930123456") {
		t.Errorf("the unfiltered view must keep both transcripts: %.600s", page)
	}

	// Garbage: a filter param that parses to nothing filters nothing.
	_, body = c.do(http.MethodGet, "/partials/history?disposition=picked-up", nil, "")
	page = string(body)
	if !strings.Contains(page, "+4930123456") || !strings.Contains(page, "+449162000123") {
		t.Errorf("an unknown disposition value must fall back to unfiltered: %.600s", page)
	}
}
