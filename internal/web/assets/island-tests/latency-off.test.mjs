// latency.js with the beacon intake OFF (no PBX_CONFIG.metrics): the
// island must never POST — the fire-and-forget instrument degrades to a
// no-op, the flag-off product stays byte-identical. Config exports
// compute at import time, so the OFF case lives in its own process (the
// audioout/audioout-unsupported precedent).
import test from "node:test";
import assert from "node:assert/strict";

import { installBrowserGlobals } from "./helpers.mjs";

installBrowserGlobals();

delete globalThis.window.PBX_CONFIG;

const posts = [];
globalThis.fetch = async (path, options) => {
  posts.push({ path, options });
  return { ok: true, status: 204 };
};

const { sendLatencyBeacon } = await import(`../island/app/latency.js?case=metrics-off`);
const { watchRemoteAudio } = await import(`../island/app/audio.js?case=metrics-off`);

const flush = () => new Promise((resolve) => setImmediate(resolve));

test("sendLatencyBeacon is a no-op without the metrics flag", async () => {
  sendLatencyBeacon("dial_to_audio", { signaling: 100 });
  sendLatencyBeacon("asr_round_trip", { parse_to_insert: 1 }, 4000);
  await flush();
  assert.equal(posts.length, 0, "no beacon may leave the island");
});

test("watchRemoteAudio answers a no-op disposer without a WebAudio graph", () => {
  // The base helper stub has no AudioContext at all: feature detection
  // must degrade to a callable disposer, never throw.
  let fired = 0;
  const stop = watchRemoteAudio({ fake: true }, () => {
    fired += 1;
  });
  assert.equal(typeof stop, "function");
  stop();
  assert.equal(fired, 0);
});
