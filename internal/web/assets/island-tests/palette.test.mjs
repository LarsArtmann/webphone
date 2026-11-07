// The command palette under node:test (F12/M30/M41): the overlay is a
// wired combobox (roles, expanded state, activedescendant), nav labels
// never carry badge digits, starts-with ranks before substring, and
// Home/End jump. Driven through the served header button and the real
// input listeners — no exports.
import test from "node:test";
import assert from "node:assert/strict";

import { installBrowserGlobals } from "./helpers.mjs";

const doc = installBrowserGlobals();
await import("../shell.js");

const fireKey = (input, key) =>
  (input.listeners.keydown ?? []).forEach((fn) =>
    fn({ key, preventDefault() {} }),
  );

// The header button TOGGLES; tests that mean "open" close any open
// overlay first so the click cannot land as a close.
const openPalette = () => {
  const current = palette();
  if (current && !current.hidden) {
    fireKey(current.querySelector(".wp-overlay-input"), "Escape");
  }
  doc.getElementById("wp-palette-open").click();
  return palette();
};

const palette = () => doc.body.querySelector("#wp-palette");

test("the palette opens as a wired combobox with a touch door", () => {
  const panel = openPalette();
  assert.ok(panel, "the header button opens the overlay");
  const input = panel.querySelector(".wp-overlay-input");
  assert.equal(input.getAttribute("role"), "combobox");
  assert.equal(input.getAttribute("aria-expanded"), "true");
  assert.equal(input.getAttribute("aria-controls"), "wp-palette-list");
  const list = panel.querySelector(".wp-overlay-list");
  assert.equal(list.getAttribute("role"), "listbox");
  const first = list.children[0];
  assert.equal(first.getAttribute("role"), "option");
  assert.equal(first.getAttribute("aria-selected"), "true");
  assert.equal(input.getAttribute("aria-activedescendant"), first.id, "the SR tracks the selection");
});

test("palette rows never read badge digits", () => {
  const nav = doc.getElementById("wp-nav");
  const link = doc.createElement();
  link.className = "wp-nav-link";
  link.selector = "#wp-nav .wp-nav-link";
  link.textContent = "Messages4";
  const badge = doc.createElement();
  badge.className = "wp-nav-badge";
  badge.textContent = "4";
  link.append(badge);
  nav.append(link);

  const panel = openPalette();
  const list = panel.querySelector(".wp-overlay-list");
  const labels = list.children.map((row) => row.textContent);
  assert.ok(
    labels.includes("Go to Messages"),
    "the nav name lands without the badge digit",
  );
  assert.ok(
    !labels.some((label) => label.includes("Messages4")),
    "no row ever reads the badge count",
  );
});

test("starts-with ranks before substring and Home/End jump", () => {
  const nav = doc.getElementById("wp-nav");
  const camel = doc.createElement();
  camel.className = "wp-nav-link";
  camel.selector = "#wp-nav .wp-nav-link";
  camel.textContent = "Camel";
  nav.append(camel);

  const panel = openPalette();
  const input = panel.querySelector(".wp-overlay-input");
  const list = panel.querySelector(".wp-overlay-list");

  input.value = "c";
  (input.listeners.input ?? []).forEach((fn) => fn({}));
  const labels = list.children.map((row) => row.textContent);
  assert.deepEqual(
    labels,
    ["Call a number", "Cycle color theme", "Go to Camel"],
    "substring-only matches rank after starts-with matches",
  );

  fireKey(input, "End");
  const last = list.children[labels.length - 1];
  assert.equal(input.getAttribute("aria-activedescendant"), last.id, "End jumps to the last row");
  fireKey(input, "Home");
  const firstRow = list.children[0];
  assert.equal(input.getAttribute("aria-activedescendant"), firstRow.id, "Home returns to the first row");
  assert.equal(firstRow.className.includes("wp-selected"), true, "Home re-paints the selection");
});
