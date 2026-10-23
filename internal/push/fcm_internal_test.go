package push

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json/v2"
	"encoding/pem"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/larsartmann/go-error-family"
	errorfamilytest "github.com/larsartmann/go-error-family/errorfamilytest"
)

// writeTestCredentials mints a real RSA service-account key file with
// the token endpoint pointed at the given server and hands back the
// key it wrote (the assertion-verification arm needs the matching
// public key).
func writeTestCredentials(t *testing.T, tokenURI string) (string, *rsa.PrivateKey) {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	der, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		t.Fatal(err)
	}
	pemKey := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der})
	raw, err := json.Marshal(serviceAccount{
		ProjectID:   "webphone-test",
		ClientEmail: "push@test.iam.gserviceaccount.com",
		PrivateKey:  string(pemKey),
		TokenURI:    tokenURI,
	})
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "fcm.json")
	if err := os.WriteFile(path, raw, 0o600); err != nil {
		t.Fatal(err)
	}
	return path, key
}

// verifyAssertion checks the JWT-bearer grant's signature against the
// service account public key and returns its claims.
func verifyAssertion(t *testing.T, assertion string, publicKey *rsa.PublicKey) map[string]any {
	t.Helper()
	parts := strings.Split(assertion, ".")
	if len(parts) != 3 {
		t.Fatalf("assertion has %d parts, want 3", len(parts))
	}
	claimsRaw, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		t.Fatalf("decode claims: %v", err)
	}
	signature, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		t.Fatalf("decode signature: %v", err)
	}
	digest := sha256.Sum256([]byte(parts[0] + "." + parts[1]))
	if err := rsa.VerifyPKCS1v15(publicKey, crypto.SHA256, digest[:], signature); err != nil {
		t.Fatalf("assertion signature: %v", err)
	}
	var claims map[string]any
	if err := json.Unmarshal(claimsRaw, &claims); err != nil {
		t.Fatalf("unmarshal claims: %v", err)
	}
	return claims
}

func TestFcmSendFullWire(t *testing.T) {
	var gotAuth string
	var gotBody map[string]any
	var gotAssertion string
	var gotGrant string
	tokenServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			t.Errorf("parse form: %v", err)
		}
		gotGrant = r.FormValue("grant_type")
		gotAssertion = r.FormValue("assertion")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"access_token":"test-access-token","expires_in":3600}`))
	}))
	defer tokenServer.Close()
	sendServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		raw, _ := io.ReadAll(r.Body)
		if err := json.Unmarshal(raw, &gotBody); err != nil {
			t.Errorf("decode body: %v", err)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer sendServer.Close()

	credentialsPath, key := writeTestCredentials(t, tokenServer.URL)
	sender, err := NewFcmSender(credentialsPath)
	if err != nil {
		t.Fatal(err)
	}
	sender.sendURL = sendServer.URL

	ring := Ring{Event: EventRing, From: "+15550000001", CallID: "call-7"}
	if err := sender.Send(context.Background(), "device-token", ring); err != nil {
		t.Fatalf("send: %v", err)
	}

	if gotGrant != fcmGrantType {
		t.Fatalf("grant_type %q, want %q", gotGrant, fcmGrantType)
	}
	claims := verifyAssertion(t, gotAssertion, &key.PublicKey)
	if claims["iss"] != "push@test.iam.gserviceaccount.com" {
		t.Fatalf("iss %v", claims["iss"])
	}
	if claims["scope"] != fcmScope {
		t.Fatalf("scope %v", claims["scope"])
	}
	if claims["aud"] != tokenServer.URL {
		t.Fatalf("aud %v", claims["aud"])
	}
	if gotAuth != "Bearer test-access-token" {
		t.Fatalf("send auth %q", gotAuth)
	}
	message, _ := gotBody["message"].(map[string]any)
	if message["token"] != "device-token" {
		t.Fatalf("token %v", message["token"])
	}
	data, _ := message["data"].(map[string]any)
	if data["event"] != "ring" || data["call_id"] != "call-7" || data["from"] != "+15550000001" {
		t.Fatalf("data %v", data)
	}
	android, _ := message["android"].(map[string]any)
	if android["priority"] != "HIGH" {
		t.Fatalf("priority %v", android["priority"])
	}
}

func TestFcmSendClassifiesDeadTokenAndRejectedCredential(t *testing.T) {
	tokenServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"access_token":"t","expires_in":3600}`))
	}))
	defer tokenServer.Close()

	t.Run("dead token", func(t *testing.T) {
		sendServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		defer sendServer.Close()
		credentialsPath, _ := writeTestCredentials(t, tokenServer.URL)
		sender, err := NewFcmSender(credentialsPath)
		if err != nil {
			t.Fatal(err)
		}
		sender.sendURL = sendServer.URL
		err = sender.Send(context.Background(), "dead", Ring{Event: EventRing})
		if !errors.Is(err, ErrUnregistered) {
			t.Fatalf("404 error = %v, want ErrUnregistered", err)
		}
		errorfamilytest.AssertCode(t, err, "push.fcm_send")
	})

	t.Run("rejected credential", func(t *testing.T) {
		rejected := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
		}))
		defer rejected.Close()
		credentialsPath, _ := writeTestCredentials(t, rejected.URL)
		sender, err := NewFcmSender(credentialsPath)
		if err != nil {
			t.Fatal(err)
		}
		err = sender.Send(context.Background(), "t", Ring{Event: EventRing})
		if !errors.Is(err, ErrUnauthorized) {
			t.Fatalf("401 error = %v, want ErrUnauthorized", err)
		}
		errorfamilytest.AssertFamily(t, err, errorfamily.Rejection)
	})
}

func TestFcmCredentialsFailClosed(t *testing.T) {
	dir := t.TempDir()

	missing := filepath.Join(dir, "missing.json")
	_, err := NewFcmSender(missing)
	if err == nil {
		t.Fatal("missing file must fail")
	}
	errorfamilytest.AssertFamily(t, err, errorfamily.Rejection)
	errorfamilytest.AssertCode(t, err, "config.push.fcm")

	garbage := filepath.Join(dir, "garbage.json")
	if err := os.WriteFile(garbage, []byte("not json"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := NewFcmSender(garbage); err == nil {
		t.Fatal("garbage file must fail")
	}

	incomplete := filepath.Join(dir, "incomplete.json")
	if err := os.WriteFile(incomplete, []byte(`{"project_id":"p"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := NewFcmSender(incomplete); err == nil {
		t.Fatal("file without client_email must fail")
	}
}

func TestFcmSenderDisabledWhenUnconfigured(t *testing.T) {
	sender, err := NewFcmSender("")
	if err != nil {
		t.Fatal(err)
	}
	if sender.Enabled() {
		t.Fatal("empty credentials path must disable the sender")
	}
}
