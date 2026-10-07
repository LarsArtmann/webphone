package views

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/a-h/templ"
	"github.com/larsartmann/webphone/internal/domain"
	"github.com/larsartmann/webphone/internal/store"
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

// TestTranscriptRendersDaySeparatorsAndTheUnreadDivider pins the
// transcript's group structure: one day head per calendar day with the
// visible-group count, the unread divider ahead of the thread's trailing
// unread inbound bubbles on the page-0 open, the top-of-window fallback
// when more unread exist than the page holds, and the notifier-safe
// plain shape (no divider) from Transcript itself.
func TestTranscriptRendersDaySeparatorsAndTheUnreadDivider(t *testing.T) {
	now := time.Now()
	yesterday := now.Add(-24 * time.Hour)
	msgs := []domain.Message{
		{ID: domain.GenerateMessageID(), Direction: domain.DirectionOutbound, Body: "old out", CreatedAt: yesterday},
		{ID: domain.GenerateMessageID(), Direction: domain.DirectionInbound, Body: "old in", CreatedAt: yesterday},
		{ID: domain.GenerateMessageID(), Direction: domain.DirectionInbound, Body: "new in 1", CreatedAt: now},
		{ID: domain.GenerateMessageID(), Direction: domain.DirectionInbound, Body: "new in 2", CreatedAt: now},
		{ID: domain.GenerateMessageID(), Direction: domain.DirectionOutbound, Body: "new out", CreatedAt: now},
	}

	open := renderComponent(t, TranscriptAt(msgs, LangEN, 2, false))
	if got := strings.Count(open, `class="wp-day-head"`); got != 2 {
		t.Errorf("open transcript: day heads = %d, want 2; rendered:\n%s", got, open)
	}
	if !strings.Contains(open, "Today") || !strings.Contains(open, "Yesterday") {
		t.Errorf("open transcript: missing en day labels; rendered:\n%s", open)
	}
	for _, count := range []string{"\u00b7 2", "\u00b7 3"} {
		if !strings.Contains(open, count) {
			t.Errorf("open transcript: missing day count %q; rendered:\n%s", count, open)
		}
	}
	dividerAt := strings.Index(open, "wp-unread-divider")
	if dividerAt < 0 {
		t.Fatalf("open transcript: unread divider missing; rendered:\n%s", open)
	}
	if !strings.Contains(open, "\u2014 2 unread \u2014") {
		t.Errorf("open transcript: divider text wrong; rendered:\n%s", open)
	}
	if at := strings.Index(open, "new in 1"); at < dividerAt {
		t.Errorf("open transcript: divider must sit ahead of the first unread bubble (divider %d, bubble %d)", dividerAt, at)
	}
	if at := strings.Index(open, "old in"); at > dividerAt {
		t.Errorf("open transcript: divider must sit after the already-read bubbles (divider %d, bubble %d)", dividerAt, at)
	}

	// German keeps its labels and its ungelesen count.
	de := renderComponent(t, TranscriptAt(msgs, LangDE, 2, false))
	if !strings.Contains(de, "Heute") || !strings.Contains(de, "Gestern") || !strings.Contains(de, "\u2014 2 ungelesen \u2014") {
		t.Errorf("de transcript: labels or divider wrong; rendered:\n%s", de)
	}

	// More unread than the window holds: the divider lands at the very
	// top, ahead of the oldest bubble in the page.
	flood := renderComponent(t, TranscriptAt(msgs, LangEN, 5, false))
	floodAt := strings.Index(flood, "wp-unread-divider")
	if oldest := strings.Index(flood, "old out"); oldest < floodAt {
		t.Errorf("flooded transcript: divider must lead the window (divider %d, oldest %d)", floodAt, oldest)
	}

	// No unread (or the older-page render, which routes through
	// Transcript) never renders the divider.
	plain := renderComponent(t, Transcript(msgs, LangEN, false))
	if strings.Contains(plain, "wp-unread-divider") {
		t.Errorf("plain transcript: divider must be TranscriptAt-only; rendered:\n%s", plain)
	}
	if clean := renderComponent(t, TranscriptAt(msgs, LangEN, 0, false)); strings.Contains(clean, "wp-unread-divider") {
		t.Errorf("zero-unread transcript: divider rendered; rendered:\n%s", clean)
	}
}

