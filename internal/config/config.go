// Package config loads the webphone runtime configuration from an optional
// JSON file and WEBPHONE_-prefixed environment variables.
package config

import (
	"fmt"
	"maps"
	"net"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/larsartmann/go-error-family"
	"github.com/larsartmann/webphone/internal/domain"

	"github.com/knadh/koanf/parsers/json"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

// GatewayMode selects how outbound messages and faxes leave the server.
type GatewayMode string

const (
	// GatewayLoopback accepts everything instantly and marks it sent —
	// the zero-dependency development mode.
	GatewayLoopback GatewayMode = "loopback"
	// GatewayWebhook forwards outbound traffic to a provider URL.
	GatewayWebhook GatewayMode = "webhook"
)

// Config is the fully-resolved runtime configuration.
// branching-flow large-struct (baselined): THE runtime-config aggregate
// — one field per koanf/__-env section; splitting shatters the
// documented config surface (AGENTS.md).
type Config struct {
	Addr          string `json:"addr" koanf:"addr"`
	DataDir       string `json:"data_dir" koanf:"data_dir"`
	SIPDomain     string `json:"sip_domain" koanf:"sip_domain"`
	WebsocketPath string `json:"websocket_path" koanf:"websocket_path"`
	PhoneAPIURL   string `json:"phone_api_url" koanf:"phone_api_url"`
	// SessionTTL is the SLIDING idle window: activity past its halfway
	// point extends the session back to the full window, so a device in
	// regular use never re-signs-in (the island resumes from the cookie).
	SessionTTL time.Duration `json:"session_ttl" koanf:"session_ttl"`
	// SessionMaxTTL caps a session's ABSOLUTE lifetime from sign-in. The
	// idle window renews; this bound never does — it is what expires even
	// a continuously used (or stolen) session and re-signs the device in.
	SessionMaxTTL time.Duration          `json:"session_max_ttl" koanf:"session_max_ttl"`
	ICEServers    []ICEServer            `json:"ice_servers" koanf:"ice_servers"`
	Contacts      []domain.SharedContact `json:"contacts" koanf:"contacts"`
	// Identities maps an extension to the PSTN number (DID) its calls and
	// messages present to the outside world — the own-number feed (DECIDED
	// 2026-09-20: static map now, stack identity endpoint as the upgrade
	// path). Display-only: the value renders verbatim, so owner-formatted
	// numbers ("+49 30 12345678") survive; keys must be normalized
	// extensions or lookups by the signed-in extension silently miss.
	Identities map[string]string `json:"identities" koanf:"identities"`
	Gateway    Gateway           `json:"gateway" koanf:"gateway"`
	TURN       TURNREST          `json:"turn_rest" koanf:"turn_rest"`
	CSRF       CSRF              `json:"csrf" koanf:"csrf"`
	// CRM is the OPTIONAL integration with the Ledger CRM (contact name
	// enrichment + call-activity logging). Zero value = disabled.
	CRM CRM `json:"crm" koanf:"crm"`
	// Paperless is the OPTIONAL downstream archive for inbound faxes
	// (Paperless-ngx via go-paperless). Zero value = disabled: faxes stay
	// blob-store-only, which is and remains the storage truth.
	Paperless Paperless `json:"paperless" koanf:"paperless"`
	// ASR is the OPTIONAL speech-to-text seam behind live transcription
	// of every audio stream (calls, voicemail, MMS audio). Zero value =
	// disabled: no /api/transcribe route exists and the island hides
	// every transcribe affordance.
	ASR ASR `json:"asr" koanf:"asr"`
	// TextAgent is the OPTIONAL seam to the stack's supervised
	// text-message agent (services.telephony.textagent review API):
	// the Messages-tab draft panel and the compose "draft with AI"
	// affordance. Zero value = disabled: neither surface renders.
	TextAgent TextAgent `json:"textagent" koanf:"textagent"`
	// Matrix is the OPTIONAL Matrix chat-network lane: a thin
	// Client-Server API client (send + sync receiver) behind the
	// matrix config gate. Zero value = disabled: no sender registers
	// for the matrix network and the product stays byte-identical.
	Matrix Matrix `json:"matrix" koanf:"matrix"`
	// Signal is the OPTIONAL Signal chat-network lane: a JSON-RPC
	// client over a local signal-cli daemon unix socket. Zero value =
	// disabled.
	Signal Signal `json:"signal" koanf:"signal"`
	// Push is the OPTIONAL push seam behind ring-while-closed: the
	// registered-device registry rides readmodels unconditionally, but
	// no transport is configured with a zero value — the device API
	// 404s and /hooks/ring acknowledges inertly. Zero value = disabled.
	Push Push `json:"push" koanf:"push"`
	// RetentionDays bounds how long stored content lives: 0 (default)
	// keeps everything forever; a positive value makes the daily sweep
	// delete messages (with attachments), fax jobs (with documents) and
	// empty threads older than that many days. Sessions always expire
	// on their own TTL — the sweep only clears their dead rows earlier.
	RetentionDays int `json:"retention_days" koanf:"retention_days"`
	// Timezone is an IANA zone name (e.g. "Europe/Berlin") that owns
	// every wall-clock rendering and log line. Empty keeps the process
	// local zone. A typo fails validation — times silently rendering in
	// UTC would be worse than a failed boot.
	Timezone string `json:"timezone" koanf:"timezone"`
	// Dashboard mounts the go-health-dashboard operator surface at
	// /health (live check cards over SSE). Zero value = disabled: no
	// /health route exists and the styled 404 answers — the default,
	// because the dashboard is an OPT-IN operator surface the deployment
	// must deliberately expose (and fence) like the probe triple.
	Dashboard Dashboard `json:"dashboard" koanf:"dashboard"`
	// Auth configures the optional passkey (WebAuthn) login mode backed
	// by the embedded cqrs-htmx/usermgmt identity layer. Zero value =
	// disabled: extension + SIP directory password stays the only login
	// and the login page renders exactly as before.
	Auth Auth `json:"auth" koanf:"auth"`
	// SIP configures the derived SIP-credential seam (credential
	// brokering). Zero value = disabled: the island keeps receiving the
	// directory password exactly as before — every surface stays
	// byte-identical until the operator sets a secret.
	SIP SIP `json:"sip" koanf:"sip"`
}

// SIP configures the derived SIP credentials: while CredentialSecret is
// set, the island's REGISTER uses a session-derived credential
// (HMAC-SHA256 over the session token) and FreeSWITCH validates it
// dynamically via the internal directory endpoint — the directory
// password never reaches the browser. Zero value = the historical
// password-wire behavior.
type SIP struct {
	// CredentialSecret is the dedicated derivation secret (env
	// WEBPHONE_SIP__CREDENTIAL_SECRET or config sip.credential_secret).
	// Never reuse the gateway webhook secret: two consumers of one
	// secret cannot rotate independently.
	CredentialSecret string `json:"credential_secret,omitempty" koanf:"credential_secret"`
	// CredentialSecretFile reads the secret from a runtime file (single
	// line, trimmed) instead — the deployment shape the module's
	// secrets pipeline prefers. At most one of the two sources.
	CredentialSecretFile string `json:"credential_secret_file,omitempty" koanf:"credential_secret_file"`
}

// DerivedCredentials reports whether the sip-credential seam is on: a
// non-empty secret (after file resolution) enables derivation, the
// internal directory endpoint, and the island's credential swap.
func (s SIP) DerivedCredentials() bool { return s.CredentialSecret != "" }

// Dashboard configures the optional go-health-dashboard mount.
// Enabled serves /health (HTML by default, JSON via Accept), /health/sse
// and the embedded Datastar SDK; the page is CSP-strict (nonce'd inline
// scripts, same-origin assets only).
type Dashboard struct {
	Enable bool `json:"enable" koanf:"enable"`
	// Title names the deployment on the dashboard page (e.g. the PBX
	// host). Empty keeps the library default.
	Title string `json:"title,omitempty" koanf:"title"`
}

// Auth selects the login modes. The extension login is ALWAYS available
// (lifeline property: signing into a phone must never depend on more
// than FreeSWITCH and the webphone process itself); the passkey and
// OIDC/SSO modes are strictly additive on top of it.
type Auth struct {
	// Users maps an account email to the extensions it may sign in as —
	// shared by every passwordless mode (passkey and OIDC), so one
	// operator edit governs both. auth.passkey.users stays supported as
	// the historical home; the two must never disagree (Load rejects a
	// divergent pair, and folds the shared map down into the passkey
	// shape so exactly one mapping reaches the runtime).
	Users map[string]PasskeyUser `json:"users,omitempty" koanf:"users"`
	// ExtensionPasswordFiles maps a normalized extension to the runtime
	// file carrying its SIP directory password (shared like Users).
	ExtensionPasswordFiles map[string]string `json:"extension_password_files,omitempty" koanf:"extension_password_files"`
	Passkey                Passkey           `json:"passkey" koanf:"passkey"`
	OAuth2                 OAuth2            `json:"oauth2" koanf:"oauth2"`
}

// OAuth2 configures the OpenID Connect login mode (Pocket ID for the PBX
// stack): authorization-code + PKCE in the browser, a server-side token
// exchange, then the same mapping + directory verification + session
// minting every other login mode rides.
//
// Enabled is DERIVED: any field being set counts, and validation fails
// closed on half a configuration. Discovery is lazy, so an unreachable
// IdP breaks logins (with an honest error) but never the boot.
type OAuth2 struct {
	// Issuer is the provider's discovery base URL (e.g.
	// https://auth.example.org). The ID token's issuer claim must equal
	// it exactly.
	Issuer string `json:"issuer" koanf:"issuer"`
	// ClientID is the registered OIDC client id (webphone).
	ClientID string `json:"client_id" koanf:"client_id"`
	// ClientSecretFile is the runtime file holding the confidential
	// client's secret (single line, trimmed, read per exchange so
	// rotation needs no restart). Empty or missing fails CLOSED.
	ClientSecretFile string `json:"client_secret_file" koanf:"client_secret_file"`
	// RedirectURL is the registered callback, absolute
	// (https://<host>/api/auth/oauth2/callback). Must match verbatim.
	RedirectURL string `json:"redirect_url" koanf:"redirect_url"`
	// Scopes defaults to openid/email/profile; "openid" is mandatory.
	Scopes []string `json:"scopes,omitempty" koanf:"scopes"`
	// AllowedGroups is the in-app IdP group allowlist. Empty (the
	// default) disables the check entirely — admission stays with the
	// users mapping and the IdP-side client-group link. Non-empty, every
	// login must carry at least one of these groups in its token claims;
	// a login without any of them is denied with a named reason before
	// the mapping is consulted. This is a second gate on top of the IdP
	// link, not a replacement: the link governs who the IdP will even
	// authenticate for this client.
	AllowedGroups []string `json:"allowed_groups,omitempty" koanf:"allowed_groups"`
}

// Enabled reports whether the OIDC mode is configured at all.
func (o OAuth2) Enabled() bool {
	return o.Issuer != "" || o.ClientID != "" || o.ClientSecretFile != "" || o.RedirectURL != "" || len(o.AllowedGroups) > 0
}

// resolvedAuth folds the shared mapping into the passkey shape so the
// runtime has exactly one mapping. A divergent pair is a config bug (two
// sources of truth for who may sign in), never a silent winner.
func resolvedAuth(a Auth) (Auth, error) {
	if len(a.Users) == 0 {
		return a, nil
	}
	if len(a.Passkey.Users) > 0 && !sameUserMap(a.Users, a.Passkey.Users) {
		return a, errorfamily.NewRejection("config.auth.users",
			"auth.users and auth.passkey.users disagree: declare the email→extension mapping ONCE (auth.users is the shared home)")
	}
	if len(a.ExtensionPasswordFiles) > 0 {
		for ext, file := range a.ExtensionPasswordFiles {
			if current, ok := a.Passkey.ExtensionPasswordFiles[ext]; ok && current != file {
				return a, errorfamily.Newf(errorfamily.Rejection, "config.auth.extension_password_files",
					"auth.extension_password_files[%q] (%s) disagrees with auth.passkey.extension_password_files (%s): declare the password files ONCE", ext, file, current)
			}
		}
	}
	a.Passkey.Users = a.Users
	merged := make(map[string]string, len(a.Passkey.ExtensionPasswordFiles)+len(a.ExtensionPasswordFiles))
	maps.Copy(merged, a.Passkey.ExtensionPasswordFiles)
	maps.Copy(merged, a.ExtensionPasswordFiles)
	a.Passkey.ExtensionPasswordFiles = merged
	return a, nil
}

func sameUserMap(a, b map[string]PasskeyUser) bool {
	if len(a) != len(b) {
		return false
	}
	for email, user := range a {
		other, ok := b[email]
		if !ok || other.DisplayName != user.DisplayName || len(other.Extensions) != len(user.Extensions) {
			return false
		}
		for i, ext := range user.Extensions {
			if other.Extensions[i] != ext {
				return false
			}
		}
	}
	return true
}

// Passkey enables passkey logins via the embedded usermgmt service.
// Enabled is DERIVED: any of RPID, RPOrigins or Users being set counts as
// enabled, and validation fails closed on half a configuration — the
// CRM/Paperless both-or-neither doctrine.
// branching-flow COMPOSITION_mixin (baselined): config mirror of
// userauth.PasskeyRuntime; a mixin would make userauth import config —
// forbidden (userauth imports ONLY config-free types).
type Passkey struct {
	// RPID is the WebAuthn Relying Party ID: the registrable domain the
	// passkeys are bound to (e.g. "pbx.example.org"). Must equal the
	// host of every entry in RPOrigins.
	RPID string `json:"rp_id" koanf:"rp_id"`
	// RPDisplayName names the relying party in the browser's passkey
	// prompt. Empty keeps "webphone".
	RPDisplayName string `json:"rp_display_name,omitempty" koanf:"rp_display_name"`
	// RPOrigins lists the absolute origins the ceremony accepts (e.g.
	// "https://pbx.example.org") — the same operator-declared shape as
	// csrf.trusted_origins, because the app behind a TLS proxy cannot
	// derive its public origin itself.
	RPOrigins []string `json:"rp_origins" koanf:"rp_origins"`
	// Users maps a usermgmt account email to the extensions that account
	// may sign in as. The FIRST extension is the one the session binds
	// (all stores are extension-scoped); every listed extension's DID
	// (config identities) renders in the whoami line. The map is the
	// admin surface, exactly like identities — changing it is a config
	// edit, never a UI flow.
	Users map[string]PasskeyUser `json:"users" koanf:"users"`
	// ExtensionPasswordFiles maps a normalized extension to a runtime
	// file carrying its SIP directory password (single line). Passkey
	// login never asks the user for the password, but the session row
	// must carry it (the island REGISTER and the /phone-api proxy are
	// password-based by design), so the server sources it the same way
	// the bridge sources its secrets: an operator-managed file, read at
	// login, never cached. Every extension listed in Users needs one.
	ExtensionPasswordFiles map[string]string `json:"extension_password_files" koanf:"extension_password_files"`
}

// PasskeyUser is one email's extension mapping (see Passkey.Users).
// branching-flow COMPOSITION_mixin (baselined): config mirror of
// userauth.MappedUser (the Passkey-mirror class; see Passkey).
type PasskeyUser struct {
	Extensions  []string `json:"extensions" koanf:"extensions"`
	DisplayName string   `json:"display_name,omitempty" koanf:"display_name"`
}

// Enabled reports whether the passkey mode is configured. Any set field
// counts: validation enforces the rest, so a half configuration fails
// the boot instead of silently running half a login mode.
func (p Passkey) Enabled() bool {
	return p.RPID != "" || len(p.RPOrigins) > 0 || len(p.Users) > 0
}

// CRM configures the optional Ledger CRM integration. Both fields must be
// set together (a URL without a token cannot authenticate; a token without
// a URL has nowhere to go) — config validation fails closed on half a
// configuration so a typo can never silently disable or enable the client.
// branching-flow DUPLICATE_TYPE+COMPOSITION_mixin (baselined):
// independent seam twin of Paperless by owner ruling 2026-10-11 —
// URL+Token stays per-seam both-or-neither; detector noise routed
// upstream.
type CRM struct {
	URL   string `json:"url" koanf:"url"`
	Token string `json:"token" koanf:"token"`
}

// Paperless configures the optional Paperless-ngx archive for inbound
// faxes. Both fields must be set together, exactly like the CRM seam —
// half a configuration would mint a client that cannot authenticate (or
// one with nowhere to go), and config validation fails closed on that.
// branching-flow DUPLICATE_TYPE (baselined): independent seam twin of
// CRM by owner ruling 2026-10-11 (see CRM).
type Paperless struct {
	URL   string `json:"url" koanf:"url"`
	Token string `json:"token" koanf:"token"`
}

// ASR configures the optional speech-to-text provider. Two provider
// kinds share the seam: "openai" (default — any OpenAI-compatible
// /v1/audio/transcriptions server, e.g. Speaches or whisper.cpp) and
// "google" (Google Cloud Speech-to-Text V2, EU-resident regional
// endpoint). Unlike the CRM/Paperless both-or-neither pairs, a token
// alone is MEANINGLESS on the openai wire (a credential with nowhere to
// send it) but a URL alone is a COMPLETE, legitimate configuration: a
// self-hosted whisper.cpp server on loopback usually accepts anonymous
// requests. So the rule is asymmetric — URL enables, token is optional,
// and token-without-URL fails closed. The google kind shifts the
// anchor: the project id is what the recognizer path needs, so project
// enables and location/token stay optional overrides.
type ASR struct {
	// Provider selects the wire: "" or "openai" = OpenAI-compatible
	// multipart; "google" = Google Cloud Speech-to-Text V2 JSON.
	Provider string `json:"provider,omitempty" koanf:"provider"`
	// URL is the provider base (e.g. http://127.0.0.1:8081) on the
	// openai wire; an optional endpoint override on the google wire.
	// Empty = disabled on openai, derived regional endpoint on google.
	URL string `json:"url" koanf:"url"`
	// Token is the bearer credential on the openai wire, the
	// x-goog-api-key on the google wire; sent only when set.
	Token string `json:"token,omitempty" koanf:"token"`
	// Model names the transcription model: whisper-style model id on
	// the openai wire (e.g. "Systran/faster-whisper-large-v3-turbo"),
	// the V2 model variant on google (default "telephony"). Empty
	// keeps the client default.
	Model string `json:"model,omitempty" koanf:"model"`
	// Project is the Google Cloud project id (google wire only, part
	// of the recognizer resource path — required when provider is
	// google).
	Project string `json:"project,omitempty" koanf:"project"`
	// Location is the Google region (google wire only). Empty keeps
	// europe-west3: the regional endpoint is what keeps caller audio
	// inside the EU.
	Location string `json:"location,omitempty" koanf:"location"`
	// Language is the fallback BCP-47 code for untagged requests
	// (google wire only — V2 has no auto-detect). Empty keeps de-DE.
	Language string `json:"language,omitempty" koanf:"language"`
}

// Google reports whether the google provider kind is selected.
func (a ASR) Google() bool { return a.Provider == "google" }

// TextAgent configures the optional supervised text-agent seam (the
// stack's services.telephony.textagent review API on loopback). URL
// and SecretFile are a both-or-neither pair, exactly like the CRM and
// Paperless seams: half a configuration would mint a client that
// cannot authenticate (or one with nowhere to go), and validation
// fails closed on that. The stack's telephony module auto-wires both
// from its own options, so a deployment never sets them by hand.
type TextAgent struct {
	// URL is the review-API base (e.g. http://127.0.0.1:8071).
	URL string `json:"url,omitempty" koanf:"url"`
	// SecretFile reads the bearer secret from a runtime file (single
	// line, trimmed) — the SAME file the review API authenticates
	// against; one secret, two consumers, like the gateway pair.
	SecretFile string `json:"secret_file,omitempty" koanf:"secret_file"`
}

// Enabled reports whether the text-agent seam is configured. The
// composition root builds the client only when this is true; nil
// client = no surface renders.
func (t TextAgent) Enabled() bool { return t.URL != "" }

// Matrix configures the optional Matrix chat-network lane. When set,
// ALL four fields are required (fail closed on half configurations —
// the same doctrine as the CRM/Paperless pairs): a client that cannot
// authenticate, or conversations with no owning extension, must fail
// at boot, not at the first message.
type Matrix struct {
	// Homeserver is the Client-Server API base (e.g. https://matrix.example.org).
	Homeserver string `json:"homeserver,omitempty" koanf:"homeserver"`
	// UserID is the bot account (e.g. @webphone:example.org).
	UserID string `json:"user_id,omitempty" koanf:"user_id"`
	// AccessToken authenticates the account; a long-lived token minted
	// once via POST /login (deliver via environmentFile — a secret).
	AccessToken string `json:"access_token,omitempty" koanf:"access_token"`
	// Owner is the webphone extension whose Messages tab owns every
	// matrix conversation (the account serves the deployment; threads
	// are owner-scoped everywhere and need their home).
	Owner string `json:"owner,omitempty" koanf:"owner"`
}

// Enabled reports whether the matrix lane is configured.
func (m Matrix) Enabled() bool { return m.Homeserver != "" }

// Signal configures the optional Signal chat-network lane over a
// signal-cli daemon's unix socket (signal-cli daemon --socket). When
// set, all three fields are required, exactly like Matrix.
type Signal struct {
	// SocketPath is the signal-cli JSON-RPC unix socket (e.g.
	// /run/signal-cli/socket). Filesystem permissions are the trust
	// boundary — the same loopback posture as the gateway.
	SocketPath string `json:"socket_path,omitempty" koanf:"socket_path"`
	// Account is the linked Signal number (+49...) signal-cli sends
	// from; the multi-account daemon needs it to pick the lane.
	Account string `json:"account,omitempty" koanf:"account"`
	// Owner is the webphone extension owning every signal conversation.
	Owner string `json:"owner,omitempty" koanf:"owner"`
}

// Enabled reports whether the signal lane is configured.
func (s Signal) Enabled() bool { return s.SocketPath != "" }

// Enabled reports whether an ASR provider is configured. The openai
// wire enables on URL alone; the google wire on a project id.
// Validation rejects half-configurations before this matters.
func (a ASR) Enabled() bool {
	if a.Google() {
		return a.Project != ""
	}
	return a.URL != ""
}

// ICEServer is one STUN/TURN server entry handed to the browser island.
// The JSON tags ARE the window.PBX_CONFIG wire contract (see README).
type ICEServer struct {
	URLs       []string `json:"urls" koanf:"urls"`
	Username   string   `json:"username,omitempty" koanf:"username"`
	Credential string   `json:"credential,omitempty" koanf:"credential"`
}

// Gateway configures the outbound message/fax gateway.
type Gateway struct {
	Mode          GatewayMode `json:"mode" koanf:"mode"`
	WebhookURL    string      `json:"webhook_url,omitempty" koanf:"webhook_url"`
	WebhookSecret string      `json:"webhook_secret,omitempty" koanf:"webhook_secret"`
	// WebhookSecretFile reads the shared secret from a runtime file at
	// startup (single line, surrounding whitespace trimmed) instead of an
	// env var / config value — a deployment then maintains ONE file that
	// both the gateway and the receiving bridge read, instead of keeping
	// an env copy and the receiver's copy in sync by hand. Exactly one
	// of webhook_secret / webhook_secret_file may be set.
	WebhookSecretFile string `json:"webhook_secret_file,omitempty" koanf:"webhook_secret_file"`
}

// errGatewaySecretBothSources: exactly-one-of gateway.webhook_secret /
// gateway.webhook_secret_file — two sources for one secret invite drift.
// Classified Rejection: the operator must fix
// the deployment's config before the service can start.
var errGatewaySecretBothSources error = errorfamily.NewRejection("config.gateway_secret_sources", "gateway: set at most one of webhook_secret / webhook_secret_file")

// errSIPSecretBothSources: exactly-one-of sip.credential_secret /
// sip.credential_secret_file — two sources for one secret invite drift.
// Classified Rejection: the operator must fix
// the deployment's config before the service can start.
var errSIPSecretBothSources error = errorfamily.NewRejection("config.sip.credential_secret_sources", "sip: set at most one of credential_secret / credential_secret_file")

// TURNREST switches TURN authentication from static config passwords to
// coturn's REST API (draft-uberti-behave-turn-rest): while the secret is
// empty, ice_servers entries ship verbatim; once set, /config.js derives a
// short-lived username/credential pair (username = unix expiry,
// credential = base64(HMAC-SHA1(secret, username))) for every ice_servers
// entry that carries a turn:/turns: URL, so a long-lived TURN password
// never sits in config, the browser, or a cached response. The SAME
// secret must ride coturn (use-auth-secret / static-auth-secret).
type TURNREST struct {
	Secret string        `json:"secret" koanf:"secret"`
	TTL    time.Duration `json:"ttl" koanf:"ttl"`
}

// CSRF teaches the double-submit CSRF middleware about the fronting
// deployment shape. The webphone listener speaks plain HTTP while nginx
// terminates TLS, so a truthful browser request carries
// Origin: https://host — which the middleware, judging from its own
// plain-HTTP view, would call a forged same-origin attestation and reject
// (403 on every POST, logins included). TrustedProxies names the local
// proxies whose X-Forwarded-Proto may be believed (loopback nginx);
// TrustedOrigins lists the browser-facing origins that count as
// same-origin. Both empty is correct for direct (unfronted) exposure.
type CSRF struct {
	TrustedProxies []string `json:"trusted_proxies,omitempty" koanf:"trusted_proxies"`
	TrustedOrigins []string `json:"trusted_origins,omitempty" koanf:"trusted_origins"`
}

// defaults keeps the zero-config path working: loopback gateway, local
// ports, /var/lib/webphone data dir.
func defaults() Config {
	return Config{
		Addr:          ":8080",
		DataDir:       "/var/lib/webphone",
		WebsocketPath: "/sip",
		SessionTTL:    7 * 24 * time.Hour,
		SessionMaxTTL: 30 * 24 * time.Hour,
		Gateway:       Gateway{Mode: GatewayLoopback},
	}
}

// Load reads the optional JSON config file (path from WEBPHONE_CONFIG,
// default /etc/webphone/config.json), then overrides from WEBPHONE_*
// environment variables (double underscore nests: WEBPHONE_GATEWAY__MODE
// → gateway.mode; single underscores stay literal: WEBPHONE_DATA_DIR →
// data_dir).
func Load() (Config, error) {
	k := koanf.New(".")

	cfg := defaults()
	if err := k.Load(confmap.Provider(map[string]any{
		"addr":           cfg.Addr,
		"data_dir":       cfg.DataDir,
		"websocket_path": cfg.WebsocketPath,
		"session_ttl":    cfg.SessionTTL.String(),
		"gateway.mode":   string(cfg.Gateway.Mode),
		"turn_rest.ttl":  (48 * time.Hour).String(),
	}, "."), nil); err != nil {
		return Config{}, errorfamily.WrapOrchestrationf(err, "config.defaults", "config defaults")
	}

	path := os.Getenv("WEBPHONE_CONFIG")
	if path == "" {
		path = "/etc/webphone/config.json"
	}
	if _, err := os.Stat(path); err == nil {
		if err := k.Load(file.Provider(path), json.Parser()); err != nil {
			return Config{}, errorfamily.WrapRejectionf(err, "config.file", "config file %s", path)
		}
	}

	if err := k.Load(env.Provider("WEBPHONE_", ".", func(s string) string {
		return envKeyToPath(s)
	}), nil); err != nil {
		return Config{}, errorfamily.WrapRejectionf(err, "config.env", "config env")
	}

	if err := k.Unmarshal("", &cfg); err != nil {
		return Config{}, errorfamily.WrapRejectionf(err, "config.unmarshal", "config unmarshal")
	}

	if cfg.Gateway.WebhookSecretFile != "" && cfg.Gateway.WebhookSecret != "" {
		// Shape check BEFORE the file read: both sources set is a config
		// drift bug regardless of whether the file is readable.
		return Config{}, fmt.Errorf("config invalid: %w", errGatewaySecretBothSources) //nolint:erraudit // family-neutral propagation: the inner error owns the family
	}
	if cfg.Gateway.WebhookSecretFile != "" {
		data, err := os.ReadFile(cfg.Gateway.WebhookSecretFile)
		if err != nil {
			return Config{}, errorfamily.WrapRejectionf(err, "config.webhook_secret_file", "gateway.webhook_secret_file %s", cfg.Gateway.WebhookSecretFile)
		}
		secret := strings.TrimSpace(string(data))
		if secret == "" {
			return Config{}, errorfamily.Newf(errorfamily.Rejection, "config.webhook_secret_file", "gateway.webhook_secret_file %s is empty", cfg.Gateway.WebhookSecretFile)
		}
		cfg.Gateway.WebhookSecret = secret
	}

	// sip.credential_secret / sip.credential_secret_file: at-most-one-of,
	// same drift argument as the gateway secret pair.
	if cfg.SIP.CredentialSecretFile != "" && cfg.SIP.CredentialSecret != "" {
		// Shape check BEFORE the file read: both sources set is a config
		// drift bug regardless of whether the file is readable.
		return Config{}, fmt.Errorf("config invalid: %w", errSIPSecretBothSources) //nolint:erraudit // family-neutral propagation: the inner error owns the family
	}
	if cfg.SIP.CredentialSecretFile != "" {
		data, err := os.ReadFile(cfg.SIP.CredentialSecretFile)
		if err != nil {
			return Config{}, errorfamily.WrapRejectionf(err, "config.sip.credential_secret_file", "sip.credential_secret_file %s", cfg.SIP.CredentialSecretFile)
		}
		secret := strings.TrimSpace(string(data))
		if secret == "" {
			return Config{}, errorfamily.Newf(errorfamily.Rejection, "config.sip.credential_secret_file", "sip.credential_secret_file %s is empty", cfg.SIP.CredentialSecretFile)
		}
		cfg.SIP.CredentialSecret = secret
	}

	auth, err := resolvedAuth(cfg.Auth)
	if err != nil {
		return Config{}, fmt.Errorf("config invalid: %w", err) //nolint:erraudit // family-neutral propagation: the inner error owns the family
	}
	cfg.Auth = auth

	if err := validate(cfg); err != nil {
		return Config{}, fmt.Errorf("config invalid: %w", err) //nolint:erraudit // family-neutral propagation: the inner error owns the family
	}

	return cfg, nil
}

// envKeyToPath maps env names to config paths: WEBPHONE_GATEWAY__MODE →
// gateway.mode (double underscore nests, single underscores stay literal —
// WEBPHONE_DATA_DIR → data_dir).
func envKeyToPath(key string) string {
	lowered := strings.ToLower(key)
	lowered = strings.TrimPrefix(lowered, "webphone_")
	if lowered == "" {
		return ""
	}
	return strings.ReplaceAll(lowered, "__", ".")
}

func validate(cfg Config) error {
	switch cfg.Gateway.Mode {
	case GatewayLoopback, GatewayWebhook:
	case "":
		return errorfamily.NewRejection("config.gateway.mode", "gateway.mode is empty")
	default:
		return errorfamily.Newf(errorfamily.Rejection, "config.gateway.mode", "gateway.mode %q is not one of loopback|webhook", cfg.Gateway.Mode)
	}
	if cfg.Gateway.Mode == GatewayWebhook && cfg.Gateway.WebhookURL == "" {
		return errorfamily.NewRejection("config.gateway.webhook_url", "gateway.webhook_url is required in webhook mode")
	}
	if cfg.Addr == "" {
		return errorfamily.NewRejection("config.addr", "addr is empty")
	}
	if cfg.DataDir == "" {
		return errorfamily.NewRejection("config.data_dir", "data_dir is empty")
	}
	if cfg.SessionTTL <= 0 {
		return errorfamily.NewRejection("config.session_ttl", "session_ttl must be positive")
	}
	if cfg.RetentionDays < 0 {
		return errorfamily.Newf(errorfamily.Rejection, "config.retention_days", "retention_days must not be negative (0 = keep forever, N = delete content older than N days)")
	}
	if cfg.Timezone != "" {
		if _, err := time.LoadLocation(cfg.Timezone); err != nil {
			return errorfamily.WrapRejectionf(err, "config.timezone", "timezone %q is not an IANA zone name", cfg.Timezone)
		}
	}
	if cfg.TURN.Secret != "" {
		if cfg.TURN.TTL <= 0 {
			return errorfamily.NewRejection("config.turn_rest.ttl", "turn_rest.ttl must be positive when turn_rest.secret is set")
		}
		hasTurnURL := false
		for _, server := range cfg.ICEServers {
			for _, raw := range server.URLs {
				u, err := url.Parse(raw)
				if err == nil && (u.Scheme == "turn" || u.Scheme == "turns") {
					hasTurnURL = true
				}
			}
		}
		if !hasTurnURL {
			return errorfamily.NewRejection("config.turn_rest.dead_secret", "turn_rest.secret is set but no ice_servers entry carries a turn:/turns: URL: the secret would be dead config (coturn never sees a derived credential)")
		}
	}
	if cfg.SessionMaxTTL < cfg.SessionTTL {
		return errorfamily.Newf(errorfamily.Rejection, "config.session_max_ttl", "session_max_ttl (%s) must be >= session_ttl (%s): the absolute cap cannot be shorter than the idle window it bounds", cfg.SessionMaxTTL, cfg.SessionTTL)
	}
	for _, origin := range cfg.CSRF.TrustedOrigins {
		u, err := url.Parse(origin)
		if err != nil || u.Scheme == "" || u.Host == "" {
			return errorfamily.Newf(errorfamily.Rejection, "config.csrf.trusted_origins", "csrf.trusted_origins: %q is not an absolute origin (want scheme://host, e.g. https://pbx.example.org)", origin)
		}
	}
	for _, proxy := range cfg.CSRF.TrustedProxies {
		if net.ParseIP(proxy) == nil {
			if _, _, err := net.ParseCIDR(proxy); err != nil {
				return errorfamily.Newf(errorfamily.Rejection, "config.csrf.trusted_proxies", "csrf.trusted_proxies: %q is not an IP address or CIDR network", proxy)
			}
		}
	}
	for ext, did := range cfg.Identities {
		parsed, err := domain.ParseExtension(ext)
		if err != nil || parsed.String() != ext {
			return errorfamily.Newf(errorfamily.Rejection, "config.identities", "identities: key %q is not a normalized extension (sanitize it exactly as it appears after login)", ext)
		}
		if _, err := domain.ParsePhone(did); err != nil {
			return errorfamily.Newf(errorfamily.Rejection, "config.identities", "identities: DID for extension %q has no dialable characters", ext)
		}
	}
	for _, contact := range cfg.Contacts {
		if _, err := domain.ParsePhone(contact.Number); err != nil {
			return errorfamily.Newf(errorfamily.Rejection, "config.contacts", "contacts: shared contact %q has no dialable characters in %q (fix the config file — a silently un-dialable directory entry is worse than a failed boot)", contact.Name, contact.Number)
		}
	}
	if cfg.Paperless.URL != "" || cfg.Paperless.Token != "" {
		switch {
		case cfg.Paperless.URL == "":
			return errorfamily.NewRejection("config.paperless.url", "paperless.token is set without paperless.url: the integration needs both (or neither)")
		case cfg.Paperless.Token == "":
			return errorfamily.NewRejection("config.paperless.token", "paperless.url is set without paperless.token: the archive API rejects anonymous uploads")
		}
		if u, err := url.Parse(cfg.Paperless.URL); err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
			return errorfamily.Newf(errorfamily.Rejection, "config.paperless.url", "paperless.url %q is not an absolute http(s) URL (e.g. http://127.0.0.1:2280)", cfg.Paperless.URL)
		}
	}
	if err := validatePasskey(cfg.Auth.Passkey); err != nil {
		return err
	}
	if err := validateOAuth2(cfg.Auth.OAuth2); err != nil {
		return err
	}
	switch {
	case cfg.TextAgent.URL == "" && cfg.TextAgent.SecretFile == "":
	case cfg.TextAgent.URL == "":
		return errorfamily.NewRejection("config.textagent.url", "textagent.secret_file is set without textagent.url: the seam needs both (or neither)")
	case cfg.TextAgent.SecretFile == "":
		return errorfamily.NewRejection("config.textagent.secret_file", "textagent.url is set without textagent.secret_file: the review API rejects anonymous callers")
	}
	if cfg.TextAgent.URL != "" {
		u, err := url.Parse(cfg.TextAgent.URL)
		if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
			return errorfamily.Newf(errorfamily.Rejection, "config.textagent.url", "textagent.url %q is not an absolute http(s) URL (e.g. http://127.0.0.1:8071)", cfg.TextAgent.URL)
		}
	}
	if err := validateMatrix(cfg.Matrix); err != nil {
		return err
	}
	if err := validateSignal(cfg.Signal); err != nil {
		return err
	}
	switch {
	case cfg.ASR.Provider != "" && cfg.ASR.Provider != "openai" && cfg.ASR.Provider != "google":
		return errorfamily.Newf(errorfamily.Rejection, "config.asr.provider", "asr.provider %q is not one of \"openai\", \"google\"", cfg.ASR.Provider)
	case cfg.ASR.Google() && cfg.ASR.Project == "":
		return errorfamily.NewRejection("config.asr.project", "asr.provider is google without asr.project: the recognizer path needs the cloud project id")
	case cfg.ASR.Google() && (strings.ContainsAny(cfg.ASR.Project, "/ ") || strings.ContainsAny(cfg.ASR.Location, "/ ")):
		return errorfamily.NewRejection("config.asr.project", "asr.project/asr.location must be a bare id, not a path (no slashes or spaces)")
	case !cfg.ASR.Google() && cfg.ASR.URL == "" && cfg.ASR.Token != "":
		return errorfamily.NewRejection("config.asr.url", "asr.token is set without asr.url: a credential with nowhere to send it (set asr.url, or clear the token for an unauthenticated provider)")
	}
	if cfg.ASR.URL != "" {
		u, err := url.Parse(cfg.ASR.URL)
		if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
			return errorfamily.Newf(errorfamily.Rejection, "config.asr.url", "asr.url %q is not an absolute http(s) URL (e.g. http://127.0.0.1:8081)", cfg.ASR.URL)
		}
	}
	switch {
	case cfg.CRM.URL == "" && cfg.CRM.Token == "":
		// Both unset: nothing to check — the tail validates the next seam.
	case cfg.CRM.URL == "":
		return errorfamily.NewRejection("config.crm.url", "crm.token is set without crm.url: the integration needs both (or neither)")
	case cfg.CRM.Token == "":
		return errorfamily.NewRejection("config.crm.token", "crm.url is set without crm.token: the CRM machine API answers 401 without a bearer token")
	default:
		u, err := url.Parse(cfg.CRM.URL)
		if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
			return errorfamily.Newf(errorfamily.Rejection, "config.crm.url", "crm.url %q is not an absolute http(s) URL (e.g. http://127.0.0.1:8080)", cfg.CRM.URL)
		}
	}
	return validatePush(cfg.Push)
}

