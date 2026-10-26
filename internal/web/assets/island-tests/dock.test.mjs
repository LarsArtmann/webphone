// The call dock (multi-call train): N>1 collapses the cards into a
// chip strip plus ONE expanded focused card; the chip click is the
// hold-&-take gesture through the existing hold pending machine; per-
// call volume rides the shared remote-audio element; [ ] cycles
// established calls only.
import test from "node:test";
import assert from "node:assert/strict";

import { installBrowserGlobals } from "./helpers.mjs";

const doc = installBrowserGlobals();
globalThis.Element = class Element {};
globalThis.SIP = {
  SessionState: {
    Establishing: "Establishing",
    Established: "Established",
    Terminating: "Terminating",
    Terminated: "Terminated",
  },
  Inviter: class {},
};

const { sessions, state } = await import("../island/app/state.js");
const { bindSession, cycleFocus, focusSession, renderCalls } = await import(
  "../island/app/calls.js"
);
const { t } = await import("../island/app/i18n.js");

const makeCard = () => ({
  dataset: {},
  hidden: false,
  className: "",
  classList: {
    toggle() {},
  },
  querySelector(selector) {
    if (selector === ".call-state-text") return stateText;
    if (selector === ".volume-slider") return this.slider;
    return { textContent: "" };
  },
  slider: null,
});
const stateText = { textContent: "" };

let seq = 0;
function addEstablished(target, { held = false } = {}) {
  const id = `call-${++seq}`;
  const entry = {
    session: {
      state: globalThis.SIP.SessionState.Established,
      invite: () => Promise.resolve(),
    },
    target,
    held,
    holdPending: null,
    holdQueued: undefined,
    muted: false,
    startedAt: Date.now(),
    dom: makeCard(),
  };
  sessions.set(id, entry);
  return id;
}

// A REAL card (the addCallCard builder) for tests that drive the actual
// wiring: sliders, buttons, transcript nodes.
function addRealEstablished(target) {
  const fake = {
    id: `real-${++seq}`,
    state: globalThis.SIP.SessionState.Established,
    invite: () => Promise.resolve(),
    stateChange: { addListener() {} },
  };
  bindSession(fake, target);
  return fake.id;
}

const reset = () => {
  sessions.clear();
  state.focusedId = null;
  renderCalls();
  doc.getElementById("calls").replaceChildren();
};

const dockEl = () => {
  const calls = doc.getElementById("calls");
  return calls.children.find((kid) => kid.className === "wp-call-dock") || null;
};

const lastToast = () => doc.getElementById("toasts").children.at(-1);

test("N=2 renders two chips plus one expanded focused card", () => {
  reset();
  const a = addEstablished("+493011111111");
  const b = addEstablished("+493022222222", { held: true });
  state.focusedId = a;
  renderCalls();

  const dock = dockEl();
  assert.ok(dock, "the dock strip exists with two live calls");
  assert.equal(dock.children.length, 2, "one chip per live call");
  assert.equal(dock.children[0].id, `wp-dock-chip-${a}`);
  assert.equal(dock.children[1].id, `wp-dock-chip-${b}`);
  assert.match(dock.children[0].children[1].textContent, /\+493011111111/);
  assert.equal(
    sessions.get(a).dom.hidden,
    false,
    "the focused call keeps its expanded card",
  );
  assert.equal(
    sessions.get(b).dom.hidden,
    true,
    "the other call collapses into its chip",
  );
  assert.match(dock.children[1].children[2].textContent, new RegExp(t("onHold").replace(/[.*+?^${}()|[\]\\]/g, "\\$&")));
});

test("a single call renders no dock (plain card stays)", () => {
  reset();
  const a = addEstablished("+493011111111");
  state.focusedId = a;
  renderCalls();
  assert.equal(dockEl(), null);
  assert.equal(sessions.get(a).dom.hidden, false);
});

test("the chip click is hold & take through the hold pending machine", async () => {
  reset();
  const a = addEstablished("+493011111111");
  const b = addEstablished("+493022222222", { held: true });
  state.focusedId = a;
  renderCalls();

  // Mid-flight hold on the TARGET: the swap must queue, not lie.
  sessions.get(b).holdPending = "holding";
  const dockRow = dockEl();
  const chip = dockRow.children.find((kid) => kid.id === `wp-dock-chip-${b}`);
  chip.click();
  assert.equal(
    sessions.get(b).holdQueued,
    false,
    "the resume wish queues behind the pending hold",
  );
  sessions.get(b).holdPending = null;

  // A settled chip take: focus moves, the other call is held.
  dockRow.children.find((kid) => kid.id === `wp-dock-chip-${a}`).click();
  assert.equal(state.focusedId, a);
  renderCalls();
  assert.equal(sessions.get(a).dom.hidden, false);
  assert.equal(sessions.get(b).dom.hidden, true);
});

test("the chip pill mirrors the live state line", () => {
  reset();
  const a = addEstablished("+493011111111");
  const b = addEstablished("+493022222222", { held: true });
  state.focusedId = a;
  renderCalls();
  const pill = dockEl().children.find((kid) => kid.id === `wp-dock-chip-${b}`).children[2];
  assert.match(pill.textContent, new RegExp(t("onHold")));
  sessions.get(b).held = false;
  renderCalls();
  assert.match(pill.textContent, new RegExp(t("inCall")));
});

test("per-call volume rides the shared audio element and persists", () => {
  reset();
  const a = addRealEstablished("+493011111111");
  const b = addEstablished("+493022222222");
  state.focusedId = a;
  renderCalls();
  const audio = doc.getElementById("remote-audio");
  audio.volume = 1;

  // The slider the REAL card builder made (bindSession path) carries
  // the input listener; drive it like the browser would.
  const sliderA = sessions.get(a).dom.querySelector(".volume-slider");
  assert.ok(sliderA && sliderA.listeners.input, "the card wires the slider");
  sliderA.value = "40";
  sliderA.listeners.input[0]({});
  assert.equal(Math.round(audio.volume * 100), 40, "the slider drives playback");
  assert.equal(sessions.get(a).volume, 0.4);

  renderCalls();
  assert.equal(
    sessions.get(a).dom.querySelector(".volume-slider").value,
    "40",
    "the remembered level survives re-renders",
  );
});

test("[ ] cycles established calls and announces the party; rings are skipped", () => {
  reset();
  const a = addEstablished("+493011111111");
  const b = addEstablished("+493022222222");
  const ringing = {
    session: { state: globalThis.SIP.SessionState.Establishing },
    target: "+493033333333",
    held: false,
    holdPending: null,
    muted: false,
    startedAt: Date.now(),
    dom: makeCard(),
  };
  sessions.set("call-ring", ringing);
  state.focusedId = a;
  renderCalls();

  cycleFocus(1);
  assert.equal(state.focusedId, b, "forward lands on the other ESTABLISHED call");
  assert.match(lastToast().textContent, /\+493022222222/);
  assert.match(lastToast().textContent, new RegExp(t("focusChanged")("+493022222222").replace(/[.*+?^${}()|[\]\\]/g, "\\$&")));

  cycleFocus(-1);
  assert.equal(state.focusedId, a, "back lands on the first call, never the ring");

  cycleFocus(1);
  cycleFocus(1);
  assert.equal(state.focusedId, a, "the cycle wraps across two live calls");
});
