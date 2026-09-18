// Calls: the session table, call cards, hold/focus/mute, blind and
// attended transfer, DTMF, hangup and outgoing dialing.
//
// Transfer: FreeSWITCH executes the actual transfer server-side on the
// REFER (the REFERing party's partner leg is re-routed through the
// dialplan on blind; the two existing calls are bridged on attended via
// Replaces) — the browser only sends REFER and waits for the NOTIFY
// sipfrag verdict.

import { ringbackStart, ringbackStop } from "./audio.js";
import { getUserAgent } from "./connection.js";
import { sipDomain } from "./config.js";
import { startIcePanel, stopIcePanel } from "./ice.js";
import { t } from "./i18n.js";
import {
  recordHistory,
  refreshServerHistory,
  scheduleVoicemailRefresh,
} from "./panels.js";
import { ringToneStop } from "./audio.js";
import { titleFlashStop } from "./notify.js";
import { sessions, state } from "./state.js";
import { announce, els, log, showDialError } from "./ui.js";

function outgoingCount() {
  let n = 0;
  sessions.forEach(({ session }) => {
    if (
      session instanceof SIP.Inviter &&
      session.state === SIP.SessionState.Establishing
    )
      n++;
  });
  return n;
}

function attachRemoteAudio(id) {
  const entry = sessions.get(id);
  if (!entry || !entry.session.sessionDescriptionHandler) return;
  const pc = entry.session.sessionDescriptionHandler.peerConnection;
  const remoteStream = new MediaStream();
  pc.getReceivers().forEach((receiver) => {
    if (receiver.track) remoteStream.addTrack(receiver.track);
  });
  els.remoteAudio.srcObject = remoteStream;
  els.remoteAudio.play().catch((err) => {
    log(`audio playback blocked: ${err.message}`, "warn");
    announce(t("audioBlocked"), "warn");
  });
}

function setTracks(entry, { recv, send }) {
  const sdh = entry.session.sessionDescriptionHandler;
  if (!sdh) return;
  if (recv !== undefined && sdh.enableReceiverTracks)
    sdh.enableReceiverTracks(recv);
  if (send !== undefined && sdh.enableSenderTracks)
    sdh.enableSenderTracks(send);
}

async function holdSession(id, hold) {
  const entry = sessions.get(id);
  if (!entry || entry.session.state !== SIP.SessionState.Established) return;
  if (entry.held === hold) return;
  entry.held = hold; // preemptive; undone if the re-INVITE fails
  try {
    await entry.session.invite();
    setTracks(entry, { recv: !hold, send: !hold && !entry.muted });
  } catch (err) {
    entry.held = !hold;
    log(`hold toggle failed: ${err.message}`, "error");
    announce(t("holdFailed")(err.message), "error");
  }
  renderCalls();
}

function focusSession(id) {
  if (!sessions.has(id)) return;
  state.focusedId = id;
  sessions.forEach((entry, otherId) => {
    if (otherId === id) {
      if (entry.held) holdSession(otherId, false);
    } else if (
      entry.session.state === SIP.SessionState.Established &&
      !entry.held
    ) {
      holdSession(otherId, true);
    }
  });
  attachRemoteAudio(id);
  renderCalls();
}

export function teardownSession(id) {
  const entry = sessions.get(id);
  if (!entry) return;
  if (entry.timer) clearInterval(entry.timer);
  if (entry.dom) entry.dom.remove();
  sessions.delete(id);
  if (state.focusedId === id) {
    const next = sessions.keys().next();
    state.focusedId = next.done ? null : next.value;
    if (state.focusedId) attachRemoteAudio(state.focusedId);
  }
  renderCalls();
}

export function teardownAll() {
  [...sessions.keys()].forEach(teardownSession);
  state.focusedId = null;
}

function durationLabel(startedAt) {
  const sec = Math.max(0, Math.floor((Date.now() - startedAt) / 1000));
  return `${Math.floor(sec / 60)}:${String(sec % 60).padStart(2, "0")}`;
}

export function renderCalls() {
  sessions.forEach((entry, id) => {
    if (!entry.dom) return;
    const sessionState = entry.session.state;
    const stateEl = entry.dom.querySelector(".call-state-text");
    if (sessionState === SIP.SessionState.Established) {
      stateEl.textContent = entry.transferring
        ? t("transferring")
        : `${entry.held ? t("onHold") : t("inCall")} · ${durationLabel(entry.startedAt)}`;
    } else if (sessionState === SIP.SessionState.Establishing) {
      stateEl.textContent = t("ringing");
    } else if (
      sessionState === SIP.SessionState.Terminating ||
      sessionState === SIP.SessionState.Terminated
    ) {
      stateEl.textContent = t("ending");
    }
    entry.dom.classList.toggle("focused", id === state.focusedId);
    const holdBtn = entry.dom.querySelector(".hold-btn");
    holdBtn.textContent = entry.held ? t("resume") : t("hold");
    const muteBtn = entry.dom.querySelector(".mute-btn");
    muteBtn.textContent = entry.muted ? t("unmute") : t("mute");
    const focusBtn = entry.dom.querySelector(".focus-btn");
    if (focusBtn) focusBtn.textContent = t("focus");
    const endBtn = entry.dom.querySelector(".hangup-btn");
    if (endBtn) endBtn.textContent = t("end");
  });
  const established =
    state.focusedId &&
    sessions.get(state.focusedId) &&
    sessions.get(state.focusedId).session.state ===
      SIP.SessionState.Established;
  els.keypad.hidden = !established;
  els.iceWrap.hidden = sessions.size === 0;
  if (sessions.size > 0) startIcePanel();
  else stopIcePanel();
  if (outgoingCount() === 0) ringbackStop();
}

