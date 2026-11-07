// Runtime config injected by the serving PBX as window.PBX_CONFIG
// (config.js, rendered at boot because it carries short-lived TURN
// credentials). Every key is optional: without a config the phone
// degrades to same-host defaults.
const config = window.PBX_CONFIG || {};

export const sipDomain = config.sipDomain || location.hostname;

// The SIP WebSocket scheme follows the page protocol: the stack serves
// https (caddy terminates TLS, /sip rides wss), while a bare-HTTP dev
// boot (the loopback quickstart, scripts/ui-capture.py) serves ws —
// hardcoding wss left the island unable to connect outside the stack.
export const websocketUrl = `${location.protocol === "https:" ? "wss" : "ws"}://${
  location.host
}${config.websocketPath || "/sip"}`;

export const iceServers = Array.isArray(config.iceServers)
  ? config.iceServers
  : [];

export const phoneApiEnabled = config.phoneApi === true;

// Derived SIP credentials (credential brokering): the server mints the
// session BEFORE the REGISTER and answers a session-derived credential;
// the directory password never reaches the browser. Off by default.
export const sipCredentialMode = config.sipCredential === true;

// Optional Ledger CRM integration: when the server has one configured it
// also accepts post-call reports on /api/calls (session-scoped). Off by
// default; the server still answers 204 for a stale island that reports
// anyway.
export const crmEnabled = config.crm === true;

// Optional speech-to-text seam: when the server has a provider configured
// it accepts audio on /api/transcribe and the island shows the transcribe
// affordances (live calls, voicemail). Off by default.
export const asrEnabled = config.asr === true;

// Where the provider processes the audio (config-honest trust label,
// M17): "local" names a loopback/private provider host the operator
// runs, "remote" an external one. Empty when the seam is off.
export const asrLocality =
  config.asrLocality === "local" || config.asrLocality === "remote"
    ? config.asrLocality
    : "";

// Optional text-agent seam (post-call summarize, M12): when the server
// has the review API configured, the ended call card offers "Summarize
// → draft SMS". Off by default — the post-call row carries only Close.
export const agentEnabled = config.agent === true;

export const sharedContacts = Array.isArray(config.contacts)
  ? config.contacts
  : [];

// The contact sources behind the dial typeahead AND the ring surfaces
// (F7): shared contacts ship in the config, per-extension personal
// contacts are fed by panels.js once the session exists. Both resolve
// here so a name resolves in exactly one way everywhere.
/** @type {PBXSharedContact[]} */
const personalContacts = [];

/** @param {unknown} contacts */
export function setExtraContacts(contacts) {
  const list = Array.isArray(contacts) ? contacts : [];
  personalContacts.length = 0;
  for (const contact of list) {
    if (contact && contact.number) personalContacts.push(contact);
  }
}

/** @returns {PBXSharedContact[]} */
export function contactSource() {
  const personalNumbers = new Set(personalContacts.map((c) => c.number));
  return [
    ...personalContacts,
    ...sharedContacts.filter((c) => !personalNumbers.has(c.number)),
  ];
}

/**
 * The contact for a number across the contact sources (personal wins
 * over shared); null when unknown. Exact match first, then a
 * digits-only match — "+49…" and "030 …" forms of one number resolve
 * to the same card.
 * @param {string} number
 * @returns {PBXSharedContact | null}
 */
export function findContact(number) {
  const wanted = String(number || "").trim();
  if (!wanted) return null;
  const source = contactSource();
  for (const contact of source) {
    if (String(contact.number) === wanted) return contact;
  }
  const digits = wanted.replace(/\D/g, "");
  if (!digits) return null;
  for (const contact of source) {
    if (String(contact.number || "").replace(/\D/g, "") === digits) {
      return contact;
    }
  }
  return null;
}

/**
 * The display name for a number; "" when unknown — an unknown caller
 * renders raw digits, which is the honest fallback.
 * @param {string} number
 * @returns {string}
 */
export function contactName(number) {
  const contact = findContact(number);
  return contact ? String(contact.name || "") : "";
}

/**
 * "Name (digits)" for a known caller, the raw digits otherwise — the
 * one shape every ring surface (banner, queue slots, the assertive
 * announce, the missed-call toast, dock chips) shows. The digits shown
 * are the CONTACT's stored form: it is the display the user wrote.
 * @param {string} number
 * @returns {string}
 */
export function callerDisplay(number) {
  const wanted = String(number || "").trim();
  if (!wanted) return "";
  const contact = findContact(wanted);
  if (contact && contact.name) return `${contact.name} (${contact.number})`;
  return wanted;
}
