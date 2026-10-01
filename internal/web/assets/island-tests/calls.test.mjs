// calls.js under node:test — the call-state chip + its aria-live
// announcements (plan T20a): the card carries data-state for the CSS
// chip, and only state TRANSITIONS are announced (the once-per-second
// duration tick must stay silent).
import test from "node:test";
import assert from "node:assert/strict";

import { installBrowserGlobals } from "./helpers.mjs";

const doc = installBrowserGlobals();
globalThis.SIP = {
  SessionState: {
    Establishing: "Establishing",
    Established: "Established",
    Terminating: "Terminating",
    Terminated: "Terminated",
  },
  Inviter: class {},
};

const { sessions } = await import("../island/app/state.js");
const { initDialWarm, renderCalls } = await import("../island/app/calls.js");
const { t } = await import("../island/app/i18n.js");

const stateText = { textContent: "" };
const makeCard = () => ({
  dataset: {},
  classList: { toggle() {} },
  querySelector(selector) {
    if (selector === ".call-state-text") return stateText;
    return { textContent: "" };
  },
});

const lastToast = () => doc.getElementById("toasts").children.at(-1);

test("the call card's data-state drives the chip and only transitions announce", () => {
  const card = makeCard();
  const entry = {
    session: { state: globalThis.SIP.SessionState.Establishing },
    dom: card,
    target: "+493012345678",
    held: false,
    muted: false,
    startedAt: Date.now(),
  };
  sessions.set("c1", entry);

  renderCalls();
  assert.equal(card.dataset.state, "ringing");
  assert.match(lastToast().textContent, /ringing/i);

  entry.session.state = globalThis.SIP.SessionState.Established;
  renderCalls();
  assert.equal(card.dataset.state, "established");
  assert.match(lastToast().textContent, /connected/i);

  // The per-second duration re-render must NOT re-announce.
  const toastsBefore = doc.getElementById("toasts").children.length;
  renderCalls();
  assert.equal(doc.getElementById("toasts").children.length, toastsBefore);

  entry.session.state = globalThis.SIP.SessionState.Terminated;
  renderCalls();
  assert.equal(card.dataset.state, "ending");
  assert.match(lastToast().textContent, /ended/i);

  sessions.delete("c1");
});

test("the announcement copy exists in both languages", () => {
  assert.match(t("callRinging")("+49"), /ringing/i);
  assert.match(t("callEstablished")("+49"), /connected/i);
  assert.match(t("callEnded")("+49"), /ended/i);
});

test("placeCall reports whether the INVITE went out (dest-clear contract)", async () => {
  const invited = [];
  globalThis.SIP.Inviter = class {
    constructor(_ua, uri) {
      this.uri = uri;
      this.id = "inv-" + invited.length;
      this.state = globalThis.SIP.SessionState.Establishing;
      this.stateChange = { addListener() {} };
    }
    async invite() {
      invited.push(this.uri.toString());
    }
  };
  globalThis.SIP.UserAgent = { makeURI: (raw) => ({ toString: () => raw }) };
  // bindSession starts the ringback for an outgoing Inviter — a silent
  // audio-context stub keeps node:test quiet.
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
  const { state } = await import("../island/app/state.js");

  state.userAgent = {};
  assert.equal(await placeCall(""), false, "empty input must not dial");
  assert.equal(await placeCall("  –  …  "), false, "nothing dialable must not dial");
  assert.equal(await placeCall("+4930-123456"), true, "real dial reports true");
  assert.equal(invited.length, 1, "exactly one INVITE");
  assert.match(invited[0], /sip:\+4930123456@/, "sanitized target in the URI");

  state.userAgent = null;
  assert.equal(await placeCall("1001"), false, "no agent must not dial");
  state.userAgent = {};
});

