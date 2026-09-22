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
const { renderCalls } = await import("../island/app/calls.js");
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
  const { placeCall } = await import("../island/app/calls.js");
  const { state } = await import("../island/app/state.js");

  state.userAgent = {};
  assert.equal(await placeCall(""), false, "empty input must not dial");
  assert.equal(await placeCall("+49 30 x"), false, "nothing dialable must not dial");
  assert.equal(await placeCall("+4930-123456"), true, "real dial reports true");
  assert.equal(invited.length, 1, "exactly one INVITE");
  assert.match(invited[0], /sip:\+4930123456@/, "sanitized target in the URI");

  state.userAgent = null;
  assert.equal(await placeCall("1001"), false, "no agent must not dial");
  state.userAgent = {};
});
