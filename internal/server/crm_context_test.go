package server

import (
	"net/http"
	"regexp"
	"strings"
	"testing"

	"github.com/larsartmann/webphone/internal/crm"
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
