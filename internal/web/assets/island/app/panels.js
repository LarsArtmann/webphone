// The panels below the call area: call history (local + server CDR),
// contacts (shared config + personal contacts on the SERVER store, one
// home with the Contacts tab) and voicemail (phone API). All server
// access rides the extension's session via authedFetch — there is no
// second login.

import {
  asrEnabled,
  crmEnabled,
  phoneApiEnabled,
  sharedContacts,
} from "./config.js";
import {
  authedFetch,
  authHeaderValue,
  getCredentials,
  noteThrottled,
} from "./auth.js";
import { t } from "./i18n.js";
import { transcribeNow } from "./transcribe.js";
import { initDialTypeahead, setExtraContacts } from "./typeahead.js";
import { announce, dialFromUi, els, errorMessage, log } from "./ui.js";

// Inline SVG icons for JS-created buttons (the island twins of the
// server's templ-components icons) — platform emoji rendered
// differently per OS and read as decoration.
/** @param {string} path @returns {string} */
const iconSvg = (path) =>
  `<svg class="wp-icon" viewBox="0 0 24 24" fill="currentColor" aria-hidden="true"><path d="${path}"></path></svg>`;
const playIcon = () => iconSvg("M8 5.14v13.72L19 12 8 5.14z");
const trashIcon = () =>
  iconSvg(
    "M9 3a1 1 0 0 0-1 1H5.5a1 1 0 0 0 0 2H6l.8 12.1A2 2 0 0 0 8.8 20h6.4a2 2 0 0 0 2-1.9L18 6h.5a1 1 0 1 0 0-2H16a1 1 0 0 0-1-1H9zm0 5a1 1 0 0 1 2 0v7a1 1 0 1 1-2 0V8zm4 0a1 1 0 1 1 2 0v7a1 1 0 1 1-2 0V8z",
  );

// --- call history ----------------------------------------------------------

const HISTORY_KEY = "pbx-history";
const HISTORY_MAX = 20;
const CONTACTS_KEY = "pbx-contacts";
const CONTACTS_MAX = 50;

// Server-backed rows (CDR), fetched when the phone API is enabled; kept
// separate from the local session history so reloads never lose either.
/** @type {IslandServerHistoryRow[]} */
let serverHistory = [];

/** @returns {IslandHistoryEntry[]} */
export function readHistory() {
  try {
    const raw = JSON.parse(localStorage.getItem(HISTORY_KEY) || "[]");
    return Array.isArray(raw) ? raw : [];
  } catch {
    return [];
  }
}

/** @param {IslandHistoryEntry} entry */
export function recordHistory(entry) {
  const list = [entry, ...readHistory()].slice(0, HISTORY_MAX);
  localStorage.setItem(HISTORY_KEY, JSON.stringify(list));
  renderHistory();
}

// recordCrmCall reports one finished call to the CRM integration. The
// server resolves the number against the CRM and journals the call on the
// matching contact; unknown numbers are dropped server-side by design (the
// integration never mints contacts). A failure surfaces as a warn toast —
// the user expects the call in their CRM and must not lose it silently.
// Every report carries a fresh UUID key so a browser-level retry or
// double-fire replays the SAME key and the server journals the call once.
// The disposition reports only what this side observed: a call that never
// carried media is "missed"; one that did is "unknown" — media cannot
// distinguish a human from a voicemail greeting, and claiming "answered"
// from connection state journaled voicemail as answered (issue #6).
// An operator confirm upgrades unknown to a specific resolution.
/**
 * @param {{dir: "in" | "out", target: string, dur: number, established: boolean}} call
 * @returns {Promise<void>}
 */
