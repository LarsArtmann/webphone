package sso

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"maps"
	"math/big"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"

	"github.com/larsartmann/go-error-family/errorfamilytest"

	"github.com/larsartmann/webphone/internal/authmap"
	"github.com/larsartmann/webphone/internal/domain"
)

const (
	stubClientID     = "webphone"
	stubRedirect     = "https://pbx.example.org/api/auth/oauth2/callback"
	stubPassword     = "dir-password"
	stubUserEmail    = "lars@example.com"
	stubGoodCode     = "good-code"
	stubKid          = "test-key"
	stubSubject      = "user-1"
	stubClientSecret = "s3cret"
)

// stubIDP is a minimal OpenID provider: discovery, JWKS, token and
// userinfo endpoints, RS256-signed ID tokens. It is deliberately hand-rolled
// (no JWT library) so the verification path is tested against the wire
// shape go-oidc actually consumes.
type stubIDP struct {
	t              *testing.T
	server         *httptest.Server
	key            *rsa.PrivateKey
	extraClaims    map[string]any
	userinfoEmail  string
	lastVerifier   string
	lastCode       string
	lastAuthHeader string
}

func newStubIDP(t *testing.T) *stubIDP {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	idp := &stubIDP{
		t:             t,
		key:           key,
		userinfoEmail: stubUserEmail,
		extraClaims: map[string]any{
			"email":          stubUserEmail,
			"email_verified": true,
		},
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/.well-known/openid-configuration", idp.discovery)
	mux.HandleFunc("/jwks", idp.jwks)
	mux.HandleFunc("/token", idp.token)
	mux.HandleFunc("/userinfo", idp.userinfo)
	idp.server = httptest.NewTLSServer(mux)
	t.Cleanup(idp.server.Close)
	return idp
}

func (s *stubIDP) issuer() string { return s.server.URL }

func (s *stubIDP) discovery(w http.ResponseWriter, _ *http.Request) {
	writeStubJSON(s.t, w, map[string]any{
		"issuer":                                s.issuer(),
		"authorization_endpoint":                s.issuer() + "/authorize",
		"token_endpoint":                        s.issuer() + "/token",
		"jwks_uri":                              s.issuer() + "/jwks",
		"userinfo_endpoint":                     s.issuer() + "/userinfo",
		"response_types_supported":              []string{"code"},
		"subject_types_supported":               []string{"public"},
		"id_token_signing_alg_values_supported": []string{"RS256"},
	})
}

func (s *stubIDP) jwks(w http.ResponseWriter, _ *http.Request) {
	writeStubJSON(s.t, w, map[string]any{"keys": []map[string]any{{
		"kty": "RSA",
		"kid": stubKid,
		"use": "sig",
		"alg": "RS256",
		"n":   base64.RawURLEncoding.EncodeToString(s.key.N.Bytes()),
		"e":   base64.RawURLEncoding.EncodeToString(big.NewInt(int64(s.key.E)).Bytes()),
	}}})
}

func (s *stubIDP) token(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad form", http.StatusBadRequest)
		return
	}
	s.lastAuthHeader = r.Header.Get("Authorization")
	// client_secret_basic is the ONLY accepted auth (Pocket ID shape):
	// BasicAuth decodes WITHOUT trimming, so a padded or newline-riding
	// secret fails here exactly like it did against the real IdP.
	id, secret, ok := r.BasicAuth()
	if !ok || id != stubClientID || secret != stubClientSecret {
		http.Error(w, `{"error":"invalid_client"}`, http.StatusUnauthorized)
		return
	}
	s.lastCode = r.PostForm.Get("code")
	s.lastVerifier = r.PostForm.Get("code_verifier")
	if s.lastCode != stubGoodCode {
		http.Error(w, `{"error":"invalid_grant"}`, http.StatusBadRequest)
		return
	}
	claims := map[string]any{
		"iss": s.issuer(),
		"aud": stubClientID,
		"sub": stubSubject,
		"exp": 4102444800, // 2100-01-01: no clock dependence
		"iat": 1700000000,
	}
	maps.Copy(claims, s.extraClaims)
	writeStubJSON(s.t, w, map[string]any{
		"access_token": "stub-access-token",
		"token_type":   "Bearer",
		"expires_in":   3600,
		"id_token":     s.sign(s.t, claims),
	})
}

