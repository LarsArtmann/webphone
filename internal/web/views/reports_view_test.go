package views

import (
	"strings"
	"testing"
	"time"

	"github.com/larsartmann/webphone/internal/reports"
)

// The totals grid is the Reports tab's headline arithmetic, and
// reportsStat is its ONE home (dedup registry 2026-11-06: the six
// hand-rolled card blocks were the clone group that red-ed the art-dupl
// gate; the extraction is render-diff-proven byte-identical). These pins
// freeze the card bytes and the alert-only-on-missed contract, so a
// future edit cannot silently re-roll the duplication or the tone.
func TestReportsStatRendersTheCardContract(t *testing.T) {
	plain := renderComponent(t, reportsStat("7", "calls", false))
	wantPlain := `<div class="wp-stat"><span class="wp-stat-value">7</span> <span class="wp-stat-label">calls</span></div>`
	if plain != wantPlain {
		t.Errorf("plain card bytes:\n%s\nwant:\n%s", plain, wantPlain)
	}

	alert := renderComponent(t, reportsStat("2", "missed", true))
	wantAlert := `<div class="wp-stat wp-stat-alert"><span class="wp-stat-value">2</span> <span class="wp-stat-label">missed</span></div>`
	if alert != wantAlert {
		t.Errorf("alert card bytes:\n%s\nwant:\n%s", alert, wantAlert)
	}
}

func TestReportsTotalsGridRendersSixCardsAlertingOnlyMissed(t *testing.T) {
	props := ReportsPanelProps{
		Enabled: true,
		Report: reports.Report{
			Totals: reports.Totals{Calls: 7, Inbound: 3, Outbound: 4, Answered: 5, Missed: 2, Talk: 42 * time.Minute},
		},
		Lang: LangEN,
	}
	rendered := renderComponent(t, reportsTotals(props))
	want := `<section class="wp-reports-totals"><div class="wp-stat"><span class="wp-stat-value">7</span> <span class="wp-stat-label">Calls</span></div><div class="wp-stat"><span class="wp-stat-value">42:00</span> <span class="wp-stat-label">Talk time</span></div><div class="wp-stat"><span class="wp-stat-value">3</span> <span class="wp-stat-label">Incoming</span></div><div class="wp-stat"><span class="wp-stat-value">4</span> <span class="wp-stat-label">Outgoing</span></div><div class="wp-stat"><span class="wp-stat-value">5</span> <span class="wp-stat-label">Answered</span></div><div class="wp-stat wp-stat-alert"><span class="wp-stat-value">2</span> <span class="wp-stat-label">Missed</span></div></section>`
	if rendered != want {
		t.Errorf("totals grid bytes:\n%s\nwant:\n%s", rendered, want)
	}
}

// undatedProps is a window whose every record resisted its timestamp:
// real totals and contacts, no day bucket at all.
func undatedProps() ReportsPanelProps {
	return ReportsPanelProps{
		Enabled: true,
		Report: reports.Report{
			Totals:  reports.Totals{Calls: 73, Outbound: 73, Answered: 24, Talk: 116 * time.Minute},
			Hours:   make([]reports.Hour, reports.HoursInDay),
			Undated: 73,
			Contacts: []reports.Contact{
				{Number: "+4915112345678", Calls: 1, Talk: 105 * time.Minute},
			},
			ThisWeek: reports.Week{Start: time.Date(2026, 11, 9, 0, 0, 0, 0, time.UTC)},
		},
		Lang: LangEN,
	}
}