export async function recordCrmCall({ dir, target, dur, established }) {
  if (!crmEnabled) return;
  try {
    const res = await authedFetch("/api/calls", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        number: target,
        direction: dir,
        seconds: dur > 0 ? dur : 0,
        disposition: established ? "unknown" : "missed",
        key: crypto.randomUUID(),
      }),
    });
    if (!res.ok) {
      const detail = `HTTP ${res.status}`;
      log(`crm call log failed: ${detail}`, "warn");
      announce(t("crmLogFailed")(detail), "warn");
    }
  } catch (err) {
    log(`crm call log failed: ${errorMessage(err)}`, "warn");
    announce(t("crmLogFailed")(errorMessage(err)), "warn");
  }
}

export async function refreshServerHistory() {
  if (!phoneApiEnabled || !authHeaderValue()) return;
  try {
    const res = await authedFetch("/phone-api/history?limit=20");
    if (!res.ok) return;
    const data = await res.json();
    serverHistory = Array.isArray(data.entries) ? data.entries : [];
    renderHistory();
  } catch (err) {
    log(`server history unavailable: ${errorMessage(err)}`, "error");
  }
}

/**
 * @param {{dir: string, target: string, whenText: string, durText: string, number: string}} row
 * @returns {HTMLLIElement}
 */
function makeHistoryRow({ dir, target, whenText, durText, number }) {
  const li = document.createElement("li");
  const dirEl = document.createElement("span");
  dirEl.className = "dir";
  dirEl.textContent = dir;
  const targetEl = document.createElement("strong");
  targetEl.textContent = target;
  const when = document.createElement("span");
  when.className = "when";
  when.textContent = whenText;
  const dur = document.createElement("span");
  dur.className = "when";
  dur.textContent = durText;
  li.append(dirEl, targetEl, when, dur);
  // Redial + save-as-contact: dialing gets fast, contacts stay local.
  const redial = document.createElement("button");
  redial.className = "ghost small redial";
  redial.textContent = "↻";
  redial.title = number;
  redial.addEventListener("click", () => dialFromUi(number));
  const save = document.createElement("button");
  save.className = "ghost small save";
  save.textContent = "☆";
  save.title = t("contactSave");
  save.addEventListener("click", () => {
    // saveContact re-renders when the server answers (async now).
    saveContact(number, number);
  });
  li.append(redial, save);
  return li;
}

export function renderHistory() {
  const list = readHistory();
  els.historyWrap.hidden = list.length === 0 && serverHistory.length === 0;
  const localRows = list.map((entry) =>
    makeHistoryRow({
      dir: entry.dir === "in" ? "←" : "→",
      target: entry.target,
      whenText: new Date(entry.at).toLocaleString(),
      durText:
        entry.dur > 0
          ? `${Math.floor(entry.dur / 60)}:${String(entry.dur % 60).padStart(2, "0")}`
          : "—",
      number: entry.target,
    }),
  );
  // The proxy forwards FreeSWITCH's CDR fields verbatim, so any of them can
  // be ABSENT (a hung-up-before-answer row carries no billsec, a rejected
  // leg may carry no caller id). Each fallback below is what the operator
  // reads instead of the word "undefined".
  const serverRows = serverHistory.map((row) => {
    const inbound = row.context === "public";
    const number =
      (inbound ? row.caller_id_number : row.destination_number) || "—";
    const billsec = row.billsec || 0;
    return makeHistoryRow({
      dir: inbound ? "←" : "→",
      target: number,
      whenText: row.start || "—",
      durText:
        billsec > 0
          ? `${Math.floor(billsec / 60)}:${String(billsec % 60).padStart(2, "0")}`
          : "—",
      number,
    });
  });
  els.history.replaceChildren(...localRows, ...serverRows);
}

// --- contacts (server home: /api/contacts) ----------------------------------
//
// Personal contacts live in the server's per-extension store — the same
// home the Contacts tab renders from — so they survive browser loss and
// never fork. Shared contacts still come from config (they render
// pre-login). The legacy localStorage list ("pbx-contacts") is imported
// once after login and cleared only after the server accepted every
// row; a failed import keeps the local rows rendering and retries on
// the next login (the store upserts by number, so re-import cannot
// duplicate).

