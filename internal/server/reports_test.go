package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/larsartmann/webphone/internal/config"
)

// reportsStub serves the login credential probe plus a fixed CDR page.
// The starts are generated RELATIVE to now so the "this week" assertions
// hold whenever the suite runs — a hardcoded 2026 date would silently
// drift out of the week window and turn a real check into a no-op.
func reportsStub(t *testing.T) *httptest.Server {
	t.Helper()

	now := time.Now()
	today := now.Format("2006-01-02")
	yesterday := now.AddDate(0, 0, -1).Format("2006-01-02")

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
			{"context":"public","caller_id_number":"+441632960961","caller_id_name":"Alice","destination_number":"1001","start":"` + yesterday + ` 09:00","billsec":120},
			{"context":"public","caller_id_number":"+441632960961","caller_id_name":"Alice","destination_number":"1001","start":"` + today + ` 10:00","billsec":60},
			{"context":"public","caller_id_number":"+491700000000","caller_id_name":"Bob","destination_number":"1001","start":"` + today + ` 11:00","billsec":0},
			{"context":"from-internal","caller_id_number":"1001","caller_id_name":"","destination_number":"+493012345678","start":"` + today + ` 12:00","billsec":30}
		]}`))
	}))
	t.Cleanup(srv.Close)

	return srv
}

// TestReportsPanelFoldsTheCallLog pins issue #12's acceptance: the tab
// answers "how much did I actually talk, and to whom" from the CDR page
// alone — totals, the day ladder, the busiest hours and the contact
// ranking all visible without SQL, and the budget bar only when one is
// configured.
func TestReportsPanelFoldsTheCallLog(t *testing.T) {
	server := newTestServerWithPhoneAPI(t, reportsStub(t).URL)
	c := signIn(t, server)

	resp, body := c.do(http.MethodGet, "/partials/reports", nil, "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET /partials/reports = %d", resp.StatusCode)
	}

	page := string(body)
	for _, want := range []string{
		`id="reports-panel"`,
		"This week",
		"Talk time",
		"Most talk time",
		"+441632960961", // the contact that consumed the most time
		"+493012345678",
		"Busiest hours",
		"By day",
	} {
		if !strings.Contains(page, want) {
			t.Errorf("reports panel missing %q", want)
		}
	}

	// 120s + 60s + 0s + 30s = 3:30 of talk, four calls.
	if !strings.Contains(page, "4") {
		t.Error("reports panel does not render the call count")
	}

	if !strings.Contains(page, "3:30") {
		t.Errorf("reports panel does not render the window's talk time (want 3:30): %.600s", page)
	}

	// No budget configured: no bar, no over-budget claim.
	if strings.Contains(page, "wp-budget-meter") {
		t.Error("an unconfigured budget must not render a bar")
	}
}

// TestReportsBudgetBarRendersOnlyWhenConfigured pins the one-number-one-
// bar knob: the budget is config, so the panel renders it exactly when
// the deployment set one — and flags the overspend instead of silently
// clamping a bar at full.
func TestReportsBudgetBarRendersOnlyWhenConfigured(t *testing.T) {
	server := newTestServerWithConfig(t, reportsStub(t).URL, func(cfg *config.Config) {
		cfg.Reports.WeeklyTalkBudget = time.Minute
	})
	c := signIn(t, server)

	_, body := c.do(http.MethodGet, "/partials/reports", nil, "")
	page := string(body)

	if !strings.Contains(page, "wp-budget-meter") {
		t.Fatalf("configured budget did not render a bar: %.600s", page)
	}

	if !strings.Contains(page, "wp-budget-over") {
		t.Error("3:30 of talk against a 1m budget must read as over budget")
	}
}

// TestReportsPanelWithoutPhoneAPISaysSo pins the honest degradation: a
// deployment without the operator phone API has no call records, and the
// panel says that instead of rendering a week of zeros that looks like a
// real, quiet week.
func TestReportsPanelWithoutPhoneAPISaysSo(t *testing.T) {
	server := newTestServer(t)
	c := signIn(t, server)

	_, body := c.do(http.MethodGet, "/partials/reports", nil, "")
	page := string(body)

	if !strings.Contains(page, "phone_api_url") {
		t.Errorf("disabled phone API must name the missing knob: %.400s", page)
	}

	if strings.Contains(page, "wp-stat-value") {
		t.Error("a deployment without records must not render statistic zeros")
	}
}

// TestReportsPanelEmptyWindow pins the empty state: the API answered and
// held nothing, which is a different fact from "unreachable" — the panel
// shows the empty state, never an error banner.
func TestReportsPanelEmptyWindow(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.RequestURI(), "/phone-api/voicemail/") && strings.HasSuffix(r.URL.RequestURI(), "/summary") {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"new":0,"old":0}`))
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"entries":[]}`))
	}))
	t.Cleanup(upstream.Close)

	server := newTestServerWithPhoneAPI(t, upstream.URL)
	c := signIn(t, server)

	_, body := c.do(http.MethodGet, "/partials/reports", nil, "")
	page := string(body)

	if !strings.Contains(page, "No calls to report on yet.") {
		t.Errorf("empty window must render the empty state: %.400s", page)
	}

	if strings.Contains(page, "wp-error") {
		t.Error("an empty window is not an error")
	}
}

// TestReportsTabIsReachableAndNavigable pins the surface itself: the
// route renders the full shell, and the nav offers the tab (a tab nobody
// can see is not a feature).
func TestReportsTabIsReachableAndNavigable(t *testing.T) {
	server := newTestServerWithPhoneAPI(t, reportsStub(t).URL)
	c := signIn(t, server)

	resp, body := c.do(http.MethodGet, "/reports", nil, "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET /reports = %d", resp.StatusCode)
	}

	page := string(body)
	if !strings.Contains(page, `data-tab="reports"`) {
		t.Error("nav does not offer the Reports tab")
	}

	if !strings.Contains(page, `hx-get="/partials/reports"`) {
		t.Error("nav link does not target the Reports partial")
	}
}
