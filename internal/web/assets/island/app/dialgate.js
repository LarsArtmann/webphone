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
