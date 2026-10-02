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
    // Real sip.js: re-registering a Registerer that never left
    // Registered re-sends the REGISTER and resolves WITHOUT a
    // stateChange (Registered -> Registered is no transition).
    if (this.state !== "Registered") this.fire("Registered");
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

let capturedMediaStreamFactory;

globalThis.SIP = {
  UserAgent: StubUserAgent,
  Registerer: StubRegisterer,
  Web: {
    defaultSessionDescriptionHandlerFactory: (mediaStreamFactory) => {
      capturedMediaStreamFactory = mediaStreamFactory;
      return () => ({});
    },
  },
  RegistererState: {
    Initial: "Initial",
    Registered: "Registered",
    Unregistered: "Unregistered",
    Terminated: "Terminated",
  },
  SessionState: { Terminated: "Terminated" },
};

// Silent audio contexts: onInvite starts the ring tone.
const silentNode = () => ({ connect: () => silentNode(), start() {}, stop() {} });
globalThis.AudioContext = class {
  constructor() {
    this.currentTime = 0;
    this.destination = {};
  }
  createOscillator() {
    return { frequency: {}, connect: silentNode, start() {}, stop() {} };
  }
  createGain() {
    return { gain: {}, connect: silentNode };
  }
};

// Each case loads a FRESH module instance (module-level registerer and
// flags must not leak between scenarios); the shared stub document keeps
// pill assertions meaningful.
const loadConnection = (tag) => import(`../island/app/connection.js?case=${tag}`);

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
const logTexts = () => doc.getElementById("log").children.map((li) => li.textContent);

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

