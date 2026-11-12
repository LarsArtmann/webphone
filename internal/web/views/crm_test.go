package views

import (
	"strings"
	"testing"
	"time"

	"github.com/dustin/go-humanize"
)

// TestCRMNameLinkIsTheOneLinkHome pins the deep-link contract from both
// sides: with a template the name is an anchor that opens the CRM in a NEW
// TAB (a same-tab navigation would unload the SIP island mid-call), and
// without one the name is plain text, so a zero-config deployment is
// byte-identical to the pre-integration product.
func TestCRMNameLinkIsTheOneLinkHome(t *testing.T) {
	t.Parallel()

	linked := CRMContacts{"+493012345678": {
		Name: "Ada Lovelace",
		URL:  "https://crm.example/contacts/01M",
	}}

	page := renderComponent(t, crmNameLink("+493012345678", linked, LangEN))
	for _, want := range []string{
		`<a class="wp-crm-link"`,
		`href="https://crm.example/contacts/01M"`,
		`target="_blank"`,
		`rel="noopener noreferrer"`,
		"Ada Lovelace",
		// The accessible NAME carries the destination. It rides aria-label,
		// not a visually-hidden suffix: the first attempt used a made-up
		// `wp-sr-only` class, so "(Open in the CRM)" rendered visibly inside
		// the 24px heading (caught by the visual harness, not by any gate).
		// aria-label cannot be styled wrong.
		`aria-label="Ada Lovelace, Open in the CRM"`,
	} {
		if !strings.Contains(page, want) {
			t.Errorf("linked name missing %q; got:\n%s", want, page)
		}
	}

	// Nothing may render the destination phrase as VISIBLE text: the link's
	// label is the name, and a hidden-text workaround is how the last defect
	// got in.
	if strings.Contains(page, ">(") {
		t.Errorf("the link must not carry a visible parenthesised suffix:\n%s", page)
	}

	plain := renderComponent(t, crmNameLink("+493012345678",
		CRMContacts{"+493012345678": {Name: "Ada Lovelace"}}, LangEN))
	if strings.Contains(plain, "<a") {
		t.Errorf("a contact without a URL must render no anchor:\n%s", plain)
	}
	if !strings.Contains(plain, "Ada Lovelace") {
		t.Errorf("plain name lost: %q", plain)
	}

	unknown := renderComponent(t, crmNameLink("+44999999", linked, LangEN))
	if strings.Contains(unknown, "<a") || !strings.Contains(unknown, "+44999999") {
		t.Errorf("an unknown number must render the raw number, no link: %q", unknown)
	}
}

// TestCRMCreateLinkArms pins the create affordance's three arms: an
// absent template renders NOTHING (the byte-identical zero-config
// product), a MISS (a number the CRM does not know) renders the anchor
// with the {phone} placeholder substituted query-escaped (the leading
// "+" is a query-param hazard), and a MATCHED number renders NOTHING —
// its profile link is the affordance there, and two links to the same
// CRM surface in one heading would be noise.
func TestCRMCreateLinkArms(t *testing.T) {
	t.Parallel()

	const tmpl = "https://crm.example/contacts/new?phone={phone}"
	linked := CRMContacts{"+493012345678": {
		Name: "Ada Lovelace",
		URL:  "https://crm.example/contacts/01M",
	}}

	absent := renderComponent(t, crmCreateLink("+44999999", "", linked, LangEN))
	if absent != "" {
		t.Errorf("empty template must render nothing, got:\n%s", absent)
	}

	miss := renderComponent(t, crmCreateLink("+44999999", tmpl, linked, LangEN))
	for _, want := range []string{
		`<a class="wp-crm-create"`,
		`href="https://crm.example/contacts/new?phone=%2B44999999"`,
		`target="_blank"`,
		`rel="noopener noreferrer"`,
		"New contact in the CRM",
	} {
		if !strings.Contains(miss, want) {
			t.Errorf("create-link miss missing %q; got:\n%s", want, miss)
		}
	}
	if strings.Contains(miss, "{phone}") {
		t.Errorf("the placeholder must never survive into markup:\n%s", miss)
	}

	matched := renderComponent(t, crmCreateLink("+493012345678", tmpl, linked, LangEN))
	if matched != "" {
		t.Errorf("a CRM-matched number must render no create link (its profile link is the affordance), got:\n%s", matched)
	}
}

// TestCRMContextStripRendersNothingWithoutContext is the degradation
// contract: absence means NO markup at all, never an empty shell. A bordered
// band with no facts would be worse than no band.
func TestCRMContextStripRendersNothingWithoutContext(t *testing.T) {
	t.Parallel()

	if got := renderComponent(t, crmContextStrip(nil, LangEN)); got != "" {
		t.Fatalf("nil context must render nothing, got %q", got)
	}
}

