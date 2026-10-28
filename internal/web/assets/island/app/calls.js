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
import {
  asrEnabled,
  asrLocality,
  agentEnabled,
  crmEnabled,
  sipDomain,
} from "./config.js";
import { authedFetch } from "./auth.js";
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
  CAPTION_LINE_CAP,
  callSegments,
  clearCallSegments,
  isTranscribing,
  pauseLiveTranscription,
  saveTranscriptSegment,
  startLiveTranscription,
  stopLiveTranscription,
  subscribeCallSegments,
} from "./transcribe.js";
import {
  announce,
  els,
  errorMessage,
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

/** @param {string} id */
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
    log(`audio playback blocked: ${errorMessage(err)}`, "warn");
    announce(t("audioBlocked"), "warn");
  });
}

/**
 * setTracks flips one call's media directions through the 0.21 hold helpers.
 * Both flags are optional: a caller that only mutes passes `send` alone.
 * @param {IslandCallEntry} entry
 * @param {{recv?: boolean, send?: boolean}} tracks
 */
function setTracks(entry, { recv, send }) {
  const sdh = entry.session.sessionDescriptionHandler;
  if (!sdh) return;
  if (recv !== undefined && sdh.enableReceiverTracks)
    sdh.enableReceiverTracks(recv);
  if (send !== undefined && sdh.enableSenderTracks)
    sdh.enableSenderTracks(send);
}

/** @param {string} id @param {boolean} hold */
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
    log(`hold toggle failed: ${errorMessage(err)}`, "error");
    announce(
      t(hold ? "holdFailed" : "resumeFailed")(errorMessage(err)),
      "error",
    );
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

/** @param {string} id */
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

