// i18n (de/en). Static strings carry data-i18n attributes applied by
// applyI18n(); dynamic strings (status pill, call cards, errors) go
// through t().
//
// The event log stays English on purpose — it is operator-facing
// diagnostics and the runbook greps these phrasings.

const LANG_KEY = "pbx-lang";

// The message tables are a heterogeneous bag: plain strings and
// formatter functions (the call site writes t("key")(arg)). defineMessages
// is an identity function that exists only to give the literal a contextual
// signature — that is what types the formatter parameters (no implicit any)
// while KEEPING the key union, which an index-signature annotation would
// flatten to string.
/**
 * @template {Record<string, string | ((...args: any[]) => string)>} T
 * @param {T} table
 * @returns {T}
 */
const defineMessages = (table) => table;

/** One dictionary entry: a finished string or the formatter that builds it. */
/** @typedef {string | ((...args: any[]) => string)} IslandMessage */
/** The two shipped languages. */
/** @typedef {"en" | "de"} IslandLang */
/** Every key the tables carry (derived from `en`, never hand-listed). */
/** @typedef {keyof typeof I18N.en} IslandI18nKey */

// Exported for the node:test suite (island-tests/): the en/de parity test
// enumerates keys exactly like the Go-side i18n sync test does.
export const I18N = {
  en: defineMessages({
    regState: "Registration state",
    signin: "Sign in to your extension",
    extension: "Extension",
    password: "Password",
    remember: "remember extension on this device (never the password)",
    connect: "Connect",
    signedInAs: "signed in as",
    signOut: "sign out",
    signOutTitle: "Unregister and sign out",
    dialPlaceholder: "Number or extension",
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
    advancedSection: "Diagnostics",
    deviceTest: "Test audio devices",
    offline: "offline",
    registered: "registered",
    offlineBanner: "You are offline — calls cannot be made or received.",
    browserOffline: "browser offline",
    regRejected: "registration rejected — check credentials",
    regRebuilding: "rebuilding connection…",
    onHold: "on hold",
    inCall: "in call",
    volume: "Volume",
    dockLabel: "Other calls",
    focusChanged: (target) => `Switched to ${target}`,
    callHolding: "holding…",
    callResuming: "resuming…",
    calling: "calling…",
    ringing: "ringing…",
    callRinging: (target) => `ringing ${target}`,
    callEstablished: (target) => `connected with ${target}`,
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
    contactSaved: (n) => `contact ${n} saved`,
    contactSaveFailed: (detail) =>
      `contact save failed (${detail}); kept locally`,
    contactsMigrated: (n) => `imported ${n} local contact(s) to the server`,
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
    passkeySection: "Sign in with a passkey",
    passkeySignin: "Sign in with passkey",
    passkeyBreakglass: "Use extension and password instead",
    ssoSection: "Sign in with Pocket ID",
    ssoSignin: "Continue to Pocket ID",
    ssoPasskeyHint:
      "Passkeys and every other sign-in method your account has work in Pocket ID.",
    ssoErrorUnmapped:
      "Your Pocket ID account has no phone mapping here — ask the operator to add your email address.",
    ssoErrorDirectory:
      "Your phone account's password could not be verified — ask the operator to refresh it, then try again.",
    ssoErrorUnverified:
      "Your Pocket ID email is not verified — verify it in Pocket ID, then try again.",
    ssoErrorGroups:
      "Your Pocket ID account is missing the required access group for this phone — ask the operator to add it, then try again.",
    ssoErrorStale: "That sign-in attempt expired — please start again.",
    ssoErrorConfig:
      "Sign-in is not configured correctly on the server — the operator finds the details in the journal.",
    ssoErrorUnavailable:
      "Pocket ID is not reachable right now — try again in a minute, or use extension and password.",
    ssoErrorFailed:
      "Sign-in with Pocket ID failed — try again, or use extension and password.",
    enrollVerified: (email) =>
      `Token verified for ${email}. Follow your device to create the passkey.`,
    enrollSuccess: "Passkey created. Go to the main page and sign in with it.",
    enrollFailed: (status) =>
      `Enrollment failed (HTTP ${status}) — the token may be invalid, expired or already used; ask for a new link if this repeats.`,
    enrollDismissed:
      "Passkey creation was cancelled — submit the token again to retry.",
    enrollNetFailed: "Network error — try again.",
    enrollTokenMissing: "Paste the token from your enrollment link first.",
    callFailed: (detail) => `call failed: ${detail}`,
    callEnded: (dur) => `call ended · ${dur}`,
    missedCall: (from) => `missed call from ${from}`,
    incomingCall: (from) => `Incoming call from ${from}`,
    noActiveCall: "no active call",
    notConnected: "not connected — sign in first",
    audioBlocked:
      "browser blocked audio playback — click the page to enable sound",
    dialEmpty: "enter a number to call",
    nothingDialable: "no dialable characters — use digits, letters, +, * or #",
    invalidDest: "invalid destination number",
    holdFailed: (detail) => `hold failed: ${detail}`,
    resumeFailed: (detail) => `resume failed: ${detail}`,
    acceptFailed: (detail) => `could not accept the call: ${detail}`,
    hangupFailed: (detail) => `could not end the call: ${detail}`,
    dtmfFailed: (detail) => `could not send tone: ${detail}`,
    vmDeleteFailed: (detail) => `could not delete message: ${detail}`,
    vmPlayFailed: (detail) => `could not play message: ${detail}`,
    sessionFailed401:
      "The server rejected the tab sign-in — check extension and password.",
    sessionThrottled: "Too many attempts — wait a moment, then try again.",
    sessionFailed: (status) =>
      `Tab sign-in failed (HTTP ${status}) — calls keep working, tabs stay locked.`,
    sessionNetFailed:
      "Could not reach the server for tab sign-in — calls keep working.",
    sseDropped:
      "Live tab updates lost — check the connection; calls keep working.",
    sseRestored: "Live tab updates restored.",
    ssePillLive: "Live tab updates: connected",
    ssePillDown: "Live tab updates: not connected",
    ssePillRetry: "Live tab updates: reconnecting…",
    resuming: "resuming session…",
    resumeRejected: (message) =>
      `Your saved sign-in was rejected (${message}) — the password may have changed. Sign in again.`,
    crmLogFailed: (detail) =>
      `Call not recorded in the CRM (${detail}) — you can log it manually.`,
    typeaheadLabel: "Contact suggestions",
    audioOutput: "Speaker",
    audioDefault: "System default",
    transcribe: "Transcribe",
    transcribeStop: "Stop transcription",
    transcribeBusy: "Transcribing…",
    transcribeEmpty: "No speech detected",
    copyTranscript: "Copy transcript",
    clearTranscript: "Clear",
    transcriptCleared: "Live transcript cleared.",
    copied: "Copied to clipboard.",
    copyFailed: "Could not copy — select the text manually.",
    transcribeUnsupported:
      "Live transcription is not supported in this browser",
    transcribeFailed: (detail) => `transcription failed: ${detail}`,
    transcribeUnavailable:
      "Live transcription keeps failing — the provider may be unreachable.",
    transcribeStarted: "Live transcription started",
    transcribeStopped: "Live transcription stopped",
    transcriptDelta: (text) => `Transcript: ${text}`,
    transcriptSavedToHistory: (n) =>
      `Saved to History — ${n} segment${n === 1 ? "" : "s"}.`,
    transcriptSaveFailed: (detail) =>
      `transcript not saved (${detail}) — it stays on the call card only`,
    captionsTitle: "Live transcript",
    captionsLocal: "audio processed on this server",
    captionsRemote: "audio sent to the configured provider",
    ended: "Ended",
    close: "Close",
    summarizeDraft: "Summarize → draft SMS",
    summarizing: "Summarizing…",
    summarizeReady: "Follow-up draft ready in Messages.",
    summarizeDuplicate: "Already drafted — see Messages.",
    summarizeNoTranscript: "No transcript for that call.",
    summarizeFailed: "The AI agent is unreachable right now.",
    summarizeOffline:
      "You are offline — the draft request never left the browser.",
    reviewQueued: "Review queued — the draft follows in Messages.",
    callMissed: "Missed call — the caller gave up before it was answered.",
    callBusy: "Busy — the line was not free.",
    callNotFound: "Not reachable — the number does not exist.",
    callBlocked: "Call not allowed — blocked by the far end.",
    callUnreachable: "Unavailable — the far end did not answer.",
    callCancelled: "Cancelled before it connected.",
    callNoAnswer: "No answer in time.",
    callFailedCode: (code) => `Call failed (code ${code}).`,
    callFailedUnknown: "The call failed before it connected.",
    summarizeSettled: "Draft in Messages",
    reviewQueued: "Review queued",
    dispositionAsk: "How did this call end?",
    dispositionAnswered: "Answered",
    dispositionVoicemail: "Voicemail",
    dispositionIvr: "IVR",
    dispositionSaved: (label) => `Verdict saved: ${label}`,
    dispositionFailed: "Saving the verdict failed — try again.",
  }),
  de: defineMessages({
    regState: "Registrierungsstatus",
    signin: "Anmeldung an Ihrer Nebenstelle",
    extension: "Nebenstelle",
    password: "Passwort",
    remember: "Nebenstelle auf diesem Gerät merken (niemals das Passwort)",
    connect: "Verbinden",
    passkeySection: "Mit Passkey anmelden",
    passkeySignin: "Mit Passkey anmelden",
    passkeyBreakglass: "Stattdessen Nebenstelle und Passwort verwenden",
    ssoSection: "Mit Pocket ID anmelden",
    ssoSignin: "Weiter zu Pocket ID",
    ssoPasskeyHint:
      "Passkeys und alle anderen Anmeldeverfahren Ihres Kontos funktionieren in Pocket ID.",
    ssoErrorUnmapped:
      "Ihr Pocket ID-Konto hat hier keine Telefon-Zuordnung — bitten Sie den Betreiber, Ihre E-Mail-Adresse einzutragen.",
    ssoErrorDirectory:
      "Das Passwort Ihres Telefon-Kontos konnte nicht verifiziert werden — bitten Sie den Betreiber, es zu erneuern, und versuchen Sie es erneut.",
    ssoErrorUnverified:
      "Ihre Pocket ID-E-Mail ist nicht bestätigt — bestätigen Sie sie in Pocket ID und versuchen Sie es erneut.",
    ssoErrorGroups:
      "Ihrem Pocket ID-Konto fehlt die erforderliche Zugriffsgruppe für dieses Telefon — bitten Sie den Betreiber, sie einzutragen, und versuchen Sie es erneut.",
    ssoErrorStale: "Dieser Anmeldeversuch ist abgelaufen — bitte neu starten.",
    ssoErrorConfig:
      "Die Anmeldung ist serverseitig nicht korrekt konfiguriert — der Betreiber findet die Details im Journal.",
    ssoErrorUnavailable:
      "Pocket ID ist gerade nicht erreichbar — versuchen Sie es in einer Minute erneut oder nutzen Sie Nebenstelle und Passwort.",
    ssoErrorFailed:
      "Anmeldung mit Pocket ID fehlgeschlagen — erneut versuchen oder Nebenstelle und Passwort verwenden.",
    enrollVerified: (email) =>
      `Token verifiziert für ${email}. Folgen Sie Ihrem Gerät, um den Passkey zu erstellen.`,
    enrollSuccess:
      "Passkey erstellt. Gehen Sie zur Hauptseite und melden Sie sich damit an.",
    enrollFailed: (status) =>
      `Hinterlegung fehlgeschlagen (HTTP ${status}) — das Token ist möglicherweise ungültig, abgelaufen oder bereits benutzt; fordern Sie im Zweifel einen neuen Link an.`,
    enrollDismissed:
      "Passkey-Erstellung abgebrochen — Token erneut absenden, um es noch einmal zu versuchen.",
    enrollNetFailed: "Netzwerkfehler — erneut versuchen.",
    enrollTokenMissing: "Fügen Sie zuerst das Token aus Ihrem Link ein.",
    signedInAs: "angemeldet als",
    signOut: "abmelden",
    signOutTitle: "Abmelden und Registrierung lösen",
    dialPlaceholder: "Nummer oder Durchwahl",
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
    advancedSection: "Diagnose",
    deviceTest: "Audiogeräte testen",
    offline: "offline",
    registered: "registriert",
    offlineBanner: "Sie sind offline — Anrufe sind nicht möglich.",
    browserOffline: "Browser offline",
    regRejected: "Registrierung abgelehnt — Zugangsdaten prüfen",
    regRebuilding: "Verbindung wird neu aufgebaut…",
    onHold: "gehalten",
    inCall: "im Gespräch",
    volume: "Lautstärke",
    dockLabel: "Weitere Anrufe",
    focusChanged: (target) => `Umgeschaltet auf ${target}`,
    callHolding: "wird gehalten…",
    callResuming: "wird fortgesetzt…",
    calling: "wird gewählt…",
    ringing: "klingelt…",
    callRinging: (target) => `klingelt bei ${target}`,
    callEstablished: (target) => `verbunden mit ${target}`,
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
    contactSaved: (n) => `Kontakt ${n} gespeichert`,
    contactSaveFailed: (detail) =>
      `Kontakt speichern fehlgeschlagen (${detail}); lokal behalten`,
    contactsMigrated: (n) => `${n} lokale Kontakte auf den Server übernommen`,
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
    incomingCall: (from) => `Eingehender Anruf von ${from}`,
    noActiveCall: "kein aktives Gespräch",
    notConnected: "nicht verbunden — bitte zuerst anmelden",
    audioBlocked:
      "Browser hat die Audiowiedergabe blockiert — Seite anklicken, um Ton zu aktivieren",
    dialEmpty: "bitte eine Nummer eingeben",
    nothingDialable:
      "keine wählbaren Zeichen — Ziffern, Buchstaben, +, * oder # verwenden",
    invalidDest: "ungültige Zielnummer",
    holdFailed: (detail) => `Halten fehlgeschlagen: ${detail}`,
    resumeFailed: (detail) => `Fortsetzen fehlgeschlagen: ${detail}`,
    acceptFailed: (detail) => `Anruf konnte nicht angenommen werden: ${detail}`,
    hangupFailed: (detail) => `Anruf konnte nicht beendet werden: ${detail}`,
    dtmfFailed: (detail) => `Ton konnte nicht gesendet werden: ${detail}`,
    vmDeleteFailed: (detail) =>
      `Nachricht konnte nicht gelöscht werden: ${detail}`,
    vmPlayFailed: (detail) =>
      `Nachricht konnte nicht abgespielt werden: ${detail}`,
    sessionFailed401:
      "Die Server-Anmeldung wurde abgelehnt — Nebenstelle und Passwort prüfen.",
    sessionThrottled:
      "Zu viele Versuche — bitte kurz warten und erneut versuchen.",
    sessionFailed: (status) =>
      `Server-Anmeldung fehlgeschlagen (HTTP ${status}) — Anrufe funktionieren weiter, Tabs bleiben gesperrt.`,
    sessionNetFailed:
      "Server für die Tab-Anmeldung nicht erreichbar — Anrufe funktionieren weiter.",
    sseDropped:
      "Live-Aktualisierungen verloren — Verbindung prüfen; Anrufe funktionieren weiter.",
    sseRestored: "Live-Aktualisierungen wiederhergestellt.",
    ssePillLive: "Live-Aktualisierungen: verbunden",
    ssePillDown: "Live-Aktualisierungen: nicht verbunden",
    ssePillRetry: "Live-Aktualisierungen: verbinde neu…",
    resuming: "Sitzung wird fortgesetzt…",
    resumeRejected: (message) =>
      `Gespeicherte Anmeldung abgelehnt (${message}) — das Passwort hat sich möglicherweise geändert. Bitte erneut anmelden.`,
    crmLogFailed: (detail) =>
      `Anruf nicht im CRM erfasst (${detail}) — bitte manuell nachtragen.`,
    typeaheadLabel: "Kontaktvorschläge",
    audioOutput: "Ausgabe",
    audioDefault: "Systemstandard",
    transcribe: "Transkribieren",
    transcribeStop: "Transkription stoppen",
    transcribeBusy: "Transkribiere…",
    transcribeEmpty: "Keine Sprache erkannt",
    copyTranscript: "Transkript kopieren",
    clearTranscript: "Leeren",
    transcriptCleared: "Live-Transkript geleert.",
    copied: "In die Zwischenablage kopiert.",
    copyFailed: "Kopieren fehlgeschlagen — Text manuell auswählen.",
    transcribeUnsupported:
      "Live-Transkription wird in diesem Browser nicht unterstützt",
    transcribeFailed: (detail) => `Transkription fehlgeschlagen: ${detail}`,
    transcribeUnavailable:
      "Live-Transkription schlägt wiederholt fehl — der Anbieter ist womöglich nicht erreichbar.",
    transcribeStarted: "Live-Transkription gestartet",
    transcribeStopped: "Live-Transkription gestoppt",
    transcriptDelta: (text) => `Transkript: ${text}`,
    transcriptSavedToHistory: (n) =>
      `In der Historie gespeichert — ${n} Segment${n === 1 ? "" : "e"}.`,
    transcriptSaveFailed: (detail) =>
      `Transkript nicht gespeichert (${detail}) — es bleibt nur auf der Anrufkarte`,
    captionsTitle: "Live-Transkript",
    captionsLocal: "Audio wird auf diesem Server verarbeitet",
    captionsRemote: "Audio geht an den konfigurierten Anbieter",
    ended: "Beendet",
    close: "Schließen",
    summarizeDraft: "Zusammenfassen → SMS-Entwurf",
    summarizing: "Fasse zusammen…",
    summarizeReady: "Folge-Entwurf in Nachrichten bereit.",
    summarizeDuplicate: "Bereits entworfen — siehe Nachrichten.",
    summarizeNoTranscript: "Keine Mitschrift für diesen Anruf.",
    summarizeFailed: "Der KI-Agent ist gerade nicht erreichbar.",
    summarizeOffline:
      "Sie sind offline — die Entwurfsanfrage hat den Browser nie verlassen.",
    reviewQueued: "Prüfung eingereiht — der Entwurf folgt in Nachrichten.",
    callMissed:
      "Verpasster Anruf — der Anrufer hat aufgelegt, bevor jemand geantwortet hat.",
    callBusy: "Besetzt — die Leitung war nicht frei.",
    callNotFound: "Nicht erreichbar — die Nummer existiert nicht.",
    callBlocked: "Anruf nicht erlaubt — von der Gegenseite blockiert.",
    callUnreachable: "Nicht verfügbar — die Gegenseite hat nicht geantwortet.",
    callCancelled: "Abgebrochen, bevor die Verbindung stand.",
    callNoAnswer: "Keine rechtzeitige Antwort.",
    callFailedCode: (code) => `Anruf fehlgeschlagen (Code ${code}).`,
    callFailedUnknown: "Der Anruf ist vor dem Verbinden fehlgeschlagen.",
    summarizeSettled: "Entwurf in Nachrichten",
    reviewQueued: "Einschätzung vorgemerkt",
    dispositionAsk: "Wie endete dieses Gespräch?",
    dispositionAnswered: "Angenommen",
    dispositionVoicemail: "Voicemail",
    dispositionIvr: "Automat",
    dispositionSaved: (label) => `Einschätzung gespeichert: ${label}`,
    dispositionFailed:
      "Speichern der Einschätzung fehlgeschlagen — bitte erneut versuchen.",
  }),
};

