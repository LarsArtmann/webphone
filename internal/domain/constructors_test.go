package domain

import (
	"testing"
	"time"
)

var constructorAt = time.Unix(1760000000, 0)

// The constructors' contract: direction, channel and status are
// consequences of the construction, never parameters — the illegal
// pairs (inbound row with a lifecycle, outbound row without one,
// channel disagreeing with the attachments) have no call site shape
// that can express them.
func TestMessageConstructorsMakeIllegalPairsUnbuildable(t *testing.T) {
	out := NewOutboundMessage(
		GenerateMessageID(), GenerateThreadID(),
		MustParseExtension("1001"), MustParsePhone("+441632960961"),
		"hi", 2, constructorAt,
	)
	if out.Direction != DirectionOutbound || out.Status != StatusQueued || out.Channel != ChannelMMS {
		t.Errorf("outbound ctor: %s/%s/%s, want out/queued/mms", out.Direction, out.Status, out.Channel)
	}

	in := NewInboundMessage(
		GenerateMessageID(), GenerateThreadID(),
		MustParseExtension("1001"), MustParsePhone("+441632960961"),
		"hi", 0, constructorAt,
	)
	if in.Direction != DirectionInbound || in.Status != "" || in.Channel != ChannelSMS {
		t.Errorf("inbound ctor: %s/%s/%s, want in/zero-status/sms — a lifecycle on an inbound row must be unbuildable", in.Direction, in.Status, in.Channel)
	}
}

func TestFaxConstructorsPinBirthStates(t *testing.T) {
	out := NewOutboundFaxJob(GenerateFaxID(), MustParseExtension("1001"), MustParsePhone("+441632960961"), "faxes/x.pdf", constructorAt)
	if out.Direction != FaxOutbound || out.Status != FaxQueued || !out.CreatedAt.Equal(out.UpdatedAt) {
		t.Errorf("outbound fax ctor: %s/%s, want out/queued with equal birth stamps", out.Direction, out.Status)
	}

	in := NewInboundFaxJob(GenerateFaxID(), MustParseExtension("1001"), MustParsePhone("+441632960961"), "faxes/y.pdf", "ref-7", 3, constructorAt)
	if in.Direction != FaxInbound || in.Status != FaxReceived || in.Pages != 3 || in.ProviderRef != "ref-7" {
		t.Errorf("inbound fax ctor: %s/%s pages=%d ref=%q, want in/received/3/ref-7", in.Direction, in.Status, in.Pages, in.ProviderRef)
	}
}
