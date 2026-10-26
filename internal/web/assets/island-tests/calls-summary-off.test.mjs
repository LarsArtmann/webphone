// The ended card with the agent seam OFF (M12): the post-call row is
// Close only — no summarize affordance exists, and nothing about the
// card leaks the seam. No PBX_CONFIG at all: the strictest seam-off boot.
import test from "node:test";
import assert from "node:assert/strict";

import { installBrowserGlobals } from "./helpers.mjs";

const doc = installBrowserGlobals();

globalThis.SIP = {
  SessionState: {
    Establishing: "Establishing",
    Established: "Established",
    Terminating: "Terminating",
    Terminated: "Terminated",
  },
  Inviter: class {
    constructor() {
      this.id = "call-noagent";
      this.state = globalThis.SIP.SessionState.Establishing;
      this.stateChange = {
        listeners: [],
        addListener(fn) {
          this.listeners.push(fn);
        },
      };
    }
    invite() {
      return Promise.resolve();
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

const { placeCall } = await import("../island/app/calls.js");
const { sessions, state } = await import("../island/app/state.js");
const { els } = await import("../island/app/ui.js");

els.remoteAudio.play = () => Promise.resolve();

const flush = async (rounds = 12) => {
  while (rounds--) await new Promise((resolve) => setImmediate(resolve));
};

test("ended card with the seam off: Close only, no summarize row", async () => {
  state.userAgent = {};
  await placeCall("+4930123456");
  const entry = sessions.get("call-noagent");
  entry.session.state = globalThis.SIP.SessionState.Established;
  entry.session.stateChange.listeners.forEach((fn) => fn("Established"));
  entry.session.state = globalThis.SIP.SessionState.Terminated;
  entry.session.stateChange.listeners.forEach((fn) => fn("Terminated"));
  await flush();

  assert.equal(sessions.has("call-noagent"), false);
  assert.ok(entry.dom.parent, "the ended card stays visible");
  assert.ok(entry.dom.classList.contains("wp-ended-card"));
  assert.equal(
    entry.dom.querySelector(".summarize-btn"),
    null,
    "the summarize affordance does not exist without the seam",
  );
  assert.ok(entry.dom.querySelector(".close-ended-btn"), "Close remains");
});