// Push configures the optional push seam (ring-while-closed). The
// two transports are independent — either, both or neither may be
// configured — but each fails closed on its own half states: an ntfy
// token without a URL is a credential with nowhere to send it, and a
// malformed FCM credentials file refuses the boot (parsed in the
// composition root, not here — the file's EXISTENCE is runtime
// state, its path shape is config).
type Push struct {
	// Ntfy is the self-hosted transport (no Google, the de-Googled
	// device's lane).
	Ntfy PushNtfy `json:"ntfy" koanf:"ntfy"`
	// Fcm is the FCM HTTP v1 transport (service-account credentials,
	// the GMS device's lane).
	Fcm PushFcm `json:"fcm" koanf:"fcm"`
}

// PushNtfy configures the ntfy transport. The URL enables the
// transport; the token is optional (a public self-hosted server
// accepts anonymous publishes) and rides the Bearer header when set.
// The per-device topics live in the registry, never here — a shared
// config topic would ring every device on every call.
type PushNtfy struct {
	// URL is the ntfy server base (e.g. https://ntfy.example.org).
	URL string `json:"url,omitempty" koanf:"url"`
	// Token is the optional access-token credential (env-borne: the
	// environmentFile is the right home for it).
	Token string `json:"token,omitempty" koanf:"token"`
}

