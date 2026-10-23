package push_test

import (
	"context"
	"errors"
	"testing"

	"github.com/larsartmann/go-error-family"
	errorfamilytest "github.com/larsartmann/go-error-family/errorfamilytest"
	"github.com/larsartmann/webphone/internal/push"
)

// fakeSender records sends; enabled toggles its liveness.
type fakeSender struct {
	enabled bool
	got     []push.Ring
	err     error
}

func (f *fakeSender) Enabled() bool { return f.enabled }
func (f *fakeSender) Describe() string { return "fake" }
func (f *fakeSender) Send(_ context.Context, _ string, ring push.Ring) error {
	f.got = append(f.got, ring)
	return f.err
}

func TestFanoutRoutesPerTransport(t *testing.T) {
	ntfy := &fakeSender{enabled: true}
	fcm := &fakeSender{enabled: true}
	fanout := push.NewFanout(map[string]push.Sender{
		push.TransportNtfy: ntfy,
		push.TransportFcm:  fcm,
	})
	if !fanout.Enabled() || fanout.Describe() != "fcm+ntfy" {
		t.Fatalf("fanout %q enabled=%v", fanout.Describe(), fanout.Enabled())
	}
	if !fanout.Knows(push.TransportNtfy) || fanout.Knows("apns") {
		t.Fatal("Knows must reflect the live senders")
	}

	ring := push.Ring{Event: push.EventRing, From: "+15550000001", CallID: "c1"}
	if err := fanout.Send(context.Background(), push.TransportNtfy, "topic", ring); err != nil {
		t.Fatal(err)
	}
	if len(ntfy.got) != 1 || len(fcm.got) != 0 {
		t.Fatalf("ntfy got %d, fcm got %d", len(ntfy.got), len(fcm.got))
	}

	err := fanout.Send(context.Background(), "apns", "t", ring)
	if err == nil {
		t.Fatal("unknown transport must reject")
	}
	errorfamilytest.AssertFamily(t, err, errorfamily.Rejection)
	errorfamilytest.AssertCode(t, err, "push.transport")
}

func TestFanoutDropsDeadSenders(t *testing.T) {
	dead := &fakeSender{enabled: false}
	fanout := push.NewFanout(map[string]push.Sender{push.TransportNtfy: dead})
	if fanout.Enabled() {
		t.Fatal("a fanout of dead senders is disabled")
	}
	if err := fanout.Send(context.Background(), push.TransportNtfy, "t", push.Ring{}); !errors.Is(err, push.ErrDisabled) {
		t.Fatalf("error = %v, want ErrDisabled", err)
	}
	var nilFanout *push.Fanout
	if nilFanout.Enabled() || nilFanout.Describe() != "off" {
		t.Fatal("nil fanout must behave as disabled")
	}
}
