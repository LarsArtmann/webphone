// The verdict lane's failure arms (the deepening plan's T15): rejected
// and timeout each settle the toggle back to the PRIOR state and
// announce the honest reason with the toggle's direction. The
// accept/refuse-resume/send-failed/pending arms live in calls.test.mjs
// (the UI-truth spec); this file owns the arm only the stubs make
// reachable — the 30s verdict timeout under mock timers — plus the
// en/de copy pins for all three verdict words.
import test from "node:test";
import assert from "node:assert/strict";

import { installBrowserGlobals } from "./helpers.mjs";

const doc = installBrowserGlobals();
globalThis.window.PBX_CONFIG = { sipDomain: "sip.test" };
globalThis.SIP = {
  SessionState: {
    Establishing: "Establishing",
    Established: "Established",
    Terminating: "Terminating",
    Terminated: "Terminated",
  },
};

const settle = async () => {
  for (let i = 0; i < 12; i++) await new Promise((r) => setImmediate(r));
};

/**
 * The established-call fixture: placeCall into a re-INVITE stub whose
 * FIRST toggle is gated (the test holds the handle and settles it
 * explicitly).
 *
 * @returns {Promise<{entry: object, refuse: Function, accept: Function}>}
 */
async function establishedCall(t, id) {
  const gates = [];
  let gateNext = false;
  class HoldInviter {
    constructor() {
      this.id = id;
      this.state = globalThis.SIP.SessionState.Establishing;
      this.stateChange = {
        listeners: [],
        addListener(fn) {
          this.listeners.push(fn);
        },
      };
    }
    invite(opts) {
      if (opts?.requestDelegate) {
        if (!gateNext) {
          opts.requestDelegate.onAccept?.();
          return Promise.resolve();
        }
        return new Promise((resolve, reject) =>
          gates.push({
            resolve,
            reject,
            accept: () => opts.requestDelegate.onAccept?.(),
            refuse: () => opts.requestDelegate.onReject(),
          }),
        );
      }
      return Promise.resolve();
    }
  }
  globalThis.SIP.Inviter = HoldInviter;
  globalThis.SIP.UserAgent = { makeURI: (raw) => ({ toString: () => raw }) };
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
  const { placeCall } = await import("../island/app/calls.js");
  const { sessions, state } = await import("../island/app/state.js");
  state.userAgent = {};
  assert.equal(await placeCall("1004"), true);
  const entry = sessions.get(id);
  entry.session.state = globalThis.SIP.SessionState.Established;
  entry.session.stateChange.listeners.forEach((fn) => fn("Established"));
  return {
    entry,
    gate: async () => {
      gateNext = true;
      entry.dom.querySelector(".hold-btn").listeners.click[0]();
      await settle();
      return gates.pop();
    },
  };
}

test("a timed-out re-INVITE keeps the settled state and names the verdict", async (t) => {
  t.mock.timers.enable({ apis: ["setTimeout"] });
  const { entry, gate } = await establishedCall(t, "s-timeout");
  const toggle = gate(); // gated: the re-INVITE never answers

  assert.equal(entry.dom.dataset.state, "holding", "pending work shows");
  t.mock.timers.tick(30_000); // the verdict lane's cap fires
  await toggle;
  await settle();

  assert.equal(
    entry.dom.dataset.state,
    "established",
    "the un-held truth restored (the timed-out toggle was a HOLD)",
  );
  const toasts = [...doc.getElementById("toasts").children].map((el) => el.textContent);
  assert.ok(
    toasts.some((txt) => /hold failed/i.test(txt)),
    "hold direction named",
  );
  assert.ok(
    toasts.some((txt) => /timeout/i.test(txt)),
    "verdict named",
  );
  assert.equal(entry.dom.querySelector(".hold-btn").disabled, false);

  entry.session.state = globalThis.SIP.SessionState.Terminated;
  entry.session.stateChange.listeners.forEach((fn) => fn("Terminated"));
});

test("a refused re-INVITE names the direction and the rejected verdict", async (t) => {
  const { entry, gate } = await establishedCall(t, "s-rejected");
  const handle = await gate();
  handle.refuse();
  await settle();

  assert.equal(entry.dom.dataset.state, "established", "refused HOLD restores the un-held truth");
  // The settled re-render re-announces "connected" AFTER the failure
  // toast — the failure copy sits somewhere in the stack, not on top
  // (the same ordering calls.test.mjs documents).
  const toasts = [...doc.getElementById("toasts").children].map((el) => el.textContent);
  assert.ok(
    toasts.some((txt) => /hold failed/i.test(txt)),
    "hold direction named",
  );
  assert.ok(
    toasts.some((txt) => /rejected/i.test(txt)),
    "verdict named",
  );
  assert.equal(entry.held, false);

  entry.session.state = globalThis.SIP.SessionState.Terminated;
  entry.session.stateChange.listeners.forEach((fn) => fn("Terminated"));
});

test("the three failure verdicts render in both languages", async () => {
  const { I18N } = await import("../island/app/i18n.js");
  const verdicts = ["rejected", "timeout", "send-failed"];
  for (const lang of ["en", "de"]) {
    for (const key of ["holdFailed", "resumeFailed"]) {
      const compose = I18N[lang][key];
      assert.equal(typeof compose, "function", `${lang}.${key} composes`);
      for (const verdict of verdicts) {
        const sentence = compose(verdict);
        assert.ok(
          sentence.includes(verdict),
          `${lang}.${key}(${verdict}) names the verdict verbatim: ${sentence}`,
        );
      }
    }
  }
});
