package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// scrubEnv clears every WEBPHONE_* variable for the duration of the test:
// the env provider reads the whole process environment, so one stray
// WEBPHONE_ADDR on the host would flip assertions.
func scrubEnv(t *testing.T) {
	t.Helper()
	for _, kv := range os.Environ() {
		key, value, _ := strings.Cut(kv, "=")
		if !strings.HasPrefix(key, "WEBPHONE_") {
			continue
		}
		if err := os.Unsetenv(key); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = os.Setenv(key, value) })
	}
}

func absentConfigFile(t *testing.T) string {
	t.Helper()
	return filepath.Join(t.TempDir(), "missing.json")
}

func writeConfigFile(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestLoadDefaults(t *testing.T) {
	scrubEnv(t)
	t.Setenv("WEBPHONE_CONFIG", absentConfigFile(t))

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Addr != ":8080" {
		t.Errorf("addr: got %q, want :8080", cfg.Addr)
	}
	if cfg.DataDir != "/var/lib/webphone" {
		t.Errorf("data_dir: got %q", cfg.DataDir)
	}
	if cfg.WebsocketPath != "/sip" {
		t.Errorf("websocket_path: got %q", cfg.WebsocketPath)
	}
	if cfg.SessionTTL != 24*time.Hour {
		t.Errorf("session_ttl: got %s", cfg.SessionTTL)
	}
	if cfg.Gateway.Mode != GatewayLoopback {
		t.Errorf("gateway.mode: got %q", cfg.Gateway.Mode)
	}
	if len(cfg.ICEServers) != 0 {
		t.Errorf("ice_servers: got %d, want none", len(cfg.ICEServers))
	}
	if len(cfg.Contacts) != 0 {
		t.Errorf("contacts: got %d, want none", len(cfg.Contacts))
	}
}

func TestLoadEnvOverrides(t *testing.T) {
	scrubEnv(t)
	t.Setenv("WEBPHONE_CONFIG", absentConfigFile(t))
	t.Setenv("WEBPHONE_ADDR", "127.0.0.1:9099")
	t.Setenv("WEBPHONE_DATA_DIR", "/tmp/webphone-data") // single underscore stays literal
	t.Setenv("WEBPHONE_GATEWAY__MODE", "webhook")       // double underscore nests
	t.Setenv("WEBPHONE_GATEWAY__WEBHOOK_URL", "http://provider.example:9000")
	t.Setenv("WEBPHONE_GATEWAY__WEBHOOK_SECRET", "provider-secret")

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Addr != "127.0.0.1:9099" {
		t.Errorf("addr: got %q", cfg.Addr)
	}
	if cfg.DataDir != "/tmp/webphone-data" {
		t.Errorf("data_dir: got %q (single underscores must stay literal)", cfg.DataDir)
	}
	if cfg.Gateway.Mode != GatewayWebhook {
		t.Errorf("gateway.mode: got %q", cfg.Gateway.Mode)
	}
	if cfg.Gateway.WebhookURL != "http://provider.example:9000" {
		t.Errorf("gateway.webhook_url: got %q (single underscore inside the leaf key)", cfg.Gateway.WebhookURL)
	}
	if cfg.Gateway.WebhookSecret != "provider-secret" {
		t.Errorf("gateway.webhook_secret: got %q", cfg.Gateway.WebhookSecret)
	}
}

func TestLoadJSONFile(t *testing.T) {
	scrubEnv(t)
	t.Setenv("WEBPHONE_CONFIG", writeConfigFile(t, `{
		"addr": ":9090",
		"data_dir": "/srv/webphone",
		"sip_domain": "pbx.example.org",
		"websocket_path": "/wss",
		"phone_api_url": "http://127.0.0.1:8081",
		"session_ttl": "12h30m",
		"ice_servers": [
			{"urls": ["stun:stun.example.org:3478"]},
			{"urls": ["turn:turn.example.org:3478"], "username": "u", "credential": "pw"}
		],
		"contacts": [{"name": "Support", "number": "2000"}],
		"gateway": {"mode": "loopback"}
	}`))

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Addr != ":9090" || cfg.DataDir != "/srv/webphone" {
		t.Errorf("scalars: addr %q data_dir %q", cfg.Addr, cfg.DataDir)
	}
	if cfg.SIPDomain != "pbx.example.org" || cfg.WebsocketPath != "/wss" {
		t.Errorf("sip config: domain %q path %q", cfg.SIPDomain, cfg.WebsocketPath)
	}
	if cfg.PhoneAPIURL != "http://127.0.0.1:8081" {
		t.Errorf("phone_api_url: got %q", cfg.PhoneAPIURL)
	}
	if cfg.SessionTTL != 12*time.Hour+30*time.Minute {
		t.Errorf("session_ttl: got %s", cfg.SessionTTL)
	}
	if len(cfg.ICEServers) != 2 {
		t.Fatalf("ice_servers: got %d, want 2", len(cfg.ICEServers))
	}
	if len(cfg.ICEServers[0].URLs) != 1 || cfg.ICEServers[0].URLs[0] != "stun:stun.example.org:3478" {
		t.Errorf("ice_servers[0]: %+v", cfg.ICEServers[0])
	}
	if cfg.ICEServers[1].Username != "u" || cfg.ICEServers[1].Credential != "pw" {
		t.Errorf("ice_servers[1] credentials: %+v", cfg.ICEServers[1])
	}
	if len(cfg.Contacts) != 1 || cfg.Contacts[0].Name != "Support" {
		t.Errorf("contacts: %+v", cfg.Contacts)
	}
}

func TestLoadFileOverriddenByEnv(t *testing.T) {
	scrubEnv(t)
	t.Setenv("WEBPHONE_CONFIG", writeConfigFile(t, `{"addr": ":9090"}`))
	t.Setenv("WEBPHONE_ADDR", ":7070")

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Addr != ":7070" {
		t.Errorf("env must win over file: got %q", cfg.Addr)
	}
}

func TestLoadRejectsInvalidConfigs(t *testing.T) {
	cases := []struct {
		name    string
		env     map[string]string
		wantErr string
	}{
		{
			name:    "unknown gateway mode",
			env:     map[string]string{"WEBPHONE_GATEWAY__MODE": "carrier-pigeon"},
			wantErr: "not one of loopback|webhook",
		},
		{
			name:    "webhook mode without url",
			env:     map[string]string{"WEBPHONE_GATEWAY__MODE": "webhook"},
			wantErr: "webhook_url is required",
		},
		{
			name:    "malformed json file",
			wantErr: "config file",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			scrubEnv(t)
			if tc.name == "malformed json file" {
				t.Setenv("WEBPHONE_CONFIG", writeConfigFile(t, `{"addr": `))
			} else {
				t.Setenv("WEBPHONE_CONFIG", absentConfigFile(t))
				for key, value := range tc.env {
					t.Setenv(key, value)
				}
			}
			if _, err := Load(); err == nil || !strings.Contains(err.Error(), tc.wantErr) {
				t.Errorf("error %v, want containing %q", err, tc.wantErr)
			}
		})
	}
}

