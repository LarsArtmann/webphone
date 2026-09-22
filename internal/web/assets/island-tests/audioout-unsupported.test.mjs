// audioout.js under node:test (unsupported environments, own process
// for module isolation): without setSinkId or with fewer than two
// physical outputs the picker must never appear and must stay empty.
import test from "node:test";
import assert from "node:assert/strict";

import { installBrowserGlobals } from "./helpers.mjs";

const doc = installBrowserGlobals();

Object.defineProperty(globalThis, "navigator", {
  value: {
    language: "en-US",
    // Enumeration says only one physical output (plus the default
    // alias) — no real choice, so the picker stays hidden.
    mediaDevices: {
      enumerateDevices: () =>
        Promise.resolve([
          { kind: "audiooutput", deviceId: "default", label: "Default" },
          { kind: "audiooutput", deviceId: "spk-1", label: "Speakers" },
        ]),
    },
  },
  configurable: true,
});

const { initAudioOutput } = await import("../island/app/audioout.js");
initAudioOutput();
await new Promise((resolve) => setImmediate(resolve));

test("a single physical output keeps the picker hidden", () => {
  const wrap = doc.getElementById("audio-out-wrap");
  const select = doc.getElementById("audio-output");
  assert.equal(wrap.hidden, true, "no choice, no picker");
  assert.equal(select.children.length, 0, "no options populated");
});
