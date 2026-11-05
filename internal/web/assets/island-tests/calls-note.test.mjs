// noteRow under node:test — the ended card's post-call note editor
// (issue #11): the hangup moment PUTs the same slot History renders,
// a 204 collapses the row to the saved line, and the 409/failure paths
// keep the draft in the editor (hand-written content is never silently
// expunged). Own file per the calls-disposition precedent.
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

let notePuts = [];
globalThis.fetch = async (input, init) => {
  if (String(input).includes("/note")) {
    notePuts.push({ url: String(input), body: init?.body ? JSON.parse(init.body) : null });
    return { ok: true, status: 204, headers: new Headers() };
  }
  return { ok: true, status: 204, headers: new Headers() };
};

const { noteRow } = await import("../island/app/calls.js?test=note");

test("the editor carries the placeholder and the save affordance", () => {
  const row = noteRow("call-n1");
  assert.equal(row.className, "call-note-row");
  const input = row.querySelector(".call-note-input");
  assert.ok(input, "the draft textarea exists");
  assert.match(input.placeholder, /Add a note for this call/);
  assert.equal(input.maxLength, 4000, "the server's note budget, enforced client-side");
  const save = row.querySelector(".call-note-save");
  assert.ok(save, "the save button exists");
  assert.match(save.textContent, /Save note/);
});

test("saving PUTs the draft to the note slot and collapses the row", async () => {
  notePuts = [];
  const row = noteRow("call-n2");
  const input = row.querySelector(".call-note-input");
  input.value = "called about the invoice; call back Thursday";
  row.querySelector(".call-note-save").click();
  await new Promise((resolve) => setTimeout(resolve, 0));

  assert.equal(notePuts.length, 1, "one note PUT");
  assert.match(notePuts[0].url, /\/api\/calls\/call-n2\/note$/);
  assert.deepEqual(notePuts[0].body, { body: "called about the invoice; call back Thursday" });
  assert.equal(row.querySelector(".call-note-input"), null, "the editor collapses on success");
  assert.equal(row.querySelector(".call-note-save"), null);
  const saved = row.querySelector(".call-note-saved");
  assert.ok(saved, "the saved line replaces the editor");
  assert.match(saved.textContent, /Note saved/);
});

test("a failed save keeps the draft and the retry affordance", async () => {
  const previousFetch = globalThis.fetch;
  for (const status of [409, 502]) {
    notePuts = [];
    globalThis.fetch = async (input) => {
      if (String(input).includes("/note")) {
        notePuts.push({ url: String(input), body: null });
        return { ok: false, status, headers: new Headers() };
      }
      return { ok: true, status: 204, headers: new Headers() };
    };
    const row = noteRow(`call-n-${status}`);
    const input = row.querySelector(".call-note-input");
    input.value = "draft survives the failure";
    row.querySelector(".call-note-save").click();
    await new Promise((resolve) => setTimeout(resolve, 0));

    assert.equal(notePuts.length, 1);
    const keptInput = row.querySelector(".call-note-input");
    assert.ok(keptInput, `status ${status}: the draft stays editable`);
    assert.equal(
      keptInput.value,
      "draft survives the failure",
      `status ${status}: the draft text survives`,
    );
    assert.equal(
      row.querySelector(".call-note-save").disabled,
      false,
      `status ${status}: the retry stays open`,
    );
  }
  globalThis.fetch = previousFetch;
});