/** @param {string} id */
export function teardownSession(id) {
  const entry = sessions.get(id);
  if (!entry) return;
  stopLiveTranscription(id);
  if (entry.captionUnsub) entry.captionUnsub();
  clearCallSegments(id);
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

// settleEndedCall is the NORMAL termination pipeline (transcription
// stops, queued segments flush, timers clear, the session-table entry
// goes exactly as teardownSession would) — but the card STAYS as a
// settled ended card: the post-call row replaces the live controls
// (the summarize affordance when the agent seam is on and the call
// actually carried media — only established calls have a transcript —
// plus Close always). Nothing reads the session map afterwards: the
// DOM outlives the call, the state does not. Error-path teardowns
// (failed setup, watchdog rebuilds) keep teardownSession's instant
// removal — a card for a call that never lived offers nothing.
/** @param {string} id */
export function settleEndedCall(id) {
  const entry = sessions.get(id);
  if (!entry) return;
  stopLiveTranscription(id);
  if (entry.captionUnsub) entry.captionUnsub();
  // The segment STORE dies with the live call (it feeds the next live
  // card's pre-render); the ended card's strip lines live in the DOM and
  // keep serving Copy/Clear after it.
  clearCallSegments(id);
  flushTranscriptSaves(entry);
  if (entry.timer) clearInterval(entry.timer);
  morphToEndedCard(id, entry);
  showEndedReason(id, entry);
  // The AUTOMATIC review (issue #8): a media call's settle dispatches
  // the review job server-side (202, detached, idempotent with the
  // manual button) — the operator stops clicking for the standard path.
  // The button is the status surface: queued while the job runs,
  // back to the manual affordance on failure.
  if (agentEnabled && entry.established) {
    const card = entry.dom;
    const summarizeBtn = /** @type {HTMLButtonElement | null} */ (
      card ? card.querySelector(".summarize-btn") : null
    );
    if (summarizeBtn) dispatchAutoReview(id, summarizeBtn);
  }
  sessions.delete(id);
  if (state.focusedId === id) {
    const next = sessions.keys().next();
    state.focusedId = next.done ? null : next.value;
    if (state.focusedId) attachRemoteAudio(state.focusedId);
  }
  renderCalls();
}

// morphToEndedCard turns one live call card into its settled ended
// shape: the ended state line, the post-call action row, and the
// wp-ended-card hook the stylesheet uses to mute the card.
/** @param {string} id @param {IslandCallEntry} entry */
function morphToEndedCard(id, entry) {
  const card = entry.dom;
  if (!card) return;
  card.classList.add("wp-ended-card");
  card.dataset.callId = id;
  const stateEl = card.querySelector(".call-state-text");
  if (stateEl) {
    stateEl.textContent = entry.established
      ? `${t("ended")} · ${durationLabel(entry.startedAt)}`
      : t("ended");
  }
  const controls = card.querySelector(".controls");
  if (!controls) return;
  controls.replaceChildren();
  if (crmEnabled) {
    controls.append(dispositionRow(id, entry));
  }
  if (agentEnabled && entry.established) {
    controls.append(summarizeButton(id));
  }
  const close = document.createElement("button");
  close.className = "ghost close-ended-btn";
  close.textContent = t("close");
  close.addEventListener("click", () => card.remove());
  controls.append(close);
}

// dispositionRow is the operator-confirm affordance (issue #6): three
// chips answering "how did this call end" — the ONE authority that may
// upgrade the auto-report's honest "unknown" to a claimed answer (the
// island cannot see who picked up; only the human can). Success
// collapses the row to the saved verdict; a failure keeps the chips
// clickable (the server kept the verdict on 502 — only the CRM
// correction is retryable; re-click re-POSTs the same verdict, which
// the server dedupes per call+verdict).
/** @param {string} id @param {IslandCallEntry} entry @returns {HTMLDivElement} */
export function dispositionRow(id, entry) {
  const row = document.createElement("div");
  row.className = "disposition-row";
  const ask = document.createElement("span");
  ask.className = "disposition-ask";
  ask.textContent = t("dispositionAsk");
  row.append(ask);
  /** @type {Array<[string, IslandMessageKey]>} */
  const chips = [
    ["answered", "dispositionAnswered"],
    ["voicemail", "dispositionVoicemail"],
    ["ivr", "dispositionIvr"],
  ];
  for (const [value, key] of chips) {
    const chip = document.createElement("button");
    chip.className = "ghost disposition-chip";
    chip.dataset.disposition = value;
    chip.textContent = t(key);
    chip.addEventListener("click", () => {
      confirmDisposition(row, id, entry, value, key).catch((err) =>
        log(`disposition confirm failed: ${errorMessage(err)}`, "warn"),
      );
    });
    row.append(chip);
  }
  return row;
}

/**
 * @param {HTMLElement} row
 * @param {string} id
 * @param {IslandCallEntry} entry
 * @param {string} value
 * @param {IslandMessageKey} labelKey
 * @returns {Promise<void>}
 */
async function confirmDisposition(row, id, entry, value, labelKey) {
  const chips = /** @type {NodeListOf<HTMLButtonElement>} */ (
    row.querySelectorAll(".disposition-chip")
  );
  for (const chip of chips) chip.disabled = true;
  try {
    const res = await authedFetch(
      `/api/calls/${encodeURIComponent(id)}/disposition`,
      {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          disposition: value,
          number: entry.target,
          direction: entry.session instanceof SIP.Inviter ? "out" : "in",
          seconds:
            entry.startedAt > 0
              ? Math.floor((Date.now() - entry.startedAt) / 1000)
              : 0,
        }),
      },
    );
    if (!res.ok) {
      for (const chip of chips) chip.disabled = false;
      announce(t("dispositionFailed"), "warn");
      return;
    }
    row.replaceChildren(
      Object.assign(document.createElement("span"), {
        className: "disposition-ask",
        textContent: t("dispositionSaved")(t(labelKey)),
      }),
    );
    announce(t("dispositionSaved")(t(labelKey)));
  } catch (err) {
    for (const chip of chips) chip.disabled = false;
    announce(t("dispositionFailed"), "warn");
    throw err;
  }
}

