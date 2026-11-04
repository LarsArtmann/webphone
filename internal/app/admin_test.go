package app

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/larsartmann/webphone/internal/config"
	"github.com/larsartmann/webphone/internal/domain"
)

// AdminService adapter pins: the shared mapping (with the historical
// passkey fallback) flows into the service with the admin flag, and the
// boot-compatibility degrade leaves a broken mapping WITHOUT an admin
// at nil (the pre-admin boot behavior) while an admin on a broken
// mapping fails loudly.

func adminPassword(t *testing.T, dir, password string) string {
	t.Helper()
	path := filepath.Join(dir, password+".txt")
	if err := os.WriteFile(path, []byte(password+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestAdminServiceResolvesTheSharedMapping(t *testing.T) {
	dir := t.TempDir()
	cfg := config.Config{
		Auth: config.Auth{
			Users: map[string]config.PasskeyUser{
				"boss@example.com":   {Extensions: []string{"1001"}, Admin: true},
				"worker@example.com": {Extensions: []string{"1002"}},
			},
			ExtensionPasswordFiles: map[string]string{
				"1001": adminPassword(t, dir, "boss"),
				"1002": adminPassword(t, dir, "worker"),
			},
		},
		Identities: map[string]string{"1001": "+49301111"},
	}
	service, err := AdminService(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if !service.IsAdmin("boss@example.com") {
		t.Error("the shared map's admin flag must reach the service")
	}
	if service.IsAdmin("worker@example.com") {
		t.Error("a non-admin mapping must stay non-admin")
	}
	users := service.Users()
	if len(users) != 2 || users[0].Email != "boss@example.com" || !users[0].Admin {
		t.Errorf("users = %+v, want boss (admin) first, sorted", users)
	}
	password, err := service.SIPPassword(domain.MustParseExtension("1001"))
	if err != nil || password != "boss" {
		t.Errorf("SIPPassword = %q, %v; want the mapped line's password", password, err)
	}
}

func TestAdminServiceEmptyConfigIsHarmless(t *testing.T) {
	service, err := AdminService(config.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if service == nil {
		t.Fatal("an empty config still builds the (empty) service; the gate answers false for everyone")
	}
	if service.IsAdmin("anyone@example.com") || service.Users() != nil || service.Numbers() != nil {
		t.Error("an empty mapping must list nothing and admit nobody")
	}
}

func TestAdminServiceDegradesBrokenMappingWithoutAdmin(t *testing.T) {
	// A mapped extension WITHOUT a password file: invalid for every
	// login mode, but with both modes off such a config historically
	// booted — the admin adapter must not regress that boot.
	cfg := config.Config{
		Auth: config.Auth{
			Users:                  map[string]config.PasskeyUser{"worker@example.com": {Extensions: []string{"1002"}}},
			ExtensionPasswordFiles: map[string]string{},
		},
	}
	service, err := AdminService(cfg)
	if err != nil {
		t.Fatalf("no admin configured: the broken mapping must degrade to a nil service, got %v", err)
	}
	if service != nil {
		t.Error("no admin on a broken mapping: the service must be nil (byte-identical off state)")
	}
}

func TestAdminServiceFailsLoudlyForAdminOnBrokenMapping(t *testing.T) {
	cfg := config.Config{
		Auth: config.Auth{
			Users:                  map[string]config.PasskeyUser{"boss@example.com": {Extensions: []string{"1001"}, Admin: true}},
			ExtensionPasswordFiles: map[string]string{},
		},
	}
	if service, err := AdminService(cfg); err == nil || service != nil {
		t.Fatal("an admin configured on a mapping that cannot source credentials must fail the boot loudly")
	}
}
