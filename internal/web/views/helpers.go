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

// avatarFor derives the two-glyph avatar label for a contact name or
// phone number: the country signum of a number ("+1", "+4", "0"), the
// first letters of a name's words ("AK"), or "?" for blanks.
// Deterministic, so the same peer renders the same mark everywhere.
func avatarFor(nameOrNumber string) string {
	trimmed := strings.TrimSpace(nameOrNumber)
	runes := []rune(trimmed)
	if len(runes) == 0 {
		return "?"
	}
	if runes[0] == '+' || (runes[0] >= '0' && runes[0] <= '9') {
		if runes[0] == '+' {
			for _, r := range runes[1:] {
				if r >= '0' && r <= '9' {
					return "+" + string(r)
				}
			}
			return "+"
		}
		return string(runes[0])
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

// isSelfThread reports whether the thread's remote number is the
// extension's own presented DID (config identities): providers refuse
// self-addressed sends (Telnyx 40310), so the thread view warns at
// intent time instead of letting the user discover it by failing. Both
// sides run through ParsePhone because config DIDs may carry spacing
// the sanitized thread remote never has; an unparseable or absent
// identity simply never warns.
func isSelfThread(identity string, remote domain.Phone) bool {
	if identity == "" {
		return false
	}
	own, err := domain.ParsePhone(identity)
	if err != nil {
		return false
	}
	return own == remote
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
