package crm

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"math/big"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	errorfamily "github.com/larsartmann/go-error-family"
)

// testCA is an in-test private certificate authority: the shape of the
// deployment issue #9 describes (a CRM vhost signed by a CA the host's
// trust store does not know).
type testCA struct {
	cert *x509.Certificate
	key  *ecdsa.PrivateKey
	pem  []byte
}

func newTestCA(t *testing.T) *testCA {
	t.Helper()

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("ca key: %v", err)
	}

	template := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "webphone test CA"},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(time.Hour),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageDigitalSignature,
		BasicConstraintsValid: true,
		IsCA:                  true,
	}

	der, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	if err != nil {
		t.Fatalf("ca cert: %v", err)
	}

	cert, err := x509.ParseCertificate(der)
	if err != nil {
		t.Fatalf("parse ca cert: %v", err)
	}

	return &testCA{
		cert: cert,
		key:  key,
		pem:  pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}),
	}
}

// leafKeyPair signs a server certificate for the loopback listener and
// returns it as a usable pair whose chain carries the signing CA.
func (ca *testCA) leafKeyPair(t *testing.T, ip net.IP) tls.Certificate {
	t.Helper()

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("leaf key: %v", err)
	}

	template := &x509.Certificate{
		SerialNumber: big.NewInt(2),
		Subject:      pkix.Name{CommonName: "crm.test"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		DNSNames:     []string{"crm.test"},
		IPAddresses:  []net.IP{ip},
	}

	der, err := x509.CreateCertificate(rand.Reader, template, ca.cert, &key.PublicKey, ca.key)
	if err != nil {
		t.Fatalf("leaf cert: %v", err)
	}

	keyDER, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		t.Fatalf("leaf key marshal: %v", err)
	}

	chainPEM := append(
		pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}),
		ca.pem...)

	pair, err := tls.X509KeyPair(chainPEM,
		pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDER}))
	if err != nil {
		t.Fatalf("leaf keypair: %v", err)
	}

	return pair
}

// caBundleFile writes the CA PEM where a deployment would find it.
func caBundleFile(t *testing.T, contents []byte) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "crm-ca.pem")
	if err := os.WriteFile(path, contents, 0o600); err != nil {
		t.Fatalf("write ca bundle: %v", err)
	}

	return path
}

// newTLSStubServer serves the stub over a leaf signed by the test CA.
func newTLSStubServer(t *testing.T, ca *testCA, stub *stubCRM) *httptest.Server {
	t.Helper()

	srv := httptest.NewUnstartedServer(http.HandlerFunc(stub.handler))
	srv.TLS = &tls.Config{
		Certificates: []tls.Certificate{ca.leafKeyPair(t, net.ParseIP("127.0.0.1"))},
		MinVersion:   tls.VersionTLS12,
	}
	srv.StartTLS()
	t.Cleanup(srv.Close)

	return srv
}

// TestClientTrustsOnlyTheConfiguredCA pins issue #9's CA lane: the same
// private-CA vhost is unreachable without the bundle and reachable with
// it, and the trust it buys stays inside this client — the transport is
// the seam's own, never http.DefaultTransport, so a CRM CA can never
// vouch for the OIDC issuer or the archive.
func TestClientTrustsOnlyTheConfiguredCA(t *testing.T) {
	ca := newTestCA(t)
	stub := &stubCRM{lookupStatus: http.StatusOK, lookupJSON: `{"results":[]}`}
	srv := newTLSStubServer(t, ca, stub)

	untrusting, err := NewClient(srv.URL, "sekrit", "")
	if err != nil {
		t.Fatalf("NewClient without ca: %v", err)
	}

	if _, err := untrusting.LookupByPhone(t.Context(), "+4930123456"); errorfamily.Code(err) != "crm.transport" {
		t.Fatalf("untrusted vhost error %v (code %q), want crm.transport", err, errorfamily.Code(err))
	}

	trusting, err := NewClient(srv.URL, "sekrit", caBundleFile(t, ca.pem))
	if err != nil {
		t.Fatalf("NewClient with ca: %v", err)
	}

	// A completed handshake on an empty result set is ErrNotFound: the
	// seam read the answer, which is the whole point of the CA lane.
	if _, err := trusting.LookupByPhone(t.Context(), "+4930123456"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("configured CA lookup error %v, want ErrNotFound (handshake + auth + decode)", err)
	}

	transport, ok := trusting.client.Transport.(*http.Transport)
	if !ok || transport == http.DefaultTransport {
		t.Fatal("the CA lane must own its transport, never http.DefaultTransport")
	}

	if transport.TLSClientConfig == nil || transport.TLSClientConfig.RootCAs == nil {
		t.Fatal("the CA lane must install the extended root pool on its transport")
	}
}

// TestCAFileFailuresAreLoud pins the boot-side posture: an unreadable
// path and a bundle that holds no certificate both refuse the client with
// the crm.ca_file code, so a typo names itself at startup instead of
// degrading enrichment forever.
func TestCAFileFailuresAreLoud(t *testing.T) {
	t.Run("missing file", func(t *testing.T) {
		_, err := NewClient("https://crm.example.org", "sekrit", filepath.Join(t.TempDir(), "absent.pem"))
		if errorfamily.Code(err) != "crm.ca_file" {
			t.Fatalf("error %v (code %q), want crm.ca_file", err, errorfamily.Code(err))
		}
	})

	t.Run("no certificate in the bundle", func(t *testing.T) {
		_, err := NewClient("https://crm.example.org", "sekrit",
			caBundleFile(t, []byte("this is not a certificate\n")))
		if errorfamily.Code(err) != "crm.ca_file" {
			t.Fatalf("error %v (code %q), want crm.ca_file", err, errorfamily.Code(err))
		}
	})
}