// PushFcm configures the FCM HTTP v1 transport.
type PushFcm struct {
	// CredentialsFile is the Google service-account JSON key file
	// (read+parsed at boot; a broken file refuses the boot).
	CredentialsFile string `json:"credentials_file,omitempty" koanf:"credentials_file"`
}

// Enabled reports whether any push transport is configured.
func (p Push) Enabled() bool { return p.Ntfy.Enabled() || p.Fcm.Enabled() }

// Enabled reports whether the ntfy transport is configured.
func (n PushNtfy) Enabled() bool { return n.URL != "" }

// Enabled reports whether the FCM transport is configured.
func (f PushFcm) Enabled() bool { return f.CredentialsFile != "" }

// validatePush rejects the push seam's half configurations: an ntfy
// token with no URL is a credential with nowhere to send it; a
// non-absolute ntfy URL cannot carry a publish.
func validatePush(cfg Push) error {
	switch {
	case cfg.Ntfy.URL == "" && cfg.Ntfy.Token != "":
		return errorfamily.NewRejection("config.push.ntfy.url", "push.ntfy.token is set without push.ntfy.url: a credential with nowhere to send it (set the url, or clear the token for an anonymous server)")
	}
	if cfg.Ntfy.URL != "" {
		u, err := url.Parse(cfg.Ntfy.URL)
		if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
			return errorfamily.Newf(errorfamily.Rejection, "config.push.ntfy.url", "push.ntfy.url %q is not an absolute http(s) URL (e.g. https://ntfy.example.org)", cfg.Ntfy.URL)
		}
	}
	return nil
}

