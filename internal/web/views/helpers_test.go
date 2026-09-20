package views

import "testing"

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