// The honesty contract: between the click and the re-INVITE's verdict the
// card shows the WORK ("holding…" / "resuming…", pulsing chip, disabled
// button) — never the hoped-for end state. A failed re-INVITE returns the
// card to the SETTLED state it actually still has.
test("hold UI shows the pending truth until the re-INVITE settles", async () => {
  const gates = [];
  let holdMode = "resolve"; // how the NEXT invite() settles
  class HoldInviter {
    constructor() {
      this.id = "s-hold";
      this.state = globalThis.SIP.SessionState.Establishing;
      this.stateChange = {
        listeners: [],
        addListener(fn) {
          this.listeners.push(fn);
        },
      };
    }
    invite() {
      if (holdMode === "resolve") return Promise.resolve();
      return new Promise((resolve, reject) => gates.push({ resolve, reject }));
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
  assert.equal(await placeCall("1002"), true);
  const entry = sessions.get("s-hold");
  entry.session.state = globalThis.SIP.SessionState.Established;
  entry.session.stateChange.listeners.forEach((fn) => fn("Established"));
  const card = entry.dom;
  const holdBtn = card.querySelector(".hold-btn");
  const chipText = () => card.querySelector(".call-state-text").textContent;
  const settle = async () => {
    for (let i = 0; i < 12; i++) await new Promise((r) => setImmediate(r));
  };

  assert.equal(card.dataset.state, "established");
  assert.equal(holdBtn.textContent, "Hold");
  assert.equal(holdBtn.disabled, false);

  holdMode = "gate";
  holdBtn.listeners.click[0]();
  await settle();
  assert.equal(card.dataset.state, "holding", "pending hold must show WORK");
  assert.match(chipText(), /holding/i);
  assert.equal(holdBtn.textContent, t("callHolding"));
  assert.equal(holdBtn.disabled, true, "no second toggle while in flight");

  gates.pop().resolve();
  await settle();
  assert.equal(card.dataset.state, "held", "settled: actually on hold");
  assert.match(chipText(), /on hold/i);
  assert.equal(holdBtn.textContent, t("resume"));
  assert.equal(holdBtn.disabled, false);

  holdBtn.listeners.click[0]();
  await settle();
  assert.equal(card.dataset.state, "resuming", "pending resume must show WORK");
  assert.match(chipText(), /resuming/i);
  assert.equal(holdBtn.textContent, t("callResuming"));
  assert.equal(holdBtn.disabled, true);

  gates.pop().reject(new Error("boom"));
  await settle();
  assert.equal(card.dataset.state, "held", "failure keeps the settled truth");
  assert.match(chipText(), /on hold/i);
  assert.equal(holdBtn.disabled, false);
  // The failed toggle above was a RESUME: the copy must name the
  // direction (island-honesty follow-up — split holdFailed/resumeFailed).
  assert.match(lastToast().textContent, /resume failed/i);

  // Terminated while a toggle is still in flight: the pending state
  // must not outlive the session (the re-INVITE answer never arrives;
  // the watchdog-rebuild teardown lands here too). The session is still
  // HELD, so this toggle is a resume — either direction must settle.
  holdMode = "gate";
  holdBtn.listeners.click[0]();
  await settle();
  assert.equal(entry.holdPending, "resuming", "the toggle is in flight");
  entry.session.state = globalThis.SIP.SessionState.Terminated;
  entry.session.stateChange.listeners.forEach((fn) => fn("Terminated"));
  assert.equal(entry.holdPending, null, "Terminated settles the pending toggle");
  assert.equal(entry.holdQueued, undefined);
  assert.equal(sessions.has("s-hold"), false, "teardown cleans the timer");
});

test("a failed hold announces the hold direction", async () => {
  const gates = [];
  let holdMode = "resolve"; // the INITIAL invite must settle; only the toggle gates
  class HoldInviter {
    constructor() {
      this.id = "s-hold2";
      this.state = globalThis.SIP.SessionState.Establishing;
      this.stateChange = {
        listeners: [],
        addListener(fn) {
          this.listeners.push(fn);
        },
      };
    }
    invite() {
      if (holdMode === "resolve") return Promise.resolve();
      return new Promise((resolve, reject) => gates.push({ resolve, reject }));
    }
  }
  globalThis.SIP.Inviter = HoldInviter;
  globalThis.SIP.UserAgent = { makeURI: (raw) => ({ toString: () => raw }) };
  const { placeCall } = await import("../island/app/calls.js");
  const { sessions, state } = await import("../island/app/state.js");

  state.userAgent = {};
  assert.equal(await placeCall("1003"), true);
  const entry = sessions.get("s-hold2");
  entry.session.state = globalThis.SIP.SessionState.Established;
  entry.session.stateChange.listeners.forEach((fn) => fn("Established"));
  const settle = async () => {
    for (let i = 0; i < 12; i++) await new Promise((r) => setImmediate(r));
  };

  holdMode = "gate";
  entry.dom.querySelector(".hold-btn").listeners.click[0]();
  await settle();
  gates.pop().reject(new Error("nope"));
  await settle();
  // The failed HOLD re-renders the settled state, which re-announces
  // "connected" — so the hold-failed toast is somewhere in the stack,
  // not necessarily the last one.
  const toasts = doc.getElementById("toasts").children;
  assert.ok(
    toasts.some((el) => /hold failed/i.test(el.textContent)),
    "the hold direction is named in the failure copy",
  );
  assert.equal(entry.held, false, "the settled truth stays un-held");

  entry.session.state = globalThis.SIP.SessionState.Terminated;
  entry.session.stateChange.listeners.forEach((fn) => fn("Terminated"));
  state.userAgent = null;
});

test("dial focus starts a bounded mic warm (the outgoing pre-warm)", async () => {
  const acquisitions = [];
  Object.defineProperty(globalThis, "navigator", {
    value: {
      mediaDevices: {
        getUserMedia(constraints) {
          acquisitions.push(constraints);
          return new Promise(() => {});
        },
      },
    },
    configurable: true,
  });

  initDialWarm();
  const dest = doc.getElementById("dest");
  dest.listeners.focus[0]();
  await new Promise((resolve) => setImmediate(resolve));

  assert.equal(acquisitions.length, 1, "one acquisition at dial focus");
  assert.deepEqual(
    acquisitions[0],
    { audio: true, video: false },
    "the warm asks for the mic only",
  );

  // Refocusing while the acquisition is pending must not stack requests.
  dest.listeners.focus[0]();
  await new Promise((resolve) => setImmediate(resolve));
  assert.equal(acquisitions.length, 1, "idempotent while pending");
});

test("sanitizeDialable keeps the dialable charset and strips formatting", () => {
  assert.equal(sanitizeDialable("+49 (30) 1234-56"), "+493012345 6".replace(" ", ""));
  assert.equal(sanitizeDialable("\u202D1003\u202C"), "1003", "direction marks are stripped");
  assert.equal(sanitizeDialable("*#abcABC"), "*#abcABC");
  assert.equal(sanitizeDialable("   "), "");
});

test("the dial hint previews the dialable form while typing", async () => {
  initDialHint();
  const dest = doc.getElementById("dest");
  const hint = doc.getElementById("wp-dial-hint");
  const type = (value) => {
    dest.value = value;
    dest.listeners.input[0]();
  };

  type("+49 (30) 1234-56");
  assert.equal(hint.hidden, false, "a formatted number reveals the hint");
  assert.equal(hint.textContent, "→ +493012345 6".replace(" ", ""), "the hint shows the dialable form");

  type("1003");
  assert.equal(hint.hidden, true, "an already-dialable number stays silent");
  assert.equal(hint.textContent, "");

  type("+49 30 \u202D1234\u202C");
  assert.equal(hint.hidden, false, "invisible direction marks still trigger the preview");
});
