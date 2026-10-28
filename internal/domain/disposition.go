package domain

import (
	"github.com/larsartmann/go-error-family"
)

// Disposition is what a call ENDED AS — the vocabulary the CRM journal
// and the operator confirm (issue #6) speak. The island can observe
// whether media flowed; it canNOT observe who picked up, so "answered"
// is claimable only by a human confirm, never derived from connection
// state. The 2026-10-27 data-corruption defect (voicemail deposits
// journaled as "answered") came from exactly that derivation.
type Disposition string

const (
	// DispositionAnswered: a human conversed. Set ONLY by an operator
	// confirm — no automated path may mint it.
	DispositionAnswered Disposition = "answered"
	// DispositionMissed: the call never carried media (no-answer, busy,
	// caller gave up, died in setup).
	DispositionMissed Disposition = "missed"
	// DispositionVoicemail: resolved to a voicemail deposit.
	DispositionVoicemail Disposition = "voicemail"
	// DispositionIVR: resolved to an IVR / automated attendant.
	DispositionIVR Disposition = "ivr"
	// DispositionFailed: teardown via the failure class (transport or
	// server error), never a remote party's choice.
	DispositionFailed Disposition = "failed"
	// DispositionUnknown: media flowed but human-vs-machine is
	// undetermined. The honest default for every established call until
	// an operator confirms.
	DispositionUnknown Disposition = "unknown"
)

// ParseDisposition validates a wire or stored disposition string.
func ParseDisposition(raw string) (Disposition, error) {
	switch Disposition(raw) {
	case DispositionAnswered, DispositionMissed, DispositionVoicemail, DispositionIVR, DispositionFailed, DispositionUnknown:
		return Disposition(raw), nil
	default:
		return "", errorfamily.Newf(errorfamily.Rejection, "domain.disposition", "call disposition %q is not one of answered|missed|voicemail|ivr|failed|unknown", raw)
	}
}

// callEndFact is the island-observable end of a call: whether media
// flowed and whether any measurable duration accumulated. This is the
// full observation axis — deliberately, because anything richer
// (audio classification, DTMF menus) would re-introduce the guesswork
// this vocabulary exists to retire.
type callEndFact struct {
	established bool
	mediaTime   bool
}

// endFactDispositions is the ONE home mapping an observed call end to
// its disposition (issue #6 machine table). Rows:
//
//   - no media, no time → missed: the definitive never-connected shape.
//   - every other row → unknown: media or unexplained time means the
//     far end MIGHT be a human, and the table may not guess. Before
//     this table the established row claimed "answered" — the defect.
var endFactDispositions = map[callEndFact]Disposition{
	{established: false, mediaTime: false}: DispositionMissed,
	{established: false, mediaTime: true}:  DispositionUnknown,
	{established: true, mediaTime: false}:  DispositionUnknown,
	{established: true, mediaTime: true}:   DispositionUnknown,
}

// DispositionFromEnd maps the observed end of a call to its
// disposition. The table is total over its fact axis (2×2), so the
// lookup cannot miss; tightening a row to answer more precisely is an
// owner call — edit the rows here and nowhere else.
func DispositionFromEnd(established bool, seconds int) Disposition { //nolint:branching-flow // bool-first reads as the fact axis of the 2x2 table (established x media); reordering would invert the documented axis order
	return endFactDispositions[callEndFact{established: established, mediaTime: seconds > 0}]
}

// legacyOutcomeDispositions maps the retired outcome wire field to the
// disposition vocabulary. The "answered" row IS the stop-the-lying
// ruling: the legacy value was derived from connection state, which
// cannot distinguish a human from a voicemail greeting, so it demotes
// to unknown. It exists for the stale-island compat window and for
// historical rows; the new island sends dispositions directly.
var legacyOutcomeDispositions = map[string]Disposition{
	"answered": DispositionUnknown,
	"missed":   DispositionMissed,
	"":         DispositionUnknown,
}

// LegacyOutcomeDisposition demotes a legacy outcome value. Values
// outside the table (and the empty value) land on unknown — the
// acceptance bar of issue #6: anything unrecognized maps into the
// non-answered class, never the answered class.
func LegacyOutcomeDisposition(outcome string) Disposition {
	if d, ok := legacyOutcomeDispositions[outcome]; ok {
		return d
	}
	return DispositionUnknown
}

// AnsweredClass reports whether this disposition may be counted as a
// human answer. ONLY DispositionAnswered qualifies — unknown maps to
// the failure class by ruling (issue #6 acceptance bar), so journal
// aggregation can never inflate answered counts from guesses.
func (d Disposition) AnsweredClass() bool {
	return d == DispositionAnswered
}

// String renders the wire form (the vocabulary's lowercase member).
func (d Disposition) String() string { return string(d) }
