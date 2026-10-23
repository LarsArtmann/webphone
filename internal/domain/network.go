package domain

import (
	"strings"

	"github.com/larsartmann/go-error-family"
)

// Network is the messaging network a conversation rides: the transport
// plus the identity system of the far end. It is immutable at thread
// birth — a conversation never migrates networks — and it owns the
// validation rule for the remote Address (see Address).
//
// This axis is deliberately DISTINCT from Channel (message.go): Channel
// is the message FORMAT (sms text vs mms attachments); Network is the
// TRANSPORT (telephony SMS vs Signal vs Matrix...). A Signal message
// with a photo is network=signal, channel=mms.
type Network string

const (
	// NetworkSMS is the telephony SMS/MMS lane via the PBX gateway —
	// the network every thread rode before this type existed.
	NetworkSMS Network = "sms"
	// NetworkSignal rides the Signal network through a signal-cli
	// daemon (JSON-RPC over a local unix socket).
	NetworkSignal Network = "signal"
	// NetworkSession rides the Session network. No implementation
	// exists today (research verdict 2026-10-23: no Go library, no
	// headless path); the network is modeled so the data model is
	// complete and the seam can adopt one without another migration.
	NetworkSession Network = "session"
	// NetworkMatrix rides the Matrix protocol via the Client-Server
	// API against a homeserver.
	NetworkMatrix Network = "matrix"
)

// ParseNetwork validates a stored or wire network string.
func ParseNetwork(raw string) (Network, error) {
	switch Network(raw) {
	case NetworkSMS, NetworkSignal, NetworkSession, NetworkMatrix:
		return Network(raw), nil
	default:
		return "", errorfamily.Newf(errorfamily.Rejection, "domain.network", "network %q is not one of sms|signal|session|matrix", raw)
	}
}

// MustNetwork is ParseNetwork for row-scan sites: the writer only ever
// stores parsed networks, so an unparsable value is a corrupt row.
func MustNetwork(raw string) Network { return must(ParseNetwork(raw)) }

// Address is a validated remote endpoint of a conversation: an
// E.164-ish dialable number on the sms and signal networks, a Matrix
// user id (@local:server) on matrix, a Session id (05-prefixed hex) on
// session. The network owns the validation rule; ParseAddress is the
// only door. Like Phone, the struct makes two addresses of different
// networks non-interchangeable at compile time.
type Address struct {
	network Network
	value   string
}

// ParseAddress validates and brands a remote endpoint for one network.
// The sms and signal rules reuse the dialable alphabet (Signal
// addresses are phone numbers); session requires the canonical 66-char
// 05-prefixed hex Session id; matrix requires an @localpart:server user
// id (the thread identity for a DM — the room mapping lives in the
// matrix client, so a conversation stays keyed on the person, like
// every other network).
func ParseAddress(network Network, raw string) (Address, error) {
	switch network {
	case NetworkSMS, NetworkSignal:
		clean := sanitizeDialable(raw)
		if clean == "" {
			return Address{}, errorfamily.Newf(errorfamily.Rejection, "domain.address", "address %q has no dialable characters", raw)
		}
		if len(clean) > 32 {
			return Address{}, errorfamily.Newf(errorfamily.Rejection, "domain.address", "address %q longer than 32 characters", raw)
		}
		return Address{network: network, value: clean}, nil
	case NetworkSession:
		lower := strings.ToLower(raw)
		if len(lower) != 66 || !strings.HasPrefix(lower, "05") || !isHex(lower[2:]) {
			return Address{}, errorfamily.Newf(errorfamily.Rejection, "domain.address", "address %q is not a Session id (05 + 64 hex characters)", raw)
		}
		return Address{network: network, value: lower}, nil
	case NetworkMatrix:
		at := strings.Index(raw, "@")
		colon := strings.LastIndex(raw, ":")
		if at != 0 || colon < 2 || colon == len(raw)-1 || len(raw) > 255 {
			return Address{}, errorfamily.Newf(errorfamily.Rejection, "domain.address", "address %q is not a Matrix user id (@localpart:server)", raw)
		}
		return Address{network: network, value: raw}, nil
	default:
		return Address{}, errorfamily.Newf(errorfamily.Rejection, "domain.network", "network %q is not one of sms|signal|session|matrix", network)
	}
}

// MustParseAddress is ParseAddress for literals in tests and seeds.
func MustParseAddress(network Network, raw string) Address {
	return must(ParseAddress(network, raw))
}

// Network reports the network the address rides.
func (a Address) Network() Network { return a.network }

// String returns the raw address value.
func (a Address) String() string { return a.value }

// IsZero reports whether the address is the zero value.
func (a Address) IsZero() bool { return a.value == "" }

// isHex reports whether every character is a lowercase hex digit.
func isHex(s string) bool {
	for i := 0; i < len(s); i++ {
		c := s[i]
		if (c < '0' || c > '9') && (c < 'a' || c > 'f') {
			return false
		}
	}
	return len(s) > 0
}
