// i18n (de/en). Static strings carry data-i18n attributes applied by
// applyI18n(); dynamic strings (status pill, call cards, errors) go
// through t().
//
// The event log stays English on purpose — it is operator-facing
// diagnostics and the runbook greps these phrasings.

const LANG_KEY = "pbx-lang";

const I18N = {
  en: {
    regState: "Registration state",
    signin: "Sign in to your extension",
    extension: "Extension",
    password: "Password",
    remember: "remember extension on this device (never the password)",
    connect: "Connect",
    signedInAs: "signed in as",
    signOut: "sign out",
    signOutTitle: "Unregister and sign out",
    dialPlaceholder: "Number, e.g. 1001, 2000, +441632960961",
    call: "Call",
    hold: "Hold",
    resume: "Resume",
    focus: "Focus",
    mute: "Mute",
    unmute: "Unmute",
    end: "End call",
    incoming: "Incoming call from",
    accept: "Accept",
    reject: "Reject",
    recentCalls: "Recent calls",
    eventLog: "Event log",
    offline: "offline",
    registered: "registered",
    regRejected: "registration rejected — check credentials",
    onHold: "on hold",
    inCall: "in call",
    calling: "calling…",
    ringing: "ringing…",
    ending: "ending…",
    transfer: "Transfer",
    transferPrompt: "transfer to",
    transferBlind: "Blind",
    transferAttended: "Attended",
    transferring: "transferring…",
    transferFailed: (detail) => `transfer failed: ${detail}`,
    transferNoPartner: "no second established call for an attended transfer",
    transferComplete: "transfer completed by the network",
    reconnectPreserved: (n) => `re-registered; ${n} call(s) preserved`,
    contacts: "Contacts",
    contactShared: "shared",
    contactSave: "save",
    contactRemove: "remove",
    voicemail: "Voicemail",
    vmRefresh: "Refresh",
    vmEmpty: "no messages",
    vmDelete: "delete",
    vmAuthFailed: "voicemail needs the operator API (not reachable)",
    vmNewCount: (n) => `${n} new`,
    connectionQuality: "Connection quality",
    iceNoMedia:
      "no media path yet — if the other side stays silent, TURN (ports 3478/5349 UDP) may be blocked",
    iceRelay:
      "media relays through TURN (expected behind strict NAT; adds a little latency)",
    iceSrflx: "direct path via STUN (router hole-punched)",
    iceHost: "direct local-network path",
    iceLoss: (lost) => `${lost} packets lost on receive — network congestion?`,
    iceFailed: "connection failed — media blocked between the networks",
    reconnecting: (delay, attempt) =>
      `reconnecting in ${delay}s (try ${attempt})`,
    loginError: (message) =>
      `Could not connect: ${message}. Check extension/password and that your browser trusts the server certificate.`,
    callFailed: (detail) => `call failed: ${detail}`,
    callEnded: (dur) => `call ended · ${dur}`,
    missedCall: (from) => `missed call from ${from}`,
    noActiveCall: "no active call",
    notConnected: "not connected — sign in first",
    audioBlocked:
      "browser blocked audio playback — click the page to enable sound",
    rejectedSecond: "second incoming call rejected (one call at a time)",
    dialEmpty: "enter a number to call",
    nothingDialable: "no dialable characters — use digits, letters, +, * or #",
    invalidDest: "invalid destination number",
    holdFailed: (detail) => `hold failed: ${detail}`,
    acceptFailed: (detail) => `could not accept the call: ${detail}`,
    hangupFailed: (detail) => `could not end the call: ${detail}`,
    dtmfFailed: (detail) => `could not send tone: ${detail}`,
    vmDeleteFailed: (detail) => `could not delete message: ${detail}`,
    vmPlayFailed: (detail) => `could not play message: ${detail}`,
  },
  de: {
    regState: "Registrierungsstatus",
    signin: "Anmeldung an Ihrer Nebenstelle",
    extension: "Nebenstelle",
    password: "Passwort",
    remember: "Nebenstelle auf diesem Gerät merken (niemals das Passwort)",
    connect: "Verbinden",
    signedInAs: "angemeldet als",
    signOut: "abmelden",
    signOutTitle: "Abmelden und Registrierung lösen",
    dialPlaceholder: "Nummer, z. B. 1001, 2000, +441632960961",
    call: "Anrufen",
    hold: "Halten",
    resume: "Fortsetzen",
    focus: "Aktivieren",
    mute: "Stumm",
    unmute: "Stumm aus",
    end: "Auflegen",
    incoming: "Eingehender Anruf von",
    accept: "Annehmen",
    reject: "Ablehnen",
    recentCalls: "Letzte Anrufe",
    eventLog: "Ereignisprotokoll",
    offline: "offline",
    registered: "registriert",
    regRejected: "Registrierung abgelehnt — Zugangsdaten prüfen",
    onHold: "gehalten",
    inCall: "im Gespräch",
    calling: "wird gewählt…",
    ringing: "klingelt…",
    ending: "wird beendet…",
    transfer: "Weiterleiten",
    transferPrompt: "weiterleiten an",
    transferBlind: "sofort",
    transferAttended: "Rückfrage",
    transferring: "wird weitergeleitet…",
    transferFailed: (detail) => `Weiterleitung fehlgeschlagen: ${detail}`,
    transferNoPartner:
      "kein zweiter bestehender Anruf für Rückfrage-Weiterleitung",
    transferComplete: "Weiterleitung vom Netz bestätigt",
    reconnectPreserved: (n) => `neu registriert; ${n} Gespräch(e) erhalten`,
    contacts: "Kontakte",
    contactShared: "gemeinsam",
    contactSave: "merken",
    contactRemove: "entfernen",
    voicemail: "Mailbox",
    vmRefresh: "Aktualisieren",
    vmEmpty: "keine Nachrichten",
    vmDelete: "löschen",
    vmAuthFailed: "Mailbox-API nicht erreichbar",
    vmNewCount: (n) => `${n} neu`,
    connectionQuality: "Verbindungsqualität",
    iceNoMedia:
      "noch kein Medienweg — falls die Gegenseite stumm bleibt, ist vermutlich TURN (UDP 3478/5349) blockiert",
    iceRelay:
      "Medien laufen über TURN (hinter strengem NAT normal; etwas mehr Laufzeit)",
    iceSrflx: "direkter Weg über STUN (Router-Lochbohrung)",
    iceHost: "direkter Weg im lokalen Netz",
    iceLoss: (lost) => `${lost} Pakete verloren — Netzüberlastung?`,
    iceFailed:
      "Verbindung fehlgeschlagen — Medien zwischen den Netzen blockiert",
    reconnecting: (delay, attempt) =>
      `Neuverbindung in ${delay}s (Versuch ${attempt})`,
    loginError: (message) =>
      `Verbindung fehlgeschlagen: ${message}. Prüfen Sie Nebenstelle/Passwort und ob Ihr Browser dem Serverzertifikat vertraut.`,
    callFailed: (detail) => `Anruf fehlgeschlagen: ${detail}`,
    callEnded: (dur) => `Anruf beendet · ${dur}`,
    missedCall: (from) => `verpasster Anruf von ${from}`,
    noActiveCall: "kein aktives Gespräch",
    notConnected: "nicht verbunden — bitte zuerst anmelden",
    audioBlocked:
      "Browser hat die Audiowiedergabe blockiert — Seite anklicken, um Ton zu aktivieren",
    rejectedSecond:
      "zweiter eingehender Anruf abgelehnt (ein Gespräch gleichzeitig)",
    dialEmpty: "bitte eine Nummer eingeben",
    nothingDialable:
      "keine wählbaren Zeichen — Ziffern, Buchstaben, +, * oder # verwenden",
    invalidDest: "ungültige Zielnummer",
    holdFailed: (detail) => `Halten fehlgeschlagen: ${detail}`,
    acceptFailed: (detail) => `Anruf konnte nicht angenommen werden: ${detail}`,
    hangupFailed: (detail) => `Anruf konnte nicht beendet werden: ${detail}`,
    dtmfFailed: (detail) => `Ton konnte nicht gesendet werden: ${detail}`,
    vmDeleteFailed: (detail) =>
      `Nachricht konnte nicht gelöscht werden: ${detail}`,
    vmPlayFailed: (detail) =>
      `Nachricht konnte nicht abgespielt werden: ${detail}`,
  },
};

let lang =
  localStorage.getItem(LANG_KEY) ||
  ((navigator.language || "en").toLowerCase().startsWith("de") ? "de" : "en");

export function t(key) {
  const table = I18N[lang] || I18N.en;
  return table[key] !== undefined ? table[key] : I18N.en[key];
}

export function getLang() {
  return lang;
}

export function setLang(next) {
  lang = next;
  localStorage.setItem(LANG_KEY, lang);
  // Tell the server: tabs and SSE fragments render in the same language
  // (cookie read by every request; no reload — the island never unloads).
  document.cookie = `wp-lang=${lang}; path=/; samesite=strict; max-age=31536000`;
}

export function applyI18n() {
  document.documentElement.lang = lang;
  document.querySelectorAll("[data-i18n]").forEach((el) => {
    const value = t(el.dataset.i18n);
    if (typeof value === "string") el.textContent = value;
  });
  document.querySelectorAll("[data-i18n-placeholder]").forEach((el) => {
    el.placeholder = t(el.dataset.i18nPlaceholder);
  });
  document.querySelectorAll("[data-i18n-title]").forEach((el) => {
    el.title = t(el.dataset.i18nTitle);
  });
}
