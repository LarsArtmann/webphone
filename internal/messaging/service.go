// Package messaging is the application service for SMS/MMS threads: it
// validates sends, persists messages and attachments, drives the outbound
// gateway, ingests inbound traffic from webhooks, and notifies listeners
// (SSE) about every change.
package messaging

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"path/filepath"
	"strings"
	"time"

	"github.com/larsartmann/webphone/internal/blob"
	"github.com/larsartmann/webphone/internal/domain"
	"github.com/larsartmann/webphone/internal/gateway"
	"github.com/larsartmann/webphone/internal/store"
)

// Limits keep uploads and bodies honest without a config knob each.
const (
	MaxAttachments    = 5
	MaxAttachmentSize = 10 << 20 // 10 MiB per attachment
	MaxBodyLength     = 1600
	messagePageSize   = 200
)

// ErrInvalidSend describes a rejected send with a user-facing message.
type ErrInvalidSend struct{ Reason string }

func (e *ErrInvalidSend) Error() string { return e.Reason }

// Upload is one attachment as it arrives from the composer form.
type Upload struct {
	Name     string
	MimeType string
	Bytes    []byte
}

// ChangeFunc is called after any thread mutation with the owning
// extension; listeners re-render the affected thread (SSE fan-out in the
// server layer).
type ChangeFunc func(ctx context.Context, owner domain.Extension, threadID domain.ThreadID)

// Service wires the message store, blob store, and outbound gateway.
type Service struct {
	messages *store.Messages
	blobs    *blob.Store
	gateway  gateway.MessageGateway
	onChange ChangeFunc
	clock    func() time.Time
}

// New builds the messaging service. onChange may be nil.
func New(messages *store.Messages, blobs *blob.Store, gw gateway.MessageGateway, onChange ChangeFunc) *Service {
	return &Service{
		messages: messages,
		blobs:    blobs,
		gateway:  gw,
		onChange: onChange,
		clock:    time.Now,
	}
}

// Send validates and delivers one message. It returns the persisted message
// (already carrying its final status) and the thread it belongs to.
func (s *Service) Send(
	ctx context.Context, owner domain.Extension, to domain.Phone, body string, uploads []Upload,
) (domain.Message, error) {
	body = strings.TrimSpace(body)
	if body == "" && len(uploads) == 0 {
		return domain.Message{}, &ErrInvalidSend{Reason: "a message needs text or an attachment"}
	}
	if len(body) > MaxBodyLength {
		return domain.Message{}, &ErrInvalidSend{Reason: fmt.Sprintf("message longer than %d characters", MaxBodyLength)}
	}
	if len(uploads) > MaxAttachments {
		return domain.Message{}, &ErrInvalidSend{Reason: fmt.Sprintf("at most %d attachments per message", MaxAttachments)}
	}
	for _, upload := range uploads {
		if len(upload.Bytes) > MaxAttachmentSize {
			return domain.Message{}, &ErrInvalidSend{
				Reason: fmt.Sprintf("attachment %s larger than %d MiB", upload.Name, MaxAttachmentSize>>20),
			}
		}
	}

	now := s.clock()
	threadID, err := s.messages.FindThread(ctx, owner, to, now)
	if err != nil {
		return domain.Message{}, fmt.Errorf("resolve thread: %w", err)
	}

	msg := domain.Message{
		ID:        domain.GenerateMessageID(),
		ThreadID:  threadID,
		Owner:     owner,
		Remote:    to,
		Direction: domain.DirectionOutbound,
		Channel:   domain.ChannelOf(body, len(uploads)),
		Body:      body,
		Status:    domain.StatusQueued,
		CreatedAt: now,
	}

	outbound := gateway.OutboundMessage{Owner: owner, To: to, Body: body}
	for _, upload := range uploads {
		path, err := s.blobs.Save("attachments", extensionOf(upload), upload.Bytes)
		if err != nil {
			return domain.Message{}, fmt.Errorf("store attachment: %w", err)
		}
		attachment := domain.Attachment{
			ID:        domain.GenerateAttachmentID(),
			MessageID: msg.ID,
			Name:      sanitizeFilename(upload.Name),
			MimeType:  upload.MimeType,
			SizeBytes: int64(len(upload.Bytes)),
			Path:      path,
		}
		msg.Attachments = append(msg.Attachments, attachment)
		outbound.Attachments = append(outbound.Attachments, gateway.OutboundAttachment{
			Name: attachment.Name, MimeType: attachment.MimeType, Path: s.blobs.Abs(path),
		})
	}

	if err := s.messages.AppendMessage(ctx, msg); err != nil {
		return domain.Message{}, fmt.Errorf("persist message: %w", err)
	}
	s.notify(ctx, owner, threadID)

	receipt, err := s.gateway.SendMessage(ctx, outbound)
	switch {
	case err != nil:
		msg.Status = domain.StatusFailed
		sendErr := s.messages.UpdateOutboundStatus(ctx, msg.ID, domain.StatusFailed, "")
		s.notify(ctx, owner, threadID)
		if sendErr != nil {
			slog.Warn("messaging: mark failed", "error", sendErr)
		}
		return msg, fmt.Errorf("gateway: %w", err)
	default:
		msg.Status = domain.StatusSent
		msg.ProviderRef = receipt.ProviderRef
		if err := s.messages.UpdateOutboundStatus(ctx, msg.ID, domain.StatusSent, receipt.ProviderRef); err != nil {
			slog.Warn("messaging: mark sent", "error", err)
		}
		s.notify(ctx, owner, threadID)
		return msg, nil
	}
}

