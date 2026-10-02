package assets

import (
	"io/fs"
	"regexp"
	"strings"
	"testing"
)

// tokenBlock extracts the {…} body following the first occurrence of
// marker (brace-matched, so nested blocks would survive too).
func tokenBlock(t *testing.T, css, marker string) string {
	t.Helper()
	i := strings.Index(css, marker)
	if i < 0 {
		t.Fatalf("marker %q not found in stylesheet", marker)
	}
	open := strings.Index(css[i:], "{")
	if open < 0 {
		t.Fatalf("marker %q has no opening brace", marker)
	}
	depth := 0
	for j := i + open; j < len(css); j++ {
		switch css[j] {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return css[i+open+1 : j]
			}
		}
	}
	t.Fatalf("marker %q block is unterminated", marker)
	return ""
}

var tokenRe = regexp.MustCompile(`--[a-zA-Z0-9-]+\s*:\s*[^;]+;`)

func parseTokens(body string) map[string]string {
	out := map[string]string{}
	for _, m := range tokenRe.FindAllString(body, -1) {
		parts := strings.SplitN(m, ":", 2)
		name := strings.TrimSpace(parts[0])
		out[name] = strings.TrimSpace(strings.TrimSuffix(parts[1], ";"))
	}
	return out
}

// themeDependent must be defined by BOTH stylesheets in BOTH themes:
// a missing definition would break the other stylesheet's theme
// propagation (the manual data-theme override reaches island surfaces
// only because app.css's override blocks carry the full shared set).
// --danger-soft is app-only today (the island consumes no soft-danger
// surface) — shared-name equality still covers it the day it crosses.
var themeDependent = []string{
	"--bg", "--surface", "--surface-2", "--surface-3",
	"--border", "--border-strong", "--text", "--muted",
	"--accent", "--accent-strong", "--accent-soft",
	"--danger", "--warn", "--ok",
	"--shadow",
}

// themeStable must exist in both dark :root blocks (values identical in
// every theme, so the light blocks inherit them by design).
var themeStable = []string{
	"--on-accent", "--scrim", "--radius-sm", "--radius", "--radius-lg",
}

func assertSharedEqual(t *testing.T, label string, a, b map[string]string) {
	t.Helper()
	for name, va := range a {
		if vb, ok := b[name]; ok && va != vb {
			t.Errorf("%s: token %s diverges between the mirrored stylesheets: %q vs %q", label, name, va, vb)
		}
	}
}

func assertPresent(t *testing.T, label string, tokens map[string]string, names []string) {
	t.Helper()
	for _, name := range names {
		if _, ok := tokens[name]; !ok {
			t.Errorf("%s: token %s is missing (the mirror rule or the theme propagation depends on it)", label, name)
		}
	}
}

// TestCSSTokenBlocksAreMirrored pins the UI token system: the :root and
// light-theme token blocks of app.css and island/style.css stay mirrored
// (shared names carry identical values, and the theme-dependent core is
// defined in every theme block of both files). The island has no
// [data-theme] selectors of its own — manual theming reaches island
// surfaces purely through app.css's higher-specificity override blocks,
// which only works while the shared set stays aligned.
func TestCSSTokenBlocksAreMirrored(t *testing.T) {
	appCSS, err := fs.ReadFile(embedded, "app.css")
	if err != nil {
		t.Fatalf("read app.css: %v", err)
	}
	islandCSS, err := fs.ReadFile(embedded, "island/style.css")
	if err != nil {
		t.Fatalf("read island/style.css: %v", err)
	}
	app := string(appCSS)
	island := string(islandCSS)

	appDark := parseTokens(tokenBlock(t, app, ":root,\n:root[data-theme=\"dark\"]"))
	appLightAuto := parseTokens(tokenBlock(t, app, ":root:not([data-theme])"))
	appLightManual := parseTokens(tokenBlock(t, app, ":root[data-theme=\"light\"]"))

	light := strings.Index(island, "@media (prefers-color-scheme: light)")
	if light < 0 {
		t.Fatal("island stylesheet lost its light media block")
	}
	islandDark := parseTokens(tokenBlock(t, island[:light], ":root"))
	islandLight := parseTokens(tokenBlock(t, island[light:], ":root"))

	// app.css's two light blocks are one fact in two places: identical.
	if len(appLightAuto) != len(appLightManual) {
		t.Errorf("app.css light blocks diverge in size: auto %d vs manual %d tokens", len(appLightAuto), len(appLightManual))
	}
	assertSharedEqual(t, "app.css light-auto vs light-manual", appLightAuto, appLightManual)

	assertSharedEqual(t, "dark :root", appDark, islandDark)
	assertSharedEqual(t, "light theme", appLightManual, islandLight)

	assertPresent(t, "app.css dark", appDark, append(append([]string{}, themeDependent...), themeStable...))
	assertPresent(t, "island dark", islandDark, append(append([]string{}, themeDependent...), themeStable...))
	assertPresent(t, "app.css light-auto", appLightAuto, themeDependent)
	assertPresent(t, "app.css light-manual", appLightManual, themeDependent)
	assertPresent(t, "island light", islandLight, themeDependent)
}
