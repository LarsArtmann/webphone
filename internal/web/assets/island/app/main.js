// Wiring: event handlers, view switching, boot sequence. Everything
// else lives in its feature module — this file is the composition root.

import { ringToneStop } from "./audio.js";
import { connect, disconnect } from "./connection.js";
import {
  answerIncoming,
  placeCall,
  rejectIncoming,
  renderCalls,
  sendDtmf,
  teardownAll,
} from "./calls.js";
import { sipDomain, websocketUrl } from "./config.js";
import "./ice.js";
import { applyI18n, getLang, setLang, t } from "./i18n.js";
import {
  cancelVoicemailRefresh,
  clearContacts,
  refreshServerHistory,
  refreshVoicemail,
  renderContacts,
  renderHistory,
} from "./panels.js";
import { requestNotifications, titleFlashStop } from "./notify.js";
import {
  connectLiveUpdates,
  createSession,
  destroySession,
  fetchLiveSession,
  initSseLiveIndicator,
  signOutQuiet,
} from "./session.js";
import { initShortcuts } from "./shortcuts.js";
import { sessions, state } from "./state.js";
import { announce, els, log, setRegStatus, toastKindFor } from "./ui.js";

const REMEMBER_KEY = "pbx-extension";

if (localStorage.getItem(REMEMBER_KEY)) {
  els.ext.value = localStorage.getItem(REMEMBER_KEY);
  els.remember.checked = true;
}
renderHistory();

if (els.lang) {
  els.lang.value = getLang();
  els.lang.addEventListener("change", () => {
    setLang(els.lang.value);
    applyI18n();
    renderCalls();
    // Re-fetch the open tab partial so the server-rendered tabs switch
    // language too (htmx.ajax keeps the island alive — no reload). The
    // nav lives outside the swap target; shell.js re-labels it on this
    // event.
    const active = document.querySelector(".wp-nav .wp-nav-link.wp-active");
    if (active && window.htmx) {
      window.htmx.ajax("GET", active.getAttribute("hx-get"), {
        target: "#tab-content",
        swap: "innerHTML",
      });
    }
    document.dispatchEvent(new CustomEvent("wp:lang-changed"));
    log(`language switched to ${els.lang.value}`);
  });
}
applyI18n();

// --- boot: resume a live cookie session before showing the login form ------
//
// A returning browser holds a live server session (sliding idle window);
// the resume probe hands its SIP credentials back and the island
// re-registers silently. The form stays hidden during the probe so a
// resumed load never flashes it; EVERY failure path restores it, so a
// broken probe can never leave a dead page.
function showLogin() {
  els.loginView.hidden = false;
}

async function resumeSession() {
  const session = await fetchLiveSession();
  if (!session) {
    showLogin();
    return;
  }
  setRegStatus("status-offline", t("resuming"));
  try {
    await connect(session.extension, session.password);
  } catch (err) {
    // The row's credentials no longer register (the PBX password
    // changed since sign-in): the server session is stale — drop it
    // quietly and put the reason where the user is looking.
    await signOutQuiet();
    els.loginError.textContent = t("resumeRejected")(err.message);
    els.loginError.hidden = false;
    showLogin();
    log(`resumed credentials rejected; server session dropped (${err.message})`, "error");
    return;
  }
  els.whoami.textContent = `${session.extension}@${sipDomain}`;
  els.loginView.hidden = true;
  els.phoneView.hidden = false;
  connectLiveUpdates();
  log("session resumed; registered");
  // The wp:session-opened listener appends the did and refreshes the
  // session-gated panels — the exact post-login wiring, reused.
  document.dispatchEvent(
    new CustomEvent("wp:session-opened", { detail: { did: session.did || "" } }),
  );
}

els.loginView.hidden = true;
resumeSession().catch((err) => {
  log(`session resume failed: ${err.message}`, "error");
  showLogin();
});

