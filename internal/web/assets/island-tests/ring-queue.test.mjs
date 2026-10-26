// The ringing queue (2026-10-26 multi-call train): N-deep invites
// answered oldest-first, Esc pinned to the newest ring, per-slot
// removal on reject/remote-cancel. Drives connection.js onInvite to
// fill the queue and calls.js/shortcuts.js to work it.
import test from "node:test";
import assert from "node:assert/strict";

import { installBrowserGlobals } from "./helpers.mjs";

const doc = installBrowserGlobals();
globalThis.Element = class Element {
  // The shortcuts gate reads target.closest("input, …"); a BODY target
  // matches no typing selector, so null is the honest stub answer.
  closest() {
    return null;
  }
};
globalThis.SIP = {
  SessionState: {
    Establishing: "Establishing",
    Established: "Established",
    Terminating: "Terminating",
    Terminated: "Terminated",
  },
  Inviter: class {},
  RegistererState: { Initial: "Initial", Registered: "Registered" },
};

const node = () => ({
  connect() {
    return node();
  },
  start() {},
  stop() {},
});
globalThis.AudioContext = class {
  constructor() {
    this.currentTime = 0;
    this.destination = {};
  }
  createOscillator() {
    return { frequency: {}, connect: node, start() {}, stop() {} };
  }
  createGain() {
    return { gain: {}, connect: node };
  }
};

const { sessions, state } = await import("../island/app/state.js");
const { answerIncoming, rejectIncoming, teardownAll } = await import(
  "../island/app/calls.js"
);
const { initShortcuts } = await import("../island/app/shortcuts.js");
const connection = await import("../island/app/connection.js");

const flushes = async (rounds = 12) => {
  while (rounds--) await new Promise((resolve) => setImmediate(resolve));
};

class StubUserAgent {
  static makeURI(uri) {
    return { toString: () => uri };
  }
  constructor(options) {
    this.delegate = options.delegate;
  }
  async start() {}
  async stop() {}
  isConnected() {
    return true;
  }
  async reconnect() {}
}
class StubRegisterer {
  constructor() {
    this.stateChange = { addListener() {} };
  }
  async register() {}
  async unregister() {}
}
globalThis.SIP.UserAgent = StubUserAgent;
globalThis.SIP.Registerer = StubRegisterer;
globalThis.SIP.Web = {
  defaultSessionDescriptionHandlerFactory: () => () => ({}),
};

const reset = () => {
  sessions.clear();
  state.incomingSessions.length = 0;
  state.focusedId = null;
  state.userAgent = null;
};

// A ringing invite stub: one stateChange emitter, one accept. The stub
// models the stack's state machine as far as calls.js reads it: accept
// reaches Established, and a hangup's bye() settles into Terminated —
// that arrival is what tears the session card down (bindSession's
// listener), exactly as the real sip.js stack notifies.
const invitation = (caller) => {
  const listeners = [];
  const inv = {
    caller,
    remoteIdentity: { uri: { user: caller } },
    state: globalThis.SIP.SessionState.Establishing,
    stateChange: {
      addListener: (fn) => listeners.push(fn),
      fire: (value) => {
        for (const fn of listeners) fn(value);
      },
    },
    acceptCalls: 0,
    rejectCalls: 0,
    byeCalls: 0,
    accept() {
      this.acceptCalls += 1;
      this.state = globalThis.SIP.SessionState.Established;
      return Promise.resolve();
    },
    reject() {
      this.rejectCalls += 1;
      this.state = globalThis.SIP.SessionState.Terminated;
    },
    bye() {
      this.byeCalls += 1;
      this.state = globalThis.SIP.SessionState.Terminated;
      this.stateChange.fire(globalThis.SIP.SessionState.Terminated);
      return Promise.resolve();
    },
    id: `invite-${caller}`,
  };
  return inv;
};

const ring = async (agent, invite) => {
  agent.delegate.onInvite(invite);
  await flushes();
};

test("two concurrent invites queue; the banner stack shows both", async () => {
  reset();
  await connection.connect("1001", "pw");
  const agent = state.userAgent;
  const first = invitation("+493011111111");
  const second = invitation("+493022222222");
  await ring(agent, first);
  await ring(agent, second);

  assert.equal(state.incomingSessions.length, 2, "both invites hold a slot");
  assert.equal(
    doc.getElementById("incoming-from").textContent,
    "+493011111111",
    "slot 1 displays the oldest caller",
  );
  const rows = doc.getElementById("incoming-queue").children;
  assert.equal(rows.length, 1, "slots 2..N render one row each");
  assert.match(
    rows[0].children[0].textContent,
    /\+493022222222/,
    "the extra row names its own caller",
  );

  // Remote cancel mid-queue removes exactly its own slot.
  first.stateChange.fire(globalThis.SIP.SessionState.Terminated);
  assert.equal(state.incomingSessions.length, 1);
  assert.equal(state.incomingSessions[0].invitation, second);
  assert.equal(
    doc.getElementById("incoming-from").textContent,
    "+493022222222",
    "the surviving ring claims slot 1",
  );

  second.stateChange.fire(globalThis.SIP.SessionState.Terminated);
  assert.equal(state.incomingSessions.length, 0);
  assert.equal(
    doc.getElementById("incoming-call").hidden,
    true,
    "an empty queue hides the banner",
  );
});

