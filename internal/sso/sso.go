// Package sso logs a browser in through an external OpenID Connect
// provider (Pocket ID for this stack) and hands the caller a mapped
// identity plus the SIP directory password — the same payload every other
// login mode mints its session from.
//
// Design rules:
//   - the webphone session stays the single session truth: the OIDC tokens
//     are used once, here, and never reach the browser;
//   - discovery is LAZY: a provider that is unreachable must fail the
//     login attempt with an actionable error, never the process boot (the
//     PBX phone must keep answering while the IdP is down);
//   - the client secret is read from its runtime file on every exchange,
//     so rotation via the secrets pipeline needs no restart;
//   - identity mapping is authmap's — an unmapped email fails closed
//     exactly like a passkey login for an unmapped account.
package sso

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/larsartmann/go-error-family"
	"golang.org/x/oauth2"

	"github.com/larsartmann/webphone/internal/authmap"
	"github.com/larsartmann/webphone/internal/domain"
)

// discoveryTimeout bounds the lazy provider lookup: the IdP normally
// answers in milliseconds on loopback; a hung one must not hang a login
// request (and with it a worker) for minutes.
const discoveryTimeout = 10 * time.Second

// providerTTL bounds the positive discovery cache: a long-running
// process re-discovers after this window so IdP endpoint or key
// rotation is picked up without a restart. Failures are never cached —
// an outage costs a login attempt, never the boot.
const providerTTL = 15 * time.Minute

// Config is the resolved OIDC configuration: the operator-declared client
// shape plus the mapping the composition root translated from config.
type Config struct {
	Issuer                 string
	ClientID               string
	ClientSecretFile       string
	RedirectURL            string
	Scopes                 []string
	Users                  map[string]authmap.User
	ExtensionPasswordFiles map[string]string
	// AllowedGroups is the in-app IdP group allowlist (empty = the
	// check is off). Non-empty, a login must carry at least one of
	// these groups in its claims or it is denied before the mapping
	// is consulted.
	AllowedGroups []string
}

// Challenge is one authorization attempt: the URL the browser must visit
// plus the state/nonce/verifier values the caller stores in short-lived
// cookies and hands back to Complete.
type Challenge struct {
	URL      string
	State    string
	Nonce    string
	Verifier string
}

// Client performs the authorization-code + PKCE dance against one
// provider. Safe for concurrent use.
type Client struct {
	cfg    Config
	mapped *authmap.Map
	http   *http.Client

	mu        sync.Mutex
	provider  *oidc.Provider
	fetchedAt time.Time
	inflight  chan providerResult
}

// providerResult is the single-flight handoff: every goroutine that
// arrives while one discovery is running shares its outcome instead of
// stacking its own network attempt behind the mutex.
type providerResult struct {
	p   *oidc.Provider
	err error
}

// New validates the static shape and builds the mapping. No network is
// touched: an IdP outage is a login failure, not a boot failure.
func New(cfg Config) (*Client, error) {
	switch {
	case cfg.Issuer == "":
		return nil, errorfamily.NewRejection("sso.issuer", "auth.oauth2.issuer is required (the IdP's discovery base URL, e.g. https://auth.example.org)")
	case cfg.ClientID == "":
		return nil, errorfamily.NewRejection("sso.client_id", "auth.oauth2.client_id is required")
	case cfg.ClientSecretFile == "":
		return nil, errorfamily.NewRejection("sso.client_secret_file", "auth.oauth2.client_secret_file is required (the OIDC client secret, read per login)")
	case cfg.RedirectURL == "":
		return nil, errorfamily.NewRejection("sso.redirect_url", "auth.oauth2.redirect_url is required (the registered callback, e.g. https://pbx.example.org/api/auth/oauth2/callback)")
	}
	mapped, err := authmap.New(cfg.Users, cfg.ExtensionPasswordFiles)
	if err != nil {
		return nil, errorfamily.WrapRejection(err, "sso.mapping", "build the email→extension mapping") //nolint:erraudit // pure config verdict: nothing failed, no ctx to carry
	}
	if len(cfg.Scopes) == 0 {
		cfg.Scopes = []string{oidc.ScopeOpenID, "email", "profile"}
	}
	return &Client{
		cfg:    cfg,
		mapped: mapped,
		http:   &http.Client{Timeout: 15 * time.Second},
	}, nil
}