// failureReasonOf maps one call's outcome to a human reason line for
// the ended card: the SIP cause first, then the missed-call class for
// an inbound call that never connected. Established calls ended
// normally and carry no reason. THE mapping (unknown stays a failure
// class, never "answered") is the one TODO_LIST #6's disposition work
// must align with.
/** @param {IslandCallEntry} entry @returns {string | null} */
function failureReasonOf(entry) {
  if (entry.established) return null;
  if (!(entry.session instanceof SIP.Inviter)) return t("callMissed");
  const cause = entry.rejectCause || 0;
  switch (cause) {
    case 486:
    case 600:
    case 603:
      return t("callBusy");
    case 404:
      return t("callNotFound");
    case 403:
    case 604:
      return t("callBlocked");
    case 480:
    case 488:
    case 502:
    case 503:
    case 504:
      return t("callUnreachable");
    case 487:
      return t("callCancelled");
    case 408:
      return t("callNoAnswer");
    default:
      return cause ? t("callFailedCode")(cause) : t("callFailedUnknown");
  }
}

// showEndedReason renders the failure reason line on the settled card
// and announces it once. It runs one microtask after the Terminated
// state change because the library fires the cause delegate
// synchronously AFTER the state change this card settled in.
/** @param {string} id @param {IslandCallEntry} entry */
function showEndedReason(id, entry) {
  queueMicrotask(() => {
    if (entry.endedReasonShown) return;
    const reason = failureReasonOf(entry);
    if (!reason) return;
    entry.endedReasonShown = true;
    announce(reason, "warn");
    const stateEl = entry.dom && entry.dom.querySelector(".call-state-text");
    if (!stateEl || !stateEl.parentElement) return;
    const line = document.createElement("div");
    line.className = "call-reason-text";
    line.textContent = reason;
    stateEl.parentElement.insertBefore(line, stateEl.nextSibling);
  });
}

// summarizeButton is the ended card's one AI affordance: the ask is
// idempotent server-side, the busy state disables the row (a double
// click must not double-speak), and every settle is honest — success
// and duplicates point at Messages (nav nudge), a failure returns the
// SETTLED retryable state with the reason as a toast.
/** @param {string} id @returns {HTMLButtonElement} */
function summarizeButton(id) {
  const btn = document.createElement("button");
  btn.className = "ghost summarize-btn";
  btn.textContent = t("summarizeDraft");
  btn.addEventListener("click", () => summarizeCall(id, btn));
  return btn;
}

/** @param {string} id @param {HTMLButtonElement} btn @returns {Promise<void>} */
async function summarizeCall(id, btn) {
  btn.disabled = true;
  btn.textContent = t("summarizing");
  let status = 0;
  let duplicate = false;
  try {
    const res = await authedFetch(
      `/api/calls/${encodeURIComponent(id)}/summary-draft`,
      { method: "POST" },
    );
    status = res.status;
    if (res.ok) {
      const body = await res.json().catch(() => ({}));
      duplicate = body.duplicate === true;
    }
  } catch {
    status = 0;
  }
  if (status === 200) {
    announce(duplicate ? t("summarizeDuplicate") : t("summarizeReady"), "ok");
    document.dispatchEvent(new CustomEvent("wp:drafts-changed"));
    btn.textContent = t("summarizeSettled");
    return;
  }
  btn.disabled = false;
  btn.textContent = t("summarizeDraft");
  if (status === 404) announce(t("summarizeNoTranscript"), "warn");
  else if (status === 0) announce(t("summarizeOffline"), "warn");
  else announce(t("summarizeFailed"), "error");
}

