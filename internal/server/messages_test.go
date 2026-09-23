package server

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"strings"
	"testing"

	"github.com/larsartmann/webphone/internal/config"
)

func TestMessageSendAndThreadFlow(t *testing.T) {
	c := newClient(t)
	c.login("1001", "pw")

	form, contentType := multipartBody(t, map[string]string{"to": "+441632960961", "body": "contract test"}, nil)
	resp, body := c.do(http.MethodPost, "/messages/send", form, contentType)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("send: %d %s", resp.StatusCode, body)
	}

	_, body = c.do(http.MethodGet, "/partials/messages", nil, "")
	page := string(body)
	if !strings.Contains(page, "+441632960961") || !strings.Contains(page, "contract test") {
		t.Fatalf("thread list missing the new thread: %s", page[:min(200, len(page))])
	}
	// T20c: the LIST row offers a dial button OUTSIDE the anchor (a
	// button inside the link would steal the row's navigation).
	if !strings.Contains(page, `data-dial="+441632960961"`) || strings.Contains(page, `<a class="wp-thread-row"[^>]*>[^<]*<button`) {
		t.Errorf("thread list row missing its dial button (or it nests inside the anchor)")
	}
	if !regexp.MustCompile(`<div class="wp-thread-rowwrap"><a[^>]*class="wp-thread-row`).MatchString(page) {
		t.Errorf("row wrapper structure wrong: %.300s", page)
	}

	match := regexp.MustCompile(`href="(/messages/[^"]+)"`).FindSubmatch(body)
	if match == nil {
		t.Fatal("no thread link in list")
	}
	_, body = c.do(http.MethodGet, "/partials"+string(match[1]), nil, "")
	threadView := string(body)
	if !strings.Contains(threadView, "contract test") || !strings.Contains(threadView, "wp-status-sent") {
		t.Fatal("thread view missing message or sent badge (loopback gateway)")
	}
	// The open conversation offers a call to the remote number.
	if !strings.Contains(threadView, `data-dial="+441632960961"`) {
		t.Errorf("thread view missing the call button: %.400s", threadView)
	}
}

func TestMMSAttachmentRoundTrip(t *testing.T) {
	c := newClient(t)
	c.login("1001", "pw")

	png := append([]byte("\x89PNG\r\n\x1a\n"), bytes.Repeat([]byte{0}, 16)...)
	form, contentType := multipartBody(t,
		map[string]string{"to": "+441632960961", "body": "with media"},
		map[string]struct {
			Name    string
			Content []byte
		}{"attachment": {Name: "pic.png", Content: png}})
	resp, body := c.do(http.MethodPost, "/messages/send", form, contentType)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("mms send: %d %s", resp.StatusCode, body)
	}

	_, body = c.do(http.MethodGet, "/partials/messages", nil, "")
	match := regexp.MustCompile(`href="(/messages/[^"]+)"`).FindSubmatch(body)
	_, body = c.do(http.MethodGet, "/partials"+string(match[1]), nil, "")
	attachment := regexp.MustCompile(`href="(/attachments/[^"]+)"`).FindSubmatch(body)
	if attachment == nil {
		t.Fatal("attachment link missing in thread view")
	}
	// Image attachments render inline (plan T21c): a lazy CSS-scaled
	// thumbnail rides the download link.
	if !strings.Contains(string(body), `class="wp-thumb"`) || !strings.Contains(string(body), `loading="lazy"`) {
		t.Fatal("image attachment missing its inline thumbnail")
	}
	resp, body = c.do(http.MethodGet, string(attachment[1]), nil, "")
	if resp.StatusCode != http.StatusOK || !bytes.Equal(body, png) {
		t.Fatalf("attachment round trip broken: %d bytes=%d", resp.StatusCode, len(body))
	}
}