// The 2026-09-22 E2E root cause: a reconnect cycle that succeeds while
// the Registerer never left Registered fires NO stateChange, so the
// pill must be refreshed by the success path itself — otherwise it
// shows the last backoff state while the phone is re-registered.
test("reconnect success refreshes the pill despite no state transition", async (tc) => {
  resetStubs();
  const connection = await loadConnection("stale-pill");
  await connection.connect("1001", "pw");
  tc.mock.timers.enable({ apis: ["setTimeout"] });
  agents.at(-1).delegate.onDisconnect(new Error("ws closed"));
  await tc.mock.timers.tick(2000); // try 1: reconnect succeeds, no state event
  await flushes();
  assert.equal(agents.length, 1, "no rebuild: the cycle succeeded");
  assert.equal(pill(), t("registered"), "the pill must not stay on the stale backoff text");
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

// The re-entrancy guard (SUPERB T17a): two rebuild triggers landing in
// the same tick (a registration-lost event racing the cycle deadline,
// or Unregistered immediately followed by Terminated) must collapse
// into ONE teardown+rebuild — the second caller sees resetting=true,
// logs "already in progress", and returns without touching the fresh
// agent.
test("concurrent rebuild triggers collapse into a single rebuild", async () => {
  resetStubs();
  const connection = await loadConnection("reentrancy");
  await connection.connect("1001", "pw");
  const dead = registerers.at(-1);
  dead.fire("Unregistered");
  dead.fire("Terminated");
  await flushes();
  assert.equal(agents.length, 2, "exactly one fresh agent, not one per trigger");
  assert.equal(registerers.length, 2, "exactly one fresh registerer");
  assert.ok(agents[0].stopped, "the old agent was torn down once");
  assert.ok(
    logTexts().some((line) => line.includes("rebuild already in progress")),
    "the collapsed trigger must say so in #log",
  );
  assert.equal(pill(), t("registered"), "the single rebuild re-registers");
});

// The transient rebuilding pill (SUPERB T17c): the moment the watchdog
// decides to rebuild, the pill says so in the UI language — before the
// fresh REGISTER lands and flips it to "registered" for real.
test("a triggered rebuild first shows the transient rebuilding pill", async () => {
  resetStubs();
  const connection = await loadConnection("rebuilding-pill");
  await connection.connect("1001", "pw");
  const dead = registerers.at(-1);
  dead.fire("Unregistered");
  // One flush: the rebuild has STARTED (pill set synchronously at its
  // top) but the fresh register() may not have completed yet.
  await new Promise((resolve) => setImmediate(resolve));
  const early = pill();
  assert.ok(
    [t("regRebuilding"), t("registered")].includes(early),
    `early pill must be rebuilding or already recovered, got ${early}`,
  );
  await flushes();
  assert.equal(pill(), t("registered"), "the rebuild lands on registered");
});

// The offline banner is the loud honest surface: visible whenever the
// phone cannot call (transport down, reconnecting, rebuilding,
// registration rejected), gone exactly when the REGISTER re-lands.
const banner = () => doc.getElementById("offline-banner");

test("connect hides the banner only once registered", async () => {
  resetStubs();
  const connection = await loadConnection("banner-happy");
  const pending = connection.connect("1001", "pw");
  // Between connect() and the REGISTER's verdict the banner is ON.
  assert.equal(banner().hidden, false, "unregistered phone cannot call");
  await pending;
  assert.equal(banner().hidden, true, "registered hides the banner");
});

test("the banner tracks transport loss and re-registered recovery", async (tc) => {
  resetStubs();
  const connection = await loadConnection("banner-cycle");
  await connection.connect("1001", "pw");
  assert.equal(banner().hidden, true);
  tc.mock.timers.enable({ apis: ["setTimeout"] });
  agents.at(-1).delegate.onDisconnect(new Error("ws closed"));
  assert.equal(banner().hidden, false, "transport loss shows the banner");
  await tc.mock.timers.tick(2000); // try 1: reconnect + register succeed
  await flushes();
  assert.equal(banner().hidden, true, "re-registered hides the banner");
  assert.equal(pill(), t("registered"));
});

test("a rejected registration keeps the banner up", async () => {
  resetStubs();
  const connection = await loadConnection("banner-rejected");
  StubRegisterer.nextRejectWith = new Error("403 Forbidden");
  await assert.rejects(() => connection.connect("1001", "bogus"));
  await flushes();
  assert.equal(banner().hidden, false, "rejected phone cannot call");
});

// networkOnline is the browser "online" event's nudge: it may schedule a
// recovery ONLY when the transport is actually down and nothing is
// pending — and it never claims registered by itself.
test("networkOnline nudges recovery only for a down, idle transport", async (tc) => {
  resetStubs();
  const connection = await loadConnection("banner-netback");
  await connection.connect("1001", "pw");
  const agent = agents.at(-1);
  tc.mock.timers.enable({ apis: ["setTimeout"] });

  agent.connected = false;
  connection.networkOnline(); // down + no pending retry: schedule one
  await tc.mock.timers.tick(2000);
  await flushes();
  assert.equal(pill(), t("registered"), "the nudge recovered the transport");
  assert.equal(banner().hidden, true);

  connection.networkOnline(); // connected: inert
  agent.connected = false;
  agents.at(-1).delegate.onDisconnect(new Error("ws closed")); // retry pending
  const before = agents.length;
  connection.networkOnline(); // a retry is already scheduled: inert
  await tc.mock.timers.tick(2000);
  await flushes();
  assert.equal(agents.length, before, "no double recovery");
  assert.equal(pill(), t("registered"));
});

// --- mic pre-warm wiring --------------------------------------------------
// The factory handed to sip.js is the island's mic module (warm handoff
// at accept time), the ice gathering wait is capped below the sip.js 5 s
// default, an incoming call acquires the mic while it rings, and a
// missed call releases the device.

const micModule = await import("../island/app/mic.js");
const { state } = await import("../island/app/state.js");
const micTrack = () => ({
  stopped: false,
  stop() {
    this.stopped = true;
  },
  addEventListener() {},
});
const micStream = () => {
  const tracks = [micTrack(), micTrack()];
  return { tracks, getTracks: () => tracks };
};

const stubMic = () => {
  const gumCalls = [];
  const streams = [];
  Object.defineProperty(globalThis, "navigator", {
    value: {
      language: "en-US",
      mediaDevices: {
        getUserMedia: (constraints) => {
          gumCalls.push(constraints);
          const stream = micStream();
          streams.push(stream);
          return Promise.resolve(stream);
        },
      },
    },
    configurable: true,
  });
  return { gumCalls, streams };
};

const sdhOptions = () => agents.at(-1).options.sessionDescriptionHandlerFactoryOptions;

const incomingInvitation = () => {
  const listeners = [];
  return {
    remoteIdentity: { uri: { user: "+493012345678" } },
    stateChange: {
      addListener: (fn) => listeners.push(fn),
      fire: (value) => {
        for (const fn of listeners) fn(value);
      },
    },
    reject() {},
  };
};

test("connect wires the mic factory and caps ice gathering", async () => {
  resetStubs();
  capturedMediaStreamFactory = undefined;
  const connection = await loadConnection("mic-wiring");
  await connection.connect("1001", "pw");
  assert.equal(sdhOptions().iceGatheringTimeout, 1000);
  assert.equal(
    capturedMediaStreamFactory,
    micModule.micMediaStreamFactory,
    "sip.js receives the island's mic factory",
  );
});

test("an incoming call starts acquiring the mic while it rings", async () => {
  resetStubs();
  micModule.releaseWarmMic();
  state.incomingSession = null;
  const { gumCalls } = stubMic();
  const connection = await loadConnection("mic-warm");
  await connection.connect("1001", "pw");

  agents.at(-1).delegate.onInvite(incomingInvitation());
  await flushes();
  assert.equal(doc.getElementById("incoming-call").hidden, false);
  assert.equal(gumCalls.length, 1, "the mic is acquired during the ring");
  assert.deepEqual(gumCalls[0], { audio: true, video: false });
  // M17 J9: the ringing call is announced assertively (role=alert) — the
  // hidden→visible banner swap alone is invisible to screen readers.
  const toast = doc.getElementById("toasts").children.at(-1);
  assert.equal(toast.textContent, t("incomingCall")("+493012345678"));
  assert.equal(toast.getAttribute("role"), "alert");
});

test("a missed call releases the warm mic", async () => {
  resetStubs();
  micModule.releaseWarmMic();
  state.incomingSession = null;
  const { gumCalls, streams } = stubMic();
  const connection = await loadConnection("mic-missed");
  await connection.connect("1001", "pw");

  const invitation = incomingInvitation();
  agents.at(-1).delegate.onInvite(invitation);
  await flushes();
  assert.equal(gumCalls.length, 1);

  invitation.stateChange.fire(globalThis.SIP.SessionState.Terminated);
  assert.equal(
    streams[0].tracks.every((track) => track.stopped),
    true,
    "the device is released when the caller gives up",
  );
});

// The second-invite guard: one call at a time. A second INVITE arriving
// while one rings is rejected on the spot and must not touch the first
// call's banner, session or warm mic.
test("a second incoming call is rejected without disturbing the first", async () => {
  resetStubs();
  micModule.releaseWarmMic();
  state.incomingSession = null;
  const { gumCalls } = stubMic();
  const connection = await loadConnection("second-invite");
  await connection.connect("1001", "pw");

  const first = incomingInvitation();
  agents.at(-1).delegate.onInvite(first);
  await flushes();
  let secondRejected = false;
  const second = {
    remoteIdentity: { uri: { user: "+493099998888" } },
    stateChange: { addListener() {} },
    reject() {
      secondRejected = true;
    },
  };
  agents.at(-1).delegate.onInvite(second);
  await flushes();

  assert.equal(secondRejected, true, "the second INVITE is rejected");
  assert.equal(state.incomingSession, first, "the first call keeps the banner");
  assert.equal(doc.getElementById("incoming-call").hidden, false);
  assert.equal(gumCalls.length, 1, "no second microphone acquisition");

  first.stateChange.fire(globalThis.SIP.SessionState.Terminated);
  state.incomingSession = null;
});

// The watchdog rebuild (recover a wedged agent) must NOT release the
// warm mic: the call that is ringing is still ringing — only a missed
// call, a reject or logout returns the device.
test("a watchdog rebuild keeps the warm mic", async () => {
  resetStubs();
  micModule.releaseWarmMic();
  state.incomingSession = null;
  const { gumCalls } = stubMic();
  const connection = await loadConnection("mic-survives-rebuild");
  await connection.connect("1001", "pw");

  const invitation = incomingInvitation();
  agents.at(-1).delegate.onInvite(invitation);
  await flushes();
  assert.equal(gumCalls.length, 1, "the ring warmed the mic");

  // A registration lost after it was established forces a full rebuild.
  registerers.at(-1).fire("Unregistered");
  await flushes();
  assert.equal(agents.length, 2, "the watchdog rebuilt the agent");
  const stillWarm = micModule.takeWarmMic();
  assert.notEqual(stillWarm, null, "the rebuild did not release the warm mic");
  assert.equal(
    stillWarm.tracks.every((track) => !track.stopped),
    true,
    "the warm tracks stay live across the rebuild",
  );
  micModule.releaseWarmMic();
  invitation.stateChange.fire(globalThis.SIP.SessionState.Terminated);
  state.incomingSession = null;
});

// M25 A6 incoming focus mode: the ring dims the shell (root class) and
// moves focus to Accept; a user mid-type keeps their keystrokes; every
// banner end (caller gave up) clears the dim.
test("a ringing call dims the shell and focuses Accept", async () => {
  resetStubs();
  micModule.releaseWarmMic();
  state.incomingSession = null;
  stubMic();
  const connection = await loadConnection("focus-mode");
  doc.activeElement = null;
  await connection.connect("1001", "pw");

  const invitation = incomingInvitation();
  doc.getElementById("accept-btn").focused = false;
  agents.at(-1).delegate.onInvite(invitation);
  await flushes();
  assert.ok(
    doc.documentElement.classList.contains("wp-incoming-focus"),
    "the root focus-mode class is set while ringing",
  );
  assert.equal(
    doc.getElementById("accept-btn").focused,
    true,
    "focus moves to Accept when the user was not typing",
  );

  invitation.stateChange.fire(globalThis.SIP.SessionState.Terminated);
  assert.ok(
    !doc.documentElement.classList.contains("wp-incoming-focus"),
    "the caller giving up clears the dim",
  );
  doc.activeElement = null;
});

test("a user mid-type keeps their focus when a call rings", async () => {
  resetStubs();
  micModule.releaseWarmMic();
  state.incomingSession = null;
  stubMic();
  const connection = await loadConnection("focus-typing");
  doc.activeElement = { tagName: "INPUT" };
  doc.getElementById("accept-btn").focused = false;
  await connection.connect("1001", "pw");

  agents.at(-1).delegate.onInvite(incomingInvitation());
  await flushes();
  assert.ok(doc.documentElement.classList.contains("wp-incoming-focus"));
  assert.equal(
    doc.getElementById("accept-btn").focused,
    false,
    "Accept does not steal focus from a typing user",
  );
  doc.activeElement = null;
});
