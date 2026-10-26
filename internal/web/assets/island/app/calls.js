// Calls: the session table, call cards, hold/focus/mute, blind and
// attended transfer, DTMF, hangup and outgoing dialing.
//
// Transfer: FreeSWITCH executes the actual transfer server-side on the
// REFER (the REFERing party's partner leg is re-routed through the
// dialplan on blind; the two existing calls are bridged on attended via
// Replaces) — the browser only sends REFER and waits for the NOTIFY
// sipfrag verdict.

import {
  resumeAudio,
  ringbackStart,
  ringbackStop,
  ringToneStop,
} from "./audio.js";
import { asrEnabled, sipDomain } from "./config.js";
import { getLang, t } from "./i18n.js";
import { releaseWarmMic, warmMic } from "./mic.js";
import {
  recordCrmCall,
  recordHistory,
  refreshServerHistory,
  scheduleVoicemailRefresh,
} from "./panels.js";
import { titleFlashStop } from "./notify.js";
import { applyTransition, sessions, state, viewStateOf } from "./state.js";
import {
  isTranscribing,
  pauseLiveTranscription,
  saveTranscriptSegment,
  startLiveTranscription,
  stopLiveTranscription,
} from "./transcribe.js";
import {
  announce,
  els,
  hideIncomingBanner,
  log,
  renderIncomingSlots,
  setIncomingSlotActions,
  showDialError,
} from "./ui.js";

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
  // Per-call volume (the dock's slider): the focused call's remembered
  // level rides every (re)attach — the shared element plays exactly one
  // call at a time (held calls carry no receiver audio).
  els.remoteAudio.volume = entry.volume ?? 1;
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
  // Mid-flight toggle: queue the latest wish (the honesty contract —
  // a focus preemption must never be lost), BEFORE the table gate.
  if (entry.holdPending) {
    entry.holdQueued = hold;
    return;
  }
  // The declarative gate (CALL_TRANSITIONS): asking for the state you
  // already have (hold from active-hold, resume from active) is simply
  // illegal in the table — one rule, no re-derived guards. A genuinely
  // illegal move from any other state would land here too: logged and
  // kept, never rendered.
  const to = hold ? "holding" : "resuming";
  const move = applyTransition(viewStateOf(entry), to);
  if (!move.changed) return;
  entry.holdPending = hold ? "holding" : "resuming";
  renderCalls();
  try {
    await entry.session.invite();
    entry.held = hold;
    setTracks(entry, { recv: !hold, send: !hold && !entry.muted });
    // A held call carries no audio either way — pause the capture so
    // the provider never receives (and hallucinates on) silence.
    pauseLiveTranscription(id, hold);
  } catch (err) {
    log(`hold toggle failed: ${err.message}`, "error");
    announce(t(hold ? "holdFailed" : "resumeFailed")(err.message), "error");
  } finally {
    entry.holdPending = null;
    const queued = entry.holdQueued;
    entry.holdQueued = undefined;
    if (
      queued !== undefined &&
      queued !== entry.held &&
      entry.session.state === SIP.SessionState.Established
    ) {
      holdSession(id, queued);
    } else {
      renderCalls();
    }
  }
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
  stopLiveTranscription(id);
  // The call's undelivered transcript segments get one last flush
  // attempt before the entry goes away (the queue drains in the
  // background; a still-failing save stays logged, never blocking).
  flushTranscriptSaves(entry);
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

// liveStateLine derives the card/chip state line for one entry — the
// ONE derivation both surfaces read (the dock chip and the expanded
// card can never disagree about what a call is doing).
function liveStateLine(entry) {
  const sessionState = entry.session.state;
  if (sessionState === SIP.SessionState.Established) {
    const dataState =
      entry.holdPending || (entry.held ? "held" : "established");
    const text = entry.transferring
      ? t("transferring")
      : entry.holdPending
        ? t(entry.holdPending === "holding" ? "callHolding" : "callResuming")
        : `${entry.held ? t("onHold") : t("inCall")} · ${durationLabel(entry.startedAt)}`;
    return { dataState, text };
  }
  if (sessionState === SIP.SessionState.Establishing) {
    return { dataState: "ringing", text: t("ringing") };
  }
  return { dataState: "ending", text: t("ending") };
}

