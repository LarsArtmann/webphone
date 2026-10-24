package push

import (
	"bytes"
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json/v2"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/larsartmann/go-error-family"
)

// FCM HTTP v1 wire: a service account (JSON key file) mints a
// short-lived OAuth2 access token via a signed JWT grant, and each
// send POSTs one message to the project's messages:send endpoint with
// the device's registration token as the address. Both round trips
// live in this file; no Google SDK dependency — the wire is two POSTs
// and an RS256 signature, all stdlib.
const (
	// fcmRequestTimeout bounds one send (token exchange or publish).
	fcmRequestTimeout = 10 * time.Second
	// fcmSendURLTemplate is the messages:send endpoint; %s is the
	// project id from the service account.
	fcmSendURLTemplate = "https://fcm.googleapis.com/v1/projects/%s/messages:send"
	// fcmScope is the one OAuth2 scope Firebase messaging needs.
	fcmScope = "https://www.googleapis.com/auth/firebase.messaging"
	// fcmGrantType is the JWT-bearer grant (RFC 7523).
	fcmGrantType = "urn:ietf:params:oauth:grant-type:jwt-bearer"
	// tokenExpirySlack refreshes the cached access token this long
	// before its exp (clock skew + a send in flight at the boundary).
	tokenExpirySlack = 60 * time.Second
	// jwtLifetime is the assertion's validity window.
	jwtLifetime = time.Hour
	// fcmBodyLimit bounds how much of an error body is buffered for
	// classification before the socket is returned.
	fcmBodyLimit = 64 << 10
)

// serviceAccount is the subset of Google's service-account JSON the
// sender needs.
type serviceAccount struct {
	ProjectID   string `json:"project_id"`
	ClientEmail string `json:"client_email"`
	PrivateKey  string `json:"private_key"`
	TokenURI    string `json:"token_uri"`
}

// FcmSender publishes ring facts over FCM HTTP v1.
type FcmSender struct {
	project  string
	email    string
	key      *rsa.PrivateKey
	tokenURI string
	sendURL  string
	client   *http.Client

	mu          sync.Mutex
	tokenValue  string
	tokenExpiry time.Time
}

// NewFcmSender loads the service-account credentials file and builds
// the sender. A missing or malformed file fails closed (Rejection):
// a half-configured FCM lane must refuse the boot, not the first ring.
// An empty path means "not configured".
func NewFcmSender(credentialsFile string) (*FcmSender, error) {
	if credentialsFile == "" {
		return &FcmSender{}, nil
	}
	raw, err := os.ReadFile(credentialsFile)
	if err != nil {
		return nil, errorfamily.WrapRejectionf(err, "config.push.fcm", "read fcm credentials file %q", credentialsFile)
	}
	var account serviceAccount
	if err := json.Unmarshal(raw, &account); err != nil {
		return nil, errorfamily.WrapRejectionf(err, "config.push.fcm", "parse fcm credentials file %q", credentialsFile)
	}
	switch {
	case account.ProjectID == "":
		return nil, errorfamily.NewRejection("config.push.fcm", fmt.Sprintf("fcm credentials file %q lacks project_id", credentialsFile))
	case account.ClientEmail == "":
		return nil, errorfamily.NewRejection("config.push.fcm", fmt.Sprintf("fcm credentials file %q lacks client_email", credentialsFile))
	case account.PrivateKey == "":
		return nil, errorfamily.NewRejection("config.push.fcm", fmt.Sprintf("fcm credentials file %q lacks private_key", credentialsFile))
	case account.TokenURI == "":
		return nil, errorfamily.NewRejection("config.push.fcm", fmt.Sprintf("fcm credentials file %q lacks token_uri", credentialsFile))
	}

	block, _ := pem.Decode([]byte(account.PrivateKey))
	if block == nil {
		return nil, errorfamily.NewRejection("config.push.fcm", fmt.Sprintf("fcm private_key in %q is not PEM", credentialsFile))
	}
	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, errorfamily.WrapRejectionf(err, "config.push.fcm", "fcm private_key in %q is not PKCS#8", credentialsFile)
	}
	rsaKey, ok := key.(*rsa.PrivateKey)
	if !ok {
		return nil, errorfamily.NewRejection("config.push.fcm", fmt.Sprintf("fcm private_key in %q is %T, want RSA", credentialsFile, key))
	}

	return &FcmSender{
		project:  account.ProjectID,
		email:    account.ClientEmail,
		key:      rsaKey,
		tokenURI: account.TokenURI,
		sendURL:  fmt.Sprintf(fcmSendURLTemplate, account.ProjectID),
		client:   &http.Client{Timeout: fcmRequestTimeout},
	}, nil
}

// Enabled reports whether the transport is configured.
func (s *FcmSender) Enabled() bool { return s != nil && s.project != "" }

// Describe names the transport for logs.
func (s *FcmSender) Describe() string { return TransportFcm }

// fcmMessage is the messages:send body. Data (not notification) is
// deliberate: the shell owns its CallStyle rendering; a stock FCM
// notification payload would double-ring with a shade entry the shell
// never styled.
type fcmMessage struct {
	Message struct {
		Token   string            `json:"token"`
		Data    map[string]string `json:"data"`
		Android struct {
			Priority string `json:"priority"`
		} `json:"android"`
	} `json:"message"`
}