// TestNavPartialAndLiveMarkRead covers the live-polish contract: the nav
// partial re-renders labels (the island's language switch re-fetches it,
// so nav labels follow without a full reload), and the explicit read
// endpoint clears the badge for an open conversation after an SSE swap —
// a live push never re-GETs the partial, so only the client can mark read.
func TestNavPartialAndLiveMarkRead(t *testing.T) {
	server := newTestServer(t)
	c := signIn(t, server)

	deliverInbound(t, server, "+441632960961", "live polish")
	threadRow := regexp.MustCompile(`hx-get="/partials/messages/([^"]+)"`)
	_, listBody := c.do(http.MethodGet, "/partials/messages", nil, "")
	match := threadRow.FindSubmatch(listBody)
	if match == nil {
		t.Fatal("no thread row rendered")
	}
	threadID := string(match[1])

	if badge := unreadBadge(t, c); badge != "1" {
		t.Fatalf("badge before live mark-read: got %q, want 1", badge)
	}

	// Signed-in nav partial: active tab honored, badge fresh, language
	// from the wp-lang cookie (same negotiation the tabs use).
	req, err := http.NewRequest(http.MethodGet, server.URL+"/partials/nav?active=history", nil)
	if err != nil {
		t.Fatal(err)
	}
	var jar []string
	base, _ := url.Parse(server.URL)
	for _, cookie := range c.http.Jar.Cookies(base) {
		jar = append(jar, cookie.Name+"="+cookie.Value)
	}
	req.Header.Set("Cookie", strings.Join(append(jar, "wp-lang=de"), "; "))
	resp, err := c.http.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	navBody := string(readAll(t, resp))
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("nav partial: %d", resp.StatusCode)
	}
	activeLink := regexp.MustCompile(`data-tab="history"[^>]*`).FindString(navBody)
	if !strings.Contains(navBody, `data-tab="history"`) || !strings.Contains(activeLink+navBody, "wp-active") {
		t.Errorf("nav partial lost the active tab: %.200s", navBody)
	}
	if !strings.Contains(navBody, "Nachrichten") {
		t.Errorf("nav partial ignored the wp-lang cookie: %.200s", navBody)
	}
	if !strings.Contains(navBody, `wp-nav-badge">1<`) {
		t.Errorf("nav partial missing the unread badge: %.200s", navBody)
	}

	// The live-swap read endpoint: clears the badge, idempotent.
	if resp, body := c.do(http.MethodPost, "/messages/"+threadID+"/read", nil, ""); resp.StatusCode != http.StatusNoContent {
		t.Fatalf("mark read: %d %s", resp.StatusCode, body)
	}
	if badge := unreadBadge(t, c); badge != "" {
		t.Fatalf("badge after live mark-read: got %q, want none", badge)
	}
	if resp, _ := c.do(http.MethodPost, "/messages/"+threadID+"/read", nil, ""); resp.StatusCode != http.StatusNoContent {
		t.Fatalf("mark read must be idempotent, got %d", resp.StatusCode)
	}

	// Anonymous visitors get labels (no session, no badge): the nav is
	// visible pre-login and the language switch works there too.
	anon := newClient(t)
	resp, body := anon.do(http.MethodGet, "/partials/nav", nil, "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("anonymous nav partial: %d", resp.StatusCode)
	}
	navBody = string(body)
	if !strings.Contains(navBody, `data-tab="messages"`) {
		t.Errorf("anonymous nav partial missing links: %.200s", navBody)
	}
	if strings.Contains(navBody, "wp-nav-badge") {
		t.Errorf("anonymous nav partial must not carry badges: %.200s", navBody)
	}
}

