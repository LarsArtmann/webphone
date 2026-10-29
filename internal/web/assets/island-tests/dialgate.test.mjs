// The advisory dial gate (issue #7) under node:test: a denied verdict
// renders the inline chip (never a modal), the override spends once
// server-side and re-dials with force, null verdicts (404/unreachable)
// dial without asking twice, and the remaining-window phrasing rides
// the wire's last_at + window.
import test from "node:test";
import assert from "node:assert/strict";

import { installBrowserGlobals } from "./helpers.mjs";

const doc = installBrowserGlobals();
const { t } = await import("../island/app/i18n.js");
await import("../island/app/ui.js");
const dialgate = await import("../island/app/dialgate.js");

// A denied verdict's chip lands after the dial error line.
const dialForm = doc.createElement("form");
dialForm.id = "dial-form";
const dialError = doc.createElement("p");
dialError.id = "dial-error";
dialForm.append(dialError);
doc.body.append(dialForm);
doc.querySelector = (selector) => (selector === "#dial-form" ? dialForm : null);

function useFetch(handler) {
  globalThis.fetch = async (url, opts) => handler(url, opts);
}

test("an unreachable gate dials (null verdict, warn logged)", async () => {
  useFetch(() => {
    throw new Error("network gone");
  });
  assert.equal(await dialgate.consultDialGate("+1715"), null);
});

test("a denied verdict carries the wire's state", async () => {
  useFetch(() => ({
    ok: true,
    status: 200,
    json: async () => ({ allowed: false, last_at: 1_000_000, window: 3600 }),
  }));
  const verdict = await dialgate.consultDialGate("+1715");
  assert.ok(verdict);
  assert.equal(verdict.allowed, false);
  assert.equal(verdict.window, 3600);
});

test("overrideDialGate spends the one-shot and answers honestly", async () => {
  let asked = null;
  useFetch((_url, opts) => {
    asked = JSON.parse(opts.body);
    return { ok: true, status: 200, json: async () => ({ allowed: true }) };
  });
  assert.equal(await dialgate.overrideDialGate("+1715"), true);
  assert.deepEqual(asked, { remote: "+1715", override: true });
  useFetch(() => ({
    ok: true,
    status: 200,
    json: async () => ({ allowed: false }),
  }));
  assert.equal(await dialgate.overrideDialGate("+1715"), false);
  useFetch(() => ({ ok: false, status: 502, json: async () => ({}) }));
  assert.equal(await dialgate.overrideDialGate("+1715"), false);
});

test("agoLabel phrases stamps the draftAge way", () => {
  const now = Date.now() / 1000;
  assert.equal(dialgate.agoLabel(now), "now");
  assert.equal(dialgate.agoLabel(now - 47 * 60), "47m");
  assert.equal(dialgate.agoLabel(now - 2 * 3600), "2h");
  assert.equal(dialgate.agoLabel(now - 26 * 3600), "1d");
});

// The 404 latch is module state, so this test runs LAST: earlier
// tests still exercise the live gate.
test("404 pins the seam off: later asks never round-trip", async () => {
  let calls = 0;
  useFetch(() => {
    calls += 1;
    return { ok: false, status: 404, json: async () => ({}) };
  });
  assert.equal(await dialgate.consultDialGate("+1715"), null);
  assert.equal(await dialgate.consultDialGate("+1715"), null);
  assert.equal(calls, 1, "the 404 is learned, not re-asked");
});

// The chip render lives in calls.js (placeCall's flow); its phrasing is
// pinned here so both languages keep the state line honest.
test("the cooldown notice names the wait", () => {
  assert.equal(t("cooldownNotice")("47m", "13m"), "Last call 47m ago — redial paused (13m left).");
  assert.equal(t("dialAnyway"), "Dial anyway");
});

test("the slots cache feeds cooldownFor with normalized matching", async () => {
  const now = Math.floor(Date.now() / 1000);
  let slotsFetched = 0;
  useFetch((url) => {
    if (url === "/api/dials") {
      slotsFetched += 1;
      return {
        ok: true,
        status: 200,
        json: async () => ({
          slots: [
            { remote: "+4989123456", last_at: now - 600, window: 3600 },
          ],
        }),
      };
    }
    throw new Error("unexpected fetch " + url);
  });
  await dialgate.refreshDialSlots(true);
  assert.ok(slotsFetched >= 1, "the slots endpoint was asked");

  // A display form of the same number matches the stored remote.
  const slot = dialgate.cooldownFor("+49 (89) 123456");
  assert.ok(slot, "display form matches the sanitized remote");
  assert.equal(slot.window, 3600);
  assert.equal(dialgate.cooldownFor("+1715"), null, "unmatched numbers decorate nothing");
  assert.equal(dialgate.cooldownFor(""), null, "blank input matches nothing");

  // The remainder phrases neutral units and drains with the stamp.
  const left = dialgate.cooldownRemainder(slot);
  assert.ok(left === "50m" || left === "49m", `remainder near 50m, got ${left}`);
  assert.equal(
    dialgate.cooldownRemainder({ last_at: now - 7200, window: 3600 }),
    "",
    "an expired slot phrases nothing",
  );

  // A failed refresh keeps the previous cache (degrade to hidden).
  useFetch(() => {
    throw new Error("network gone");
  });
  await dialgate.refreshDialSlots(true);
  assert.ok(dialgate.cooldownFor("+4989123456"), "failed refresh keeps server truth");
});
