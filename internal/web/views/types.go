// The view-owned mirror shapes for the two external seams whose wire
// types used to leak into the templates (the arch review's P2
// "reach-through" finding): the PBX voicemail feed (pbx.Voicemail*)
// and the supervised-agent client (textagent.Reviews/Draft). The server
// boundary converts once (panels.go owns the one home per conversion);
// the templates speak only these shapes, so an upstream wire change
// stops at a converter instead of rippling through templ.
//
// readmodels.ThreadSummary and reports.* are NOT mirrors: those two
// packages are the ratified homes of their view shapes (the 2026-11-12
// vocabulary move + the reports pure-projection ruling), so views
// consuming them is the architecture, not a leak.
package views

// VoicemailSummary is the panel's New/Old counter line.
type VoicemailSummary struct {
	New int
	Old int
}

// VoicemailItem is one stored voicemail row: the fields the panel and
// its rows render.
type VoicemailItem struct {
	UUID      string
	CIDNumber string
	CIDName   string
	Seconds   int
	Created   int64
	Read      bool
	AudioURL  string
}

// AgentDraft is one pending AI draft in the review queue.
type AgentDraft struct {
	ID         string
	Contact    string
	Trigger    string
	Text       string
	EnGloss    string
	Language   string
	Confidence int
	CreatedAt  string
}

// AgentFrozen is one frozen-contact note row.
type AgentFrozen struct {
	Contact string
	Reason  string
}

// AgentSuppressed is one suppressed-keyword note row.
type AgentSuppressed struct {
	Contact string
	Keyword string
}

// AgentReviews is the drafts panel's whole state: the queues plus the
// mode vocabulary.
type AgentReviews struct {
	Drafts      []AgentDraft
	Frozen      []AgentFrozen
	Suppressed  []AgentSuppressed
	Mode        string
	ModeSource  string
	DefaultMode string
	Modes       []string
}
