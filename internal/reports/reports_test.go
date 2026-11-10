package reports_test

import (
	"testing"
	"time"

	"github.com/larsartmann/webphone/internal/pbx"
	"github.com/larsartmann/webphone/internal/reports"
)

// berlin is fixed so the projection's numbers never depend on the zone
// the test machine runs in.
var berlin = mustLoadBerlin()

func mustLoadBerlin() *time.Location {
	loc, err := time.LoadLocation("Europe/Berlin")
	if err != nil {
		panic(err)
	}

	return loc
}

func cdr(context, caller, dest, start string, billsec int) pbx.CDR {
	return pbx.CDR{Context: context, CallerIDNumber: caller, DestinationNumber: dest, Start: start, Billsec: billsec}
}

// TestAggregateFoldsDirectionOutcomeAndTalk pins the headline arithmetic
// and the independence of its axes: direction (in/out) and outcome
// (answered/missed) are different questions, and an outbound call that
// never connected belongs to neither outcome bucket.
func TestAggregateFoldsDirectionOutcomeAndTalk(t *testing.T) {
	report := reports.Aggregate(reports.Input{
		Entries: []pbx.CDR{
			cdr("public", "+4930111", "1001", "2026-11-04 09:05:00", 90),
			cdr("public", "+4930222", "1001", "2026-11-04 10:15:00", 0),
			cdr("internal", "1001", "+4930333", "2026-11-04 11:30:00", 30),
			cdr("internal", "1001", "+4930444", "2026-11-04 12:00:00", 0),
		},
		Location: berlin,
		Now:      time.Date(2026, 11, 4, 18, 0, 0, 0, berlin),
	})

	want := reports.Totals{Calls: 4, Inbound: 2, Outbound: 2, Answered: 2, Missed: 1, NoAnswer: 1, Talk: 2 * time.Minute}
	if report.Totals != want {
		t.Fatalf("totals = %+v, want %+v", report.Totals, want)
	}

	if report.Undated != 0 {
		t.Fatalf("undated = %d, want 0", report.Undated)
	}

	if len(report.Days) != 1 {
		t.Fatalf("days = %d, want 1", len(report.Days))
	}

	day := report.Days[0]
	if day.Calls != 4 || day.Inbound != 2 || day.Outbound != 2 || day.Answered != 2 || day.Missed != 1 {
		t.Fatalf("day buckets = %+v", day)
	}

	if day.Date.Format("2006-01-02") != "2026-11-04" {
		t.Fatalf("day date = %s", day.Date)
	}
}

// TestNoAnswerCountsOutboundThatNeverConnected pins the outcome
// partition: Answered carried talk, Missed arrived inbound and carried
// none, NoAnswer is the outbound that never connected — every call in
// exactly one bucket, so the three outcomes always sum to Calls.
func TestNoAnswerCountsOutboundThatNeverConnected(t *testing.T) {
	report := reports.Aggregate(reports.Input{
		Entries: []pbx.CDR{
			cdr("internal", "1001", "+4930111", "2026-11-04 09:00:00", 0), // outbound, never connected
			cdr("internal", "1001", "+4930122", "2026-11-04 09:05:00", 0), // outbound, never connected
			cdr("internal", "1001", "+4930133", "2026-11-04 09:10:00", 45),
			cdr("public", "+4930144", "1001", "2026-11-04 09:15:00", 0), // inbound, missed
			cdr("public", "+4930155", "1001", "2026-11-04 09:20:00", 30),
		},
		Location: berlin,
	})

	want := reports.Totals{Calls: 5, Inbound: 2, Outbound: 3, Answered: 2, Missed: 1, NoAnswer: 2, Talk: 75 * time.Second}
	if report.Totals != want {
		t.Fatalf("totals = %+v, want %+v", report.Totals, want)
	}

	if report.Totals.Answered+report.Totals.Missed+report.Totals.NoAnswer != report.Totals.Calls {
		t.Fatalf("outcomes %d + %d + %d do not partition %d calls",
			report.Totals.Answered, report.Totals.Missed, report.Totals.NoAnswer, report.Totals.Calls)
	}
}

