package domain

import "context"

// Event is a service-seam change notification: what happened, to whose
// data, with which aggregate id. Services emit these through their
// ChangeFunc seam; the server's projection turns each kind into SSE
// pushes. The set is closed (unexported marker) — a new notification
// shape is a new event type, never a flag on an old one.
type Event interface{ isDomainEvent() }

// MessageReceived: an inbound message landed in its thread.
type MessageReceived struct {
	Owner  Extension
	Thread ThreadID
}

// MessageDeliveryAdvanced: an outbound message's delivery state moved
// (send accepted or refused, provider verdict landed).
type MessageDeliveryAdvanced struct {
	Owner  Extension
	Thread ThreadID
}

// ThreadFlagChanged: the user toggled a thread organization flag. No
// thread id on purpose — the projection is list-only (a transcript
// push would wipe whatever conversation is open).
type ThreadFlagChanged struct {
	Owner Extension
}

// FaxAdvanced: a fax job moved through its lifecycle (submit, settle,
// verdict, receive).
type FaxAdvanced struct {
	Owner Extension
	Fax   FaxID
}

func (MessageReceived) isDomainEvent()         {}
func (MessageDeliveryAdvanced) isDomainEvent() {}
func (ThreadFlagChanged) isDomainEvent()       {}
func (FaxAdvanced) isDomainEvent()             {}

// ChangeFunc is the ONE seam shape for change notifications: services
// call it with a typed Event, the server projects (which SSE nudges
// follow is projection policy, not service knowledge).
type ChangeFunc func(ctx context.Context, e Event)
