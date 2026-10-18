// Mutable state shared across modules. Kept in one place so the module
// graph stays acyclic (calls, ice and connection all need the live
// session table; none of them may import each other for it).

// id -> { session, target, held, holdPending, holdQueued, muted, startedAt, timer, dom }
export const sessions = new Map();

export const state = {
  focusedId: null,
  incomingSession: null,
  // The live SIP.UserAgent, written by connection.js only. Lives here so
  // calls and ice never import the connection module for it.
  userAgent: null,
};

// CALL_TRANSITIONS is the island's declarative call-lifecycle table
// (D-train): every UI-visible call state and its legal successors live
// here as data, never as scattered branches in calls.js. View states:
//   ringing   — invite sent, not answered (sip.js Establishing)
//   active    — media flows
//   holding / resuming — a hold toggle is IN FLIGHT (the honest pending
//               state: the UI shows the work, never the hope)
//   held      — the settled hold
//   ending / ended — tearing down and gone
// The honesty contract rides the table: a FAILED toggle returns to the
// SETTLED state (holding → active|held per where it came from), never
// the hoped-for one; applyTransition refuses illegal moves outright so
// a wiring bug fails loudly (logged + state kept) instead of rendering
// a lie.
export const CALL_TRANSITIONS = {
  ringing: ["active", "ending", "ended"],
  active: ["holding", "held", "ending", "ended"],
  holding: ["active", "held", "ending", "ended"],
  resuming: ["active", "held", "ending", "ended"],
  held: ["resuming", "active", "ending", "ended"],
  ending: ["ended"],
  ended: [],
};

// applyTransition proposes one move. Illegal proposals keep the current
// state (changed=false) — the caller logs it; legality questions are
// answered by the table, never re-derived at call sites.
export function applyTransition(current, next) {
  const legal = (CALL_TRANSITIONS[current] ?? []).includes(next);
  return { state: legal ? next : current, changed: legal };
}

// viewStateOf derives the CURRENT view state of a call entry from the
// same three fields the renderer reads (session state, hold pending,
// settled hold) — the table's vocabulary maps 1:1 onto them, so the
// data flow stays single-brained: mutation sites propose through
// applyTransition, rendering derives through viewStateOf.
export function viewStateOf(entry) {
  if (!entry || !entry.session) return "ended";
  const s = entry.session.state;
  if (s === "Established") {
    if (entry.holdPending === "holding") return "holding";
    if (entry.holdPending === "resuming") return "resuming";
    return entry.held ? "held" : "active";
  }
  if (s === "Establishing" || s === "Progress") return "ringing";
  if (s === "Terminating") return "ending";
  return "ended";
}
