// Black-box behavior specs for the session surface: what a signed-in
// extension (or an attacker with her browser) can observe through HTTP.
// These run beside the white-box suite in package server; the Ginkgo
// specs describe the login/logout contract from the outside, the way the
// island experiences it.
package server_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"regexp"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/larsartmann/webphone/internal/blob"
	"github.com/larsartmann/webphone/internal/config"
	"github.com/larsartmann/webphone/internal/fax"
	"github.com/larsartmann/webphone/internal/gateway"
	"github.com/larsartmann/webphone/internal/messaging"
	"github.com/larsartmann/webphone/internal/pbx"
	"github.com/larsartmann/webphone/internal/server"
	"github.com/larsartmann/webphone/internal/session"
	"github.com/larsartmann/webphone/internal/store"
)

// webphone is the subject under test: one running instance on a random
// port, plus the cookie-carrying browser client that talks to it.
type webphone struct {
	base   string
	client *http.Client
	token  string
}

// start boots a fresh instance; phoneAPIURL is the upstream the login
// credential proof rides ("" = loopback dev mode, no verification).
func start(t GinkgoTInterface, phoneAPIURL string) *webphone {
	db, err := store.Open(":memory:")
	Expect(err).NotTo(HaveOccurred())
	t.Cleanup(func() { _ = db.Close() })

	blobs, err := blob.New(t.TempDir())
	Expect(err).NotTo(HaveOccurred())
	phoneAPI, err := pbx.NewClient(phoneAPIURL)
	Expect(err).NotTo(HaveOccurred())

	hubs := server.NewHubs()
	notifier := server.NewNotifier(hubs, store.NewMessages(db), store.NewFaxes(db))
	cfg := config.Config{
		Addr: ":0", DataDir: t.TempDir(), WebsocketPath: "/sip",
		SessionTTL: time.Hour,
		Gateway:    config.Gateway{Mode: config.GatewayLoopback, WebhookSecret: "test-secret"},
	}
	messages := store.NewMessages(db)
	faxes := store.NewFaxes(db)
	handler := server.New(server.Deps{
		Config:    cfg,
		Sessions:  session.NewStore(time.Hour),
		Messages:  messages,
		Faxes:     faxes,
		Contacts:  store.NewContacts(db),
		Messaging: messaging.New(messages, blobs, gateway.NewMessageGateway(cfg.Gateway, gateway.DefaultClient()), notifier.MessagesChanged),
		Fax:       fax.New(faxes, blobs, gateway.NewFaxGateway(cfg.Gateway, gateway.DefaultClient()), notifier.FaxChanged),
		PhoneAPI:  phoneAPI,
		Hubs:      hubs,
		DB:        db,
		BlobRoot:  blobs.Root(),
	})

	jars, err := cookiejar.New(nil)
	Expect(err).NotTo(HaveOccurred())
	httpClient := &http.Client{Jar: jars}
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)

	wp := &webphone{base: srv.URL, client: httpClient}
	wp.token = wp.csrfToken()
	return wp
}

func mustURL(base string) *url.URL {
	parsed, err := url.Parse(base)
	Expect(err).NotTo(HaveOccurred())
	return parsed
}

var _ = Describe("Signing in", func() {
	var wp *webphone

	When("the PBX directory accepts the credentials (loopback dev: no phone API)", func() {
		BeforeEach(func() {
			wp = start(GinkgoT(), "")
		})

		It("opens a session cookie that unlocks the tab data", func() {
			wp.loginExpecting("1001", "secret", http.StatusCreated)
			Expect(wp.sessionCookie()).NotTo(BeNil())

			resp, _ := wp.get("/partials/nav")
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
		})

		It("rotates the CSRF token so the very next POST still validates", func() {
			wp.loginExpecting("1001", "secret", http.StatusCreated)
			wp.adoptFreshCSRFToken()

			resp, _ := wp.postRaw("/messages/not-a-thread/read", nil)
			Expect(resp.StatusCode).To(Equal(http.StatusNotFound))
		})
	})

	When("the extension is malformed", func() {
		BeforeEach(func() {
			wp = start(GinkgoT(), "")
		})

		It("rejects the login without minting a session", func() {
			wp.loginExpecting("not an extension!", "secret", http.StatusBadRequest)
			Expect(wp.sessionCookie()).To(BeNil())
		})
	})

	DescribeTable("rejecting broken login bodies",
		func(body string, wantStatus int) {
			wp = start(GinkgoT(), "")
			resp, _ := wp.postRaw("/api/session", []byte(body))
			Expect(resp.StatusCode).To(Equal(wantStatus))
		},
		Entry("a missing password", `{"extension":"1001"}`, http.StatusBadRequest),
		Entry("invalid JSON", `{oops`, http.StatusBadRequest),
	)
})

