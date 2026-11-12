// The dial-cooldown seam's GATE + VIEW surface: the interface the
// server's handlers consume and the projection rows the journal
// backend folds. The engine-backed store that used to implement the
// gate here retired with the readmodels engine (2026-11-11
// storage-ladder train C) — internal/cqrs/dialwindows.go is the one
// implementation now.
package readmodels

import (
	"context"
	"time"

	"github.com/larsartmann/webphone/internal/domain"
)

// DialWindowsMaxPerExtension bounds one extension's cooldown rows (the
// transcripts posture: nothing the owner authored is ever lost to this
// cap — cooldown rows are SYSTEM facts, so the trim expunges the
// oldest instead of rejecting).
const DialWindowsMaxPerExtension = 500

// DialSlotsMaxPerResponse bounds the typeahead's active-window read.
const DialSlotsMaxPerResponse = 50

// DialWindowView is the projection row ONE journal fold produces for
// an (owner, remote) cooldown slot: the LAST allowed dial's timestamp.
//
// branching-flow COMPOSITION_mixin (baselined): the projection row
// mirrors the domain fact verbatim (the fold is a plain insert).
// brandid-lint:ignore(BD003) engine projection row: the json tags are the storage schema (strong-id wire ruling)
type DialWindowView struct {
	Key    domain.DialWindowKey `json:"key"`
	Owner  string               `json:"owner"`
	Remote string               `json:"remote"`
	At     int64                `json:"at"`
}

// DialOverrideView is the projection row for the one-shot bypass slot
// (issue #7): Consumed=true marks an override the gate already spent.
//
// branching-flow COMPOSITION_mixin (baselined): the projection row
// mirrors the domain fact verbatim (the fold is a plain insert).
// brandid-lint:ignore(BD003) engine projection row: the json tags are the storage schema (strong-id wire ruling)
type DialOverrideView struct {
	Key      domain.DialWindowKey `json:"key"`
	Owner    string               `json:"owner"`
	Remote   string               `json:"remote"`
	Consumed bool                 `json:"consumed"`
	At       int64                `json:"at"`
}

// ActiveSlot is one live cooldown window: the remote is inside its
// cooldown until LastAt+Window.
//
// branching-flow COMPOSITION_mixin (baselined): mirrors DialWindowView
// plus the window length the verdict needs (the projection row stores
// only the stamp). brandid-lint:ignore(BD003) wire DTO: the remote is the dialable string (strong-id wire ruling)
type ActiveSlot struct {
	Remote string `json:"remote"`
	LastAt int64  `json:"last_at"`
	Window int64  `json:"window"`
}

// DialGate is the dial-cooldown seam (issue #7) as the server's
// handlers see it: the gate's check+record transaction, the one-shot
// override arm, and the typeahead's active-slot read. The
// journal-backed service (internal/cqrs) is the implementation.
type DialGate interface {
	Attempt(ctx context.Context, owner domain.Extension, remote string, cooldown time.Duration) (allowed bool, stamp time.Time, err error)
	ArmOverride(ctx context.Context, owner domain.Extension, remote string) error
	Last(ctx context.Context, owner domain.Extension, remote string) (time.Time, bool, error)
	Active(ctx context.Context, owner domain.Extension, cooldown time.Duration) ([]ActiveSlot, error)
}
