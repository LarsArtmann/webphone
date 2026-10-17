package domain

import (
	"strings"
	"testing"
)

// TestEnumParsersRoundTrip pins the Must* vocabulary closes: every
// stored enum value parses (and Musts) back to itself, and every
// foreign string is rejected with its stable seam code — the corrupt-
// row panic path the store scan sites ride on.
func TestEnumParsersRoundTrip(t *testing.T) {
	t.Run("direction", func(t *testing.T) {
		for _, raw := range []string{"in", "out"} {
			if got := MustDirection(raw); string(got) != raw {
				t.Errorf("MustDirection(%q) = %q", raw, got)
			}
		}
		if _, err := ParseDirection("sideways"); err == nil || !strings.Contains(err.Error(), "domain.direction") {
			t.Errorf("ParseDirection(sideways): %v (want the domain.direction code)", err)
		}
	})
	t.Run("channel", func(t *testing.T) {
		for _, raw := range []string{"sms", "mms"} {
			if got := MustChannel(raw); string(got) != raw {
				t.Errorf("MustChannel(%q) = %q", raw, got)
			}
		}
		if _, err := ParseChannel("email"); err == nil || !strings.Contains(err.Error(), "domain.channel") {
			t.Errorf("ParseChannel(email): %v (want the domain.channel code)", err)
		}
	})
	t.Run("outbound status incl. the inbound zero", func(t *testing.T) {
		for _, raw := range []string{"", "queued", "sent", "delivered", "failed"} {
			if got := MustOutboundStatus(raw); string(got) != raw {
				t.Errorf("MustOutboundStatus(%q) = %q", raw, got)
			}
		}
		if _, err := ParseOutboundStatus("yolo"); err == nil || !strings.Contains(err.Error(), "domain.outbound_status") {
			t.Errorf("ParseOutboundStatus(yolo): %v (want the domain.outbound_status code)", err)
		}
	})
	t.Run("fax status", func(t *testing.T) {
		for _, raw := range []string{"queued", "sending", "transmitted", "received", "failed"} {
			if got := MustFaxStatus(raw); string(got) != raw {
				t.Errorf("MustFaxStatus(%q) = %q", raw, got)
			}
		}
		if _, err := ParseFaxStatus("maybe"); err == nil || !strings.Contains(err.Error(), "domain.fax_status") {
			t.Errorf("ParseFaxStatus(maybe): %v (want the domain.fax_status code)", err)
		}
	})
	t.Run("fax direction", func(t *testing.T) {
		for _, raw := range []string{"in", "out"} {
			if got := MustFaxDirection(raw); string(got) != raw {
				t.Errorf("MustFaxDirection(%q) = %q", raw, got)
			}
		}
		if _, err := ParseFaxDirection("diagonal"); err == nil || !strings.Contains(err.Error(), "domain.fax_direction") {
			t.Errorf("ParseFaxDirection(diagonal): %v (want the domain.fax_direction code)", err)
		}
	})
}

// TestMustEnumsPanicOnCorruptRows pins the fail-fast contract: a value
// the writer never writes surfaces as a panic naming the enum, not as a
// silently rendered mystery state.
func TestMustEnumsPanicOnCorruptRows(t *testing.T) {
	for name, fn := range map[string]func(){
		"direction":     func() { _ = MustDirection("??") },
		"channel":       func() { _ = MustChannel("??") },
		"status":        func() { _ = MustOutboundStatus("??") },
		"fax status":    func() { _ = MustFaxStatus("??") },
		"fax direction": func() { _ = MustFaxDirection("??") },
	} {
		t.Run(name, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Fatalf("Must* on a corrupt %s value must panic", name)
				}
			}()
			fn()
		})
	}
}
