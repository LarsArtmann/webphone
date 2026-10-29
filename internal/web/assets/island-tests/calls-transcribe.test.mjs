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

  // One segment: the strip grows by one line, the delta is announced,
  // the text is persisted fire-and-forget to the owner-scoped store.
  recorder.emit({ size: 10, type: "audio/webm" });
  await flush();
  const transcriptEl = entry.dom.querySelector(".call-transcript");
  assert.equal(transcriptEl.children.length, 1, "the segment is one strip line");
  assert.match(transcriptEl.children[0].textContent, /hello world/);
  // F053: the strip is aria-hidden — the announce path already speaks
  // each segment politely; the strip must not double-speak.
  const strip = entry.dom.querySelector(".wp-captions");
  assert.ok(strip, "the strip wraps the lines when the seam is on");
  assert.equal(strip.getAttribute("aria-hidden"), "true");
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
  assert.match(
    entry.dom.querySelector(".call-transcript").children.at(-1).textContent,
    /still live/,
  );
  assert.ok(
    [...doc.getElementById("toasts").children].some((el) => /not saved/i.test(el.textContent)),
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
  assert.equal(
    fetchCalls.filter((c) => c.url === "/api/transcripts").length,
    1,
    "the queue tries one save at a time while failing",
  );

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

  // With text: the whole joined transcript lands on the clipboard. The
  // words arrive the honest way — two REAL segments through the capture
  // path — because the strip's lines are the copy source now.
  const words = ["erste worte", "zweite worte"];
  fetchImpl = async (url) => {
    if (String(url).startsWith("/api/transcribe")) {
      const text = words.length ? words[0] : "";
      if (words.length) words.shift();
      return { ok: true, status: 200, json: async () => ({ text }) };
    }
    return { ok: true, status: 204 };
  };
  const recorder = globalThis.__recorders.at(-1);
  recorder.emit({ size: 8, type: "audio/webm" });
  await flush(2);
  recorder.emit({ size: 8, type: "audio/webm" });
  await flush(2);
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

test("the transcribe button label follows the capture state across renders", async () => {
  const { renderCalls } = await import("../island/app/calls.js");
  const { t } = await import("../island/app/i18n.js");
  const { startLiveTranscription, stopLiveTranscription } =
    await import("../island/app/transcribe.js");

  globalThis.SIP.Inviter = class {
    constructor() {
      this.id = "call-label";
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
  };
  fetchImpl = async () => ({ ok: true, status: 204 });
  state.userAgent = {};
  assert.equal(await placeCall("+4930887766"), true);
  const entry = sessions.get("call-label");
  assert.ok(entry, "the call entry exists");

  const btn = { textContent: "" };
  entry.dom.querySelector = (selector) =>
    selector === ".transcribe-btn" ? btn : { textContent: "" };

  renderCalls();
  assert.equal(btn.textContent, t("transcribe"), "idle capture renders the start label");

  startLiveTranscription("call-label", entry, () => {});
  renderCalls();
  assert.equal(btn.textContent, t("transcribeStop"), "live capture flips the label to stop");

  stopLiveTranscription("call-label");
  renderCalls();
  assert.equal(btn.textContent, t("transcribe"), "torn-down capture returns the start label");
});

test("the live card's transcript DOM is capped, keeping the newest tail (#41)", async () => {
  // The label test above leaves its own Inviter stub (id "call-label")
  // installed; re-stub the file's original shape so this test's entry
  // is addressable as call-t1 again.
  globalThis.SIP.Inviter = class {
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
  };
  state.userAgent = {};
  assert.equal(await placeCall("+49301110000"), true);
  const entry = sessions.get("call-t1");

  let segment = 0;
  fetchImpl = async (url) => {
    if (String(url).startsWith("/api/transcribe")) {
      const n = (segment += 1);
      return { ok: true, status: 200, json: async () => ({ text: `seg-${n}` }) };
    }
    return { ok: true, status: 204 };
  };
  entry.session.state = globalThis.SIP.SessionState.Established;
  entry.session.stateChange.listeners.forEach((fn) => fn("Established"));
  await flush();
  const recorder = globalThis.__recorders.at(-1);
  // 35 segments: past the strip's 30-line cap. The kept tail is the
  // NEWEST 30 lines; the durable whole lives server-side (History).
  for (let i = 0; i < 35; i += 1) {
    recorder.emit({ size: 10, type: "audio/webm" });
    await flush();
  }
  const transcriptEl = entry.dom.querySelector(".call-transcript");
  assert.equal(transcriptEl.children.length, 30, "the strip stays bounded at 30 lines");
  assert.equal(
    transcriptEl.children[0].textContent,
    "seg-6",
    "the cap drops the OLDEST head (seg-1..5)",
  );
  assert.equal(
    transcriptEl.children.at(-1).textContent,
    "seg-35",
    "the newest segment is the last line",
  );

  entry.session.state = globalThis.SIP.SessionState.Terminated;
  entry.session.stateChange.listeners.forEach((fn) => fn("Terminated"));
  await flush();
});

test("the card's clear button wipes the live view without touching capture (#39)", async () => {
  state.userAgent = {};
  assert.equal(await placeCall("+49301110001"), true);
  const entry = [...sessions.values()].at(-1);
  const transcriptEl = entry.dom.querySelector(".call-transcript");
  const clearBtn = entry.dom.querySelector(".clear-transcript-btn");
  assert.ok(clearBtn, "the ASR-gated card carries the clear affordance");
  const click = () => (clearBtn.listeners.click ?? []).forEach((fn) => fn({}));

  // Clearing an empty transcript announces that honestly.
  click();
  await flush(2);
  assert.ok(
    [...doc.getElementById("toasts").children].some((el) =>
      /no speech detected/i.test(el.textContent),
    ),
  );

  // With text: the view empties, the announce scopes itself to the
  // LIVE card (the stored History transcript is a different surface).
  // The line arrives the honest way: one REAL segment.
  fetchImpl = async (url) => {
    if (String(url).startsWith("/api/transcribe")) {
      return { ok: true, status: 200, json: async () => ({ text: "alter text" }) };
    }
    return { ok: true, status: 204 };
  };
  entry.session.state = globalThis.SIP.SessionState.Established;
  entry.session.stateChange.listeners.forEach((fn) => fn("Established"));
  await flush();
  const recorder = globalThis.__recorders.at(-1);
  recorder.emit({ size: 8, type: "audio/webm" });
  await flush(2);
  assert.equal(transcriptEl.children.length, 1, "the segment rendered");
  click();
  await flush(2);
  assert.equal(transcriptEl.children.length, 0, "the strip emptied");
  assert.equal(transcriptEl.textContent, "");
  assert.ok(
    [...doc.getElementById("toasts").children].some((el) =>
      /live transcript cleared/i.test(el.textContent),
    ),
  );

  entry.session.state = globalThis.SIP.SessionState.Terminated;
  entry.session.stateChange.listeners.forEach((fn) => fn("Terminated"));
  await flush();
});

test("a confirmed save surfaces the saved-to-History count, announced once", async () => {
  state.userAgent = {};
  assert.equal(await placeCall("+4930123456"), true);
  const entry = sessions.get("call-t1");
  fetchImpl = async (url) => {
    if (String(url).startsWith("/api/transcribe")) {
      return { ok: true, status: 200, json: async () => ({ text: "hi" }) };
    }
    return { ok: true, status: 204 };
  };
  entry.session.state = globalThis.SIP.SessionState.Established;
  entry.session.stateChange.listeners.forEach((fn) => fn("Established"));
  await flush();
  const [recorder] = globalThis.__recorders.slice(-1);
  const savedToasts = () =>
    [...doc.getElementById("toasts").children].filter((el) =>
      /Saved to History/i.test(el.textContent),
    ).length;

  recorder.emit({ size: 8, type: "audio/webm" });
  await flush();
  const marker = entry.dom.querySelector(".call-transcript-saved");
  assert.ok(marker, "the marker node exists when the seam is on");
  assert.equal(marker.hidden, false, "the first confirmed save reveals it");
  assert.match(marker.textContent, /Saved to History — 1 segment/);
  assert.equal(savedToasts(), 1, "the first save announces politely");

  recorder.emit({ size: 8, type: "audio/webm" });
  await flush();
  assert.match(marker.textContent, /2 segments/);
  assert.equal(
    savedToasts(),
    1,
    "the count is the ongoing affordance; the announce fires once per call",
  );

  entry.session.state = globalThis.SIP.SessionState.Terminated;
  entry.session.stateChange.listeners.forEach((fn) => fn("Terminated"));
  await flush();
});

test("a segment that lands after the call ended still saves but never announces", async () => {
  state.userAgent = {};
  // Shared module state: an earlier spec may have left the transcribe
  // backoff armed. A clock past every window keeps this about staleness.
  const realNow = Date.now;
  const base = realNow();
  Date.now = () => base + 7_200_000;
  try {
    assert.equal(await placeCall("+4930999"), true);
    const entry = sessions.get("call-t1");
    let release = () => {};
    fetchCalls = [];
    fetchImpl = (url) => {
      if (String(url).startsWith("/api/transcribe")) {
        return new Promise((resolve) => {
          release = () =>
            resolve({
              ok: true,
              status: 200,
              json: async () => ({ text: "late tail" }),
            });
        });
      }
      return Promise.resolve({ ok: true, status: 204 });
    };

    entry.session.state = globalThis.SIP.SessionState.Established;
    entry.session.stateChange.listeners.forEach((fn) => fn("Established"));
    await flush();

    const recorder = globalThis.__recorders.at(-1);
    recorder.emit({ size: 10, type: "audio/webm" });
    await flush();

    // The call ends while the provider is still answering.
    entry.session.state = globalThis.SIP.SessionState.Terminated;
    entry.session.stateChange.listeners.forEach((fn) => fn("Terminated"));
    await flush();
    assert.equal(sessions.has("call-t1"), false, "the call is settled");

    release();
    await flush();

    const saves = fetchCalls
      .filter((call) => call.url === "/api/transcripts")
      .map((call) => JSON.parse(call.options.body));
    assert.equal(saves.length, 1, "the tail still reaches the durable store");
    assert.equal(saves[0].text, "late tail");
    assert.equal(saves[0].callId, "call-t1", "attribution survives the race");
    assert.ok(
      ![...doc.getElementById("toasts").children].some((el) =>
        /late tail/i.test(el.textContent),
      ),
      "a settled call must not speak a caption for audio nobody is on",
    );
  } finally {
    Date.now = realNow;
  }
});
