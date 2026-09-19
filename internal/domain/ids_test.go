package domain

import "testing"

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

func TestParsePhoneSanitizesLikeTheIsland(t *testing.T) {
	// The island dials with raw.replace(/[^\d+*#]/g, ""); threads must key
	// on the identical normalization or history rows and threads diverge.
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

func cut(s string, sep byte) (string, string, bool) {
	for i := 0; i < len(s); i++ {
		if s[i] == sep {
			return s[:i], s[i+1:], true
		}
	}
	return s, "", false
}
