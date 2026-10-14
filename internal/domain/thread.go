package domain

import "time"

// Thread is the append-owned slice of a message thread: the state one
// message append may touch. Pinned and muted are deliberately absent —
// they are pure user toggles (SetThreadFlag) with no append interaction;
// archived IS here because an inbound append un-archives its thread.
type Thread struct {
	ID             ThreadID
	Owner          Extension
	Remote         Phone
	LastActivityAt time.Time
	Unread         int
	Archived       bool
}

// AppendToThread folds one message into its thread and returns the
// advanced copy; the input is never mutated. It is the ONE home of the
// append invariants:
//
//   - last activity always follows the appended message's creation
//     (parity with the historical upsert: the write is unconditional,
//     even if the stamp is older than the thread's);
//   - unread increments on inbound only;
//   - an inbound message un-archives its thread (the conversation is
//     live again; outbound sends never touch the flag);
//   - a zero thread (no ID yet) is creation: it adopts the message's
//     thread id, owner and remote.
//
// The store's append transaction computes its row through this
// function; the SQL carries no parallel copy of these rules.
func AppendToThread(thread Thread, msg Message) Thread {
	if thread.ID == (ThreadID{}) {
		thread.ID = msg.ThreadID
		thread.Owner = msg.Owner
		thread.Remote = msg.Remote
	}
	thread.LastActivityAt = msg.CreatedAt
	if msg.Direction == DirectionInbound {
		thread.Unread++
		thread.Archived = false
	}
	return thread
}