// Challenge builds one authorization URL with fresh state, nonce and PKCE
// verifier values.
func (c *Client) Challenge(ctx context.Context) (Challenge, error) {
	provider, err := c.providerFor(ctx)
	if err != nil {
		return Challenge{}, err
	}
	state, err := randomToken()
	if err != nil {
		return Challenge{}, errorfamily.WrapInfrastructure(err, "sso.state", "generate the OIDC state")
	}
	nonce, err := randomToken()
	if err != nil {
		return Challenge{}, errorfamily.WrapInfrastructure(err, "sso.nonce", "generate the OIDC nonce")
	}
	verifier, err := randomToken()
	if err != nil {
		return Challenge{}, errorfamily.WrapInfrastructure(err, "sso.verifier", "generate the PKCE verifier")
	}
	oc := c.oauthConfig(provider.Endpoint(), "")
	url := oc.AuthCodeURL(state,
		oauth2.SetAuthURLParam("nonce", nonce),
		oauth2.SetAuthURLParam("code_challenge", pkceChallenge(verifier)),
		oauth2.SetAuthURLParam("code_challenge_method", "S256"),
	)
	return Challenge{URL: url, State: state, Nonce: nonce, Verifier: verifier}, nil
}

// Complete exchanges the authorization code, verifies the ID token
// (signature, issuer, audience, expiry, nonce) and resolves the mapped
// identity plus its SIP directory password. Every failure is a
// classification-carrying error, actionable in the journal.
func (c *Client) Complete(ctx context.Context, code, verifier, nonce string) (authmap.User, string, error) {
	empty := authmap.User{}
	if code == "" {
		return empty, "", errorfamily.NewRejection("sso.code", "the IdP redirected back without an authorization code")
	}
	if verifier == "" {
		return empty, "", errorfamily.NewRejection("sso.verifier", "the PKCE verifier cookie is missing — restart the sign-in from this browser")
	}
	provider, err := c.providerFor(ctx)
	if err != nil {
		return empty, "", err
	}
	secret, err := c.clientSecret()
	if err != nil {
		return empty, "", err
	}
	oc := c.oauthConfig(provider.Endpoint(), secret)
	token, err := oc.Exchange(ctx, code, oauth2.SetAuthURLParam("code_verifier", verifier))
	if err != nil {
		return empty, "", errorfamily.WrapRejection(err, "sso.exchange",
			"exchange the authorization code with the IdP (check the registered callback URL and the client secret)")
	}
	rawIDToken, _ := token.Extra("id_token").(string)
	if rawIDToken == "" {
		return empty, "", errorfamily.NewRejection("sso.id_token", "the IdP's token response carried no id_token (is the openid scope enabled?)")
	}
	idToken, err := provider.Verifier(&oidc.Config{ClientID: c.cfg.ClientID}).Verify(ctx, rawIDToken)
	if err != nil {
		return empty, "", errorfamily.WrapRejection(err, "sso.id_token.verify", "verify the IdP's ID token")
	}
	if idToken.Nonce != nonce {
		return empty, "", errorfamily.NewRejection("sso.nonce.mismatch", "the ID token nonce does not match this sign-in attempt — restart the sign-in from this browser")
	}
	claims, err := c.claims(ctx, provider, token, idToken)
	if err != nil {
		return empty, "", err
	}
	// Audit line: who knocked, with which IdP-side groups. The groups
	// CLAIM is admission evidence for the journal only — it is never
	// copied onto the mapped user or the session (the mapping and the
	// IdP-side group link are the two gates, not token contents).
	slog.Info("webphone: oidc login claims",
		"email", claims.Email,
		"username", claims.PreferredUsername,
		"groups", strings.Join(claims.Groups, ","))
	if claims.EmailVerified != nil && !*claims.EmailVerified {
		return empty, "", errorfamily.NewRejection("sso.email_unverified",
			"the IdP reports %q as unverified — verify the address in the IdP before signing in").WithContext("email", claims.Email)
	}
	// The in-app group allowlist: the second gate, after the IdP-side
	// client-group link. Fail-closed both ways — a token with no groups
	// claim never passes a non-empty allowlist, and the check is off
	// entirely when the operator configures none.
	if len(c.cfg.AllowedGroups) > 0 {
		allowed := false
		for _, claim := range claims.Groups {
			for _, want := range c.cfg.AllowedGroups {
				if claim == want {
					allowed = true
					break
				}
			}
			if allowed {
				break
			}
		}
		if !allowed {
			return empty, "", errorfamily.NewRejection("sso.groups_denied",
				"the IdP account %q carries none of the allowed groups (%s) — add the account to a required access group in the IdP").WithContext("email", claims.Email).WithContext("allowed_groups", strings.Join(c.cfg.AllowedGroups, ", "))
		}
	}
	mapped, ok := c.mapped.Resolve(claims.Email)
	if !ok {
		return empty, "", errorfamily.NewRejection("sso.unmapped_email",
			"the IdP account %q is not in the users mapping (auth.users) — add the mapping or remove the account").WithContext("email", claims.Email)
	}
	password, err := c.mapped.SIPPassword(mapped.SessionExtension())
	if err != nil {
		return empty, "", err //nolint:wrapcheck // authmap owns the classification and the operator copy
	}
	return mapped, password, nil
}

