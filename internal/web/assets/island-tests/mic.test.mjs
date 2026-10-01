// mic.js under node:test — the pre-warmed microphone state machine.
// The warm stream is acquired while a call RINGS so the mic is live the
// moment the user accepts; the factory hands it to sip.js exactly once
// (sip.js owns the tracks from there and stops them at hangup), and
// every non-answer exit (reject, caller gave up, logout) releases the
// device. The takeover rule: a take/release during a PENDING acquisition
// must stop the late stream instead of caching it — otherwise the mic
// indicator would stay lit with nobody owning the tracks.
import test from "node:test";
import assert from "node:assert/strict";

import { installBrowserGlobals } from "./helpers.mjs";

installBrowserGlobals();

globalThis.MediaStream = class MediaStream {};

const loadMic = (tag) => import(`../island/app/mic.js?case=${tag}`);

const flushes = async (rounds = 12) => {
  while (rounds--) await new Promise((resolve) => setImmediate(resolve));
};

const makeTrack = () => ({
  stopped: false,
  listeners: {},
  stop() {
    this.stopped = true;
  },
  addEventListener(type, fn) {
    (this.listeners[type] ??= []).push(fn);
  },
});

const makeStream = () => {
  const tracks = [makeTrack(), makeTrack()];
  return {
    tracks,
    getTracks() {
      return tracks;
    },
  };
};

const deferred = () => {
  let resolve;
  let reject;
  const promise = new Promise((res, rej) => {
    resolve = res;
    reject = rej;
  });
  return { promise, resolve, reject };
};

const setMediaDevices = (getUserMedia) => {
  Object.defineProperty(globalThis, "navigator", {
    value: {
      language: "en-US",
      mediaDevices: getUserMedia ? { getUserMedia } : undefined,
    },
    configurable: true,
  });
};

const recorder = () => {
  const calls = [];
  calls.streams = [];
  calls.fn = (constraints) => {
    calls.push(constraints);
    const stream = makeStream();
    calls.streams.push(stream);
    return Promise.resolve(stream);
  };
  return calls;
};

test("warm hands the stream off exactly once", async () => {
  const mic = await loadMic("handoff-once");
  const gum = recorder();
  setMediaDevices(gum.fn);

  mic.warmMic();
  await flushes();
  assert.equal(gum.length, 1);
  assert.deepEqual(gum[0], { audio: true, video: false });

  mic.warmMic();
  await flushes();
  assert.equal(gum.length, 1, "a second warm while warm is a no-op");

  const stream = mic.takeWarmMic();
  assert.equal(stream, gum.streams[0], "the warm stream is handed off");
  assert.equal(mic.takeWarmMic(), null, "the handoff is one-shot");
  assert.equal(
    stream.tracks.every((track) => !track.stopped),
    true,
    "handoff does not stop the tracks (sip.js owns them now)",
  );
});

test("a second warm during a pending acquisition does not re-acquire", async () => {
  const mic = await loadMic("double-warm-pending");
  const gum = recorder();
  gum.fn = (constraints) => {
    gum.push(constraints);
    return new Promise(() => {});
  };
  setMediaDevices(gum.fn);

  mic.warmMic();
  mic.warmMic();
  await flushes();
  assert.equal(gum.length, 1, "one device acquisition while pending");
});

test("takeover during a pending warm stops the late stream", async () => {
  const mic = await loadMic("takeover-pending");
  const late = deferred();
  const stream = makeStream();
  setMediaDevices(() => late.promise);

  mic.warmMic();
  assert.equal(mic.takeWarmMic(), null, "nothing warm yet");
  late.resolve(stream);
  await flushes();
  assert.equal(
    stream.tracks.every((track) => track.stopped),
    true,
    "the unowned late stream is released, not cached",
  );
});

test("release stops the warm stream's tracks", async () => {
  const mic = await loadMic("release-warm");
  const gum = recorder();
  setMediaDevices(gum.fn);

  mic.warmMic();
  await flushes();
  mic.releaseWarmMic();
  assert.equal(
    gum.streams[0].tracks.every((track) => track.stopped),
    true,
    "release returns the device",
  );
  assert.equal(mic.takeWarmMic(), null);
});

test("release during a pending warm stops the late stream", async () => {
  const mic = await loadMic("release-pending");
  const late = deferred();
  const stream = makeStream();
  setMediaDevices(() => late.promise);

  mic.warmMic();
  mic.releaseWarmMic();
  late.resolve(stream);
  await flushes();
  assert.equal(
    stream.tracks.every((track) => track.stopped),
    true,
    "the cancelled acquisition never lingers",
  );
});

test("an ended track invalidates the warm cache", async () => {
  const mic = await loadMic("ended-invalidate");
  const gum = recorder();
  setMediaDevices(gum.fn);

  mic.warmMic();
  await flushes();
  gum.streams[0].tracks[0].listeners.ended[0]();
  assert.equal(mic.takeWarmMic(), null, "a dead device is not handed off");
});

