// The push-devices aggregate (SUPERB Phase 3, A24): the ring lane's
// registered-device store as a per-owner stream. The 8-per-owner cap
// for NEW device ids and the rotation-replaces semantics live in the
// DECISION; the wake token stays write-only in exactly the same way
// (the projection row carries it for the server-side fan-out; the
// LIST wire shape never echoes it — that mapping stays at the API
// door). ErrListFull survives the swap verbatim.
package cqrs

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/larsartmann/go-cqrs-lite/command/v4"
	"github.com/larsartmann/go-cqrs-lite/decider/v4"
	"github.com/larsartmann/go-cqrs-lite/event/v4"
	"github.com/larsartmann/go-cqrs-lite/metaengine/v4"
	"github.com/larsartmann/go-cqrs-lite/record/v4"
	"github.com/larsartmann/go-cqrs-lite/system/v4"
	"github.com/larsartmann/go-error-family"

	"github.com/larsartmann/webphone/internal/domain"
	"github.com/larsartmann/webphone/internal/readmodels"
)

// PushDevicesStreamType keys per-owner device streams.
const PushDevicesStreamType = "push_devices"

// pushDevicesMaxPerOwner mirrors readmodels.PushDevicesMaxPerExtension
// at the aggregate.
const pushDevicesMaxPerOwner = readmodels.PushDevicesMaxPerExtension

// SavePushDevice registers (or re-registers) one device (the field
// name DeviceID, never ID — that would shadow the embedded command's
// ID() method).
type SavePushDevice struct {
	*command.BasicCommand
	Owner     string
	DeviceID  string
	Transport string
	Token     string
	Now       time.Time
}

// RemovePushDevice deregisters one device; an unknown id decides NO
// events (the deregistration outcome is the same).
type RemovePushDevice struct {
	*command.BasicCommand
	Owner    string
	DeviceID string
}

// pushDeviceSlot is one device's aggregate state.
type pushDeviceSlot struct {
	Transport string
	Token     string
	UpdatedAt int64 // unix nanos
}

// pushDevicesState is the owner's registered-device set.
type pushDevicesState struct {
	Devices map[string]pushDeviceSlot
}

func (pushDevicesState) initial() pushDevicesState {
	return pushDevicesState{Devices: map[string]pushDeviceSlot{}}
}

func applyPushDevicesState(s pushDevicesState, evt event.Event) (pushDevicesState, error) {
	switch evt.Type() {
	case event.Type(metaengine.EventTypeName(domain.PushDeviceRegistered{})):
		fact, err := event.DecodePayloadAuto[domain.PushDeviceRegistered](evt)
		if err != nil {
			return s, errorfamily.WrapCorruptionf(err, "cqrs.push_device_fold", "decode PushDeviceRegistered")
		}
		s.Devices[fact.Key.DeviceID] = pushDeviceSlot{Transport: fact.Transport, Token: fact.Token, UpdatedAt: fact.UpdatedAt.UnixNano()}
	case event.Type(metaengine.EventTypeName(domain.PushDeviceRemoved{})):
		fact, err := event.DecodePayloadAuto[domain.PushDeviceRemoved](evt)
		if err != nil {
			return s, errorfamily.WrapCorruptionf(err, "cqrs.push_device_fold", "decode PushDeviceRemoved")
		}
		delete(s.Devices, fact.Key.DeviceID)
	}
	return s, nil
}

func decideSavePushDevice(s pushDevicesState, ver event.Version, cmd SavePushDevice) ([]event.Event, error) {
	if _, known := s.Devices[cmd.DeviceID]; !known && len(s.Devices) >= pushDevicesMaxPerOwner {
		return nil, errorfamily.WrapRejectionf(readmodels.ErrListFull, "cqrs.push_device_list_full",
			"%s at %d push devices", cmd.Owner, pushDevicesMaxPerOwner)
	}
	evt, err := event.New(event.Type(metaengine.EventTypeName(domain.PushDeviceRegistered{})), cmd.StreamID(), PushDevicesStreamType, ver+1, domain.PushDeviceRegistered{
		Key:       domain.PushDeviceKey{Owner: cmd.Owner, DeviceID: cmd.DeviceID},
		Transport: cmd.Transport,
		Token:     cmd.Token,
		UpdatedAt: cmd.Now,
	})
	if err != nil {
		return nil, errorfamily.WrapInfrastructuref(err, "cqrs.push_device_save", "build PushDeviceRegistered")
	}
	return []event.Event{evt}, nil
}

