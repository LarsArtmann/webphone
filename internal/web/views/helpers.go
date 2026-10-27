package views

import (
	"strconv"
	"strings"
	"time"

	"github.com/larsartmann/webphone/internal/domain"
)

// fmtInt renders counts for templ (templ children cannot call strconv
// directly with a plain int conversion).
func fmtInt(n int) string { return strconv.Itoa(n) }

// dispositionFilterOptions builds the History verdict filter's rows: the
// empty "any" row plus the five verdicts an operator can confirm.
// DispositionUnknown is deliberately absent — it is a machine observation,
// not a verdict anyone confirms, so it never filters. Labels are resolved
// here so the option component stays free of i18n.
func dispositionFilterOptions(lang Lang, current string) []filterOption {
	options := make([]filterOption, 0, len(dispositionFilterLabels)+1)
	options = append(options, filterOption{Label: T(lang, "history.disposition.any")})
	for _, verdict := range dispositionFilterLabels {
		options = append(options, filterOption{
			Value:    string(verdict.value),
			Label:    T(lang, verdict.labelKey),
			Selected: current == string(verdict.value),
		})
	}
	return options
}

// dispositionFilterLabel pairs a verdict with its dictionary key.
type dispositionFilterLabel struct {
	value    domain.Disposition
	labelKey string
}

// dispositionFilterLabels is the ONE home of the filterable verdicts, in
// presentation order.
var dispositionFilterLabels = []dispositionFilterLabel{
	{domain.DispositionAnswered, "history.disposition.answered"},
	{domain.DispositionVoicemail, "history.disposition.voicemail"},
	{domain.DispositionIVR, "history.disposition.ivr"},
	{domain.DispositionMissed, "history.disposition.missed"},
	{domain.DispositionFailed, "history.disposition.failed"},
}

// avatarFor derives the two-glyph avatar label for a contact name or
// phone number: the LAST TWO DIGITS of a number ("61"), the first
// letters of a name's words ("AK"), or "?" for blanks. The tail — not
// the country signum the label once showed — because every peer on a
// single-country deployment shared one meaningless "+4"; the digits
// differ per peer at a glance. Deterministic, so the same peer renders
// the same mark everywhere.
func avatarFor(nameOrNumber string) string {
	trimmed := strings.TrimSpace(nameOrNumber)
	runes := []rune(trimmed)
	if len(runes) == 0 {
		return "?"
	}
	if runes[0] == '+' || (runes[0] >= '0' && runes[0] <= '9') {
		digits := make([]rune, 0, len(runes))
		for _, r := range runes {
			if r >= '0' && r <= '9' {
				digits = append(digits, r)
			}
		}
		switch len(digits) {
		case 0:
			return "+"
		case 1:
			return string(digits)
		default:
			return string(digits[len(digits)-2:])
		}
	}
	letters := make([]rune, 0, 2)
	wantLetter := true
	for _, r := range runes {
		if r == ' ' {
			wantLetter = true
			continue
		}
		if wantLetter {
			letters = append(letters, r)
			wantLetter = false
			if len(letters) == 2 {
				break
			}
		}
	}
	return string(letters)
}

// avatarHue maps a name or number onto a stable hue (0-359) so each peer
// keeps a consistent avatar tint across every surface.
func avatarHue(nameOrNumber string) int {
	sum := 0
	for _, r := range nameOrNumber {
		sum = (sum*31 + int(r)) % 360
	}
	return sum
}

// avatarHueClass buckets the stable hue into a 10° CSS class. The page
// CSP is style-src 'self' without unsafe-inline, so a per-avatar style
// ATTRIBUTE is blocked by the browser (the hue never applies); the hue
// rides a wp-av-h<deg> class instead. 10° buckets are visually
// indistinguishable from the exact hue.
func avatarHueClass(nameOrNumber string) string {
	return "wp-av-h" + fmtInt(avatarHue(nameOrNumber)/10*10)
}

// displayName prefers the CRM-resolved contact name and falls back to the
// raw number, so no surface ever renders blank when the integration is off
// or the number is unknown. The number itself stays on the functional
// attributes (data-dial, hidden inputs, avatar seed) — only the visible
// label switches to the name.
func displayName(number string, names map[string]string) string {
	if name, ok := names[number]; ok && name != "" {
		return name
	}
	return number
}

// isSelfThread reports whether the thread's remote address is the
// extension's own presented DID (config identities): providers refuse
// self-addressed sends (Telnyx 40310), so the thread view warns at
// intent time instead of letting the user discover it by failing. Both
// sides run through the phone parse because config DIDs may carry
// spacing the sanitized thread remote never has; an unparseable or
// absent identity simply never warns. Chat-network threads can never
// be self-threads — the DID is a telephony identity.
func isSelfThread(identity string, remote domain.Address) bool {
	if identity == "" || remote.Network() != domain.NetworkSMS {
		return false
	}
	own, err := domain.ParsePhone(identity)
	if err != nil {
		return false
	}
	return own.String() == remote.String()
}

// networkLabel is the chat-network prefix beside a thread identity:
// the proper noun for chat networks, "" for telephony — the default
// lane wears no badge (most conversations are SMS; a badge everywhere
// is noise). Proper nouns stay untranslated in both languages.
func networkLabel(n domain.Network) string {
	switch n {
	case domain.NetworkSignal:
		return "Signal · "
	case domain.NetworkMatrix:
		return "Matrix · "
	case domain.NetworkSession:
		return "Session · "
	default:
		return ""
	}
}

// formatFor is the one home of the language switch behind the timestamp
// helpers: German renders its de convention, English keeps its en form
// byte-stable so existing pins and the stack E2E never churn.
func formatFor(lang Lang, t time.Time, de, en string) string {
	if lang == LangDE {
		return t.Local().Format(de)
	}
	return t.Local().Format(en)
}

// formatClock renders a wall-clock time for bubble meta: German gets
// the 24h convention ("16:09"), English the meridiem form
// ("4:09PM") it has always shown.
func formatClock(lang Lang, t time.Time) string {
	return formatFor(lang, t, "15:04", time.Kitchen)
}

// formatStamp renders a beyond-24h timestamp for row meta: German gets
// the numeric date convention ("22.09. 16:09"), English its
// existing "Sep 22, 16:09".
func formatStamp(lang Lang, t time.Time) string {
	return formatFor(lang, t, "02.01. 15:04", "Jan 2, 15:04")
}

// fullStamp renders the unambiguous absolute moment (date, year, time)
// for hover titles: relative labels ("3h") and clock-only meta ("4:09PM")
// gain the exact instant on hover instead of hiding it. German uses the
// ISO-shaped numeric convention, English its meridiem form.
func fullStamp(lang Lang, t time.Time) string {
	return formatFor(lang, t, "02.01.2006 15:04", "Jan 2, 2006, 3:04PM")
}
