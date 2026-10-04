package views

import (
	"strings"
	"testing"
)

// TestLoginCardAdaptsToPasskeyMode pins the login card's two shapes:
// off (the default) renders EXACTLY the extension login — no passkey
// ids anywhere — while on promotes the passkey form to the front door
// and demotes the extension login into the break-glass disclosure
// without losing a single one of its DOM-contract ids.
func TestLoginCardAdaptsToPasskeyMode(t *testing.T) {
	off := renderComponent(t, PhoneIsland(PhoneIslandProps{}))
	on := renderComponent(t, PhoneIsland(PhoneIslandProps{Passkey: true}))

	passkeyIds := []string{
		`id="passkey-login-form"`,
		`id="passkey-email"`,
		`id="passkey-login-error"`,
	}
	for _, id := range passkeyIds {
		if strings.Contains(off, id) {
			t.Errorf("passkey-off card renders %s — a disabled deployment must stay byte-shaped like the pre-passkey card", id)
		}
		if !strings.Contains(on, id) {
			t.Errorf("passkey-on card lacks %s — the front door is incomplete:\n%s", id, on)
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