// validatePasskey fails closed on half a passkey configuration: a login
// mode that is half-wired would either brick passkey logins at runtime or
// silently fall back to extension logins — both worse than a failed boot.
// The password-file existence is deliberately NOT checked here (files
// rotate at runtime); it is enforced at login time, fail-closed.
// validateMatrix fails closed on half configurations: the homeserver,
// the account, its token, and the owning extension are one unit — any
// gap would mint a client that cannot authenticate (or conversations
// with no home), and boot is the honest place to say so.
func validateMatrix(m Matrix) error {
	if !m.Enabled() {
		return nil
	}
	switch {
	case m.UserID == "":
		return errorfamily.NewRejection("config.matrix.user_id", "matrix.homeserver is set without matrix.user_id: the lane needs all of homeserver, user_id, access_token, owner")
	case m.AccessToken == "":
		return errorfamily.NewRejection("config.matrix.access_token", "matrix.homeserver is set without matrix.access_token: the homeserver rejects anonymous clients (deliver the token via environmentFile)")
	case m.Owner == "":
		return errorfamily.NewRejection("config.matrix.owner", "matrix.homeserver is set without matrix.owner: matrix conversations need the extension whose Messages tab owns them")
	}
	if u, err := url.Parse(m.Homeserver); err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return errorfamily.Newf(errorfamily.Rejection, "config.matrix.homeserver", "matrix.homeserver %q is not an absolute http(s) URL (e.g. https://matrix.example.org)", m.Homeserver)
	}
	if _, err := domain.ParseAddress(domain.NetworkMatrix, m.UserID); err != nil {
		return errorfamily.Newf(errorfamily.Rejection, "config.matrix.user_id", "matrix.user_id %q is not a Matrix user id (@localpart:server)", m.UserID)
	}
	if _, err := domain.ParseExtension(m.Owner); err != nil {
		return errorfamily.Newf(errorfamily.Rejection, "config.matrix.owner", "matrix.owner %q is not a valid extension", m.Owner)
	}
	return nil
}

