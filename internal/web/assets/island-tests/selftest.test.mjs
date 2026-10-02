// selftest.js under node:test — the pre-call device check (M25 A10):
// the happy path proves the mic and releases it, the failure path names
// the error, and a browser without mediaDevices degrades honestly.
import test from "node:test";
import assert from "node:assert/strict";

import { installBrowserGlobals } from "./helpers.mjs";

const doc = installBrowserGlobals();
const { runDeviceCheck, initDeviceCheck } = await import("../island/app/selftest.js");

const logTexts = () => doc.getElementById("log").children.map((li) => li.textContent);

const stubMedia = ({ gum, devices } = {}) => {
  const stopped = [];
  const tracks = [{ label: "USB Mic", stop: () => stopped.push("USB Mic") }];
  const stream = { getTracks: () => tracks };
  Object.defineProperty(globalThis, "navigator", {
    value: {
      language: "en-US",
      mediaDevices: {
        getUserMedia: () => (gum ? Promise.reject(gum) : Promise.resolve(stream)),
        enumerateDevices: () =>
          Promise.resolve(
            devices || [
              { kind: "audiooutput", deviceId: "a" },
              { kind: "audioinput", deviceId: "b" },
              { kind: "audiooutput", deviceId: "c" },
            ],
          ),
      },
    },
    configurable: true,
  });
  return { stopped };
};

test("the device check proves the mic, releases it, and counts speakers", async () => {
  doc.getElementById("log").replaceChildren();
  const { stopped } = stubMedia();
  await runDeviceCheck();
  const texts = logTexts().join("\n");
  assert.match(texts, /device check: microphone OK \(USB Mic\)/);
  assert.match(texts, /device check: 2 audio output\(s\) visible/);
  assert.deepEqual(stopped, ["USB Mic"], "the borrowed stream is released");
});

test("a refused microphone names the error and stops there", async () => {
  doc.getElementById("log").replaceChildren();
  stubMedia({ gum: Object.assign(new Error("denied"), { name: "NotAllowedError" }) });
  await runDeviceCheck();
  const texts = logTexts().join("\n");
  assert.match(texts, /microphone FAILED — NotAllowedError/);
  assert.doesNotMatch(texts, /audio output/);
});

test("a browser without mediaDevices degrades honestly", async () => {
  doc.getElementById("log").replaceChildren();
  Object.defineProperty(globalThis, "navigator", {
    value: { language: "en-US" },
    configurable: true,
  });
  await runDeviceCheck();
  assert.match(logTexts().join("\n"), /media devices unavailable/);
});

test("the button runs the check disabled-while-pending", async () => {
  doc.getElementById("log").replaceChildren();
  stubMedia();
  const button = doc.getElementById("wp-devtest-btn");
  initDeviceCheck();
  button.disabled = false;
  assert.ok(button.listeners.click, "the click handler is wired");
  button.listeners.click.forEach((fn) => fn());
  assert.equal(button.disabled, true, "pending disables the button");
  await new Promise((resolve) => setTimeout(resolve, 0));
  assert.equal(button.disabled, false, "settled re-enables the button");
  assert.match(logTexts().join("\n"), /microphone OK/);
});