// TestContactsRankByTalkThenVolumeThenNumber pins the total order: talk
// time first (where the time went), volume next, and the number last so
// two runs over the same records can never disagree — map iteration
// order must not leak into the UI.
func TestContactsRankByTalkThenVolumeThenNumber(t *testing.T) {
	report := reports.Aggregate(reports.Input{
		Entries: []pbx.CDR{
			cdr("public", "+49b", "1001", "2026-11-04 09:00:00", 60),
			cdr("public", "+49a", "1001", "2026-11-04 09:10:00", 60),
			cdr("internal", "1001", "+49c", "2026-11-04 09:20:00", 300),
			cdr("public", "+49a", "1001", "2026-11-04 09:30:00", 0),
		},
		Location: berlin,
	})

	got := make([]string, 0, len(report.Contacts))
	for _, contact := range report.Contacts {
		got = append(got, contact.Number)
	}

	want := []string{"+49c", "+49a", "+49b"}
	if len(got) != len(want) {
		t.Fatalf("contacts = %v, want %v", got, want)
	}

	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("contacts = %v, want %v", got, want)
		}
	}

	if report.Contacts[1].Calls != 2 {
		t.Fatalf("+49a calls = %d, want 2", report.Contacts[1].Calls)
	}
}

// TestBusiestHoursOrdersAndSkipsEmptyHours pins the hour ladder: every
// bucket exists (24), but the busiest list carries only hours that rang,
// busiest first.
func TestBusiestHoursOrdersAndSkipsEmptyHours(t *testing.T) {
	report := reports.Aggregate(reports.Input{
		Entries: []pbx.CDR{
			cdr("public", "+4930", "1001", "2026-11-04 17:05:00", 30),
			cdr("public", "+4931", "1001", "2026-11-04 17:45:00", 30),
			cdr("internal", "1001", "+4932", "2026-11-04 08:10:00", 10),
		},
		Location: berlin,
	})

	if len(report.Hours) != reports.HoursInDay {
		t.Fatalf("hours ladder = %d, want %d", len(report.Hours), reports.HoursInDay)
	}

	busy := report.BusiestHours()
	if len(busy) != 2 {
		t.Fatalf("busy hours = %d, want 2 (%+v)", len(busy), busy)
	}

	if busy[0].Hour != 17 || busy[0].Calls != 2 {
		t.Fatalf("busiest = %+v, want hour 17 with 2 calls", busy[0])
	}

	if busy[1].Hour != 8 || busy[1].Calls != 1 {
		t.Fatalf("runner-up = %+v, want hour 8 with 1 call", busy[1])
	}
}

// TestThisWeekCountsOnlyTheMondayWindow pins the week window: earlier
// calls stay in the totals and the day list but never inflate the
// current week, and a Sunday-night call belongs to the week that just
// ended (ISO weeks start Monday).
func TestThisWeekCountsOnlyTheMondayWindow(t *testing.T) {
	// 2026-11-04 is a Wednesday; that week started Monday 2026-11-02.
	report := reports.Aggregate(reports.Input{
		Entries: []pbx.CDR{
			cdr("public", "+4930", "1001", "2026-10-28 09:00:00", 600), // previous week
			cdr("public", "+4931", "1001", "2026-11-01 23:30:00", 120), // Sunday: previous week
			cdr("public", "+4932", "1001", "2026-11-02 00:05:00", 60),  // Monday: this week
			cdr("public", "+4933", "1001", "2026-11-04 09:00:00", 60),  // Wednesday: this week
		},
		Location: berlin,
		Now:      time.Date(2026, 11, 4, 12, 0, 0, 0, berlin),
	})

	if report.Totals.Calls != 4 {
		t.Fatalf("totals calls = %d, want 4 (the window does not gate the totals)", report.Totals.Calls)
	}

	if report.ThisWeek.Calls != 2 || report.ThisWeek.Talk != 2*time.Minute {
		t.Fatalf("this week = %+v, want 2 calls / 2m", report.ThisWeek)
	}

	if got := report.ThisWeek.Start.Format("2006-01-02"); got != "2026-11-02" {
		t.Fatalf("week start = %s, want 2026-11-02", got)
	}

	if report.ThisWeek.OverBudget() {
		t.Fatal("no budget configured must never read as over budget")
	}

	if report.ThisWeek.Fraction() != 0 {
		t.Fatalf("fraction without a budget = %v, want 0", report.ThisWeek.Fraction())
	}
}

// TestBudgetBarClampsAndFlagsOverspend pins the one-number-one-bar
// contract: the bar fills toward 1 and never draws past it, while
// OverBudget stays true so the copy can say what happened.
func TestBudgetBarClampsAndFlagsOverspend(t *testing.T) {
	report := reports.Aggregate(reports.Input{
		Entries:  []pbx.CDR{cdr("public", "+4930", "1001", "2026-11-04 09:00:00", 300)},
		Location: berlin,
		Now:      time.Date(2026, 11, 4, 12, 0, 0, 0, berlin),
		Budget:   2 * time.Minute,
	})

	if !report.ThisWeek.OverBudget() {
		t.Fatal("5m of talk against a 2m budget must read as over budget")
	}

	if report.ThisWeek.Fraction() != 1 {
		t.Fatalf("fraction = %v, want a clamped 1", report.ThisWeek.Fraction())
	}

	half := reports.Aggregate(reports.Input{
		Entries:  []pbx.CDR{cdr("public", "+4930", "1001", "2026-11-04 09:00:00", 60)},
		Location: berlin,
		Now:      time.Date(2026, 11, 4, 12, 0, 0, 0, berlin),
		Budget:   2 * time.Minute,
	})

	if got := half.ThisWeek.Fraction(); got < 0.49 || got > 0.51 {
		t.Fatalf("half-spent fraction = %v, want ~0.5", got)
	}
}