// --- the call dock (multi-call train) ---------------------------------------
//
// More than one live call collapses the cards into a dock: one chip per
// call (stable id wp-dock-chip-<callId>) plus ONE expanded card for the
// focused call. The chip click IS hold & take — focusSession holds the
// current call and resumes the target through the existing
// holdPending/holdQueued pending machine, so the swap stays honest (the
// UI shows the work; a mid-flight toggle queues, never lies).

const dockChips = new Map();
let dockEl = null;

function renderCallDock() {
  if (sessions.size < 2) {
    dockChips.forEach((chip) => chip.remove());
    dockChips.clear();
    if (dockEl) {
      dockEl.remove();
      dockEl = null;
    }
    return;
  }
  if (!dockEl) dockEl = createCallDock();
  dockChips.forEach((chip, id) => {
    if (!sessions.has(id)) {
      chip.remove();
      dockChips.delete(id);
    }
  });
  sessions.forEach((entry, id) => {
    let chip = dockChips.get(id);
    if (!chip) {
      chip = buildDockChip(id, entry);
      dockChips.set(id, chip);
      dockEl.append(chip);
    }
    updateDockChip(chip, id, entry);
  });
}

function createCallDock() {
  const dock = document.createElement("div");
  dock.id = "wp-call-dock";
  dock.className = "wp-call-dock";
  dock.setAttribute("role", "toolbar");
  dock.setAttribute("aria-label", t("dockLabel"));
  els.calls.prepend(dock);
  return dock;
}

function buildDockChip(id, entry) {
  const chip = document.createElement("div");
  chip.className = "wp-dock-chip";
  chip.id = `wp-dock-chip-${id}`;
  chip.dataset.callId = id;
  chip.tabIndex = 0;
  chip.append(chipAvatar(entry.target), chipName(entry.target));
  const pill = document.createElement("span");
  pill.className = "wp-chip-pill";
  chip.append(pill);
  // The whole chip is the one-gesture hold & take: focus (hold current,
  // resume this one) lands where the user clicked.
  const take = () => focusSession(id);
  chip.addEventListener("click", take);
  chip.addEventListener("keydown", (event) => {
    if (event.key === "Enter" || event.key === " ") {
      event.preventDefault();
      take();
    }
  });
  return chip;
}

function updateDockChip(chip, id, entry) {
  const { dataState, text } = liveStateLine(entry);
  chip.dataset.state = dataState;
  chip.classList.toggle("focused", id === state.focusedId);
  const pill = chip.children[2];
  pill.textContent = text;
}

// chipAvatar mirrors the server's desk-phone vernacular: a hue-bucketed
// disc showing the number's LAST TWO DIGITS (never a country signum).
// Classes, never inline styles — the strict CSP silently drops those.
function chipAvatar(target) {
  const digits = String(target).replace(/\D/g, "");
  const label =
    digits.length >= 2 ? digits.slice(-2) : String(target).slice(-2) || "·";
  let sum = 0;
  for (const ch of String(target)) sum = (sum * 31 + ch.charCodeAt(0)) % 360;
  const av = document.createElement("span");
  av.className = `wp-chip-av h${Math.floor(sum / 30) * 30}`;
  av.textContent = label;
  av.setAttribute("aria-hidden", "true");
  return av;
}

function chipName(target) {
  const name = document.createElement("span");
  name.className = "wp-chip-name";
  name.textContent = target;
  return name;
}

