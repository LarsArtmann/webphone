package domain

import "time"

// PushDeviceKey identifies one registered push device in the read-model
// store: the owning extension plus the app-generated per-install device
// id. The device id is stable across token rotations (an FCM
// registration token churns; the identity must not), so re-registering
// the same install replaces its address instead of growing a row.
// brandid-lint:ignore(BD003) the device id is the shell's client-generated install identity (a UUID), not a domain entity (strong-id wire ruling)
type PushDeviceKey struct {
	Owner    string
	DeviceID string
}

// PushDeviceRegistered is the read-model fact inserting or replacing
// one device row: which transport wakes it ("ntfy" or "fcm") and the
// device's address on that transport (the ntfy topic or the FCM
// registration token).
type PushDeviceRegistered struct {
	Key       PushDeviceKey
	Transport string
	Token     string
	UpdatedAt time.Time
}

// PushDeviceRemoved is the read-model fact removing one device (the
// shell deregisters on logout; the ring lane prunes dead tokens).
type PushDeviceRemoved struct {
	Key PushDeviceKey
}
