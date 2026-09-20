// Parser edge table (plan T18, 2026-09-20): ParseExtension / ParsePhone
// and the branded id parsers as one table-driven spec — prefix, length,
// charset, empty — so the sanitization contract (island parity, DECIDED
// 2026-09-20) cannot drift silently.
package domain

import (
	"strings"
	"testing"
)

func TestDialableParserEdgeTable(t *testing.T) {
	cases := []struct {
		name    string
		raw     string
		want    string
		wantErr bool
	}{
		{"empty", "", "", true},
		{"plain extension", "1001", "1001", false},
		{"leading plus", "+441632960961", "+441632960961", false},
		{"letters kept (island parity)", "support", "support", false},
		{"mixed alphanumeric", "1a2b*3c#", "1a2b*3c#", false},
		{"spaces stripped", " 100 1 ", "1001", false},
		{"dashes stripped", "100-1", "1001", false},
		{"parens stripped", "(0) 1632", "01632", false},
		{"unicode direction marks stripped (pasted RTL number)",
			"+44​1632\u200f960961", "+441632960961", false},
		{"emoji stripped", "1001📱", "1001", false},
		{"punctuation stripped to empty", "…", "", true},
		{"32 chars is the boundary", strings.Repeat("a", 32), strings.Repeat("a", 32), false},
		{"33 chars rejected", strings.Repeat("a", 33), "", true},
	}
	for _, tc := range cases {
		t.Run("extension/"+tc.name, func(t *testing.T) {
			got, err := ParseExtension(tc.raw)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("ParseExtension(%q) accepted, want error", tc.raw)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseExtension(%q): %v", tc.raw, err)
			}
			if got.String() != tc.want {
				t.Errorf("ParseExtension(%q) = %q, want %q", tc.raw, got.String(), tc.want)
			}
		})
		t.Run("phone/"+tc.name, func(t *testing.T) {
			got, err := ParsePhone(tc.raw)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("ParsePhone(%q) accepted, want error", tc.raw)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParsePhone(%q): %v", tc.raw, err)
			}
			if got.String() != tc.want {
				t.Errorf("ParsePhone(%q) = %q, want %q", tc.raw, got.String(), tc.want)
			}
		})
	}
}

func TestMustParsersPanicOnGarbage(t *testing.T) {
	for name, fn := range map[string]func(){
		"MustParseExtension": func() { MustParseExtension("…") },
		"MustParsePhone":     func() { MustParsePhone("🎉") },
	} {
		t.Run(name, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Fatal("Must parser accepted garbage without panicking")
				}
			}()
			fn()
		})
	}
}

func TestBrandedIDParsersRejectJunk(t *testing.T) {
	cases := []struct {
		name string
		bad  string
	}{
		{"empty", ""},
		{"whitespace", "   "},
		{"unknown brand prefix", "Fax:not-a-real-one"},
		{"wrong brand for type", "Contact:abc"},
		{"garbage body", "Thread:\x00\x01"},
	}
	parse := map[string]func(string) error{
		"ThreadID":    func(s string) error { _, err := ParseThreadID(s); return err },
		"MessageID":   func(s string) error { _, err := ParseMessageID(s); return err },
		"AttachmentID": func(s string) error { _, err := ParseAttachmentID(s); return err },
		"FaxID":       func(s string) error { _, err := ParseFaxID(s); return err },
		"ContactID":   func(s string) error { _, err := ParseContactID(s); return err },
	}
	for kind, parseFn := range parse {
		for _, tc := range cases {
			t.Run(kind+"/"+tc.name, func(t *testing.T) {
				if err := parseFn(tc.bad); err == nil {
					t.Fatalf("%s.Parse(%q) accepted", kind, tc.bad)
				}
			})
		}
	}
}

func TestBrandedIDRoundTripAcceptsBothForms(t *testing.T) {
	thread := GenerateThreadID()
	// Raw form (database rows) and branded form (rendered strings) both
	// parse back to the same id — the go-branded-id String() renders
	// "Brand:value".
	for _, form := range []string{
		thread.String(),                      // branded "Thread:xxx"
		strings.SplitN(thread.String(), ":", 2)[1], // raw nanoid form (database rows)
	} {
		got, err := ParseThreadID(form)
		if err != nil {
			t.Fatalf("ParseThreadID(%q): %v", form, err)
		}
		if got != thread {
			t.Errorf("round trip drifted: %q", form)
		}
	}
}