export function renderCalls() {
  renderCallDock();
  sessions.forEach((entry, id) => {
    if (!entry.dom) return;
    const { dataState, text } = liveStateLine(entry);
    const stateEl = entry.dom.querySelector(".call-state-text");
    stateEl.textContent = text;
    entry.dom.dataset.state = dataState;
    announceCallState(entry, dataState);
    entry.dom.classList.toggle("focused", id === state.focusedId);
    // Dock mode: the chips carry the other calls; only the focused call
    // keeps its expanded card. A single call renders the plain card.
    entry.dom.hidden = sessions.size > 1 && id !== state.focusedId;
    const holdBtn = entry.dom.querySelector(".hold-btn");
    holdBtn.disabled = Boolean(entry.holdPending);
    holdBtn.textContent = entry.holdPending
      ? t(entry.holdPending === "holding" ? "callHolding" : "callResuming")
      : entry.held
        ? t("resume")
        : t("hold");
    const muteBtn = entry.dom.querySelector(".mute-btn");
    muteBtn.textContent = entry.muted ? t("unmute") : t("mute");
    const focusBtn = entry.dom.querySelector(".focus-btn");
    if (focusBtn) focusBtn.textContent = t("focus");
    const endBtn = entry.dom.querySelector(".hangup-btn");
    if (endBtn) endBtn.textContent = t("end");
    const slider = entry.dom.querySelector(".volume-slider");
    if (slider) slider.value = String(Math.round((entry.volume ?? 1) * 100));
    const transcribeBtn = entry.dom.querySelector(".transcribe-btn");
    if (transcribeBtn)
      transcribeBtn.textContent = isTranscribing(id)
        ? t("transcribeStop")
        : t("transcribe");
    const savedEl = entry.dom.querySelector(".call-transcript-saved");
    if (savedEl && entry.transcriptSaved) {
      savedEl.textContent = t("transcriptSavedToHistory")(
        entry.transcriptSaved,
      );
    }
  });
  const established =
    state.focusedId &&
    sessions.get(state.focusedId) &&
    sessions.get(state.focusedId).session.state ===
      SIP.SessionState.Established;
  els.keypad.hidden = !established;
  // ice.js listens for this (calls and ice never import each other —
  // the module-graph invariant the arch test enforces).
  document.dispatchEvent(new CustomEvent("wp:calls-changed"));
  if (outgoingCount() === 0) ringbackStop();
}

// announceCallState speaks only state TRANSITIONS through the toast
// live region: the state text itself swaps silently, and once
// established it ticks once per second — announcing those would be
// noise, not signal (screen readers get ringing/connected/ended as
// discrete events instead).
function announceCallState(entry, dataState) {
  if (entry.announcedState === dataState) return;
  entry.announcedState = dataState;
  if (dataState === "ringing") {
    announce(t("callRinging")(entry.target), "info");
  } else if (dataState === "established") {
    announce(t("callEstablished")(entry.target), "ok");
  } else if (dataState === "ending") {
    announce(t("callEnded")(entry.target), "info");
  }
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
  // Per-call volume (multi-call train): the focused call's slider
  // drives the shared remote-audio element directly and is remembered
  // per entry, so switching back to a call restores its level.
  const volume = document.createElement("input");
  volume.type = "range";
  volume.className = "volume-slider";
  volume.min = "0";
  volume.max = "100";
  volume.value = "100";
  volume.setAttribute("aria-label", t("volume"));
  volume.addEventListener("input", () => {
    const entry = sessions.get(id);
    if (!entry) return;
    entry.volume = Number(volume.value) / 100;
    els.remoteAudio.volume = entry.volume;
  });
  controls.append(volume);
  // Live transcription (only when the ASR seam is configured): a toggle
  // that captures the call audio and appends each segment's text below.
  const transcript = document.createElement("div");
  transcript.className = "call-transcript";
  transcript.id = `call-transcript-${id}`;
  if (asrEnabled) {
    controls.append(
      mkBtn(t("transcribe"), "ghost transcribe-btn", () =>
        toggleTranscription(id, transcript),
      ),
      mkBtn(t("copyTranscript"), "ghost copy-transcript-btn", () =>
        copyCardTranscript(transcript),
      ),
      mkBtn(t("clearTranscript"), "ghost clear-transcript-btn", () =>
        clearCardTranscript(transcript),
      ),
    );
  }
  card.append(head, controls, transcript);
  if (asrEnabled) {
    const saved = document.createElement("div");
    saved.className = "call-transcript-saved";
    saved.hidden = true;
    card.append(saved);
  }
  els.calls.append(card);
  return card;
}