function addCallCard(id, target) {
  const card = document.createElement("div");
  card.className = "call-card";
  const head = document.createElement("div");
  head.className = "call-state";
  const targetEl = document.createElement("span");
  targetEl.className = "target";
  targetEl.textContent = target;
  const stateEl = document.createElement("span");
  stateEl.className = "call-state-text";
  stateEl.textContent = t("calling");
  head.append(stateEl, targetEl);
  const controls = document.createElement("div");
  controls.className = "controls";
  const mkBtn = (label, cls, onClick) => {
    const b = document.createElement("button");
    b.textContent = label;
    if (cls) b.className = cls;
    b.addEventListener("click", onClick);
    return b;
  };
  controls.append(
    mkBtn(t("focus"), "ghost focus-btn", () => focusSession(id)),
    mkBtn(t("mute"), "mute-btn", () => {
      const entry = sessions.get(id);
      if (!entry) return;
      entry.muted = !entry.muted;
      setTracks(entry, { send: !entry.held && !entry.muted });
      renderCalls();
    }),
    mkBtn(t("hold"), "hold-btn", () => {
      const entry = sessions.get(id);
      if (entry) holdSession(id, !entry.held);
    }),
    mkBtn(t("transfer"), "ghost transfer-btn", () =>
      toggleTransferRow(card, id),
    ),
    mkBtn(t("end"), "danger hangup-btn", () => hangup(id)),
  );
  card.append(head, controls);
  els.calls.append(card);
  return card;
}

// Transfer row: inline destination input with blind/attended actions.
function toggleTransferRow(card, id) {
  const existing = card.querySelector(".transfer-row");
  if (existing) {
    existing.remove();
    return;
  }
  const row = document.createElement("div");
  row.className = "transfer-row";
  const input = document.createElement("input");
  input.inputMode = "tel";
  input.placeholder = t("transferPrompt");
  input.className = "transfer-dest";
  const blind = document.createElement("button");
  blind.className = "ghost small";
  blind.textContent = t("transferBlind");
  const attended = document.createElement("button");
  attended.className = "ghost small";
  attended.textContent = t("transferAttended");
  blind.addEventListener("click", () => {
    const dest = input.value.trim();
    if (dest) blindTransfer(id, dest);
  });
  attended.addEventListener("click", () => attendedTransfer(id));
  row.append(input, blind, attended);
  card.append(row);
  input.focus();
}

function referOnNotify(notification) {
  const body = (notification.request && notification.request.body) || "";
  const status = body.match(/^SIP\/2\.0 (\d{3})/m);
  notification.accept().catch(() => {});
  if (status && /^2/.test(status[1])) {
    log(t("transferComplete"));
    announce(t("transferComplete"), "ok");
    return;
  }
  const detail = status ? status[1] : "no final NOTIFY";
  log(t("transferFailed")(detail), "error");
  announce(t("transferFailed")(detail), "error");
}

