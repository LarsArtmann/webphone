// Account-surface pins for the 2026-11-18 line-ownership model: the
// Settings "My account" group (identity + owned lines, active line
// marked — mapped and extension logins alike) and the Admin users
// inventory rendering DID-labeled line chips instead of bare codes.
package server

import (
	"net/http"
	"strings"
	"testing"
)

// TestSettingsAccountListsOwnedLinesWithTheActiveMark: a mapped login
// sees WHO it is and EVERY line it owns, each labeled with its
// presented number, the session's own line marked active.
func TestSettingsAccountListsOwnedLinesWithTheActiveMark(t *testing.T) {
	_, c := newSwitchServer(t)
	resp, body := c.do(http.MethodGet, "/partials/settings", nil, "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("settings: %d", resp.StatusCode)
	}
	page := string(body)
	for _, want := range []string{
		`id="wp-settings-account"`,
		pkEmail,
		"Lars",
		"+17287289311", "1000",
		"+17287289312", "1001",
	} {
		if !strings.Contains(page, want) {
			t.Errorf("account section must carry %q:\n%s", want, page)
		}
	}
	if !strings.Contains(page, "active") {
		t.Errorf("account section must mark the active line:\n%s", page)
	}
}

// TestSettingsAccountNamesTheLineForExtensionLogin: the classic
// extension login owns no mapped identity, but its session still has a
// line — the section renders the active line WITHOUT claiming an
// email identity it does not have.
func TestSettingsAccountNamesTheLineForExtensionLogin(t *testing.T) {
	server := newTestServer(t)
	c := signIn(t, server)
	_, body := c.do(http.MethodGet, "/partials/settings", nil, "")
	page := string(body)
	if !strings.Contains(page, `id="wp-settings-account"`) {
		t.Fatalf("account section renders for every session:\n%s", page)
	}
	if !strings.Contains(page, "active") {
		t.Errorf("the session's own line is marked active:\n%s", page)
	}
	if strings.Contains(page, "Signed-in identity: ") {
		// The identity line renders, but must not claim an email for an
		// extension login — an identity-less session shows no email.
		if idx := strings.Index(page, "Signed-in identity"); idx >= 0 {
			tail := page[idx:min(idx+120, len(page))]
			if strings.Contains(tail, "@") {
				t.Errorf("extension login must not claim a mapped email:\n%s", tail)
			}
		}
	}
}

// TestAdminUserRowsLabelOwnedLines: the users inventory reads as phone
// numbers — each account's rows carry "DID · extension" chips, never a
// bare extension list alone.
func TestAdminUserRowsLabelOwnedLines(t *testing.T) {
	server, sessions := newAdminServer(t, "")
	admin := mintedClient(t, server, sessions, "1001", "boss@example.com")
	_, body := admin.do(http.MethodGet, "/partials/admin", nil, "")
	page := string(body)
	for _, want := range []string{`class="wp-admin-line"`, "+49301111", " · 1001", "+49302222", " · 1002"} {
		if !strings.Contains(page, want) {
			t.Errorf("admin user rows must label lines with their numbers (missing %q):\n%s", want, page)
		}
	}
}
