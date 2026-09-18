package views

import "testing"

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