// TestSendClassifiesGatewayOutageAs502 pins the error taxonomy: an
// unreachable send gateway is a server-side failure (502, message saved
// as failed) while a validation mistake stays 422. Before this pin an
// outage masqueraded as a client error with the gateway's internal
// error text leaked into the page.
func TestSendClassifiesGatewayOutageAs502(t *testing.T) {
	server := newTestServerWithConfig(t, "", func(cfg *config.Config) {
		cfg.Gateway = config.Gateway{
			Mode:          config.GatewayWebhook,
			WebhookURL:    "http://127.0.0.1:1", // nothing listens there
			WebhookSecret: "test-secret",
		}
	})
	c := clientFor(t, server)
	c.login("1001", "pw")

	form, contentType := multipartBody(t, map[string]string{"to": "+441632960961", "body": "hi"}, nil)
	resp, body := c.do(http.MethodPost, "/messages/send", form, contentType)
	if resp.StatusCode != http.StatusBadGateway {
		t.Fatalf("gateway outage: %d %s (want 502)", resp.StatusCode, body)
	}
	if !strings.Contains(string(body), "saved as failed") {
		t.Fatalf("502 body must tell the user the message was saved: %.200s", body)
	}
	// The 502 pairs an HX-Trigger toast with the rendered .wp-error
	// banner: the toast announces the failure at once, the responseHandling
	// config swaps the banner into #wp-tab-error so it survives the toast
	// fade. Dropping the header breaks the toast half of that pairing.
	if resp.Header.Get("HX-Trigger") == "" {
		t.Error("502 response lost its HX-Trigger toast header")
	}
	if !strings.Contains(string(body), `class="wp-error"`) {
		t.Error("502 body lost the .wp-error banner the error swap selects")
	}

	form, contentType = multipartBody(t, map[string]string{"to": "+441632960961", "body": ""}, nil)
	resp, _ = c.do(http.MethodPost, "/messages/send", form, contentType)
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("validation mistake: %d (want 422)", resp.StatusCode)
	}
}

// TestThreadViewWarnsOnSelfSend pins the intent-time self-send caution:
// when the thread's remote number is the extension's own DID (config
// identities), the thread view warns BEFORE the user can discover the
// provider's refusal (Telnyx 40310) by failing. A non-self thread must
// never carry the notice.
func TestThreadViewWarnsOnSelfSend(t *testing.T) {
	server := newTestServerWithConfig(t, "", func(c *config.Config) {
		c.Identities = map[string]string{"1001": "+17287289311"}
	})
	c := clientFor(t, server)
	c.login("1001", "pw")

	form, contentType := multipartBody(t, map[string]string{"to": "+17287289311", "body": "self"}, nil)
	resp, body := c.do(http.MethodPost, "/messages/send", form, contentType)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("self send (loopback accepts): %d %s", resp.StatusCode, body)
	}
	_, body = c.do(http.MethodGet, "/partials/messages", nil, "")
	selfLink := regexp.MustCompile(`href="(/messages/[^"]+)"`).FindSubmatch(body)
	if selfLink == nil {
		t.Fatal("no self thread in list")
	}
	_, body = c.do(http.MethodGet, "/partials"+string(selfLink[1]), nil, "")
	view := string(body)
	if !strings.Contains(view, `class="wp-notice"`) || !strings.Contains(view, "This is your own number") {
		t.Errorf("self thread view missing the notice: %.300s", view)
	}

	form, contentType = multipartBody(t, map[string]string{"to": "+441632960961", "body": "other"}, nil)
	resp, body = c.do(http.MethodPost, "/messages/send", form, contentType)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("other send: %d %s", resp.StatusCode, body)
	}
	_, body = c.do(http.MethodGet, "/partials/messages", nil, "")
	for _, link := range regexp.MustCompile(`href="(/messages/[^"]+)"`).FindAllSubmatch(body, -1) {
		if string(link[1]) == string(selfLink[1]) {
			continue
		}
		_, other := c.do(http.MethodGet, "/partials"+string(link[1]), nil, "")
		if strings.Contains(string(other), `class="wp-notice"`) {
			t.Error("non-self thread carries the self-send notice")
		}
	}
}