// toggleTranscription starts or stops live-call transcription for one
// call. Starting reports each segment's text through onCallTranscript;
// when the browser cannot capture (no MediaRecorder/AudioContext, or no
// media yet) the transcript shows an honest "unsupported" notice instead
// of a silent dead button.
function toggleTranscription(id, target) {
  const entry = sessions.get(id);
  if (!entry) return;
  if (isTranscribing(id)) {
    stopLiveTranscription(id);
    announce(t("transcribeStopped"), "info");
  } else if (!startCallTranscription(id, entry, target)) {
    target.textContent = t("transcribeUnsupported");
  }
  renderCalls();
}

// startCallTranscription is the ONE start path (auto-start on connect
// and the manual toggle both ride it): it wires segment text to
// onCallTranscript and passes the UI language as the provider hint.
function startCallTranscription(id, entry, target) {
  return startLiveTranscription(
    id,
    entry,
    (text) => onCallTranscript(id, entry, target, text),
    {
      language: getLang(),
    },
  );
}

// ensureTranscription is the auto-start hook: transcription begins by
// itself when the ASR seam is on (owner decision 2026-10-07) — the
// button remains the honest manual override. An honest notice replaces
// a silent failure to start.
function ensureTranscription(id, entry) {
  if (!asrEnabled) return;
  const target = entry.dom && entry.dom.querySelector(".call-transcript");
  if (!target || isTranscribing(id)) return;
  if (startCallTranscription(id, entry, target)) {
    announce(t("transcribeStarted"), "info");
    renderCalls();
  } else {
    target.textContent = t("transcribeUnsupported");
  }
}

// onCallTranscript is the per-segment sink: the card grows, screen
// readers hear the delta (polite — this IS the product for them), and
// the segment goes through the per-call save queue (a failed save costs
// the LATER re-read, never the live card).
function onCallTranscript(id, entry, target, text) {
  appendTranscript(target, text);
  announce(clipForAnnounce(text), "info");
  queueTranscriptSave(entry, {
    callId: id,
    direction: entry.session instanceof SIP.Inviter ? "out" : "in",
    remote: entry.target,
    startedAt: entry.startedAt,
    text,
  });
}

// The save queue bounds what a flaky network can eat: failed segments
// stay queued (order preserved, cap 50 per call, drop-oldest) and flush
// one-at-a-time on the next successful save — and once more when the
// call ends. No timer: the queue is honest about being best-effort, and
// the once-per-call warn toast keeps the operator informed without
// nagging at the ~15 segments/minute cadence.
const TRANSCRIPT_SAVE_QUEUE_CAP = 50;

function queueTranscriptSave(entry, segment) {
  entry.transcriptSaveQueue ??= [];
  entry.transcriptSaveQueue.push(segment);
  if (entry.transcriptSaveQueue.length > TRANSCRIPT_SAVE_QUEUE_CAP) {
    entry.transcriptSaveQueue.shift();
  }
  flushTranscriptSaves(entry);
}