func decideRemovePushDevice(s pushDevicesState, ver event.Version, cmd RemovePushDevice) ([]event.Event, error) {
	if _, known := s.Devices[cmd.DeviceID]; !known {
		return nil, nil
	}
	evt, err := event.New(event.Type(metaengine.EventTypeName(domain.PushDeviceRemoved{})), cmd.StreamID(), PushDevicesStreamType, ver+1, domain.PushDeviceRemoved{
		Key: domain.PushDeviceKey{Owner: cmd.Owner, DeviceID: cmd.DeviceID},
	})
	if err != nil {
		return nil, errorfamily.WrapInfrastructuref(err, "cqrs.push_device_remove", "build PushDeviceRemoved")
	}
	return []event.Event{evt}, nil
}

// RegisterPushDevices wires the push-devices aggregate into a System.
func RegisterPushDevices(sys *system.System) error {
	if err := system.RegisterDecider[pushDevicesState](sys, PushDevicesStreamType, decider.Decider[pushDevicesState]{
		Initial: pushDevicesState{}.initial(),
		Apply:   applyPushDevicesState,
	}); err != nil {
		return errorfamily.WrapInfrastructuref(err, "cqrs.push_device_register", "register push-devices decider")
	}
	if err := system.RegisterCommand[*SavePushDevice, pushDevicesState](sys, "push_device.save",
		func(_ context.Context, cmd *SavePushDevice) system.Op[pushDevicesState] {
			return system.Execute[pushDevicesState](context.Background(), cmd.StreamID(), PushDevicesStreamType,
				func(state pushDevicesState, ver event.Version) ([]event.Event, error) {
					return decideSavePushDevice(state, ver, *cmd)
				})
		}); err != nil {
		return errorfamily.WrapInfrastructuref(err, "cqrs.push_device_register", "register push_device.save")
	}
	if err := system.RegisterCommand[*RemovePushDevice, pushDevicesState](sys, "push_device.remove",
		func(_ context.Context, cmd *RemovePushDevice) system.Op[pushDevicesState] {
			return system.Execute[pushDevicesState](context.Background(), cmd.StreamID(), PushDevicesStreamType,
				func(state pushDevicesState, ver event.Version) ([]event.Event, error) {
					return decideRemovePushDevice(state, ver, *cmd)
				})
		}); err != nil {
		return errorfamily.WrapInfrastructuref(err, "cqrs.push_device_register", "register push_device.remove")
	}
	return nil
}

// pushDevicesProjection folds the device facts into the SAME view
// shape the fan-out reads today (readmodels.PushDeviceView).
func pushDevicesProjection() metaengine.QueryDecl[pushDevicesKeyInput, readmodels.PushDeviceView] {
	return metaengine.Query[pushDevicesKeyInput, readmodels.PushDeviceView](pushDevicesCollection,
		metaengine.OnRecord(domain.PushDeviceRegistered{},
			func(_ record.Record, e domain.PushDeviceRegistered) (domain.PushDeviceKey, readmodels.PushDeviceView) {
				return e.Key, readmodels.PushDeviceView{
					Owner:     e.Key.Owner,
					DeviceID:  e.Key.DeviceID,
					Transport: e.Transport,
					Token:     e.Token,
					UpdatedAt: e.UpdatedAt.UnixNano(),
				}
			}),
		metaengine.OnRecord(domain.PushDeviceRemoved{}, metaengine.Remove[readmodels.PushDeviceView]()),
		metaengine.FilterOnField[readmodels.PushDeviceView]("owner", metaengine.FilterEq),
		metaengine.SortOnField[readmodels.PushDeviceView]("updated_at", true),
	)
}

// pushDevicesCollection names the system engine's projection (the
// same name the readmodels engine declares — different databases, one
// shape).
const pushDevicesCollection = "push_devices"