func (s *stubIDP) sign(t *testing.T, claims map[string]any) string {
	t.Helper()
	header := map[string]any{"alg": "RS256", "kid": stubKid, "typ": "JWT"}
	headerJSON, err := json.Marshal(header)
	if err != nil {
		t.Fatal(err)
	}
	claimsJSON, err := json.Marshal(claims)
	if err != nil {
		t.Fatal(err)
	}
	signingInput := base64.RawURLEncoding.EncodeToString(headerJSON) + "." +
		base64.RawURLEncoding.EncodeToString(claimsJSON)
	digest := sha256.Sum256([]byte(signingInput))
	signature, err := rsa.SignPKCS1v15(rand.Reader, s.key, crypto.SHA256, digest[:])
	if err != nil {
		t.Fatal(err)
	}
	return signingInput + "." + base64.RawURLEncoding.EncodeToString(signature)
}

func (s *stubIDP) userinfo(w http.ResponseWriter, _ *http.Request) {
	writeStubJSON(s.t, w, map[string]any{
		"sub":            stubSubject,
		"email":          s.userinfoEmail,
		"email_verified": true,
	})
}

func writeStubJSON(t *testing.T, w http.ResponseWriter, body any) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(body); err != nil {
		t.Errorf("write JSON: %v", err)
	}
}

// testContext carries the stub server's TLS-trusting client into both
// consumers (go-oidc reads oidc.ClientContext, the exchange reads
// oauth2.HTTPClient).
func testContext(idp *stubIDP) context.Context {
	ctx := oidc.ClientContext(context.Background(), idp.server.Client())
	return context.WithValue(ctx, oauth2.HTTPClient, idp.server.Client())
}