async function blindTransfer(id, destination) {
  const entry = sessions.get(id);
  if (!entry || entry.session.state !== SIP.SessionState.Established) return;
  const target = destination.replace(/[^\d+*#]/g, "");
  const uri = SIP.UserAgent.makeURI(`sip:${target}@${sipDomain}`);
  if (!uri) {
    log(t("transferFailed")("bad destination"), "error");
    announce(t("transferFailed")("bad destination"), "error");
    return;
  }
  entry.transferring = true;
  renderCalls();
  try {
    await entry.session.refer(uri, { onNotify: referOnNotify });
    log(`blind transfer ${entry.target} → ${target} sent`);
  } catch (err) {
    entry.transferring = false;
    log(t("transferFailed")(err.message), "error");
    announce(t("transferFailed")(err.message), "error");
    renderCalls();
  }
}

async function attendedTransfer(id) {
  const entry = sessions.get(id);
  if (!entry || entry.session.state !== SIP.SessionState.Established) return;
  // The other ESTABLISHED call is the consult leg; Replaces points at
  // its dialog so the network bridges the two far ends.
  let partner = null;
  sessions.forEach((other, otherId) => {
    if (
      otherId !== id &&
      other.session.state === SIP.SessionState.Established &&
      !other.transferring
    )
      partner = other;
  });
  if (!partner) {
    log(t("transferNoPartner"), "warn");
    announce(t("transferNoPartner"), "warn");
    return;
  }
  entry.transferring = true;
  renderCalls();
  try {
    await entry.session.refer(partner.session, { onNotify: referOnNotify });
    log(`attended transfer ${entry.target} ↔ ${partner.target} sent`);
  } catch (err) {
    entry.transferring = false;
    log(t("transferFailed")(err.message), "error");
    announce(t("transferFailed")(err.message), "error");
    renderCalls();
  }
}

// Diagnostic handle: the E2E suite reads getStats() from these to
// prove real media flows (bytes on the wire), not just signaling.
window.__pcs = window.__pcs || new Map();

export function bindSession(newSession, target) {
  newSession.stateChange.addListener((sessionState) => {
    if (sessionState === SIP.SessionState.Established) {
      const pc =
        newSession.sessionDescriptionHandler &&
        newSession.sessionDescriptionHandler.peerConnection;
      if (pc) window.__pcs.set(newSession.id, pc);
    }
  });
  const id = newSession.id;
  const entry = {
    session: newSession,
    target,
    held: false,
    muted: false,
    established: false,
    startedAt: Date.now(),
    timer: null,
    dom: addCallCard(id, target),
  };
  sessions.set(id, entry);
  state.focusedId = state.focusedId || id;
  if (newSession instanceof SIP.Inviter) ringbackStart();

  newSession.stateChange.addListener((sessionState) => {
    const live = sessions.get(id);
    if (!live) return;
    log(`call ${target} ${sessionState}`);
    if (sessionState === SIP.SessionState.Established) {
      live.established = true;
      live.startedAt = Date.now();
      if (!live.timer) live.timer = setInterval(renderCalls, 1000);
      focusSession(id);
      // Call is no longer ringing: the incoming UX stops either way.
      titleFlashStop();
      ringToneStop();
    } else if (sessionState === SIP.SessionState.Terminated) {
      const dur = Math.floor((Date.now() - live.startedAt) / 1000);
      if (live.established) {
        announce(t("callEnded")(durationLabel(live.startedAt)));
      }
      recordHistory({
        dir: newSession instanceof SIP.Inviter ? "out" : "in",
        target,
        at: Date.now(),
        dur: dur > 0 ? dur : 0,
      });
      teardownSession(id);
      // A just-ended call may have left a voicemail deposit.
      scheduleVoicemailRefresh();
      refreshServerHistory();
    }
    renderCalls();
  });
  renderCalls();
}

export async function hangup(id) {
  const entry = sessions.get(id);
  if (!entry) return;
  const current = entry.session;
  try {
    if (
      current instanceof SIP.Inviter &&
      current.state === SIP.SessionState.Initial
    ) {
      await current.cancel();
    } else if (current.state === SIP.SessionState.Established) {
      await current.bye();
    } else if (current instanceof SIP.Inviter) {
      await current.cancel();
    } else {
      await current.reject();
    }
  } catch (err) {
    log(`hangup: ${err.message}`, "error");
    announce(t("hangupFailed")(err.message), "error");
    teardownSession(id);
  }
}

export function sendDtmf(tone) {
  const entry = state.focusedId && sessions.get(state.focusedId);
  if (!entry || entry.session.state !== SIP.SessionState.Established) {
    announce(t("noActiveCall"), "info");
    return;
  }
  // application/dtmf-relay with "Signal=<d>" (equals): that is the
  // only form mod_sofia parses, and only with the profile flag
  // extended-info-parsing enabled (the generated profiles set it).
  // The colon form is 200-OK'd and silently dropped.
  const body = {
    contentDisposition: "render",
    contentType: "application/dtmf-relay",
    content: `Signal=${tone}\r\nDuration: 2000`,
  };
  entry.session
    .info({ requestOptions: { body } })
    .then(() => log(`dtmf ${tone}`))
    .catch((err) => {
      log(`dtmf failed: ${err.message}`, "error");
      announce(t("dtmfFailed")(err.message), "error");
    });
}

export async function placeCall(raw) {
  const userAgent = getUserAgent();
  if (!userAgent) {
    announce(t("notConnected"), "error");
    return;
  }
  if (!raw) {
    showDialError(t("dialEmpty"));
    return;
  }

  // Pasted numbers routinely carry invisible Unicode direction marks
  // (macOS/phone apps add them around telephone numbers) and formatting
  // (spaces, dashes, parentheses). makeURI rejects all of that, so strip
  // everything that is not dialable before building the SIP URI.
  const target = raw.replace(/[^\d+*#]/g, "");
  if (!target) {
    log(`nothing dialable in "${raw}" — enter digits, or + * #`, "warn");
    showDialError(t("nothingDialable"));
    return;
  }

  const uri = SIP.UserAgent.makeURI(`sip:${target}@${sipDomain}`);
  if (!uri) {
    log(`invalid destination "${target}"`, "warn");
    showDialError(t("invalidDest"));
    return;
  }

  const inviter = new SIP.Inviter(userAgent, uri, {
    sessionDescriptionHandlerOptions: {
      constraints: { audio: true, video: false },
    },
  });
  bindSession(inviter, target);
  try {
    await inviter.invite();
  } catch (err) {
    log(`invite failed: ${err.message}`, "error");
    showDialError(t("callFailed")(err.message));
    teardownSession(inviter.id);
  }
}
