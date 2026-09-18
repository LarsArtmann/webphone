// Runtime config injected by the serving PBX as window.PBX_CONFIG
// (config.js, rendered at boot because it carries short-lived TURN
// credentials). Every key is optional: without a config the phone
// degrades to same-host defaults.
const config = window.PBX_CONFIG || {};

export const sipDomain = config.sipDomain || location.hostname;

export const websocketUrl = `wss://${
  location.host
}${config.websocketPath || "/sip"}`;

export const iceServers = Array.isArray(config.iceServers)
  ? config.iceServers
  : [];

export const phoneApiEnabled = config.phoneApi === true;

export const sharedContacts = Array.isArray(config.contacts)
  ? config.contacts
  : [];
