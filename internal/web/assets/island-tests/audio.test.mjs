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
  setSinkId(deviceId) {
    this.sinkCalls = this.sinkCalls || [];
    this.sinkCalls.push(deviceId);
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

test("setToneSink routes the tone context to the picker's device", async () => {
  const before = contexts.length;
  const audio = await loadAudio("sink-applied");
  audio.setToneSink("spk-1");
  assert.equal(
    contexts.length,
    before,
    "remembering a device must not create the context (apps pay for silence)",
  );
  audio.ringToneStart();
  const sinkCtx = contexts.at(-1);
  assert.deepEqual(
    sinkCtx.sinkCalls,
    ["spk-1"],
    "the remembered device applies the moment the context exists",
  );
  audio.setToneSink("");
  assert.deepEqual(
    sinkCtx.sinkCalls,
    ["spk-1", ""],
    "the browser default is a real pick, not a no-op",
  );
  audio.ringToneStop();
});

test("a browser without AudioContext.setSinkId keeps the room alarm", async () => {
  const saved = AudioContext.prototype.setSinkId;
  delete AudioContext.prototype.setSinkId;
  try {
    const audio = await loadAudio("sink-unsupported");
    audio.setToneSink("spk-1");
    assert.doesNotThrow(
      () => audio.ringToneStart(),
      "an unsupported sink must never break the ring",
    );
    audio.ringToneStop();
  } finally {
    AudioContext.prototype.setSinkId = saved;
  }
});

// --- the gesture unlock (2026-11-19 ring-silence + autoplay storm) ---------
// A cookie session restore boots the island without any click, so the
// first ring creates the context OUTSIDE a gesture: Chrome suspends it
// forever and prints one "not allowed to start" warning per blocked
// attempt — N warnings under a fork-storm. The unlock contract: arming
// at creation, no failed resume per ring, one interaction unlocks.

// Each fresh module instance arms its own listener pair on the SHARED
// window stub; the gesture tests reset the registry so every assertion
// counts exactly its own module's pair.
const resetWindowListeners = () => {
  globalThis.window.listeners = {};
};
const pointerBaseline = () =>
  (globalThis.window.listeners.pointerdown ?? []).length;

test("a never-gestured ring does not resume; the first interaction unlocks", async () => {
  resetWindowListeners();
  const baseline = pointerBaseline();
  const audio = await loadAudio("gesture-unlock");
  audio.ringToneStart();
  const ctx = contexts.at(-1);
  assert.equal(ctx.state, "suspended");
  assert.equal(
    ctx.resumeCalls,
    0,
    "a ring carries no gesture: no doomed resume attempt per invite",
  );
  globalThis.window.dispatchWindow("pointerdown");
  await new Promise((resolve) => setImmediate(resolve));
  assert.equal(ctx.state, "running", "the first interaction unlocks the ring");
  assert.equal(ctx.resumeCalls, 1);
  assert.equal(
    pointerBaseline(),
    baseline,
    "the unlock disarms itself after succeeding",
  );
  globalThis.window.dispatchWindow("keydown");
  assert.equal(ctx.resumeCalls, 1, "no duplicate resume from the keydown");
  audio.ringToneStop();
});

test("a refused resume stays armed; the next gesture retries", async () => {
  resetWindowListeners();
  const baseline = pointerBaseline();
  const audio = await loadAudio("gesture-retry");
  audio.ringToneStart();
  const ctx = contexts.at(-1);
  ctx.resume = () => Promise.reject(new Error("no user activation"));
  globalThis.window.dispatchWindow("pointerdown");
  await new Promise((resolve) => setImmediate(resolve));
  assert.equal(ctx.state, "suspended", "the refused resume changes nothing");
  assert.equal(
    pointerBaseline(),
    baseline + 1,
    "the unlock stays armed for the next interaction",
  );
  ctx.resume = function () {
    this.resumeCalls = (this.resumeCalls ?? 0) + 1;
    this.state = "running";
    return Promise.resolve();
  };
  globalThis.window.dispatchWindow("keydown");
  await new Promise((resolve) => setImmediate(resolve));
  assert.equal(ctx.state, "running", "the second gesture unlocks");
  assert.equal(pointerBaseline(), baseline, "disarmed after success");
  audio.ringToneStop();
});

test("a running context never re-arms the unlock", async () => {
  resetWindowListeners();
  const audio = await loadAudio("gesture-running");
  await audio.resumeAudio(); // explicit gesture path creates + unlocks
  const ctx = contexts.at(-1);
  globalThis.window.dispatchWindow("pointerdown"); // disarm whatever creation armed
  await new Promise((resolve) => setImmediate(resolve));
  const baseline = pointerBaseline();
  audio.ringToneStart();
  assert.equal(
    pointerBaseline(),
    baseline,
    "a running context must not arm listeners",
  );
  assert.equal(ctx.resumeCalls, 1, "no resume attempt on an already-running context");
  audio.ringToneStop();
});
