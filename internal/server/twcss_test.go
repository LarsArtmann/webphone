package server

import (
	"context"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/a-h/templ"
	"github.com/larsartmann/templ-components/display"
	"github.com/larsartmann/templ-components/forms"
	"github.com/larsartmann/templ-components/icons"
	"github.com/larsartmann/templ-components/layout"
	"github.com/larsartmann/templ-components/utils"
)

// TestTwCssCoversAdoptedComponentClasses renders every adopted
// templ-components surface and asserts each emitted class token has a
// matching selector in the built tw.css artifact. This is the drift gate
// for the /assets/tw.css adoption layer: a dependency bump that starts
// emitting a new Tailwind class (or a new component adoption) fails here
// instead of shipping unstyled markup. The artifact itself is rebuilt by
// scripts/build-tw-css.sh from tw.css.input.
func TestTwCssCoversAdoptedComponentClasses(t *testing.T) {
	css, err := os.ReadFile("../web/assets/tw.css")
	if err != nil {
		t.Fatalf("read tw.css artifact: %v", err)
	}
	artifact := string(css)

	components := map[string]templ.Component{
		"button primary md": display.Button(display.ButtonProps{
			Text: "Send", Type: display.ButtonHTMLSubmit,
			Variant: display.ButtonPrimary, Size: display.ButtonSizeMD,
		}),
		"button primary sm": display.Button(display.ButtonProps{
			Text: "Send", Variant: display.ButtonPrimary, Size: display.ButtonSizeSM,
		}),
		"button secondary md": display.Button(display.ButtonProps{
			Text: "Import", Variant: display.ButtonSecondary, Size: display.ButtonSizeMD,
		}),
		"button with icon": display.Button(display.ButtonProps{
			Text: "Fax", Variant: display.ButtonPrimary,
			Icon: icons.Icon(icons.Fax, "h-4 w-4"),
		}),
		"input bare": forms.Input(forms.InputProps{
			BaseProps: utils.BaseProps{ID: "x-input", AriaLabel: "X"},
			Name:      "x",
		}),
		"input labeled error": forms.Input(forms.InputProps{
			BaseProps: utils.BaseProps{ID: "y-input"},
			Name:      "y", Label: "Y", Required: true,
			Error: "bad", HelpText: "hint",
		}),
		"textarea bare": forms.Textarea(forms.TextareaProps{
			BaseProps: utils.BaseProps{ID: "z-area", AriaLabel: "Z"},
			Name:      "z", Rows: 2,
		}),
		"textarea labeled": forms.Textarea(forms.TextareaProps{
			BaseProps: utils.BaseProps{ID: "w-area"},
			Name:      "w", Label: "W",
		}),
		"empty state with action": display.EmptyState(display.EmptyStateProps{
			Title:      "Nothing",
			Icon:       icons.Inbox,
			ActionText: "Do",
			ActionHref: "/do",
		}),
		"layout base": layout.Base(layout.PageProps{
			Title: "t", Locale: "en", NoThemeScript: true, HTMXNone: true,
		}),
	}

	classAttr := regexp.MustCompile(`class="([^"]*)"`)
	for name, comp := range components {
		var sb strings.Builder
		if err := comp.Render(context.Background(), &sb); err != nil {
			t.Fatalf("%s: render: %v", name, err)
		}
		for _, attr := range classAttr.FindAllStringSubmatch(sb.String(), -1) {
			for _, tok := range strings.Fields(attr[1]) {
				if !classesInArtifact(tok, artifact) {
					t.Errorf("%s: class %q has no selector in tw.css (rebuild via scripts/build-tw-css.sh)", name, tok)
				}
			}
		}
	}
}

// classesInArtifact reports whether the Tailwind class token has a
// corresponding escaped selector in the built CSS. Non-Tailwind classes
// (htmx contract classes, app-owned classes) return true unconditionally:
// this gate owns the Tailwind adoption layer, nothing else.
func classesInArtifact(token, css string) bool {
	switch token {
	case "htmx-indicator", "tc-auto-grow":
		return true
	}
	selector := "." + cssEscapeSelector(token)
	return strings.Contains(css, selector)
}

// cssEscapeSelector escapes the characters Tailwind escapes in generated
// class selectors: ":" becomes "\:", "." becomes "\.", "/" becomes "\/",
// "%" becomes "\%".
func cssEscapeSelector(token string) string {
	replacer := strings.NewReplacer(
		":", `\:`,
		".", `\.`,
		"/", `\/`,
		"%", `\%`,
	)
	return replacer.Replace(token)
}
