// The compensating-delete pins (TODO blob-orphan row, cheap half): a
// message whose persist fails must leave NO file in the blob tree —
// retention walks rows, so an orphaned attachment is unreachable
// forever. Both lanes pin the end state: Send and Receive each save
// their attachments, the gate refuses the append, and the store root
// must be empty again.
package messaging_test

import (
	"context"
	"errors"
	"io/fs"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/larsartmann/webphone/internal/blob"
	"github.com/larsartmann/webphone/internal/cqrs"
	"github.com/larsartmann/webphone/internal/domain"
	"github.com/larsartmann/webphone/internal/messaging"
	"github.com/larsartmann/webphone/internal/readmodels"
)

// failingAppendGate answers every FindThread through the real journal
// backend but refuses the append — the persist-failure arm without a
// full disk.
type failingAppendGate struct {
	readmodels.ThreadGate
}

func newFailingAppendGate(t *testing.T) *failingAppendGate {
	t.Helper()
	sys, err := cqrs.NewThreadsSystem(context.Background(), "file:"+filepath.Join(t.TempDir(), "system.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = sys.Close() })
	return &failingAppendGate{ThreadGate: cqrs.NewThreads(sys)}
}

func (g *failingAppendGate) AppendMessage(_ context.Context, _ domain.Message) error {
	return errors.New("persist down")
}

// blobFileCount walks the store root and counts regular files — zero
// after a compensated failure is the whole assertion.
func blobFileCount(t *testing.T, root string) int {
	t.Helper()
	count := 0
	err := filepath.WalkDir(root, func(_ string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.Type().IsRegular() {
			count++
		}
		return nil
	})
	if err != nil && !strings.Contains(err.Error(), "no such file") {
		t.Fatalf("walk blob root: %v", err)
	}
	return count
}

func TestFailedSendLeavesNoOrphanedBlobs(t *testing.T) {
	gate := newFailingAppendGate(t)
	blobs, err := blob.New(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	svc := messaging.New(gate, blobs, nil, nil, nil)

	owner := domain.MustParseExtension("100")
	remote := domain.MustParseAddress(domain.NetworkSMS, "+491711")

	_, err = svc.Send(context.Background(), owner, remote, "with attachments", []domain.AttachmentContent{
		{Name: "a.txt", MimeType: "text/plain", Bytes: []byte("first")},
		{Name: "b.txt", MimeType: "text/plain", Bytes: []byte("second")},
	})
	if err == nil || !strings.Contains(err.Error(), "persist down") {
		t.Fatalf("Send verdict = %v, want the persist failure", err)
	}
	if n := blobFileCount(t, blobs.Root()); n != 0 {
		t.Fatalf("%d blob files survived a failed persist — orphans retention can never find", n)
	}
}

func TestFailedReceiveLeavesNoOrphanedBlobs(t *testing.T) {
	gate := newFailingAppendGate(t)
	blobs, err := blob.New(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	svc := messaging.New(gate, blobs, nil, nil, nil)

	_, err = svc.Receive(context.Background(), domain.InboundMessage{
		Owner:      domain.MustParseExtension("100"),
		From:       domain.MustParseAddress(domain.NetworkSMS, "+491712"),
		Body:       "inbound with media",
		ReceivedAt: time.Now(),
		Attachments: []domain.AttachmentContent{
			{Name: "pic.jpg", MimeType: "image/jpeg", Bytes: []byte("jpeg-bytes")},
		},
	})
	if err == nil || !strings.Contains(err.Error(), "persist down") {
		t.Fatalf("Receive verdict = %v, want the persist failure", err)
	}
	if n := blobFileCount(t, blobs.Root()); n != 0 {
		t.Fatalf("%d blob files survived a failed inbound persist — orphans retention can never find", n)
	}
}
