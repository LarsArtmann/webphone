package server

import (
	"net/http"
	"time"

	"github.com/a-h/templ"

	"github.com/larsartmann/webphone/internal/reports"
	"github.com/larsartmann/webphone/internal/session"
	"github.com/larsartmann/webphone/internal/web/views"
)

// reportsFetchSize is the CDR window the Reports tab folds. Wider than
// the History tab's page because aggregates over 30 rows answer "this
// afternoon", not "this week" — and narrow enough that one tab click
// stays one phone-API call. The panel's subtitle says "recent calls" for
// exactly this reason: the numbers describe the records the API
// returned, never an invented history this deployment does not hold.
const reportsFetchSize = 500

// reportsPanel folds the extension's call records into the Reports tab
// (issue #12). Pure presentation of a pure projection: the phone API owns
// the records, internal/reports owns the arithmetic, and this handler
// only supplies the window, the deployment's zone and the optional
// budget.
//
// No SSE nudge of its own: an aggregate is read at a glance, not watched,
// and the tab re-folds on every visit.
func (h *handlers) reportsPanel(r *http.Request, sess session.Session) (templ.Component, error) {
	lang := h.lang(r)
	props := views.ReportsPanelProps{Lang: lang}
	if !h.deps.PhoneAPI.Enabled() {
		return views.ReportsPanel(props), nil
	}

	page, err := h.deps.PhoneAPI.History(r.Context(), sess.PBXCredentials(), reportsFetchSize)
	if err != nil {
		props.Enabled = true
		props.Error = h.T(r, "reports.unreachable")

		return views.ReportsPanel(props), nil
	}

	report := reports.Aggregate(reports.Input{
		Entries:  page.Entries,
		Location: time.Local,
		Now:      time.Now(),
		Budget:   h.deps.Config.Reports.WeeklyTalkBudget,
	})

	props.Enabled = true
	props.Report = report
	props.CRM = h.crmDirectory(r.Context(), crmNumbers(report.TopContacts(views.TopContactRows),
		func(contact reports.Contact) string { return contact.Number }))

	return views.ReportsPanel(props), nil
}
