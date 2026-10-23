package authmap

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/larsartmann/go-error-family/errorfamilytest"

	"github.com/larsartmann/webphone/internal/domain"
)

func TestResolveNormalizesCaseAndWhitespace(t *testing.T) {
	m, err := New(map[string]User{
		"lars@example.com": {Email: "lars@example.com", DisplayName: "Lars", Extensions: []domain.Extension{domain.MustParseExtension("1000")}},
	}, map[string]string{"1000": fileWith(t, "secret")})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	user, ok := m.Resolve("  LARS@Example.COM ")
	if !ok {
		t.Fatal("mapping lookup must normalize case and whitespace")
	}
	if user.SessionExtension().String() != "1000" {
		t.Errorf("session extension = %q, want 1000", user.SessionExtension().String())
	}
	if _, ok := m.Resolve("nobody@example.com"); ok {
		t.Fatal("unmapped email must not resolve")
	}
}

func TestMappedByExtensionReverseResolves(t *testing.T) {
	ext := domain.MustParseExtension("1001")
	m, err := New(map[string]User{
		"alice@example.com": {Email: "alice@example.com", DisplayName: "Alice", Extensions: []domain.Extension{ext}},
	}, map[string]string{"1001": fileWith(t, "secret")})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	user, ok := m.MappedByExtension(ext)
	if !ok || user.Email != "alice@example.com" {
		t.Fatalf("MappedByExtension = (%v, %v), want alice", user.Email, ok)
	}
	if _, ok := m.MappedByExtension(domain.MustParseExtension("1009")); ok {
		t.Fatal("unmapped extension must not resolve")
	}
}

func TestSIPPasswordTrimsAndFailsClosed(t *testing.T) {
	ext := domain.MustParseExtension("1000")
	file := fileWith(t, "  dir-password\n")
	m, err := New(map[string]User{
		"lars@example.com": {Email: "lars@example.com", Extensions: []domain.Extension{ext}},
	}, map[string]string{"1000": file})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if got, err := m.SIPPassword(ext); err != nil || got != "dir-password" {
		t.Fatalf("SIPPassword = (%q, %v), want the trimmed password", got, err)
	}

	// Empty file: the 2026-10-01 outage class.
	if err := os.WriteFile(file, []byte(" \n"), 0o640); err != nil {
		t.Fatal(err)
	}
	_, err = m.SIPPassword(ext)
	errorfamilytest.AssertCode(t, err, "authmap.password_file.empty")

	// Unreadable file: the read code, never a silent empty password.
	if err := os.WriteFile(file, []byte("x"), 0o640); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(file, 0o000); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(file, 0o640) })
	_, err = m.SIPPassword(ext)
	errorfamilytest.AssertCode(t, err, "authmap.password_file.read")

	// Extension with no configured file at all.
	_, err = m.SIPPassword(domain.MustParseExtension("1009"))
	errorfamilytest.AssertCode(t, err, "authmap.password_file.missing")
}

func TestNewRejectsMappedExtensionWithoutPasswordFile(t *testing.T) {
	_, err := New(map[string]User{
		"lars@example.com": {Email: "lars@example.com", Extensions: []domain.Extension{domain.MustParseExtension("1000")}},
	}, nil)
	errorfamilytest.AssertCode(t, err, "authmap.password_file.missing")
}

func TestNilMapFailsClosed(t *testing.T) {
	var m *Map
	if _, ok := m.Resolve("lars@example.com"); ok {
		t.Fatal("nil map must not resolve")
	}
	if _, err := m.SIPPassword(domain.MustParseExtension("1000")); err == nil {
		t.Fatal("nil map must reject a password read")
	}
}

func fileWith(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "secret")
	if err := os.WriteFile(path, []byte(content), 0o640); err != nil {
		t.Fatal(err)
	}
	return path
}