function flushTranscriptSaves(entry) {
  if (entry.transcriptSaveFlushing || !entry.transcriptSaveQueue?.length)
    return;
  entry.transcriptSaveFlushing = true;
  saveTranscriptSegment(entry.transcriptSaveQueue[0])
    .then(() => {
      entry.transcriptSaveQueue.shift();
      entry.transcriptSaved = (entry.transcriptSaved || 0) + 1;
      entry.transcriptSaveFlushing = false;
      showTranscriptSaved(entry);
      flushTranscriptSaves(entry);
    })
    .catch((err) => {
      entry.transcriptSaveFlushing = false;
      log(`transcript save failed: ${err.message}`, "warn");
      if (!entry.transcriptSaveWarned) {
        entry.transcriptSaveWarned = true;
        announce(t("transcriptSaveFailed")(err.message), "warn");
      }
    });
}

// showTranscriptSaved is the #13 discoverability affordance: a CONFIRMED
// save surfaces a persistent "saved to History" count on the card and
// announces once per call — the live transcript must not read as a
// card-only artifact that vanishes on hangup.
function showTranscriptSaved(entry) {
  const marker = entry.dom && entry.dom.querySelector(".call-transcript-saved");
  if (marker) {
    marker.textContent = t("transcriptSavedToHistory")(entry.transcriptSaved);
    marker.hidden = false;
  }
  if (!entry.transcriptSavedAnnounced) {
    entry.transcriptSavedAnnounced = true;
    announce(t("transcriptSavedToHistory")(entry.transcriptSaved), "info");
  }
}

// clipForAnnounce bounds what the live region reads per segment: a
// full paragraph would jam the SR queue at the ~15 segments/minute cadence.
function clipForAnnounce(text) {
  const clipped = text.length > 140 ? `${text.slice(0, 139)}…` : text;
  return t("transcriptDelta")(clipped);
}

// The live card's DOM cap (#41): a marathon call must not grow an
// unbounded transcript node. The kept tail is a few screens of live
// context; the durable whole lives server-side (History) regardless.
const TRANSCRIPT_DOM_CAP = 8000;

function appendTranscript(target, text) {
  if (!text) return;
  let next = target.textContent ? `${target.textContent} ${text}` : text;
  if (next.length > TRANSCRIPT_DOM_CAP) {
    next = `… ${next.slice(next.length - TRANSCRIPT_DOM_CAP)}`;
  }
  target.textContent = next;
  target.scrollTop = target.scrollHeight;
}

// clearCardTranscript is the per-call clear affordance (#39): the live
// card is presentation only — clearing it never touches the stored
// History transcript, and the announce says exactly that scope.
function clearCardTranscript(target) {
  if (!target.textContent) {
    announce(t("transcribeEmpty"), "warn");
    return;
  }
  target.textContent = "";
  announce(t("transcriptCleared"), "info");
}

