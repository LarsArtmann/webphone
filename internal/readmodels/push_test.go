package readmodels

import (
	"context"
	"errors"
	"testing"

	errorfamilytest "github.com/larsartmann/go-error-family/errorfamilytest"
	"github.com/larsartmann/webphone/internal/domain"
	"github.com/larsartmann/webphone/internal/push"
)

func TestPushDevicesRegisterListRemove(t *testing.T) {
	engine := openTestEngine(t)
	devices := engine.PushDevices()
	ctx := context.Background()
	owner := domain.MustParseExtension("1001")

	if err := devices.Upsert(ctx, owner, "device-a", push.TransportNtfy, "topic-a"); err != nil {
		t.Fatalf("register a: %v", err)
	}
	if err := devices.Upsert(ctx, owner, "device-b", push.TransportFcm, "fcm-token-b"); err != nil {
		t.Fatalf("register b: %v", err)
	}

	// Owner scoping: another extension sees nothing.
	other := domain.MustParseExtension("1002")
	otherList, err := devices.List(ctx, other)
	if err != nil {
		t.Fatalf("list other: %v", err)
	}
	if len(otherList) != 0 {
		t.Fatalf("other owner sees %d devices, want 0", len(otherList))
	}

	list, err := devices.List(ctx, owner)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("list length %d, want 2", len(list))
	}

	// Rotation: re-registering the same device id replaces the row
	// (token churn must not grow the registry).
	if err := devices.Upsert(ctx, owner, "device-a", push.TransportNtfy, "topic-a-rotated"); err != nil {
		t.Fatalf("rotate a: %v", err)
	}
	list, err = devices.List(ctx, owner)
	if err != nil {
		t.Fatalf("list after rotation: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("list after rotation %d, want 2", len(list))
	}
	for _, view := range list {
		if view.DeviceID == "device-a" && view.Token != "topic-a-rotated" {
			t.Fatalf("rotation did not replace the token: %+v", view)
		}
	}

	// Remove: scoped to the owner, inert when unknown.
	if err := devices.Remove(ctx, owner, "device-a"); err != nil {
		t.Fatalf("remove a: %v", err)
	}
	if err := devices.Remove(ctx, owner, "device-a"); err != nil {
		t.Fatalf("remove a again: %v", err)
	}
	if err := devices.Remove(ctx, other, "device-b"); err != nil {
		t.Fatalf("remove other's device: %v", err)
	}
	list, err = devices.List(ctx, owner)
	if err != nil {
		t.Fatalf("list after remove: %v", err)
	}
	if len(list) != 1 || list[0].DeviceID != "device-b" {
		t.Fatalf("after remove: %+v", list)
	}
}

func TestPushDevicesCap(t *testing.T) {
	engine := openTestEngine(t)
	devices := engine.PushDevices()
	ctx := context.Background()
	owner := domain.MustParseExtension("1001")

	for i := range PushDevicesMaxPerExtension {
		if err := devices.Upsert(ctx, owner, "device-"+string(rune('a'+i)), push.TransportNtfy, "topic"); err != nil {
			t.Fatalf("register %d: %v", i, err)
		}
	}
	err := devices.Upsert(ctx, owner, "device-overflow", push.TransportNtfy, "topic")
	if !errors.Is(err, ErrListFull) {
		t.Fatalf("overflow error = %v, want ErrListFull", err)
	}
	errorfamilytest.AssertCode(t, err, "readmodels.push_list_full")

	// Rotation stays legal at the cap (a known device id replaces).
	if err := devices.Upsert(ctx, owner, "device-a", push.TransportFcm, "fcm-token"); err != nil {
		t.Fatalf("rotate at cap: %v", err)
	}
}