// The wp-lang cookie is what the SERVER already renders the shell, tabs
// and <html lang> from (pages.go) — read it FIRST so a cookie≠navigator
// user gets one language everywhere: German tabs under a German phone,
// screen readers announcing with the right phonemes. localStorage only
// leads before the first setLang (which is what writes the cookie).
/** @returns {IslandLang | ""} */
function cookieLang() {
  const match = document.cookie.match(/(?:^|;\s*)wp-lang=(de|en)\b/);
  return match ? /** @type {IslandLang} */ (match[1]) : "";
}

/** @type {IslandLang} */
let lang =
  cookieLang() ||
  /** @type {IslandLang} */ (localStorage.getItem(LANG_KEY)) ||
  ((navigator.language || "en").toLowerCase().startsWith("de") ? "de" : "en");

// The table value is a finished string OR the formatter for it, and the
// call sites write both t("key") and t("key")(arg). Returning the union
// would force a cast at all 155 call sites for no safety (the KEY is what
// the type gate can check); returning any keeps that split at the seam.
/**
 * @param {IslandI18nKey} key
 * @returns {any}
 */
export function t(key) {
  /** @type {typeof I18N.en} */
  const table = I18N[lang] || I18N.en;
  return table[key] !== undefined ? table[key] : I18N.en[key];
}

/** @returns {IslandLang} */
export function getLang() {
  return lang;
}

/** @param {IslandLang} next */
export function setLang(next) {
  lang = next;
  localStorage.setItem(LANG_KEY, lang);
  // Tell the server: tabs and SSE fragments render in the same language
  // (cookie read by every request; no reload — the island never unloads).
  document.cookie = `wp-lang=${lang}; path=/; samesite=strict; max-age=31536000`;
}

export function applyI18n() {
  document.documentElement.lang = lang;
  document.querySelectorAll("[data-i18n]").forEach((node) => {
    const el = /** @type {HTMLElement} */ (node);
    const value = t(/** @type {IslandI18nKey} */ (el.dataset.i18n));
    if (typeof value === "string") el.textContent = value;
  });
  document.querySelectorAll("[data-i18n-placeholder]").forEach((node) => {
    const el = /** @type {HTMLInputElement} */ (node);
    el.placeholder = t(
      /** @type {IslandI18nKey} */ (el.dataset.i18nPlaceholder),
    );
  });
  document.querySelectorAll("[data-i18n-title]").forEach((node) => {
    const el = /** @type {HTMLElement} */ (node);
    el.title = t(/** @type {IslandI18nKey} */ (el.dataset.i18nTitle));
  });
}
