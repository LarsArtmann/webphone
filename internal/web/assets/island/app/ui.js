// DOM handles and feedback primitives (event log, toasts, status pill,
// inline dial error). Every id in `els` is part of the published DOM
// contract — external E2E harnesses drive this page through them (see
// AGENTS.md before renaming anything).

export const $ = (id) => document.getElementById(id);

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
  remoteAudio: $("remote-audio"),
};

export function log(message, level = "info") {
  const entry = document.createElement("li");
  entry.textContent = `${new Date().toISOString().slice(11, 19)} ${message}`;
  if (level !== "info") entry.dataset.level = level;
  els.log.prepend(entry);
  while (els.log.children.length > 100) els.log.lastChild.remove();
  // Devtools mirror: the on-page log is the operator trail, the browser
  // console is where a user actually looks when something misbehaves.
  console[level === "info" ? "info" : level]("webphone:", message);
}

// --- toasts: action feedback visible without opening the event log -------

const TOAST_MAX = 4;
const TOAST_MS = { info: 4000, ok: 4000, warn: 6000, error: 8000 };

export function announce(message, kind = "info") {
  const toast = document.createElement("div");
  toast.className = `toast toast-${kind}`;
  toast.textContent = message;
  toast.addEventListener("click", () => toast.remove());
  els.toasts.append(toast);
  while (els.toasts.children.length > TOAST_MAX) els.toasts.firstChild.remove();
  setTimeout(() => toast.remove(), TOAST_MS[kind] || TOAST_MS.info);
}

// Inline dial-form error: the failure lands where the user is looking.
export function showDialError(message) {
  els.dialError.textContent = message;
  els.dialError.hidden = false;
  clearTimeout(showDialError.timer);
  showDialError.timer = setTimeout(() => {
    els.dialError.hidden = true;
  }, 6000);
}

export function setRegStatus(state, text) {
  els.regStatus.textContent = text;
  els.regStatus.className = `status ${state}`;
}

export function dialFromUi(number) {
  els.dest.value = number;
  els.dialForm.requestSubmit();
}
