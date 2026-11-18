// The post-call card family: everything the settled ended card offers
// after the conversation is over — the failure-reason line, the note
// editor, the disposition chips, and the two AI-draft affordances.
// calls.js owns the settle machinery (morphToEndedCard) and imports the
// pieces from here; this module imports only ui/auth/i18n, never calls.js.

import { authedFetch } from "./auth.js";
import { t } from "./i18n.js";
import { announce, errorMessage, focusAfterReplace, log } from "./ui.js";

// q850CauseOf reads the Q.850 cause out of a SIP Reason header value
// ("Q.850;cause=21;text=...") — the carrier verdict that rides UNDER a
// 403 wrapper. 0 when the header is absent or names another numbering
// plan. THE parser for the trunk's live failure signature (the settle
// path in calls.js reads it too).
/** @param {string | undefined} reason @returns {number} */
export function q850CauseOf(reason) {
  if (!reason) return 0;
  const match = /Q\.850;\s*cause=(\d+)/i.exec(reason);
  return match ? Number(match[1]) : 0;
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
    const saved = Object.assign(document.createElement("span"), {
      className: "disposition-ask",
      textContent: t("dispositionSaved")(t(labelKey)),
    });
    focusAfterReplace(
      row,
      () =>
        /** @type {HTMLElement | null} */ (
          row.querySelector(".disposition-ask")
        ),
      () => row.replaceChildren(saved),
    );
    announce(t("dispositionSaved")(t(labelKey)));
  } catch (err) {
    for (const chip of chips) chip.disabled = false;
    announce(t("dispositionFailed"), "warn");
    throw err;
  }
}

// noteRow is the post-call note editor (issue #11 + the 2026-11-18
// autosave train): the hangup moment gets the SAME note slot History
// renders later (PUT /api/calls/{id}/note), so the conversation is
// written down with zero navigation. The note saves itself as the
// owner types — every keystroke re-arms a 250ms debounce, blur (the
// change event) flushes, and every save rides the same event-sourced
// PUT (the server drops same-body echoes). The explicit button stays
// the "done" affordance: it saves AND collapses the row to the saved
// line; the 409/failure paths keep the draft in the editor
// (hand-written content is never silently expunged); dismissing the
// card flushes the unsaved tail first (calls.js), so it loses nothing
// already typed.

// The autosave tuning seam (the clock-seam precedent): tests shrink
// the debounce instead of sleeping real 250ms windows.
export const noteAutosave = { delayMs: 250 };

/**
 * @typedef {Object} NoteAutosaveEntry
 * @property {string} id
 * @property {string} saved - the last CONFIRMED body
 * @property {ReturnType<typeof setTimeout> | null} timer
 * @property {boolean} inflight
 * @property {boolean} announced - one failure announcement per episode
 */
/** @type {Map<HTMLElement, NoteAutosaveEntry>} */
const noteAutosaveEntries = new Map();

// flushPendingCallNotes fires every editor's unsaved tail NOW: the
// ended card's Close button calls it before the DOM goes away, and the
// pagehide flush keeps the last PUT alive past teardown (keepalive).
/** @param {boolean} [keepalive] */
export function flushPendingCallNotes(keepalive = false) {
  for (const entry of noteAutosaveEntries.values()) {
    if (entry.timer) {
      clearTimeout(entry.timer);
      entry.timer = null;
    }
  }
  for (const [row, entry] of noteAutosaveEntries) {
    if (!entry.inflight) runNoteAutosave(row, entry, keepalive);
  }
}

document.addEventListener("pagehide", () => flushPendingCallNotes(true));

/**
 * @param {HTMLElement} row
 * @param {string} text
 */
function setNoteStatus(row, text) {
  const status = row.querySelector(".call-note-status");
  if (status) status.textContent = text;
}

/**
 * @param {HTMLElement} row
 * @param {NoteAutosaveEntry} entry
 * @param {number} status - the HTTP verdict; 0 = transport failure
 */
