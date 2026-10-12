package sipcred

import "testing"

// The wire vector pins the derivation contract across repos: the stack's
// directory bridge (and its VM test fixture) must agree byte-for-byte
// with this server. Changing Derive means changing BOTH sides plus this
// vector — treat it as a cross-repo contract, not an implementation
// detail.
const wireVector = "S9cjQ8oET4qrHZjwdgbNsc9H3wwIn_e1st9E5A2GmXA"

func TestDeriveMatchesTheWireVector(t *testing.T) {
	got := Derive("test-secret", "test-token")
	if got != wireVector {
		t.Fatalf("Derive(secret, token) = %q, want the pinned vector %q", got, wireVector)
	}
}

func TestDeriveIsDeterministic(t *testing.T) {
	first := Derive("s", "token")
	second := Derive("s", "token")
	if first != second {
		t.Fatalf("Derive not deterministic: %q vs %q", first, second)
	}
}

func TestDeriveBindsTheToken(t *testing.T) {
	if Derive("s", "token-a") == Derive("s", "token-b") {
		t.Fatal("different session tokens derived the same credential")
	}
}

func TestDeriveBindsTheSecret(t *testing.T) {
	if Derive("secret-a", "token") == Derive("secret-b", "token") {
		t.Fatal("different secrets derived the same credential")
	}
}

func TestDeriveIsBase64URLOfAHMAC(t *testing.T) {
	got := Derive("k", "t")
	// 32 raw bytes -> 43 base64url chars, no padding.
	if len(got) != 43 {
		t.Fatalf("derived credential length = %d, want 43 (sha256 base64url)", len(got))
	}
	for _, r := range got {
		if r == '+' || r == '/' || r == '=' {
			t.Fatalf("derived credential %q is not unpadded base64url", got)
		}
	}
}
