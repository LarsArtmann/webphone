package server

import (
	"context"
	"encoding/json/v2"
	"fmt"
	"net/http"
	"testing"

	"github.com/larsartmann/webphone/internal/push"
	"github.com/larsartmann/webphone/internal/readmodels"
)

// pushStubTransport satisfies push.Sender for the API surface: the
// registry endpoints never send, they only consult the fanout's Knows.
type pushStubTransport struct{}

func (pushStubTransport) Enabled() bool                                 { return true }
func (pushStubTransport) Describe() string                              { return "stub" }
func (pushStubTransport) Send(context.Context, string, push.Ring) error { return nil }

func newPushServer(t *testing.T) *testServer {
	t.Helper()
	engine := newTestEngine(t)
	fanout := push.NewFanout(map[string]push.Sender{
		push.TransportNtfy: pushStubTransport{},
		push.TransportFcm:  pushStubTransport{},
	})
	return newTestServerWithPhoneAPI(t, "", func(d *Deps) {
		d.Push = fanout
		d.PushDevices = engine.PushDevices()
	})
}

type apiPushDevices struct {
	Devices []struct {
		DeviceID  string `json:"device_id"`
		Transport string `json:"transport"`
		UpdatedAt int64  `json:"updated_at"`
	} `json:"devices"`
}

func listPushDevices(t *testing.T, c *client) apiPushDevices {
	t.Helper()
	resp, body := c.do(http.MethodGet, "/api/push-devices", nil, "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("list push devices: %d %s", resp.StatusCode, body)
	}
	var got apiPushDevices
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatalf("list push devices body %q: %v", body, err)
	}
	return got
}

func registerPushDevice(t *testing.T, c *client, deviceID, transport, token string) *http.Response {
	t.Helper()
	body, err := json.Marshal(map[string]string{
		"device_id": deviceID,
		"transport": transport,
		"token":     token,
	})
	if err != nil {
		t.Fatal(err)
	}
	resp, respBody := c.do(http.MethodPost, "/api/push-devices", body, "application/json")
	if resp.StatusCode >= 500 {
		t.Fatalf("register push device: %d %s", resp.StatusCode, respBody)
	}
	return resp
}

func TestPushDeviceRegistryLifecycle(t *testing.T) {
	server := newPushServer(t)
	c := clientFor(t, server)
	c.login("1001", "pass")

	if resp := registerPushDevice(t, c, "device-1111-aaaa", "ntfy", "wp-8f2a-topic"); resp.StatusCode != http.StatusNoContent {
		t.Fatalf("register status %d, want 204", resp.StatusCode)
	}

	got := listPushDevices(t, c)
	if len(got.Devices) != 1 || got.Devices[0].DeviceID != "device-1111-aaaa" || got.Devices[0].Transport != "ntfy" {
		t.Fatalf("list = %+v", got.Devices)
	}
	if got.Devices[0].UpdatedAt == 0 {
		t.Fatal("list must carry the registration timestamp")
	}

	// Rotation replaces, never grows.
	if resp := registerPushDevice(t, c, "device-1111-aaaa", "fcm", "fcm-token-xyz"); resp.StatusCode != http.StatusNoContent {
		t.Fatalf("rotate status %d", resp.StatusCode)
	}
	got = listPushDevices(t, c)
	if len(got.Devices) != 1 || got.Devices[0].Transport != "fcm" {
		t.Fatalf("after rotation = %+v", got.Devices)
	}

	// Deregister is idempotent-inert.
	for range 2 {
		resp, _ := c.do(http.MethodDelete, "/api/push-devices?id=device-1111-aaaa", nil, "")
		if resp.StatusCode != http.StatusNoContent {
			t.Fatalf("deregister status %d, want 204", resp.StatusCode)
		}
	}
	got = listPushDevices(t, c)
	if len(got.Devices) != 0 {
		t.Fatalf("after deregister = %+v", got.Devices)
	}
}