// dispatchAutoReview fires the AUTOMATIC review ask once per settle
// (issue #8): the server detaches the job and answers 202 — duplicate
// names the idem replay when the manual button won the race, or the
// job is already running. The button is the status surface: queued
// text while accepted, back to the manual affordance on failure (the
// failure is the server's too — the job audited and stayed retryable).
/** @param {string} id @param {HTMLButtonElement} btn @returns {Promise<void>} */
async function dispatchAutoReview(id, btn) {
  btn.disabled = true;
  btn.textContent = t("summarizing");
  let status = 0;
  let duplicate = false;
  try {
    const res = await authedFetch(`/api/calls/${encodeURIComponent(id)}/review`, {
      method: "POST",
    });
    status = res.status;
    if (res.ok) {
      const body = await res.json().catch(() => ({}));
      duplicate = body.duplicate === true;
    }
  } catch {
    status = 0;
  }
  if (status === 202) {
    // A duplicate replay means the job was already queued (or the manual
    // button won the race): say so instead of pretending it just started.
    const label = t(duplicate ? "summarizeDuplicate" : "reviewQueued");
    announce(label, "ok");
    btn.textContent = label;
    return;
  }
  btn.disabled = false;
  btn.textContent = t("summarizeDraft");
  if (status === 404) announce(t("summarizeNoTranscript"), "warn");
  else if (status === 0) announce(t("summarizeOffline"), "warn");
  else announce(t("summarizeFailed"), "error");
}

export function teardownAll() {
  [...sessions.keys()].forEach(teardownSession);
  state.focusedId = null;
}

/** @param {number} startedAt @returns {string} */
function durationLabel(startedAt) {
  const sec = Math.max(0, Math.floor((Date.now() - startedAt) / 1000));
  return `${Math.floor(sec / 60)}:${String(sec % 60).padStart(2, "0")}`;
}

// liveStateLine derives the card/chip state line for one entry — the
// ONE derivation both surfaces read (the dock chip and the expanded
// card can never disagree about what a call is doing).
/**
 * @param {IslandCallEntry} entry
 * @returns {{dataState: IslandCallDockState, text: string}}
 */
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

/** @type {Map<string, HTMLDivElement>} */
const dockChips = new Map();
/** @type {HTMLDivElement | null} */
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
  // Bound once: the forEach callbacks below are closures, so the module-level
  // `let` does not stay narrowed inside them.
  const dock = dockEl;
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
      dock.append(chip);
    }
    updateDockChip(chip, id, entry);
  });
}

/** @returns {HTMLDivElement} */
function createCallDock() {
  const dock = document.createElement("div");
  dock.id = "wp-call-dock";
  dock.className = "wp-call-dock";
  dock.setAttribute("role", "toolbar");
  dock.setAttribute("aria-label", t("dockLabel"));
  els.calls.prepend(dock);
  return dock;
}

/** @param {string} id @param {IslandCallEntry} entry @returns {HTMLDivElement} */
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

/** @param {HTMLDivElement} chip @param {string} id @param {IslandCallEntry} entry */
function updateDockChip(chip, id, entry) {
  const { dataState, text } = liveStateLine(entry);
  chip.dataset.state = dataState;
  chip.classList.toggle("focused", id === state.focusedId);
  const pill = /** @type {HTMLElement} */ (chip.children[2]);
  pill.textContent = text;
}

// chipAvatar mirrors the server's desk-phone vernacular: a hue-bucketed
// disc showing the number's LAST TWO DIGITS (never a country signum).
// Classes, never inline styles — the strict CSP silently drops those.
/** @param {string} target @returns {HTMLSpanElement} */
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

/** @param {string} target @returns {HTMLSpanElement} */
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
    const stateEl = /** @type {HTMLElement} */ (
      entry.dom.querySelector(".call-state-text")
    );
    stateEl.textContent = text;
    entry.dom.dataset.state = dataState;
    announceCallState(entry, dataState);
    entry.dom.classList.toggle("focused", id === state.focusedId);
    // Dock mode: the chips carry the other calls; only the focused call
    // keeps its expanded card. A single call renders the plain card.
    entry.dom.hidden = sessions.size > 1 && id !== state.focusedId;
    const holdBtn = /** @type {HTMLButtonElement} */ (
      entry.dom.querySelector(".hold-btn")
    );
    holdBtn.disabled = Boolean(entry.holdPending);
    holdBtn.textContent = entry.holdPending
      ? t(entry.holdPending === "holding" ? "callHolding" : "callResuming")
      : entry.held
        ? t("resume")
        : t("hold");
    const muteBtn = /** @type {HTMLButtonElement} */ (
      entry.dom.querySelector(".mute-btn")
    );
    muteBtn.textContent = entry.muted ? t("unmute") : t("mute");
    const focusBtn = entry.dom.querySelector(".focus-btn");
    if (focusBtn) focusBtn.textContent = t("focus");
    const endBtn = entry.dom.querySelector(".hangup-btn");
    if (endBtn) endBtn.textContent = t("end");
    const slider = /** @type {HTMLInputElement | null} */ (
      entry.dom.querySelector(".volume-slider")
    );
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
  const focused = state.focusedId ? sessions.get(state.focusedId) : undefined;
  els.keypad.hidden = !(
    focused && focused.session.state === SIP.SessionState.Established
  );
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
/** @param {IslandCallEntry} entry @param {IslandCallDockState} dataState */
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

