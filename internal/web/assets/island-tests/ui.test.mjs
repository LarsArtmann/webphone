// ui.js under node:test — the toast behavior the Go asset tripwires can
// only grep for. announce() is the consumer of the server's HX-Trigger
// ToastDetail payload ({message, kind} → `.toast .toast-<kind>`).
import test from "node:test";
import assert from "node:assert/strict";

import { installBrowserGlobals } from "./helpers.mjs";

installBrowserGlobals();
const ui = await import("../island/app/ui.js");

test("announce renders a toast element with kind and message", () => {
  const toasts = document.getElementById("toasts");
  ui.announce("Message sent.", "ok");
  assert.equal(toasts.children.length, 1);
  assert.equal(toasts.children[0].className, "toast toast-ok");
  assert.equal(toasts.children[0].textContent, "Message sent.");
});

test("announce caps the stack at four and drops the oldest", () => {
  const toasts = document.getElementById("toasts");
  for (let i = 0; i < 6; i++) ui.announce(`m${i}`, "info");
  assert.equal(toasts.children.length, 4);
  assert.equal(toasts.children[0].textContent, "m2");
});

test("unknown kinds fall back to the info lifetime without crashing", () => {
  const toasts = document.getElementById("toasts");
  ui.announce("odd kind", "nope");
  const last = toasts.children.at(-1);
  assert.equal(last.className, "toast toast-nope");
});
