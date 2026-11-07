// The transfer row (T19 M24): an empty destination is an ANNOUNCED
// mistake (the row stays open for editing, never a silent no-op), and
// Escape closes the row without sending anything.
import test from "node:test";
import assert from "node:assert/strict";

import { installBrowserGlobals } from "./helpers.mjs";

const doc = installBrowserGlobals();
globalThis.SIP = {
  SessionState: {
    Established: "Established",
    Terminated: "Terminated",
  },
  Inviter: class {},
  UserAgent: { makeURI: () => ({}) },
};

const { toggleTransferRow } = await import("../island/app/calls.js?test=transfer");
const { t } = await import("../island/app/i18n.js?test=transfer");

const lastToast = () => doc.getElementById("toasts").children.at(-1);

test("an empty transfer destination announces and keeps the row", () => {
  const card = doc.createElement();
  card.className = "call-card";
  const row = toggleTransferRow(card, "transfer-x1");

  const blind = row.children.find((el) => el.textContent === t("transferBlind"));
  blind.click();
  assert.ok(
    lastToast().textContent.includes(t("transferEmpty")),
    `the empty-destination mistake is announced, got ${lastToast().textContent}`,
  );
  assert.ok(card.children.includes(row), "the row stays open for editing");
});

test("Escape closes the transfer row", () => {
  const card = doc.createElement();
  card.className = "call-card";
  const row = toggleTransferRow(card, "transfer-x2");
  const input = row.children[0];

  input.listeners.keydown.forEach((fn) => fn({ key: "Escape" }));
  assert.ok(!card.children.includes(row), "the row is gone");

  const row2 = toggleTransferRow(card, "transfer-x3");
  input.listeners.keydown.forEach((fn) => fn({ key: "Enter" }));
  assert.ok(card.children.includes(row2), "other keys leave the row open");
  row2.remove();
});