test("answer takes the OLDEST ring; Esc rejects the NEWEST", async () => {
  reset();
  await connection.connect("1001", "pw");
  const agent = state.userAgent;
  const oldest = invitation("+493011111111");
  const newest = invitation("+493022222222");
  await ring(agent, oldest);
  await ring(agent, newest);

  assert.equal(answerIncoming(), true, "the A path answers");
  assert.equal(oldest.acceptCalls, 1, "the oldest invite was accepted");
  assert.equal(newest.acceptCalls, 0);
  assert.equal(state.incomingSessions.length, 1, "the newest ring survives");
  assert.equal(
    doc.getElementById("incoming-from").textContent,
    "+493022222222",
    "the stack re-renders onto the survivor",
  );
  assert.equal(sessions.size, 1, "the answer minted one session card");

  rejectIncoming();
  assert.equal(newest.rejectCalls, 1, "Esc (no target) rejects the newest");
  assert.equal(state.incomingSessions.length, 0);
  assert.equal(
    doc.getElementById("incoming-call").hidden,
    true,
    "the last slot out stands the ring down",
  );
  teardownAll();
});

test("the per-slot buttons answer/reject exactly their own caller", async () => {
  reset();
  await connection.connect("1001", "pw");
  const agent = state.userAgent;
  const first = invitation("+493011111111");
  const second = invitation("+493022222222");
  await ring(agent, first);
  await ring(agent, second);

  // The extra row's reject button routes the SECOND invite (its own).
  const row = doc.getElementById("incoming-queue").children[0];
  const rejectBtn = row.children[1].children[1];
  assert.equal(rejectBtn.className, "danger");
  rejectBtn.click();
  assert.equal(second.rejectCalls, 1, "the row rejects its own caller");
  assert.equal(first.rejectCalls, 0);
  assert.equal(state.incomingSessions.length, 1);

  // The pinned Accept (slot 1) answers the remaining (now oldest) ring.
  doc.getElementById("accept-btn").click();
  assert.equal(first.acceptCalls, 1, "the pinned slot answers the survivor");
  assert.equal(second.acceptCalls, 0, "the rejected caller stays rejected");
  assert.equal(state.incomingSessions.length, 0);
  teardownAll();
});

test("Esc rejects the newest ring and never hangs up the live call", async () => {
  reset();
  initShortcuts();
  await connection.connect("1001", "pw");
  const agent = state.userAgent;
  const live = invitation("+493011111111");
  await ring(agent, live);
  assert.equal(answerIncoming(), true);
  assert.equal(sessions.size, 1, "the live call is up");

  const fresh = invitation("+493022222222");
  await ring(agent, fresh);

  const target = new globalThis.Element();
  target.tagName = "BODY";
  doc.dispatch("keydown", { key: "Escape", target });
  assert.equal(fresh.rejectCalls, 1, "Esc cleared the newest ring");
  assert.equal(sessions.size, 1, "the live call stayed untouched");
  assert.equal(state.incomingSessions.length, 0);

  // With an empty queue Esc hangs the focused call up again.
  doc.dispatch("keydown", { key: "Escape", target });
  assert.equal(sessions.size, 0, "the second Esc ended the live call");
});

test("the answer path is silent with an empty queue", async () => {
  reset();
  assert.equal(answerIncoming(), false, "nothing to answer");
  rejectIncoming();
  assert.equal(state.incomingSessions.length, 0);
});

test("the missed-call contract fires per remotely-cancelled caller", async () => {
  reset();
  await connection.connect("1001", "pw");
  const agent = state.userAgent;
  const missed = [];
  doc.addEventListener("wp:call-missed", (event) => missed.push(event.detail.target));
  const first = invitation("+493011111111");
  const second = invitation("+493022222222");
  await ring(agent, first);
  await ring(agent, second);
  second.stateChange.fire(globalThis.SIP.SessionState.Terminated);
  assert.deepEqual(missed, ["+493022222222"], "exactly the cancelled caller");
  first.stateChange.fire(globalThis.SIP.SessionState.Terminated);
  assert.deepEqual(missed, ["+493022222222", "+493011111111"]);
});
