package pbx

import (
	"context"
	"encoding/base64"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func newUpstream(t *testing.T, handler http.HandlerFunc) (*Client, *httptest.Server) {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	client, err := NewClient(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	return client, srv
}

var creds = Credentials{Extension: "1001", Password: "s3cret"}

func TestDisabledClientReturnsErrDisabled(t *testing.T) {
	client, err := NewClient("")
	if err != nil {
		t.Fatal(err)
	}
	if client.Enabled() {
		t.Fatal("empty base URL must mean disabled")
	}
	ctx := context.Background()
	if _, err := client.History(ctx, creds, 30); !errors.Is(err, ErrDisabled) {
		t.Errorf("History: %v", err)
	}
	if _, err := client.VoicemailSummary(ctx, creds); !errors.Is(err, ErrDisabled) {
		t.Errorf("VoicemailSummary: %v", err)
	}
	if _, err := client.VoicemailMessages(ctx, creds); !errors.Is(err, ErrDisabled) {
		t.Errorf("VoicemailMessages: %v", err)
	}
	if err := client.DeleteVoicemail(ctx, creds, "uuid-1"); !errors.Is(err, ErrDisabled) {
		t.Errorf("DeleteVoicemail: %v", err)
	}
}

// The Enabled guard is nil-receiver-safe, so a nil *Client (a Deps field
// nobody wired) short-circuits through do() to ErrDisabled instead of
// panicking on the base-URL deref.
func TestNilClientReturnsErrDisabled(t *testing.T) {
	var client *Client
	if client.Enabled() {
		t.Fatal("nil client must report disabled")
	}
	ctx := context.Background()
	if _, err := client.History(ctx, creds, 30); !errors.Is(err, ErrDisabled) {
		t.Errorf("History: %v", err)
	}
	if _, err := client.VoicemailSummary(ctx, creds); !errors.Is(err, ErrDisabled) {
		t.Errorf("VoicemailSummary: %v", err)
	}
	if _, err := client.VoicemailMessages(ctx, creds); !errors.Is(err, ErrDisabled) {
		t.Errorf("VoicemailMessages: %v", err)
	}
	if err := client.DeleteVoicemail(ctx, creds, "uuid-1"); !errors.Is(err, ErrDisabled) {
		t.Errorf("DeleteVoicemail: %v", err)
	}
	if err := client.VerifyCredentials(ctx, creds); !errors.Is(err, ErrDisabled) {
		t.Errorf("VerifyCredentials: %v", err)
	}
}

func TestNewClientRejectsInvalidBaseURL(t *testing.T) {
	if _, err := NewClient("://no-scheme"); err == nil {
		t.Fatal("invalid base URL must be an error")
	}
}

func TestVerifyCredentials(t *testing.T) {
	t.Run("accepts valid directory credentials", func(t *testing.T) {
		client, _ := newUpstream(t, func(w http.ResponseWriter, r *http.Request) {
			user, pass, ok := r.BasicAuth()
			if !ok || user != creds.Extension || pass != creds.Password {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"new":1,"old":2}`))
		})
		if err := client.VerifyCredentials(context.Background(), creds); err != nil {
			t.Fatalf("valid credentials rejected: %v", err)
		}
	})
	t.Run("rejects invalid directory credentials with sentinel", func(t *testing.T) {
		client, _ := newUpstream(t, func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
		})
		if err := client.VerifyCredentials(context.Background(), creds); !errors.Is(err, ErrUnauthorized) {
			t.Fatalf("want ErrUnauthorized, got %v", err)
		}
	})
	t.Run("maps 403 to the sentinel too", func(t *testing.T) {
		client, _ := newUpstream(t, func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusForbidden)
		})
		if err := client.VerifyCredentials(context.Background(), creds); !errors.Is(err, ErrUnauthorized) {
			t.Fatalf("want ErrUnauthorized, got %v", err)
		}
	})
	t.Run("surfaces PBX outages as non-sentinel errors", func(t *testing.T) {
		client, _ := newUpstream(t, func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		})
		err := client.VerifyCredentials(context.Background(), creds)
		if err == nil || errors.Is(err, ErrUnauthorized) {
			t.Fatalf("outage must not look like bad credentials: %v", err)
		}
	})
	t.Run("disabled client reports ErrDisabled", func(t *testing.T) {
		client, err := NewClient("")
		if err != nil {
			t.Fatal(err)
		}
		if err := client.VerifyCredentials(context.Background(), creds); !errors.Is(err, ErrDisabled) {
			t.Fatalf("want ErrDisabled, got %v", err)
		}
	})
}

func TestHistorySendsBasicAuthAndDecodes(t *testing.T) {
	var gotPath, gotAuth, gotMethod string
	client, _ := newUpstream(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath, gotAuth, gotMethod = r.URL.RequestURI(), r.Header.Get("Authorization"), r.Method
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"entries":[{"caller_id_number":"+441632960961","destination_number":"1001","start":"2026-09-18 10:00","billsec":42}]}`))
	})

	page, err := client.History(context.Background(), creds, 30)
	if err != nil {
		t.Fatal(err)
	}
	if gotMethod != http.MethodGet {
		t.Errorf("method: %s", gotMethod)
	}
	if gotPath != "/phone-api/history?limit=30" {
		t.Errorf("request URI: %s", gotPath)
	}
	wantAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte("1001:s3cret"))
	if gotAuth != wantAuth {
		t.Errorf("authorization: got %q, want %q", gotAuth, wantAuth)
	}
	if len(page.Entries) != 1 || page.Entries[0].CallerIDNumber != "+441632960961" || page.Entries[0].Billsec != 42 {
		t.Errorf("decoded entries: %+v", page.Entries)
	}
}

func TestVoicemailEndpointsHitTheirPaths(t *testing.T) {
	var paths []string
	client, _ := newUpstream(t, func(w http.ResponseWriter, r *http.Request) {
		paths = append(paths, r.URL.Path)
		_, _ = w.Write([]byte(`{}`))
	})
	ctx := context.Background()

	if _, err := client.VoicemailSummary(ctx, creds); err != nil {
		t.Fatal(err)
	}
	if _, err := client.VoicemailMessages(ctx, creds); err != nil {
		t.Fatal(err)
	}
	if err := client.DeleteVoicemail(ctx, creds, "uuid-9"); err != nil {
		t.Fatal(err)
	}

	want := []string{
		"/phone-api/voicemail/1001/summary",
		"/phone-api/voicemail/1001/messages",
		"/phone-api/voicemail/1001/messages/uuid-9",
	}
	if len(paths) != len(want) {
		t.Fatalf("requests: %v", paths)
	}
	for i, path := range want {
		if paths[i] != path {
			t.Errorf("request %d: got %s, want %s", i, paths[i], path)
		}
	}
}

func TestErrorPaths(t *testing.T) {
	cases := []struct {
		name    string
		status  int
		body    string
		wantErr []string
	}{
		{name: "credentials rejected", status: http.StatusUnauthorized, body: "no", wantErr: []string{"rejected the credentials"}},
		{name: "server error", status: http.StatusInternalServerError, body: "boom", wantErr: []string{"HTTP 500"}},
		{name: "malformed response", status: http.StatusOK, body: `{"entries":[`, wantErr: []string{"decode phone api response"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			client, _ := newUpstream(t, func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(tc.status)
				_, _ = w.Write([]byte(tc.body))
			})
			_, err := client.History(context.Background(), creds, 30)
			if err == nil {
				t.Fatal("expected an error")
			}
			for _, want := range tc.wantErr {
				if !strings.Contains(err.Error(), want) {
					t.Errorf("error %q does not contain %q", err.Error(), want)
				}
			}
		})
	}
}

func TestTransportFailureIsWrapped(t *testing.T) {
	client, _ := newUpstream(t, func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(2 * time.Second)
	})
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	if _, err := client.History(ctx, creds, 30); err == nil || !strings.Contains(err.Error(), "phone api call") {
		t.Fatalf("transport failure must be wrapped as phone api call error, got %v", err)
	}
}

func TestResolvePathKeepsQuery(t *testing.T) {
	client, err := NewClient("https://pbx.example.com")
	if err != nil {
		t.Fatal(err)
	}
	got := client.ResolvePath("/voicemail/1001/messages", "limit=5")
	if got != "https://pbx.example.com/phone-api/voicemail/1001/messages?limit=5" {
		t.Errorf("resolved %q", got)
	}
}