// validateSignal mirrors validateMatrix for the signal-cli lane: the
// socket, the sending account, and the owning extension are one unit.
func validateSignal(s Signal) error {
	if !s.Enabled() {
		return nil
	}
	switch {
	case s.Account == "":
		return errorfamily.NewRejection("config.signal.account", "signal.socket_path is set without signal.account: the multi-account daemon needs the number to send from")
	case s.Owner == "":
		return errorfamily.NewRejection("config.signal.owner", "signal.socket_path is set without signal.owner: signal conversations need the extension whose Messages tab owns them")
	}
	if !strings.HasPrefix(s.SocketPath, "/") {
		return errorfamily.Newf(errorfamily.Rejection, "config.signal.socket_path", "signal.socket_path %q is not an absolute path (e.g. /run/signal-cli/socket)", s.SocketPath)
	}
	if _, err := domain.ParseExtension(s.Owner); err != nil {
		return errorfamily.Newf(errorfamily.Rejection, "config.signal.owner", "signal.owner %q is not a valid extension", s.Owner)
	}
	return nil
}

func validatePasskey(p Passkey) error {
	if !p.Enabled() {
		return nil
	}
	if p.RPID == "" {
		return errorfamily.NewRejection("config.auth.passkey.rp_id", "auth.passkey: rp_id is required when the passkey mode is configured (the registrable domain the passkeys bind to, e.g. pbx.example.org)")
	}
	if len(p.RPOrigins) == 0 {
		return errorfamily.NewRejection("config.auth.passkey.rp_origins", "auth.passkey: rp_origins is required (at least one absolute origin, e.g. https://pbx.example.org)")
	}
	for _, origin := range p.RPOrigins {
		u, err := url.Parse(origin)
		if err != nil || u.Scheme != "https" && u.Scheme != "http" || u.Host == "" {
			return errorfamily.Newf(errorfamily.Rejection, "config.auth.passkey.rp_origins", "auth.passkey.rp_origins: %q is not an absolute origin (want scheme://host, e.g. https://pbx.example.org)", origin)
		}
		// WebAuthn binds the RP ID to the origin's DOMAIN — ports are
		// irrelevant (a loopback dev origin carries one), so the honest
		// mismatch test compares hostnames, not host:port.
		if u.Hostname() != p.RPID {
			return errorfamily.Newf(errorfamily.Rejection, "config.auth.passkey.rp_id", "auth.passkey: rp_id %q must equal the host of every rp_origins entry (got %q): a mismatch binds passkeys the browser then refuses to use", p.RPID, u.Hostname())
		}
	}
	if len(p.Users) == 0 {
		return errorfamily.NewRejection("config.auth.passkey.users", "auth.passkey: users is required — without an email→extension mapping no passkey login can mint a session")
	}
	for email, user := range p.Users {
		if !strings.Contains(email, "@") || strings.HasPrefix(email, "@") || strings.HasSuffix(email, "@") {
			return errorfamily.Newf(errorfamily.Rejection, "config.auth.passkey.users", "auth.passkey.users: key %q is not an email address", email)
		}
		if len(user.Extensions) == 0 {
			return errorfamily.Newf(errorfamily.Rejection, "config.auth.passkey.users", "auth.passkey.users: %q has no extensions — a passkey login needs at least one extension to bind its session to", email)
		}
		for _, raw := range user.Extensions {
			parsed, err := domain.ParseExtension(raw)
			if err != nil || parsed.String() != raw {
				return errorfamily.Newf(errorfamily.Rejection, "config.auth.passkey.users", "auth.passkey.users: extension %q of %q is not a normalized extension (sanitize it exactly as it appears in the directory)", raw, email)
			}
			file, ok := p.ExtensionPasswordFiles[raw]
			if !ok || file == "" {
				return errorfamily.Newf(errorfamily.Rejection, "config.auth.passkey.extension_password_files", "auth.passkey.extension_password_files: no password file for extension %q (mapped by %q) — passkey login must source the SIP directory password server-side", raw, email)
			}
		}
	}
	return nil
}

