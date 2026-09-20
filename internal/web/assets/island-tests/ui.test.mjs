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

const resetToasts = () => {
  const host = document.getElementById("toasts");
  while (host.firstChild) host.firstChild.remove();
  return host;
};

test("toasts are keyboard-dismissable and never steal focus on creation", () => {
  const toasts = resetToasts();
  ui.announce("kbd me", "warn");
  const toast = toasts.children[0];
  assert.equal(toast.tabIndex, 0, "toast must be focusable for dismissal");
  assert.notEqual(toast, document.activeElement);
  toast.listeners.keydown[0]({ key: "Escape", preventDefault() {} });
  assert.equal(toasts.children.length, 0, "Escape must remove the toast");

  ui.announce("enter me", "info");
  const second = toasts.children[0];
  second.listeners.keydown[0]({ key: "Enter", preventDefault() {} });
  assert.equal(toasts.children.length, 0, "Enter must remove the toast");

  ui.announce("ignore me", "info");
  toasts.children[0].listeners.keydown[0]({ key: "Tab", preventDefault() {} });
  assert.equal(toasts.children.length, 1, "other keys must not dismiss");
});

test("announce leaves the live-region host attributes untouched", () => {
  // The server renders #toasts with role="status" aria-live="polite"
  // (phone.templ; pinned on the served page by the Go contract test).
  // announce() appends children only — a regression that overwrote the
  // host attributes (e.g. setting aria-hidden) would blind screen
  // readers to every toast.
  const toasts = resetToasts();
  toasts.setAttribute("role", "status");
  toasts.setAttribute("aria-live", "polite");
  ui.announce("host check", "ok");
  assert.equal(toasts.getAttribute("role"), "status");
  assert.equal(toasts.getAttribute("aria-live"), "polite");
  assert.equal(toasts.getAttribute("aria-hidden"), null);
});

test("identical consecutive toasts are deduped, different ones are not", () => {
  const toasts = resetToasts();
  ui.announce("same", "error");
  ui.announce("same", "error");
  assert.equal(toasts.children.length, 1, "identical repeat must not stack");
  ui.announce("different", "ok");
  assert.equal(toasts.children.length, 2, "a different message passes");
  ui.announce("same", "error");
  assert.equal(
    toasts.children.length,
    3,
    "dedup only compares against the LAST toast",
  );
});
