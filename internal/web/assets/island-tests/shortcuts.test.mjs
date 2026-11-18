// shortcuts.js's keymap as a whole (ring-queue.test.mjs pins the Esc
// queue semantics and the typing gate; this file pins the REST of the
// map): the letter keys a/h/m/p and the focus brackets [ ] all act
// (preventDefault fired), unhandled keys pass through untouched, and
// every entry stays inert under ctrl/meta/alt — those belong to the
// browser, and the shell's own Ctrl/Cmd-K palette rides shell.js.
import test from "node:test";
import assert from "node:assert/strict";

import { installBrowserGlobals } from "./helpers.mjs";

installBrowserGlobals();
globalThis.Element = class Element {
  closest() {
    return null;
  }
};
globalThis.SIP = {
  SessionState: {
    Establishing: "Establishing",
    Established: "Established",
    Terminating: "Terminating",
    Terminated: "Terminated",
  },
  Inviter: class {},
  RegistererState: { Initial: "Initial", Registered: "Registered" },
};
const node = () => ({
  connect() {
    return node();
  },
  start() {},
  stop() {},
});
globalThis.AudioContext = class {
  constructor() {
    this.currentTime = 0;
    this.destination = {};
  }
  createOscillator() {
    return { frequency: {}, connect: node, start() {}, stop() {} };
  }
  createGain() {
    return { gain: {}, connect: node };
  }
};

await import("../island/app/state.js");
await import("../island/app/calls.js");
const { initShortcuts } = await import("../island/app/shortcuts.js");
initShortcuts();

const keydown = (key, extra = {}) => {
  const event = {
    type: "keydown",
    key,
    ctrlKey: false,
    metaKey: false,
    altKey: false,
    target: new globalThis.Element(),
    preventDefaultCalled: false,
    preventDefault() {
      this.preventDefaultCalled = true;
    },
    ...extra,
  };
  document.dispatchEvent(event);
  return event;
};

test("the letter and bracket keys are the acting map", () => {
  for (const key of ["a", "A", "h", "H", "m", "M", "p", "P", "[", "]"]) {
    const event = keydown(key);
    assert.ok(event.preventDefaultCalled, `key ${key} must act (preventDefault)`);
  }
});

test("media keys act too", () => {
  for (const key of ["MediaPlayPause", "MediaStop"]) {
    const event = keydown(key);
    assert.ok(event.preventDefaultCalled, `key ${key} must act`);
  }
});

test("unmapped keys pass through untouched", () => {
  for (const key of ["x", "1", "Enter", "ArrowDown"]) {
    const event = keydown(key);
    assert.ok(!event.preventDefaultCalled, `unmapped key ${key} must be left to the browser`);
  }
});

test("modifier chords never act", () => {
  for (const mod of ["ctrlKey", "metaKey", "altKey"]) {
    const event = keydown("h", { [mod]: true });
    assert.ok(!event.preventDefaultCalled, `${mod}+h belongs to the browser`);
  }
});
