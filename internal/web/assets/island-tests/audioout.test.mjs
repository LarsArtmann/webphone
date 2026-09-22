// audioout.js under node:test (full environment): with setSinkId and two
// physical outputs the picker must populate, restore a stored pick,
// persist a new one, and survive device churn. The module's init guard
// is once-per-process, so one init drives all specs sequentially here;
// the "unsupported" branches live in audioout-unsupported.test.mjs.
import test from "node:test";
import assert from "node:assert/strict";

import { installBrowserGlobals } from "./helpers.mjs";

const doc = installBrowserGlobals();

const device = (kind, deviceId, label = "") => ({ kind, deviceId, label });

const flush = () => new Promise((resolve) => setImmediate(resolve));

const applied = [];
let outputs = [
  device("audioinput", "mic-1"),
  device("audiooutput", "default", "System default"),
  device("audiooutput", "spk-1", "Speakers"),
  device("audiooutput", "spk-2", "Headphones"),
];
const els = (await import("../island/app/ui.js")).els;
els.remoteAudio.setSinkId = (id) => {
  applied.push(id);
  return Promise.resolve();
};

globalThis.localStorage.setItem("wp-sink", "spk-2");

const listeners = new Map();
Object.defineProperty(globalThis, "navigator", {
  value: {
    language: "en-US",
    mediaDevices: {
      enumerateDevices: () => Promise.resolve(outputs),
      addEventListener(type, fn) {
        (listeners.get(type) ?? listeners.set(type, []).get(type)).push(fn);
      },
    },
  },
  configurable: true,
});

const { initAudioOutput, physicalOutputs } = await import("../island/app/audioout.js");
initAudioOutput();
await flush();

const select = doc.getElementById("audio-output");
const wrap = doc.getElementById("audio-out-wrap");

test("physicalOutputs keeps only addressable, non-alias audio outputs", () => {
  const outs = physicalOutputs([
    device("audioinput", "mic-1"),
    device("audiooutput", "default", "Default"),
    device("audiooutput", "communications", "Communications"),
    device("audiooutput", "", "Unaddressable"),
    device("audiooutput", "spk-1", "Speakers"),
    device("audiooutput", "spk-2", "Headphones"),
  ]);
  assert.deepEqual(
    outs.map((d) => d.deviceId),
    ["spk-1", "spk-2"],
  );
});

test("the picker appears with a real choice and restores the stored pick", () => {
  assert.equal(wrap.hidden, false, "two physical outputs unhide the picker");
  assert.equal(select.children.length, 3, "default option plus two outputs");
  assert.equal(select.children[0].textContent, "System default");
  assert.equal(select.children[2].textContent, "Headphones");
  assert.equal(select.value, "spk-2", "the stored pick is preselected");
  assert.deepEqual(applied, ["spk-2"], "the stored pick is applied to the remote audio");
});

test("choosing an output applies it to the remote audio and persists", () => {
  select.value = "spk-1";
  select.listeners.change.at(-1)({ target: select });
  assert.deepEqual(applied, ["spk-2", "spk-1"], "the pick is applied immediately");
  assert.equal(globalThis.localStorage.getItem("wp-sink"), "spk-1", "the pick persists");

  select.value = "";
  select.listeners.change.at(-1)({ target: select });
  assert.equal(
    globalThis.localStorage.getItem("wp-sink"),
    "",
    "returning to the default persists too",
  );
});

test("device churn keeps a live selection and falls back off a dead one", async () => {
  select.value = "spk-2";
  select.listeners.change.at(-1)({ target: select });

  // Device churn: spk-2 vanishes, a new one appears.
  outputs = [
    device("audiooutput", "spk-1", "Speakers"),
    device("audiooutput", "spk-3", "New dock"),
  ];
  for (const fn of listeners.get("devicechange") ?? []) {
    fn({});
  }
  await flush();

  assert.equal(select.children.at(-1).textContent, "New dock", "options are repopulated");
  assert.equal(select.value, "", "a vanished device falls back to the default");
  assert.equal(wrap.hidden, false, "the picker survives device churn");

  // A still-present selection survives the same churn untouched.
  select.value = "spk-1";
  select.listeners.change.at(-1)({ target: select });
  for (const fn of listeners.get("devicechange") ?? []) {
    fn({});
  }
  await flush();
  assert.equal(select.value, "spk-1", "a live selection is kept across churn");
});
