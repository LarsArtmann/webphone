package domain

import "time"

// PushDeviceKey identifies one registered push device in the read-model
// store: the owning extension plus the app-generated per-install device
// id. The device id is stable across token rotations (an FCM
// registration token churns; the identity must not), so re-registering
// the same install replaces its address instead of growing a row.
type PushDeviceKey struct {
	Owner    string
	DeviceID string
}

// PushDeviceRegistered is the read-model fact inserting or replacing
// one device row: which transport wakes it ("ntfy" or "fcm") and the
// device's address on that transport (the ntfy topic or the FCM
// registration token).
//
// brandid-lint:ignore(BD003) read-model fact: wire shape; the transport
// is a foreign protocol vocabulary and the token a foreign credential,
// neither owns a domain brand (strong-id wire ruling).
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