function noteAutosaveFailed(row, entry, status) {
  setNoteStatus(row, t("callNoteUnsaved"));
  // Autosave must not announce on every typing-pause retry: one
  // announcement per failure episode, reset on the next success.
  if (!entry.announced) {
    announce(status === 409 ? t("callNoteFull") : t("callNoteFailed"), "warn");
    entry.announced = true;
  }
}

/**
 * The quiet save lane: latest keystroke wins, no collapse, no success
 * announcement (the status line carries it). A newer body that lands
 * mid-flight re-runs on completion — PUTs never race out of order.
 * @param {HTMLElement} row
 * @param {NoteAutosaveEntry} entry
 * @param {boolean} keepalive
 */
async function runNoteAutosave(row, entry, keepalive) {
  const input = /** @type {HTMLTextAreaElement | null} */ (
    row.querySelector(".call-note-input")
  );
  if (!input || entry.inflight) return;
  const body = input.value;
  if (body === entry.saved) return;
  entry.inflight = true;
  const carried = body;
  let landed = false;
  setNoteStatus(row, t("callNoteSaving"));
  try {
    const res = await authedFetch(
      `/api/calls/${encodeURIComponent(entry.id)}/note`,
      {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ body }),
        keepalive,
      },
    );
    if (!res.ok) {
      noteAutosaveFailed(row, entry, res.status);
      return;
    }
    entry.saved = body;
    entry.announced = false;
    landed = true;
    setNoteStatus(row, t("callNoteStatusSaved"));
  } catch (err) {
    noteAutosaveFailed(row, entry, 0);
    log(`note autosave failed: ${errorMessage(err)}`, "warn");
  } finally {
    entry.inflight = false;
    // Only input NEWER than this flight's body re-dispatches — a
    // FAILED save whose text is still sitting in the editor must not
    // retry itself in a loop; the next keystroke (or the re-arm below)
    // carries it. Success carries immediately; failure re-arms the
    // debounce once, bounded by the unchanged-body guard above.
    const live = /** @type {HTMLTextAreaElement | null} */ (
      row.querySelector(".call-note-input")
    );
    if (!live || entry.timer || live.value === carried) return;
    if (landed) {
      runNoteAutosave(row, entry, false);
      return;
    }
    entry.timer = setTimeout(() => {
      entry.timer = null;
      runNoteAutosave(row, entry, false);
    }, noteAutosave.delayMs);
  }
}

/** @param {string} id @returns {HTMLDivElement} */
export function noteRow(id) {
  const row = document.createElement("div");
  row.className = "call-note-row";
  const input = document.createElement("textarea");
  input.className = "call-note-input";
  input.placeholder = t("callNotePlaceholder");
  input.maxLength = 4000;
  input.rows = 3;
  const status = document.createElement("span");
  status.className = "call-note-status";
  /** @type {NoteAutosaveEntry} */
  const entry = {
    id,
    saved: "",
    timer: null,
    inflight: false,
    announced: false,
  };
  noteAutosaveEntries.set(row, entry);
  input.addEventListener("input", () => {
    if (entry.timer) clearTimeout(entry.timer);
    entry.timer = setTimeout(() => {
      entry.timer = null;
      runNoteAutosave(row, entry, false);
    }, noteAutosave.delayMs);
  });
  // Blur is the natural "done typing" beat: change fires only when the
  // value moved, so the flush never PUTs a settled box.
  input.addEventListener("change", () => {
    if (entry.timer) {
      clearTimeout(entry.timer);
      entry.timer = null;
    }
    runNoteAutosave(row, entry, false);
  });
  const save = document.createElement("button");
  save.className = "ghost call-note-save";
  save.textContent = t("callNoteSave");
  save.addEventListener("click", () => {
    if (entry.timer) {
      clearTimeout(entry.timer);
      entry.timer = null;
    }
    saveCallNote(row, entry).catch((err) =>
      log(`note save failed: ${errorMessage(err)}`, "warn"),
    );
  });
  row.append(input, save, status);
  return row;
}