// TestSendFormsDisableWhileInFlight pins the double-submit guard: every
// outbound send form (new message, thread reply, fax) disables its
// submit button for the duration of the request. Evidence: a live
// self-send test produced TWO identical failed messages because the
// multi-second gateway round-trip left the Send button live.
func TestSendFormsDisableWhileInFlight(t *testing.T) {
	c := newClient(t)
	c.login("1001", "pw")

	const directive = `hx-disabled-elt="find button[type=submit]"`
	_, body := c.do(http.MethodGet, "/partials/messages", nil, "")
	if got := strings.Count(string(body), directive); got != 1 {
		t.Errorf("new-message form: %d disable directives, want 1", got)
	}

	form, contentType := multipartBody(t, map[string]string{"to": "+441632960961", "body": "contract test"}, nil)
	resp, body := c.do(http.MethodPost, "/messages/send", form, contentType)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("send: %d %s", resp.StatusCode, body)
	}
	_, body = c.do(http.MethodGet, "/partials/messages", nil, "")
	match := regexp.MustCompile(`href="(/messages/[^"]+)"`).FindSubmatch(body)
	if match == nil {
		t.Fatal("no thread link in list")
	}
	_, body = c.do(http.MethodGet, "/partials"+string(match[1]), nil, "")
	if got := strings.Count(string(body), directive); got != 1 {
		t.Errorf("reply form: %d disable directives, want 1", got)
	}

	_, body = c.do(http.MethodGet, "/partials/fax", nil, "")
	if !strings.Contains(string(body), directive) {
		t.Error("fax form missing the disable directive")
	}
}

// TestComposerCarriesSegmentCounterAndTextarea pins the composer UX
// contract server-side: the body fields are auto-growing textareas
// (Enter-to-send lives in shell.js §4), each message composer carries
// the segment-counter span, and the reply + fax forms carry the
// attachment-chip container. Without these the shell behaviors have
// nothing to attach to after a swap.
func TestComposerCarriesSegmentCounterAndTextarea(t *testing.T) {
	c := newClient(t)
	c.login("1001", "pw")

	_, body := c.do(http.MethodGet, "/partials/messages", nil, "")
	list := string(body)
	if !strings.Contains(list, `<textarea name="body" class="wp-compose-body"`) {
		t.Errorf("new-message composer body is not a textarea: %.300s", list)
	}
	if got := strings.Count(list, `class="wp-segcount"`); got != 1 {
		t.Errorf("new-message composer: %d segment counters, want 1", got)
	}

	form, contentType := multipartBody(t, map[string]string{"to": "+441632960961", "body": "clock"}, nil)
	resp, body := c.do(http.MethodPost, "/messages/send", form, contentType)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("send: %d %s", resp.StatusCode, body)
	}
	_, body = c.do(http.MethodGet, "/partials/messages", nil, "")
	match := regexp.MustCompile(`href="(/messages/[^"]+)"`).FindSubmatch(body)
	if match == nil {
		t.Fatal("no thread link in list")
	}
	_, body = c.do(http.MethodGet, "/partials"+string(match[1]), nil, "")
	view := string(body)
	if !strings.Contains(view, `<textarea name="body" class="wp-compose-body"`) {
		t.Errorf("reply composer body is not a textarea")
	}
	if got := strings.Count(view, `class="wp-segcount"`); got != 1 {
		t.Errorf("reply composer: %d segment counters, want 1", got)
	}
	if got := strings.Count(view, `class="wp-attach"`); got != 1 {
		t.Errorf("reply composer: %d attachment containers, want 1", got)
	}

	_, body = c.do(http.MethodGet, "/partials/fax", nil, "")
	if !strings.Contains(string(body), `class="wp-attach"`) {
		t.Error("fax form missing the attachment container")
	}
}

