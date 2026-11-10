// Package reports folds call detail records into the feedback a call log
// cannot give on its own: how much this week actually cost, which
// direction and outcome the calls fell into, when the phone rings, and
// with whom the time went.
//
// It is a PURE projection. No storage, no clock, no I/O: the caller
// supplies the records, the zone the deployment renders in, "now" for
// the current-week window, and the optional weekly budget. The same
// inputs always produce the same Report, which is what makes the tab's
// numbers reproducible (rebuild the projection from the same CDR page
// and every figure repeats exactly) and lets a second consumer — a
// future export, an alert — fold the same records without a read model
// of its own.
package reports

import (
	"cmp"
	"slices"
	"time"

	"github.com/larsartmann/webphone/internal/pbx"
)

// startLayout is the phone API's CDR start format ("2026-11-04 18:42:13").
// FreeSWITCH writes start_stamp as %Y-%m-%d %T (switch_channel.c) and the
// operator API passes it through verbatim, so the seconds are load-bearing:
// parsing minute-only left every live record Undated (2026-11-10 triage).
// It matches the History tab's day bucketing (Start[:10]): both read the
// same wire field, so a format change has to move both.
const startLayout = "2006-01-02 15:04:05"

// HoursInDay is the size of the Report.Hours ladder.
const HoursInDay = 24

// Report is the whole projection of one CDR set.
type Report struct {
	Totals Totals
	// Days are the days that hold calls, oldest first. An empty slice
	// means the window held no call at all.
	Days []Day
	// Hours always holds HoursInDay buckets (index = hour of day in the
	// report's zone), so a caller can render a stable ladder.
	Hours []Hour
	// Contacts are every distinct remote party, ordered by talk time.
	Contacts []Contact
	// ThisWeek is the window [Monday 00:00, next Monday 00:00) in the
	// report's zone.
	ThisWeek Week
	// Undated counts records whose start could not be parsed. They stay
	// in Totals and Contacts (the call happened) but cannot land in a
	// day, hour or week bucket; the view says so instead of pretending
	// the totals add up.
	Undated int
	// NoAnswerByCause counts the hangup causes behind the NoAnswer
	// bucket (outbound calls that never connected), in FreeSWITCH's own
	// vocabulary — translating it would invent a second truth about the
	// same call. An empty-cause row counts as UNKNOWN. Nil when the
	// window held no never-connected outbound call.
	NoAnswerByCause map[string]int
}

// Totals is the window's headline arithmetic.
//
// Direction (Inbound/Outbound) and outcome (Answered/Missed/NoAnswer)
// are different questions, and every call lands in exactly one outcome
// bucket: Answered carried talk time, Missed arrived inbound and carried
// none, NoAnswer is an outbound that never connected. That is the
// History tab's own vocabulary, kept identical so the two tabs never
// disagree about the same calls.
type Totals struct {
	Calls    int
	Inbound  int
	Outbound int
	Answered int
	Missed   int
	NoAnswer int
	Talk     time.Duration
}

// Day is one calendar day of calls.
type Day struct {
	Date     time.Time
	Calls    int
	Inbound  int
	Outbound int
	Answered int
	Missed   int
	Talk     time.Duration
}

// Hour is one hour-of-day bucket across the whole window: the shape of
// the working day.
type Hour struct {
	Hour  int
	Calls int
	Talk  time.Duration
}

// Contact is one remote party's share of the window.
type Contact struct {
	Number string
	Calls  int
	Talk   time.Duration
}

// Week is the current week against the optional talk-time budget.
type Week struct {
	// Start is Monday 00:00 of the week "now" falls in, in the report's
	// zone.
	Start  time.Time
	Calls  int
	Talk   time.Duration
	Budget time.Duration
}

// Fraction is the budget bar's fill, clamped to 0..1. Zero budget (no
// budget configured) reports 0 so a caller cannot render a bar nobody
// asked for.
func (w Week) Fraction() float64 {
	if w.Budget <= 0 {
		return 0
	}

	ratio := float64(w.Talk) / float64(w.Budget)
	if ratio > 1 {
		return 1
	}

	if ratio < 0 {
		return 0
	}

	return ratio
}

// OverBudget reports whether the week has spent more talk time than the
// configured budget. Always false without a budget.
func (w Week) OverBudget() bool { return w.Budget > 0 && w.Talk > w.Budget }

// Input is everything the projection needs.
type Input struct {
	// Entries are the CDR rows, in any order.
	Entries []pbx.CDR
	// Location is the zone the deployment renders in. A nil value means
	// UTC: the projection never reaches for the process zone, so a
	// test's numbers do not depend on where it runs.
	Location *time.Location
	// Now anchors the current-week window. The zero value means the
	// window is empty (no week math) — a caller without a clock renders
	// totals without the budget bar rather than a guessed week.
	Now time.Time
	// Budget is the optional weekly talk-time budget. Zero means "no
	// budget configured": the bar simply does not render.
	Budget time.Duration
}

