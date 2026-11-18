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

// TestSettingsAccountLeadsWithTheActiveLine: after a line switch the
// session's NEW line leads the owned-lines list — mapping order alone
// would bury the line the operator is acting on (owner ruling
// 2026-11-18: active first, mapping order after).
func TestSettingsAccountLeadsWithTheActiveLine(t *testing.T) {
	_, c := newSwitchServer(t)
	if resp, body := c.do(http.MethodPost, switchPath, []byte(`{"extension":"1001"}`), "application/json"); resp.StatusCode != http.StatusOK {
		t.Fatalf("switch: %d %s", resp.StatusCode, body)
	}
	_, body := c.do(http.MethodGet, "/partials/settings", nil, "")
	page := string(body)
	active := strings.Index(page, "+17287289312")
	idle := strings.Index(page, "+17287289311")
	if active < 0 || idle < 0 || active > idle {
		t.Errorf("the active line (+…312, switched to 1001) must lead the account lines:\n%s", page)
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
	// The fallback rule (authmap.Line.Display): a line with no declared
	// DID names itself ONCE — the bare extension, never "1000 · 1000".
	if strings.Contains(page, " · 1000") {
		t.Errorf("a DID-less line must not repeat its extension as its own secondary label:\n%s", page)
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
// bare extension list alone; a DID-less line names itself once (the
// authmap.Line.Display fallback), and the numbers inventory renders
// the SAME chip shape (symmetric line naming across both sections).
func TestAdminUserRowsLabelOwnedLines(t *testing.T) {
	server, sessions := newAdminServer(t, "")
	admin := mintedClient(t, server, sessions, "1001", "boss@example.com")
	_, body := admin.do(http.MethodGet, "/partials/admin", nil, "")
	page := string(body)
	for _, want := range []string{`class="wp-admin-line"`, "+49301111", " · 1001", "+49302222", " · 1002", "1003"} {
		if !strings.Contains(page, want) {
			t.Errorf("admin rows must label lines with their numbers (missing %q):\n%s", want, page)
		}
	}
	if strings.Contains(page, "1003 · 1003") {
		t.Errorf("a DID-less line must not repeat its extension as its own secondary label:\n%s", page)
	}
	// The numbers inventory carries the same labeled chips plus the
	// no-DID marker for intern's line.
	numbers := page[strings.Index(page, "wp-admin-numbers"):]
	if !strings.Contains(numbers, "1003") || !strings.Contains(numbers, "no DID configured") {
		t.Errorf("the numbers inventory must chip 1003 and mark its missing DID:\n%s", numbers)
	}
}
