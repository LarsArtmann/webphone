package server

import (
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/larsartmann/webphone/internal/domain"
)

func mustURL(t *testing.T, raw string) *url.URL {
	t.Helper()
	parsed, err := url.Parse(raw)
	if err != nil {
		t.Fatal(err)
	}
	return parsed
}

func setLangCookie(t *testing.T, c *client, lang string) {
	t.Helper()
	c.http.Jar.SetCookies(mustURL(t, c.base), []*http.Cookie{{Name: "wp-lang", Value: lang}})
}

func TestGermanLanguageViaCookieAndHeader(t *testing.T) {
	server := newTestServer(t)
	c := signIn(t, server)

	// The island's language switch writes wp-lang; the tabs follow.
	setLangCookie(t, c, "de")
	_, body := c.do(http.MethodGet, "/partials/contacts", nil, "")
	page := string(body)
	for _, want := range []string{"Kontakte", "Speichern", "vCard exportieren", "Anrufen"} {
		if !strings.Contains(page, want) {
			t.Errorf("German contacts panel missing %q: %.300s", want, page)
		}
	}

	// The shell nav follows too.
	_, body = c.do(http.MethodGet, "/", nil, "")
	page = string(body)
	for _, want := range []string{">Nachrichten<", ">Mailbox<", ">Anrufe<", ">Einstellungen<"} {
		if !strings.Contains(page, want) {
			t.Errorf("German nav missing %q", want)
		}
	}

	// English stays the default (no cookie, plain client).
	plain := signIn(t, server)
	_, body = plain.do(http.MethodGet, "/partials/contacts", nil, "")
	if !strings.Contains(string(body), ">Save<") || strings.Contains(string(body), "Speichern") {
		t.Errorf("default language must be English: %.300s", body)
	}

	// Accept-Language: de without a cookie.
	headerClient := signIn(t, server)
	req, err := http.NewRequest(http.MethodGet, server.URL+"/partials/contacts", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("X-CSRF-Token", headerClient.token)
	req.Header.Set("Accept-Language", "de-DE,de;q=0.9")
	resp, err := headerClient.http.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	bodyBytes := readAll(t, resp)
	_ = resp.Body.Close()
	if !strings.Contains(string(bodyBytes), "vCard exportieren") {
		t.Errorf("Accept-Language de must translate: %.300s", bodyBytes)
	}
}

func TestHubRemembersExtensionLanguage(t *testing.T) {
	hubs := NewHubs()
	ext := domain.MustParseExtension("1001")

	if got := hubs.Lang(ext); got != "en" {
		t.Errorf("unknown extension must default to en, got %q", got)
	}
	hubs.SetLang(ext, "de")
	if got := hubs.Lang(ext); got != "de" {
		t.Errorf("hub must remember the extension language, got %q", got)
	}
}
