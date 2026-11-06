package views

import (
	"os"
	"testing"
	"time"

	"github.com/larsartmann/webphone/internal/reports"
)

// TestReportsPanelRenderSnapshot is TEMPORARY scaffolding for the
// reportsStat extraction's old-vs-new render diff (dedup protocol:
// templ refactors proven with a render diff). Writes the rendered
// ReportsPanel bytes to $REPORTS_SNAPSHOT when set; deleted after the
// extraction lands.
func TestReportsPanelRenderSnapshot(t *testing.T) {
	path := os.Getenv("REPORTS_SNAPSHOT")
	if path == "" {
		t.Skip("REPORTS_SNAPSHOT not set")
	}
	hours := make([]reports.Hour, reports.HoursInDay)
	hours[9] = reports.Hour{Hour: 9, Calls: 3, Talk: 10 * time.Minute}
	hours[14] = reports.Hour{Hour: 14, Calls: 4, Talk: 32 * time.Minute}
	props := ReportsPanelProps{
		Enabled: true,
		Report: reports.Report{
			Totals:   reports.Totals{Calls: 7, Inbound: 3, Outbound: 4, Answered: 5, Missed: 2, Talk: 42 * time.Minute},
			Days:     []reports.Day{{Date: time.Date(2026, 11, 4, 0, 0, 0, 0, time.UTC), Calls: 4, Talk: 20 * time.Minute}, {Date: time.Date(2026, 11, 5, 0, 0, 0, 0, time.UTC), Calls: 3, Talk: 22 * time.Minute}},
			Hours:    hours,
			Contacts: []reports.Contact{{Number: "+4915112345678", Calls: 3, Talk: 15 * time.Minute}, {Number: "+17287289311", Calls: 4, Talk: 27 * time.Minute}},
			ThisWeek: reports.Week{Start: time.Date(2026, 11, 2, 0, 0, 0, 0, time.UTC), Calls: 6, Talk: 38 * time.Minute, Budget: time.Hour},
			Undated:  1,
		},
		CRM:  CRMContacts{},
		Lang: LangEN,
	}
	rendered := renderComponent(t, ReportsPanel(props))
	if err := os.WriteFile(path, []byte(rendered), 0o600); err != nil {
		t.Fatalf("write snapshot: %v", err)
	}
}
