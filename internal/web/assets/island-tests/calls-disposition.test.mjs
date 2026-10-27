// dispositionRow under node:test — the ended card's operator-confirm
// chips (issue #6): the chips POST the verdict with the call facts, a
// success collapses the row to the saved label, a failure re-enables
// the chips for the retry, and the row exists only when the CRM seam
// is on (boot-level config, own file per the panels-crm precedent).
import test from "node:test";
import assert from "node:assert/strict";

import { installBrowserGlobals } from "./helpers.mjs";

const doc = installBrowserGlobals();
globalThis.window.PBX_CONFIG = { crm: true };
globalThis.SIP = {
  SessionState: {
    Establishing: "Establishing",
    Established: "Established",
    Terminating: "Terminating",
    Terminated: "Terminated",
  },
  Inviter: class {},
};

let dispositionPosts = [];
globalThis.fetch = async (input, init) => {
  if (String(input).includes("/disposition")) {
    dispositionPosts.push({ url: String(input), body: init?.body ? JSON.parse(init.body) : null });
    return { ok: true, status: 204, headers: new Headers() };
  }
  return { ok: true, status: 204, headers: new Headers() };
};

const { dispositionRow } = await import("../island/app/calls.js?test=disposition");
const { t } = await import("../island/app/i18n.js?test=disposition");

const endedEntry = () => ({
  session: new globalThis.SIP.Inviter(),
  target: "+493012345678",
  startedAt: Date.now() - 42_000,
});

test("the confirm row carries the ask and the three verdict chips", () => {
  const row = dispositionRow("call-x1", endedEntry());
  assert.equal(row.className, "disposition-row");
  const chips = row.querySelectorAll(".disposition-chip");
  assert.equal(chips.length, 3, "answered, voicemail, ivr — the human-verdictable set");
  assert.deepEqual(
    chips.map((c) => c.dataset.disposition),
    ["answered", "voicemail", "ivr"],
  );
  assert.match(row.querySelector(".disposition-ask").textContent, /How did this call end/);
});

test("clicking a chip POSTs the verdict with the call facts and collapses the row", async () => {
  dispositionPosts = [];
  const row = dispositionRow("call-x2", endedEntry());
  const answered = row.querySelectorAll(".disposition-chip")[0];
  answered.click();
  await new Promise((resolve) => setTimeout(resolve, 0));

  assert.equal(dispositionPosts.length, 1, "one verdict POST");
  assert.match(dispositionPosts[0].url, /\/api\/calls\/call-x2\/disposition$/);
  assert.deepEqual(dispositionPosts[0].body, {
    disposition: "answered",
    number: "+493012345678",
    direction: "out",
    seconds: 42,
  });
  // The row collapses to the saved label — no chips remain. (The stub
  // does not derive a parent's textContent from element children, so
  // the label is read off the replacement span itself.)
  assert.equal(row.querySelectorAll(".disposition-chip").length, 0);
  assert.match(row.querySelector(".disposition-ask").textContent, /Verdict saved/);
});

test("a failed POST keeps the chips clickable for the retry", async () => {
  dispositionPosts = [];
  const previousFetch = globalThis.fetch;
  globalThis.fetch = async (input) => {
    if (String(input).includes("/disposition")) {
      return { ok: false, status: 502, headers: new Headers() };
    }
    return { ok: true, status: 204, headers: new Headers() };
  };
  const row = dispositionRow("call-x3", endedEntry());
  const chip = row.querySelectorAll(".disposition-chip")[1];
  chip.click();
  await new Promise((resolve) => setTimeout(resolve, 0));

  assert.equal(dispositionPosts.length, 1);
  assert.equal(chip.disabled, false, "the retry stays open after a failure");
  globalThis.fetch = previousFetch;
});
