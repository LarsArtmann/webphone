// Package vcard encodes and decodes the small vCard subset the contacts
// tab needs: name (FN, or N as fallback) plus one phone number per card.
// Parsing is deliberately tolerant — address books in the wild fold
// lines, quote parameters, and escape punctuation.
package vcard

import (
	"strings"
)

// Card is one imported/exported contact.
type Card struct {
	Name   string
	Number string
}

// Decode parses every contact card in a vCard 3.0/4.0 document. Cards
// without any TEL are skipped; a missing FN falls back to the N field.
func Decode(data []byte) []Card {
	lines := unfold(strings.Split(strings.ReplaceAll(string(data), "\r\n", "\n"), "\n"))

	var (
		cards   []Card
		current *Card
		name    string
		nField  string
	)
	flush := func() {
		if current == nil {
			return
		}
		if current.Number != "" {
			current.Name = firstNonEmpty(name, displayNameFromN(nField), current.Number)
			cards = append(cards, *current)
		}
		current = nil
		name, nField = "", ""
	}

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		switch strings.ToUpper(trimmed) {
		case "BEGIN:VCARD":
			flush()
			current = &Card{}
			continue
		case "END:VCARD":
			flush()
			continue
		}
		if current == nil {
			continue
		}

		prop, value, ok := splitProperty(trimmed)
		if !ok {
			continue
		}
		switch prop.Name {
		case "FN":
			name = unescape(value)
		case "N":
			nField = value
		case "TEL":
			if current.Number == "" {
				current.Number = sanitizeNumber(value)
			}
		}
	}
	flush()
	return cards
}

// Encode renders cards as a vCard 3.0 document (the most widely accepted
// flavor for import elsewhere).
func Encode(cards []Card) []byte {
	var out strings.Builder
	for _, card := range cards {
		out.WriteString("BEGIN:VCARD\r\n")
		out.WriteString("VERSION:3.0\r\n")
		out.WriteString("FN:")
		out.WriteString(escape(card.Name))
		out.WriteString("\r\n")
		out.WriteString("N:")
		out.WriteString(escape(card.Name))
		out.WriteString(";;;;\r\n")
		out.WriteString("TEL;TYPE=CELL:")
		out.WriteString(card.Number)
		out.WriteString("\r\n")
		out.WriteString("END:VCARD\r\n")
	}
	return []byte(out.String())
}

// unfold joins RFC continuation lines (leading space or tab).
func unfold(lines []string) []string {
	unfolded := make([]string, 0, len(lines))
	for _, line := range lines {
		if len(line) > 0 && (line[0] == ' ' || line[0] == '\t') && len(unfolded) > 0 {
			unfolded[len(unfolded)-1] += line[1:]
			continue
		}
		unfolded = append(unfolded, line)
	}
	return unfolded
}

// property is one "NAME;param=value:value" line, parameters dropped.
type property struct {
	Name  string
	Value string
}

// splitProperty separates the property name (and its parameters) from
// the value at the first colon outside a quoted string.
func splitProperty(line string) (property, string, bool) {
	var inQuotes bool
	for i := 0; i < len(line); i++ {
		switch line[i] {
		case '"':
			inQuotes = !inQuotes
		case ':':
			if inQuotes {
				continue
			}
			head := line[:i]
			if cut, _, found := strings.Cut(head, ";"); found {
				head = cut
			}
			return property{Name: strings.ToUpper(strings.TrimSpace(head))}, line[i+1:], true
		}
	}
	return property{}, "", false
}

// sanitizeNumber strips separators that leak out of phone fields and
// the "tel:" URI scheme vCard 4.0 allows.
func sanitizeNumber(value string) string {
	value = strings.TrimSpace(value)
	if scheme, rest, found := strings.Cut(value, ":"); found && strings.EqualFold(scheme, "tel") {
		value = rest
	}
	return strings.TrimSpace(strings.Map(func(r rune) rune {
		switch r {
		case ' ', '-', '(', ')', '.':
			return -1
		}
		return r
	}, value))
}

// unescape resolves the vCard text escapes (\n \, \; \\).
func unescape(value string) string {
	var out strings.Builder
	for i := 0; i < len(value); i++ {
		if value[i] == '\\' && i+1 < len(value) {
			switch value[i+1] {
			case 'n', 'N':
				out.WriteByte(' ')
				i++
				continue
			case ',', ';', '\\':
				out.WriteByte(value[i+1])
				i++
				continue
			}
		}
		out.WriteByte(value[i])
	}
	return out.String()
}

// escape applies the same escapes on the way out.
func escape(value string) string {
	replacer := strings.NewReplacer("\\", "\\\\", ",", "\\,", ";", "\\;", "\n", "\\n")
	return replacer.Replace(value)
}

func displayNameFromN(n string) string {
	parts := strings.Split(n, ";")
	if len(parts) < 2 {
		return ""
	}
	return strings.Trim(strings.TrimSpace(parts[1])+" "+strings.TrimSpace(parts[0]), " ")
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}