/** @param {string} id @param {string} target @returns {HTMLDivElement} */
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
  /**
   * @param {string} label
   * @param {string} cls
   * @param {(event: MouseEvent) => void} onClick
   * @returns {HTMLButtonElement}
   */
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
  // Live captions (only when the ASR seam is configured): the strip is
  // the line-based reading surface under the card (last CAPTION_LINE_CAP
  // segments, aria-hidden — the polite announce path already speaks each
  // segment; the strip must not double-speak), with a config-honest
  // locality chip naming where the audio is processed. The inner
  // .call-transcript node stays the copy/clear surface.
  const transcript = document.createElement("div");
  transcript.className = "call-transcript";
  transcript.id = `call-transcript-${id}`;
  if (asrEnabled) {
    const strip = document.createElement("div");
    strip.className = "wp-captions";
    strip.id = `wp-captions-${id}`;
    strip.setAttribute("aria-hidden", "true");
    const captionsHead = document.createElement("div");
    captionsHead.className = "wp-captions-head";
    const title = document.createElement("span");
    title.className = "wp-captions-title";
    title.textContent = t("captionsTitle");
    captionsHead.append(title);
    if (asrLocality) {
      const locality = document.createElement("span");
      locality.className = "wp-captions-locality";
      locality.textContent = t(
        asrLocality === "local" ? "captionsLocal" : "captionsRemote",
      );
      captionsHead.append(locality);
    }
    strip.append(captionsHead, transcript);
    card.append(head, controls, strip);
    for (const line of callSegments(id)) appendTranscript(transcript, line);
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
  } else {
    card.append(head, controls, transcript);
  }
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
/** @param {string} id @param {HTMLElement} target */
function toggleTranscription(id, target) {
  const entry = sessions.get(id);
  if (!entry) return;
  if (isTranscribing(id)) {
    stopLiveTranscription(id);
    announce(t("transcribeStopped"), "info");
  } else if (!startCallTranscription(id, entry)) {
    target.textContent = t("transcribeUnsupported");
  }
  renderCalls();
}

// startCallTranscription is the ONE start path (auto-start on connect
// and the manual toggle both ride it): it wires segment text to
// onCallTranscript and passes the UI language as the provider hint.
/** @param {string} id @param {IslandCallEntry} entry @returns {boolean} */
function startCallTranscription(id, entry) {
  return startLiveTranscription(
    id,
    entry,
    (text) => onCallTranscript(id, entry, text),
    {
      language: getLang(),
    },
  );
}

// ensureTranscription is the auto-start hook: transcription begins by
// itself when the ASR seam is on (owner decision 2026-10-07) — the
// button remains the honest manual override. An honest notice replaces
// a silent failure to start.
/** @param {string} id @param {IslandCallEntry} entry */
function ensureTranscription(id, entry) {
  if (!asrEnabled) return;
  const target = entry.dom && entry.dom.querySelector(".call-transcript");
  if (!target || isTranscribing(id)) return;
  if (startCallTranscription(id, entry)) {
    announce(t("transcribeStarted"), "info");
    renderCalls();
  } else {
    target.textContent = t("transcribeUnsupported");
  }
}

