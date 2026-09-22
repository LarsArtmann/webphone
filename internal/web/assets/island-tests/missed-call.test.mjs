// The two genuinely-missed inbound call paths dispatch wp:call-missed
// for the shell's header badge (shell.test.mjs owns the badge half):
// 1. the caller gave up before the user answered (connection.js onInvite),
// 2. an accepted call died before any media (calls.js bindSession).
// A deliberate REJECT is a SEEN call and must never dispatch.
import test from "node:test";
import assert from "node:assert/strict";

import { installBrowserGlobals } from "./helpers.mjs";

const doc = installBrowserGlobals();

class StubEmitter {
  constructor() {
    this.listeners = [];
  }
  addListener(fn) {
    this.listeners.push(fn);
  }
  fire(value) {
    for (const fn of this.listeners) fn(value);
  }
}

const seen = [];
doc.addEventListener("wp:call-missed", (event) => seen.push(event.detail));
const since = (mark) => seen.slice(mark);

// --- path 1: caller gives up while the incoming banner is up -------------

const agents = [];
class StubRegisterer {
  constructor() {
    this.stateChange = new StubEmitter();
  }
  async register() {
    this.stateChange.fire("Registered");
  }
}
class StubUserAgent {
  constructor(options) {
    this.delegate = options.delegate;
    agents.push(this);
  }
  static makeURI(uri) {
    return { toString: () => uri };
  }
  isConnected() {
    return true;
  }
  async start() {}
  async stop() {}
  async reconnect() {}
}

globalThis.SIP = {
  UserAgent: StubUserAgent,
  Registerer: StubRegisterer,
  Web: { defaultSessionDescriptionHandlerFactory: () => () => ({}) },
  RegistererState: { Registered: "Registered" },
  SessionState: { Terminated: "Terminated" },
  Inviter: class {},
};

// Silent audio contexts (ring tone starts with the incoming banner).
const node = () => ({ connect: () => node(), start() {}, stop() {} });
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

test("a caller who gives up before the answer dispatches wp:call-missed", async () => {
  const mark = seen.length;
  const connection = await import("../island/app/connection.js");
  await connection.connect("1001", "pw");
  const invitation = {
    remoteIdentity: { uri: { user: "+493012345678" } },
    stateChange: new StubEmitter(),
    reject() {},
  };
  agents.at(-1).delegate.onInvite(invitation);
  assert.equal(doc.getElementById("incoming-call").hidden, false);

  invitation.stateChange.fire(globalThis.SIP.SessionState.Terminated);
  const events = since(mark);
  assert.equal(events.length, 1, "exactly one missed-call event");
  assert.equal(events[0].target, "+493012345678");
  assert.equal(doc.getElementById("incoming-call").hidden, true);
});

test("a deliberate REJECT never dispatches wp:call-missed", async () => {
  const mark = seen.length;
  const { rejectIncoming } = await import("../island/app/calls.js");
  const connection = await import("../island/app/connection.js");
  const invitation = {
    remoteIdentity: { uri: { user: "+493098765432" } },
    stateChange: new StubEmitter(),
    reject() {},
  };
  agents.at(-1).delegate.onInvite(invitation);

  rejectIncoming();
  invitation.stateChange.fire(globalThis.SIP.SessionState.Terminated);
  assert.equal(since(mark).length, 0, "a seen (rejected) call is not missed");
  connection.disconnect();
});

// --- path 2: accepted call dies before any media --------------------------

test("an accepted call that dies before media dispatches wp:call-missed", async () => {
  const mark = seen.length;
  const { sessions } = await import("../island/app/state.js");
  const { bindSession } = await import("../island/app/calls.js");
  const { state } = await import("../island/app/state.js");

  const card = {
    dataset: {},
    classList: { toggle() {} },
    remove() {},
    querySelector: () => ({ textContent: "" }),
  };
  const session = {
    id: "dead-inbound",
    state: globalThis.SIP.SessionState.Establishing,
    stateChange: new StubEmitter(),
  };
  bindSession(session, "+441632960111");

  session.state = globalThis.SIP.SessionState.Terminated;
  session.stateChange.fire(globalThis.SIP.SessionState.Terminated);

  const events = since(mark);
  assert.equal(events.length, 1, "the dead inbound call counts as missed");
  assert.equal(events[0].target, "+441632960111");
  assert.equal(sessions.has("dead-inbound"), false, "session torn down");
  state.focusedId = null;
});

test("an OUTBOUND call that never connects is not missed", async () => {
  const mark = seen.length;
  const { sessions } = await import("../island/app/state.js");
  const { bindSession } = await import("../island/app/calls.js");
  const { state } = await import("../island/app/state.js");

  const card = {
    dataset: {},
    classList: { toggle() {} },
    remove() {},
    querySelector: () => ({ textContent: "" }),
  };
  const session = new globalThis.SIP.Inviter();
  session.id = "dead-outbound";
  session.state = globalThis.SIP.SessionState.Establishing;
  session.stateChange = new StubEmitter();
  bindSession(session, "+441632960222");

  session.state = globalThis.SIP.SessionState.Terminated;
  session.stateChange.fire(globalThis.SIP.SessionState.Terminated);

  assert.equal(since(mark).length, 0, "outbound failures are not missed calls");
  sessions.delete("dead-outbound");
  state.focusedId = null;
});
