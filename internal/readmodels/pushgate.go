// The push-device seam's GATE + VIEW surface: the interface the push
// surfaces consume and the projection row the journal backend folds.
// The engine-backed store that used to implement the gate here retired
// with the readmodels engine (2026-11-11 storage-ladder closeout) —
// internal/cqrs/pushdevices.go is the one implementation now.
package readmodels

import (
	"context"

	"github.com/larsartmann/webphone/internal/domain"
)

// PushDevicesMaxPerExtension bounds one extension's registered devices
// (the contacts-cap posture: the bound is the durability story).
const PushDevicesMaxPerExtension = 8

// PushDeviceView is the projection row ONE journal fold produces for a
// registered device. The LIST never echoes the token to the browser —
// write-only credential doctrine — but the row carries it for the
// server-side fan-out.
//
// brandid-lint:ignore(BD003) engine projection row: the json tags are the storage schema (strong-id wire ruling)
type PushDeviceView struct {
	Owner     string `json:"owner"`
	DeviceID  string `json:"device_id"`
	Transport string `json:"transport"`
	Token     string `json:"token"`
	UpdatedAt int64  `json:"updated_at"`
}

// PushDeviceGate is the registered-device seam as the push surfaces
// see it (register/rotate, fan-out list, deregister). The
// journal-backed service (internal/cqrs) is the implementation.
type PushDeviceGate interface {
	Upsert(ctx context.Context, owner domain.Extension, deviceID, transport, token string) error
	List(ctx context.Context, owner domain.Extension) ([]PushDeviceView, error)
	Remove(ctx context.Context, owner domain.Extension, deviceID string) error
}