// SignInHint (the pre-login welcome tab) is dismissible per browser:
// the compact line and the dismiss button must render in BOTH session
// languages — the copy stays server-owned while shell.js only toggles
// the collapse class (M18).
func TestSignInHintRendersDismissAndCompactLineBothLanguages(t *testing.T) {
	for _, lang := range []Lang{LangEN, LangDE} {
		rendered := renderComponent(t, SignInHint(lang))
		for _, want := range []string{
			`type="button"`,
			`class="wp-mini wp-welcome-dismiss"`,
			`class="wp-welcome-compact-line"`,
			T(lang, "welcome.compact"),
			T(lang, "welcome.dismiss"),
		} {
			if !strings.Contains(rendered, want) {
				t.Errorf("%s welcome: missing %q; rendered:\n%s", lang, want, rendered)
			}
		}
	}
}

// TestThreadRowFlagControlsBothLanguages pins the M22 row controls: the
// stable wrapper id (D14), action labels that track the row's state
// (the same expression feeds label and posted value), the muted row's
// suppressed badge with its state glyph, and the archived row's
// unarchive-only shape — in both UI languages.
func TestThreadRowFlagControlsBothLanguages(t *testing.T) {
	newSummary := func(flags domain.Thread) store.ThreadSummary {
		return store.ThreadSummary{Thread: flags}
	}
	plain := newSummary(domain.Thread{
		ID: domain.GenerateThreadID(), Remote: domain.MustParsePhone("+441632960961"), Unread: 2,
	})
	pinned := plain
	pinned.Thread.Pinned = true
	muted := plain
	muted.Thread.Muted = true
	archived := plain
	archived.Thread.Archived = true

	en := renderComponent(t, ThreadRow(plain, nil, LangEN))
	if !strings.Contains(en, `id="thread-`+plain.Thread.ID.String()+`"`) {
		t.Errorf("row wrapper id must be the stable thread-<id>: %s", en)
	}
	if !strings.Contains(en, `aria-label="Pin"`) || !strings.Contains(en, plain.Thread.ID.String()+`/pin?on=1`) {
		t.Errorf("plain row must offer Pin posting on=1: %s", en)
	}
	if !strings.Contains(en, `aria-label="Mute"`) || !strings.Contains(en, plain.Thread.ID.String()+`/mute?on=1`) {
		t.Errorf("plain row must offer Mute posting on=1: %s", en)
	}
	if !strings.Contains(en, `aria-label="Archive"`) || !strings.Contains(en, plain.Thread.ID.String()+`/archive?on=1`) {
		t.Errorf("plain row must offer Archive posting on=1: %s", en)
	}
	if !strings.Contains(en, `class="wp-nav-badge"`) {
		t.Errorf("unmuted unread row must show the badge: %s", en)
	}

	pinnedEN := renderComponent(t, ThreadRow(pinned, nil, LangEN))
	if !strings.Contains(pinnedEN, `aria-label="Unpin"`) || !strings.Contains(pinnedEN, pinned.Thread.ID.String()+`/pin?on=0`) {
		t.Errorf("pinned row must offer Unpin posting on=0: %s", pinnedEN)
	}

	mutedEN := renderComponent(t, ThreadRow(muted, nil, LangEN))
	if strings.Contains(mutedEN, `class="wp-nav-badge"`) {
		t.Errorf("muted row must suppress the badge (D8): %s", mutedEN)
	}
	if !strings.Contains(mutedEN, `aria-label="Unmute"`) || !strings.Contains(mutedEN, ">🔇</button>") {
		t.Errorf("muted row must offer Unmute with the speaker-off glyph: %s", mutedEN)
	}

	archivedEN := renderComponent(t, ThreadRow(archived, nil, LangEN))
	if !strings.Contains(archivedEN, `aria-label="Unarchive"`) || !strings.Contains(archivedEN, archived.Thread.ID.String()+`/archive?on=0`) {
		t.Errorf("archived row must offer Unarchive posting on=0: %s", archivedEN)
	}
	if strings.Contains(archivedEN, `/pin?`) || strings.Contains(archivedEN, `/mute?`) {
		t.Errorf("archived row keeps no pins or mutes: %s", archivedEN)
	}

	de := renderComponent(t, ThreadRow(plain, nil, LangDE))
	for _, want := range []string{`aria-label="Anheften"`, `aria-label="Stummschalten"`, `aria-label="Archivieren"`} {
		if !strings.Contains(de, want) {
			t.Errorf("German row labels missing %s: %s", want, de)
		}
	}
}