// TestBubbleClockFollowsLanguage pins the German 24h convention: with
// wp-lang=de the bubble meta carries a zero-padded 24h clock and never
// a meridiem; the default English rendering keeps its meridiem form.
func TestBubbleClockFollowsLanguage(t *testing.T) {
	c := newClient(t)
	c.login("1001", "pw")

	form, contentType := multipartBody(t, map[string]string{"to": "+441632960961", "body": "clock"}, nil)
	resp, body := c.do(http.MethodPost, "/messages/send", form, contentType)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("send: %d %s", resp.StatusCode, body)
	}
	_, body = c.do(http.MethodGet, "/partials/messages", nil, "")
	match := regexp.MustCompile(`href="(/messages/[^"]+)"`).FindSubmatch(body)
	if match == nil {
		t.Fatal("no thread link in list")
	}

	_, body = c.do(http.MethodGet, "/partials"+string(match[1]), nil, "")
	if !regexp.MustCompile(`\d{1,2}:\d{2}[AP]M`).MatchString(string(body)) {
		t.Errorf("english bubble clock lost its meridiem form")
	}

	req, err := http.NewRequest(http.MethodGet, c.base+"/partials"+string(match[1]), nil)
	if err != nil {
		t.Fatal(err)
	}
	var jar []string
	base, _ := url.Parse(c.base)
	for _, cookie := range c.http.Jar.Cookies(base) {
		jar = append(jar, cookie.Name+"="+cookie.Value)
	}
	req.Header.Set("Cookie", strings.Join(append(jar, "wp-lang=de"), "; "))
	resp, err = c.http.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	deBody := string(readAll(t, resp))
	_ = resp.Body.Close()
	if !regexp.MustCompile(`\b\d{2}:\d{2}\b`).MatchString(deBody) {
		t.Errorf("german bubble clock missing a 24h timestamp")
	}
	if regexp.MustCompile(`\d{1,2}:\d{2}[AP]M`).MatchString(deBody) {
		t.Errorf("german bubble clock still renders a meridiem")
	}
}

