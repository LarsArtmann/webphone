package views

import (
	"strings"
	"testing"
	"time"
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
		// The screen-reader suffix names the destination: a bare "Ada
		// Lovelace" link in a heading does not say where it goes.
		"Open in the CRM",
	} {
		if !strings.Contains(page, want) {
			t.Errorf("linked name missing %q; got:\n%s", want, page)
		}
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
		if got := formatThousands(in); got != want {
			t.Errorf("formatThousands(%d) = %q, want %q", in, got, want)
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