/** @returns {IslandContactRow[]} */
function readPersonalContacts() {
  try {
    const raw = JSON.parse(localStorage.getItem(CONTACTS_KEY) || "[]");
    return Array.isArray(raw) ? raw : [];
  } catch {
    return [];
  }
}

/** @type {IslandContactRow[]} */
let personalContacts = [];
// Seeded once at load; emptied only by a confirmed import.
/** @type {IslandContactRow[]} */
let legacyContacts = readPersonalContacts();

export function clearContacts() {
  personalContacts = [];
  legacyContacts = [];
}

// loadContacts runs when the server session exists (the island's login
// fires wp:session-opened after the cookie is minted and the fresh CSRF
// token adopted — POSTing earlier would ride a dead token).
export async function loadContacts() {
  try {
    const res = await authedFetch("/api/contacts");
    if (!res.ok) return;
    const data = await res.json();
    personalContacts = Array.isArray(data.personal) ? data.personal : [];
    renderContacts();
    await migrateLegacyContacts();
  } catch (err) {
    log(`server contacts unavailable: ${errorMessage(err)}`, "error");
  }
}
document.addEventListener("wp:session-opened", () => {
  loadContacts();
});

// Cross-surface live updates: every contacts mutation (island API call,
// the Contacts tab forms, a vCard import — this browser or another one)
// publishes the payload-less "contacts" SSE nudge. The server's contacts
// panel re-fetches its own partial via hx-trigger="sse:contacts"; that
// subscription also fires a bubbling htmx:sseMessage whose detail is the
// raw MessageEvent (detail.type === "contacts"), which is what this
// listener keys on — the island's dropdown then re-fetches through its
// own session credentials, mirroring the voicemail nudge pattern.
document.addEventListener("htmx:sseMessage", (event) => {
  const detail = /** @type {CustomEvent} */ (event).detail;
  if (detail && detail.type === "contacts") {
    loadContacts();
  }
});

async function migrateLegacyContacts() {
  if (legacyContacts.length === 0) return;
  // Numbers already on the server keep their server-side names.
  const known = new Set(personalContacts.map((c) => c.number));
  const missing = legacyContacts.filter(
    (c) => c.number && !known.has(c.number),
  );
  try {
    for (const contact of missing) {
      const res = await authedFetch("/api/contacts", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          name: contact.name || contact.number,
          number: contact.number,
        }),
      });
      if (!res.ok) throw new Error(`HTTP ${res.status}`);
    }
    localStorage.removeItem(CONTACTS_KEY);
    legacyContacts = [];
    log(`imported ${missing.length} local contact(s) into the server store`);
    // The migration is otherwise invisible: the toast says it HAPPENED
    // (and how many rows moved) so the one-time import is not a mystery.
    announce(t("contactsMigrated")(missing.length), "ok");
    if (missing.length > 0) await loadContacts();
  } catch (err) {
    log(
      `local contact import failed (${errorMessage(err)}); retrying next login`,
      "error",
    );
  }
}

/** @param {string} number @param {string} name */
function saveLegacyContact(number, name) {
  legacyContacts = [
    { name: name || number, number },
    ...legacyContacts.filter((c) => c.number !== number),
  ].slice(0, CONTACTS_MAX);
  localStorage.setItem(CONTACTS_KEY, JSON.stringify(legacyContacts));
}

/** @param {string} number @param {string} name @returns {Promise<void>} */
export async function saveContact(number, name) {
  const clean = String(number).replace(/[^\d+*#a-zA-Z]/g, "");
  if (!clean) return;
  try {
    const res = await authedFetch("/api/contacts", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ name: name || clean, number: clean }),
    });
    if (!res.ok) throw new Error(`HTTP ${res.status}`);
    await loadContacts();
    announce(t("contactSaved")(clean), "ok");
  } catch (err) {
    // Server unreachable: keep the pre-migration behavior — the row
    // lands in the local list and migrates on a later login.
    log(`contact save failed (${errorMessage(err)}); kept locally`, "error");
    announce(t("contactSaveFailed")(errorMessage(err)), "error");
    saveLegacyContact(clean, name);
    renderContacts();
  }
}