// onCallTranscript is the per-segment sink: the screen reader hears the
// delta (polite — this IS the product for them), and the segment goes
// through the per-call save queue (a failed save costs the LATER
// re-read, never the live card). The strip renders via the
// transcribe.js segment subscription — this path never touches the DOM
// (one renderer per surface).
/** @param {string} id @param {IslandCallEntry} entry @param {string} text */
function onCallTranscript(id, entry, text) {
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

/** @param {IslandCallEntry} entry @param {IslandTranscriptSegment} segment */
function queueTranscriptSave(entry, segment) {
  const queue = (entry.transcriptSaveQueue ??= []);
  queue.push(segment);
  if (queue.length > TRANSCRIPT_SAVE_QUEUE_CAP) {
    queue.shift();
  }
  flushTranscriptSaves(entry);
}

/** @param {IslandCallEntry} entry */
function flushTranscriptSaves(entry) {
  if (entry.transcriptSaveFlushing || !entry.transcriptSaveQueue?.length)
    return;
  entry.transcriptSaveFlushing = true;
  const queue = /** @type {IslandTranscriptSegment[]} */ (
    entry.transcriptSaveQueue
  );
  saveTranscriptSegment(queue[0])
    .then(() => {
      queue.shift();
      entry.transcriptSaved = (entry.transcriptSaved || 0) + 1;
      entry.transcriptSaveFlushing = false;
      showTranscriptSaved(entry);
      flushTranscriptSaves(entry);
    })
    .catch((err) => {
      entry.transcriptSaveFlushing = false;
      log(`transcript save failed: ${errorMessage(err)}`, "warn");
      if (!entry.transcriptSaveWarned) {
        entry.transcriptSaveWarned = true;
        announce(t("transcriptSaveFailed")(errorMessage(err)), "warn");
      }
    });
}

// showTranscriptSaved is the #13 discoverability affordance: a CONFIRMED
// save surfaces a persistent "saved to History" count on the card and
// announces once per call — the live transcript must not read as a
// card-only artifact that vanishes on hangup.
/** @param {IslandCallEntry} entry */
function showTranscriptSaved(entry) {
  const marker = /** @type {HTMLElement | null} */ (
    entry.dom && entry.dom.querySelector(".call-transcript-saved")
  );
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
/** @param {string} text @returns {string} */
function clipForAnnounce(text) {
  const clipped = text.length > 140 ? `${text.slice(0, 139)}…` : text;
  return t("transcriptDelta")(clipped);
}

// The live card's DOM cap (multi-call train): the transcript renders as
// bounded LINE divs (CAPTION_LINE_CAP, scalar-bounded like the server
// store cap) — a marathon call cannot grow the strip. The durable whole
// lives server-side (History) regardless.
/** @param {HTMLElement} target @param {string} text */
function appendTranscript(target, text) {
  if (!text) return;
  const line = document.createElement("div");
  line.className = "wp-caption-line";
  line.textContent = text;
  target.append(line);
  while (target.children.length > CAPTION_LINE_CAP) target.firstChild?.remove();
  target.scrollTop = target.scrollHeight;
}

// clearCardTranscript is the per-call clear affordance (#39): the live
// card is presentation only — clearing it never touches the stored
// History transcript (nor the per-call segment log), and the announce
// says exactly that scope.
/** @param {HTMLElement} target */
function clearCardTranscript(target) {
  if (!target.children.length && !target.textContent) {
    announce(t("transcribeEmpty"), "warn");
    return;
  }
  target.replaceChildren();
  announce(t("transcriptCleared"), "info");
}

// transcriptText reads the call's words for the clipboard: the line
// divs join with spaces (the browser DOM would concatenate bare); a
// plain text node (the unsupported notice) passes through.
/** @param {HTMLElement} target @returns {string} */
function transcriptText(target) {
  if (!target.children.length) return (target.textContent || "").trim();
  return Array.from(target.children, (line) => line.textContent)
    .join(" ")
    .trim();
}

// copyCardTranscript puts the card's accumulated transcript on the
// clipboard (secure context — the stack fronts TLS). Honest feedback
// either way: an empty transcript announces that instead of copying
// nothing, a missing clipboard API announces the manual fallback.
/** @param {HTMLElement} target */
function copyCardTranscript(target) {
  const text = transcriptText(target);
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
/** @param {HTMLElement} card @param {string} id */
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

/** @param {SipNotification} notification */
function referOnNotify(notification) {
  const body = (notification.request && notification.request.body) || "";
  const status = body.match(/^SIP\/2\.0 (\d{3})/m);
  // accept() rejects only on prompt dismissal; the transfer verdict
  // below carries the user surface.
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

/** @param {string} id @param {string} destination @returns {Promise<void>} */
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
    log(t("transferFailed")(errorMessage(err)), "error");
    announce(t("transferFailed")(errorMessage(err)), "error");
    renderCalls();
  }
}

/** @param {string} id @returns {Promise<void>} */
async function attendedTransfer(id) {
  const entry = sessions.get(id);
  if (!entry || entry.session.state !== SIP.SessionState.Established) return;
  // The other ESTABLISHED call is the consult leg; Replaces points at
  // its dialog so the network bridges the two far ends.
  /** @type {IslandCallEntry | null} */
  let partner = null;
  sessions.forEach((other, otherId) => {
    if (
      otherId !== id &&
      other.session.state === SIP.SessionState.Established &&
      !other.transferring
    )
      partner = other;
  });
  // The forEach above ran in a closure, so TypeScript still sees the
  // initializer (`null`) here — the cast names what the closure wrote.
  const other = /** @type {IslandCallEntry | null} */ (partner);
  if (!other) {
    log(t("transferNoPartner"), "warn");
    announce(t("transferNoPartner"), "warn");
    return;
  }
  entry.transferring = true;
  renderCalls();
  try {
    await entry.session.refer(other.session, { onNotify: referOnNotify });
    log(`attended transfer ${entry.target} ↔ ${other.target} sent`);
  } catch (err) {
    entry.transferring = false;
    log(t("transferFailed")(errorMessage(err)), "error");
    announce(t("transferFailed")(errorMessage(err)), "error");
    renderCalls();
  }
}

// Diagnostic handle: the E2E suite reads getStats() from these to
// prove real media flows (bytes on the wire), not just signaling.
window.__pcs = window.__pcs || new Map();

/** @param {SipSession} newSession @param {string} target */
export function bindSession(newSession, target) {
  newSession.stateChange.addListener((sessionState) => {
    if (sessionState === SIP.SessionState.Established) {
      const pc =
        newSession.sessionDescriptionHandler &&
        newSession.sessionDescriptionHandler.peerConnection;
      if (pc && window.__pcs) window.__pcs.set(newSession.id, pc);
    }
  });
  const id = newSession.id;
  /** @type {IslandCallEntry} */
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
  // The caption strip rides the per-call segment subscription: new
  // segments render as bounded lines; the subscription dies with the
  // card (teardownSession unsubscribes and clears the log).
  if (asrEnabled) {
    const target = /** @type {HTMLElement | null} */ (
      entry.dom.querySelector(".call-transcript")
    );
    if (target) {
      entry.captionUnsub = subscribeCallSegments(id, (text) =>
        appendTranscript(target, text),
      );
    }
  }

  newSession.stateChange.addListener((sessionState) => {
    const live = sessions.get(id);
    if (!live) return;
    log(`call ${target} ${sessionState}`);
    // Emission through the declarative table (D-train): every sip.js
    // state arrival maps onto a view state and is legality-checked
    // against CALL_TRANSITIONS. A refused arrival (Established twice,
    // traffic after ended) is a wiring bug — logged loudly here, the
    // view-state derivation below stays single-sourced either way.
    /** @type {Partial<Record<SipSessionStateName, IslandCallViewState>>} */
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
      settleEndedCall(id);
      // A just-ended call may have left a voicemail deposit.
      scheduleVoicemailRefresh();
      refreshServerHistory();
    }
    renderCalls();
  });
  renderCalls();
}

/** @param {string} id @returns {Promise<void>} */
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
    log(`hangup: ${errorMessage(err)}`, "error");
    announce(t("hangupFailed")(errorMessage(err)), "error");
    teardownSession(id);
  }
}

