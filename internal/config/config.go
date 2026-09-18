// Package config loads the webphone runtime configuration from an optional
// JSON file and WEBPHONE_-prefixed environment variables.
package config

import (
	"fmt"
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
	Addr          string        `json:"addr" koanf:"addr"`
	DataDir       string        `json:"data_dir" koanf:"data_dir"`
	SIPDomain     string        `json:"sip_domain" koanf:"sip_domain"`
	WebsocketPath string        `json:"websocket_path" koanf:"websocket_path"`
	PhoneAPIURL   string        `json:"phone_api_url" koanf:"phone_api_url"`
	SessionTTL    time.Duration `json:"session_ttl" koanf:"session_ttl"`
	ICEServers    []ICEServer   `json:"ice_servers" koanf:"ice_servers"`
	Contacts      []domain.SharedContact `json:"contacts" koanf:"contacts"`
	Gateway       Gateway       `json:"gateway" koanf:"gateway"`
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

// defaults keeps the zero-config path working: loopback gateway, local
// ports, /var/lib/webphone data dir.
func defaults() Config {
	return Config{
		Addr:          ":8080",
		DataDir:       "/var/lib/webphone",
		WebsocketPath: "/sip",
		SessionTTL:    24 * time.Hour,
		Gateway:       Gateway{Mode: GatewayLoopback},
	}
}

// Load reads the optional JSON config file (path from WEBPHONE_CONFIG,
// default /etc/webphone/config.json), then overrides from WEBPHONE_*
// environment variables (dots become underscores: WEBPHONE_GATEWAY_MODE).
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
	return nil
}
