package server

import (
	"net/http"
	"regexp"
	"strings"
	"testing"

	"github.com/larsartmann/webphone/internal/crm"
	"github.com/larsartmann/webphone/internal/domain"
)

// TestCRMContactsProjectsNameAndDeepLink pins the seam-to-view boundary: the
// view projection carries the CRM's published NAME always, and a URL only when
// this deployment configured a deep-link template. No CRM id survives into the
// view, so a surface cannot mint a link the CRM never named.
func TestCRMContactsProjectsNameAndDeepLink(t *testing.T) {
	t.Parallel()

	people := map[string]crm.Person{
		"+493012345678": {Ref: crm.NewContactRef("01M"), Name: "Ada Lovelace"},
	}

	linked := crmContacts(people, crm.NewResolver(nil, "https://crm.example/contacts/{id}", nil))
	if got := linked["+493012345678"]; got.Name != "Ada Lovelace" || got.URL != "https://crm.example/contacts/01M" {
		t.Fatalf("linked projection: %+v", got)
	}

	plain := crmContacts(people, crm.NewResolver(nil, "", nil))
	if got := plain["+493012345678"]; got.Name != "Ada Lovelace" || got.URL != "" {
		t.Fatalf("no template must yield no URL: %+v", got)
	}

	if nilResolver := crmContacts(people, nil); nilResolver["+493012345678"].URL != "" {
		t.Fatalf("a nil resolver must never mint a URL: %+v", nilResolver)
	}

	if empty := crmContacts(nil, nil); empty != nil {
		t.Fatalf("no people must project to nil, got %v", empty)
	}
}

const stubContextJSON = `{
  "id": "01M",
  "name": "Ada Lovelace",
  "company": "Analytical Engines Ltd",
  "open_deals": [{"title": "Retainer 2027", "stage": "proposal", "value_cents": 1200000, "currency": "EUR"}],
  "open_deal_count": 3,
  "open_tasks": 2,
  "next_task": {"title": "Send the revised offer", "due_at": "2026-11-02T09:00:00Z"},
  "last_activity_at": "2026-10-24T11:03:00Z"
}`

// openThreadWithCRM seeds one conversation with the CRM's known number and
// renders its partial through a server whose CRM publishes stubContextJSON.
func openThreadWithCRM(t *testing.T, stub *crmStub) string {
	t.Helper()

	resolver := crmResolverFor(t, stub)
	server := newTestServerWithPhoneAPI(t, "", func(d *Deps) { d.CRM = resolver })
	c := signIn(t, server)

	deliverInbound(t, server, "+493012345678", "hello from Ada")

	_, list := c.do(http.MethodGet, "/partials/messages", nil, "")
	rows := regexp.MustCompile(`class="wp-thread-row[^"]*" href="/messages/([^"]+)"`).FindAllSubmatch(list, -1)
	if len(rows) != 1 {
		t.Fatalf("expected one thread, got %d: %.400s", len(rows), list)
	}

	_, body := c.do(http.MethodGet, "/partials/messages/"+string(rows[0][1]), nil, "")

	return string(body)
}

// TestThreadPanelRendersTheCRMContextStrip is the end-to-end proof of this
// train's whole point: opening a conversation with a known contact renders the
// CRM's published context (company, deal, outstanding task, last touch) beside
// the head, from ONE upstream call.
func TestThreadPanelRendersTheCRMContextStrip(t *testing.T) {
	stub := &crmStub{contextJSON: stubContextJSON, contextStatus: http.StatusOK}
	view := openThreadWithCRM(t, stub)

	for _, want := range []string{
		`id="wp-crm-context"`,
		"Analytical Engines Ltd",
		"Retainer 2027",
		"EUR 12,000",
		"+2", // open_deal_count 3, one deal shipped in open_deals
		"Send the revised offer",
		// The contact NAME still reaches the head through the same seam.
		"Ada Lovelace",
	} {
		if !strings.Contains(view, want) {
			t.Errorf("thread view missing %q; got:\n%.1200s", want, view)
		}
	}
}

// TestThreadPanelOmitsTheStripWhenTheCRMHasNoContext: a 404 (a CRM build
// without the product, or a vanished contact) must render NO strip and must
// not degrade the conversation. The name still renders.
func TestThreadPanelOmitsTheStripWhenTheCRMHasNoContext(t *testing.T) {
	stub := &crmStub{contextStatus: http.StatusNotFound}
	view := openThreadWithCRM(t, stub)

	if strings.Contains(view, `id="wp-crm-context"`) {
		t.Fatalf("a 404 must render no strip at all:\n%.1200s", view)
	}
	if !strings.Contains(view, "Ada Lovelace") {
		t.Fatalf("the contact name must still resolve:\n%.1200s", view)
	}
}

