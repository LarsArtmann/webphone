package vcard

import (
	"strings"
	"testing"
)

func TestDecodeHandlesRealWorldFiles(t *testing.T) {
	document := "BEGIN:VCARD\r\n" +
		"VERSION:3.0\r\n" +
		"FN:Alice A. Wonderland\r\n" +
		"N:Wonderland;Alice;A.;;;\r\n" +
		"TEL;TYPE=CELL:+44 1632 960961\r\n" +
		"END:VCARD\r\n" +
		"BEGIN:VCARD\r\n" +
		"VERSION:4.0\r\n" +
		"N:Bob;Robert;;;\r\n" +
		"TEL;TYPE=\"home,voice\":tel:+491700000000\r\n" +
		"END:VCARD\r\n" +
		"BEGIN:VCARD\r\n" +
		"FN:No Number\r\n" +
		"END:VCARD\r\n"

	cards := Decode([]byte(document))
	if len(cards) != 2 {
		t.Fatalf("cards: got %d, want 2 (numberless card skipped): %+v", len(cards), cards)
	}
	if cards[0].Name != "Alice A. Wonderland" || cards[0].Number != "+441632960961" {
		t.Errorf("card 0: %+v", cards[0])
	}
	if cards[1].Name != "Robert Bob" || cards[1].Number != "+491700000000" {
		t.Errorf("card 1 (N fallback, quoted param): %+v", cards[1])
	}
}

func TestDecodeUnfoldsAndUnescapes(t *testing.T) {
	document := "BEGIN:VCARD\r\n" +
		"FN:Sem icolon\r\n" +
		" N;\r\n" +
		" ame With Continuation\r\n" +
		"TEL:+49123\r\n" +
		"END:VCARD\r\n"
	cards := Decode([]byte(document))
	if len(cards) != 1 {
		t.Fatalf("cards: %+v", cards)
	}
	if cards[0].Number != "+49123" {
		t.Errorf("folded lines must not break TEL: %+v", cards[0])
	}
}

func TestDecodeToleratesGarbage(t *testing.T) {
	if cards := Decode([]byte("this is not a vCard")); len(cards) != 0 {
		t.Errorf("garbage must decode to zero cards: %+v", cards)
	}
	if cards := Decode(nil); len(cards) != 0 {
		t.Errorf("empty input must decode to zero cards")
	}
}

func TestEncodeRoundTrips(t *testing.T) {
	source := []Card{
		{Name: "Comma, Semi; Colon", Number: "+441632960961"},
		{Name: "Plain", Number: "+493012345678"},
	}
	encoded := Encode(source)
	if !strings.Contains(string(encoded), `FN:Comma\, Semi\; Colon`) {
		t.Errorf("name must be escaped: %s", encoded)
	}

	decoded := Decode(encoded)
	if len(decoded) != 2 {
		t.Fatalf("round trip cards: %+v", decoded)
	}
	if decoded[0].Name != "Comma, Semi; Colon" || decoded[0].Number != "+441632960961" {
		t.Errorf("round trip card 0: %+v", decoded[0])
	}
	if decoded[1].Name != "Plain" || decoded[1].Number != "+493012345678" {
		t.Errorf("round trip card 1: %+v", decoded[1])
	}
}
