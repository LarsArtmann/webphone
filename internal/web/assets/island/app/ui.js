// DOM handles and feedback primitives (event log, toasts, status pill,
// inline dial error). Every id in `els` is part of the published DOM
// contract — external E2E harnesses drive this page through them (see
// AGENTS.md before renaming anything).

import { sipDomain } from "./config.js";
import { t } from "./i18n.js";
import { state } from "./state.js";

/**
 * The DOM handle lookup. Every id it accepts is served markup pinned by
 * TestServedPageHoldsTheDomContract (docs/dom-contract.md), so the result
 * is non-null BY CONTRACT and an unknown id is a compile error rather
 * than a silent `null`.
 * @template {keyof IslandDomIds} K
 * @param {K} id
 * @returns {IslandDomIds[K]}
 */
export const $ = (id) =>
  /** @type {IslandDomIds[K]} */ (document.getElementById(id));

export const els = {
  regStatus: $("reg-status"),
  lang: $("lang"),
  loginView: $("login-view"),
  loginForm: $("login-form"),
  loginError: $("login-error"),
  ext: $("ext"),
  pass: $("pass"),
  remember: $("remember"),
  phoneView: $("phone-view"),
  whoami: $("whoami-ext"),
  logout: $("logout"),
  dialForm: $("dial-form"),
  dest: $("dest"),
  calls: $("calls"),
  keypad: $("keypad"),
  history: $("history-list"),
  historyWrap: $("history-wrap"),
  incoming: $("incoming-call"),
  incomingFrom: $("incoming-from"),
  incomingQueue: $("incoming-queue"),
  accept: $("accept-btn"),
  reject: $("reject-btn"),
  contactsWrap: $("contacts-wrap"),
  contactsList: $("contacts-list"),
  vmWrap: $("vm-wrap"),
  vmBadge: $("vm-badge"),
  vmList: $("vm-list"),
  vmRefresh: $("vm-refresh"),
  vmStatus: $("vm-status"),
  iceWrap: $("ice-wrap"),
  icePanel: $("ice-panel"),
  log: $("log"),
  toasts: $("toasts"),
  dialError: $("dial-error"),
  dialHint: $("wp-dial-hint"),
  offlineBanner: $("offline-banner"),
  remoteAudio: $("remote-audio"),
};

/**
 * One event-log line. English by policy: #log is the operator trail in
 * both UI languages.
 * @param {string} message
 * @param {IslandToastKind} [level]
 */
export function log(message, level = "info") {
  const entry = document.createElement("li");
  entry.textContent = `${new Date().toISOString().slice(11, 19)} ${message}`;
  if (level !== "info") entry.dataset.level = level;
  els.log.prepend(entry);
  while (els.log.children.length > 100) els.log.lastChild?.remove();
  // Devtools mirror: the on-page log is the operator trail, the browser
  // console is where a user actually looks when something misbehaves.
  const mirror =
    level === "error"
      ? console.error
      : level === "warn"
        ? console.warn
        : console.info;
  mirror("webphone:", message);
}

/**
 * The message of a caught unknown. A throw site may throw anything;
 * the island only ever renders a string, so this is the ONE narrowing
 * point (a bare `err.message` is a compile error under checkJs and a
 * `undefined` in the copy when someone throws a string).
 * @param {unknown} err
 * @returns {string}
 */
export function errorMessage(err) {
  if (err instanceof Error) return err.message;
  if (typeof err === "string") return err;
  return String(err);
}

// --- toasts: action feedback visible without opening the event log -------

const TOAST_MAX = 4;
const TOAST_MS = { info: 4000, ok: 4000, warn: 6000, error: 8000 };

