// audio.js under node:test — the shared-AudioContext contract. Chrome
// leaves a context created outside a user gesture `suspended`, so the
// incoming ring (which has no gesture) would be SILENT. The module must
// therefore (a) never construct the context at import, (b) share ONE
// context across ringback and ring tone, and (c) expose resumeAudio()
// for the gesture handlers in calls.js to bring it to running.
import test from "node:test";
import assert from "node:assert/strict";

import { installBrowserGlobals } from "./helpers.mjs";

installBrowserGlobals();

const contexts = [];
const node = () => ({ connect: () => node(), start() {}, stop() {} });
globalThis.AudioContext = class {
  constructor() {
    this.currentTime = 0;
    this.destination = {};
    this.state = "suspended";
    this.resumeCalls = 0;
    contexts.push(this);
  }
  createOscillator() {
    return { frequency: {}, connect: node, start() {}, stop() {} };
  }
  createGain() {
    return { gain: {}, connect: node };
  }
  resume() {
    this.resumeCalls += 1;
    this.state = "running";
    return Promise.resolve();
  }
};

const loadAudio = (tag) => import(`../island/app/audio.js?case=${tag}`);

test("importing constructs no AudioContext; the first tone creates it lazily", async () => {
  const before = contexts.length;
  const audio = await loadAudio("lazy-create");
  assert.equal(contexts.length, before, "import must not construct an AudioContext");
  audio.ringToneStart();
  assert.equal(contexts.length, before + 1, "the ring tone lazily creates the context");
  audio.ringToneStop();
});

test("ringback and ring tone share ONE context", async () => {
  const before = contexts.length;
  const audio = await loadAudio("shared-ctx");
  audio.ringToneStart();
  audio.ringbackStart();
  assert.equal(contexts.length, before + 1, "one shared context for both tones");
  audio.ringToneStop();
  audio.ringbackStop();
});

test("resumeAudio brings a suspended context to running, once", async () => {
  const audio = await loadAudio("resume");
  await audio.resumeAudio();
  const ctx = contexts.at(-1);
  assert.equal(ctx.state, "running");
  assert.equal(ctx.resumeCalls, 1);
  await audio.resumeAudio();
  assert.equal(ctx.resumeCalls, 1, "no duplicate resume once running");
});
