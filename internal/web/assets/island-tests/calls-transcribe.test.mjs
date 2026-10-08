// calls.js live-transcription wiring under node:test: auto-start on
// connect (owner decision 2026-10-07), per-segment persistence to
// /api/transcripts, polite announcements, and teardown with the call.
// The ASR gate is forced ON before any island import (config.js reads
// window.PBX_CONFIG at import time).
import test from "node:test";
import assert from "node:assert/strict";

import { installBrowserGlobals } from "./helpers.mjs";

const doc = installBrowserGlobals();
globalThis.window.PBX_CONFIG = { asr: true, sipDomain: "sip.test" };

globalThis.SIP = {
  SessionState: {
    Establishing: "Establishing",
    Established: "Established",
    Terminating: "Terminating",
    Terminated: "Terminated",
  },
  Inviter: class {
    constructor() {
      this.id = "call-t1";
      this.state = globalThis.SIP.SessionState.Establishing;
      this.stateChange = {
        listeners: [],
        addListener(fn) {
          this.listeners.push(fn);
        },
      };
      this.sessionDescriptionHandler = {
        peerConnection: {
          getReceivers: () => [{ track: { kind: "audio" } }],
          getSenders: () => [{ track: { kind: "audio" } }],
        },
      };
    }
    invite() {
      return Promise.resolve();
    }
    async bye() {}
  },
};
globalThis.SIP.UserAgent = { makeURI: (raw) => ({ toString: () => raw }) };

// One AudioContext stub serves both consumers: audio.js's ringback and
// transcribe.js's leg mixing (the mixing methods track their tracks so
// the teardown spec can prove the capture graph is released).
const node = () => ({
  connect() {
    return node();
  },
  start() {},
  stop() {},
});
globalThis.__contexts = [];
globalThis.AudioContext = class {
  constructor() {
    this.currentTime = 0;
    this.destination = {};
    globalThis.__contexts.push(this);
  }
  createOscillator() {
    return { frequency: {}, connect: node, start() {}, stop() {} };
  }
  createGain() {
    return { gain: {}, connect: node };
  }
  createMediaStreamDestination() {
    const tracks = [
      {
        stopped: false,
        stop() {
          this.stopped = true;
        },
      },
    ];
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
// transcribe.js's capture reads window.AudioContext (the helpers' window
// is a distinct object from globalThis — both must see the same class).
globalThis.window.AudioContext = globalThis.AudioContext;
globalThis.MediaStream = class {
  constructor() {}
  addTrack() {}
};
globalThis.__recorders = [];
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
    this.state = "inactive";
    for (const fn of this.listeners.stop ?? []) fn({});
  }
  emit(data) {
    for (const fn of this.listeners.dataavailable ?? []) fn({ data });
  }
};

let fetchCalls = [];
let fetchImpl = async () => ({ ok: true, status: 200, json: async () => ({ text: "" }) });
globalThis.fetch = async (url, options = {}) => {
  fetchCalls.push({ url, options });
  return fetchImpl(url, options);
};

const { placeCall } = await import("../island/app/calls.js");
const { sessions, state } = await import("../island/app/state.js");
const { isTranscribing } = await import("../island/app/transcribe.js");
const { els } = await import("../island/app/ui.js");

// The fake session carries a real-shaped sessionDescriptionHandler (the
// capture path needs it), so focusSession reaches remote-audio playback
// — give the stub element an unlocked play().
els.remoteAudio.play = () => Promise.resolve();

const flush = async (rounds = 12) => {
  while (rounds--) await new Promise((resolve) => setImmediate(resolve));
};

