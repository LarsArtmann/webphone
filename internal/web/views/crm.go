package views

import (
	"fmt"
	"strings"
	"time"
)

// CRMContact is the optional CRM integration's answer for one phone
// number, as a SURFACE renders it: the published display name and — only
// when the deployment configured a deep-link template — the URL of that
// contact in the CRM's own UI.
//
// This is a projection, never a copy. No CRM id reaches the view, so no
// render path can invent a link the CRM never named, and nothing a surface
// produces can be persisted as a replica of a contact record. The CRM owns
// the contact; the phone owns the conversation.
type CRMContact struct {
	Name string
	URL  string
}

// CRMContacts indexes the integration's answers by the number the surface
// keys on. A number ABSENT from the map is simply unknown — not empty —
// so displayName's raw-number fallback needs no extra policy, and an
// unknown caller renders exactly as it did before the integration existed.
type CRMContacts map[string]CRMContact

// CRMDeal is one open CRM deal, reduced to what a strip can say at a
// glance. Value stays integer cents: the CRM's journal is EUR-cents at
// rest, and inventing sub-cent precision here would be a lie.
type CRMDeal struct {
	Title      string
	Stage      string
	ValueCents int64
	Currency   string
}

// CRMContext is the CRM's published context for ONE contact: who they are
// inside the CRM and what is outstanding with them. Rendered only on the
// surfaces where a human studies a single conversation — never on a list,
// where N contacts × M facts would be an N+1 storm for data nobody reads
// at list density.
//
// A nil *CRMContext means "nothing to say": the integration is off, the
// number matches no contact, the CRM does not publish context, or it is
// unreachable. The surface then renders no strip at all, which is the same
// product a CRM-less deployment has always had.
type CRMContext struct {
	Company      string
	Deals        []CRMDeal
	MoreDeals    int
	OpenTasks    int
	NextTask     string
	NextTaskDue  time.Time
	Overdue      bool
	LastActivity time.Time
}

// crmProfileURL is the number's CRM URL, or "" when the deployment
// configured no deep-link template (or the number is unknown). An empty
// answer is the common case: zero CRM config must render byte-identically
// to the pre-integration product, so every link site degrades to text.
func crmProfileURL(number string, contacts CRMContacts) string {
	contact, ok := contacts[number]
	if !ok {
		return ""
	}

	return contact.URL
}

// dealValue renders a deal's size without locale games: thousands counts
// the cents the CRM journals, then the currency code. No symbol table, no
// locale formatting — a wrong € for a £ is worse than a plain code.
func dealValue(deal CRMDeal) string {
	currency := deal.Currency
	if currency == "" {
		currency = "EUR"
	}

	major := deal.ValueCents / 100
	if deal.ValueCents%100 != 0 {
		return fmt.Sprintf("%s %s.%02d", currency, formatThousands(major), deal.ValueCents%100)
	}

	return currency + " " + formatThousands(major)
}

// formatThousands groups an integer with no locale dependency (no i18n money
// library in this build, and a wrong symbol is worse than a plain code). The
// sign is stripped FIRST: grouping the sign as a digit produced "-,999".
func formatThousands(n int64) string {
	digits := fmt.Sprintf("%d", n)

	sign := ""
	if strings.HasPrefix(digits, "-") {
		sign, digits = "-", digits[1:]
	}

	if len(digits) <= 3 {
		return sign + digits
	}

	var b strings.Builder

	head := len(digits) % 3
	if head > 0 {
		b.WriteString(digits[:head])
	}

	for i := head; i < len(digits); i += 3 {
		if b.Len() > 0 {
			b.WriteString(",")
		}

		b.WriteString(digits[i : i+3])
	}

	return sign + b.String()
}

// crmTaskDue is the strip's due phrase: "overdue" or "due 2 Nov". A plain
// short date rather than "in 3 days" keeps the phrase deterministic (no
// wall-clock dependence in the render), plural-free (the i18n dictionaries
// hold plain strings, not format templates) and honest about what the CRM
// published, which is a timestamp, not a promise.
func crmTaskDue(ctx CRMContext, lang Lang) string {
	if ctx.NextTask == "" {
		return ""
	}

	if ctx.Overdue {
		return T(lang, "crm.overdue")
	}

	// An undated task has no due phrase at all: formatting the zero time
	// would print "due 1 Jan" (year 1) and make the UI look broken.
	if ctx.NextTaskDue.IsZero() {
		return ""
	}

	return T(lang, "crm.duePrefix") + " " + ctx.NextTaskDue.Format("2 Jan")
}
