package views

import (
	"strings"
	"testing"
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