// pushDevicesKeyInput is the projection's key-extraction input (the
// same shape the readmodels engine declares).
type pushDevicesKeyInput struct {
	Owner    string
	DeviceID string
}

// PushDevicesDomainDecl is the push-devices aggregate's share of the
// system's DomainConfig (a registration failure is a boot-time miswire
// and panics loudly).
func PushDevicesDomainDecl() system.DomainConfig {
	return system.DomainConfig{
		Commands: func(sys *system.System) {
			if err := RegisterPushDevices(sys); err != nil {
				panic(fmt.Sprintf("cqrs: register push-devices aggregate: %v", err))
			}
		},
		Projections: []system.ProjectionDeclaration{system.RawQuery(pushDevicesProjection())},
		Events: []event.Type{
			event.Type(metaengine.EventTypeName(domain.PushDeviceRegistered{})),
			event.Type(metaengine.EventTypeName(domain.PushDeviceRemoved{})),
		},
	}
}

// ─── the backend the server keeps calling ──────────────────────────

// PushDevices is the dispatch-backed registered-device store: the
// same contract the engine-backed store exposed (cap-bounded upsert
// with rotation-replaces, newest-first List, inert Remove).
type PushDevices struct {
	sys    *system.System
	reader *metaengine.TypedReader[readmodels.PushDeviceView]
	mu     sync.Mutex
}

var _ readmodels.PushDeviceGate = (*PushDevices)(nil)

// NewPushDevices builds the backend over a System.
func NewPushDevices(sys *system.System) *PushDevices {
	return &PushDevices{sys: sys, reader: metaengine.NewReader[readmodels.PushDeviceView](sys.MetaEngine(), pushDevicesCollection)}
}

// Upsert registers (or re-registers) one device for the owner: a
// known id replaces its transport/token row (rotation, cap-free); a
// NEW id past the per-owner cap answers the ErrListFull rejection.
//
// brandid-lint:ignore(BD003) store-seam transport: the client-generated device id keys engine rows (strong-id wire ruling)
func (s *PushDevices) Upsert(ctx context.Context, owner domain.Extension, deviceID, transport, token string) error { //nolint:branching-flow // verbatim port of the retired engine store's Upsert signature (parity rule)
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := dispatchOwnerCommand(ctx, s.sys, "push_device.save", owner.String(),
		func(base *command.BasicCommand) *SavePushDevice {
			return &SavePushDevice{
				BasicCommand: base,
				Owner:        owner.String(),
				DeviceID:     deviceID,
				Transport:    transport,
				Token:        token,
				Now:          time.Now(),
			}
		}); err != nil {
		if errors.Is(err, readmodels.ErrListFull) {
			return err
		}
		return errorfamily.WrapInfrastructuref(err, "cqrs.push_device_save", "dispatch push_device.save")
	}
	return nil
}

// List returns the owner's registered devices, newest registration
// first — the ring lane's fan-out source.
func (s *PushDevices) List(ctx context.Context, owner domain.Extension) ([]readmodels.PushDeviceView, error) {
	views, err := s.reader.Scan(ctx,
		metaengine.WithFilter("owner", metaengine.FilterEq, owner.String()),
		metaengine.WithSort("updated_at", true),
		metaengine.WithLimit(pushDevicesMaxPerOwner+1),
	)
	if err != nil {
		return nil, errorfamily.WrapInfrastructuref(err, "cqrs.push_device_list", "list push devices for %s", owner.String())
	}
	return views, nil
}

// Remove drops one device scoped to its owner; removing an unknown id
// is inert (the deregistration outcome is the same).
//
// brandid-lint:ignore(BD003) store-seam transport: the client-generated device id keys engine rows (strong-id wire ruling)
func (s *PushDevices) Remove(ctx context.Context, owner domain.Extension, deviceID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := dispatchOwnerCommand(ctx, s.sys, "push_device.remove", owner.String(),
		func(base *command.BasicCommand) *RemovePushDevice {
			return &RemovePushDevice{BasicCommand: base, Owner: owner.String(), DeviceID: deviceID}
		}); err != nil {
		return errorfamily.WrapInfrastructuref(err, "cqrs.push_device_remove", "dispatch push_device.remove")
	}
	return nil
}
