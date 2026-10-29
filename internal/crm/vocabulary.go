package crm

import "github.com/larsartmann/webphone/internal/domain"

// The vocabulary bridge. Webphone and the CRM each own a closed call-outcome
// vocabulary and neither may dictate the other's: the phone says
// answered|missed|voicemail|ivr|failed|unknown (what the media path observed),
// the CRM says connected|voicemail|ivr|no_answer|wrong_number|dnc|disconnected
// (what the human relationship allows). Sending ours verbatim made the CRM
// reject every answered call with 400 once its vocabulary closed, which turned
// into a 502 toast in the island on every single call.
//
// So the translation lives here, at the consumer's edge, in ONE table: a new
// disposition on either side is a one-line change with a test, never a silent
// 400 in production.
//
// The empty string is the producer's legal "no disposition recorded", and it is
// where every honest unknown lands: claiming `no_answer` for a call whose media
// flowed would write a lie into the customer's journal.
var outcomeForDisposition = map[domain.Disposition]string{
	domain.DispositionAnswered:  "connected",
	domain.DispositionMissed:    "no_answer",
	domain.DispositionVoicemail: "voicemail",
	domain.DispositionIVR:       "ivr",
	domain.DispositionFailed:    "disconnected",
	domain.DispositionUnknown:   "",
}

// crmOutcome translates a webphone disposition into the CRM's closed outcome
// vocabulary. An unrecognised (future) disposition degrades to "no disposition
// recorded" rather than sending a value the CRM would reject.
func crmOutcome(disposition domain.Disposition) string {
	return outcomeForDisposition[disposition]
}
