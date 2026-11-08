// latency.js + the audio.js RMS watcher under node:test, metrics ON
// (PBX_CONFIG.metrics = true): the beacon posts the exact wire shape to
// /api/metrics/latency, omits the optional bytes for non-numeric values,
// and fires exactly once per completed interval; the watcher stamps on
// the first non-silent sample and the disposer stops the poll. The
// metrics-OFF no-op lives in latency-off.test.mjs (config.js computes
// its exports at import time, so one process carries one gate state —
// the audioout/audioout-unsupported precedent).
import test from "node:test";
import assert from "node:assert/strict";

import { installBrowserGlobals } from "./helpers.mjs";

installBrowserGlobals();

globalThis.window.PBX_CONFIG = { metrics: true };

const posts = [];
globalThis.fetch = async (path, options) => {
  posts.push({ path, options });
  return { ok: true, status: 204 };
};

// A scripted WebAudio graph: the analyser's first tick answers silence
// (all bytes at the 128 midpoint), the second answers one sample far off
// the midpoint — the watcher must fire on THAT tick, exactly once.
class ScriptedAnalyser {
  constructor() {
    this.fftSize = 512;
    this.ticks = 0;
    this.disconnects = 0;
  }
  getByteTimeDomainData(buf) {
    this.ticks += 1;
    const value = this.ticks >= 2 ? 200 : 128;
    for (let i = 0; i < buf.length; i++) buf[i] = value;
  }
  disconnect() {
    this.disconnects += 1;
  }
  connect() {}
}
class ScriptedSource {
  constructor() {
    this.disconnects = 0;
  }
  connect() {}
  disconnect() {
    this.disconnects += 1;
  }
}
let analyser;
globalThis.AudioContext = class {
  constructor() {
    this.destination = {};
    this.state = "running";
  }
  createMediaStreamSource() {
    this.source = new ScriptedSource();
    return this.source;
  }
  createAnalyser() {
    analyser = new ScriptedAnalyser();
    return analyser;
  }
};

const { sendLatencyBeacon } = await import(
  `../island/app/latency.js?case=metrics-on`
);
const { watchRemoteAudio } = await import(
  `../island/app/audio.js?case=metrics-on`
);

const flush = () => new Promise((resolve) => setImmediate(resolve));

test("sendLatencyBeacon posts the exact wire shape to the beacon route", async () => {
  sendLatencyBeacon("dial_to_audio", {
    island_overhead: 12,
    signaling: 1500,
    media_attach: 80,
    audio_gap: 240,
  });
  await flush();
  assert.equal(posts.length, 1);
  assert.equal(posts[0].path, "/api/metrics/latency");
  assert.equal(posts[0].options.method, "POST");
  assert.deepEqual(JSON.parse(posts[0].options.body), {
    interval: "dial_to_audio",
    legs_ms: { island_overhead: 12, signaling: 1500, media_attach: 80, audio_gap: 240 },
  });
});

test("sendLatencyBeacon carries bytes only when a non-negative size rode in", async () => {
  posts.length = 0;
  sendLatencyBeacon("asr_round_trip", { capture_to_post: 3 }, 4000);
  await flush();
  assert.deepEqual(JSON.parse(posts[0].options.body).bytes, 4000);
  posts.length = 0;
  sendLatencyBeacon("asr_round_trip", { parse_to_insert: 2 });
  await flush();
  assert.equal("bytes" in JSON.parse(posts[0].options.body), false);
});

test("watchRemoteAudio fires once on the first non-silent sample; the disposer stops it", async () => {
  let fired = 0;
  const stream = { fake: true };
  const stop = watchRemoteAudio(stream, () => {
    fired += 1;
  });
  assert.equal(typeof stop, "function", "feature-detection passed: a real disposer");
  await new Promise((resolve) => setTimeout(resolve, 130));
  assert.equal(fired, 1, "exactly one stamp after the scripted second tick");
  await new Promise((resolve) => setTimeout(resolve, 60));
  assert.equal(fired, 1, "the watcher is once-only");
  assert.equal(analyser.disconnects, 0, "the fired watcher cleaned itself up");
  // A second watcher's disposer detaches the graph without firing.
  let second = 0;
  const stop2 = watchRemoteAudio(stream, () => {
    second += 1;
  });
  stop2();
  await new Promise((resolve) => setTimeout(resolve, 60));
  assert.equal(second, 0, "a disposed watcher never fires");
  stop();
});
