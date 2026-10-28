package domain

import "testing"

// ParseComposedAddress is the compose-form door: the two chat-network
// SHAPES route to their networks, everything else is the sms dialable
// lane (signal is compose-born-inbound by design — its phone-shaped
// addresses are indistinguishable from sms).
func TestParseComposedAddress(t *testing.T) {
	t.Parallel()

	cases := []struct {
		raw     string
		network Network
		value   string
	}{
		{"@peer:example.org", NetworkMatrix, "@peer:example.org"},
		{"  @lars:matrix.org  ", NetworkMatrix, "@lars:matrix.org"},
		{
			"05" + "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
			NetworkSession,
			"05" + "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
		},
		{"+44 1632 960961", NetworkSMS, "+441632960961"},
		{"*971001", NetworkSMS, "*971001"},
	}

	for _, tc := range cases {
		got, err := ParseComposedAddress(tc.raw)
		if err != nil {
			t.Errorf("ParseComposedAddress(%q): %v", tc.raw, err)
			continue
		}
		if got.Network() != tc.network || got.String() != tc.value {
			t.Errorf("ParseComposedAddress(%q) = %v/%q, want %v/%q", tc.raw, got.Network(), got.String(), tc.network, tc.value)
		}
	}

	// A session id with a non-hex tail is NOT a session address — it
	// falls through to the SMS rule, whose own verdict (here: the
	// 32-character dialable cap) is the answer the composer sees.
	badSession := "05" + "zz" + "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcd"
	got, err := ParseComposedAddress(badSession)
	if err == nil || got.Network() == NetworkSession {
		t.Fatalf("a non-hex session shape must not parse as session, got %v err=%v", got, err)
	}
}