// TestLoadValidatesCSRFConfig pins the fronting-deployment validation: the
// trusted origins must be absolute origins and the trusted proxies valid
// IP/CIDR entries, because a typo would otherwise surface as a library
// panic at handler construction instead of a startup error naming the key.
func TestLoadValidatesCSRFConfig(t *testing.T) {
	cases := []struct {
		name    string
		file    string
		wantErr string
	}{
		{
			name:    "relative origin",
			file:    `{"csrf": {"trusted_origins": ["pbx.example.org"]}}`,
			wantErr: "not an absolute origin",
		},
		{
			name:    "origin without host",
			file:    `{"csrf": {"trusted_origins": ["https://"]}}`,
			wantErr: "not an absolute origin",
		},
		{
			name:    "proxy neither IP nor CIDR",
			file:    `{"csrf": {"trusted_proxies": ["nginx-local"]}}`,
			wantErr: "not an IP address or CIDR network",
		},
		{
			name:    "valid csrf section loads",
			file:    `{"csrf": {"trusted_proxies": ["127.0.0.1", "10.8.0.0/24"], "trusted_origins": ["https://pbx.example.org"]}}`,
			wantErr: "",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			scrubEnv(t)
			t.Setenv("WEBPHONE_CONFIG", writeConfigFile(t, tc.file))
			_, err := Load()
			if tc.wantErr == "" {
				if err != nil {
					t.Fatalf("valid config rejected: %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
				t.Errorf("error %v, want containing %q", err, tc.wantErr)
			}
		})
	}
}

func TestEnvKeyToPath(t *testing.T) {
	cases := []struct{ env, want string }{
		{"WEBPHONE_ADDR", "addr"},
		{"WEBPHONE_DATA_DIR", "data_dir"},
		{"WEBPHONE_SIP_DOMAIN", "sip_domain"},
		{"WEBPHONE_GATEWAY__MODE", "gateway.mode"},
		{"WEBPHONE_GATEWAY__WEBHOOK_URL", "gateway.webhook_url"},
		{"webphone_gateway__webhook_secret", "gateway.webhook_secret"},
	}
	for _, tc := range cases {
		if got := envKeyToPath(tc.env); got != tc.want {
			t.Errorf("envKeyToPath(%q) = %q, want %q", tc.env, got, tc.want)
		}
	}
}

// TestLoadValidatesIdentities pins the own-number feed's config contract:
// owner-formatted DIDs are allowed (the value renders verbatim), but a
// non-normalized extension key would make the signed-in extension's
// lookup silently miss, and a DID without dialable characters is a typo.
func TestLoadValidatesIdentities(t *testing.T) {
	scrubEnv(t)
	t.Setenv("WEBPHONE_CONFIG", writeConfigFile(t, `{"identities":{"1001":"+49 30 12345678"}}`))
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Identities["1001"] != "+49 30 12345678" {
		t.Fatalf("identities: got %v", cfg.Identities)
	}

	t.Setenv("WEBPHONE_CONFIG", writeConfigFile(t, `{"identities":{"10 01":"+493012345678"}}`))
	if _, err := Load(); err == nil || !strings.Contains(err.Error(), "normalized extension") {
		t.Fatalf("non-normalized key: err %v, want normalized-extension rejection", err)
	}

	t.Setenv("WEBPHONE_CONFIG", writeConfigFile(t, `{"identities":{"1001":"!!!"}}`))
	if _, err := Load(); err == nil || !strings.Contains(err.Error(), "dialable") {
		t.Fatalf("dialable-less DID: err %v, want dialable rejection", err)
	}
}