// Aggregate folds the CDR rows into a Report.
func Aggregate(in Input) Report {
	loc := in.Location
	if loc == nil {
		loc = time.UTC
	}

	report := Report{Hours: make([]Hour, HoursInDay)}
	for hour := range report.Hours {
		report.Hours[hour].Hour = hour
	}

	byDay := make(map[string]*Day, 8)
	byContact := make(map[string]*Contact, 16)
	weekStart := weekStartOf(in.Now, loc)

	for _, cdr := range in.Entries {
		report.Totals.Calls++

		inbound := cdr.Context == "public"
		if inbound {
			report.Totals.Inbound++
		} else {
			report.Totals.Outbound++
		}

		talk := secondsToDuration(cdr.Billsec)
		report.Totals.Talk += talk

		answered := cdr.Billsec > 0
		switch {
		case answered:
			report.Totals.Answered++
		case inbound:
			report.Totals.Missed++
		default:
			report.Totals.NoAnswer++
			cause := cdr.HangupCause
			if cause == "" {
				cause = "UNKNOWN"
			}
			if report.NoAnswerByCause == nil {
				report.NoAnswerByCause = make(map[string]int, 4)
			}
			report.NoAnswerByCause[cause]++
		}

		remote := remoteOf(cdr)
		if remote != "" {
			contact := byContact[remote]
			if contact == nil {
				contact = &Contact{Number: remote}
				byContact[remote] = contact
			}

			contact.Calls++
			contact.Talk += talk
		}

		started, ok := parseStart(cdr.Start, loc)
		if !ok {
			report.Undated++

			continue
		}

		key := started.Format("2006-01-02")
		day := byDay[key]
		if day == nil {
			day = &Day{Date: started}
			byDay[key] = day
		}

		day.Calls++
		day.Talk += talk

		if inbound {
			day.Inbound++
		} else {
			day.Outbound++
		}

		switch {
		case answered:
			day.Answered++
		case inbound:
			day.Missed++
		}

		hour := &report.Hours[started.Hour()]
		hour.Calls++
		hour.Talk += talk

		if !weekStart.IsZero() && !started.Before(weekStart) && started.Before(weekStart.AddDate(0, 0, 7)) {
			report.ThisWeek.Calls++
			report.ThisWeek.Talk += talk
		}
	}

	report.ThisWeek.Start = weekStart
	report.ThisWeek.Budget = in.Budget

	for _, day := range byDay {
		report.Days = append(report.Days, *day)
	}

	slices.SortFunc(report.Days, func(a, b Day) int { return a.Date.Compare(b.Date) })

	report.Contacts = make([]Contact, 0, len(byContact))
	for _, contact := range byContact {
		report.Contacts = append(report.Contacts, *contact)
	}

	// Talk time first (where the time went), then volume, then the
	// number — the last key only to keep the order total, because a map
	// iteration order would make two identical runs differ.
	slices.SortFunc(report.Contacts, func(a, b Contact) int {
		return cmp.Or(
			cmp.Compare(b.Talk, a.Talk),
			cmp.Compare(b.Calls, a.Calls),
			cmp.Compare(a.Number, b.Number),
		)
	})

	return report
}

// TopContacts returns at most limit contacts, already ranked. The cap
// lives with the caller because it is a display budget, not a fact about
// the window — and naming the very same rows the ladder will render is
// what keeps CRM lookups down to what a reader actually sees.
func (r Report) TopContacts(limit int) []Contact {
	if limit <= 0 || len(r.Contacts) == 0 {
		return nil
	}

	if limit > len(r.Contacts) {
		limit = len(r.Contacts)
	}

	return r.Contacts[:limit]
}

// BusiestHours returns the hours with at least one call, busiest first.
// Callers cap the list to what fits; ordering is total (calls, then talk,
// then hour) so a test can pin it.
func (r Report) BusiestHours() []Hour {
	busy := make([]Hour, 0, HoursInDay)
	for _, hour := range r.Hours {
		if hour.Calls > 0 {
			busy = append(busy, hour)
		}
	}

	slices.SortFunc(busy, func(a, b Hour) int {
		return cmp.Or(
			cmp.Compare(b.Calls, a.Calls),
			cmp.Compare(b.Talk, a.Talk),
			cmp.Compare(a.Hour, b.Hour),
		)
	})

	return busy
}

// MaxDayTalk is the largest daily talk time, the denominator a daily bar
// ladder needs. Zero when no day holds talk time.
func (r Report) MaxDayTalk() time.Duration {
	var max time.Duration

	for _, day := range r.Days {
		if day.Talk > max {
			max = day.Talk
		}
	}

	return max
}

// remoteOf is the OTHER party: the caller for an inbound record, the
// dialled destination for an outbound one. The History tab asks the same
// question (views.cdrTarget), so the two tabs name a contact identically.
func remoteOf(cdr pbx.CDR) string {
	if cdr.Context == "public" {
		return cdr.CallerIDNumber
	}

	return cdr.DestinationNumber
}

func parseStart(raw string, loc *time.Location) (time.Time, bool) {
	if raw == "" {
		return time.Time{}, false
	}

	parsed, err := time.ParseInLocation(startLayout, raw, loc)
	if err != nil {
		return time.Time{}, false
	}

	return parsed, true
}

// weekStartOf is Monday 00:00 of the week the instant falls in. Zero
// "now" (no clock) yields the zero time, which the week accumulators read
// as "no window".
func weekStartOf(now time.Time, loc *time.Location) time.Time {
	if now.IsZero() {
		return time.Time{}
	}

	local := now.In(loc)
	midnight := time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, loc)

	// Go's Weekday has Sunday = 0; the ISO week starts Monday.
	offset := (int(midnight.Weekday()) + 6) % 7

	return midnight.AddDate(0, 0, -offset)
}

func secondsToDuration(seconds int) time.Duration {
	if seconds <= 0 {
		return 0
	}

	return time.Duration(seconds) * time.Second
}