// DTMF feedback (A8): the pressed key pulses — on INVOCATION, not
// delivery, so keyboard-triggered tones flash too and the press is
// acknowledged instantly; the #log line records the verdict.
/** @param {string} tone */
function flashKey(tone) {
  const key = /** @type {HTMLElement | null} */ (
    els.keypad.querySelector(`button[data-tone="${tone}"]`)
  );
  if (!key || !key.classList) return;
  key.classList.remove("wp-key-sent");
  void key.offsetWidth; // restart the animation on rapid repeats
  key.classList.add("wp-key-sent");
}

// The transient tones-sent trail under the keypad: the last ≤12 digits,
// fading to hidden after 3 s of silence. #log stays the durable record;
// this is the in-the-moment confirmation for mid-call menus.
/** @type {string[]} */
let tonesSent = [];
/** @type {ReturnType<typeof setTimeout> | null} */
let tonesTimer = null;

/** @param {string} tone */
function noteToneSent(tone) {
  /** @type {HTMLElement | null} */
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

/** @param {string} tone */
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
      log(`dtmf failed: ${errorMessage(err)}`, "error");
      announce(t("dtmfFailed")(errorMessage(err)), "error");
    });
}

// sanitizeDialable strips everything that is not dialable: pasted
// numbers routinely carry invisible Unicode direction marks (macOS/phone
// apps add them around telephone numbers) and formatting (spaces,
// dashes, parentheses), and makeURI rejects all of it. ONE home for the
// charset — placeCall dials it, initDialHint previews it.
/** @param {string} raw @returns {string} */
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
/** @param {string} raw @returns {Promise<boolean>} */
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
    // A SIP rejection never rejects invite(): sip.js resolves once the
    // INVITE is sent, the verdict arrives through the requestDelegate.
    // The cause lands on the session entry so the ended card can tell
    // the user WHY the call failed (ONE mapping, failureReasonOf).
    await inviter.invite({
      requestDelegate: {
        onReject: (
          /** @type {{statusCode?: number} | null | undefined} */ res,
        ) => {
          const entry = sessions.get(inviter.id);
          if (entry)
            entry.rejectCause = res && res.statusCode ? res.statusCode : 0;
        },
      },
    });
    return true;
  } catch (err) {
    log(`invite failed: ${errorMessage(err)}`, "error");
    showDialError(t("callFailed")(errorMessage(err)));
    teardownSession(inviter.id);
    return false;
  }
}

