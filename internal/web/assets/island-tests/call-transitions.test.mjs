// D-train (F55/F56): the island's declarative call-lifecycle table —
// CALL_TRANSITIONS + applyTransition + viewStateOf live in state.js so
// every module can read them without cycles. These specs pin the table
// itself and the honesty contract it encodes.
import test from "node:test";
import assert from "node:assert/strict";

import { CALL_TRANSITIONS, applyTransition, viewStateOf } from "../island/app/state.js";

test("the call table allows every lifecycle walk and refuses the lies", () => {
  // The happy paths.
  for (const [from, to] of [
    ["ringing", "active"],
    ["active", "holding"],
    ["holding", "held"],
    ["held", "resuming"],
    ["resuming", "active"],
    ["active", "ending"],
    ["ending", "ended"],
    ["ringing", "ended"], // a dropped call never reaches active
  ]) {
    assert.deepEqual(
      applyTransition(from, to),
      { state: to, changed: true },
      `${from} -> ${to} is a legal move`,
    );
  }

  // The honesty contract's core: a FAILED hold toggle returns to the
  // SETTLED state — both settle outcomes stay legal from in-flight
  // states because where you came from decides where you land.
  assert.equal(applyTransition("holding", "active").changed, true);
  assert.equal(applyTransition("holding", "held").changed, true);
  assert.equal(applyTransition("resuming", "held").changed, true);

  // The lies: asking for what you have, walking backwards, moving on
  // from the grave.
  for (const [from, to] of [
    ["active", "resuming"], // resume while live: nothing to resume
    ["held", "holding"], // hold while held: nothing to hold
    ["ended", "active"], // no resurrection
    ["ending", "active"], // no takebacks mid-teardown
    ["ringing", "held"], // hold before media exists
  ]) {
    assert.deepEqual(
      applyTransition(from, to),
      { state: from, changed: false },
      `${from} -> ${to} must be refused (state kept)`,
    );
  }
});

test("viewStateOf derives the same vocabulary the renderer branches on", () => {
  const session = (state) => ({ state });
  assert.equal(viewStateOf({ session: session("Establishing") }), "ringing");
  assert.equal(viewStateOf({ session: session("Established") }), "active");
  assert.equal(viewStateOf({ session: session("Established"), held: true }), "held");
  assert.equal(viewStateOf({ session: session("Established"), holdPending: "holding" }), "holding");
  assert.equal(
    viewStateOf({ session: session("Established"), holdPending: "resuming" }),
    "resuming",
  );
  assert.equal(viewStateOf({ session: session("Terminating") }), "ending");
  assert.equal(viewStateOf({ session: session("Terminated") }), "ended");
  assert.equal(viewStateOf(null), "ended");
  assert.equal(viewStateOf({}), "ended");
});
