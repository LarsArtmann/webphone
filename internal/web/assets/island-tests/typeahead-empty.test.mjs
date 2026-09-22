// typeahead.js degradation: with NO contacts in the served PBX_CONFIG
// (the common small-PBX case) the typeahead never wires anything — the
// dial field stays the plain input it always was. Own file because
// config.js evaluates at import time (per-file process isolation).
import test from "node:test";
import assert from "node:assert/strict";

import { installBrowserGlobals } from "./helpers.mjs";

const doc = installBrowserGlobals();
globalThis.window = { PBX_CONFIG: {} };

const { els } = await import("../island/app/ui.js");
const { rankContacts, initDialTypeahead } = await import("../island/app/typeahead.js");

test("without contacts the typeahead never wires anything", () => {
  initDialTypeahead();
  assert.equal(
    els.dialForm.children.find((el) => el.id === "dial-suggest"),
    undefined,
    "no listbox created",
  );
  assert.deepEqual(rankContacts([], "ann"), []);
});