// The server's HX-Trigger toasts carry the ISLAND's kind vocabulary
// ("ok"/"error"/"warn"/"info" — notifyToast in toast.go owns it; the
// library only owns the wire shape). The dispatch-layer vocabulary
// (success/warning) is accepted too so a future library-emitted toast
// degrades to the right color instead of plain info.
/** @type {Record<string, IslandToastKind>} */
const TOAST_KINDS = {
  ok: "ok",
  error: "error",
  warn: "warn",
  info: "info",
  success: "ok",
  warning: "warn",
};

/**
 * @param {string} kind
 * @returns {IslandToastKind}
 */
export function toastKindFor(kind) {
  return TOAST_KINDS[kind] || "info";
}

// Politeness policy (M17 J9 audit): every announcement is POLITE by
// default — toasts are status, never interruptions. The single
// assertive exception is the incoming call: a ringing phone must break
// through whatever the screen reader is reading, so that one toast
// carries role="alert" (aria-live assertive) on the node itself while
// the host region stays polite for everything else.
/**
 * @param {string} message
 * @param {IslandToastKind} [kind]
 * @param {{ assertive?: boolean }} [options]
 */
export function announce(message, kind = "info", { assertive = false } = {}) {
  // Identical-consecutive dedup: a retry loop (or a repeated identical
  // failure) must not stack look-alike toasts; a different message or an
  // expired first toast passes freely.
  const host = els.toasts;
  const last = host.children[host.children.length - 1];
  if (last && last.textContent === message) return;
  const toast = document.createElement("div");
  toast.className = `toast toast-${kind}`;
  if (assertive) toast.setAttribute("role", "alert");
  toast.textContent = message;
  // Focusable so keyboard users can dismiss (Enter/Space/Escape); the
  // container's role="status" aria-live="polite" already announced it to
  // screen readers — focus is never stolen on creation.
  toast.tabIndex = 0;
  const dismiss = () => toast.remove();
  toast.addEventListener("click", dismiss);
  toast.addEventListener("keydown", (event) => {
    if (event.key === "Enter" || event.key === " " || event.key === "Escape") {
      event.preventDefault();
      dismiss();
    }
  });
  els.toasts.append(toast);
  while (els.toasts.children.length > TOAST_MAX)
    els.toasts.firstChild?.remove();
  setTimeout(dismiss, TOAST_MS[kind] ?? TOAST_MS.info);
}

/** @type {ReturnType<typeof setTimeout> | null} */
let dialErrorTimer = null;
// Inline dial-form error: the failure lands where the user is looking.
/** @param {string} message */
export function showDialError(message) {
  els.dialError.textContent = message;
  els.dialError.hidden = false;
  if (dialErrorTimer) clearTimeout(dialErrorTimer);
  dialErrorTimer = setTimeout(() => {
    els.dialError.hidden = true;
  }, 6000);
}

/**
 * @param {string} state
 * @param {string} text
 */
export function setRegStatus(state, text) {
  els.regStatus.textContent = text;
  els.regStatus.className = `status ${state}`;
}

// The offline banner: the loud honest surface for "you cannot call right
// now". Visibility is the only API — its text lives in the markup with
// data-i18n so a language switch re-localizes it; the precise reason
// stays in the #reg-status pill.
/** @param {boolean} visible */
export function setOfflineBanner(visible) {
  els.offlineBanner.hidden = !visible;
}

// Incoming-call focus mode (M25 A6): while a call rings, the shell dims
// (app.css rides the root class — the island owns the ringing state)
// and every banner path funnels through these helpers so the class can
// never outlive the banner. The dim is visual only: nothing is
// pointer-blocked, a user mid-task keeps full control.
/** @param {boolean} active */
export function setIncomingFocusMode(active) {
  document.documentElement.classList.toggle("wp-incoming-focus", active);
}

// The ringing queue's slot rendering (the N-deep banner stack). Slot 1
// IS the served markup: the pinned ids (#incoming-from, #accept-btn,
// #reject-btn) belong to the OLDEST ring, so the stack's browser E2E
// keeps driving the first invite through them. Slots 2..N render as
// per-invite rows into #incoming-queue, each with its own answer/reject
// pair. The extra-slot buttons ride the actions registry below — ui.js
// must not import calls.js (calls imports ui; the registry keeps the
// graph one-way).
/** @type {IslandIncomingSlotActions | null} */
let slotActions = null;