// copyCardTranscript puts the card's accumulated transcript on the
// clipboard (secure context — the stack fronts TLS). Honest feedback
// either way: an empty transcript announces that instead of copying
// nothing, a missing clipboard API announces the manual fallback.
function copyCardTranscript(target) {
  const text = (target.textContent || "").trim();
  if (!text) {
    announce(t("transcribeEmpty"), "warn");
    return;
  }
  const clip = navigator.clipboard;
  if (!clip || !clip.writeText) {
    announce(t("copyFailed"), "warn");
    return;
  }
  clip.writeText(text).then(
    () => announce(t("copied"), "info"),
    () => announce(t("copyFailed"), "warn"),
  );
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
  const target = destination.replace(/[^\d+*#a-zA-Z]/g, "");
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
    holdPending: null,
    holdQueued: undefined,
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
    // Emission through the declarative table (D-train): every sip.js
    // state arrival maps onto a view state and is legality-checked
    // against CALL_TRANSITIONS. A refused arrival (Established twice,
    // traffic after ended) is a wiring bug — logged loudly here, the
    // view-state derivation below stays single-sourced either way.
    const sipToView = {
      [SIP.SessionState.Established]: "active",
      [SIP.SessionState.Terminating]: "ending",
      [SIP.SessionState.Terminated]: "ended",
    };
    const proposed = sipToView[sessionState];
    if (proposed) {
      const from = viewStateOf(live);
      if (from !== proposed) {
        const move = applyTransition(from, proposed);
        if (!move.changed) {
          log(
            `call ${target}: refused illegal view transition ${from} -> ${proposed} (${sessionState})`,
            "warn",
          );
        }
      }
    }
    if (sessionState === SIP.SessionState.Established) {
      live.established = true;
      live.startedAt = Date.now();
      if (!live.timer) live.timer = setInterval(renderCalls, 1000);
      focusSession(id);
      // Auto-start (owner decision 2026-10-07): with the ASR seam on,
      // transcription begins with the call's media — the transcribe
      // button stays as the manual stop/restart override.
      ensureTranscription(id, live);
      // Call is no longer ringing: the incoming UX stops either way.
      titleFlashStop();
      ringToneStop();
    } else if (sessionState === SIP.SessionState.Terminated) {
      // A dead session settles its hold state machine: the re-INVITE
      // answer never arrives, so no pending chip may outlive the call
      // (the watchdog-rebuild teardown lands here too).
      live.holdPending = null;
      live.holdQueued = undefined;
      const dur = Math.floor((Date.now() - live.startedAt) / 1000);
      if (live.established) {
        announce(t("callEnded")(durationLabel(live.startedAt)));
      }
      if (!(newSession instanceof SIP.Inviter) && !live.established) {
        // Inbound call that never carried media: the caller gave up
        // before the user answered, or the accepted call died in setup.
        // shell.js mirrors this into the missed-call header badge; a
        // user REJECT never reaches this path (rejected calls are seen
        // calls).
        document.dispatchEvent(
          new CustomEvent("wp:call-missed", { detail: { target } }),
        );
      }
      if (!(newSession instanceof SIP.Inviter)) releaseWarmMic();
      recordHistory({
        dir: newSession instanceof SIP.Inviter ? "out" : "in",
        target,
        at: Date.now(),
        dur: dur > 0 ? dur : 0,
      });
      // Fire-and-forget CRM report: enrichment for the journal, never a
      // call-path dependency — a dead CRM must not delay teardown.
      recordCrmCall({
        dir: newSession instanceof SIP.Inviter ? "out" : "in",
        target,
        dur: dur > 0 ? dur : 0,
        established: live.established,
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

// DTMF feedback (A8): the pressed key pulses — on INVOCATION, not
// delivery, so keyboard-triggered tones flash too and the press is
// acknowledged instantly; the #log line records the verdict.
function flashKey(tone) {
  const key = els.keypad.querySelector(`button[data-tone="${tone}"]`);
  if (!key || !key.classList) return;
  key.classList.remove("wp-key-sent");
  void key.offsetWidth; // restart the animation on rapid repeats
  key.classList.add("wp-key-sent");
}

// The transient tones-sent trail under the keypad: the last ≤12 digits,
// fading to hidden after 3 s of silence. #log stays the durable record;
// this is the in-the-moment confirmation for mid-call menus.
let tonesSent = [];
let tonesTimer = null;

function noteToneSent(tone) {
  let trail = document.getElementById("wp-tones");
  if (!trail) {
    trail = document.createElement("p");
    trail.id = "wp-tones";
    trail.className = "wp-tones";
    trail.setAttribute("aria-hidden", "true");
    els.keypad.append(trail);
  }
  tonesSent.push(tone);
  trail.textContent = tonesSent.slice(-12).join("·");
  trail.hidden = false;
  if (tonesTimer) clearTimeout(tonesTimer);
  tonesTimer = setTimeout(() => {
    trail.hidden = true;
    trail.textContent = "";
    tonesSent = [];
  }, 3000);
}

export function sendDtmf(tone) {
  const entry = state.focusedId && sessions.get(state.focusedId);
  if (!entry || entry.session.state !== SIP.SessionState.Established) {
    announce(t("noActiveCall"), "info");
    return;
  }
  flashKey(tone);
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
    .then(() => {
      log(`dtmf ${tone}`);
      noteToneSent(tone);
    })
    .catch((err) => {
      log(`dtmf failed: ${err.message}`, "error");
      announce(t("dtmfFailed")(err.message), "error");
    });
}

// sanitizeDialable strips everything that is not dialable: pasted
// numbers routinely carry invisible Unicode direction marks (macOS/phone
// apps add them around telephone numbers) and formatting (spaces,
// dashes, parentheses), and makeURI rejects all of it. ONE home for the
// charset — placeCall dials it, initDialHint previews it.
export function sanitizeDialable(raw) {
  return raw.replace(/[^\d+*#a-zA-Z]/g, "");
}

// Input-time normalization preview (A5): shows exactly what placeCall
// would dial, while the number is still being typed — a pasted
// "+49 (30) 1234-56" reveals its dialable form before the Call button
// commits to it. Visual aid only (aria-hidden): screen readers already
// hear the input, and dial errors announce themselves.
export function initDialHint() {
  els.dest.addEventListener("input", () => {
    const raw = els.dest.value;
    const dialable = sanitizeDialable(raw);
    if (dialable && dialable !== raw.trim()) {
      els.dialHint.textContent = `→ ${dialable}`;
      els.dialHint.hidden = false;
    } else {
      els.dialHint.textContent = "";
      els.dialHint.hidden = true;
    }
  });
}

// Dial-focus mic warm (the outgoing mirror of the incoming onInvite
// warm): the first focus in the dial field IS dial intent — start the
// getUserMedia acquisition while the user is still typing, so the
// INVITE's session-description handler finds a warm stream instead of
// paying the device-open latency on the call path. The TTL bounds the
// intent: an abandoned dial must not keep the mic indicator lit.
// Session-open is deliberately NOT a trigger — lighting the mic right
// after login expresses no call intent at all.
const DIAL_WARM_TTL_MS = 45_000;

export function initDialWarm() {
  els.dest.addEventListener("focus", () => warmMic(DIAL_WARM_TTL_MS));
}

// placeCall returns whether an INVITE actually went out, so the dial
// form can clear itself ONLY then (a validation error keeps the typed
// text for editing).
export async function placeCall(raw) {
  // The dial click is a user gesture: unlock the shared AudioContext so
  // the ringback is audible (audio.js).
  resumeAudio().catch(() => {});
  const userAgent = state.userAgent;
  if (!userAgent) {
    announce(t("notConnected"), "error");
    return false;
  }
  if (!raw) {
    showDialError(t("dialEmpty"));
    return false;
  }

  const target = sanitizeDialable(raw);
  if (!target) {
    log(
      `nothing dialable in "${raw}" — enter digits or letters, or + * #`,
      "warn",
    );
    showDialError(t("nothingDialable"));
    return false;
  }

  const uri = SIP.UserAgent.makeURI(`sip:${target}@${sipDomain}`);
  if (!uri) {
    log(`invalid destination "${target}"`, "warn");
    showDialError(t("invalidDest"));
    return false;
  }

  const inviter = new SIP.Inviter(userAgent, uri, {
    sessionDescriptionHandlerOptions: {
      constraints: { audio: true, video: false },
    },
  });
  bindSession(inviter, target);
  try {
    await inviter.invite();
    return true;
  } catch (err) {
    log(`invite failed: ${err.message}`, "error");
    showDialError(t("callFailed")(err.message));
    teardownSession(inviter.id);
    return false;
  }
}

// Keyboard-shortcut surface (shortcuts.js drives these; the call-card
// buttons and the incoming-call banner share the same primitives).

export function toggleMuteFocused() {
  const entry = sessions.get(state.focusedId);
  if (!entry) return;
  entry.muted = !entry.muted;
  setTracks(entry, { send: !entry.held && !entry.muted });
  renderCalls();
}

export function toggleHoldFocused() {
  const entry = sessions.get(state.focusedId);
  if (entry) holdSession(state.focusedId, !entry.held);
}

export function hangupFocused() {
  if (state.focusedId) hangup(state.focusedId);
}

// cycleFocus walks the ESTABLISHED calls ([ and ]) — ringing-only
// sessions are skipped (a ring is answered or rejected, not focused).
// The move announces politely with the remote party's name so a screen
// reader user knows whose voice just came up.
export function cycleFocus(step) {
  const established = [...sessions.entries()].filter(
    ([, entry]) => entry.session.state === SIP.SessionState.Established,
  );
  if (established.length < 2) return;
  const at = established.findIndex(([id]) => id === state.focusedId);
  const next =
    established[(Math.max(at, 0) + step + established.length) % established.length];
  focusSession(next[0]);
  announce(t("focusChanged")(next[1].target), "info");
}

// The ringing queue's answer/reject primitives. The no-arg answer pops
// the OLDEST ring (the A key and the pinned Accept button land here);
// the no-arg reject takes the NEWEST ring (Esc) — the freshest
// interruption clears first and the focused live call is never touched.
// An explicit INVITE (the banner stack's per-slot buttons) targets that
// exact slot, so a button can never reject a different caller than the
// one it labels.
function takeIncomingEntry(target) {
  const queue = state.incomingSessions;
  if (!queue.length) return null;
  let at;
  if (target === undefined) at = 0;
  else if (typeof target === "number") at = target;
  else at = queue.findIndex((entry) => entry.invitation === target);
  if (at < 0 || at >= queue.length) return null;
  return queue.splice(at, 1)[0];
}

// After one slot leaves the queue: remaining rings re-render the stack
// and keep ringing; the LAST one out hides the banner and stands the
// ring UX down (the warm mic stays for an answer, is released on a
// reject-to-empty by the reject path).
function settleIncomingAfterRemoval() {
  if (state.incomingSessions.length > 0) {
    renderIncomingSlots();
    return;
  }
  hideIncomingBanner();
  ringToneStop();
  titleFlashStop();
}

export function answerIncoming(target) {
  const entry = takeIncomingEntry(target);
  if (!entry) return false;
  // The accept click is a user gesture — bring the shared AudioContext
  // out of Chrome's suspended state so call audio starts immediately.
  resumeAudio().catch(() => {});
  settleIncomingAfterRemoval();
  bindSession(entry.invitation, entry.from);
  entry.invitation
    .accept({
      sessionDescriptionHandlerOptions: {
        constraints: { audio: true, video: false },
      },
    })
    .catch((err) => {
      log(`accept failed: ${err.message}`, "error");
      announce(t("acceptFailed")(err.message), "error");
      teardownSession(entry.invitation.id);
    });
  return true;
}

export function rejectIncoming(target) {
  let entry;
  if (target === undefined) {
    const queue = state.incomingSessions;
    entry = queue.length ? queue.pop() : null;
  } else {
    entry = takeIncomingEntry(target);
  }
  if (!entry) return;
  // Gesture: unlock the shared context for every later tone.
  resumeAudio().catch(() => {});
  entry.invitation.reject();
  settleIncomingAfterRemoval();
  if (!state.incomingSessions.length) releaseWarmMic();
}

// The banner stack's per-slot buttons route through here (ui.js owns
// the rows, calls.js owns the actions — the import stays one-way).
setIncomingSlotActions({ answer: answerIncoming, reject: rejectIncoming });
// The PINNED Accept/Reject buttons label slot 1 (the OLDEST ring the
// #incoming-from line displays): Accept answers the queue head; Reject
// must reject exactly that caller, never a newer one down the stack.
// The wiring lives here — next to the actions they dispatch — so the
// buttons work in any page that loads the island, and the queue's
// tests can drive them without importing main.js.
els.accept.addEventListener("click", () => answerIncoming());
els.reject.addEventListener("click", () => rejectIncoming(0));
