package views

import (
	"strings"
	"testing"
)

// TestLoginPageExtraAdaptsToPasskeyMode pins the /login page's consumer
// section (views.LoginPageExtra, hosted inside the cqrs-htmx loginpage
// card) in its shapes: passkey-off renders the extension login as the
// PRIMARY method (bare form), passkey-on demotes it into the break-glass
// disclosure, and the SSO door renders above either shape without losing
// a single DOM-contract id. The island card no longer renders any login
// form — /login is the one front door.
func TestLoginPageExtraAdaptsToPasskeyMode(t *testing.T) {
	primary := renderComponent(t, LoginPageExtra(true, false))
	breakglass := renderComponent(t, LoginPageExtra(false, false))
	withSSO := renderComponent(t, LoginPageExtra(false, true))

	extensionIds := []string{
		`id="login-form"`,
		`id="ext"`,
		`id="pass"`,
		`id="remember"`,
		`id="login-error"`,
	}
	for _, id := range extensionIds {
		if !strings.Contains(primary, id) || !strings.Contains(breakglass, id) {
			t.Errorf("extension login %s must survive in BOTH shapes (primary=%v breakglass=%v)", id,
				strings.Contains(primary, id), strings.Contains(breakglass, id))
		}
	}

	if !strings.Contains(breakglass, "Use extension and password instead") {
		t.Error("passkey-on page lacks the break-glass disclosure around the extension login")
	}
	if strings.Contains(primary, "Use extension and password instead") {
		t.Error("passkey-off page hides its PRIMARY sign-in method behind a disclosure")
	}

	if !strings.Contains(withSSO, `id="sso-login"`) {
		t.Errorf("SSO variant lacks the Pocket ID door:\n%s", withSSO)
	}
	if !strings.Contains(withSSO, `href="/api/auth/oauth2/begin"`) {
		t.Errorf("SSO door must navigate into the OIDC begin route:\n%s", withSSO)
	}
	if !strings.Contains(primary, `src="/assets/login-page.js"`) {
		t.Error("the section must load its standalone module (same-origin 'self' script, no nonce needed)")
	}

	// The hand-rolled ceremony form is DEAD: its ids must never come
	// back on any shape (the library page owns the lp-* contract).
	for _, dead := range []string{`id="passkey-login-form"`, `id="passkey-email"`, `id="passkey-login-error"`} {
		if strings.Contains(primary, dead) || strings.Contains(withSSO, dead) {
			t.Errorf("dead hand-rolled id %s rendered — the library login page owns the ceremony UI", dead)
		}
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