// TestThreadPanelToleratesAnUnreachableCRM: a 5xx on the context product is a
// degradation, never an error surface in the phone's UI.
func TestThreadPanelToleratesAnUnreachableCRM(t *testing.T) {
	stub := &crmStub{contextStatus: http.StatusInternalServerError}
	view := openThreadWithCRM(t, stub)

	if strings.Contains(view, `id="wp-crm-context"`) {
		t.Fatalf("a failing CRM must render no strip:\n%.1200s", view)
	}
	if strings.Contains(view, "wp-error") {
		t.Fatalf("a decorative seam failure must not surface as an error banner:\n%.1200s", view)
	}
}

// TestCRMContextRefusesNonPhoneShapedRemotes pins the telephony guard at the
// seam's door: the CRM indexes phone numbers, so a Matrix user id or a Session
// id must never cost a request. Signal IS phone-shaped (its addresses are
// numbers), so it stays eligible — otherwise a Signal conversation would show
// a linked name from the identity product and no context, for no reason a user
// could understand.
func TestCRMContextRefusesNonPhoneShapedRemotes(t *testing.T) {
	t.Parallel()

	session := "05" + strings.Repeat("ab", 32)

	cases := []struct {
		name    string
		network domain.Network
		raw     string
		wantNil bool
	}{
		{"matrix", domain.NetworkMatrix, "@ada:example.org", true},
		{"session", domain.NetworkSession, session, true},
		{"sms", domain.NetworkSMS, "+493012345678", false},
		{"signal", domain.NetworkSignal, "+493012345678", false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			stub := &crmStub{contextJSON: stubContextJSON, contextStatus: http.StatusOK}
			h := &handlers{deps: Deps{CRM: crmResolverFor(t, stub)}}
			addr := domain.MustParseAddress(tc.network, tc.raw)

			got := h.crmContextFor(t.Context(), addr)

			if tc.wantNil && got != nil {
				t.Fatalf("%s must yield no strip, got %+v", tc.name, got)
			}
			if !tc.wantNil && got == nil {
				t.Fatalf("%s must yield a strip", tc.name)
			}

			stub.mu.Lock()
			lookups := len(stub.lookupNumbers)
			stub.mu.Unlock()

			wantLookups := 1
			if tc.wantNil {
				wantLookups = 0
			}
			if lookups != wantLookups {
				t.Fatalf("%s made %d CRM lookups, want %d", tc.name, lookups, wantLookups)
			}
		})
	}
}

// TestCRMNumberForAnswersOnlyPhoneShapedRemotes keeps the identity feed and the
// context strip on ONE rule. Before this, a Matrix thread's raw address reached
// /api/contacts/by-phone, which answers 400 for a value with no digits — a
// failure the resolver does not cache, so every render repeated it.
func TestCRMNumberForAnswersOnlyPhoneShapedRemotes(t *testing.T) {
	t.Parallel()

	if got := crmNumberFor(domain.MustParseAddress(domain.NetworkMatrix, "@ada:example.org")); got != "" {
		t.Fatalf("matrix remote must not be asked about, got %q", got)
	}
	if got := crmNumberFor(domain.MustParseAddress(domain.NetworkSMS, "+49 30 12345678")); got != "+493012345678" {
		t.Fatalf("sms remote = %q", got)
	}
	if got := crmNumberFor(domain.MustParseAddress(domain.NetworkSignal, "+493012345678")); got != "+493012345678" {
		t.Fatalf("signal remote = %q", got)
	}
}

// TestMatrixThreadNeverReachesTheCRM is the wiring pin: crmNumberFor being
// correct is worth nothing if a panel still passes raw remotes. A Matrix
// conversation renders through the real partial path here, and the stub must
// see ZERO lookups.
func TestMatrixThreadNeverReachesTheCRM(t *testing.T) {
	stub := &crmStub{}
	resolver := crmResolverFor(t, stub)

	var deps *Deps

	server := newTestServerWithPhoneAPI(t, "", func(d *Deps) {
		d.CRM = resolver
		deps = d
	})
	c := signIn(t, server)

	matrix := domain.MustParseAddress(domain.NetworkMatrix, "@ada:example.org")
	if _, err := deps.Messaging.Receive(t.Context(), domain.InboundMessage{
		Owner: domain.MustParseExtension("1001"),
		From:  matrix,
		Body:  "hello from a chat network",
	}); err != nil {
		t.Fatalf("seed matrix thread: %v", err)
	}

	deliverInbound(t, server, "+493012345678", "hello from a phone")

	resp, list := c.do(http.MethodGet, "/partials/messages", nil, "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("messages panel: %d", resp.StatusCode)
	}
	if !strings.Contains(string(list), "hello from a chat network") {
		t.Fatalf("the matrix thread must render:\n%.800s", list)
	}

	stub.mu.Lock()
	asked := append([]string(nil), stub.lookupNumbers...)
	stub.mu.Unlock()

	for _, number := range asked {
		if number == matrix.String() {
			t.Fatalf("the matrix address reached the CRM lookup: %v", asked)
		}
	}
	if len(asked) != 1 || asked[0] != "+493012345678" {
		t.Fatalf("lookups = %v, want exactly the phone-shaped remote", asked)
	}
}
