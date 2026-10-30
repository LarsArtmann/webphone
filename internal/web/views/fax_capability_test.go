package views

import (
	"strings"
	"testing"

	"github.com/larsartmann/webphone/internal/domain"
)

// TestFaxPanelDeclaresTheDeploymentCapability pins the honesty contract of
// the Fax tab: the panel offers a send form ONLY when the deployment
// declares outbound fax, states the missing direction up front (never as a
// failed send), and keeps the both-directions rendering byte-identical to
// the pre-capability markup.
func TestFaxPanelDeclaresTheDeploymentCapability(t *testing.T) {
	both := FaxCapability{Send: true, Receive: true}
	sendOnly := FaxCapability{Send: true}
	receiveOnly := FaxCapability{Receive: true}
	none := FaxCapability{}

	cases := []struct {
		name       string
		cap        FaxCapability
		wantForm   bool
		wantNotice string // "" = no notice at all
		wantSub    string
		wantEmpty  string
	}{
		{
			name:      "both directions keep the plain panel",
			cap:       both,
			wantForm:  true,
			wantSub:   "fax.subtitle",
			wantEmpty: "fax.empty",
		},
		{
			name:       "no send form without an outbound provider",
			cap:        receiveOnly,
			wantForm:   false,
			wantNotice: "fax.sendOff",
			wantSub:    "fax.subtitle.nosend",
			wantEmpty:  "fax.empty.nosend",
		},
		{
			name:       "no receive says the list cannot grow inwards",
			cap:        sendOnly,
			wantForm:   true,
			wantNotice: "fax.receiveOff",
			wantSub:    "fax.subtitle.noreceive",
			wantEmpty:  "fax.empty.noreceive",
		},
		{
			name:       "neither direction is one combined notice",
			cap:        none,
			wantForm:   false,
			wantNotice: "fax.off",
			wantSub:    "fax.subtitle.off",
			wantEmpty:  "fax.empty.off",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			for _, lang := range []Lang{LangEN, LangDE} {
				panel := renderComponent(t, FaxPanel(FaxPanelProps{Capability: tc.cap, Lang: lang}))

				if got := strings.Contains(panel, `class="wp-compose wp-compose-fax"`); got != tc.wantForm {
					t.Errorf("%s/%s: send form present = %v, want %v", tc.name, lang, got, tc.wantForm)
				}
				if tc.wantNotice == "" {
					if strings.Contains(panel, "wp-fax-off") {
						t.Errorf("%s/%s: unexpected capability notice in a complete deployment", tc.name, lang)
					}
				} else {
					if got := strings.Count(panel, "wp-fax-off"); got != 1 {
						t.Errorf("%s/%s: capability notices = %d, want exactly 1", tc.name, lang, got)
					}
					if !strings.Contains(panel, T(lang, tc.wantNotice)) {
						t.Errorf("%s/%s: notice copy %q missing from panel", tc.name, lang, T(lang, tc.wantNotice))
					}
				}
				if !strings.Contains(panel, T(lang, tc.wantSub)) {
					t.Errorf("%s/%s: subtitle %q missing from panel", tc.name, lang, T(lang, tc.wantSub))
				}
				if !strings.Contains(panel, T(lang, tc.wantEmpty)) {
					t.Errorf("%s/%s: empty state %q missing from panel", tc.name, lang, T(lang, tc.wantEmpty))
				}
			}
		})
	}
}

// TestFaxPanelDefaultMarkupIsUnchanged guards the zero-config promise: a
// deployment that declares both directions renders exactly the markup the
// panel produced before capabilities existed (no notice, the historical
// subtitle and empty-state copy).
func TestFaxPanelDefaultMarkupIsUnchanged(t *testing.T) {
	panel := renderComponent(t, FaxPanel(FaxPanelProps{Capability: FaxCapability{Send: true, Receive: true}, Lang: LangEN}))

	for _, want := range []string{
		`<p class="wp-panel-sub">Send PDFs as faxes; received faxes land here.</p>`,
		`No faxes yet — send a PDF above.`,
		`hx-post="/fax/send"`,
	} {
		if !strings.Contains(panel, want) {
			t.Errorf("default panel lost %q; rendered:\n%s", want, panel)
		}
	}
	if strings.Contains(panel, "wp-fax-off") {
		t.Error("default panel must not render a capability notice")
	}
}

// TestFaxListEmptyStateFollowsTheCapability pins the SSE fragment: the
// pushed list is the same component the full render uses, so it must never
// tell a user to "send a PDF above" on a deployment without a send form.
func TestFaxListEmptyStateFollowsTheCapability(t *testing.T) {
	cases := []struct {
		cap  FaxCapability
		want string
	}{
		{FaxCapability{Send: true, Receive: true}, "fax.empty"},
		{FaxCapability{Receive: true}, "fax.empty.nosend"},
		{FaxCapability{Send: true}, "fax.empty.noreceive"},
		{FaxCapability{}, "fax.empty.off"},
	}
	for _, tc := range cases {
		got := renderComponent(t, FaxList(nil, CRMContacts{}, tc.cap, LangEN))
		if !strings.Contains(got, T(LangEN, tc.want)) {
			t.Errorf("cap %+v: empty state %q missing; rendered:\n%s", tc.cap, T(LangEN, tc.want), got)
		}
	}
}

// TestFaxRowRendersWithoutCapability pins that the capability gates
// AFFORDANCES only: stored faxes (history from a previous configuration)
// still render, with their document links, on a deployment that can no
// longer send or receive.
func TestFaxRowRendersWithoutCapability(t *testing.T) {
	job := domain.FaxJob{
		Direction:    domain.FaxOutbound,
		Status:       domain.FaxTransmitted,
		Remote:       domain.MustParsePhone("+491512345678"),
		DocumentPath: "blob/fax.pdf",
	}
	panel := renderComponent(t, FaxPanel(FaxPanelProps{Jobs: []domain.FaxJob{job}, Capability: FaxCapability{}, Lang: LangEN}))

	if !strings.Contains(panel, "wp-fax-row") {
		t.Error("stored fax row must survive a capability change")
	}
	if !strings.Contains(panel, "/fax/"+job.ID.String()+"/document") {
		t.Errorf("stored document link must survive a capability change; rendered:\n%s", panel)
	}
}