// Send ships one ring fact to the device's registration token.
func (s *FcmSender) Send(ctx context.Context, target string, ring Ring) error {
	if !s.Enabled() {
		return ErrDisabled
	}
	accessToken, err := s.fetchAccessToken(ctx)
	if err != nil {
		return err
	}

	priority := "NORMAL"
	if ring.Event == EventRing {
		priority = "HIGH"
	}
	var body fcmMessage
	body.Message.Token = target
	body.Message.Data = map[string]string{
		"event":   string(ring.Event),
		"from":    ring.From,
		"call_id": ring.CallID,
	}
	body.Message.Android.Priority = priority
	payload, err := json.Marshal(body)
	if err != nil {
		return errorfamily.WrapInfrastructuref(err, "push.fcm_send", "encode fcm message")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.sendURL, bytes.NewReader(payload))
	if err != nil {
		return errorfamily.WrapInfrastructuref(err, "push.fcm_send", "build fcm request")
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := s.client.Do(req)
	if err != nil {
		return errorfamily.WrapInfrastructuref(err, "push.fcm_send", "publish to fcm")
	}
	errorBody, _ := io.ReadAll(io.LimitReader(resp.Body, fcmBodyLimit))
	_ = resp.Body.Close() //nolint:erraudit // best-effort socket return; the body is drained

	switch {
	case resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden:
		return errorfamily.WrapRejectionf(ErrUnauthorized, "push.fcm_send", "fcm rejected the access token (%d)", resp.StatusCode)
	case resp.StatusCode == http.StatusNotFound || strings.Contains(string(errorBody), fcmUnregisteredMarker):
		return errorfamily.WrapInfrastructuref(ErrUnregistered, "push.fcm_send", "fcm reports the device token dead (%d)", resp.StatusCode)
	case statusOK(resp.StatusCode):
		return nil
	default:
		return errorfamily.Newf(errorfamily.Infrastructure, "push.fcm_send", "fcm answered %d", resp.StatusCode)
	}
}

// fcmUnregisteredMarker is the error reason FCM carries for a dead
// registration token (the 404 lane and the 200-with-error body variant
// both use it; the 404 status alone already decides, the marker
// catches the rest).
const fcmUnregisteredMarker = "UNREGISTERED"

// fetchAccessToken returns a cached OAuth2 access token, minting a
// fresh JWT-bearer grant when the cached one is stale.
func (s *FcmSender) fetchAccessToken(ctx context.Context) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.tokenValue != "" && time.Now().Before(s.tokenExpiry.Add(-tokenExpirySlack)) {
		return s.tokenValue, nil
	}
	assertion, expiry, err := s.signedJWT()
	if err != nil {
		return "", err
	}
	form := url.Values{
		"grant_type": {fcmGrantType},
		"assertion":  {assertion},
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.tokenURI, strings.NewReader(form.Encode()))
	if err != nil {
		return "", errorfamily.WrapInfrastructuref(err, "push.fcm_token", "build token request")
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := s.client.Do(req)
	if err != nil {
		return "", errorfamily.WrapInfrastructuref(err, "push.fcm_token", "exchange fcm access token")
	}
	grantBody, _ := io.ReadAll(io.LimitReader(resp.Body, fcmBodyLimit))
	_ = resp.Body.Close() //nolint:erraudit // best-effort socket return; the body is drained

	var grant struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}
	if resp.StatusCode == http.StatusOK {
		if err := json.Unmarshal(grantBody, &grant); err != nil {
			return "", errorfamily.WrapInfrastructuref(err, "push.fcm_token", "decode fcm token grant")
		}
	}
	if resp.StatusCode != http.StatusOK || grant.AccessToken == "" {
		if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
			return "", errorfamily.WrapRejectionf(ErrUnauthorized, "push.fcm_token", "fcm token endpoint rejected the service account (%d)", resp.StatusCode)
		}
		return "", errorfamily.Newf(errorfamily.Infrastructure, "push.fcm_token", "fcm token endpoint answered %d", resp.StatusCode)
	}
	if grant.ExpiresIn <= 0 {
		grant.ExpiresIn = int(time.Until(expiry).Seconds())
	}
	s.tokenValue = grant.AccessToken
	s.tokenExpiry = time.Now().Add(time.Duration(grant.ExpiresIn) * time.Second)
	return s.tokenValue, nil
}

// signedJWT mints the RS256 assertion for the JWT-bearer grant and
// returns it with its expiry.
func (s *FcmSender) signedJWT() (string, time.Time, error) {
	now := time.Now()
	expiry := now.Add(jwtLifetime)
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"RS256","typ":"JWT"}`))
	claims, err := json.Marshal(map[string]any{
		"iss":   s.email,
		"scope": fcmScope,
		"aud":   s.tokenURI,
		"iat":   now.Unix(),
		"exp":   expiry.Unix(),
	})
	if err != nil {
		return "", time.Time{}, errorfamily.WrapInfrastructuref(err, "push.fcm_token", "encode jwt claims")
	}
	signingInput := header + "." + base64.RawURLEncoding.EncodeToString(claims)
	digest := sha256.Sum256([]byte(signingInput))
	signature, err := rsa.SignPKCS1v15(rand.Reader, s.key, crypto.SHA256, digest[:])
	if err != nil {
		return "", time.Time{}, errorfamily.WrapInfrastructuref(err, "push.fcm_token", "sign jwt")
	}
	return signingInput + "." + base64.RawURLEncoding.EncodeToString(signature), expiry, nil
}
