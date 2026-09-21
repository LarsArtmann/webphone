// Package config loads the webphone runtime configuration from an optional
// JSON file and WEBPHONE_-prefixed environment variables.
package config

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"strings"
	"time"

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
type Config struct {
	Addr          string                 `json:"addr" koanf:"addr"`
	DataDir       string                 `json:"data_dir" koanf:"data_dir"`
	SIPDomain     string                 `json:"sip_domain" koanf:"sip_domain"`
	WebsocketPath string                 `json:"websocket_path" koanf:"websocket_path"`
	PhoneAPIURL   string                 `json:"phone_api_url" koanf:"phone_api_url"`
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
	CSRF       CSRF              `json:"csrf" koanf:"csrf"`
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
	}, "."), nil); err != nil {
		return Config{}, fmt.Errorf("config defaults: %w", err)
	}

	path := os.Getenv("WEBPHONE_CONFIG")
	if path == "" {
		path = "/etc/webphone/config.json"
	}
	if _, err := os.Stat(path); err == nil {
		if err := k.Load(file.Provider(path), json.Parser()); err != nil {
			return Config{}, fmt.Errorf("config file %s: %w", path, err)
		}
	}

	if err := k.Load(env.Provider("WEBPHONE_", ".", func(s string) string {
		return envKeyToPath(s)
	}), nil); err != nil {
		return Config{}, fmt.Errorf("config env: %w", err)
	}

	if err := k.Unmarshal("", &cfg); err != nil {
		return Config{}, fmt.Errorf("config unmarshal: %w", err)
	}

	if err := validate(cfg); err != nil {
		return Config{}, fmt.Errorf("config invalid: %w", err)
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
		return fmt.Errorf("gateway.mode is empty")
	default:
		return fmt.Errorf("gateway.mode %q is not one of loopback|webhook", cfg.Gateway.Mode)
	}
	if cfg.Gateway.Mode == GatewayWebhook && cfg.Gateway.WebhookURL == "" {
		return fmt.Errorf("gateway.webhook_url is required in webhook mode")
	}
	if cfg.Addr == "" {
		return fmt.Errorf("addr is empty")
	}
	if cfg.DataDir == "" {
		return fmt.Errorf("data_dir is empty")
	}
	if cfg.SessionTTL <= 0 {
		return fmt.Errorf("session_ttl must be positive")
	}
	if cfg.SessionMaxTTL < cfg.SessionTTL {
		return fmt.Errorf("session_max_ttl (%s) must be >= session_ttl (%s): the absolute cap cannot be shorter than the idle window it bounds", cfg.SessionMaxTTL, cfg.SessionTTL)
	}
	for _, origin := range cfg.CSRF.TrustedOrigins {
		u, err := url.Parse(origin)
		if err != nil || u.Scheme == "" || u.Host == "" {
			return fmt.Errorf("csrf.trusted_origins: %q is not an absolute origin (want scheme://host, e.g. https://pbx.example.org)", origin)
		}
	}
	for _, proxy := range cfg.CSRF.TrustedProxies {
		if net.ParseIP(proxy) == nil {
			if _, _, err := net.ParseCIDR(proxy); err != nil {
				return fmt.Errorf("csrf.trusted_proxies: %q is not an IP address or CIDR network", proxy)
			}
		}
	}
	for ext, did := range cfg.Identities {
		parsed, err := domain.ParseExtension(ext)
		if err != nil || parsed.String() != ext {
			return fmt.Errorf("identities: key %q is not a normalized extension (sanitize it exactly as it appears after login)", ext)
		}
		if _, err := domain.ParsePhone(did); err != nil {
			return fmt.Errorf("identities: DID for extension %q has no dialable characters", ext)
		}
	}
	return nil
}