// Receive ingests an inbound message from a gateway/webhook and returns
// its persisted form.
func (s *Service) Receive(ctx context.Context, inbound domain.InboundMessage) (domain.Message, error) {
	now := inbound.ReceivedAt
	if now.IsZero() {
		now = s.clock()
	}
	threadID, err := s.messages.FindThread(ctx, inbound.Owner, inbound.From, now)
	if err != nil {
		return domain.Message{}, fmt.Errorf("resolve thread: %w", err)
	}

	msg := domain.Message{
		ID:        domain.GenerateMessageID(),
		ThreadID:  threadID,
		Owner:     inbound.Owner,
		Remote:    inbound.From,
		Direction: domain.DirectionInbound,
		Channel:   domain.ChannelOf(inbound.Body, len(inbound.Attachments)),
		Body:      inbound.Body,
		CreatedAt: now,
	}
	for _, att := range inbound.Attachments {
		path, err := s.blobs.Save("attachments", extensionOf(Upload(att)), att.Bytes)
		if err != nil {
			return domain.Message{}, fmt.Errorf("store inbound attachment: %w", err)
		}
		msg.Attachments = append(msg.Attachments, domain.Attachment{
			ID:        domain.GenerateAttachmentID(),
			MessageID: msg.ID,
			Name:      sanitizeFilename(att.Name),
			MimeType:  att.MimeType,
			SizeBytes: int64(len(att.Bytes)),
			Path:      path,
		})
	}

	if err := s.messages.AppendMessage(ctx, msg); err != nil {
		return domain.Message{}, fmt.Errorf("persist inbound message: %w", err)
	}
	s.notify(ctx, inbound.Owner, threadID)

	return msg, nil
}

// Threads lists the owner's threads for the sidebar.
func (s *Service) Threads(ctx context.Context, owner domain.Extension) ([]store.ThreadSummary, error) {
	return s.messages.ListThreads(ctx, owner)
}

// Thread returns one thread's transcript (oldest first).
func (s *Service) Thread(
	ctx context.Context, owner domain.Extension, id domain.ThreadID,
) (domain.Thread, []domain.Message, error) {
	thread, err := s.messages.GetThread(ctx, owner, id)
	if err != nil {
		return domain.Thread{}, nil, err
	}
	msgs, err := s.messages.ListMessages(ctx, owner, id, messagePageSize)
	if err != nil {
		return domain.Thread{}, nil, err
	}
	return thread, msgs, nil
}

// MarkRead zeroes the unread counter.
func (s *Service) MarkRead(ctx context.Context, owner domain.Extension, id domain.ThreadID) error {
	return s.messages.MarkThreadRead(ctx, owner, id)
}

// AttachmentByID returns one attachment scoped to the owner's messages.
func (s *Service) AttachmentByID(ctx context.Context, owner domain.Extension, id domain.AttachmentID) (domain.Attachment, error) {
	return s.messages.AttachmentByID(ctx, owner, id)
}

// OpenAttachment streams one attachment's bytes.
func (s *Service) OpenAttachment(path string) (io.ReadSeekCloser, error) {
	return s.blobs.Open(path)
}

func (s *Service) notify(ctx context.Context, owner domain.Extension, threadID domain.ThreadID) {
	if s.onChange != nil {
		s.onChange(ctx, owner, threadID)
	}
}

func extensionOf(upload Upload) string {
	ext := strings.ToLower(filepath.Ext(upload.Name))
	switch ext {
	case ".jpg", ".jpeg", ".png", ".gif", ".webp", ".pdf", ".txt", ".vcf", ".amr", ".mp3", ".ogg", ".mp4", ".webm":
		return ext
	default:
		if ext == "" && strings.HasPrefix(upload.MimeType, "image/") {
			return ".img"
		}
		return ".bin"
	}
}

func sanitizeFilename(name string) string {
	name = filepath.Base(strings.TrimSpace(name))
	if name == "" || name == "." || name == ".." {
		return "attachment"
	}
	if len(name) > 128 {
		name = name[len(name)-128:]
	}
	return name
}

// ErrThreadNotFound marks a missing thread for handlers mapping to 404.
var ErrThreadNotFound = errors.New("thread not found")
