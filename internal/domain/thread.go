package domain

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
// Pinned and Muted are untouched — they are pure user toggles with no
// append interaction. The store's append transaction computes its row
// through this function; the SQL carries no parallel copy of these
// rules.
func AppendToThread(thread Thread, msg Message) Thread {
	if thread.ID == (ThreadID{}) {
		thread.ID = msg.ThreadID
		thread.Owner = msg.Owner
		thread.Network = msg.Network
		thread.Remote = msg.Remote
	}
	thread.LastActivityAt = msg.CreatedAt
	if msg.Direction == DirectionInbound {
		thread.Unread++
		thread.Archived = false
	}
	return thread
}
