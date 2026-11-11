// The live-transcription announce paths under node:test (the 2026-11-08
// harvest row): the caption-lag warn fires once per laggy streak and a
// recovered fast segment re-arms it; the unavailable warn fires exactly
// at the third consecutive failure and stays quiet while the streak
// continues. Date.now is replaced by a steppable clock because both
// arms are wall-clock decisions (LAG_WARN_MS 12s, BACKOFF_BASE_MS 2s).
import test from "node:test";
import assert from "node:assert/strict";

import { installBrowserGlobals } from "./helpers.mjs";

const doc = installBrowserGlobals();
globalThis.window.PBX_CONFIG = { asr: true, sipDomain: "sip.test" };

// Steppable wall clock: every Date.now() read in the module goes through
// this, so a fetch stub advances it to age a segment past the lag bar or
// past the failure backoff.
let nowMs = 1_000;
globalThis.Date.now = () => nowMs;

const node = () => ({
  connect: () => node(),
  start() {},
  stop() {},
});
globalThis.AudioContext = class {
  constructor() {
    this.state = "running";
    this.destination = {};
  }
  createMediaStreamDestination() {
    const tracks = [];
    return { stream: { id: "mixed", getTracks: () => tracks }, connect: () => {} };
  }
  createMediaStreamSource() {
    return { connect: () => {} };
  }
  close() {
    this.state = "closed";
    return Promise.resolve();
  }
};
globalThis.window.AudioContext = globalThis.AudioContext;
globalThis.MediaStream = class {
  addTrack() {}
};
globalThis.__announceRecorders = [];
globalThis.MediaRecorder = class {
  constructor(stream, options) {
    this.listeners = {};
    this.state = "inactive";
    globalThis.__announceRecorders.push(this);
  }
  addEventListener(type, fn) {
    (this.listeners[type] ??= []).push(fn);
  }
  start(ms) {
    this.state = "recording";
  }
  emit(data) {
    for (const fn of this.listeners.dataavailable ?? []) fn({ data });
  }
};

const { t } = await import("../island/app/i18n.js");
const { els } = await import("../island/app/ui.js");
const { startLiveTranscription, stopLiveTranscription } =
  await import("../island/app/transcribe.js");

const toasts = () => els.toasts.children.map((c) => c.textContent);
const countToast = (needle, list) => list.filter((s) => s === needle).length;

const entry = () => ({
  session: {
    sessionDescriptionHandler: {
      peerConnection: { getReceivers: () => [], getSenders: () => [] },
    },
  },
});

/** @param {"ok" | "laggy" | "fail"} arm */
const fetchImplFor = (arm) => async () => {
  if (arm === "laggy") {
    nowMs += 13_000;
    return { ok: true, status: 200, json: async () => ({ text: "seg" }) };
  }
  if (arm === "ok") {
    nowMs += 100;
    return { ok: true, status: 200, json: async () => ({ text: "seg" }) };
  }
  nowMs += 100;
  throw { status: 500, message: "provider down" };
};

const start = async (id) => {
  await startLiveTranscription(id, entry(), () => {}, { language: "de" });
  const recorder = globalThis.__announceRecorders.at(-1);
  recorder.start(4_000);
  return recorder;
};

const emit = async (recorder, arm) => {
  let fetch;
  const realFetch = globalThis.fetch;
  globalThis.fetch = async (...args) => fetch(...args);
  fetch = fetchImplFor(arm);
  recorder.emit({ data: { size: 4_000 } });
  await new Promise((resolve) => setTimeout(resolve, 10));
  globalThis.fetch = realFetch;
};

test("the lag warn announces once per streak and recovery re-arms it", async () => {
  const recorder = await start("lag-c1");
  try {
    await emit(recorder, "laggy"); // lag 13s > 12s bar: announce
    await emit(recorder, "laggy"); // still lagging: no second announce
    await emit(recorder, "fail"); // failure 1: below the toast bar
    nowMs += 3_000; // step past the first backoff window
    await emit(recorder, "fail"); // failure 2
    nowMs += 5_000;
    await emit(recorder, "fail"); // failure 3: announce unavailable
    nowMs += 8_001;
    await emit(recorder, "fail"); // failure 4: streak continues, quiet
    nowMs += 61_000; // step past the 60s backoff cap before recovery
    await emit(recorder, "ok"); // fast success: the streak flags reset
    await emit(recorder, "laggy"); // lagging again: announce re-armed

    const seen = toasts();
    assert.equal(countToast(t("transcribeLag"), seen), 2, `lag toasts: ${JSON.stringify(seen)}`);
    assert.equal(
      countToast(t("transcribeUnavailable"), seen),
      1,
      `failure toasts: ${JSON.stringify(seen)}`,
    );
    assert.deepEqual(seen, [t("transcribeLag"), t("transcribeUnavailable"), t("transcribeLag")]);
  } finally {
    stopLiveTranscription("lag-c1");
  }
});

test("the unavailable warn waits for the third consecutive failure", async () => {
  const recorder = await start("fail-c2");
  try {
    // A laggy-but-successful segment first: it resets the failure flags
    // (a fresh streak) and puts a DIFFERENT message last, so the dedup
    // cannot swallow the unavailable announce below. Count-based asserts
    // stay immune to toast expiry.
    await emit(recorder, "laggy");
    const lagBaseline = countToast(t("transcribeLag"), toasts());
    const failBaseline = countToast(t("transcribeUnavailable"), toasts());
    await emit(recorder, "fail");
    nowMs += 3_000;
    await emit(recorder, "fail");
    assert.equal(
      countToast(t("transcribeUnavailable"), toasts()),
      failBaseline,
      "two failures stay quiet",
    );
    nowMs += 5_000;
    await emit(recorder, "fail");
    assert.equal(
      countToast(t("transcribeUnavailable"), toasts()),
      failBaseline + 1,
      "the third failure announces",
    );
    assert.equal(countToast(t("transcribeLag"), toasts()), lagBaseline);
    nowMs += 8_001;
    await emit(recorder, "fail");
    assert.equal(
      countToast(t("transcribeUnavailable"), toasts()),
      failBaseline + 1,
      "the fourth stays quiet",
    );
  } finally {
    stopLiveTranscription("fail-c2");
  }
});