test("live transcription auto-starts on connect, persists segments, and tears down with the call", async () => {
  state.userAgent = {};
  assert.equal(await placeCall("+4930123456"), true);
  const entry = sessions.get("call-t1");
  assert.equal(isTranscribing("call-t1"), false, "no capture before media flows");

  fetchImpl = async (url) => {
    if (String(url).startsWith("/api/transcribe")) {
      return { ok: true, status: 200, json: async () => ({ text: "hello world" }) };
    }
    return { ok: true, status: 204 };
  };
  fetchCalls = [];
  entry.session.state = globalThis.SIP.SessionState.Established;
  entry.session.stateChange.listeners.forEach((fn) => fn("Established"));
  await flush();

  assert.equal(isTranscribing("call-t1"), true, "transcription begins with the call");
  const [recorder] = globalThis.__recorders;
  assert.ok(recorder, "a recorder was created");
  assert.equal(recorder.startedWith, 4000);
  const toasts = doc.getElementById("toasts").children;
  assert.ok(
    toasts.some((el) => /live transcription started/i.test(el.textContent)),
    "the auto-start announces itself politely",
  );

  // One segment: the card grows, the delta is announced, the text is
  // persisted fire-and-forget to the owner-scoped store.
  recorder.emit({ size: 10, type: "audio/webm" });
  await flush();
  const transcriptEl = entry.dom.querySelector(".call-transcript");
  assert.match(transcriptEl.textContent, /hello world/);
  const saved = fetchCalls.find((c) => c.url === "/api/transcripts");
  assert.ok(saved, "segment persisted via POST /api/transcripts");
  assert.deepEqual(JSON.parse(saved.options.body), {
    callId: "call-t1",
    direction: "out",
    remote: "+4930123456",
    startedAt: entry.startedAt,
    text: "hello world",
  });
  assert.ok(
    [...doc.getElementById("toasts").children].some((el) =>
      /Transcript: hello world/i.test(el.textContent),
    ),
    "the segment delta rides the polite live region",
  );

  // Teardown with the call: capture ends, the mixing context closes.
  const mixCtx = globalThis.__contexts.find((ctx) => ctx.mixedTracks);
  entry.session.state = globalThis.SIP.SessionState.Terminated;
  entry.session.stateChange.listeners.forEach((fn) => fn("Terminated"));
  await flush();
  assert.equal(isTranscribing("call-t1"), false);
  assert.equal(recorder.stopped, true);
  assert.equal(sessions.has("call-t1"), false);
  assert.ok(mixCtx.closed, "the mixing AudioContext is closed on teardown");
});

test("a failed save warns once per call and never blocks the capture", async () => {
  state.userAgent = {};
  assert.equal(await placeCall("+49800123456"), true);
  const entry = sessions.get("call-t1");
  fetchImpl = async (url) => {
    if (String(url).startsWith("/api/transcribe")) {
      return { ok: true, status: 200, json: async () => ({ text: "still live" }) };
    }
    return { ok: false, status: 500 };
  };
  entry.session.state = globalThis.SIP.SessionState.Established;
  entry.session.stateChange.listeners.forEach((fn) => fn("Established"));
  await flush();
  const [recorder] = globalThis.__recorders.slice(-1);
  const toastCount = () => doc.getElementById("toasts").children.length;

  recorder.emit({ size: 8, type: "audio/webm" });
  await flush();
  assert.match(entry.dom.querySelector(".call-transcript").textContent, /still live/);
  assert.ok(
    [...doc.getElementById("toasts").children].some((el) =>
      /not saved/i.test(el.textContent),
    ),
    "the failed persistence surfaces once",
  );

  const afterFirst = toastCount();
  recorder.emit({ size: 8, type: "audio/webm" });
  await flush();
  assert.equal(toastCount(), afterFirst, "subsequent failures stay quiet (log only)");

  entry.session.state = globalThis.SIP.SessionState.Terminated;
  entry.session.stateChange.listeners.forEach((fn) => fn("Terminated"));
  await flush();
});

test("failed saves queue and flush in order on the next success", async () => {
  state.userAgent = {};
  assert.equal(await placeCall("+49800123456"), true);
  const entry = sessions.get("call-t1");
  let savesOk = false;
  fetchImpl = async (url) => {
    if (String(url).startsWith("/api/transcribe")) {
      return { ok: true, status: 200, json: async () => ({ text: "words" }) };
    }
    return savesOk ? { ok: true, status: 204 } : { ok: false, status: 503 };
  };
  fetchCalls = [];
  entry.session.state = globalThis.SIP.SessionState.Established;
  entry.session.stateChange.listeners.forEach((fn) => fn("Established"));
  await flush();
  const [recorder] = globalThis.__recorders.slice(-1);

  recorder.emit({ size: 8, type: "audio/webm" });
  recorder.emit({ size: 8, type: "audio/webm" });
  await flush(20);
  assert.equal(fetchCalls.filter((c) => c.url === "/api/transcripts").length, 1,
    "the queue tries one save at a time while failing");

  // The network heals: the next segment flushes the whole queue, in order.
  savesOk = true;
  fetchCalls = [];
  recorder.emit({ size: 8, type: "audio/webm" });
  await flush(30);
  const bodies = fetchCalls
    .filter((c) => c.url === "/api/transcripts")
    .map((c) => JSON.parse(c.options.body).text);
  assert.equal(bodies.length, 3, "queued + new segment all persisted");
  assert.deepEqual(bodies, ["words", "words", "words"], "order preserved, none dropped");

  entry.session.state = globalThis.SIP.SessionState.Terminated;
  entry.session.stateChange.listeners.forEach((fn) => fn("Terminated"));
  await flush();
});

