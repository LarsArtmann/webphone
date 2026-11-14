// The ended card with the agent seam ON but ASR OFF: no transcript can
// exist for the call, so the summarize affordance must not render —
// today it 404s into an announce after the click (the defect the
// Summarize micro-cluster pins). The agent gate alone is not enough:
// the transcript seam is the producer of the text the agent summarizes.
import test from "node:test";
import assert from "node:assert/strict";

import { installBrowserGlobals } from "./helpers.mjs";

installBrowserGlobals();
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
      this.id = "call-noasr";
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

test("ended card with agent on but ASR off: Close only, no summarize row", async () => {
  state.userAgent = {};
  await placeCall("+4930123456");
  const entry = sessions.get("call-noasr");
  entry.session.state = globalThis.SIP.SessionState.Established;
  entry.session.stateChange.listeners.forEach((fn) => fn("Established"));
  entry.session.state = globalThis.SIP.SessionState.Terminated;
  entry.session.stateChange.listeners.forEach((fn) => fn("Terminated"));
  await flush();

  assert.equal(sessions.has("call-noasr"), false);
  assert.ok(entry.dom.parent, "the ended card stays visible");
  assert.ok(entry.dom.classList.contains("wp-ended-card"));
  assert.equal(
    entry.dom.querySelector(".summarize-btn"),
    null,
    "no summarize affordance without the transcript seam — the button cannot work",
  );
  assert.ok(entry.dom.querySelector(".close-ended-btn"), "Close remains");
});
