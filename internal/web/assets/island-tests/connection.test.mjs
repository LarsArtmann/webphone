// connection.js under node:test — the registration wedge detector.
// The 1001 E2E anomaly (registration gone at the server while the
// island kept its dead Registerer) must end in a full agent rebuild;
// a never-registered rejection must stay a pill (bogus-credentials
// login UX parity, no rebuild loop). Drives the module through a
// minimal SIP.js stub.
import test from "node:test";
import assert from "node:assert/strict";

import { installBrowserGlobals } from "./helpers.mjs";

const doc = installBrowserGlobals();
const { t } = await import("../island/app/i18n.js");

const agents = [];
const registerers = [];

class StubEmitter {
  constructor() {
    this.listeners = [];
  }
  addListener(fn) {
    this.listeners.push(fn);
  }
}

class StubRegisterer {
  // Optional class-level script for the NEXT constructed registerer:
  // makes its first register() reject (firing Unregistered first, like
  // the real state machine does on a rejected REGISTER).
  static nextRejectWith = null;

  constructor() {
    this.state = "Initial";
    this.stateChange = new StubEmitter();
    this.registerCalls = 0;
    this.rejectWith = StubRegisterer.nextRejectWith;
    this.silentThrow = null;
    StubRegisterer.nextRejectWith = null;
    registerers.push(this);
  }

  async register() {
    this.registerCalls += 1;
    if (this.silentThrow) throw this.silentThrow;
    if (this.rejectWith) {
      this.fire("Unregistered");
      throw this.rejectWith;
    }
    this.fire("Registered");
  }

  async unregister() {
    this.fire("Unregistered");
    this.fire("Terminated");
  }

  fire(regState) {
    this.state = regState;
    for (const fn of this.stateChange.listeners) fn(regState);
  }
}

class StubUserAgent {
  constructor(options) {
    this.options = options;
    this.delegate = options.delegate;
    this.stopped = false;
    this.connected = true;
    this.hangReconnect = false;
    agents.push(this);
  }

  static makeURI(uri) {
    return { toString: () => uri };
  }

  isConnected() {
    return this.connected;
  }

  async start() {}

  async stop() {
    this.stopped = true;
  }

  async reconnect() {
    if (this.hangReconnect) return new Promise(() => {});
  }
}

globalThis.SIP = {
  UserAgent: StubUserAgent,
  Registerer: StubRegisterer,
  Web: { defaultSessionDescriptionHandlerFactory: () => () => ({}) },
  RegistererState: {
    Initial: "Initial",
    Registered: "Registered",
    Unregistered: "Unregistered",
    Terminated: "Terminated",
  },
  SessionState: { Terminated: "Terminated" },
};

// Each case loads a FRESH module instance (module-level registerer and
// flags must not leak between scenarios); the shared stub document keeps
// pill assertions meaningful.
const loadConnection = (tag) =>
  import(`../island/app/connection.js?case=${tag}`);

const flushes = async (rounds = 12) => {
  while (rounds--) await new Promise((resolve) => setImmediate(resolve));
};

// The stub arrays are file-global; every scenario starts from zero so
// the absolute-count assertions below stay readable.
const resetStubs = () => {
  agents.length = 0;
  registerers.length = 0;
  StubRegisterer.nextRejectWith = null;
};

const pill = () => doc.getElementById("reg-status").textContent;

test("connect builds one agent and reaches the registered pill", async () => {
  resetStubs();
  const connection = await loadConnection("happy");
  await connection.connect("1001", "pw");
  assert.equal(agents.length, 1);
  assert.equal(registerers.length, 1);
  assert.equal(registerers[0].registerCalls, 1);
  assert.equal(pill(), t("registered"));
});

