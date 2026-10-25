package views

import (
	"strings"
	"testing"
)

// TestLoginCardAdaptsToPasskeyMode pins the login card's two shapes:
// off (the default) renders EXACTLY the extension login — no passkey
// ids anywhere — while on renders the /login front-door link (the
// cqrs-htmx loginpage owns the ceremony; the island card only links to
// it) and demotes the extension login into the break-glass disclosure
// without losing a single one of its DOM-contract ids.
func TestLoginCardAdaptsToPasskeyMode(t *testing.T) {
	off := renderComponent(t, PhoneIsland(PhoneIslandProps{}))
	on := renderComponent(t, PhoneIsland(PhoneIslandProps{Passkey: true}))

	link := `id="passkey-login"`
	if strings.Contains(off, link) {
		t.Errorf("passkey-off card renders %s — a disabled deployment must stay byte-shaped like the pre-passkey card", link)
	}
	if !strings.Contains(on, link) {
		t.Errorf("passkey-on card lacks %s — the /login front door is incomplete:\n%s", link, on)
	}
	if !strings.Contains(on, `id="passkey-login" class="primary wp-passkey-link" href="/login"`) {
		t.Errorf("passkey-on card must link to /login (the library login page):\n%s", on)
	}

	// The hand-rolled ceremony form is DEAD: its ids must never come
	// back on either shape (the library page owns the lp-* contract).
	for _, dead := range []string{`id="passkey-login-form"`, `id="passkey-email"`, `id="passkey-login-error"`} {
		if strings.Contains(off, dead) || strings.Contains(on, dead) {
			t.Errorf("dead hand-rolled id %s rendered — the library login page owns the ceremony UI", dead)
		}
	}

	extensionIds := []string{
		`id="login-form"`,
		`id="ext"`,
		`id="pass"`,
		`id="remember"`,
		`id="login-error"`,
	}
	for _, id := range extensionIds {
		if !strings.Contains(off, id) || !strings.Contains(on, id) {
			t.Errorf("extension login %s must survive in BOTH shapes (off=%v on=%v)", id,
				strings.Contains(off, id), strings.Contains(on, id))
		}
	}

	if !strings.Contains(on, "passkey-breakglass") {
		t.Error("passkey-on card lacks the break-glass disclosure around the extension login")
	}
	if strings.Contains(off, "passkey-breakglass") {
		t.Error("passkey-off card carries a break-glass disclosure with nothing to break glass for")
	}
}

// TestEnrollPageRendersStandalone pins the one-time enrollment page's
// wire ids (assets/enroll/enroll.js drives them) and the fact it loads
// NO island runtime — the module tag is the enroll asset, not main.js.
func TestEnrollPageRendersStandalone(t *testing.T) {
	page := renderComponent(t, EnrollPage(EnrollProps{Lang: LangEN, CSRFToken: "tok"}))
	for _, id := range []string{
		`id="enroll-view"`,
		`id="enroll-form"`,
		`id="enroll-token"`,
		`id="enroll-credential-name"`,
		`id="enroll-status"`,
		`id="enroll-error"`,
	} {
		if !strings.Contains(page, id) {
			t.Errorf("enroll page lacks %s:\n%s", id, page)
		}
	}
	if !strings.Contains(page, `src="/assets/enroll/enroll.js"`) {
		t.Error("enroll page must load its own module, not the island runtime")
	}
	if strings.Contains(page, "island/app/main.js") {
		t.Error("enroll page must not boot the island (no sip.js, no session surface)")
	}
	if !strings.Contains(page, `name="csrf-token" content="tok"`) {
		t.Error("enroll page must carry the CSRF meta its module POSTs with")
	}
}
