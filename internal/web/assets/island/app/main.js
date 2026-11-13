// Wiring: event handlers, view switching, boot sequence. Everything
// else lives in its feature module — this file is the composition root.

import { ringToneStop } from "./audio.js";
import {
  connect,
  disconnect,
  networkOnline,
  registerWasRejected,
} from "./connection.js";
import {
  initDialHint,
  initDialWarm,
  placeCall,
  renderCalls,
  sendDtmf,
  teardownAll,
} from "./calls.js";
import { sipDomain } from "./config.js";
import "./ice.js";
import { applyI18n, getLang, setLang, t } from "./i18n.js";
import {
  cancelVoicemailRefresh,
  clearContacts,
  refreshServerHistory,
  refreshVoicemail,
  renderHistory,
} from "./panels.js";
import { titleFlashStop } from "./notify.js";
import {
  connectLiveUpdates,
  destroySession,
  fetchLiveSession,
  initSseLiveIndicator,
  renderLineSwitcher,
  signOutQuiet,
} from "./session.js";
import { initShortcuts } from "./shortcuts.js";
import { sessions, state } from "./state.js";
import { initDialTypeahead, relabelTypeahead } from "./typeahead.js";
import {
  announce,
  els,
  errorMessage,
  initAdvancedToggle,
  log,
  setOfflineBanner,
  setRegStatus,
  toastKindFor,
  whoamiLine,
} from "./ui.js";
import { initAudioOutput } from "./audioout.js";
import { initDeviceCheck } from "./selftest.js";

renderHistory();

if (els.lang) {
  els.lang.value = getLang();
  els.lang.addEventListener("change", () => {
    setLang(els.lang.value === "de" ? "de" : "en");
    applyI18n();
    renderCalls();
    // Re-fetch the open tab partial so the server-rendered tabs switch
    // language too (htmx.ajax keeps the island alive — no reload). The
    // nav lives outside the swap target; shell.js re-labels it on this
    // event.
    const active = document.querySelector(".wp-nav .wp-nav-link.wp-active");
    const path = active && active.getAttribute("hx-get");
    if (path && window.htmx) {
      window.htmx.ajax("GET", path, {
        target: "#tab-content",
        swap: "innerHTML",
      });
    }
    document.dispatchEvent(new CustomEvent("wp:lang-changed"));
    relabelTypeahead(t("typeaheadLabel"));
    log(`language switched to ${els.lang.value}`);
  });
}
applyI18n();

// --- boot: resume a live cookie session ------------------------------------
//
// A returning browser holds a live server session (sliding idle window);
// the resume probe hands its SIP credentials back and the island
// re-registers silently. The server only renders the shell with a live
// session (anonymous / redirects to /login), so a probe that finds
// nothing either raced an expiry or hit a broken network answer — both
// go back to the front door instead of a dead island.

async function resumeSession() {
  const session = await fetchLiveSession();
  if (!session) {
    window.location.replace("/login");
    return;
  }
  setRegStatus("status-offline", t("resuming"));
  try {
    await connect(
      session.extension,
      session.sip_credential || session.password,
    );
    log("session resumed; registered");
  } catch (err) {
    if (registerWasRejected()) {
      // The row's credentials no longer register (the PBX password
      // changed since sign-in): the server session is stale — drop it
      // quietly and let the login page say why (its ?error= alert is
      // the one surface; the English sentence is server copy there).
      await signOutQuiet();
      log(
        `resumed credentials rejected; server session dropped (${errorMessage(err)})`,
        "error",
      );
      window.location.replace(
        `/login?error=${encodeURIComponent("Your saved sign-in was rejected by the phone system; sign in again.")}`,
      );
      return;
    }
    // A transport failure (PBX/WS down at load) is NOT a stale
    // session: keep the cookie, show the phone in its honest offline
    // state, and let the reconnect backoff own recovery. The tabs
    // stay usable — the server session was never the problem.
    networkOnline();
    log(`session resumed; SIP unreachable (${errorMessage(err)})`, "error");
  }
  // The whoami line honors the session's identity fields (a passkey
  // session carries display_name/numbers; an extension session keeps
  // the plain ext@sipDomain line) — one renderer for both.
  els.whoami.textContent = whoamiLine(session);
  // The line switcher renders only for a multi-line mapped user (D1);
  // single-line and extension sessions render nothing.
  renderLineSwitcher(session);
  els.phoneView.hidden = false;
  connectLiveUpdates();
  // The notification-permission state line: the login page's submit
  // requested the prompt under its own user gesture; the boot states
  // the RESULT where the operator trail lives (#log, English).
  if (typeof Notification !== "undefined") {
    log(`notifications: ${Notification.permission}`);
  }
  // The wp:session-opened listener appends the did and refreshes the
  // session-gated panels — the exact post-login wiring, reused. The
  // agent_mode* fields feed the shell's drafts pill (M13).
  document.dispatchEvent(
    new CustomEvent("wp:session-opened", {
      detail: {
        did: session.did || "",
        agentMode: session.agent_mode || "",
        agentModeSource: session.agent_mode_source || "",
        agentDefaultMode: session.agent_default_mode || "",
        agentModes: Array.isArray(session.agent_modes)
          ? session.agent_modes
          : [],
      },
    }),
  );
}

