package views

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func TestTFallsBackToEnglishThenKey(t *testing.T) {
	if got := T(LangDE, "tab.messages"); got != "Nachrichten" {
		t.Errorf("de lookup: got %q", got)
	}
	if got := T(LangEN, "tab.messages"); got != "Messages" {
		t.Errorf("en lookup: got %q", got)
	}
	if got := T(Lang("fr"), "tab.messages"); got != "Messages" {
		t.Errorf("unsupported language must fall back to English: got %q", got)
	}
	if got := T(LangDE, "key.that.does.not.exist"); got != "key.that.does.not.exist" {
		t.Errorf("unknown key must surface itself: got %q", got)
	}
}

func TestParseLang(t *testing.T) {
	if ParseLang("de") != LangDE {
		t.Error("de must parse")
	}
	if ParseLang("en-GB") != LangEN {
		t.Error("en-GB must fall back to en")
	}
	if ParseLang("") != LangEN {
		t.Error("empty must fall back to en")
	}
}

func TestDictionariesStayInSync(t *testing.T) {
	en := dictionaries[LangEN]
	de := dictionaries[LangDE]
	for key := range en {
		if _, ok := de[key]; !ok {
			t.Errorf("German dictionary missing %q (English users would see it, German ones the raw key fallback)", key)
		}
	}
	for key := range de {
		if _, ok := en[key]; !ok {
			t.Errorf("German dictionary has orphan key %q", key)
		}
	}
}

// TestFormatVerbsMatchAcrossLanguages pins that every key carries the same
// number of format verbs in both languages — e.g. err.messageRejected's
// single %s. A mismatched translation would render %!s(MISSING) (or eat the
// argument) in exactly one language, invisible to any single-locale test.
func TestFormatVerbsMatchAcrossLanguages(t *testing.T) {
	for key, enValue := range dictionaries[LangEN] {
		deValue, ok := dictionaries[LangDE][key]
		if !ok {
			continue // key-set parity is TestDictionariesStayInSync's job
		}
		if got, want := countFormatVerbs(enValue), countFormatVerbs(deValue); got != want {
			t.Errorf(
				"%q: EN carries %d format verb(s), DE carries %d — one language renders %%!s(MISSING)",
				key, got, want,
			)
		}
	}
}

// TestNoUnusedDictionaryKeys is the dictionary's dead-key guard (plan
// T19/M24): every key must appear as a quoted literal in the view
// sources (templ or Go) or the server sources — i18n.go itself is
// excluded, it is where the keys are DEFINED. A key nothing reads is
// drift from a removed surface: it makes the dictionaries lie about
// the copy surface and slows every future audit.
func TestNoUnusedDictionaryKeys(t *testing.T) {
	var scans []string
	templ, err := filepath.Glob("*.templ")
	if err != nil {
		t.Fatal(err)
	}
	scans = append(scans, templ...)
	for _, name := range []string{"*.go", filepath.Join("..", "..", "server", "*.go")} {
		matches, err := filepath.Glob(name)
		if err != nil {
			t.Fatal(err)
		}
		for _, match := range matches {
			if !strings.HasSuffix(match, "_test.go") && !strings.HasSuffix(match, "i18n.go") {
				scans = append(scans, match)
			}
		}
	}
	sources := make([]string, 0, len(scans))
	for _, name := range scans {
		raw, err := os.ReadFile(name)
		if err != nil {
			t.Fatalf("read %s: %v", name, err)
		}
		sources = append(sources, string(raw))
	}
	for key := range dictionaries[LangEN] {
		quoted := `"` + key + `"`
		referenced := false
		for _, source := range sources {
			if strings.Contains(source, quoted) {
				referenced = true
				break
			}
		}
		if !referenced {
			t.Errorf("dictionary key %q is referenced nowhere (views or server) — remove it or wire it", key)
		}
	}
}

func countFormatVerbs(s string) int {
	count := 0
	for i := 0; i < len(s); i++ {
		if s[i] != '%' {
			continue
		}
		if i+1 < len(s) && s[i+1] == '%' { // %% is an escaped literal
			i++
			continue
		}
		count++
	}
	return count
}

// TestEveryReferencedKeyExists pins the dynamic-template key-sync
// guard (plan T27d): every T(lang, "…") literal in the view sources
// must exist in the ENGLISH dictionary. T() surfaces unknown keys as
// the raw key in the page (deliberate at runtime), so a typo'd or
// renamed key renders garbage for users — this fails the build instead.
func TestEveryReferencedKeyExists(t *testing.T) {
	sources, err := filepath.Glob("*.templ")
	if err != nil || len(sources) == 0 {
		t.Fatalf("glob view sources: %v (%d)", err, len(sources))
	}
	re := regexp.MustCompile(`T\((?:props\.Lang|lang)\w*,\s*"([^"]+)"\)`)
	missing := map[string]bool{}
	for _, name := range sources {
		raw, err := os.ReadFile(name)
		if err != nil {
			t.Fatalf("read %s: %v", name, err)
		}
		for _, m := range re.FindAllStringSubmatch(string(raw), -1) {
			key := m[1]
			if _, ok := dictionaries[LangEN][key]; !ok {
				missing[name+": "+key] = true
			}
		}
	}
	for key := range missing {
		t.Errorf("referenced key missing from the dictionary: %s", key)
	}
}