// TestArchivedToggleAndSearchHiding pins the D6 panel shapes: the
// active view links to the archived list only when something is filed,
// and the archived view swaps the link for the way back and drops the
// search box (search walks the active list only).
func TestArchivedToggleAndSearchHiding(t *testing.T) {
	active := renderComponent(t, ThreadsPanel(ThreadsPanelProps{ArchivedCount: 2, Lang: LangEN}))
	if !strings.Contains(active, ">Archived (2)</a>") || !strings.Contains(active, `hx-get="/partials/messages?archived=1"`) {
		t.Errorf("active view must link the archived list: %s", active)
	}
	if !strings.Contains(active, `class="wp-thread-search"`) {
		t.Errorf("active view keeps the search box: %s", active)
	}

	empty := renderComponent(t, ThreadsPanel(ThreadsPanelProps{Lang: LangEN}))
	if strings.Contains(empty, "wp-archived-toggle") {
		t.Errorf("no archived link when nothing is filed: %s", empty)
	}

	archived := renderComponent(t, ThreadsPanel(ThreadsPanelProps{Archived: true, Lang: LangEN}))
	if !strings.Contains(archived, ">← Back to messages</a>") || !strings.Contains(archived, `hx-push-url="/messages"`) {
		t.Errorf("archived view must offer the way back: %s", archived)
	}
	if strings.Contains(archived, `class="wp-thread-search"`) {
		t.Errorf("archived view must hide the search box: %s", archived)
	}
	archivedDE := renderComponent(t, ThreadsPanel(ThreadsPanelProps{Archived: true, Lang: LangDE}))
	if !strings.Contains(archivedDE, "Zurück zu den Nachrichten") {
		t.Errorf("German back link missing: %s", archivedDE)
	}
}

func testSnippets(count, quick int) []domain.Snippet {
	snippets := make([]domain.Snippet, 0, count)
	for i := range count {
		snippets = append(snippets, domain.Snippet{
			ID: domain.GenerateSnippetID(), Owner: domain.MustParseExtension("1001"),
			Body: fmt.Sprintf("snippet %02d", i), Quick: i < quick, CreatedAt: time.Now(),
		})
	}
	return snippets
}