// The shell's data-save-contact affordance (history/voicemail rows)
// bridges here: the island owns /api/contacts writes and their
// feedback, the shell only forwards the gesture.
document.addEventListener("wp:save-contact", (event) => {
  /** @type {{number?: string, name?: string}} */
  const detail = /** @type {CustomEvent} */ (event).detail || {};
  if (detail.number) saveContact(detail.number, detail.name || "");
});

/** @param {string} number */
function removeLegacyContact(number) {
  legacyContacts = legacyContacts.filter((c) => c.number !== number);
  localStorage.setItem(CONTACTS_KEY, JSON.stringify(legacyContacts));
  renderContacts();
}

/** @param {IslandContactRow} contact @returns {Promise<void>} */
async function removeServerContact(contact) {
  const { id } = contact;
  // Only a server row carries an id (the caller checks too); a legacy row
  // has nothing to delete server-side.
  if (!id) return;
  try {
    const res = await authedFetch(
      `/api/contacts?id=${encodeURIComponent(id)}`,
      {
        method: "DELETE",
      },
    );
    if (!res.ok) throw new Error(`HTTP ${res.status}`);
    personalContacts = personalContacts.filter((c) => c.id !== id);
    renderContacts();
  } catch (err) {
    log(`contact removal failed: ${errorMessage(err)}`, "error");
  }
}

export function renderContacts() {
  // Every contact render also refreshes the dial typeahead's source
  // (and re-attempts its boot — at first load it may have bailed with
  // no contacts at all). Legacy-only contacts are deliberately not
  // suggested: they migrate to the server on first login.
  setExtraContacts(personalContacts);
  initDialTypeahead();
  if (!phoneApiEnabled && sharedContacts.length === 0) return;
  els.contactsWrap.hidden = false;
  // Server rows first; legacy rows render only for numbers the server
  // does not know yet (pending or failed migration).
  const serverNumbers = new Set(personalContacts.map((c) => c.number));
  const personal = [
    ...personalContacts,
    ...legacyContacts.filter((c) => !serverNumbers.has(c.number)),
  ];
  /**
   * @param {IslandContactRow} contact
   * @param {boolean} shared
   * @returns {HTMLLIElement}
   */
  const mk = (contact, shared) => {
    const li = document.createElement("li");
    const name = document.createElement("span");
    name.className = "name";
    name.textContent = contact.name;
    const number = document.createElement("span");
    number.className = "number";
    number.textContent = contact.number;
    const origin = document.createElement("span");
    origin.className = "origin";
    origin.textContent = shared ? t("contactShared") : "";
    const call = document.createElement("button");
    call.className = "ghost small call";
    call.textContent = "☎";
    call.addEventListener("click", () => dialFromUi(contact.number));
    li.append(name, number, origin);
    if (shared) {
      li.append(call);
    } else {
      const remove = document.createElement("button");
      remove.className = "ghost small";
      remove.textContent = "✕";
      remove.title = t("contactRemove");
      remove.addEventListener("click", () => {
        if (contact.id) removeServerContact(contact);
        else removeLegacyContact(contact.number);
      });
      li.append(call, remove);
    }
    return li;
  };
  els.contactsList.replaceChildren(
    ...sharedContacts.map((c) => mk(c, true)),
    ...personal.map((c) => mk(c, false)),
  );
}

// --- voicemail (phone API) ---------------------------------------------------

/** @type {ReturnType<typeof setTimeout> | null} */
let voicemailPoll = null;

