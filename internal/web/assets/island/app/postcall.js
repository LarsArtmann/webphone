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

// noteRow is the post-call note editor (issue #11): the hangup moment
// gets the SAME note slot History renders later (PUT /api/calls/{id}/note),
// so the conversation is written down with zero navigation. A 204
// collapses the row to the saved line; the note-list-full 409 and any
// other failure keep the draft in the editor (hand-written content is
// never silently expunged); dismissing the card loses nothing already
// saved.
/** @param {string} id @returns {HTMLDivElement} */
export function noteRow(id) {
  const row = document.createElement("div");
  row.className = "call-note-row";
  const input = document.createElement("textarea");
  input.className = "call-note-input";
  input.placeholder = t("callNotePlaceholder");
  input.maxLength = 4000;
  input.rows = 3;
  const save = document.createElement("button");
  save.className = "ghost call-note-save";
  save.textContent = t("callNoteSave");
  save.addEventListener("click", () => {
    saveCallNote(row, id, input.value).catch((err) =>
      log(`note save failed: ${errorMessage(err)}`, "warn"),
    );
  });
  row.append(input, save);
  return row;
}

/**
 * @param {HTMLElement} row
 * @param {string} id
 * @param {string} body
 * @returns {Promise<void>}
 */
async function saveCallNote(row, id, body) {
  const save = /** @type {HTMLButtonElement | null} */ (
    row.querySelector(".call-note-save")
  );
  const input = /** @type {HTMLTextAreaElement | null} */ (
    row.querySelector(".call-note-input")
  );
  if (save) save.disabled = true;
  try {
    const res = await authedFetch(`/api/calls/${encodeURIComponent(id)}/note`, {
      method: "PUT",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ body }),
    });
    if (!res.ok) {
      if (save) save.disabled = false;
      if (res.status === 409) announce(t("callNoteFull"), "warn");
      else announce(t("callNoteFailed"), "warn");
      return;
    }
    const saved = Object.assign(document.createElement("span"), {
      className: "call-note-saved",
      textContent: t("callNoteSaved"),
    });
    focusAfterReplace(
      row,
      () =>
        /** @type {HTMLElement | null} */ (
          row.querySelector(".call-note-saved")
        ),
      () => {
        if (input) input.remove();
        if (save) save.remove();
        row.append(saved);
      },
    );
    announce(t("callNoteSaved"));
  } catch (err) {
    if (save) save.disabled = false;
    announce(t("callNoteFailed"), "warn");
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