test("dialing after a missed call acquires fresh, never the stopped stream", async () => {
  const mic = await loadMic("dial-after-missed");
  const gum = recorder();
  setMediaDevices(gum.fn);

  mic.warmMic();
  await flushes();
  // The missed-call exit (connection.js fires releaseWarmMic when the
  // caller gives up): the warm stream's tracks stop.
  mic.releaseWarmMic();
  const handed = await mic.micMediaStreamFactory({ audio: true, video: false });
  assert.equal(gum.length, 2, "the dial triggers a new device acquisition");
  assert.equal(handed, gum.streams[1], "the outgoing call gets the fresh stream");
  assert.equal(
    gum.streams[0].tracks.every((track) => track.stopped),
    true,
    "the missed call's stream stays released",
  );
});

test("factory hands the warm stream to audio-only constraints", async () => {
  const mic = await loadMic("factory-warm");
  const gum = recorder();
  setMediaDevices(gum.fn);

  mic.warmMic();
  await flushes();
  const handed = await mic.micMediaStreamFactory({
    audio: true,
    video: false,
  });
  assert.equal(handed, gum.streams[0], "the warm stream itself, not a copy");
  assert.equal(gum.length, 1, "no acquisition at factory time");
});

test("factory falls back to getUserMedia for video and keeps the warm stream", async () => {
  const mic = await loadMic("factory-video");
  const gum = recorder();
  setMediaDevices(gum.fn);

  mic.warmMic();
  await flushes();
  const constraints = { audio: true, video: true };
  const fallback = await mic.micMediaStreamFactory(constraints);
  assert.equal(fallback, gum.streams[1], "fallback used getUserMedia");
  assert.deepEqual(gum[1], constraints);
  assert.equal(mic.takeWarmMic(), gum.streams[0], "warm not consumed");
});

test("factory mirrors the empty-constraints default without acquiring", async () => {
  const mic = await loadMic("factory-empty");
  const gum = recorder();
  setMediaDevices(gum.fn);

  const empty = await mic.micMediaStreamFactory({ audio: false, video: false });
  assert.ok(empty instanceof globalThis.MediaStream);
  assert.equal(gum.length, 0, "no device acquisition");
  assert.equal(mic.takeWarmMic(), null);
});

test("factory rejects without media devices like the upstream default", async () => {
  const mic = await loadMic("factory-insecure");
  setMediaDevices(null);

  await assert.rejects(() => mic.micMediaStreamFactory({ audio: true, video: false }), {
    message: "Media devices not available in insecure contexts.",
  });
});

test("warm is a silent no-op without media devices", async () => {
  const mic = await loadMic("warm-insecure");
  setMediaDevices(null);

  mic.warmMic();
  await flushes();
  assert.equal(mic.takeWarmMic(), null);
});

test("a timed warm expires and returns the device", async () => {
  const mic = await loadMic("ttl-expires");
  const gum = recorder();
  setMediaDevices(gum.fn);

  mic.warmMic(20);
  await flushes();
  assert.equal(gum.length, 1, "the dial warm acquired the mic");
  await new Promise((resolve) => setTimeout(resolve, 60));
  assert.equal(
    gum.streams[0].tracks.every((track) => track.stopped),
    true,
    "an expired dial warm releases the mic (abandoned dial intent decays)",
  );
  assert.equal(mic.takeWarmMic(), null);
});

test("consuming a timed warm before expiry keeps the stream alive", async () => {
  const mic = await loadMic("ttl-consumed");
  const gum = recorder();
  setMediaDevices(gum.fn);

  mic.warmMic(25);
  await flushes();
  const stream = mic.takeWarmMic();
  await new Promise((resolve) => setTimeout(resolve, 60));
  assert.equal(
    stream.tracks.every((track) => !track.stopped),
    true,
    "a consumed warm is never expired — sip.js owns the tracks",
  );
  assert.equal(mic.takeWarmMic(), null);
});

test("a stale dial-warm timer never kills a later warm", async () => {
  const mic = await loadMic("ttl-stale-timer");
  const gum = recorder();
  setMediaDevices(gum.fn);

  mic.warmMic(25);
  await flushes();
  // The warm device dies on its own (unplug): the cache invalidates but
  // the expiry timer stays armed — then an UNTIMED incoming warm lands.
  gum.streams[0].tracks.forEach((track) => {
    (track.listeners.ended ?? []).forEach((fn) => fn());
  });
  assert.equal(mic.takeWarmMic(), null);
  mic.warmMic();
  await flushes();
  await new Promise((resolve) => setTimeout(resolve, 70));
  const later = mic.takeWarmMic();
  assert.equal(later, gum.streams[1], "the later warm survives the stale timer");
  assert.equal(
    later.tracks.every((track) => !track.stopped),
    true,
    "the identity guard held",
  );
});
