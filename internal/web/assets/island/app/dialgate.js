// The advisory dial gate (issue #7): before every SIP invite the island
// asks POST /api/dials whether this remote is inside its cooldown
// window. The gate ADVISES, it never traps — 404 (seam off), a store
// failure, or a network error all dial anyway; a phone must dial. A
// denied verdict renders inline, never a modal: the dial area grows the
// state line (last call age + remaining window) and one "Dial anyway"
// chip whose click spends the remote's one-shot override server-side.

import { csrfToken } from "./csrf.js";
import { errorMessage, log } from "./ui.js";

// One 404 pins the seam off for this page's life: every later dial
// skips the round-trip (the server would answer 404 forever — the
// config does not change under a running page).
let gateOff = false;

/**
 * @typedef {object} DialGateVerdict
 * @property {boolean} allowed
 * @property {number} [last_at] unix seconds of the slot's last allowed dial
 * @property {number} [window] the configured cooldown window, seconds
 */

/**
 * The pre-invite ask. Returns null when the island must dial without a
 * verdict (seam off, unreachable, malformed answer) — the advisory
 * contract keeps every failure on the side of dialing.
 * @param {string} remote the sanitized dialable remote
 * @returns {Promise<DialGateVerdict | null>}
 */
export async function consultDialGate(remote) {
  if (gateOff) return null;
  try {
    const res = await fetch("/api/dials", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "X-CSRF-Token": csrfToken(),
      },
      body: JSON.stringify({ remote }),
    });
    if (res.status === 404) {
      gateOff = true;
      return null;
    }
    if (!res.ok) {
      // 502/503: the gate itself failed — dial anyway, but say so.
      log(`dial gate unavailable (HTTP ${res.status}); dialing anyway`, "warn");
      return null;
    }
    const data = await res.json();
    if (!data || typeof data.allowed !== "boolean") return null;
    return data;
  } catch (err) {
    log(`dial gate unreachable (${errorMessage(err)}); dialing anyway`, "warn");
    return null;
  }
}

/**
 * The click-through: spend the remote's one-shot override. True means
 * the dial may proceed; false keeps the cooldown line on screen.
 * @param {string} remote
 * @returns {Promise<boolean>}
 */
export async function overrideDialGate(remote) {
  try {
    const res = await fetch("/api/dials", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "X-CSRF-Token": csrfToken(),
      },
      body: JSON.stringify({ remote, override: true }),
    });
    if (!res.ok) return false;
    const data = await res.json();
    return Boolean(data && data.allowed);
  } catch {
    return false;
  }
}

// --- typeahead row state (GET /api/dials) --------------------------------
//
// The typeahead decorates contact rows that are inside their cooldown
// window. The slots cache is SERVER TRUTH, fetched debounced: a local
// guess (localStorage of recent dials) would lie after a reload or on
// the second device. A failed fetch hides the decoration — it never
// blocks or lies.

/** @type {{remote: string, last_at: number, window: number}[]} */
let slots = [];
let slotsFetchedAt = 0;

const SLOTS_TTL_MS = 30_000;

/**
 * Fetch the owner's active cooldown windows (debounced). Never
 * rejects: failures leave the previous cache (or an empty one) and the
 * decoration simply stays hidden.
 * @param {boolean} [force] skip the debounce window (after a dial)
 * @returns {Promise<void>}
 */
export async function refreshDialSlots(force = false) {
  if (gateOff) return;
  if (!force && Date.now() - slotsFetchedAt < SLOTS_TTL_MS) return;
  slotsFetchedAt = Date.now();
  try {
    const res = await fetch("/api/dials", {
      headers: { "X-CSRF-Token": csrfToken() },
    });
    if (res.status === 404) {
      gateOff = true;
      return;
    }
    if (!res.ok) return;
    const data = await res.json();
    slots = Array.isArray(data && data.slots) ? data.slots : [];
  } catch {
    // Unreachable: keep the previous cache; the decoration degrades to
    // hidden (the advisory contract — a phone must dial).
  }
}

/**
 * The active cooldown slot for a contact's display number, or null.
 * Both sides normalize to the dialable alphabet before comparing (the
 * server stores the sanitized remote; contacts carry display forms).
 * @param {string} raw the contact's display number
 * @returns {{remote: string, last_at: number, window: number} | null}
 */
export function cooldownFor(raw) {
  const needle = normalizeDialable(raw);
  if (!needle) return null;
  for (const slot of slots) {
    if (normalizeDialable(slot.remote) === needle) return slot;
  }
  return null;
}

/** @param {string} value @returns {string} */
function normalizeDialable(value) {
  return String(value || "")
    .toLowerCase()
    .replace(/[^\d+*#a-z]/g, "");
}

/**
 * The wait left in a slot's window ("43m"), or "" when it has run out
 * (or the slot lacks the stamp/window pair — the allowed verdict's
 * shape rides the same optional fields). Neutral labels, the agoLabel
 * convention.
 * @param {{last_at?: number, window?: number}} slot
 * @returns {string}
 */
export function cooldownRemainder(slot) {
  if (typeof slot.last_at !== "number" || typeof slot.window !== "number") {
    return "";
  }
  const left = slot.window - (Date.now() / 1000 - slot.last_at);
  if (left <= 0) return "";
  if (left < 3600) return `${Math.ceil(left / 60)}m`;
  return `${Math.ceil(left / 3600)}h`;
}

// agoLabel phrases a past unix stamp the draftAge way — neutral labels
// ("now", "47m", "2h") so the line reads identically in both languages.
/** @param {number} unixSeconds @returns {string} */
export function agoLabel(unixSeconds) {
  const seconds = Math.max(0, Date.now() / 1000 - unixSeconds);
  if (seconds < 60) return "now";
  if (seconds < 3600) return `${Math.floor(seconds / 60)}m`;
  if (seconds < 86400) return `${Math.floor(seconds / 3600)}h`;
  return `${Math.floor(seconds / 86400)}d`;
}