func TestPushDeviceRegistryValidation(t *testing.T) {
	server := newPushServer(t)
	c := clientFor(t, server)
	c.login("1001", "pass")

	for _, tc := range []struct {
		name       string
		deviceID   string
		transport  string
		token      string
		wantStatus int
	}{
		{"unknown transport", "device-1111-aaaa", "apns", "t", http.StatusUnprocessableEntity},
		{"short device id", "short", "ntfy", "wp-topic", http.StatusUnprocessableEntity},
		{"junk device id", "device 1111 !!!", "ntfy", "wp-topic", http.StatusUnprocessableEntity},
		{"empty token", "device-1111-aaaa", "ntfy", "", http.StatusUnprocessableEntity},
		{"topic with slash", "device-1111-aaaa", "ntfy", "wp/topic", http.StatusUnprocessableEntity},
		{"valid ntfy", "device-1111-aaaa", "ntfy", "wp-topic", http.StatusNoContent},
		{"valid fcm", "device-2222-bbbb", "fcm", "dKj3:APA91b-token-shape", http.StatusNoContent},
	} {
		if resp := registerPushDevice(t, c, tc.deviceID, tc.transport, tc.token); resp.StatusCode != tc.wantStatus {
			t.Fatalf("%s: status %d, want %d", tc.name, resp.StatusCode, tc.wantStatus)
		}
	}
}

func TestPushDeviceRegistryCap(t *testing.T) {
	server := newPushServer(t)
	c := clientFor(t, server)
	c.login("1001", "pass")

	for i := 0; i < readmodels.PushDevicesMaxPerExtension; i++ {
		deviceID := fmt.Sprintf("device-%d", i)
		if resp := registerPushDevice(t, c, deviceID, "ntfy", "wp-topic"); resp.StatusCode != http.StatusNoContent {
			t.Fatalf("register %d: status %d", i, resp.StatusCode)
		}
	}
	if resp := registerPushDevice(t, c, "device-overflow", "ntfy", "wp-topic"); resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("overflow status %d, want 422", resp.StatusCode)
	}
}

func TestPushDeviceRegistryRequiresSession(t *testing.T) {
	server := newPushServer(t)
	anon := clientFor(t, server)

	for _, tc := range []struct{ method, path string }{
		{http.MethodGet, "/api/push-devices"},
		{http.MethodPost, "/api/push-devices"},
		{http.MethodDelete, "/api/push-devices?id=x"},
	} {
		resp, _ := anon.do(tc.method, tc.path, nil, "")
		if resp.StatusCode != http.StatusUnauthorized {
			t.Fatalf("%s %s: status %d, want 401", tc.method, tc.path, resp.StatusCode)
		}
	}
}

func TestPushDeviceRegistryOwnerScoped(t *testing.T) {
	server := newPushServer(t)
	alice := clientFor(t, server)
	alice.login("1001", "pass")
	bob := clientFor(t, server)
	bob.login("1002", "pass")

	if resp := registerPushDevice(t, alice, "device-1111-aaaa", "ntfy", "wp-topic"); resp.StatusCode != http.StatusNoContent {
		t.Fatal("alice register failed")
	}

	// Bob sees only his own (none).
	if got := listPushDevices(t, bob); len(got.Devices) != 0 {
		t.Fatalf("bob sees alice's devices: %+v", got.Devices)
	}

	// Bob's deregistration of alice's id is a plain inert 204.
	resp, _ := bob.do(http.MethodDelete, "/api/push-devices?id=device-1111-aaaa", nil, "")
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("cross-owner delete status %d", resp.StatusCode)
	}
	if got := listPushDevices(t, alice); len(got.Devices) != 1 {
		t.Fatalf("cross-owner delete removed alice's device: %+v", got.Devices)
	}
}

func TestPushDeviceRegistrySeamOffAnswers404(t *testing.T) {
	// The standard test server wires no push seam: the surface does not
	// exist (the shell must not believe registration succeeded).
	server := newTestServer(t)
	c := clientFor(t, server)
	c.login("1001", "pass")

	for _, tc := range []struct{ method, path string }{
		{http.MethodGet, "/api/push-devices"},
		{http.MethodPost, "/api/push-devices"},
	} {
		resp, _ := c.do(tc.method, tc.path, nil, "")
		if resp.StatusCode != http.StatusNotFound {
			t.Fatalf("%s %s: status %d, want 404", tc.method, tc.path, resp.StatusCode)
		}
	}
}
