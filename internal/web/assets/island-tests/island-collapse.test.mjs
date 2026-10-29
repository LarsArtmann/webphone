// The collapsible phone panel (P2) at LOAD time.
//
// A separate test module on purpose: node gives each test file its own
// module registry, so importing shell.js here evaluates it fresh — which is
// the only way to observe the load-time re-assert. The main shell suite
// imports shell.js once and cannot re-run it (query-string re-imports
// return the cached CommonJS module).
//
// What is being pinned: theme-preload.js puts `wp-island-collapsed` on
// <html> before first paint, long before the button exists. shell.js must
// then describe reality — a button that says aria-expanded="true" while the
// island is hidden is a lying control, and screen readers are the only
// surface where that lie is visible.
import test from "node:test";
import assert from "node:assert/strict";

import { installBrowserGlobals } from "./helpers.mjs";

const doc = installBrowserGlobals();
doc.documentElement.classList.add("wp-island-collapsed");
await import("../shell.js");

test("a pre-paint collapsed island is reflected in aria-expanded on load", () => {
  assert.equal(doc.getElementById("wp-island-toggle").getAttribute("aria-expanded"), "false");
});
