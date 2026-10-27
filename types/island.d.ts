// Island-wide type declarations: the DOM handle contract and the wire
// shapes the island reads.
//
// This is a SCRIPT-style .d.ts (no imports/exports) on purpose: every
// declaration here is global, so a module never imports a type — it just
// uses it, exactly like the JSDoc annotations that reference it.
//
// Every id in IslandDomIds is served markup pinned by
// TestServedPageHoldsTheDomContract (docs/dom-contract.md is the list's
// one home). Declaring them means a typo'd id or a renamed element is a
// compile error instead of a silent `null` at runtime.

interface IslandDomIds {
  "reg-status": HTMLElement;
  lang: HTMLSelectElement;
  "login-view": HTMLElement;
  "login-form": HTMLFormElement;
  "login-error": HTMLElement;
  ext: HTMLInputElement;
  pass: HTMLInputElement;
  remember: HTMLInputElement;
  "phone-view": HTMLElement;
  "whoami-ext": HTMLElement;
  logout: HTMLButtonElement;
  "dial-form": HTMLFormElement;
  dest: HTMLInputElement;
  calls: HTMLElement;
  keypad: HTMLElement;
  "history-list": HTMLElement;
  "history-wrap": HTMLDetailsElement;
  "incoming-call": HTMLElement;
  "incoming-from": HTMLElement;
  "incoming-queue": HTMLElement;
  "accept-btn": HTMLButtonElement;
  "reject-btn": HTMLButtonElement;
  "contacts-wrap": HTMLDetailsElement;
  "contacts-list": HTMLElement;
  "vm-wrap": HTMLDetailsElement;
  "vm-badge": HTMLElement;
  "vm-list": HTMLElement;
  "vm-refresh": HTMLButtonElement;
  "vm-status": HTMLElement;
  "ice-wrap": HTMLDetailsElement;
  "ice-panel": HTMLElement;
  log: HTMLElement;
  toasts: HTMLElement;
  "dial-error": HTMLElement;
  "wp-dial-hint": HTMLElement;
  "offline-banner": HTMLElement;
  "remote-audio": HTMLAudioElement;
  "wp-adv-toggle": HTMLButtonElement;
  "wp-advanced": HTMLElement;
  "wp-devtest-btn": HTMLButtonElement;
}

/** The island's four feedback kinds (ui.js owns the vocabulary). */
type IslandToastKind = "ok" | "error" | "warn" | "info";

/** The view states CALL_TRANSITIONS declares, in one place. */
type IslandCallViewState =
  | "ringing"
  | "active"
  | "holding"
  | "resuming"
  | "held"
  | "ending"
  | "ended";
