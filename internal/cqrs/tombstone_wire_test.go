package cqrs

import (
	"bytes"
	"encoding/json/v2"
	"reflect"
	"testing"

	"github.com/larsartmann/go-codec"
	"github.com/larsartmann/go-cqrs-lite/event/v4"
	"github.com/larsartmann/go-cqrs-lite/id/v4"
	"github.com/larsartmann/go-cqrs-lite/metaengine/v4"

	"github.com/larsartmann/webphone/internal/domain"
)

// The tombstone facts carry exact-name json tags (the D014 triage fix,
// 2026-11-11) that pin the journal's wire keys to the Go field names.
// The witnesses below mirror the PRE-tag shapes: identical exported
// fields, no tags. Encoding a witness with the journal's default codec
// and decoding the bytes through the fold's own DecodePayloadAuto path
// proves old journal rows survive the tagging; re-encoding the decoded
// value byte-identically proves the tags changed nothing on the wire.
type oldShapeThreadDeleted struct {
	Key           domain.ThreadRowKey
	Owner         string
	ThreadID      string
	Network       string
	Remote        string
	MessageIDs    []string
	AttachmentIDs []string
	Reason        string
}

type oldShapeContactDeleted struct {
	Key domain.ContactKey
}

type oldShapeSnippetDeleted struct {
	Key domain.SnippetKey
}

func TestTombstoneRowsReplayThroughDecodePayloadAuto(t *testing.T) {
	oldThread := oldShapeThreadDeleted{
		Key:           domain.ThreadRowKey{ThreadID: "th-1"},
		Owner:         "1001",
		ThreadID:      "th-1",
		Network:       string(domain.NetworkSMS),
		Remote:        "+15551234567",
		MessageIDs:    []string{"m1", "m2"},
		AttachmentIDs: []string{"a1"},
		Reason:        "retention",
	}
	wantThread := domain.ThreadDeleted{
		Key:           oldThread.Key,
		Owner:         oldThread.Owner,
		ThreadID:      oldThread.ThreadID,
		Network:       oldThread.Network,
		Remote:        oldThread.Remote,
		MessageIDs:    oldThread.MessageIDs,
		AttachmentIDs: oldThread.AttachmentIDs,
		Reason:        oldThread.Reason,
	}

	oldContact := oldShapeContactDeleted{Key: domain.ContactKey{Owner: "1001", Phone: "+15551234567"}}
	wantContact := domain.ContactDeleted{Key: oldContact.Key}

	oldSnippet := oldShapeSnippetDeleted{Key: domain.SnippetKey{ID: "snp-1"}}
	wantSnippet := domain.SnippetDeleted{Key: oldSnippet.Key}

	streamID, err := id.ParseStreamID("replay-probe")
	if err != nil {
		t.Fatalf("parse stream id: %v", err)
	}

	tests := []struct {
		name      string
		eventType event.Type
		old       any
		want      any
	}{
		{"thread tombstone", event.Type(metaengine.EventTypeName(domain.ThreadDeleted{})), oldThread, wantThread},
		{"contact tombstone", event.Type(metaengine.EventTypeName(domain.ContactDeleted{})), oldContact, wantContact},
		{"snippet tombstone", event.Type(metaengine.EventTypeName(domain.SnippetDeleted{})), oldSnippet, wantSnippet},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oldBytes, err := event.DefaultCodec.Encode(tt.old)
			if err != nil {
				t.Fatalf("encode old-shape row: %v", err)
			}
			evt, err := event.NewEvent(tt.eventType, streamID, ThreadsStreamType, event.Version(1), oldBytes, event.WithEncoding(codec.EncodingCBOR))
			if err != nil {
				t.Fatalf("build event: %v", err)
			}

			switch want := tt.want.(type) {
			case domain.ThreadDeleted:
				got, err := event.DecodePayloadAuto[domain.ThreadDeleted](evt)
				if err != nil {
					t.Fatalf("decode old-shape row: %v", err)
				}
				if !reflect.DeepEqual(got, want) {
					t.Fatalf("decoded = %+v, want %+v", got, want)
				}
				assertReencodeByteNeutral(t, evt.Encoding(), got, oldBytes)
			case domain.ContactDeleted:
				got, err := event.DecodePayloadAuto[domain.ContactDeleted](evt)
				if err != nil {
					t.Fatalf("decode old-shape row: %v", err)
				}
				if !reflect.DeepEqual(got, want) {
					t.Fatalf("decoded = %+v, want %+v", got, want)
				}
				assertReencodeByteNeutral(t, evt.Encoding(), got, oldBytes)
			case domain.SnippetDeleted:
				got, err := event.DecodePayloadAuto[domain.SnippetDeleted](evt)
				if err != nil {
					t.Fatalf("decode old-shape row: %v", err)
				}
				if !reflect.DeepEqual(got, want) {
					t.Fatalf("decoded = %+v, want %+v", got, want)
				}
				assertReencodeByteNeutral(t, evt.Encoding(), got, oldBytes)
			default:
				t.Fatalf("unhandled fact type %T", tt.want)
			}
		})
	}
}

// assertReencodeByteNeutral re-encodes the decoded (tagged) value with the
// event's own codec and requires the bytes to equal the old-shape row: a
// tag that ever diverges from the field name fails here instead of on a
// live replay.
func assertReencodeByteNeutral(t *testing.T, encoding codec.Encoding, decoded any, oldBytes []byte) {
	t.Helper()
	c, err := codec.ForEncoding(codec.Encoding(encoding))
	if err != nil {
		t.Fatalf("resolve codec for %q: %v", encoding, err)
	}
	newBytes, err := c.Encode(decoded)
	if err != nil {
		t.Fatalf("encode tagged value: %v", err)
	}
	if !bytes.Equal(oldBytes, newBytes) {
		t.Fatalf("tagged re-encode differs from the old-shape row (%d vs %d bytes): the tags are not byte-neutral", len(oldBytes), len(newBytes))
	}
}

func TestTombstoneJSONKeysStayFieldNames(t *testing.T) {
	thread, _ := json.Marshal(domain.ThreadDeleted{
		Key:           domain.ThreadRowKey{ThreadID: "th-1"},
		Owner:         "1001",
		ThreadID:      "th-1",
		Network:       string(domain.NetworkSMS),
		Remote:        "+15551234567",
		MessageIDs:    []string{"m1", "m2"},
		AttachmentIDs: []string{"a1"},
		Reason:        "retention",
	})
	if got, want := string(thread), `{"Key":{"ThreadID":"th-1"},"Owner":"1001","ThreadID":"th-1","Network":"sms","Remote":"+15551234567","MessageIDs":["m1","m2"],"AttachmentIDs":["a1"],"Reason":"retention"}`; got != want {
		t.Fatalf("ThreadDeleted JSON keys drifted from the Go field names:\n got %s\nwant %s", got, want)
	}

	contact, _ := json.Marshal(domain.ContactDeleted{Key: domain.ContactKey{Owner: "1001", Phone: "+15551234567"}})
	if got, want := string(contact), `{"Key":{"Owner":"1001","Phone":"+15551234567"}}`; got != want {
		t.Fatalf("ContactDeleted JSON keys drifted from the Go field names:\n got %s\nwant %s", got, want)
	}

	snippet, _ := json.Marshal(domain.SnippetDeleted{Key: domain.SnippetKey{ID: "snp-1"}})
	if got, want := string(snippet), `{"Key":{"ID":"snp-1"}}`; got != want {
		t.Fatalf("SnippetDeleted JSON keys drifted from the Go field names:\n got %s\nwant %s", got, want)
	}
}
