package gateway

import (
	"net/http"

	"github.com/larsartmann/webphone/internal/domain"
)

// SelfSendRejection synthesizes the provider refusal a send addressed to
// the extension's own DID would earn anyway (Telnyx answers 400, error
// 40310, "Source and destination cannot be the same number"): the calling
// service fails the row locally instead of burning a provider roundtrip
// (SUPERB send-failure train C, 2026-09-24). The returned error carries
// the provider-rejection shape, so the existing 422 send-failure arm
// renders it like any provider refusal. A nil return — no identities map,
// no DID for this owner, an unparseable DID, or a different destination —
// means no local verdict; the gateway decides. detail is the lane's
// user-facing refusal text (English, operator-greppable like every
// service reason).
func SelfSendRejection(
	identities map[string]string, owner domain.Extension, to domain.Phone, detail string,
) *ErrProviderRejected {
	did, ok := identities[owner.String()]
	if !ok {
		return nil
	}
	own, err := domain.ParsePhone(did)
	if err != nil || to != own {
		return nil
	}
	return &ErrProviderRejected{Status: http.StatusBadRequest, Detail: detail}
}
