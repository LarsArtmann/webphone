package views

import (
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
