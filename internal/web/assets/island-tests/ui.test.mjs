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

test("toastKindFor keeps the server's island vocabulary exact", () => {
  // notifyToast (toast.go) emits exactly these kinds; a drift on either
  // side must fail here instead of recoloring toasts to info.
  assert.equal(ui.toastKindFor("ok"), "ok");
  assert.equal(ui.toastKindFor("error"), "error");
  assert.equal(ui.toastKindFor("warn"), "warn");
  assert.equal(ui.toastKindFor("info"), "info");
});

test("toastKindFor tolerates the dispatch-layer vocabulary and junk", () => {
  assert.equal(ui.toastKindFor("success"), "ok");
  assert.equal(ui.toastKindFor("warning"), "warn");
  assert.equal(ui.toastKindFor("unexpected"), "info");
  assert.equal(ui.toastKindFor(undefined), "info");
});
