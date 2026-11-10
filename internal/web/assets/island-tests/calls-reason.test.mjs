// The ended card's failure-reason mapping (C1): every trunk verdict gets
// its own honest line — the Q.850 cause under a 403 wrapper names the
// network rejection, the island's own no-verdict deadline records 408,
// and unknown codes still name themselves. Driven through the real
// placeCall path (onReject → Terminated → settled card), never by
// poking private helpers.
import test from "node:test";
import assert from "node:assert/strict";

import { installBrowserGlobals } from "./helpers.mjs";

const doc = installBrowserGlobals();
globalThis.window.PBX_CONFIG = { agent: true, sipDomain: "sip.test" };

globalThis.SIP = {
  SessionState: {
    Establishing: "Establishing",
    Established: "Established",
    Terminating: "Terminating",
    Terminated: "Terminated",
  },
  Inviter: class {
    constructor() {
      this.id = "call-s1";
      this.cancelled = false;
      this.state = globalThis.SIP.SessionState.Establishing;
      this.stateChange = {
        listeners: [],
        addListener(fn) {
          this.listeners.push(fn);
        },
      };
    }
    invite(options) {
      this.inviteOptions = options;
      return Promise.resolve();
    }
    cancel() {
      this.cancelled = true;
    }
  },
};
globalThis.SIP.UserAgent = { makeURI: (raw) => ({ toString: () => raw }) };

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
  createMediaStreamDestination() {
    return { stream: { id: "mixed", getTracks: () => [] } };
  }
  createMediaStreamSource() {
    return { connect() {} };
  }
  close() {
    return Promise.resolve();
  }
};
globalThis.window.AudioContext = globalThis.AudioContext;
globalThis.MediaStream = class {
  constructor() {}
  addTrack() {}
};

let fetchImpl = async () => ({ ok: true, status: 204 });
globalThis.fetch = async (url, options = {}) => fetchImpl(url, options);

const { placeCall } = await import("../island/app/calls.js");
const { sessions, state } = await import("../island/app/state.js");
const { els } = await import("../island/app/ui.js");

els.remoteAudio.play = () => Promise.resolve();

const flush = async (rounds = 12) => {
  while (rounds--) await new Promise((resolve) => setImmediate(resolve));
};

// One outbound call rejected with the given shape: the settled card's
// reason line carries the mapping's verdict.
async function rejectedCall(id, res) {
  state.userAgent = {};
  await placeCall("+4930123456");
  const entry = sessions.get(id);
  assert.ok(entry, "the call is live");
  const inviter = entry.session;
  const onReject = inviter.inviteOptions?.requestDelegate?.onReject;
  assert.ok(onReject, "the invite carries the rejection delegate");
  onReject(res);
  inviter.state = globalThis.SIP.SessionState.Terminated;
  inviter.stateChange.listeners.forEach((fn) => fn("Terminated"));
  await flush();
  return entry;
}

const reasonText = (entry) =>
  entry.dom.querySelector(".call-reason-text")?.textContent ?? "";

test("a 403 that carries Q.850;cause=21 names the network rejection", async () => {
  const entry = await rejectedCall("call-s1", {
    statusCode: 403,
    message: {
      getHeader: (name) =>
        name === "Reason" ? 'Q.850;cause=21;text="Call rejected"' : undefined,
    },
  });
  assert.match(
    reasonText(entry),
    /Rejected by the network/,
    "the carrier verdict outranks the 403 wrapper",
  );
});

test("Q.850 cause 17 reads as busy", async () => {
  const entry = await rejectedCall("call-s1", {
    statusCode: 403,
    message: { getHeader: (name) => (name === "Reason" ? "Q.850;cause=17" : undefined) },
  });
  assert.match(reasonText(entry), /Busy/);
});

test("a plain 403 without a Reason header stays a far-end block", async () => {
  const entry = await rejectedCall("call-s1", {
    statusCode: 403,
    message: { getHeader: () => undefined },
  });
  assert.match(reasonText(entry), /blocked by the far end/);
});

test("488 names the incompatible audio path", async () => {
  const entry = await rejectedCall("call-s1", { statusCode: 488 });
  assert.match(reasonText(entry), /No compatible audio path/);
});

test("401 names the credential refusal", async () => {
  const entry = await rejectedCall("call-s1", { statusCode: 401 });
  assert.match(reasonText(entry), /Sign-in refused/);
});

test("an unknown code still names itself", async () => {
  const entry = await rejectedCall("call-s1", { statusCode: 499 });
  assert.match(reasonText(entry), /Call failed \(code 499\)\./);
});

test("the island's own deadline records 408, cancels, and says no answer", async (t) => {
  t.mock.timers.enable({ apis: ["setTimeout"] });
  state.userAgent = {};
  await placeCall("+4930123456");
  const entry = sessions.get("call-s1");
  assert.ok(entry, "the call is live");
  const inviter = entry.session;
  assert.equal(entry.rejectCause, undefined, "no verdict yet");
  t.mock.timers.tick(32000);
  assert.equal(entry.rejectCause, 408, "the deadline recorded its verdict");
  assert.equal(inviter.cancelled, true, "the hanging INVITE was cancelled");
  inviter.state = globalThis.SIP.SessionState.Terminated;
  inviter.stateChange.listeners.forEach((fn) => fn("Terminated"));
  await flush();
  assert.match(reasonText(entry), /No answer in time\./);
  assert.equal(sessions.has("call-s1"), false, "the session table is clean");
});
