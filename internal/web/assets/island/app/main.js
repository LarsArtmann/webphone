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
  refreshServerHistory,
  refreshVoicemail,
  renderContacts,
  renderHistory,
} from "./panels.js";
import { requestNotifications, titleFlashStop } from "./notify.js";
import {
  createSession,
  destroySession,
  initSseLiveIndicator,
} from "./session.js";
import { initShortcuts } from "./shortcuts.js";
import { sessions, state } from "./state.js";
import { announce, els, log, setRegStatus } from "./ui.js";

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
    // language too (htmx.ajax keeps the island alive — no reload).
    const active = document.querySelector(".wp-nav .wp-nav-link.wp-active");
    if (active && window.htmx) {
      window.htmx.ajax("GET", active.getAttribute("hx-get"), {
        target: "#tab-content",
        swap: "innerHTML",
      });
    }
    log(`language switched to ${els.lang.value}`);
  });
}
applyI18n();

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
    // Phone-API backed panels: voicemail badge, server history, contacts.
    renderContacts();
    refreshVoicemail();
    refreshServerHistory();
  } catch (err) {
    els.loginError.textContent = t("loginError")(err.message);
    els.loginError.hidden = false;
    log(`connect failed: ${err.message}`, "error");
  }
});

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
// htmx dispatches it as a DOM event that bubbles to the body. The
// library's kind vocabulary maps onto the island's toast styles.
const TOAST_KINDS = { success: "ok", error: "error", warning: "warn", info: "info" };
document.body.addEventListener("showMessage", (event) => {
  const detail = event.detail || {};
  announce(detail.message || "", TOAST_KINDS[detail.kind] || "info");
});

els.keypad.querySelectorAll("button[data-tone]").forEach((button) => {
  button.addEventListener("click", () => sendDtmf(button.dataset.tone));
});

log(`webphone loaded; sip domain ${sipDomain}`);