var _ = Describe("Signing out", func() {
	It("ends the session so the tab data stops answering", func() {
		wp := start(GinkgoT(), "")
		wp.loginExpecting("1001", "secret", http.StatusCreated)
		wp.adoptFreshCSRFToken()

		resp, _ := wp.do(http.MethodDelete, "/api/session", nil)
		Expect(resp.StatusCode).To(Equal(http.StatusNoContent))

		resp, _ = wp.get("/partials/nav?active=messages")
		Expect(resp.StatusCode).To(Equal(http.StatusUnauthorized))
	})
})

var _ = Describe("A forged login without a reachable PBX", func() {
	It("fails closed instead of minting an unverified session", func() {
		// A phone API configured but unreachable must NOT fall back to
		// trust: the forged-session fix deliberately fails closed (502).
		wp := start(GinkgoT(), "http://127.0.0.1:1")
		wp.loginExpecting("1001", "anything", http.StatusBadGateway)
		Expect(wp.sessionCookie()).To(BeNil())
	})
})

// --- helpers -------------------------------------------------------------

func (wp *webphone) csrfToken() string {
	resp, err := wp.client.Get(wp.base + "/")
	Expect(err).NotTo(HaveOccurred())
	defer func() { _ = resp.Body.Close() }()
	body := readAllBody(resp)
	match := regexp.MustCompile(`name="csrf-token" content="([^"]+)"`).FindSubmatch(body)
	Expect(match).NotTo(BeNil(), "served page must carry the csrf meta tag")
	return string(match[1])
}

func (wp *webphone) adoptFreshCSRFToken() {
	resp, err := wp.client.Get(wp.base + "/api/csrf")
	Expect(err).NotTo(HaveOccurred())
	defer func() { _ = resp.Body.Close() }()
	Expect(resp.StatusCode).To(Equal(http.StatusOK))
	var got struct {
		Token string `json:"token"`
	}
	Expect(json.Unmarshal(readAllBody(resp), &got)).To(Succeed())
	Expect(got.Token).NotTo(BeEmpty())
	wp.token = got.Token
}

func (wp *webphone) loginExpecting(extension, password string, wantStatus int) {
	payload, err := json.Marshal(map[string]string{"extension": extension, "password": password})
	Expect(err).NotTo(HaveOccurred())
	resp, _ := wp.postRaw("/api/session", payload)
	Expect(resp.StatusCode).To(Equal(wantStatus))
}

func (wp *webphone) postRaw(path string, body []byte) (*http.Response, []byte) {
	req, err := http.NewRequest(http.MethodPost, wp.base+path, bytes.NewReader(body))
	Expect(err).NotTo(HaveOccurred())
	req.Header.Set("Content-Type", "application/json")
	if wp.token != "" {
		req.Header.Set("X-CSRF-Token", wp.token)
	}
	resp, err := wp.client.Do(req)
	Expect(err).NotTo(HaveOccurred())
	return resp, readAllBody(resp)
}

func (wp *webphone) post(path string, body []byte) (*http.Response, []byte) {
	return wp.postRaw(path, body)
}

func (wp *webphone) get(path string) (*http.Response, []byte) {
	return wp.do(http.MethodGet, path, nil)
}

func (wp *webphone) do(method, path string, body []byte) (*http.Response, []byte) {
	req, err := http.NewRequest(method, wp.base+path, bytes.NewReader(body))
	Expect(err).NotTo(HaveOccurred())
	if wp.token != "" {
		req.Header.Set("X-CSRF-Token", wp.token)
	}
	resp, err := wp.client.Do(req)
	Expect(err).NotTo(HaveOccurred())
	return resp, readAllBody(resp)
}

func (wp *webphone) sessionCookie() *http.Cookie {
	for _, cookie := range wp.client.Jar.Cookies(mustURL(wp.base)) {
		if cookie.Name == "webphone_session" && cookie.Value != "" {
			return cookie
		}
	}
	return nil
}

func readAllBody(resp *http.Response) []byte {
	var buf bytes.Buffer
	_, _ = buf.ReadFrom(resp.Body)
	return buf.Bytes()
}