// TestCRMContextStripRendersThePhonesSecondFactClass: the strip is the whole
// reason the train exists, so every fact it can carry is pinned here.
func TestCRMContextStripRendersThePhonesSecondFactClass(t *testing.T) {
	t.Parallel()

	ctx := &CRMContext{
		Company: "Analytical Engines Ltd",
		Deals: []CRMDeal{
			{Title: "Retainer 2027", Stage: "proposal", ValueCents: 1200000, Currency: "EUR"},
		},
		MoreDeals:    3,
		OpenTasks:    2,
		NextTask:     "Send the revised offer",
		NextTaskDue:  time.Date(2026, 11, 2, 9, 0, 0, 0, time.UTC),
		LastActivity: time.Now().Add(-5 * time.Minute),
	}

	page := renderComponent(t, crmContextStrip(ctx, LangEN))
	for _, want := range []string{
		`id="wp-crm-context"`,
		`aria-label="CRM context"`,
		"Analytical Engines Ltd",
		"Retainer 2027",
		"proposal",
		"EUR 12,000",
		"+3",
		"Send the revised offer",
		"due 2 Nov",
		"5m",
	} {
		if !strings.Contains(page, want) {
			t.Errorf("strip missing %q; got:\n%s", want, page)
		}
	}

	if strings.Contains(page, "overdue") {
		t.Errorf("a future task must not read overdue:\n%s", page)
	}
}

// TestCRMContextStripMarksOverdueTasks: the CRM computes the verdict against
// its own clock; the strip only renders it. Rendering "due 2 Nov" for a task
// that is already late would be the exact silence the honesty contract bans.
func TestCRMContextStripMarksOverdueTasks(t *testing.T) {
	t.Parallel()

	ctx := &CRMContext{
		OpenTasks:   1,
		NextTask:    "Send the revised offer",
		NextTaskDue: time.Date(2020, 1, 2, 9, 0, 0, 0, time.UTC),
		Overdue:     true,
	}

	page := renderComponent(t, crmContextStrip(ctx, LangEN))
	if !strings.Contains(page, "overdue") {
		t.Fatalf("overdue task not marked: %s", page)
	}
	if strings.Contains(page, "due 2 Jan") {
		t.Fatalf("an overdue task must not claim a plain due date: %s", page)
	}
}

// TestCRMContextStripEscapesOperatorData: CRM text is operator data and must
// never become markup (the CSP forbids inline handlers, but escaping is the
// real boundary).
func TestCRMContextStripEscapesOperatorData(t *testing.T) {
	t.Parallel()

	ctx := &CRMContext{Company: `Acme <script>alert(1)</script>`}
	page := renderComponent(t, crmContextStrip(ctx, LangEN))

	if strings.Contains(page, "<script>") {
		t.Fatalf("CRM text reached the markup unescaped: %s", page)
	}
	if !strings.Contains(page, "&lt;script&gt;") {
		t.Fatalf("escaped company text lost: %s", page)
	}
}

// TestVoicemailRowLinksTheKnownCRMContact pins the CRM data-mesh
// consistency fix: a caller the CRM resolves renders the LINKED name
// inside the row's strong (the same crmNameLink contract as every other
// surface — new tab, noopener), the vmCaller ladder otherwise. Zero
// config and CRM misses must not grow an anchor.
func TestVoicemailRowLinksTheKnownCRMContact(t *testing.T) {
	t.Parallel()

	linked := CRMContacts{"+441632960961": {
		Name: "Alice",
		URL:  "https://crm.example/contacts/01M",
	}}

	msg := VoicemailItem{UUID: "u1", CIDNumber: "+441632960961", CIDName: "Alice", Seconds: 12,
		AudioURL: "/phone-api/voicemail/1001/messages/u1/audio"}

	page := renderComponent(t, VoicemailRow(msg, linked, LangEN, false))
	for _, want := range []string{
		`<a class="wp-crm-link"`,
		`href="https://crm.example/contacts/01M"`,
		`target="_blank"`,
		"Alice",
	} {
		if !strings.Contains(page, want) {
			t.Errorf("linked voicemail caller missing %q; got:\n%s", want, page)
		}
	}

	// A name the CRM does NOT know (provider CIDName only) keeps the
	// unlinked ladder — display text with no profile must never anchor.
	miss := renderComponent(t, VoicemailRow(msg, nil, LangEN, false))
	if strings.Contains(miss, "<a class") || !strings.Contains(miss, "Alice (+441632960961)") {
		t.Errorf("CRM miss must keep the vmCaller ladder without a link:\n%s", miss)
	}

	anonymous := VoicemailItem{UUID: "u2", CIDName: "Withheld", Seconds: 8,
		AudioURL: "/phone-api/voicemail/1001/messages/u2/audio"}
	anon := renderComponent(t, VoicemailRow(anonymous, linked, LangEN, false))
	if strings.Contains(anon, "<a class") || !strings.Contains(anon, "Withheld") || strings.Contains(anon, "Withheld ()") {
		t.Errorf("a withheld caller renders the bare name, no link, no empty parens:\n%s", anon)
	}
}

