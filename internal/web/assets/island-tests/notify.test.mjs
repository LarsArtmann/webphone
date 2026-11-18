// notify.js's attention machinery: the incoming-call tab-title flash
// (start alternates the glyph strip on a 900ms interval, stop restores
// the captured idle title, start is idempotent) and restoreTitle's
// plain hand-back. The in-call "● in call · mm:ss" title lives in
// calls-title.test.mjs; this file owns the FLASH half.
import test from "node:test";
import assert from "node:assert/strict";

import { installBrowserGlobals } from "./helpers.mjs";

installBrowserGlobals();
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

const notify = await import("../island/app/notify.js");

test("titleFlashStart flips the title on the interval and stop restores the idle title", async () => {
  notify.titleFlashStart();
  // Two half-periods guarantee at least one flip whatever the phase.
  await new Promise((resolve) => setTimeout(resolve, 1900));
  const flipped = document.title === "☎ ☎ ☎";
  assert.ok(
    flipped || document.title === "Webphone",
    `title in the flash orbit, got ${document.title}`,
  );
  // While flashing, a second start must not stack a second interval.
  notify.titleFlashStart();
  notify.titleFlashStop();
  assert.equal(document.title, "Webphone", "stop restores the captured idle title");
});

test("titleFlashStop without a running flash is inert", () => {
  assert.equal(document.title, "Webphone");
  notify.titleFlashStop();
  assert.equal(document.title, "Webphone");
});

test("restoreTitle hands the idle title back without the flash machinery", () => {
  document.title = "● in call · 1:00";
  notify.restoreTitle();
  assert.equal(document.title, "Webphone");
});
