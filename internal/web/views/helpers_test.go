package views

import (
	"testing"
	"time"

	"github.com/larsartmann/webphone/internal/domain"
)

func TestAvatarForCountrySignum(t *testing.T) {
	cases := map[string]string{
		"+14155550132":  "+1",
		"+441632960961": "+4",
		"+491512345678": "+4",
		"+":             "+",
		"030 1234567":   "0",
		"1555123":       "1",
	}
	for in, want := range cases {
		if got := avatarFor(in); got != want {
			t.Errorf("avatarFor(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestAvatarForNameInitials(t *testing.T) {
	cases := map[string]string{
		"Anna Kellner":    "AK",
		"anna kellner":    "ak",
		"Mara":            "M",
		"van der Berg":    "vd",
		"  Lead  Spaces":  "LS",
		"Telekom Störung": "TS",
	}
	for in, want := range cases {
		if got := avatarFor(in); got != want {
			t.Errorf("avatarFor(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestAvatarForBlank(t *testing.T) {
	for _, in := range []string{"", "   "} {
		if got := avatarFor(in); got != "?" {
			t.Errorf("avatarFor(%q) = %q, want %q", in, got, "?")
		}
	}
}

func TestAvatarHueStableAndBounded(t *testing.T) {
	first := avatarHue("+441632960961")
	for i := 0; i < 50; i++ {
		if got := avatarHue("+441632960961"); got != first {
			t.Fatalf("avatarHue not deterministic: %d then %d", first, got)
		}
	}
	for _, in := range []string{"", "A", "+441632960961", "Anna Kellner", "展"} {
		if got := avatarHue(in); got < 0 || got > 359 {
			t.Fatalf("avatarHue(%q) = %d, out of [0,359]", in, got)
		}
	}
	if avatarHue("Anna") == avatarHue("Bolt") {
		t.Log("collision on distinct inputs is allowed, but suspicious")
	}
}

func TestIsSelfThread(t *testing.T) {
	remote := domain.MustParsePhone("+17287289311")
	cases := []struct {
		name     string
		identity string
		remote   domain.Phone
		want     bool
	}{
		{"match", "+17287289311", remote, true},
		{"match ignores config spacing", "+1 728 728 9311", remote, true},
		{"different number", "+441632960961", remote, false},
		{"no identity configured", "", remote, false},
		{"unparseable identity never warns", "!!!", remote, false},
	}
	for _, tc := range cases {
		if got := isSelfThread(tc.identity, tc.remote); got != tc.want {
			t.Errorf("%s: isSelfThread(%q) = %v, want %v", tc.name, tc.identity, got, tc.want)
		}
	}
}

func TestFormatClockAndStampFollowLanguage(t *testing.T) {
	// A fixed local-time instant (the helpers render t.Local(), so the
	// fixture rides time.Local to stay deterministic on any host TZ).
	at := time.Date(2026, 9, 22, 16, 9, 0, 0, time.Local)
	cases := []struct {
		name  string
		lang  Lang
		clock string
		stamp string
	}{
		{"english keeps the meridiem clock", LangEN, "4:09PM", "Sep 22, 16:09"},
		{"german goes 24h + numeric date", LangDE, "16:09", "22.09. 16:09"},
	}
	for _, tc := range cases {
		if got := formatClock(tc.lang, at); got != tc.clock {
			t.Errorf("%s: formatClock = %q, want %q", tc.name, got, tc.clock)
		}
		if got := formatStamp(tc.lang, at); got != tc.stamp {
			t.Errorf("%s: formatStamp = %q, want %q", tc.name, got, tc.stamp)
		}
	}
}