test("the card's copy button copies the accumulated transcript honestly", async () => {
  state.userAgent = {};
  assert.equal(await placeCall("+49123998877"), true);
  const entry = sessions.get("call-t1");
  const transcriptEl = entry.dom.querySelector(".call-transcript");
  const copyBtn = entry.dom.querySelector(".copy-transcript-btn");
  assert.ok(copyBtn, "the ASR-gated card carries the copy affordance");
  const click = () => (copyBtn.listeners.click ?? []).forEach((fn) => fn({}));
  const toasts = () => [...doc.getElementById("toasts").children];

  // Empty transcript: honest announce, nothing hits the clipboard.
  const copied = [];
  globalThis.navigator.clipboard = {
    writeText: (text) => {
      copied.push(text);
      return Promise.resolve();
    },
  };
  click();
  await flush(2);
  assert.equal(copied.length, 0);
  assert.ok(
    toasts().some((el) => /no speech detected/i.test(el.textContent)),
    "an empty transcript announces that instead of copying nothing",
  );

  // With text: the whole joined transcript lands on the clipboard.
  transcriptEl.textContent = "erste worte zweite worte";
  click();
  await flush(2);
  assert.deepEqual(copied, ["erste worte zweite worte"]);
  assert.ok(
    toasts().some((el) => /copied to clipboard/i.test(el.textContent)),
    "the copy confirms itself politely",
  );

  // A clipboard failure announces the manual fallback.
  globalThis.navigator.clipboard = {
    writeText: () => Promise.reject(new Error("denied")),
  };
  click();
  await flush(2);
  assert.ok(
    toasts().some((el) => /could not copy/i.test(el.textContent)),
    "the failure says what to do instead",
  );

  delete globalThis.navigator.clipboard;
  entry.session.state = globalThis.SIP.SessionState.Terminated;
  entry.session.stateChange.listeners.forEach((fn) => fn("Terminated"));
  await flush();
});

test("the save queue is bounded: oldest segments drop past the cap", async () => {
  state.userAgent = {};
  assert.equal(await placeCall("+49800123456"), true);
  const entry = sessions.get("call-t1");
  let savesOk = false;
  let segment = 0;
  fetchImpl = async (url) => {
    if (String(url).startsWith("/api/transcribe")) {
      // Capture eagerly: the lazy json() reads LATER (all 52 transcribe
      // fetches are initiated before any .json() resolves), so closing
      // over the live counter would stamp every segment with the LAST
      // value and make the drop-oldest assertions meaningless.
      const n = (segment += 1);
      return { ok: true, status: 200, json: async () => ({ text: `seg-${n}` }) };
    }
    return savesOk ? { ok: true, status: 204 } : { ok: false, status: 500 };
  };
  entry.session.state = globalThis.SIP.SessionState.Established;
  entry.session.stateChange.listeners.forEach((fn) => fn("Established"));
  await flush();
  const [recorder] = globalThis.__recorders.slice(-1);

  // 52 failing segments: the queue keeps the newest 50 (the first flush
  // attempt consumes #1, leaving 51 queued -> capped to 50).
  for (let i = 0; i < 52; i++) recorder.emit({ size: 6, type: "audio/webm" });
  await flush(20);

  savesOk = true;
  fetchCalls = [];
  recorder.emit({ size: 6, type: "audio/webm" });
  await flush(400);
  const savedTexts = fetchCalls
    .filter((c) => c.url === "/api/transcripts")
    .map((c) => JSON.parse(c.options.body).text);
  assert.equal(savedTexts.length, 50, "exactly the capped queue drains");
  assert.equal(savedTexts[0], "seg-4", "the oldest beyond the cap were dropped (seg-1..3)");
  assert.equal(savedTexts[49], "seg-53", "the newest segment is last");

  entry.session.state = globalThis.SIP.SessionState.Terminated;
  entry.session.stateChange.listeners.forEach((fn) => fn("Terminated"));
  await flush();
});