export async function refreshVoicemail() {
  if (!phoneApiEnabled || !authHeaderValue()) return;
  // phoneApiEnabled + a session check precede this (the credentials are
  // exactly what authHeaderValue() just minted from).
  const extension = /** @type {{extension: string}} */ (getCredentials())
    .extension;
  try {
    const headers = { Authorization: authHeaderValue() };
    const [summaryRes, listRes] = await Promise.all([
      fetch(`/phone-api/voicemail/${extension}/summary`, { headers }),
      fetch(`/phone-api/voicemail/${extension}/messages`, { headers }),
    ]);
    noteThrottled(summaryRes, "voicemail/summary");
    noteThrottled(listRes, "voicemail/messages");
    if (summaryRes.status === 401 || listRes.status === 401) {
      els.vmStatus.textContent = t("vmAuthFailed");
      return;
    }
    const summary = await summaryRes.json();
    const list = await listRes.json();
    const unread = Number.isFinite(summary.new) ? summary.new : 0;
    els.vmBadge.hidden = unread === 0;
    els.vmBadge.textContent = t("vmNewCount")(unread);
    els.vmWrap.hidden = false;
    els.vmStatus.textContent = "";
    const messages = Array.isArray(list.messages) ? list.messages : [];
    if (messages.length === 0) {
      els.vmList.replaceChildren(
        Object.assign(document.createElement("li"), {
          textContent: t("vmEmpty"),
        }),
      );
      return;
    }
    els.vmList.replaceChildren(
      .../** @type {IslandVoicemailMessage[]} */ (messages).map((msg) => {
        const li = document.createElement("li");
        if (!msg.read) li.className = "unread";
        const caller = document.createElement("span");
        caller.className = "caller";
        caller.textContent = msg.cid_number || msg.cid_name || "?";
        const len = document.createElement("span");
        len.className = "len";
        len.textContent = `${msg.seconds}s`;
        const when = document.createElement("span");
        when.className = "when";
        when.textContent = new Date(msg.created * 1000).toLocaleString();
        const play = document.createElement("button");
        play.className = "ghost small";
        play.innerHTML = playIcon();
        play.addEventListener("click", () => {
          new Audio(msg.audio_url).play().catch((err) => {
            log(`playback: ${errorMessage(err)}`, "warn");
            announce(t("vmPlayFailed")(errorMessage(err)), "warn");
          });
        });
        const del = document.createElement("button");
        del.className = "ghost small";
        del.innerHTML = trashIcon();
        del.title = t("vmDelete");
        del.addEventListener("click", async () => {
          try {
            const res = await authedFetch(
              `/phone-api/voicemail/${extension}/messages/${msg.uuid}`,
              { method: "DELETE" },
            );
            if (!res.ok) {
              log(`voicemail delete failed: HTTP ${res.status}`, "error");
              announce(t("vmDeleteFailed")(res.status), "error");
              return;
            }
            refreshVoicemail();
          } catch (err) {
            log(`voicemail delete failed: ${errorMessage(err)}`, "error");
            announce(t("vmDeleteFailed")(errorMessage(err)), "error");
          }
        });
        li.append(caller, len, when, play, del);
        if (asrEnabled) {
          const tr = document.createElement("button");
          tr.className = "ghost small";
          tr.textContent = t("transcribe");
          tr.title = t("transcribe");
          const transcript = document.createElement("div");
          transcript.className = "vm-transcript";
          tr.addEventListener("click", () =>
            transcribeNow(msg.audio_url, transcript, {
              filename: `${msg.uuid}.wav`,
            }),
          );
          li.append(tr, transcript);
        }
        return li;
      }),
    );
  } catch (err) {
    els.vmStatus.textContent = t("vmAuthFailed");
    log(`voicemail unavailable: ${errorMessage(err)}`, "error");
  }
}

export function scheduleVoicemailRefresh() {
  if (!phoneApiEnabled) return;
  if (voicemailPoll !== null) clearTimeout(voicemailPoll);
  voicemailPoll = setTimeout(refreshVoicemail, 1500);
}

export function cancelVoicemailRefresh() {
  if (voicemailPoll !== null) clearTimeout(voicemailPoll);
}
