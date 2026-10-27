package domain

import (
	"testing"

	"github.com/larsartmann/go-error-family"
	errorfamilytest "github.com/larsartmann/go-error-family/errorfamilytest"
)

func TestParseDisposition(t *testing.T) {
	for _, valid := range []Disposition{
		DispositionAnswered, DispositionMissed, DispositionVoicemail,
		DispositionIVR, DispositionFailed, DispositionUnknown,
	} {
		got, err := ParseDisposition(valid.String())
		if err != nil || got != valid {
			t.Fatalf("ParseDisposition(%q) = %v, %v", valid, got, err)
		}
	}

	_, err := ParseDisposition("answered ")
	if err == nil {
		t.Fatal("foreign strings must refuse")
	}
	errorfamilytest.AssertFamily(t, err, errorfamily.Rejection)
	_, err = ParseDisposition("Answered")
	if err == nil {
		t.Fatal("the vocabulary is lowercase-exact; capitalizations must refuse")
	}
}

// TestDispositionFromEndTable pins the issue #6 machine table: the
// only row allowed to name a specific non-unknown resolution is the
// never-connected one, and NO row — including the established row that
// used to claim "answered" — may reach the answered class.
func TestDispositionFromEndTable(t *testing.T) {
	rows := []struct {
		established bool
		seconds     int
		want        Disposition
	}{
		{false, 0, DispositionMissed},
		{false, -3, DispositionMissed},
		{false, 45, DispositionUnknown},
		{true, 0, DispositionUnknown},
		{true, 1, DispositionUnknown},
		{true, 600, DispositionUnknown},
	}
	for _, row := range rows {
		got := DispositionFromEnd(row.established, row.seconds)
		if got != row.want {
			t.Fatalf("DispositionFromEnd(%v, %d) = %q, want %q", row.established, row.seconds, got, row.want)
		}
		if got.AnsweredClass() {
			t.Fatalf("observed end (%v, %d) reached the answered class — the table may never guess", row.established, row.seconds)
		}
	}
}

// TestLegacyOutcomeDemotion pins the stop-the-lying ruling: the retired
// outcome wire value "answered" was connection-state-derived and
// DEMOTES to unknown, "missed" survives verbatim, and every
// unrecognized or empty value lands on unknown — never the answered
// class (issue #6 acceptance bar).
func TestLegacyOutcomeDemotion(t *testing.T) {
	rows := []struct {
		legacy string
		want   Disposition
	}{
		{"answered", DispositionUnknown},
		{"missed", DispositionMissed},
		{"", DispositionUnknown},
		{"voicemail", DispositionUnknown},
		{"ANSWERED", DispositionUnknown},
		{"garbage", DispositionUnknown},
	}
	for _, row := range rows {
		got := LegacyOutcomeDisposition(row.legacy)
		if got != row.want {
			t.Fatalf("LegacyOutcomeDisposition(%q) = %q, want %q", row.legacy, got, row.want)
		}
	}
}

// TestAnsweredClassIsExclusive pins the class law: exactly one
// vocabulary member is the answered class.
func TestAnsweredClassIsExclusive(t *testing.T) {
	for _, d := range []Disposition{
		DispositionAnswered, DispositionMissed, DispositionVoicemail,
		DispositionIVR, DispositionFailed, DispositionUnknown,
	} {
		if got, want := d.AnsweredClass(), d == DispositionAnswered; got != want {
			t.Fatalf("Disposition(%q).AnsweredClass() = %v, want %v", d, got, want)
		}
	}
}