/** @param {IslandIncomingSlotActions} actions */
export function setIncomingSlotActions(actions) {
  slotActions = actions;
}

export function renderIncomingSlots() {
  const queue = state.incomingSessions;
  if (!queue.length) return;
  els.incoming.hidden = false;
  setIncomingFocusMode(true);
  els.incomingFrom.textContent = queue[0].from;
  if (!els.incomingQueue) return;
  els.incomingQueue.replaceChildren();
  for (let i = 1; i < queue.length; i += 1) {
    const entry = queue[i];
    const row = document.createElement("div");
    row.className = "incoming-slot";
    row.id = `wp-incoming-slot-${i}`;
    const label = document.createElement("span");
    label.className = "incoming-slot-from";
    label.textContent = `${t("incoming")} ${entry.from}`;
    const controls = document.createElement("div");
    controls.className = "controls";
    const answer = document.createElement("button");
    answer.className = "primary";
    answer.textContent = t("accept");
    answer.addEventListener("click", () => {
      if (slotActions) slotActions.answer(i);
    });
    const reject = document.createElement("button");
    reject.className = "danger";
    reject.textContent = t("reject");
    reject.addEventListener("click", () => {
      if (slotActions) slotActions.reject(i);
    });
    controls.append(answer, reject);
    row.append(label, controls);
    els.incomingQueue.append(row);
  }
}

export function hideIncomingBanner() {
  els.incoming.hidden = true;
  setIncomingFocusMode(false);
  if (els.incomingQueue) els.incomingQueue.replaceChildren();
}

/** @param {string} number */
export function dialFromUi(number) {
  els.dest.value = number;
  els.dialForm.requestSubmit();
}

// whoamiLine renders the signed-in line for an identity-bearing session
// payload: the display name plus the user's numbers when the deployment
// knows them, the plain extension otherwise. It lives in this NEUTRAL
// module (not the passkey feature module) because BOTH login paths
// render it — the resume path and the passkey path carry the same
// fields.
/**
 * @param {IslandSessionIdentity} data
 * @returns {string}
 */
export function whoamiLine(data) {
  const numbers = Array.isArray(data.numbers)
    ? data.numbers.filter(Boolean)
    : [];
  if (data.display_name) {
    const tail = numbers.length
      ? numbers.join(", ")
      : `${data.extension}@${sipDomain}`;
    return `${data.display_name} · ${tail}`;
  }
  if (numbers.length) {
    return `${data.extension}@${sipDomain} · ${numbers.join(", ")}`;
  }
  return `${data.extension}@${sipDomain}`;
}

// The diagnostics gear (K5): collapses the ICE panel + event log on
// request. Ships OPEN — the E2E and the operator runbook must never
// meet a hidden diagnostics tree — and persists the choice, the same
// key style as the theme/extension precedents.
const ADVANCED_COLLAPSED_KEY = "wp-advanced-collapsed";

export function initAdvancedToggle() {
  const toggle = $("wp-adv-toggle");
  const advanced = $("wp-advanced");
  if (!toggle || !advanced) return;

  /** @param {boolean} collapsed */
  const apply = (collapsed) => {
    advanced.hidden = collapsed;
    toggle.setAttribute("aria-expanded", collapsed ? "false" : "true");
  };
  apply(localStorage.getItem(ADVANCED_COLLAPSED_KEY) === "1");

  toggle.addEventListener("click", () => {
    const collapsed = localStorage.getItem(ADVANCED_COLLAPSED_KEY) !== "1";
    localStorage.setItem(ADVANCED_COLLAPSED_KEY, collapsed ? "1" : "0");
    apply(collapsed);
  });
}