resumeSession().catch((err) => {
  log(`session resume failed: ${errorMessage(err)}`, "error");
  window.location.replace("/login");
});

// The server completes the whoami line with the extension's presented
// number when it knows one (config identities). Append it once, on the
// event — a passkey/SSO/resumed session knows the did only after the
// probe's answer lands.
document.addEventListener("wp:session-opened", (event) => {
  /** @type {{did?: string}} */
  const detail = /** @type {CustomEvent} */ (event).detail || {};
  const did = detail.did;
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
  // E3: the shell remembers the last-active tab; a fresh sign-in lands
  // where the user left off (messages when nothing is stored).
  const stored = localStorage.getItem("wp-last-tab") || "";
  const known = [
    "messages",
    "fax",
    "voicemail",
    "history",
    "contacts",
    "settings",
  ];
  const tab = known.includes(stored) ? stored : "messages";
  window.htmx.ajax("GET", `/partials/${tab}`, {
    target: "#tab-content",
    swap: "innerHTML",
  });
  window.htmx.ajax("GET", `/partials/nav?active=${tab}`, {
    target: "#wp-nav",
    swap: "morph:innerHTML",
  });
}

// M21: logout with a live call is a two-click move — the first click
// warns (the misclick must not hang up on the operator), the second
// within the arm window confirms. An idle phone signs out on one click.
let logoutArmedAt = 0;
els.logout.addEventListener("click", async () => {
  if (sessions.size > 0 && Date.now() - logoutArmedAt > 4000) {
    logoutArmedAt = Date.now();
    announce(t("logoutLiveCall"), "warn");
    return;
  }
  logoutArmedAt = 0;
  try {
    [...sessions.values()].forEach(({ session }) => {
      // Logout tears the page down next: a failed bye has no surface
      // left to render and the server ends the call anyway.
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
    // destroySession reloads; the anonymous page then redirects to the
    // login page (the server never renders the shell without a session).
  }
});

// The double-submit guard: two rapid Enters must not race two gate
// consults into two INVITEs (the dial form has no disabled-button
// protection while the gate round-trip is in flight — Enter submits
// bypass button state). One pending dial at a time; the flag resets
// when placeCall settles (dialed or not).
let dialPending = false;

els.dialForm.addEventListener("submit", async (event) => {
  event.preventDefault();
  if (dialPending) return;
  dialPending = true;
  const submitButton = /** @type {HTMLButtonElement | null} */ (
    els.dialForm.querySelector('button[type="submit"]')
  );
  if (submitButton) submitButton.disabled = true;
  els.dialError.hidden = true;
  // t0 of the dial_to_audio latency interval (the measure-first design):
  // the submit IS the dial activation — the click or Enter that starts
  // everything, gate consult included.
  try {
    const dialed = await placeCall(els.dest.value.trim(), {
      latencyT0: Date.now(),
    });
    // Clear only on a real INVITE: a validation error keeps the typed
    // text for editing, and a stale value must never append onto the
    // next dial (the E2E concatenated two extensions exactly once too
    // often).
    if (dialed) els.dest.value = "";
  } finally {
    dialPending = false;
    if (submitButton) submitButton.disabled = false;
  }
});

els.vmRefresh.addEventListener("click", () => refreshVoicemail());

initShortcuts();
// The passkey front door lives on the library-owned /login page now —
// the island card links to it (plain navigation, no island JS).
initSseLiveIndicator();
initDialTypeahead();
initDialHint();
initDialWarm();
initAdvancedToggle();
initDeviceCheck();
relabelTypeahead(t("typeaheadLabel"));
initAudioOutput();

// The browser's own network truth is the EARLIEST offline signal (the
// websocket notices later). "offline" says so immediately; "online"
// only nudges recovery — it never claims registered, that verdict
// stays with the REGISTER.
window.addEventListener("offline", () => {
  setRegStatus("status-offline", t("browserOffline"));
  setOfflineBanner(true);
  log("browser reports network offline", "warn");
});
window.addEventListener("online", () => {
  log("browser reports network back");
  networkOnline();
});

// Server-driven toasts: tab-action responses carry an HX-Trigger header
// ("showMessage", the cqrs-htmx ToastDetail wire shape {message, kind});
// htmx dispatches it as a DOM event that bubbles to the body. The kind
// mapping lives in ui.js (toastKindFor) so the server's vocabulary stays
// pinned by the island tests.
document.body.addEventListener("showMessage", (event) => {
  /** @type {{message?: string, kind?: string}} */
  const detail = /** @type {CustomEvent} */ (event).detail || {};
  announce(detail.message || "", toastKindFor(detail.kind || ""));
});

els.keypad.querySelectorAll("button[data-tone]").forEach((node) => {
  const button = /** @type {HTMLButtonElement} */ (node);
  button.addEventListener("click", () => sendDtmf(button.dataset.tone || ""));
});

log(`webphone loaded; sip domain ${sipDomain}`);