// validateOAuth2 fails closed on half an OIDC configuration: a login mode
// that is half-wired would either brick SSO logins at runtime or silently
// fall back to extension logins — both worse than a failed boot. The
// client-secret FILE is deliberately not read here (it rotates at
// runtime); it is enforced at login time, fail-closed.
func validateOAuth2(o OAuth2) error {
	if !o.Enabled() {
		return nil
	}
	if o.Issuer == "" {
		return errorfamily.NewRejection("config.auth.oauth2.issuer", "auth.oauth2: issuer is required when the OIDC mode is configured (the IdP discovery base URL, e.g. https://auth.example.org)")
	}
	u, err := url.Parse(o.Issuer)
	if err != nil || u.Scheme != "https" || u.Host == "" || u.RawQuery != "" || u.Fragment != "" {
		return errorfamily.Newf(errorfamily.Rejection, "config.auth.oauth2.issuer", "auth.oauth2.issuer %q is not a bare https URL (no query/fragment: the issuer claim must match verbatim)", o.Issuer)
	}
	if o.ClientID == "" {
		return errorfamily.NewRejection("config.auth.oauth2.client_id", "auth.oauth2: client_id is required (the registered OIDC client id)")
	}
	if o.ClientSecretFile == "" {
		return errorfamily.NewRejection("config.auth.oauth2.client_secret_file", "auth.oauth2: client_secret_file is required — the client is confidential and the secret must never live in the config file")
	}
	if o.RedirectURL == "" {
		return errorfamily.NewRejection("config.auth.oauth2.redirect_url", "auth.oauth2: redirect_url is required (the registered callback, e.g. https://pbx.example.org/api/auth/oauth2/callback)")
	}
	ru, err := url.Parse(o.RedirectURL)
	if err != nil || ru.Scheme != "https" || ru.Host == "" {
		return errorfamily.Newf(errorfamily.Rejection, "config.auth.oauth2.redirect_url", "auth.oauth2.redirect_url %q is not an absolute https URL", o.RedirectURL)
	}
	hasOpenID := false
	for _, scope := range o.Scopes {
		if scope == "openid" {
			hasOpenID = true
		}
	}
	if !hasOpenID {
		return errorfamily.NewRejection("config.auth.oauth2.scopes", "auth.oauth2.scopes must include \"openid\" (without it the IdP returns no id_token and no login can complete)")
	}
	return nil
}
