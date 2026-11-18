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

const { noteRow, flushPendingCallNotes, noteAutosave } = await import("../island/app/postcall.js?test=note");

// fireInput drives the element stub's input listeners the way the
// browser would (the stub has no dispatchEvent on elements).
/** @param {any} el */
const fireInput = (el) => {
  for (const fn of el.listeners.input ?? []) fn({ target: el });
};
// withFastDebounce shrinks the module's tuning seam for one test body.
/** @param {() => Promise<void>} fn */
const withFastDebounce = async (fn) => {
  const real = noteAutosave.delayMs;
  noteAutosave.delayMs = 5;
  try {
    await fn();
  } finally {
    noteAutosave.delayMs = real;
  }
};
const settleAutosave = () => new Promise((resolve) => setTimeout(resolve, 25));

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

test("typing autosaves after the debounce without collapsing the editor", async () => {
  await withFastDebounce(async () => {
    notePuts = [];
    const row = noteRow("call-a1");
    const input = row.querySelector(".call-note-input");
    input.value = "rueckruf am freitag";
    fireInput(input);
    await settleAutosave();

    assert.equal(notePuts.length, 1, "one debounced PUT");
    assert.deepEqual(notePuts[0].body, { body: "rueckruf am freitag" });
    assert.match(notePuts[0].url, /\/api\/calls\/call-a1\/note$/);
    assert.ok(row.querySelector(".call-note-input"), "the editor stays open");
    assert.match(
      row.querySelector(".call-note-status").textContent,
      /Saved/,
      "the quiet status line carries the save",
    );

    input.value = "rueckruf am freitag, nachmittag";
    fireInput(input);
    await settleAutosave();
    assert.equal(notePuts.length, 2, "the next pause saves again");
    assert.deepEqual(notePuts[1].body, { body: "rueckruf am freitag, nachmittag" });
  });
});

test("a typing burst coalesces into ONE save carrying the final body", async () => {
  await withFastDebounce(async () => {
    notePuts = [];
    const row = noteRow("call-a2");
    const input = row.querySelector(".call-note-input");
    input.value = "a";
    fireInput(input);
    input.value = "ab";
    fireInput(input);
    input.value = "abc";
    fireInput(input);
    await settleAutosave();

    assert.equal(notePuts.length, 1, "the debounce coalesces the burst");
    assert.deepEqual(notePuts[0].body, { body: "abc" });
  });
});

test("a failed autosave keeps the draft, marks the line, announces once per episode", async () => {
  await withFastDebounce(async () => {
    const previousFetch = globalThis.fetch;
    let answerStatus = 502;
    globalThis.fetch = async (inputUrl, init) => {
      if (String(inputUrl).includes("/note")) {
        notePuts.push({ url: String(inputUrl), body: init?.body ? JSON.parse(init.body) : null });
        return { ok: false, status: answerStatus, headers: new Headers() };
      }
      return { ok: true, status: 204, headers: new Headers() };
    };
    try {
      const toasts = doc.getElementById("toasts");
      notePuts = [];
      const row = noteRow("call-a3");
      const input = row.querySelector(".call-note-input");
      input.value = "draft one";
      fireInput(input);
      await settleAutosave();
      assert.equal(notePuts.length, 1);
      assert.match(row.querySelector(".call-note-status").textContent, /Not saved/);
      const announced = toasts.children.length;
      assert.ok(announced >= 1, "the first failure announces");

      input.value = "draft two";
      fireInput(input);
      await settleAutosave();
      assert.equal(notePuts.length, 2, "the next pause retries");
      assert.equal(
        toasts.children.length,
        announced,
        "the same episode does not re-announce",
      );

      answerStatus = 204;
      input.value = "draft three";
      fireInput(input);
      await settleAutosave();
      assert.equal(notePuts.length, 3);
      assert.match(row.querySelector(".call-note-status").textContent, /Saved/);
    } finally {
      globalThis.fetch = previousFetch;
    }
  });
});

test("flushPendingCallNotes sends the unsaved tail without waiting", async () => {
  await withFastDebounce(async () => {
    const previousFetch = globalThis.fetch;
    const seen = [];
    globalThis.fetch = async (inputUrl, init) => {
      if (String(inputUrl).includes("/note")) {
        seen.push({ body: init?.body ? JSON.parse(init.body) : null, init });
        return { ok: true, status: 204, headers: new Headers() };
      }
      return { ok: true, status: 204, headers: new Headers() };
    };
    try {
      const row = noteRow("call-a4");
      const input = row.querySelector(".call-note-input");
      input.value = "typed right before closing";
      fireInput(input);
      // No settle wait: the flush is the point.
      flushPendingCallNotes();
      await new Promise((resolve) => setTimeout(resolve, 0));
      assert.equal(seen.length, 1, "the tail flew immediately");
      assert.deepEqual(seen[0].body, { body: "typed right before closing" });

      // pagehide rides the same flush with keepalive: the page is going
      // away, the browser must finish the PUT past teardown.
      input.value = "and one more word";
      fireInput(input);
      doc.dispatch("pagehide");
      await new Promise((resolve) => setTimeout(resolve, 0));
      assert.equal(seen.length, 2);
      assert.equal(seen[1].init.keepalive, true, "pagehide keeps the PUT alive");
    } finally {
      globalThis.fetch = previousFetch;
    }
  });
});

test("the done button collapses an already-autosaved body without a second PUT", async () => {
  await withFastDebounce(async () => {
    notePuts = [];
    const row = noteRow("call-a5");
    const input = row.querySelector(".call-note-input");
    input.value = "all typed, all saved";
    fireInput(input);
    await settleAutosave();
    assert.equal(notePuts.length, 1);

    row.querySelector(".call-note-save").click();
    await new Promise((resolve) => setTimeout(resolve, 0));
    assert.equal(notePuts.length, 1, "no second PUT for a settled body");
    assert.equal(row.querySelector(".call-note-input"), null, "the row collapses");
    assert.ok(row.querySelector(".call-note-saved"), "the saved line lands");
  });
});