// TestReportsPanelAllUndatedNamesTheCauseRendersOnlyRealFacts pins the
// honest-degradation contract: a window where every timestamp failed
// must NOT headline "This week — 0 calls" (a fake quiet week) nor an
// empty "By day" heading — it names why the time views are absent and
// keeps only the date-independent facts: the totals, the undated count
// and the contact ranking.
func TestReportsPanelAllUndatedNamesTheCauseRendersOnlyRealFacts(t *testing.T) {
	rendered := renderComponent(t, ReportsPanel(undatedProps()))

	for _, want := range []string{
		"Time views unavailable",
		"stay empty instead of showing zeros",
		"Records without a readable timestamp", // the count must survive
		"+4915112345678",                       // contacts are date-independent facts
		`class="wp-stat"`,                      // so are the totals
	} {
		if !strings.Contains(rendered, want) {
			t.Errorf("all-undated panel missing %q:\n%.600s", want, rendered)
		}
	}

	for _, absent := range []string{
		"This week",     // the fake quiet week
		"By day",        // an empty ladder heading
		"Busiest hours", // its ladder guard already skips empty hours
	} {
		if strings.Contains(rendered, absent) {
			t.Errorf("all-undated panel must not render %q:\n%.600s", absent, rendered)
		}
	}
}

// TestReportsPanelMixedUndatedKeepsSectionsAndExplainsTheGap pins the
// partial case: some records parsed, some did not — the temporal
// sections keep the datable calls and the caveat explains the gap right
// under them, before the contact ranking.
func TestReportsPanelMixedUndatedKeepsSectionsAndExplainsTheGap(t *testing.T) {
	props := undatedProps()
	berlin, err := time.LoadLocation("Europe/Berlin")
	if err != nil {
		t.Fatal(err)
	}
	props.Report.Days = []reports.Day{{Date: time.Date(2026, 11, 4, 9, 0, 0, 0, berlin), Calls: 24, Talk: 11 * time.Minute}}
	props.Report.Undated = 49
	props.Report.ThisWeek = reports.Week{Start: time.Date(2026, 11, 2, 0, 0, 0, 0, berlin), Calls: 24, Talk: 11 * time.Minute}

	rendered := renderComponent(t, ReportsPanel(props))

	for _, want := range []string{
		"This week",
		"By day",
		"Records without a readable timestamp:",
		"+4915112345678",
	} {
		if !strings.Contains(rendered, want) {
			t.Errorf("mixed window missing %q:\n%.600s", want, rendered)
		}
	}

	if strings.Contains(rendered, "Time views unavailable") {
		t.Errorf("a partial window must NOT render the all-undated state:\n%.600s", rendered)
	}

	caveat := strings.Index(rendered, "Records without a readable timestamp:")
	contacts := strings.Index(rendered, "Most talk time")
	if caveat < 0 || contacts < 0 || caveat > contacts {
		t.Errorf("the caveat must sit under the temporal sections, before the contact ranking (caveat %d, contacts %d)", caveat, contacts)
	}
}

// TestReportsPanelDatedWindowHasNoCaveat pins the clean case: every
// record parsed, no caveat — the footnote only speaks when the numbers
// are actually short.
func TestReportsPanelDatedWindowHasNoCaveat(t *testing.T) {
	props := undatedProps()
	berlin, err := time.LoadLocation("Europe/Berlin")
	if err != nil {
		t.Fatal(err)
	}
	props.Report.Days = []reports.Day{{Date: time.Date(2026, 11, 4, 9, 0, 0, 0, berlin), Calls: 73, Talk: 116 * time.Minute}}
	props.Report.Undated = 0

	rendered := renderComponent(t, ReportsPanel(props))

	if strings.Contains(rendered, "Records without a readable timestamp") {
		t.Errorf("a fully dated window must not carry the caveat:\n%.600s", rendered)
	}

	if strings.Contains(rendered, "Time views unavailable") {
		t.Errorf("a fully dated window must not render the all-undated state:\n%.600s", rendered)
	}
}

// TestReportsDaysSkipsEmptyLadder pins the days ladder's own guard: a
// heading above nothing reads as data, not absence — the section must
// not render at all when no day bucket exists.
func TestReportsDaysSkipsEmptyLadder(t *testing.T) {
	props := undatedProps()
	props.Report.Days = nil

	if rendered := renderComponent(t, reportsDays(props)); rendered != "" {
		t.Errorf("empty days ladder rendered a heading:\n%s", rendered)
	}

	if rendered := renderComponent(t, reportsWeek(props)); rendered != "" {
		t.Errorf("week above an undatable window rendered:\n%s", rendered)
	}
}
