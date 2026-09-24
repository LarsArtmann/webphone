package views

import (
	"strings"
	"testing"

	"github.com/a-h/templ"
	"github.com/larsartmann/webphone/internal/domain"
)

// hostile doubles as content and escaping probe: every panel must render
// it inert, never raw — the error banner carries operator-facing failure
// text that may echo untrusted upstream input.
const hostile = `boom <script>`

// wantErrorNode is the byte-exact error banner every panel shares; the
// htmx responseHandling config selects `.wp-error` on 4xx/5xx swaps, so
// these bytes are a wire contract, not a style choice.
const wantErrorNode = `<p class="wp-error" role="alert">boom &lt;script&gt;</p>`

// wantIdentityNode is the byte-exact from-identity line the composer
// panels render when the deployment knows the extension's DID.
const wantIdentityNode = `<p class="wp-identity">sending as <strong>+491512345678</strong></p>`

func renderComponent(t *testing.T, component templ.Component) string {
	t.Helper()
	var buf strings.Builder
	if err := component.Render(t.Context(), &buf); err != nil {
		t.Fatalf("render: %v", err)
	}
	return buf.String()
}

func requireNode(t *testing.T, panel, rendered, want string, wantPresent bool) {
	t.Helper()
	if got := strings.Contains(rendered, want); got != wantPresent {
		t.Errorf("%s: error banner presence = %v, want %v; rendered:\n%s", panel, got, wantPresent, rendered)
	}
}

// TestPanelsRenderTheSharedErrorAndIdentityNodes pins the panel prologue
// contract across every panel that can carry it: the conditional
// wp-error banner and the conditional wp-identity line render the exact
// same node bytes everywhere, and a clean render emits neither. Any
// change to these nodes must land in every panel at once — which is why
// they route through shared components.
func TestPanelsRenderTheSharedErrorAndIdentityNodes(t *testing.T) {
	thread := domain.Thread{ID: domain.GenerateThreadID(), Remote: domain.MustParsePhone("+17287289311")}
	panels := map[string]struct {
		withFailure templ.Component
		clean       templ.Component
		hasIdentity bool
	}{
		"FaxPanel": {
			withFailure: FaxPanel(FaxPanelProps{Error: hostile, Identity: "+491512345678", Lang: LangEN}),
			clean:       FaxPanel(FaxPanelProps{Lang: LangEN}),
			hasIdentity: true,
		},
		"ThreadsPanel": {
			withFailure: ThreadsPanel(ThreadsPanelProps{Error: hostile, Identity: "+491512345678", Lang: LangEN}),
			clean:       ThreadsPanel(ThreadsPanelProps{Lang: LangEN}),
			hasIdentity: true,
		},
		"ThreadView": {
			withFailure: ThreadView(ThreadViewProps{Thread: thread, Error: hostile, Identity: "+491512345678", Lang: LangEN}),
			clean:       ThreadView(ThreadViewProps{Thread: thread, Lang: LangEN}),
			hasIdentity: true,
		},
		"HistoryPanel": {
			withFailure: HistoryPanel(HistoryPanelProps{Enabled: true, Error: hostile, Lang: LangEN}),
			clean:       HistoryPanel(HistoryPanelProps{Enabled: true, Lang: LangEN}),
		},
		"VoicemailPanel": {
			withFailure: VoicemailPanel(VoicemailPanelProps{Enabled: true, Error: hostile, Lang: LangEN}),
			clean:       VoicemailPanel(VoicemailPanelProps{Enabled: true, Lang: LangEN}),
		},
	}
	for name, panel := range panels {
		failed := renderComponent(t, panel.withFailure)
		requireNode(t, name, failed, wantErrorNode, true)
		if panel.hasIdentity {
			requireNode(t, name, failed, wantIdentityNode, true)
		}
		clean := renderComponent(t, panel.clean)
		requireNode(t, name, clean, `<p class="wp-error"`, false)
		requireNode(t, name, clean, `<p class="wp-identity"`, false)
	}
}

// TestPanelHeadPinsBothLanguages is the micro-test panelHead was born
// without (2026-09-23 sweep confession): the exact header bytes, both
// languages, so a future edit that reshapes the shared head fails here
// instead of silently drifting six panels.
func TestPanelHeadPinsBothLanguages(t *testing.T) {
	cases := map[Lang]string{
		LangEN: `<header class="wp-panel-head"><h2>Fax</h2><p class="wp-panel-sub">Send PDFs as faxes; received faxes land here.</p></header>`,
		LangDE: `<header class="wp-panel-head"><h2>Fax</h2><p class="wp-panel-sub">PDFs als Faxe senden; empfangene Faxe erscheinen hier.</p></header>`,
	}
	for lang, want := range cases {
		if got := renderComponent(t, panelHead(lang, "tab.fax", "fax.subtitle")); got != want {
			t.Errorf("panelHead(%s) =\n%s\nwant\n%s", lang, got, want)
		}
	}
}

// TestErrorBannerEscapesAndPanelErrorHidesEmpty pins the two usage
// shapes of the shared error node: the banner always renders (and
// escapes), the panel wrapper renders nothing when there is nothing to
// say — a panel without a failure must not emit an empty alert.
func TestErrorBannerEscapesAndPanelErrorHidesEmpty(t *testing.T) {
	if got := renderComponent(t, errorBanner(hostile)); got != wantErrorNode {
		t.Errorf("errorBanner(hostile) = %q, want %q", got, wantErrorNode)
	}
	if got := renderComponent(t, panelError(hostile)); got != wantErrorNode {
		t.Errorf("panelError(hostile) = %q, want %q", got, wantErrorNode)
	}
	if got := renderComponent(t, panelError("")); got != "" {
		t.Errorf("panelError(\"\") = %q, want no output", got)
	}
}

// TestIdentityLineHidesEmptyAndFollowsLanguage pins the from-identity
// line: present with the deployment's DID in both languages, absent
// when the DID is unknown.
func TestIdentityLineHidesEmptyAndFollowsLanguage(t *testing.T) {
	cases := map[Lang]string{
		LangEN: wantIdentityNode,
		LangDE: `<p class="wp-identity">Senden als <strong>+491512345678</strong></p>`,
	}
	for lang, want := range cases {
		if got := renderComponent(t, identityLine(lang, "+491512345678")); got != want {
			t.Errorf("identityLine(%s) = %q, want %q", lang, got, want)
		}
	}
	if got := renderComponent(t, identityLine(LangEN, "")); got != "" {
		t.Errorf("identityLine(en, \"\") = %q, want no output", got)
	}
}
