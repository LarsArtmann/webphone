package crm

import (
	"crypto/tls"
	"crypto/x509"
	"net/http"
	"os"

	"github.com/larsartmann/go-error-family"
)

// newHTTPClient builds the seam's HTTP client. Without caFile it is the
// stock client. With caFile it ADDS that bundle's certificates to the
// system roots for the CRM client ALONE.
//
// Scoping is the whole point. A private-CA vhost (the usual deployment:
// the CRM behind a cert the host's trust store does not know) can be
// trusted host-wide via security.pki, or process-wide via SSL_CERT_FILE —
// and both vouch for every TLS connection the machine makes afterwards,
// including the OIDC issuer, the paperless archive and the ASR provider.
// One client, one CA: the trust boundary stays exactly as wide as the
// need. Additive, never replacing, so a public-CA deployment keeps its
// chains.
func newHTTPClient(caFile string) (*http.Client, error) {
	if caFile == "" {
		// Clone (not share) DefaultTransport: a client with its own pool
		// never rides connections that another test's httptest.Server.Close
		// may retire mid-roundtrip (httptest "helpfully" closes idle
		// connections on the global transport; a shared pool made the
		// failure-classes suite flaky under parallel tests).
		return &http.Client{
			Timeout:   requestTimeout,
			Transport: http.DefaultTransport.(*http.Transport).Clone(),
		}, nil
	}

	bundle, err := os.ReadFile(caFile)
	if err != nil {
		return nil, errorfamily.WrapRejectionf(err, "crm.ca_file", "crm.ca_file %s", caFile)
	}

	pool, err := x509.SystemCertPool()
	if err != nil {
		// A host without a system bundle has no roots to extend; start
		// from an empty pool rather than lose the configured CA too.
		pool = x509.NewCertPool()
	}

	if !pool.AppendCertsFromPEM(bundle) {
		return nil, errorfamily.Newf(errorfamily.Rejection, "crm.ca_file",
			"crm.ca_file %s holds no PEM certificate: point it at the CA bundle that signed the CRM vhost", caFile)
	}

	base, ok := http.DefaultTransport.(*http.Transport)
	if !ok {
		// Unreachable while net/http ships its own *http.Transport as
		// DefaultTransport; guarded so a host that replaced the global
		// with a custom RoundTripper degrades to a zero-config transport
		// instead of panicking at boot.
		base = &http.Transport{}
	}
	transport := base.Clone()
	transport.TLSClientConfig = &tls.Config{RootCAs: pool, MinVersion: tls.VersionTLS12}

	return &http.Client{Timeout: requestTimeout, Transport: transport}, nil
}
