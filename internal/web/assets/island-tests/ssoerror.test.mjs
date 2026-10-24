// ssoerror.js under node:test — read-then-strip is the contract: the
// ?sso_error sentence renders once from the query parameter, the
// parameter then leaves the address bar (history.replaceState), and a
// simulated reload finds no parameter and shows nothing. Without the
// card element (or without any JS at all) nothing renders and nothing
// strips — the strip is address-bar hygiene, never a render
// dependency.
import test from "node:test";
import assert from "node:assert/strict";

import { installBrowserGlobals } from "./helpers.mjs";

const doc = installBrowserGlobals();
const { t } = await import("../island/app/i18n.js");
const { showSSOError } = await import("../island/app/ssoerror.js");

function card() {
  return doc.getElementById("sso-error");
}

function withLocation(search, hash = "") {
  globalThis.location.search = search;
  globalThis.location.pathname = "/";
  globalThis.location.hash = hash;
}

let lastReplaced;
globalThis.history = {
  replaceState(_state, _title, url) {
    lastReplaced = url;
  },
};

test("renders the localized sentence, then strips the parameter", () => {
  const el = card();
  withLocation("?sso_error=unmapped");
  showSSOError(t);
  assert.equal(el.hidden, false);
  assert.notEqual(el.textContent, "");
  assert.notEqual(el.textContent, "unmapped", "the raw code must never render");
  assert.ok(
    !lastReplaced.includes("sso_error"),
    `the parameter must leave the URL, got ${lastReplaced}`,
  );
});

test("a reload finds no parameter and re-shows nothing", () => {
  const el = card();
  el.hidden = true;
  el.textContent = "";
  withLocation("?"); // the stripped address the reload would see
  lastReplaced = undefined;
  showSSOError(t);
  assert.equal(el.hidden, true, "a stale error must not re-render");
  assert.equal(lastReplaced, undefined, "nothing to strip, no history write");
});

test("an unknown code still shows a sentence (the Failed fallback)", () => {
  const el = card();
  withLocation("?sso_error=future_code");
  showSSOError(t);
  assert.equal(el.hidden, false);
  assert.notEqual(el.textContent, "");
  assert.ok(!lastReplaced.includes("sso_error"));
});

test("no card element: the strip still happens (no-JS-shaped page, JS hygiene)", () => {
  withLocation("?sso_error=unavailable&x=1");
  showSSOError(t);
  assert.ok(
    !lastReplaced.includes("sso_error"),
    "the strip must not depend on the card existing",
  );
  assert.ok(lastReplaced.includes("x=1"), "unrelated parameters survive");
});

test("the directory reason renders its own sentence", () => {
  const el = card();
  withLocation("?sso_error=directory");
  showSSOError(t);
  assert.equal(el.hidden, false);
  assert.notEqual(el.textContent, "");
  assert.notEqual(el.textContent, "directory", "the raw code must never render");
  assert.ok(!lastReplaced.includes("sso_error"));
});
