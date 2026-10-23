// Sender-registry pins (chat-networks train F): the messaging service
// routes by the destination address's network — the telephony gateway
// owns the sms lane by construction, chat senders register behind their
// config gates, and a network with no sender is an honest Rejection,
// never a silent drop.
package messaging_test

import (
	"context"
	"testing"

	errorfamily "github.com/larsartmann/go-error-family"
	errorfamilytest "github.com/larsartmann/go-error-family/errorfamilytest"
	"github.com/larsartmann/webphone/internal/blob"
	"github.com/larsartmann/webphone/internal/domain"
	"github.com/larsartmann/webphone/internal/gateway"
	"github.com/larsartmann/webphone/internal/messaging"
	"github.com/larsartmann/webphone/internal/store"
)

// chatGateway records the outbound message it was handed.
type chatGateway struct {
	calls int
	last  gateway.OutboundMessage
}

func (g *chatGateway) SendMessage(_ context.Context, msg gateway.OutboundMessage) (gateway.Receipt, error) {
	g.calls++
	g.last = msg
	return gateway.Receipt{ProviderRef: "chat-ref-1", Resolution: gateway.ResolutionImmediate}, nil
}

func newSenderTestService(t *testing.T) (*messaging.Service, *chatGateway, *countingGateway) {
	t.Helper()
	db, err := store.Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	blobs, err := blob.New(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	telephony := &countingGateway{}
	svc := messaging.New(store.NewMessages(db), blobs, telephony, nil, nil)
	chat := &chatGateway{}
	return svc, chat, telephony
}

func TestRegisterSenderGuardsTheLanes(t *testing.T) {
	svc, chat, _ := newSenderTestService(t)

	if err := svc.RegisterSender(domain.NetworkSMS, chat); err == nil {
		t.Fatal("registering a second sms sender must be refused — the telephony gateway owns that lane")
	}
	if err := svc.RegisterSender(domain.NetworkMatrix, chat); err != nil {
		t.Fatalf("matrix registration: %v", err)
	}
	if err := svc.RegisterSender(domain.NetworkMatrix, chat); err == nil {
		t.Fatal("double registration must be refused — a wiring defect, not a preference")
	}
	if err := svc.RegisterSender(domain.Network("wlan"), chat); err == nil {
		t.Fatal("an unknown network must be refused at the door")
	}
}

func TestSendRoutesByNetwork(t *testing.T) {
	svc, chat, telephony := newSenderTestService(t)
	if err := svc.RegisterSender(domain.NetworkMatrix, chat); err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	owner := domain.MustParseExtension("1001")

	msg, err := svc.Send(ctx, owner, domain.MustParseAddress(domain.NetworkMatrix, "@peer:example.org"), "hello room", nil)
	if err != nil {
		t.Fatalf("matrix send: %v", err)
	}
	if chat.calls != 1 || chat.last.To.String() != "@peer:example.org" {
		t.Fatalf("chat gateway must carry the mxid, got %d calls, last %+v", chat.calls, chat.last)
	}
	if telephony.calls != 0 {
		t.Fatal("the telephony gateway must not see chat traffic")
	}
	if msg.ProviderRef != "chat-ref-1" || msg.Status != domain.StatusSent {
		t.Fatalf("immediate chat receipt must settle the row, got %s/%s", msg.ProviderRef, msg.Status)
	}

	// The same shape over telephony keeps its sms lane.
	if _, err := svc.Send(ctx, owner, domain.MustParseAddress(domain.NetworkSMS, "+441632960961"), "hello pstn", nil); err != nil {
		t.Fatalf("sms send: %v", err)
	}
	if telephony.calls != 1 || chat.calls != 1 {
		t.Fatalf("lane isolation broke: telephony=%d chat=%d", telephony.calls, chat.calls)
	}

	// An unregistered network is a user-facing refusal.
	_, err = svc.Send(ctx, owner, domain.MustParseAddress(domain.NetworkSignal, "+441632960961"), "anyone?", nil)
	if err == nil {
		t.Fatal("a network without a sender must refuse the send")
	}
	errorfamilytest.AssertFamily(t, err, errorfamily.Rejection)
}

// TestChatSendPersistsOntoItsOwnThread pins the routing's persistence
// side: two networks, same remote value, two threads — the twin-thread
// rule the v5 migration bought.
func TestChatSendPersistsOntoItsOwnThread(t *testing.T) {
	svc, chat, _ := newSenderTestService(t)
	if err := svc.RegisterSender(domain.NetworkSignal, chat); err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	owner := domain.MustParseExtension("1001")
	number := "+441632960961"

	if _, err := svc.Send(ctx, owner, domain.MustParseAddress(domain.NetworkSMS, number), "via pstn", nil); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Send(ctx, owner, domain.MustParseAddress(domain.NetworkSignal, number), "via signal", nil); err != nil {
		t.Fatal(err)
	}

	threads, err := svc.Threads(ctx, owner)
	if err != nil {
		t.Fatal(err)
	}
	if len(threads) != 2 {
		t.Fatalf("same number on two networks must be two threads, got %d", len(threads))
	}
	networks := map[domain.Network]bool{}
	for _, summary := range threads {
		networks[summary.Thread.Remote.Network()] = true
	}
	if !networks[domain.NetworkSMS] || !networks[domain.NetworkSignal] {
		t.Fatalf("want one sms and one signal thread, got %v", networks)
	}
}