// Keyboard-shortcut surface (shortcuts.js drives these; the call-card
// buttons and the incoming-call banner share the same primitives).

export function toggleMuteFocused() {
  const entry = state.focusedId ? sessions.get(state.focusedId) : undefined;
  if (!entry) return;
  entry.muted = !entry.muted;
  setTracks(entry, { send: !entry.held && !entry.muted });
  renderCalls();
}

export function toggleHoldFocused() {
  if (!state.focusedId) return;
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
/** @param {number} step */
export function cycleFocus(step) {
  const established = [...sessions.entries()].filter(
    ([, entry]) => entry.session.state === SIP.SessionState.Established,
  );
  if (established.length < 2) return;
  const at = established.findIndex(([id]) => id === state.focusedId);
  const next =
    established[
      (Math.max(at, 0) + step + established.length) % established.length
    ];
  if (!next) return;
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
/**
 * @param {SipInvitation | number | undefined} target
 *   undefined = the queue head, a number = the slot index, a SipInvitation
 *   = that exact ring (the banner stack's per-slot buttons).
 * @returns {IslandIncomingSession | null}
 */
function takeIncomingEntry(target) {
  const queue = state.incomingSessions;
  if (!queue.length) return null;
  /** @type {number} */
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

/** @param {SipInvitation | number} [target] @returns {boolean} */
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
      log(`accept failed: ${errorMessage(err)}`, "error");
      announce(t("acceptFailed")(errorMessage(err)), "error");
      teardownSession(entry.invitation.id);
    });
  return true;
}

/** @param {SipInvitation | number} [target] */
export function rejectIncoming(target) {
  /** @type {IslandIncomingSession | null | undefined} */
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
