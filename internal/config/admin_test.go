package config

import (
	"os"
	"path/filepath"
	"testing"
)

// The admin flag's config pins: it rides the SHARED user map, the
// shared/passkey pair must agree about it, and an admin without
// extensions is a config bug (an admin still signs in through a mapped
// line).

func adminAuth(t *testing.T) Auth {
	t.Helper()
	return Auth{Users: map[string]PasskeyUser{
		"boss@example.com": {Extensions: []string{"1000"}, Admin: true},
	}}
}

func TestValidateAdminUsersAcceptsMappedAdmin(t *testing.T) {
	if err := validateAdminUsers(adminAuth(t)); err != nil {
		t.Fatalf("mapped admin must pass: %v", err)
	}
	if err := validateAdminUsers(Auth{}); err != nil {
		t.Fatalf("no users (mode off everywhere) must pass: %v", err)
	}
}

func TestValidateAdminUsersRejectsAdminWithoutExtensions(t *testing.T) {
	auth := adminAuth(t)
	auth.Users["boss@example.com"] = PasskeyUser{Admin: true}
	if err := validateAdminUsers(auth); err == nil {
		t.Fatal("an admin without extensions cannot sign in anywhere; the config must say so at boot")
	}
	// The check reads the historical passkey home as a fallback too.
	passkeyHome := Auth{Passkey: Passkey{Users: map[string]PasskeyUser{
		"boss@example.com": {Admin: true},
	}}}
	if err := validateAdminUsers(passkeyHome); err == nil {
		t.Fatal("passkey.users fallback home must be validated identically")
	}
}

func TestSameUserMapComparesAdminFlag(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "pw")
	if err := os.WriteFile(file, []byte("secret\n"), 0o640); err != nil {
		t.Fatal(err)
	}
	shared := map[string]PasskeyUser{
		"boss@example.com": {Extensions: []string{"1000"}, DisplayName: "Boss"},
	}
	divergent := map[string]PasskeyUser{
		"boss@example.com": {Extensions: []string{"1000"}, DisplayName: "Boss", Admin: true},
	}
	auth := Auth{Users: shared, Passkey: Passkey{Users: divergent, ExtensionPasswordFiles: map[string]string{"1000": file}}}
	if _, err := resolvedAuth(auth); err == nil {
		t.Fatal("a divergent admin flag is two homes disagreeing about who may read every line; Load must reject it")
	}
}
