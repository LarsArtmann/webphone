// transcribe.js under node:test — the /api/transcribe client seam.
// Covers the one wire contract every audio surface rides (blob POST, URL
// fetch, live-call segment capture) with stubbed fetch/MediaRecorder.
import test from "node:test";
import assert from "node:assert/strict";

import { installBrowserGlobals } from "./helpers.mjs";

installBrowserGlobals();
// Force the ASR gate ON: config.js reads window.PBX_CONFIG at import time.
globalThis.window.PBX_CONFIG = { asr: true };

// --- fetch stub -------------------------------------------------------------

let fetchCalls = [];
let fetchImpl = async () => ({ ok: true, status: 200, json: async () => ({ text: "" }) });
globalThis.fetch = async (url, options = {}) => {
  fetchCalls.push({ url, options });
  return fetchImpl(url, options);
};

// --- MediaRecorder / AudioContext stubs ------------------------------------

globalThis.__recorders = [];
globalThis.__contexts = [];
globalThis.MediaStream = class {
  constructor() {}
  addTrack() {}
};
globalThis.window.AudioContext = class {
  constructor() {
    this.closed = false;
    globalThis.__contexts.push(this);
  }
  createMediaStreamDestination() {
    const tracks = [{ stopped: false, stop() { this.stopped = true; } }];
    this.mixedTracks = tracks;
    return { stream: { id: "mixed", getTracks: () => tracks } };
  }
  createMediaStreamSource() {
    return { connect() {} };
  }
  close() {
    this.closed = true;
    return Promise.resolve();
  }
};
globalThis.MediaRecorder = class {
  static isTypeSupported() {
    return false;
  }
  constructor(stream, options) {
    this.stream = stream;
    this.options = options;
    this.listeners = {};
    this.state = "inactive";
    globalThis.__recorders.push(this);
  }
  addEventListener(type, fn) {
    (this.listeners[type] ??= []).push(fn);
  }
  start(ms) {
    this.startedWith = ms;
    this.state = "recording";
  }
  pause() {
    this.paused = true;
    this.state = "paused";
  }
  resume() {
    this.resumed = true;
    this.state = "recording";
  }
  stop() {
    this.stopped = true;
    for (const fn of this.listeners.stop ?? []) fn({});
  }
  emit(data) {
    for (const fn of this.listeners.dataavailable ?? []) fn({ data });
  }
};

const transcribe = await import("../island/app/transcribe.js");

const flush = async (rounds = 8) => {
  while (rounds--) await new Promise((resolve) => setImmediate(resolve));
};

test("transcribeBlob posts the blob to the seam and returns trimmed text", async () => {
  fetchCalls = [];
  fetchImpl = async () => ({ ok: true, status: 200, json: async () => ({ text: "  hi there  " }) });
  const text = await transcribe.transcribeBlob(
    { type: "audio/webm", size: 3 },
    { filename: "clip.webm", language: "de" },
  );
  assert.equal(text, "hi there");
  assert.equal(fetchCalls.length, 1);
  const call = fetchCalls[0];
  assert.match(call.url, /^\/api\/transcribe\?/);
  assert.match(call.url, /filename=clip\.webm/);
  assert.match(call.url, /lang=de/);
  assert.equal(call.options.method, "POST");
  // auth.js passes a Headers instance through, so read it with .get().
  assert.equal(call.options.headers.get("Content-Type"), "audio/webm");
});

test("transcribeBlob throws on a failed response", async () => {
  fetchImpl = async () => ({ ok: false, status: 502 });
  await assert.rejects(
    () => transcribe.transcribeBlob({ type: "audio/webm", size: 3 }),
    /HTTP 502/,
  );
});

test("transcribeUrl fetches then posts the audio", async () => {
  fetchCalls = [];
  fetchImpl = async (url) => {
    if (url.startsWith("/phone-api/")) {
      return { ok: true, status: 200, blob: async () => ({ type: "audio/wav", size: 9 }) };
    }
    return { ok: true, status: 200, json: async () => ({ text: "voicemail words" }) };
  };
  const text = await transcribe.transcribeUrl("/phone-api/voicemail/1001/messages/x/audio");
  assert.equal(text, "voicemail words");
  assert.equal(fetchCalls.length, 2);
  assert.equal(fetchCalls[0].url, "/phone-api/voicemail/1001/messages/x/audio");
  assert.match(fetchCalls[1].url, /^\/api\/transcribe\?/);
});

