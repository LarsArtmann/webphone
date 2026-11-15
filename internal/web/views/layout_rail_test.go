package views

import (
	"strings"
	"testing"
)

// TestShellRendersTheSidebarRail pins the 2026-11-15 IA change at the
// SOURCE: the grouped rail (workspace / calls / people / system), one real
// route per destination and the rail's own collapse control. Navigation
// must stay an HTMX partial swap with a pushed URL — the SIP island lives
// in this document, so a plain document navigation would drop calls.
func TestShellRendersTheSidebarRail(t *testing.T) {
	html := renderComponent(t, ShellBody(ShellProps{
		Lang:      LangEN,
		SignedIn:  "1000",
		ActiveTab: TabMessages,
		Admin:     true,
	}))

	for _, want := range []string{
		`class="wp-rail"`,
		`id="wp-nav"`,
		`class="wp-nav-group"`,
		`>Workspace<`,
		`>Calls<`,
		`>People<`,
		`>System<`,
		`id="wp-rail-toggle"`,
		`aria-controls="wp-rail"`,
	} {
		if !strings.Contains(html, want) {
			t.Errorf("shell is missing %s (the rail ships groups + a collapse control)", want)
		}
	}

	for _, tab := range AllTabs {
		if !strings.Contains(html, `data-tab="`+string(tab)+`"`) {
			t.Errorf("rail lost the %s route", tab)
		}
	}
	for _, want := range []string{
		`hx-get="/partials/messages"`,
		`hx-push-url="/messages"`,
		`hx-target="#tab-content"`,
	} {
		if !strings.Contains(html, want) {
			t.Errorf("rail link lost %s", want)
		}
	}
}

// TestNavGroupsCoverEveryTab keeps the section lists honest: AllTabs is the
// DOM/route truth, navGroups is the visible IA, and a tab that exists in
// one but not the other is a destination nobody can reach (or a duplicate
// row in the rail).
func TestNavGroupsCoverEveryTab(t *testing.T) {
	seen := make(map[Tab]int, len(AllTabs))
	for _, group := range navGroups {
		if group.Key == "" || len(group.Tabs) == 0 {
			t.Errorf("empty nav group: %+v", group)
		}
		for _, tab := range group.Tabs {
			seen[tab]++
		}
	}
	for _, tab := range AllTabs {
		if seen[tab] != 1 {
			t.Errorf("tab %q appears %d times across navGroups, want exactly 1", tab, seen[tab])
		}
	}
	if len(seen) != len(AllTabs) {
		t.Errorf("navGroups carries %d tabs, AllTabs %d", len(seen), len(AllTabs))
	}
}

// TestSignedOutShellHasNoRailColumn: signed out is a door, not a floor
// plan — the body drops the rail column (the empty #wp-nav region stays as
// the post-login re-fetch target, but no links and no collapse control
// render).
func TestSignedOutShellHasNoRailColumn(t *testing.T) {
	html := renderComponent(t, ShellBody(ShellProps{Lang: LangEN}))

	if !strings.Contains(html, "wp-body-signed-out") {
		t.Error("signed-out shell must drop the rail column (wp-body-signed-out)")
	}
	if strings.Contains(html, "wp-nav-link") {
		t.Error("signed-out shell renders nav links")
	}
	if strings.Contains(html, `id="wp-rail-toggle"`) {
		t.Error("signed-out shell renders the rail toggle")
	}
}
