package server

import (
	"testing"
)

func TestGhostProbeShape(t *testing.T) {
	server := newTestServer(t)
	resp, err := server.Client().Get(server.URL + "/api/ghost")
	if err != nil {
		t.Fatal(err)
	}
	body := readAll(t, resp)
	t.Logf("status=%d ct=%q body-head=%q", resp.StatusCode, resp.Header.Get("Content-Type"), string(body[:min(120, len(body))]))
}