// TestFailedBubbleCarriesReasonAndRetry (send-failure plan D + T21):
// a transient (outage) failure renders the persisted reason under the
// bubble AND a retry affordance with the same body; a provider
// rejection renders the reason but NEVER a retry; a delivered verdict
// shows the distinct ✓ badge.
func TestFailedBubbleCarriesReasonAndRetry(t *testing.T) {
	outage := func(cfg *config.Config) {
		cfg.Gateway = config.Gateway{
			Mode:          config.GatewayWebhook,
			WebhookURL:    "http://127.0.0.1:1", // nothing listens there
			WebhookSecret: "test-secret",
		}
	}
	server := newTestServerWithConfig(t, "", outage)
	c := clientFor(t, server)
	c.login("1001", "pw")

	form, contentType := multipartBody(t, map[string]string{"to": "+441632960961", "body": "try me"}, nil)
	if resp, body := c.do(http.MethodPost, "/messages/send", form, contentType); resp.StatusCode != http.StatusBadGateway {
		t.Fatalf("transient send: %d %s (want 502)", resp.StatusCode, body)
	}
	threadView := threadViewFor(t, c, "+441632960961")
	if !strings.Contains(threadView, "wp-status-failed") || !strings.Contains(threadView, "wp-failed-detail") {
		t.Fatalf("failed badge or reason disclosure missing: %.400s", threadView)
	}
	if !strings.Contains(threadView, "wp-retry") || !strings.Contains(threadView, `name="body" value="try me"`) {
		t.Fatalf("transient failure must offer retry with the same body: %.400s", threadView)
	}

	// A provider rejection: same shape, but no retry affordance.
	rejecting := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "217022: not a valid SMS destination", http.StatusUnprocessableEntity)
	}))
	t.Cleanup(rejecting.Close)
	server2 := newTestServerWithConfig(t, "", func(cfg *config.Config) {
		cfg.Gateway = config.Gateway{
			Mode:          config.GatewayWebhook,
			WebhookURL:    rejecting.URL,
			WebhookSecret: "test-secret",
		}
	})
	c2 := clientFor(t, server2)
	c2.login("1001", "pw")
	form, contentType = multipartBody(t, map[string]string{"to": "+441632960977", "body": "no retry"}, nil)
	// Provider refusals answer 422 (send-failure train E: a refusal the
	// user can fix is input feedback, not a system fault); the pin here
	// is the BUBBLE story: reason shown, no retry affordance.
	if resp, body := c2.do(http.MethodPost, "/messages/send", form, contentType); resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("rejected send: %d %s (want 422, the refusal arm)", resp.StatusCode, body)
	}
	view := threadViewFor(t, c2, "+441632960977")
	if !strings.Contains(view, "wp-failed-detail") {
		t.Fatalf("rejection reason disclosure missing: %.400s", view)
	}
	if strings.Contains(view, "wp-retry") {
		t.Fatal("a rejection must not offer retry (it would fail identically)")
	}

	// A delivered verdict: the ✓-marked distinct badge.
	accepting := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"provider_ref":"ref-deliver-1"}`))
	}))
	t.Cleanup(accepting.Close)
	server3 := newTestServerWithConfig(t, "", func(cfg *config.Config) {
		cfg.Gateway = config.Gateway{
			Mode:          config.GatewayWebhook,
			WebhookURL:    accepting.URL,
			WebhookSecret: "test-secret",
		}
	})
	c3 := clientFor(t, server3)
	c3.login("1001", "pw")
	form, contentType = multipartBody(t, map[string]string{"to": "+441632960988", "body": "deliver me"}, nil)
	if resp, body := c3.do(http.MethodPost, "/messages/send", form, contentType); resp.StatusCode != http.StatusOK {
		t.Fatalf("accepted send: %d %s", resp.StatusCode, body)
	}
	view = threadViewFor(t, c3, "+441632960988")
	if !strings.Contains(view, "wp-status-sent") {
		t.Fatalf("sent badge missing before delivery: %.400s", view)
	}
	hookReq, err := http.NewRequest(http.MethodPost, server3.URL+"/hooks/message/status",
		strings.NewReader(`{"provider_ref":"ref-deliver-1","status":"delivered"}`))
	if err != nil {
		t.Fatal(err)
	}
	hookReq.Header.Set("Content-Type", "application/json")
	hookReq.Header.Set("Authorization", "Bearer test-secret")
	hookResp, err := server3.Client().Do(hookReq)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = hookResp.Body.Close() }()
	if hookResp.StatusCode != http.StatusAccepted {
		t.Fatalf("delivery hook: %d (want 202)", hookResp.StatusCode)
	}
	view = threadViewFor(t, c3, "+441632960988")
	if !strings.Contains(view, "wp-status-delivered") || !strings.Contains(view, `aria-hidden="true">✓`) {
		t.Fatalf("delivered badge must carry the distinct check glyph: %.400s", view)
	}
}

// threadViewFor opens the thread partial for the given remote number.
func threadViewFor(t testing.TB, c *client, remote string) string {
	t.Helper()
	_, body := c.do(http.MethodGet, "/partials/messages", nil, "")
	match := regexp.MustCompile(`href="(/messages/[^"]+)"`).FindSubmatch(body)
	if match == nil {
		t.Fatal("no thread link in list")
	}
	_, body = c.do(http.MethodGet, "/partials"+string(match[1]), nil, "")
	view := string(body)
	if !strings.Contains(view, remote) {
		t.Fatalf("thread view for %s not found", remote)
	}
	return view
}

