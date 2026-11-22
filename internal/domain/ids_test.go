package domain

import (
	"encoding/json/v2"
	"errors"
	"testing"
)

func TestParseExtension(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		want    string
		wantErr bool
	}{
		{name: "digits", raw: "1001", want: "1001"},
		{name: "strips direction marks and spaces", raw: "\u200e 10 01 \u200f", want: "1001"},
		{name: "letters allowed (some PBXs)", raw: "sales", want: "sales"},
		{name: "empty", raw: "", wantErr: true},
		{name: "only punctuation", raw: "---", wantErr: true},
		{name: "too long", raw: "123456789012345678901234567890123", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseExtension(tt.raw)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("ParseExtension(%q) = %q, want error", tt.raw, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseExtension(%q): %v", tt.raw, err)
			}
			if got.String() != tt.want {
				t.Fatalf("ParseExtension(%q) = %q, want %q", tt.raw, got, tt.want)
			}
		})
	}
}

func TestExtensionTextDoorsAreStrict(t *testing.T) {
	// The config decode door (koanf/mapstructure → UnmarshalText) is the
	// ONE place identities keys and chat owners become Extensions. Its
	// strictness is the fail-closed key rule: a form that would sanitize
	// differently is REJECTED, because a silently-normalized map key
	// misses its own lookup (one line, two spellings).
	var ext Extension
	if err := ext.UnmarshalText([]byte("1001")); err != nil || ext.String() != "1001" {
		t.Fatalf("UnmarshalText(normal) = %q, %v; want 1001, nil", ext, err)
	}
	var sloppy Extension
	if err := sloppy.UnmarshalText([]byte("10 01")); err == nil {
		t.Fatalf("UnmarshalText(non-normalized) = %q, nil; want rejection (sanitize would change it)", sloppy)
	}
	var junk Extension
	if err := junk.UnmarshalText([]byte("---")); err == nil {
		t.Fatal("UnmarshalText(no dialable chars) must reject")
	}

	// MarshalText is the map-key lane: json renders map[Extension]V keys
	// through it, and the bytes must equal MarshalJSON's (the raw value).
	text, err := MustParseExtension("1001").MarshalText()
	if err != nil || string(text) != "1001" {
		t.Fatalf("MarshalText = %q, %v; want 1001, nil", text, err)
	}
	wire, err := json.Marshal(map[Extension]string{MustParseExtension("1001"): "+4930"})
	if err != nil {
		t.Fatalf("marshal typed map: %v", err)
	}
	if string(wire) != `{"1001":"+4930"}` {
		t.Fatalf("typed map keys marshal wrong: %s", wire)
	}
}

func TestParsePhoneSanitizesLikeTheIsland(t *testing.T) {
	// The island dials with raw.replace(/[^\d+*#a-zA-Z]/g, ""); threads
	// must key on the identical normalization or history rows and threads
	// diverge. The served-asset table in internal/server pins the literal
	// regex string, so a one-sided drift fails the build.
	got, err := ParsePhone(" +44 (1632) 960-961 ")
	if err != nil {
		t.Fatal(err)
	}
	if got.String() != "+441632960961" {
		t.Fatalf("got %q, want +441632960961", got)
	}
	if got, _ := ParsePhone("*1"); got.String() != "*1" {
		t.Fatalf("star key: got %q", got)
	}
	// The heavy-asterisk glyph is NOT the dialable star; it must strip out.
	if _, err := ParsePhone("✱"); err == nil {
		t.Fatal("unicode heavy asterisk must not survive sanitization")
	}
}

func TestChannelOf(t *testing.T) {
	if ChannelOf(0) != ChannelSMS {
		t.Fatal("text only must be SMS")
	}
	if ChannelOf(1) != ChannelMMS {
		t.Fatal("attachment makes MMS")
	}
	if ChannelOf(2) != ChannelMMS {
		t.Fatal("text plus attachments makes MMS")
	}
}

func TestMustIDAcceptsBrandedAndRaw(t *testing.T) {
	id := GenerateThreadID()
	branded := id.String()
	if branded == "" {
		t.Fatal("branded string form empty")
	}
	if again := MustThreadID(branded); again != id {
		t.Fatalf("branded round trip: %q != %q", again, id)
	}
	if _, rest, found := cut(branded, ':'); !found {
		t.Fatalf("expected branded form, got %q", branded)
	} else if MustThreadID(rest) != id {
		t.Fatalf("raw round trip failed for %q", rest)
	}
}

func TestParseThreadIDRoundTripsAndRejects(t *testing.T) {
	valid := GenerateThreadID()
	got, err := ParseThreadID(valid.String())
	if err != nil {
		t.Fatalf("ParseThreadID(%q): %v", valid.String(), err)
	}
	if got != valid {
		t.Fatalf("branded round trip: %q != %q", got, valid)
	}
	if raw := valid.String(); len(raw) > 21 {
		if again, err := ParseThreadID(raw[len(raw)-21:]); err != nil || again != valid {
			t.Fatalf("raw round trip: %q, %v", again, err)
		}
	}
	for _, bad := range []string{"", "short", "Thread:tooshort", "has spaces 1234567890123"} {
		if _, err := ParseThreadID(bad); err == nil {
			t.Fatalf("ParseThreadID(%q) accepted a malformed id", bad)
		}
	}
}

func cut(s string, sep byte) (string, string, bool) {
	for i := 0; i < len(s); i++ {
		if s[i] == sep {
			return s[:i], s[i+1:], true
		}
	}
	return s, "", false
}

// TestMustUnwrapsOrPanics pins the Must engine itself: success passes
// the value through untouched, an error panics with that exact error —
// the Must forms are for literals and database rows, so a failure is a
// broken literal or a corrupt database, never bad user input.
func TestMustUnwrapsOrPanics(t *testing.T) {
	want := GenerateThreadID()
	if got := must(want, nil); got != want {
		t.Fatalf("must success: %v", got)
	}
	boom := errors.New("corrupt id")
	func() {
		defer func() {
			r := recover()
			err, ok := r.(error)
			if !ok || !errors.Is(err, boom) {
				t.Fatalf("must panic carried %v, want the original error", r)
			}
		}()
		_ = must(GenerateThreadID(), boom)
		t.Fatal("must must panic on error")
	}()
}
