package views

import (
	"strings"
	"testing"
)

// NotFoundPanel is the served 404 body: a page change here breaks the
// stack's browser E2E and the smoke suite, so its shape is pinned. The
// path is attacker-supplied and must never render as markup.
func TestNotFoundPanelRendersTheContract(t *testing.T) {
	got := renderComponent(t, NotFoundPanel("/messages/42/gone", LangEN))

	for _, want := range []string{
		`class="wp-panel wp-notfound"`,
		"There is nothing at this address.",
		"404 — page not found",
		`<code>/messages/42/gone</code>`,
		`class="wp-notfound-back" href="/"`,
		"Back to Messages",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in:\n%s", want, got)
		}
	}
	if strings.Contains(got, "data-reload") {
		t.Error("the 404 must not offer a reload button: reloading a wrong address fails again")
	}
}

func TestNotFoundPanelEscapesThePath(t *testing.T) {
	got := renderComponent(t, NotFoundPanel(`/x"><script>alert(1)</script>`, LangEN))
	if strings.Contains(got, "<script>") {
		t.Errorf("the requested path must be escaped, got:\n%s", got)
	}
	if !strings.Contains(got, "&lt;script&gt;") {
		t.Errorf("the escaped path must still be shown, got:\n%s", got)
	}
}

func TestNotFoundPanelFollowsTheLanguage(t *testing.T) {
	got := renderComponent(t, NotFoundPanel("/nope", LangDE))
	for _, want := range []string{"404 — Seite nicht gefunden", "Zurück zu Nachrichten"} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in:\n%s", want, got)
		}
	}
}
