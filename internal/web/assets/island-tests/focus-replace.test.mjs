// focusAfterReplace under node:test — the island keeps keyboard focus
// with a surface an island transition replaces (F8): the note editor's
// save collapses the row and focus lands on the saved line; a
// transition nobody was focused in stays a no-op, never stealing focus.
import test from "node:test";
import assert from "node:assert/strict";

import { installBrowserGlobals } from "./helpers.mjs";

const doc = installBrowserGlobals();
globalThis.window.PBX_CONFIG = {};
globalThis.SIP = {
  SessionState: {
    Establishing: "Establishing",
    Established: "Established",
    Terminating: "Terminating",
    Terminated: "Terminated",
  },
  Inviter: class {},
};

globalThis.fetch = async () => ({ ok: true, status: 204, headers: new Headers() });

const { noteRow } = await import("../island/app/postcall.js?test=focus");
const { focusAfterReplace } = await import("../island/app/ui.js?test=focus");

test("focus moves to the replacement when the departing subtree held it", () => {
  const container = doc.createElement();
  container.className = "wrap";
  const button = doc.createElement();
  button.className = "old";
  container.append(button);
  const next = doc.createElement();
  next.className = "next";
  doc.activeElement = button;

  focusAfterReplace(
    container,
    () => next,
    () => container.replaceChildren(next),
  );

  assert.equal(next.focused, true, "focus followed the replaced surface");
  doc.activeElement = null;
});

test("a transition nobody was focused in disturbs nobody", () => {
  const container = doc.createElement();
  container.className = "wrap2";
  const next = doc.createElement();
  next.className = "next2";
  doc.activeElement = doc.body;

  focusAfterReplace(
    container,
    () => next,
    () => container.replaceChildren(next),
  );

  assert.notEqual(next.focused, true, "no focus steal from an unrelated surface");
  doc.activeElement = null;
});

test("saving a call note keeps focus on the saved line", async () => {
  const row = noteRow("call-f1");
  const save = row.querySelector(".call-note-save");
  doc.activeElement = save;
  save.click();
  await new Promise((resolve) => setTimeout(resolve, 0));

  const saved = row.querySelector(".call-note-saved");
  assert.ok(saved, "the row collapsed to the saved line");
  assert.equal(saved.focused, true, "focus landed on the saved line");
  doc.activeElement = null;
});
