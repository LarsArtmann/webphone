// Ring-name resolution under node:test (F7): the banner, the queue
// slots and the config-level resolver speak the contact name with the
// digits beside it, personal contacts win over shared, and unknown
// callers stay raw digits.
import test from "node:test";
import assert from "node:assert/strict";

import { installBrowserGlobals } from "./helpers.mjs";

const doc = installBrowserGlobals();
globalThis.window.PBX_CONFIG = {};

const config = await import("../island/app/config.js");
const state = await import("../island/app/state.js");
const ui = await import("../island/app/ui.js");
config.setSharedContacts([{ name: "Marta", number: "+493012345678" }]);
config.setExtraContacts([{ name: "Bob Personal", number: "030 999 888" }]);

test("the banner and slots carry name (digits)", () => {
  const queue = state.state.incomingSessions;
  queue.push({ invitation: {}, from: "+493012345678" });
  queue.push({ invitation: {}, from: "030 999 888" });
  ui.renderIncomingSlots();

  assert.equal(
    doc.getElementById("incoming-from").textContent,
    "Marta (+493012345678)",
    "slot 0 answers with the contact name",
  );
  const rows = doc.getElementById("incoming-queue").children;
  assert.match(
    rows[0].children[0].textContent,
    /Bob Personal \(030 999 888\)/,
    "the extra slot resolves its own caller",
  );
  queue.length = 0;
});

test("unknown callers stay raw digits", () => {
  const queue = state.state.incomingSessions;
  queue.push({ invitation: {}, from: "+19995550000" });
  ui.renderIncomingSlots();
  assert.equal(
    doc.getElementById("incoming-from").textContent,
    "+19995550000",
    "the honest fallback is the number",
  );
  queue.length = 0;
});

test("personal contacts win and digits normalize", () => {
  assert.equal(config.contactName("030999888"), "Bob Personal");
  assert.equal(config.callerDisplay("030999888"), "Bob Personal (030 999 888)");
  assert.equal(config.contactName("+493012345678"), "Marta");
  assert.equal(config.contactName("nosuch"), "");
});