func newTestClient(t *testing.T, idp *stubIDP, tweak func(*Config)) *Client {
	t.Helper()
	secretFile := filepath.Join(t.TempDir(), "client-secret")
	if err := os.WriteFile(secretFile, []byte("  "+stubClientSecret+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	cfg := Config{
		Issuer:           idp.issuer(),
		ClientID:         stubClientID,
		ClientSecretFile: secretFile,
		RedirectURL:      stubRedirect,
		Scopes:           []string{"openid", "email", "profile"},
		Users: map[string]authmap.User{
			stubUserEmail: {Email: stubUserEmail, DisplayName: "Lars", Extensions: []domain.Extension{domain.MustParseExtension("1000")}},
		},
		ExtensionPasswordFiles: map[string]string{"1000": passwordFile(t)},
	}
	if tweak != nil {
		tweak(&cfg)
	}
	client, err := New(cfg)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return client
}

func passwordFile(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "ext1000")
	if err := os.WriteFile(path, []byte(stubPassword+"\n"), 0o640); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestChallengeCarriesThePKCEAndOIDCParameters(t *testing.T) {
	idp := newStubIDP(t)
	client := newTestClient(t, idp, nil)

	challenge, err := client.Challenge(testContext(idp))
	if err != nil {
		t.Fatalf("Challenge: %v", err)
	}
	u, err := url.Parse(challenge.URL)
	if err != nil {
		t.Fatalf("parse authorize URL: %v", err)
	}
	if got := u.Query().Get("client_id"); got != stubClientID {
		t.Errorf("client_id = %q, want %q", got, stubClientID)
	}
	if got := u.Query().Get("redirect_uri"); got != stubRedirect {
		t.Errorf("redirect_uri = %q, want %q", got, stubRedirect)
	}
	if got := u.Query().Get("response_type"); got != "code" {
		t.Errorf("response_type = %q, want code", got)
	}
	if got := u.Query().Get("state"); got != challenge.State || got == "" {
		t.Errorf("state = %q, want the minted %q", got, challenge.State)
	}
	if got := u.Query().Get("nonce"); got != challenge.Nonce || got == "" {
		t.Errorf("nonce = %q, want the minted %q", got, challenge.Nonce)
	}
	if got := u.Query().Get("code_challenge_method"); got != "S256" {
		t.Errorf("code_challenge_method = %q, want S256", got)
	}
	if got := u.Query().Get("code_challenge"); got != pkceChallenge(challenge.Verifier) {
		t.Errorf("code_challenge = %q, want the S256 of the verifier", got)
	}
	if scopes := u.Query().Get("scope"); !strings.Contains(scopes, "openid") {
		t.Errorf("scope = %q, want openid", scopes)
	}
}

func TestCompleteMintsTheMappedIdentityAndPassword(t *testing.T) {
	idp := newStubIDP(t)
	client := newTestClient(t, idp, nil)
	challenge, err := client.Challenge(testContext(idp))
	if err != nil {
		t.Fatalf("Challenge: %v", err)
	}
	idp.extraClaims["nonce"] = challenge.Nonce

	user, password, err := client.Complete(testContext(idp), stubGoodCode, challenge.Verifier, challenge.Nonce)
	if err != nil {
		t.Fatalf("Complete: %v", err)
	}
	if user.Email != stubUserEmail {
		t.Errorf("email = %q, want %q", user.Email, stubUserEmail)
	}
	if user.SessionExtension().String() != "1000" {
		t.Errorf("session extension = %q, want 1000", user.SessionExtension().String())
	}
	if password != stubPassword {
		t.Errorf("password = %q, want the file's trimmed value", password)
	}
	if idp.lastVerifier != challenge.Verifier {
		t.Errorf("token request code_verifier = %q, want %q (PKCE broke)", idp.lastVerifier, challenge.Verifier)
	}
	if idp.lastCode != stubGoodCode {
		t.Errorf("token request code = %q, want %q", idp.lastCode, stubGoodCode)
	}
}

func TestCompleteRejectsAMismatchedNonce(t *testing.T) {
	idp := newStubIDP(t)
	client := newTestClient(t, idp, nil)
	challenge, err := client.Challenge(testContext(idp))
	if err != nil {
		t.Fatalf("Challenge: %v", err)
	}
	idp.extraClaims["nonce"] = "some-other-attempt"

	_, _, err = client.Complete(testContext(idp), stubGoodCode, challenge.Verifier, challenge.Nonce)
	errorfamilytest.AssertCode(t, err, "sso.nonce.mismatch")
}

func TestCompleteRejectsAnUnverifiedEmail(t *testing.T) {
	idp := newStubIDP(t)
	client := newTestClient(t, idp, nil)
	challenge, err := client.Challenge(testContext(idp))
	if err != nil {
		t.Fatalf("Challenge: %v", err)
	}
	idp.extraClaims["nonce"] = challenge.Nonce
	idp.extraClaims["email_verified"] = false

	_, _, err = client.Complete(testContext(idp), stubGoodCode, challenge.Verifier, challenge.Nonce)
	errorfamilytest.AssertCode(t, err, "sso.email_unverified")
}

func TestCompleteRejectsAnUnmappedEmail(t *testing.T) {
	idp := newStubIDP(t)
	client := newTestClient(t, idp, nil)
	challenge, err := client.Challenge(testContext(idp))
	if err != nil {
		t.Fatalf("Challenge: %v", err)
	}
	idp.extraClaims["nonce"] = challenge.Nonce
	idp.extraClaims["email"] = "stranger@example.com"

	_, _, err = client.Complete(testContext(idp), stubGoodCode, challenge.Verifier, challenge.Nonce)
	errorfamilytest.AssertCode(t, err, "sso.unmapped_email")
}

func TestCompleteFallsBackToUserinfoWhenTheIDTokenHasNoEmail(t *testing.T) {
	idp := newStubIDP(t)
	client := newTestClient(t, idp, nil)
	challenge, err := client.Challenge(testContext(idp))
	if err != nil {
		t.Fatalf("Challenge: %v", err)
	}
	delete(idp.extraClaims, "email")
	delete(idp.extraClaims, "email_verified")
	idp.extraClaims["nonce"] = challenge.Nonce

	user, _, err := client.Complete(testContext(idp), stubGoodCode, challenge.Verifier, challenge.Nonce)
	if err != nil {
		t.Fatalf("Complete via userinfo: %v", err)
	}
	if user.Email != stubUserEmail {
		t.Errorf("email = %q, want the userinfo address", user.Email)
	}
}

func TestCompleteFailsClosedOnABadExchange(t *testing.T) {
	idp := newStubIDP(t)
	client := newTestClient(t, idp, nil)
	challenge, err := client.Challenge(testContext(idp))
	if err != nil {
		t.Fatalf("Challenge: %v", err)
	}
	_, _, err = client.Complete(testContext(idp), "rejected-code", challenge.Verifier, challenge.Nonce)
	errorfamilytest.AssertCode(t, err, "sso.exchange")
}

func TestCompleteFailsClosedOnAMissingOrEmptySecret(t *testing.T) {
	idp := newStubIDP(t)
	client := newTestClient(t, idp, nil)
	challenge, err := client.Challenge(testContext(idp))
	if err != nil {
		t.Fatalf("Challenge: %v", err)
	}
	idp.extraClaims["nonce"] = challenge.Nonce

	missing := newTestClient(t, idp, func(cfg *Config) {
		cfg.ClientSecretFile = filepath.Join(t.TempDir(), "gone")
	})
	_, _, err = missing.Complete(testContext(idp), stubGoodCode, challenge.Verifier, challenge.Nonce)
	errorfamilytest.AssertCode(t, err, "sso.client_secret.read")

	emptyFile := filepath.Join(t.TempDir(), "empty")
	if err := os.WriteFile(emptyFile, []byte(" \n"), 0o600); err != nil {
		t.Fatal(err)
	}
	empty := newTestClient(t, idp, func(cfg *Config) { cfg.ClientSecretFile = emptyFile })
	_, _, err = empty.Complete(testContext(idp), stubGoodCode, challenge.Verifier, challenge.Nonce)
	errorfamilytest.AssertCode(t, err, "sso.client_secret.empty")
}

func TestChallengeReportsAnUnreachableProvider(t *testing.T) {
	idp := newStubIDP(t)
	client := newTestClient(t, idp, func(cfg *Config) { cfg.Issuer = "https://127.0.0.1:1" })
	_, err := client.Challenge(testContext(idp))
	errorfamilytest.AssertCode(t, err, "sso.discovery")
}

func TestNewFailsClosedOnHalfAConfiguration(t *testing.T) {
	for _, tc := range []struct {
		name string
		cfg  Config
		code string
	}{
		{"no issuer", Config{ClientID: "x", ClientSecretFile: "y", RedirectURL: "https://h/cb"}, "sso.issuer"},
		{"no client id", Config{Issuer: "https://i", ClientSecretFile: "y", RedirectURL: "https://h/cb"}, "sso.client_id"},
		{"no secret file", Config{Issuer: "https://i", ClientID: "x", RedirectURL: "https://h/cb"}, "sso.client_secret_file"},
		{"no redirect", Config{Issuer: "https://i", ClientID: "x", ClientSecretFile: "y"}, "sso.redirect_url"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := New(tc.cfg)
			errorfamilytest.AssertCode(t, err, tc.code)
		})
	}
}

// The token endpoint speaks client_secret_basic ONLY (the Pocket ID
// shape, pinned 2026-10-24): the Authorization header must carry the
// base64(client_id:client_secret) pair with the TRIMMED secret
// byte-exact. A padded or newline-riding secret rides the header raw
// (oauth2 does not trim either) and was the 2026-10-14 invalid_client
// burn; client_secret_post form auth is refused outright.
func TestTokenEndpointRequiresClientSecretBasic(t *testing.T) {
	t.Parallel()

	idp := newStubIDP(t)
	client := newTestClient(t, idp, nil)

	ctx := testContext(idp)
	ch, err := client.Challenge(ctx)
	if err != nil {
		t.Fatalf("Challenge: %v", err)
	}
	idp.extraClaims["nonce"] = ch.Nonce
	if _, _, err := client.Complete(ctx, stubGoodCode, ch.Verifier, ch.Nonce); err != nil {
		t.Fatalf("Complete: %v", err)
	}

	if !strings.HasPrefix(idp.lastAuthHeader, "Basic ") {
		t.Fatalf("token exchange must use client_secret_basic, got Authorization: %q", idp.lastAuthHeader)
	}
	raw, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(idp.lastAuthHeader, "Basic "))
	if err != nil {
		t.Fatalf("Authorization header is not base64: %v", err)
	}
	id, secret, ok := strings.Cut(string(raw), ":")
	if !ok || id != stubClientID {
		t.Fatalf("Basic pair client id drifted: %q", string(raw))
	}
	if secret != stubClientSecret {
		t.Fatalf("Basic pair secret must be the trimmed file value %q byte-exact, got %q (a padded or newline-riding secret is the invalid_client burn shape)", stubClientSecret, secret)
	}

	// client_secret_post (credentials in the form) is refused: no Basic
	// header means no auth.
	form := url.Values{
		"code":          {stubGoodCode},
		"code_verifier": {ch.Verifier},
		"client_id":     {stubClientID},
		"client_secret": {stubClientSecret},
	}
	resp, err := idp.server.Client().PostForm(idp.issuer()+"/token", form)
	if err != nil {
		t.Fatalf("post-form probe: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("client_secret_post must be refused with 401, got %d", resp.StatusCode)
	}

	// A Basic header carrying the RAW file bytes (padded + newline) is
	// the exact 2026-10-14 burn twin: refused, never half-accepted.
	req, err := http.NewRequest(http.MethodPost, idp.issuer()+"/token", strings.NewReader(form.Encode()))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(stubClientID, "  "+stubClientSecret+"\n")
	resp2, err := idp.server.Client().Do(req)
	if err != nil {
		t.Fatalf("raw-secret probe: %v", err)
	}
	defer resp2.Body.Close()
	if resp2.StatusCode != http.StatusUnauthorized {
		t.Fatalf("the raw (padded+newline) secret must be refused with 401, got %d", resp2.StatusCode)
	}
}
