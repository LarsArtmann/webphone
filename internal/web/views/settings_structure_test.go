package views

import (
	"strings"
	"testing"
)

// Pins from the 2026-11-07 frontend-UX train's verification tail (P1-4). Each
// one guards a fix that a later refactor could undo without any gate noticing,
// because the regressions are copy- and structure-shaped, not compile-shaped.

// TestAdminCallLineCountNamesTranscriptLines pins the admin call row's line
// count copy. It used to read "12 lines", which is indistinguishable from a
// line count of the CALL; the row counts TRANSCRIPT segments. The leading space
// is part of the value (the template concatenates " · " + count + value), and
// the German compound is a single unbreakable token — there is no wrap point
// inside "Transkriptzeilen", so the narrow admin column wraps BETWEEN the
// segments of the row, never inside the word.
func TestAdminCallLineCountNamesTranscriptLines(t *testing.T) {
	en := T(LangEN, "admin.calls.lines")
	de := T(LangDE, "admin.calls.lines")

	if !strings.HasPrefix(en, " ") {
		t.Errorf("admin.calls.lines must keep its leading space (the row concatenates \" · \" + count + value): %q", en)
	}
	if !strings.Contains(en, "transcript") {
		t.Errorf("English copy must name the transcript, not a bare \"lines\": %q", en)
	}
	if !strings.HasPrefix(de, " ") {
		t.Errorf("German copy lost its leading space: %q", de)
	}
	if strings.Contains(strings.TrimSpace(de), " ") {
		t.Errorf("German copy must stay ONE unbreakable compound (a space would create a wrap point inside the count phrase): %q", de)
	}
}

// TestSettingsPanelReadsInThreeAudiences pins the F20 structure: the panel
// answers three different questions in a fixed order — what THIS session owns
// (preferences you can act on), what leaves with you (your data), and what the
// operator deployed (collapsed, because it is reference, not action).
//
// The order is the contract, so it is asserted as an ORDER, not a set: a
// refactor that moves the operator block above the personal one silently
// demotes the only sections a user can act on.
func TestSettingsPanelReadsInThreeAudiences(t *testing.T) {
	html := renderComponent(t, SettingsPanel(SettingsPanelProps{Lang: LangEN}))

	prefs := strings.Index(html, T(LangEN, "settings.group.prefs"))
	data := strings.Index(html, T(LangEN, "settings.group.data"))
	server := strings.Index(html, T(LangEN, "settings.group.server"))
	if prefs < 0 || data < 0 || server < 0 {
		t.Fatalf("all three audience headings must render (prefs=%d data=%d server=%d)", prefs, data, server)
	}
	if prefs >= data || data >= server {
		t.Errorf("audience order broken: prefs=%d < data=%d < server=%d expected", prefs, data, server)
	}

	// The operator block is reference material: it stays COLLAPSED inside a
	// <details>, so the panel does not open with facts a user cannot change.
	// The wrapper opens BEFORE the summary, so anchor on the details itself.
	details := strings.Index(html, `<details class="wp-settings-server"`)
	if details < 0 {
		t.Error("the server/operator group must render inside a <details> (collapsed reference)")
	} else if data >= details || details >= server {
		t.Errorf("the collapsed operator block must sit between the data group and the shortcuts (details=%d, data=%d, server=%d)", details, data, server)
	}
	if details >= 0 {
		block := html[details:]
		if end := strings.Index(block, "</details>"); end >= 0 {
			block = block[:end]
		}
		if !strings.Contains(block, "<summary") {
			t.Error("the collapsed operator group needs a <summary> — otherwise it has no visible label")
		}
	}

	// The personal group is a definition list; losing its class would drop the
	// scoped styling that separates "yours" from "the server's".
	if !strings.Contains(html, "wp-settings wp-settings-mine") {
		t.Error("the 'your data' group must carry the wp-settings-mine class")
	}
	// The shortcuts list is hand-written markup, so its row count is a real
	// "someone deleted a shortcut" detector.
	shortcuts := html[strings.Index(html, "wp-settings wp-shortcuts"):]
	if end := strings.Index(shortcuts, "</dl>"); end >= 0 {
		shortcuts = shortcuts[:end]
	}
	if got := strings.Count(shortcuts, "<dt>"); got != 10 {
		t.Errorf("shortcut rows = %d, want 10 (A H M P Esc [ ] Enter Shift+Enter Ctrl/Cmd+K ?)", got)
	}
}

// TestPaletteButtonIsADisclosure pins the coarse-pointer palette button's
// accessibility contract at the SOURCE. The button is a disclosure for the
// dialog it opens, so it must name that dialog and carry a state; shell.js
// keeps aria-expanded in step at runtime, which is a separate pin (the island
// specs cannot see served markup and this test cannot see the runtime).
func TestPaletteButtonIsADisclosure(t *testing.T) {
	html := renderComponent(t, ShellBody(ShellProps{Lang: LangEN}))

	button := ""
	for line := range strings.SplitSeq(html, "<button") {
		if strings.Contains(line, `id="wp-palette-open"`) {
			button = line
			break
		}
	}
	if button == "" {
		t.Fatal("the palette button must be served (it is the only touch door to the command palette)")
	}
	if !strings.Contains(button, `aria-controls="wp-palette"`) {
		t.Errorf("the button must name the dialog it opens (aria-controls=wp-palette): %s", button)
	}
	if !strings.Contains(button, `aria-expanded="false"`) {
		t.Errorf("the button must carry its closed state: %s", button)
	}
}