test("startLiveTranscription captures segments and reports text", async () => {
  globalThis.__recorders = [];
  fetchCalls = [];
  fetchImpl = async () => ({ ok: true, status: 200, json: async () => ({ text: "hello live" }) });

  const pc = {
    getReceivers: () => [{ track: { kind: "audio" } }],
    getSenders: () => [{ track: { kind: "audio" } }],
  };
  const entry = { session: { sessionDescriptionHandler: { peerConnection: pc } } };
  const texts = [];

  const ok = transcribe.startLiveTranscription("call-1", entry, (t) => texts.push(t));
  assert.equal(ok, true);
  assert.equal(transcribe.isTranscribing("call-1"), true);
  assert.equal(globalThis.__recorders.length, 1);
  assert.equal(globalThis.__recorders[0].startedWith, 4000);

  globalThis.__recorders[0].emit({ size: 12, type: "audio/webm" });
  await flush();
  assert.deepEqual(texts, ["hello live"]);

  transcribe.stopLiveTranscription("call-1");
  assert.equal(transcribe.isTranscribing("call-1"), false);
  assert.equal(globalThis.__recorders[0].stopped, true);
});

test("startLiveTranscription is idempotent per call and refuses without media", () => {
  globalThis.__recorders = [];
  const pc = { getReceivers: () => [], getSenders: () => [] };
  const entry = { session: { sessionDescriptionHandler: { peerConnection: pc } } };
  assert.equal(transcribe.startLiveTranscription("call-2", entry, () => {}), true);
  assert.equal(transcribe.startLiveTranscription("call-2", entry, () => {}), false);
  transcribe.stopLiveTranscription("call-2");

  // No peer connection at all: honest refusal, no recorder created.
  globalThis.__recorders = [];
  assert.equal(transcribe.startLiveTranscription("call-3", {}, () => {}), false);
  assert.equal(globalThis.__recorders.length, 0);
});

test("stopLiveTranscription on an unknown call is a no-op", () => {
  assert.doesNotThrow(() => transcribe.stopLiveTranscription("never-started"));
});

test("stopLiveTranscription closes the AudioContext and stops the mixed tracks", () => {
  globalThis.__recorders = [];
  globalThis.__contexts = [];
  const pc = { getReceivers: () => [{ track: { kind: "audio" } }], getSenders: () => [] };
  const entry = { session: { sessionDescriptionHandler: { peerConnection: pc } } };
  transcribe.startLiveTranscription("leak-check", entry, () => {});
  assert.equal(globalThis.__contexts.length, 1);
  const [ctx] = globalThis.__contexts;
  transcribe.stopLiveTranscription("leak-check");
  assert.equal(ctx.closed, true, "the mixing AudioContext must be closed (browsers cap instances)");
  assert.equal(ctx.mixedTracks[0].stopped, true, "the mixed stream tracks must be stopped");
});

test("pauseLiveTranscription mirrors the recorder's hold state", () => {
  globalThis.__recorders = [];
  globalThis.__contexts = [];
  const pc = { getReceivers: () => [{ track: { kind: "audio" } }], getSenders: () => [] };
  const entry = { session: { sessionDescriptionHandler: { peerConnection: pc } } };
  transcribe.startLiveTranscription("hold-check", entry, () => {});
  const [recorder] = globalThis.__recorders;
  transcribe.pauseLiveTranscription("hold-check", true);
  assert.equal(recorder.paused, true);
  transcribe.pauseLiveTranscription("hold-check", false);
  assert.equal(recorder.resumed, true);
  transcribe.stopLiveTranscription("hold-check");
  // Unknown call: safe on every hold flip.
  assert.doesNotThrow(() => transcribe.pauseLiveTranscription("nope", true));
});

test("startLiveTranscription forwards the language hint on every segment", async () => {
  globalThis.__recorders = [];
  globalThis.__contexts = [];
  fetchCalls = [];
  fetchImpl = async () => ({ ok: true, status: 200, json: async () => ({ text: "Guten Tag" }) });
  const pc = { getReceivers: () => [{ track: { kind: "audio" } }], getSenders: () => [] };
  const entry = { session: { sessionDescriptionHandler: { peerConnection: pc } } };
  transcribe.startLiveTranscription("lang-check", entry, () => {}, { language: "de" });
  globalThis.__recorders[0].emit({ size: 5, type: "audio/webm" });
  await flush();
  assert.match(fetchCalls[0].url, /lang=de/);
  transcribe.stopLiveTranscription("lang-check");
});

test("saveTranscriptSegment POSTs the segment to the store endpoint", async () => {
  fetchCalls = [];
  fetchImpl = async () => ({ ok: true, status: 204 });
  await transcribe.saveTranscriptSegment({
    callId: "abc",
    direction: "out",
    remote: "+4930123456",
    startedAt: 1760000000000,
    text: "hello",
  });
  assert.equal(fetchCalls.length, 1);
  assert.equal(fetchCalls[0].url, "/api/transcripts");
  assert.equal(fetchCalls[0].options.method, "POST");
  assert.deepEqual(JSON.parse(fetchCalls[0].options.body), {
    callId: "abc",
    direction: "out",
    remote: "+4930123456",
    startedAt: 1760000000000,
    text: "hello",
  });
  fetchImpl = async () => ({ ok: false, status: 500 });
  await assert.rejects(
    () => transcribe.saveTranscriptSegment({ callId: "x", text: "y" }),
    /HTTP 500/,
  );
});