// TestCRMDraftContactLinksTheKnownCaller pins the drafts panel's link
// arm (the last of the five CRM-name surfaces): the CRM-resolved
// contact renders linked with the dialable number kept beside it, the
// contactLabel ladder otherwise. Zero CRM config stays byte-identical.
func TestCRMDraftContactLinksTheKnownCaller(t *testing.T) {
	t.Parallel()

	linked := CRMContacts{"493012345678": {
		Name: "Ada Lovelace",
		URL:  "https://crm.example/contacts/01M",
	}}

	page := renderComponent(t, crmDraftContact(linked, "493012345678", LangEN))
	for _, want := range []string{
		`<a class="wp-crm-link"`,
		`href="https://crm.example/contacts/01M"`,
		"Ada Lovelace",
		"+493012345678",
	} {
		if !strings.Contains(page, want) {
			t.Errorf("linked draft contact missing %q; got:\n%s", want, page)
		}
	}

	// Known name, no URL template: the ladder keeps name · +number, no anchor.
	nameOnly := CRMContacts{"493012345678": {Name: "Ada Lovelace"}}
	plain := renderComponent(t, crmDraftContact(nameOnly, "493012345678", LangEN))
	if strings.Contains(plain, "<a") || !strings.Contains(plain, "Ada Lovelace · +493012345678") {
		t.Errorf("name-without-URL must keep the ladder without a link:\n%s", plain)
	}

	unknown := renderComponent(t, crmDraftContact(linked, "44999999", LangEN))
	if strings.Contains(unknown, "<a") || !strings.Contains(unknown, "+44999999") {
		t.Errorf("unknown number must render the ladder, no link:\n%s", unknown)
	}
}

func TestFormatThousands(t *testing.T) {
	t.Parallel()
	cases := map[int64]string{
		0:          "0",
		7:          "7",
		999:        "999",
		1000:       "1,000",
		12000:      "12,000",
		1234567:    "1,234,567",
		-12000:     "-12,000",
		-999:       "-999",
		1000000000: "1,000,000,000",
	}

	for in, want := range cases {
		if got := humanize.Comma(in); got != want {
			t.Errorf("humanize.Comma(%d) = %q, want %q", in, got, want)
		}
	}
}

func TestDealValue(t *testing.T) {
	t.Parallel()

	cases := []struct {
		deal CRMDeal
		want string
	}{
		{CRMDeal{ValueCents: 1200000, Currency: "EUR"}, "EUR 12,000"},
		{CRMDeal{ValueCents: 1200050, Currency: "EUR"}, "EUR 12,000.50"},
		{CRMDeal{ValueCents: 99}, "EUR 0.99"}, // empty currency falls back
		{CRMDeal{ValueCents: 0, Currency: "GBP"}, "GBP 0"},
		{CRMDeal{ValueCents: 123456789, Currency: "USD"}, "USD 1,234,567.89"},
	}

	for _, tc := range cases {
		if got := dealValue(tc.deal); got != tc.want {
			t.Errorf("dealValue(%+v) = %q, want %q", tc.deal, got, tc.want)
		}
	}
}

func TestCRMTaskDue(t *testing.T) {
	t.Parallel()

	none := CRMContext{}
	if got := crmTaskDue(none, LangEN); got != "" {
		t.Errorf("no task must render no due phrase, got %q", got)
	}

	overdue := CRMContext{NextTask: "x", Overdue: true}
	if got := crmTaskDue(overdue, LangEN); got != "overdue" {
		t.Errorf("overdue phrase = %q", got)
	}
	if got := crmTaskDue(overdue, LangDE); got != "überfällig" {
		t.Errorf("German overdue phrase = %q", got)
	}

	dated := CRMContext{NextTask: "x", NextTaskDue: time.Date(2026, 11, 2, 9, 0, 0, 0, time.UTC)}
	if got := crmTaskDue(dated, LangEN); got != "due 2 Nov" {
		t.Errorf("due phrase = %q", got)
	}

	undated := CRMContext{NextTask: "x"}
	if got := crmTaskDue(undated, LangEN); got != "" {
		t.Errorf("an undated task must render no due phrase, got %q", got)
	}
}

// TestCrmDueLabelLocalizesTheMonth pins the DE-locale fix: the due
// phrase names the month in the PANEL language ("2 Jan" spells the
// English month in Go regardless of the desk).
func TestCrmDueLabelLocalizesTheMonth(t *testing.T) {
	due := CRMContext{NextTask: "Follow up", NextTaskDue: time.Date(2026, 11, 4, 9, 0, 0, 0, time.UTC)}
	if got := crmTaskDue(due, LangDE); got != "fällig 4. Nov" {
		t.Fatalf("DE due = %q, want %q", got, "fällig 4. Nov")
	}
	if got := crmTaskDue(due, LangEN); got != "due 4 Nov" {
		t.Fatalf("EN due = %q, want %q", got, "due 4 Nov")
	}
}