els.loginForm.addEventListener("submit", async (event) => {
  event.preventDefault();
  els.loginError.hidden = true;
  // Permission prompts need a user gesture; the login click is the
  // natural one (a later denied state is shown in the log, not nagged).
  requestNotifications();
  try {
    const extension = els.ext.value.trim();
    await connect(extension, els.pass.value);
    // The PBX accepted the REGISTER — these credentials are good. Hand
    // them to the server so the tabs and the /phone-api proxy share the
    // same login (see session.js).
    createSession(extension, els.pass.value);
    if (els.remember.checked) localStorage.setItem(REMEMBER_KEY, extension);
    else localStorage.removeItem(REMEMBER_KEY);
    els.whoami.textContent = `${extension}@${sipDomain}`;
    els.loginView.hidden = true;
    els.phoneView.hidden = false;
    log(`connected via ${websocketUrl}`);
    // Only the shared-contacts render runs here: it needs no server
    // session. The session-gated panels (voicemail badge, server
    // history) are refreshed by the wp:session-opened listener — firing
    // them at submit raced the session POST and 401'd on every login.
    renderContacts();
  } catch (err) {
    els.loginError.textContent = t("loginError")(err.message);
    els.loginError.hidden = false;
    log(`connect failed: ${err.message}`, "error");
  }
});

// The server completes the whoami line with the extension's presented
// number when it knows one (config identities). The session POST runs
// async on purpose, so the DID arrives after the synchronous
// ext@sipDomain write above — append it once, on the event.
document.addEventListener("wp:session-opened", (event) => {
  const did = event.detail && event.detail.did;
  if (did && !els.whoami.textContent.includes(did)) {
    els.whoami.textContent += ` · ${did}`;
  }
  // Session-gated panels wait for the cookie (session.js dispatches this
  // only after the cookie is minted and the fresh CSRF adopted).
  refreshVoicemail();
  refreshServerHistory();
  restoreTabsAfterSignIn();
});

// A signed-out shell renders the welcome hint in the tab area. Once the
// session exists, load the default tab — what a signed-in page render
// would show — so the tabs work without a manual reload after the
// island login. Mirrors the lang-change seam: htmx.ajax, no reload.
function restoreTabsAfterSignIn() {
  const content = document.querySelector("#tab-content");
  if (!content || !content.querySelector(".wp-welcome")) return;
  if (!window.htmx) return;
  window.htmx.ajax("GET", "/partials/messages", {
    target: "#tab-content",
    swap: "innerHTML",
  });
  window.htmx.ajax("GET", "/partials/nav?active=messages", {
    target: "#wp-nav",
    swap: "morph:innerHTML",
  });
}

els.logout.addEventListener("click", async () => {
  try {
    [...sessions.values()].forEach(({ session }) => {
      session.bye().catch(() => {});
    });
    if (state.userAgent) await state.userAgent.stop().catch(() => {});
  } finally {
    disconnect();
    destroySession();
    teardownAll();
    ringToneStop();
    titleFlashStop();
    cancelVoicemailRefresh();
    clearContacts();
    setRegStatus("status-offline", t("offline"));
    els.phoneView.hidden = true;
    els.loginView.hidden = false;
  }
});

els.dialForm.addEventListener("submit", async (event) => {
  event.preventDefault();
  els.dialError.hidden = true;
  await placeCall(els.dest.value.trim());
});

els.accept.addEventListener("click", () => {
  answerIncoming();
});

els.reject.addEventListener("click", () => {
  rejectIncoming();
});

els.vmRefresh.addEventListener("click", () => refreshVoicemail());

initShortcuts();
initSseLiveIndicator();

// Server-driven toasts: tab-action responses carry an HX-Trigger header
// ("showMessage", the cqrs-htmx ToastDetail wire shape {message, kind});
// htmx dispatches it as a DOM event that bubbles to the body. The kind
// mapping lives in ui.js (toastKindFor) so the server's vocabulary stays
// pinned by the island tests.
document.body.addEventListener("showMessage", (event) => {
  const detail = event.detail || {};
  announce(detail.message || "", toastKindFor(detail.kind));
});

els.keypad.querySelectorAll("button[data-tone]").forEach((button) => {
  button.addEventListener("click", () => sendDtmf(button.dataset.tone));
});

log(`webphone loaded; sip domain ${sipDomain}`);