// TestTranscriptCarriesHoverStampsAndJumpChip pins the hover-stamp and
// jump-to-latest markup contract: thread rows and bubble clocks carry a
// full absolute-stamp title (hover = exact moment), and the transcript
// is wrapped with the hidden jump chip shell.js fills on scrolled-away
// live pushes. Without these the shell behaviors have nothing to
// attach to after a swap.
func TestTranscriptCarriesHoverStampsAndJumpChip(t *testing.T) {
	c := newClient(t)
	c.login("1001", "pw")

	form, contentType := multipartBody(t, map[string]string{"to": "+441632960961", "body": "hover stamp"}, nil)
	resp, body := c.do(http.MethodPost, "/messages/send", form, contentType)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("send: %d %s", resp.StatusCode, body)
	}

	_, body = c.do(http.MethodGet, "/partials/messages", nil, "")
	list := string(body)
	if !strings.Contains(list, `class="wp-thread-when" title="`) {
		t.Errorf("thread row when-span lost its absolute-stamp title: %.300s", list)
	}
	if strings.Contains(list, `class="wp-transcript-wrap"`) {
		t.Error("thread list must not render the transcript wrap")
	}

	view := threadViewFor(t, c, "+441632960961")
	if !strings.Contains(view, `class="wp-transcript-wrap"`) {
		t.Error("thread view lost the transcript wrap (jump chip anchor)")
	}
	if !regexp.MustCompile(`class="wp-jump-latest" hidden aria-label="jump to the latest messages"`).MatchString(view) {
		t.Errorf("thread view lost the hidden jump chip: %.400s", view)
	}
	if !strings.Contains(view, `class="wp-bubble-meta"><span title="`) {
		t.Errorf("bubble meta clock lost its absolute-stamp title: %.400s", view)
	}
}

// TestThreadSearchFiltersPanel pins the search contract: the panel route
// (page AND partial) takes q, matches remote numbers and message bodies,
// echoes the query into the search input, renders the quoted no-match
// empty state, and leaves the unfiltered list untouched when q is empty.
func TestThreadSearchFiltersPanel(t *testing.T) {
	c := newClient(t)
	c.login("1001", "pw")

	form, contentType := multipartBody(t, map[string]string{"to": "+441632960961", "body": "the launch code is 50%"}, nil)
	resp, body := c.do(http.MethodPost, "/messages/send", form, contentType)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("send: %d %s", resp.StatusCode, body)
	}
	form, contentType = multipartBody(t, map[string]string{"to": "+491601234567", "body": "totally unrelated"}, nil)
	resp, body = c.do(http.MethodPost, "/messages/send", form, contentType)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("send 2: %d %s", resp.StatusCode, body)
	}

	// Body match in the partial.
	_, list := c.do(http.MethodGet, "/partials/messages?q=launch+code", nil, "")
	if !strings.Contains(string(list), "+441632960961") || strings.Contains(string(list), "+491601234567") {
		t.Fatalf("body search must filter the partial list: %.400s", list)
	}
	// The input echoes the query for morph-safe focus retention.
	if !strings.Contains(string(list), `id="wp-thread-search-input" value="launch code"`) {
		t.Fatalf("search input lost the echoed query: %.400s", list)
	}

	// Remote-number match on the full page route. The composer
	// placeholder carries the example number, so assert on thread rows.
	_, page := c.do(http.MethodGet, "/messages?q=01234567", nil, "")
	pageRows := strings.Count(string(page), `wp-thread-row" href="/messages/`)
	if pageRows != 1 {
		t.Fatalf("remote search must leave exactly one row on the page: %d rows", pageRows)
	}
	if !strings.Contains(string(page), "+491601234567") {
		t.Fatalf("remaining row must be the +49 thread: %.400s", page)
	}

	// LIKE metacharacters stay literal: "50%" matches only the % body.
	_, list = c.do(http.MethodGet, "/partials/messages?q=50%25", nil, "")
	if !strings.Contains(string(list), "+441632960961") || strings.Contains(string(list), "+491601234567") {
		t.Fatalf("literal %% search must not act as a wildcard: %.400s", list)
	}

	// No match renders the quoted empty state, not the all-threads row.
	_, list = c.do(http.MethodGet, "/partials/messages?q=zzznope", nil, "")
	if !strings.Contains(string(list), "zzznope") || strings.Contains(string(list), "wp-thread-row") {
		t.Fatalf("no-match search must render the quoted empty state: %.400s", list)
	}

	// Empty q keeps the plain unfiltered list (quote-terminated match:
	// wp-thread-row must not also count wp-thread-rowwrap).
	_, list = c.do(http.MethodGet, "/partials/messages", nil, "")
	if strings.Count(string(list), `wp-thread-row"`) != 2 {
		t.Fatalf("empty q must list every thread: %.400s", list)
	}
}
