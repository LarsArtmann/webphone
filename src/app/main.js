// Wiring: event handlers, view switching, boot sequence. Everything
// else lives in its feature module — this file is the composition root.

import { ringToneStop } from "./audio.js";
import { connect, disconnect, getUserAgent } from "./connection.js";
import {
  bindSession,
  placeCall,
  renderCalls,
  sendDtmf,
  teardownAll,
  teardownSession,
} from "./calls.js";
import { sipDomain, websocketUrl } from "./config.js";
import { applyI18n, getLang, setLang, t } from "./i18n.js";
import {
  cancelVoicemailRefresh,
  refreshServerHistory,
  refreshVoicemail,
  renderContacts,
  renderHistory,
} from "./panels.js";
import { requestNotifications, titleFlashStop } from "./notify.js";
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
    if (getUserAgent())
      await getUserAgent()
        .stop()
        .catch(() => {});
  } finally {
    disconnect();
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

els.accept.addEventListener("click", async () => {
  const invitation = state.incomingSession;
  if (!invitation) return;
  els.incoming.hidden = true;
  state.incomingSession = null;
  ringToneStop();
  titleFlashStop();
  bindSession(invitation, els.incomingFrom.textContent);
  try {
    await invitation.accept({
      sessionDescriptionHandlerOptions: {
        constraints: { audio: true, video: false },
      },
    });
  } catch (err) {
    log(`accept failed: ${err.message}`, "error");
    announce(t("acceptFailed")(err.message), "error");
    teardownSession(invitation.id);
  }
});

els.reject.addEventListener("click", () => {
  if (state.incomingSession) state.incomingSession.reject();
  els.incoming.hidden = true;
  state.incomingSession = null;
  ringToneStop();
  titleFlashStop();
});

els.vmRefresh.addEventListener("click", () => refreshVoicemail());

els.keypad.querySelectorAll("button[data-tone]").forEach((button) => {
  button.addEventListener("click", () => sendDtmf(button.dataset.tone));
});

log(`webphone loaded; sip domain ${sipDomain}`);