/**
 * The explicit lane: save now and collapse the row to the saved line —
 * with autosave the button means "done", and an unchanged body still
 * collapses (nothing new to carry, the owner is finished here).
 * @param {HTMLElement} row
 * @param {NoteAutosaveEntry} entry
 * @returns {Promise<void>}
 */
async function saveCallNote(row, entry) {
  const save = /** @type {HTMLButtonElement | null} */ (
    row.querySelector(".call-note-save")
  );
  const input = /** @type {HTMLTextAreaElement | null} */ (
    row.querySelector(".call-note-input")
  );
  if (!save || !input) return;
  if (entry.inflight) return; // the flying autosave carries this body
  const body = input.value;
  if (body === entry.saved) {
    collapseNoteRow(row);
    return;
  }
  save.disabled = true;
  try {
    const res = await authedFetch(
      `/api/calls/${encodeURIComponent(entry.id)}/note`,
      {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ body }),
      },
    );
    if (!res.ok) {
      save.disabled = false;
      entry.announced = false; // a manual retry announces again
      if (res.status === 409) announce(t("callNoteFull"), "warn");
      else announce(t("callNoteFailed"), "warn");
      setNoteStatus(row, t("callNoteUnsaved"));
      return;
    }
    entry.saved = body;
    entry.announced = false;
    collapseNoteRow(row);
    announce(t("callNoteSaved"));
  } catch (err) {
    save.disabled = false;
    announce(t("callNoteFailed"), "warn");
    setNoteStatus(row, t("callNoteUnsaved"));
    throw err;
  }
}

/**
 * @param {HTMLElement} row
 */
function collapseNoteRow(row) {
  const input = row.querySelector(".call-note-input");
  const save = row.querySelector(".call-note-save");
  const status = row.querySelector(".call-note-status");
  const saved = Object.assign(document.createElement("span"), {
    className: "call-note-saved",
    textContent: t("callNoteSaved"),
  });
  noteAutosaveEntries.delete(row);
  focusAfterReplace(
    row,
    () =>
      /** @type {HTMLElement | null} */ (row.querySelector(".call-note-saved")),
    () => {
      if (input) input.remove();
      if (save) save.remove();
      if (status) status.remove();
      row.append(saved);
    },
  );
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
  // The Reason header's Q.850 cause outranks the SIP code: the trunk's
  // live failure is a 403 that carries Q.850;cause=21 (the carrier
  // rejected the call after a successful registration) — "blocked"
  // would blame the wrong side.
  switch (entry.rejectQ850 || 0) {
    case 17:
      return t("callBusy");
    case 21:
      return t("callRejectedRemote");
  }
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
    case 488:
      return t("callNotAcceptable");
    case 480:
    case 502:
    case 503:
    case 504:
      return t("callUnreachable");
    case 487:
      return t("callCancelled");
    case 408:
      return t("callNoAnswer");
    case 401:
    case 407:
      return t("callAuthFailed");
    default:
      return cause ? t("callFailedCode")(cause) : t("callFailedUnknown");
  }
}

// showEndedReason renders the failure reason line on the settled card
// and announces it once. It runs one microtask after the Terminated
// state change because the library fires the cause delegate
// synchronously AFTER the state change this card settled in.
/** @param {string} id @param {IslandCallEntry} entry */
export function showEndedReason(id, entry) {
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
export function summarizeButton(id) {
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
export async function dispatchAutoReview(id, btn) {
  btn.disabled = true;
  btn.textContent = t("summarizing");
  let status = 0;
  let duplicate = false;
  try {
    const res = await authedFetch(
      `/api/calls/${encodeURIComponent(id)}/review`,
      {
        method: "POST",
      },
    );
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