// TestSnippetChipsAndPicker pin the M21 composer lane: quick snippets
// (capped at five) render as chips, every snippet sits behind the
// picker disclosure, and an empty snippet list renders no lane at all.
func TestSnippetChipsAndPicker(t *testing.T) {
	view := renderComponent(t, ThreadView(ThreadViewProps{
		Thread: domain.Thread{ID: domain.GenerateThreadID(), Remote: domain.MustParsePhone("+441632960961")},
		Snippets: func() []domain.Snippet {
			snippets := testSnippets(7, 6)
			snippets[6] = domain.Snippet{
				ID: domain.GenerateSnippetID(), Owner: domain.MustParseExtension("1001"),
				Body:      "a deliberately long snippet body that must be shortened for the chip lane",
				CreatedAt: time.Now(),
			}
			return snippets
		}(),
		Lang: LangEN,
	}))
	if got := strings.Count(view, `class="wp-chip"`); got != 5 {
		t.Errorf("chip lane must cap at five quick snippets, got %d", got)
	}
	if !strings.Contains(view, `data-snippet="snippet 00"`) || strings.Count(view, `data-snippet="snippet 05"`) != 1 {
		t.Errorf("chips carry the first five QUICK bodies; snippet 05 lives only in the picker: %s", view)
	}
	if !strings.Contains(view, "Insert snippet") || !strings.Contains(view, `class="wp-snippet-picker"`) {
		t.Errorf("picker disclosure missing: %s", view)
	}
	if !strings.Contains(view, `data-snippet="a deliberately long snippet body that must be shortened for the chip lane"`) {
		t.Errorf("picker holds every snippet, long ones included: %s", view)
	}
	if !strings.Contains(view, "…") {
		t.Errorf("chip labels truncate with an ellipsis: %s", view)
	}

	bare := renderComponent(t, ThreadView(ThreadViewProps{
		Thread: domain.Thread{ID: domain.GenerateThreadID(), Remote: domain.MustParsePhone("+441632960961")},
		Lang:   LangEN,
	}))
	if strings.Contains(bare, "wp-snippet-bar") {
		t.Errorf("no snippet lane without snippets: %s", bare)
	}
}

// TestSettingsSnippetsSection pins the Settings management surface: the
// list rows with quick markers and delete buttons, the add form with
// the quick checkbox — and the empty state when nothing exists.
func TestSettingsSnippetsSection(t *testing.T) {
	with := renderComponent(t, SettingsPanel(SettingsPanelProps{
		SharedContacts: 1, Snippets: testSnippets(2, 1), Lang: LangEN,
	}))
	if !strings.Contains(with, "Reply snippets") || !strings.Contains(with, "Add snippet") {
		t.Errorf("section head/add form missing: %s", with)
	}
	if !strings.Contains(with, `data-quick`) && !strings.Contains(with, `>⚡<`) {
		t.Errorf("quick marker missing on the first snippet: %s", with)
	}
	if !strings.Contains(with, `hx-post="/snippets/delete?id=`) {
		t.Errorf("delete buttons missing: %s", with)
	}
	if !strings.Contains(with, `name="quick"`) || !strings.Contains(with, `hx-post="/snippets/save"`) {
		t.Errorf("add form missing its fields: %s", with)
	}

	empty := renderComponent(t, SettingsPanel(SettingsPanelProps{Lang: LangEN}))
	if !strings.Contains(empty, "No snippets yet — add the first one below.") {
		t.Errorf("empty state missing: %s", empty)
	}
}

// TestImageAttachmentCarriesLightboxAttr: the inline image attachment
// anchor carries data-lightbox (shell.js opens the dialog on it); a
// non-image attachment never does.
func TestImageAttachmentCarriesLightboxAttr(t *testing.T) {
	image := domain.Message{
		ID: domain.GenerateMessageID(), Direction: domain.DirectionInbound,
		Attachments: []domain.Attachment{{
			ID: domain.GenerateAttachmentID(), Name: "pic.png", MimeType: "image/png", SizeBytes: 4096,
		}},
	}
	file := domain.Message{
		ID: domain.GenerateMessageID(), Direction: domain.DirectionInbound,
		Attachments: []domain.Attachment{{
			ID: domain.GenerateAttachmentID(), Name: "doc.pdf", MimeType: "application/pdf", SizeBytes: 4096,
		}},
	}
	imageHTML := renderComponent(t, Bubble(image, LangEN, false))
	if !strings.Contains(imageHTML, "wp-attachment-image") || !strings.Contains(imageHTML, `data-lightbox="pic.png"`) {
		t.Errorf("image attachment must carry data-lightbox: %s", imageHTML)
	}
	fileHTML := renderComponent(t, Bubble(file, LangEN, false))
	if strings.Contains(fileHTML, "data-lightbox") {
		t.Errorf("non-image attachment must not carry data-lightbox: %s", fileHTML)
	}
}
