// Mutable state shared across modules. Kept in one place so the module
// graph stays acyclic (calls, ice and connection all need the live
// session table; none of them may import each other for it).

// id -> { session, target, held, muted, startedAt, timer, dom }
export const sessions = new Map();

export const state = {
  focusedId: null,
  incomingSession: null,
};
