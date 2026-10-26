// The caption strip with the ASR seam OFF (F050/F056): the whole
// surface is absent — no strip, no copy/clear affordances, no locality
// chip. The bare transcript node stays (the pre-caption shape).
import test from "node:test";
import assert from "node:assert/strict";

import { installBrowserGlobals } from "./helpers.mjs";

const doc = installBrowserGlobals();
// No PBX_CONFIG at all: the strictest seam-off boot.
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
      this.stateChange = { addListener() {} };
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
};

const { placeCall } = await import("../island/app/calls.js");
const { state } = await import("../island/app/state.js");
const { els } = await import("../island/app/ui.js");
els.remoteAudio.play = () => Promise.resolve();

test("an ASR-off boot renders no caption surface on the call card", async () => {
  state.userAgent = {};
  assert.equal(await placeCall("+4930123456"), true);
  const entry = [...(await import("../island/app/state.js")).sessions.values()].at(-1);
  const dom = entry.dom;

  assert.equal(dom.querySelector(".wp-captions"), null, "no caption strip");
  assert.equal(
    dom.querySelector(".copy-transcript-btn"),
    null,
    "no copy affordance",
  );
  assert.equal(
    dom.querySelector(".clear-transcript-btn"),
    null,
    "no clear affordance",
  );
  assert.equal(
    dom.querySelector(".transcribe-btn"),
    null,
    "no transcribe affordance",
  );
  assert.ok(
    dom.querySelector(".call-transcript"),
    "the bare transcript node remains (the pre-caption shape)",
  );
});
