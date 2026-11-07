// The in-call tab title + the beforeunload guard (T19 M18/M17): the 1s
// render tick carries "● in call · mm:ss" into the tab strip so other
// tabs can see the phone is busy, and the browser asks before closing a
// tab with a live call — an idle phone never nags.
import test from "node:test";
import assert from "node:assert/strict";

import { installBrowserGlobals } from "./helpers.mjs";

const doc = installBrowserGlobals();
// notify.js captures the idle title at import — a real page has one.
document.title = "Webphone";
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
const makeDom = () => ({
  dataset: {},
  classList: { toggle() {} },
  querySelector(selector) {
    if (selector === ".call-state-text") return stateText;
    return { textContent: "" };
  },
});

const establishedEntry = (id, ageMs) => {
  const entry = {
    session: new globalThis.SIP.Inviter(),
    target: "+493012345678",
    startedAt: Date.now() - ageMs,
    dom: makeDom(),
  };
  entry.session.state = globalThis.SIP.SessionState.Established;
  sessions.set(id, entry);
  return entry;
};

test("the render tick carries the established call into the tab title", () => {
  establishedEntry("title-1", 65_000);
  renderCalls();
  assert.ok(
    document.title.startsWith(`● ${t("inCall")} · 1:0`),
    `title reads "● in call · mm:ss", got ${document.title}`,
  );
  sessions.clear();
  renderCalls();
  assert.equal(document.title, "Webphone", "the idle title is handed back");
});

test("the oldest established call owns the title's clock", () => {
  establishedEntry("title-old", 125_000);
  establishedEntry("title-new", 5_000);
  renderCalls();
  assert.ok(
    document.title.includes("2:0"),
    `the oldest call's duration ticks, got ${document.title}`,
  );
  sessions.clear();
  renderCalls();
});

test("a live call arms beforeunload; an idle phone never nags", () => {
  const handlers = globalThis.window.listeners["beforeunload"] ?? [];
  assert.ok(handlers.length > 0, "the guard registers at import");
  const fired = { count: 0, returnValue: undefined };
  const event = {
    preventDefault() {
      fired.count += 1;
    },
    set returnValue(v) {
      fired.returnValue = v;
    },
  };

  handlers.forEach((fn) => fn(event));
  assert.equal(fired.count, 0, "no call — closing is silent");

  establishedEntry("unload-1", 1_000);
  handlers.forEach((fn) => fn(event));
  assert.equal(fired.count, 1, "a live call makes the browser ask");
  assert.equal(fired.returnValue, "", "the legacy returnValue channel is set");

  sessions.clear();
});