// TestUnparsableStartsAreCountedNotGuessed pins the honesty rule: a
// record whose start cannot be read still happened, so it stays in the
// totals and the contact ranking — and is counted in Undated so the view
// can say the time buckets are short, instead of silently dropping it.
func TestUnparsableStartsAreCountedNotGuessed(t *testing.T) {
	report := reports.Aggregate(reports.Input{
		Entries: []pbx.CDR{
			cdr("public", "+4930", "1001", "", 60),
			cdr("public", "+4931", "1001", "not-a-timestamp", 30),
			cdr("public", "+4932", "1001", "2026-11-04 09:00:00", 60),
			cdr("public", "+4933", "1001", "2026-11-04 09:05", 60), // minute-only: the pre-2026-11-10 layout
		},
		Location: berlin,
	})

	if report.Totals.Calls != 4 || report.Totals.Talk != 210*time.Second {
		t.Fatalf("totals = %+v, want 4 calls / 3m30s", report.Totals)
	}

	if report.Undated != 3 {
		t.Fatalf("undated = %d, want 3", report.Undated)
	}

	if len(report.Days) != 1 || len(report.Contacts) != 4 {
		t.Fatalf("days = %d contacts = %d, want 1 and 4", len(report.Days), len(report.Contacts))
	}
}

// TestZeroClockRendersWithoutAWeek pins the no-clock contract: a caller
// that never supplies "now" gets totals, days and contacts, and no week
// claim at all (rather than a guessed window).
func TestZeroClockRendersWithoutAWeek(t *testing.T) {
	report := reports.Aggregate(reports.Input{
		Entries:  []pbx.CDR{cdr("public", "+4930", "1001", "2026-11-04 09:00:00", 60)},
		Location: berlin,
	})

	if report.ThisWeek.Calls != 0 || report.ThisWeek.Talk != 0 {
		t.Fatalf("this week = %+v, want an empty window", report.ThisWeek)
	}

	if !report.ThisWeek.Start.IsZero() {
		t.Fatalf("week start = %v, want zero", report.ThisWeek.Start)
	}

	if report.Totals.Calls != 1 {
		t.Fatalf("totals calls = %d, want 1", report.Totals.Calls)
	}
}

// TestEmptyWindowReportsNothing pins the zero case the view keys its
// empty state on: no rows means no days, no contacts, and a 24-bucket
// ladder that is entirely empty.
func TestEmptyWindowReportsNothing(t *testing.T) {
	report := reports.Aggregate(reports.Input{})

	if len(report.Days) != 0 || len(report.Contacts) != 0 {
		t.Fatalf("empty window produced %d days / %d contacts", len(report.Days), len(report.Contacts))
	}

	if len(report.Hours) != reports.HoursInDay {
		t.Fatalf("hours ladder = %d, want %d", len(report.Hours), reports.HoursInDay)
	}

	for _, hour := range report.BusiestHours() {
		t.Fatalf("empty window has a busy hour: %+v", hour)
	}

	if report.Totals != (reports.Totals{}) {
		t.Fatalf("totals = %+v, want zero", report.Totals)
	}
}

// TestDaysAreAscending pins the day order: the newest call is not
// necessarily last in the CDR page, and the ladder must read oldest
// first regardless.
func TestDaysAreAscending(t *testing.T) {
	report := reports.Aggregate(reports.Input{
		Entries: []pbx.CDR{
			cdr("public", "+4930", "1001", "2026-11-04 09:00:00", 60),
			cdr("public", "+4931", "1001", "2026-11-01 09:00:00", 60),
			cdr("public", "+4932", "1001", "2026-11-03 09:00:00", 60),
		},
		Location: berlin,
	})

	want := []string{"2026-11-01", "2026-11-03", "2026-11-04"}
	for i, day := range report.Days {
		if got := day.Date.Format("2006-01-02"); got != want[i] {
			t.Fatalf("day %d = %s, want %s", i, got, want[i])
		}
	}

	if report.MaxDayTalk() != time.Minute {
		t.Fatalf("max day talk = %v, want 1m", report.MaxDayTalk())
	}
}
