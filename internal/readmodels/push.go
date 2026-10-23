package readmodels

import (
	"context"
	"sync"
	"time"

	metaengine "github.com/larsartmann/go-cqrs-lite/metaengine/v4"
	"github.com/larsartmann/go-cqrs-lite/record/v4"
	"github.com/larsartmann/go-error-family"
	"github.com/larsartmann/webphone/internal/domain"
)

// PushDeviceView is the engine-side projection of one registered push
// device (the token is the device's wake address — the ntfy topic or
// the FCM registration token).
//
// branching-flow DUPLICATE_TYPE (baselined): the engine projection row
// mirrors its wire schema on purpose (registry 2026-10-05); the json
// tags are the storage schema, not a DTO to consolidate.
// brandid-lint:ignore(BD003) engine projection row: the json tags are the storage schema (strong-id wire ruling)
type PushDeviceView struct {
	Owner     string `json:"owner"`
	DeviceID  string `json:"device_id"`
	Transport string `json:"transport"`
	Token     string `json:"token"`
	UpdatedAt int64  `json:"updated_at"`
}

// pushDevicesCollection names the engine collection (query name).
const pushDevicesCollection = "push_devices"

// PushDevicesMaxPerExtension bounds one extension's registered devices:
// a human carries a handful (phone, tablet, spare); eight leaves room
// for a de-Googled extra while keeping the ring fan-out honest.
const PushDevicesMaxPerExtension = 8

type pushDevicesInput struct{ Owner string }

// pushDevicesQuery declares the push-device read model: a keyed
// projection upserted by PushDeviceRegistered facts, removed by
// PushDeviceRemoved facts, filtered by owner. The insert fold is the
// single upsert (re-registration replaces the row — token rotation
// keeps the key), mirroring the contacts shape.
var pushDevicesQuery = metaengine.Query[pushDevicesInput, PushDeviceView](pushDevicesCollection,
	metaengine.OnRecord(domain.PushDeviceRegistered{},
		func(_ record.Record, e domain.PushDeviceRegistered) (domain.PushDeviceKey, PushDeviceView) {
			return e.Key, PushDeviceView{
				Owner:     e.Key.Owner,
				DeviceID:  e.Key.DeviceID,
				Transport: e.Transport,
				Token:     e.Token,
				UpdatedAt: e.UpdatedAt.UnixNano(),
			}
		}),
	metaengine.OnRecord(domain.PushDeviceRemoved{}, metaengine.Remove[PushDeviceView]()),
	metaengine.FilterOnField[PushDeviceView]("owner", metaengine.FilterEq),
	metaengine.SortOnField[PushDeviceView]("updated_at", true),
	metaengine.Volume(humanScaleVolume),
)

// PushDevices is the engine-backed registered-device store behind the
// push seam: owner-scoped upsert (re-registration replaces the token),
// owner-scoped list (the ring lane's fan-out source), owner-scoped
// remove. mu serializes the cap's check-then-apply sequence — the same
// single-process doctrine the contacts store documents.
//
// branching-flow COMPOSITION_mixin (baselined): the shared engine-seam
// shape (models+reader+mu, see EnrollTokens/Contacts) — deliberate
// symmetry, not a mixin candidate.
type PushDevices struct {
	models *metaengine.Store
	reader *metaengine.TypedReader[PushDeviceView]
	mu     sync.Mutex
}

func newPushDevices(models *metaengine.Store) *PushDevices {
	return &PushDevices{
		models: models,
		reader: metaengine.NewReader[PushDeviceView](models, pushDevicesCollection),
	}
}

// Upsert registers (or re-registers) one device for the owner. A known
// device id replaces its transport/token row (rotation); a NEW device
// id past the per-owner cap is refused with the ErrListFull contract.
func (s *PushDevices) Upsert(ctx context.Context, owner domain.Extension, deviceID, transport, token string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := domain.PushDeviceKey{Owner: owner.String(), DeviceID: deviceID}
	_, found, err := s.reader.Get(ctx, key)
	if err != nil {
		return errorfamily.WrapInfrastructuref(err, "readmodels.push_upsert", "load device slot")
	}
	if !found {
		if err := s.ownerHasRoom(ctx, owner); err != nil {
			return err
		}
	}

	if err := s.models.Apply(ctx, metaengine.EventTypeName(domain.PushDeviceRegistered{}), domain.PushDeviceRegistered{
		Key:       key,
		Transport: transport,
		Token:     token,
		UpdatedAt: time.Now(),
	}); err != nil {
		return errorfamily.WrapInfrastructuref(err, "readmodels.push_upsert", "apply device registration")
	}
	return nil
}

// List returns the owner's registered devices, newest registration
// first — the ring lane's fan-out source.
func (s *PushDevices) List(ctx context.Context, owner domain.Extension) ([]PushDeviceView, error) {
	views, err := s.ownerViews(ctx, owner)
	if err != nil {
		return nil, err
	}
	return views, nil
}

// Remove drops one device scoped to its owner; removing an unknown id
// is inert (the deregistration outcome is the same).
func (s *PushDevices) Remove(ctx context.Context, owner domain.Extension, deviceID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := domain.PushDeviceKey{Owner: owner.String(), DeviceID: deviceID}
	_, found, err := s.reader.Get(ctx, key)
	if err != nil {
		return errorfamily.WrapInfrastructuref(err, "readmodels.push_remove", "load device slot")
	}
	if !found {
		return nil
	}
	if err := s.models.Apply(ctx, metaengine.EventTypeName(domain.PushDeviceRemoved{}), domain.PushDeviceRemoved{Key: key}); err != nil {
		return errorfamily.WrapInfrastructuref(err, "readmodels.push_remove", "apply device removal")
	}
	return nil
}

// ownerHasRoom rejects a NEW device when the owner is at the cap.
func (s *PushDevices) ownerHasRoom(ctx context.Context, owner domain.Extension) error {
	views, err := s.ownerViews(ctx, owner)
	if err != nil {
		return err
	}
	if len(views) >= PushDevicesMaxPerExtension {
		return errorfamily.WrapRejectionf(ErrListFull, "readmodels.push_list_full",
			"%s at %d push devices", owner.String(), PushDevicesMaxPerExtension)
	}
	return nil
}

func (s *PushDevices) ownerViews(ctx context.Context, owner domain.Extension) ([]PushDeviceView, error) {
	// WithLimit is load-bearing: the cap bounds this scan at
	// PushDevicesMaxPerExtension rows.
	views, err := s.reader.Scan(ctx,
		metaengine.WithFilter("owner", metaengine.FilterEq, owner.String()),
		metaengine.WithSort("updated_at", true),
		metaengine.WithLimit(PushDevicesMaxPerExtension+1),
	)
	if err != nil {
		return nil, errorfamily.WrapInfrastructuref(err, "readmodels.push_list", "list push devices for %s", owner.String())
	}
	return views, nil
}
