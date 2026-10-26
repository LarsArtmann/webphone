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
export const asrLocality = config.asrLocality === "local" || config.asrLocality === "remote" ? config.asrLocality : "";

export const sharedContacts = Array.isArray(config.contacts)
  ? config.contacts
  : [];