test("Unregistered after a successful registration rebuilds the agent", async () => {
  resetStubs();
  const connection = await loadConnection("lost");
  await connection.connect("1001", "pw");
  const dead = registerers.at(-1);
  dead.fire("Unregistered");
  await flushes();
  assert.ok(agents[0].stopped, "the wedged agent must be stopped");
  assert.equal(agents.length, 2, "a fresh agent must be built");
  assert.equal(registerers.length, 2, "a fresh registerer must be built");
  assert.ok(registerers[1].registerCalls >= 1, "it must register");
  assert.equal(pill(), t("registered"), "the rebuild re-registers");
});

test("Terminated after a successful registration rebuilds the agent", async () => {
  resetStubs();
  const connection = await loadConnection("terminated");
  await connection.connect("1001", "pw");
  registerers.at(-1).fire("Terminated");
  await flushes();
  assert.equal(agents.length, 2);
  assert.ok(agents[0].stopped);
  assert.equal(pill(), t("registered"));
});

test("rejected FIRST registration shows the pill without rebuilding", async () => {
  resetStubs();
  const connection = await loadConnection("bogus-login");
  StubRegisterer.nextRejectWith = new Error("403 Forbidden");
  await assert.rejects(() => connection.connect("1001", "bogus"));
  await flushes();
  assert.equal(agents.length, 1, "no rebuild for a bogus login");
  assert.equal(pill(), t("regRejected"));
});

// The 2026-09-22 E2E failure class: a registration lost while the
// transport is UP gets the rebuild; lost while the transport is DOWN
// must NOT rebuild (the reconnect backoff owns an outage).
test("registration lost during a transport outage stays on the backoff", async () => {
  resetStubs();
  const connection = await loadConnection("outage");
  await connection.connect("1001", "pw");
  agents.at(-1).connected = false;
  registerers.at(-1).fire("Unregistered");
  await flushes();
  assert.equal(agents.length, 1, "no rebuild while the transport is down");
});

// The frozen-cycle class: the attempt chain wedges without settling
// (withTimeout included); the independent cycle-deadline timer must
// still fire and force the rebuild.
test("a wedged reconnect cycle hits the deadline watchdog and rebuilds", async (tc) => {
  resetStubs();
  const connection = await loadConnection("cycle-deadline");
  await connection.connect("1001", "pw");
  tc.mock.timers.enable({ apis: ["setTimeout"] });
  agents.at(-1).hangReconnect = true;
  agents.at(-1).delegate.onDisconnect(new Error("ws closed"));
  await tc.mock.timers.tick(2000); // try 1 starts and hangs
  await flushes();
  assert.equal(agents.length, 1, "the wedged attempt builds nothing");
  await tc.mock.timers.tick(15000); // cycle deadline fires
  await flushes();
  assert.ok(agents[0].stopped, "the wedged agent is torn down");
  assert.equal(agents.length, 2, "the deadline watchdog rebuilds");
  assert.equal(pill(), t("registered"), "the rebuild re-registers");
});

test("reconnect attempt on a Terminated registerer rebuilds the agent", async (tc) => {
  resetStubs();
  const connection = await loadConnection("terminated-retry");
  await connection.connect("1001", "pw");
  const dead = registerers.at(-1);
  tc.mock.timers.enable({ apis: ["setTimeout"] });
  // Transport dies (schedules the reconnect), and the registerer died
  // server-side without a state notification reaching the island.
  agents.at(-1).delegate.onDisconnect(new Error("ws closed"));
  dead.state = "Terminated";
  dead.silentThrow = new Error("Registerer has terminated");
  await tc.mock.timers.tick(2000);
  await flushes();
  assert.ok(agents[0].stopped, "the agent holding the dead registerer stops");
  assert.equal(agents.length, 2, "the retry path rebuilds instead of looping");
  assert.ok(registerers[1].registerCalls >= 1);
  assert.equal(pill(), t("registered"));
});

test("logout does not rebuild", async () => {
  resetStubs();
  const connection = await loadConnection("logout");
  await connection.connect("1001", "pw");
  await connection.disconnect();
  registerers.at(-1).fire("Unregistered");
  registerers.at(-1).fire("Terminated");
  await flushes();
  assert.equal(agents.length, 1, "teardown events during logout are inert");
});