// Resolve exposes the mapping for callers that render identities (the
// server's whoami line reverse-resolves by extension).
func (c *Client) Resolve(email string) (authmap.User, bool) { return c.mapped.Resolve(email) }

// MappedByExtension reverse-resolves a session extension to its mapped user.
func (c *Client) MappedByExtension(extension domain.Extension) (authmap.User, bool) {
	return c.mapped.MappedByExtension(extension)
}

// claims is the identity payload this login needs. The ID token carries
// them on a well-behaved IdP; Pocket ID's ACCESS tokens famously carry no
// profile claims, and its ID token may omit them for some client shapes,
// so a missing email falls back to the userinfo endpoint (the same
// merge doctrine Stalwart's native OIDC path uses).
func (c *Client) claims(ctx context.Context, provider *oidc.Provider, token *oauth2.Token, idToken *oidc.IDToken) (idClaims, error) {
	var claims idClaims
	if err := idToken.Claims(&claims); err != nil {
		return idClaims{}, errorfamily.WrapRejection(err, "sso.claims", "parse the ID token claims")
	}
	if claims.Email != "" {
		return claims, nil
	}
	info, err := provider.UserInfo(ctx, oauth2.StaticTokenSource(token))
	if err != nil {
		return idClaims{}, errorfamily.WrapRejection(err, "sso.userinfo", "fetch the userinfo claims (the ID token carried no email)")
	}
	if err := info.Claims(&claims); err != nil {
		return idClaims{}, errorfamily.WrapRejection(err, "sso.userinfo.claims", "parse the userinfo claims")
	}
	if claims.Email == "" {
		return idClaims{}, errorfamily.NewRejection("sso.email_missing", "the IdP returned no email claim (add the email scope to auth.oauth2.scopes)")
	}
	return claims, nil
}

// idClaims is the narrow claim set this login consumes.
type idClaims struct {
	Email             string   `json:"email"`
	EmailVerified     *bool    `json:"email_verified"`
	PreferredUsername string   `json:"preferred_username"`
	Groups            []string `json:"groups"`
}

func (c *Client) oauthConfig(endpoint oauth2.Endpoint, secret string) *oauth2.Config {
	endpoint.AuthStyle = oauth2.AuthStyleInHeader
	return &oauth2.Config{
		ClientID:     c.cfg.ClientID,
		ClientSecret: secret,
		Endpoint:     endpoint,
		RedirectURL:  c.cfg.RedirectURL,
		Scopes:       c.cfg.Scopes,
	}
}

// clientSecret reads the runtime secret file per exchange (rotation takes
// effect on the next login). Empty or whitespace-only fails closed — the
// same class the gates learned the hard way (a trailing newline is trimmed
// here; the consumer must never authenticate as secret+newline).
func (c *Client) clientSecret() (string, error) {
	raw, err := os.ReadFile(c.cfg.ClientSecretFile)
	if err != nil {
		return "", errorfamily.WrapRejection(err, "sso.client_secret.read",
			"read the OIDC client secret file (check the secrets pipeline and the 640 root:webphone grant)")
	}
	secret := strings.TrimSpace(string(raw))
	if secret == "" {
		return "", errorfamily.NewRejection("sso.client_secret.empty", "the OIDC client secret file is empty")
	}
	return secret, nil
}

// providerFor lazily discovers the provider and caches the success. A
// failed discovery is NOT cached: an IdP that was down for one login must
// be retried on the next, not remembered forever.
func (c *Client) providerFor(ctx context.Context) (*oidc.Provider, error) {
	c.mu.Lock()
	if c.provider != nil && time.Since(c.fetchedAt) < providerTTL {
		p := c.provider
		c.mu.Unlock()
		return p, nil
	}
	if c.inflight != nil {
		ch := c.inflight
		c.mu.Unlock()
		res := <-ch
		return res.p, res.err
	}
	ch := make(chan providerResult, 1)
	c.inflight = ch
	c.mu.Unlock()

	dctx, cancel := context.WithTimeout(ctx, discoveryTimeout)
	provider, err := oidc.NewProvider(dctx, c.cfg.Issuer)
	cancel()

	c.mu.Lock()
	c.inflight = nil
	if err == nil {
		c.provider = provider
		c.fetchedAt = time.Now()
	}
	c.mu.Unlock()
	ch <- providerResult{p: provider, err: err}
	close(ch)
	if err != nil {
		return nil, errorfamily.WrapRejection(err, "sso.discovery",
			"discover the OIDC provider (check auth.oauth2.issuer and that the IdP is reachable)")
	}
	return provider, nil
}

func randomToken() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("crypto/rand: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

func pkceChallenge(verifier string) string {
	sum := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}
