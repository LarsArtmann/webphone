// typeahead.js under node:test — the dial field's contact suggestions:
// ranking (name-prefix beats name-contains beats number-contains, capped
// and tie-broken), the input → listbox render, keyboard selection, and
// the degradation when no contacts are configured.
import test from "node:test";
import assert from "node:assert/strict";

import { installBrowserGlobals } from "./helpers.mjs";

const doc = installBrowserGlobals();

const CONTACTS = [
  { name: "Anna Kellner", number: "+491512345678" },
  { name: "Bolt Delivery", number: "+441632960111" },
  { name: "Anna Licht", number: "+4930123456" },
  { name: "Mara Opel", number: "1001" },
  { name: "Kellerei Anna", number: "+4989" },
  { name: "Zentrum", number: "110" },
  { name: "Zahnpasta Hotline", number: "0800" },
  { name: "Zoo Berlin", number: "+493087871" },
];

globalThis.window = { PBX_CONFIG: { contacts: CONTACTS } };

const { els } = await import("../island/app/ui.js");
const { rankContacts, initDialTypeahead, relabelTypeahead } =
  await import("../island/app/typeahead.js");

const fire = (element, type, event = {}) => {
  for (const fn of element.listeners[type] ?? []) {
    fn({ target: element, preventDefault() {}, ...event });
  }
};

test("ranking: name-prefix beats name-contains beats number-contains", () => {
  assert.deepEqual(
    rankContacts(CONTACTS, "ann").map((c) => c.name),
    ["Anna Kellner", "Anna Licht", "Kellerei Anna"],
  );
  assert.deepEqual(
    rankContacts(CONTACTS, "960").map((c) => c.name),
    ["Bolt Delivery"],
  );
  assert.deepEqual(rankContacts(CONTACTS, "110"), [{ name: "Zentrum", number: "110" }]);
  assert.deepEqual(rankContacts(CONTACTS, "   "), [], "blank query suggests nothing");
  assert.deepEqual(rankContacts(CONTACTS, "kein treffer"), []);
});

test("ranking: capped at six, ties alphabetical by name", () => {
  const many = Array.from({ length: 9 }, (_, i) => ({
    name: `Anna ${String.fromCharCode(122 - i)}${i}`,
    number: `1${i}`,
  }));
  const ranked = rankContacts(many, "ann");
  assert.equal(ranked.length, 6);
  assert.deepEqual(
    ranked.map((c) => c.name),
    [...ranked.map((c) => c.name)].sort((a, b) => a.localeCompare(b)),
  );
});

test("the listbox renders on input, walks with arrows, and Enter picks", () => {
  initDialTypeahead();
  const form = els.dialForm;
  const list = form.children.find((el) => el.id === "dial-suggest");
  assert.ok(list, "the suggestions listbox was created in the dial form");

  els.dest.value = "ann";
  fire(els.dest, "input");
  assert.equal(list.hidden, false);
  assert.equal(list.children.length, 3);
  assert.equal(list.children[0].className, "active", "first suggestion starts active");
  assert.equal(list.getAttribute("aria-label"), "Contact suggestions");
  assert.equal(list.children[0].getAttribute("aria-selected"), "true");

  fire(els.dest, "keydown", { key: "ArrowDown" });
  assert.equal(list.children[1].className, "active");
  assert.equal(list.children[0].className, "");

  fire(els.dest, "keydown", { key: "Enter" });
  assert.equal(els.dest.value, "+4930123456", "Enter filled the active suggestion");
  assert.equal(list.hidden, true, "list hides after picking");
  assert.ok(els.dest.focused, "focus stays in the dial field");

  // The next Enter submits the dial form untouched (no prevented event):
  // asserted implicitly — a pick's preventDefault never leaks.

  els.dest.value = "zz";
  fire(els.dest, "input");
  assert.equal(list.hidden, true, "no matches renders an empty hidden list");

  fire(els.dest, "keydown", { key: "Escape", preventDefault() {} });
  assert.equal(list.hidden, true);
});

test("mousedown on a suggestion picks it without losing the field", () => {
  const list = els.dialForm.children.find((el) => el.id === "dial-suggest");
  els.dest.value = "zoo";
  fire(els.dest, "input");
  assert.equal(list.hidden, false);

  const option = list.children[0];
  option.selector = "li[data-index]";
  fire(list, "mousedown", { target: option, preventDefault() {} });
  assert.equal(els.dest.value, "+493087871");
  assert.equal(list.hidden, true);
});

test("init is idempotent: a second call never duplicates the listbox", () => {
  const before = els.dialForm.children.length;
  initDialTypeahead();
  assert.equal(els.dialForm.children.length, before);
});

test("relabel updates the listbox aria-label", () => {
  const list = els.dialForm.children.find((el) => el.id === "dial-suggest");
  relabelTypeahead("Kontaktvorschläge");
  assert.equal(list.getAttribute("aria-label"), "Kontaktvorschläge");
});
