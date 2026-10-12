// Package sipcred derives the per-session SIP REGISTER credential: an
// HMAC-SHA256 of the session token under the deployment's dedicated
// derivation secret. The directory password never reaches the browser;
// FreeSWITCH validates the derived credential dynamically (via the
// internal directory endpoint), so revocation is inherent — the
// credential dies with the session at the next REGISTER.
package sipcred

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
)

// Derive maps (secret, sessionToken) to the REGISTER credential:
// base64url(HMAC-SHA256(secret, token)). Deterministic so the server
// endpoint and any verifier agree without shared state; the token is
// 256 bits of entropy, so the derived credential never repeats across
// sessions and is worthless once the session is gone.
func Derive(secret, sessionToken string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(sessionToken))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}
