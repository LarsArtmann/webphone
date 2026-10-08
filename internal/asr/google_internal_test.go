package asr

import "testing"

// The regional host derivation is the EU-residency guarantee for the
// google wire — a wrong host would silently route caller audio through
// the global (US) endpoint, so the shape is pinned here.
func TestGoogleEndpointDerivation(t *testing.T) {
	cases := map[string]string{
		"europe-west3": "https://speech.europe-west3.rep.googleapis.com",
		"us-central1":  "https://speech.us-central1.rep.googleapis.com",
	}
	for location, want := range cases {
		if got := googleEndpoint(location); got != want {
			t.Errorf("googleEndpoint(%q): got %q, want %q", location, got, want)
		}
	}
}
