// The panels below the call area: call history (local + server CDR),
// contacts (shared config + personal contacts on the SERVER store, one
// home with the Contacts tab) and voicemail (phone API). All server
// access rides the extension's session via authedFetch — there is no
// second login.

import { phoneApiEnabled, crmEnabled, sharedContacts } from "./config.js";
import {
  authedFetch,
  authHeaderValue,
  getCredentials,
  noteThrottled,
} from "./auth.js";
import { t } from "./i18n.js";
import { announce, dialFromUi, els, log } from "./ui.js";

// --- call history ----------------------------------------------------------

const HISTORY_KEY = "pbx-history";
const HISTORY_MAX = 20;
const CONTACTS_KEY = "pbx-contacts";
const CONTACTS_MAX = 50;

// Server-backed rows (CDR), fetched when the phone API is enabled; kept
// separate from the local session history so reloads never lose either.
let serverHistory = [];

export function readHistory() {
  try {
    const raw = JSON.parse(localStorage.getItem(HISTORY_KEY) || "[]");
    return Array.isArray(raw) ? raw : [];
  } catch {
    return [];
  }
}

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
        outcome: established ? "answered" : "missed",
      }),
    });
    if (!res.ok) {
      const detail = `HTTP ${res.status}`;
      log(`crm call log failed: ${detail}`, "warn");
      announce(t("crmLogFailed")(detail), "warn");
    }
  } catch (err) {
    log(`crm call log failed: ${err.message}`, "warn");
    announce(t("crmLogFailed")(err.message), "warn");
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
    log(`server history unavailable: ${err.message}`, "error");
  }
}

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
  const serverRows = serverHistory.map((row) =>
    makeHistoryRow({
      dir: row.context === "public" ? "←" : "→",
      target:
        row.context === "public"
          ? row.caller_id_number
          : row.destination_number,
      whenText: row.start || "—",
      durText:
        row.billsec > 0
          ? `${Math.floor(row.billsec / 60)}:${String(row.billsec % 60).padStart(2, "0")}`
          : "—",
      number:
        row.context === "public"
          ? row.caller_id_number
          : row.destination_number,
    }),
  );
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

function readPersonalContacts() {
  try {
    const raw = JSON.parse(localStorage.getItem(CONTACTS_KEY) || "[]");
    return Array.isArray(raw) ? raw : [];
  } catch {
    return [];
  }
}

let personalContacts = [];
// Seeded once at load; emptied only by a confirmed import.
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
    log(`server contacts unavailable: ${err.message}`, "error");
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
  if (event.detail && event.detail.type === "contacts") {
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
      `local contact import failed (${err.message}); retrying next login`,
      "error",
    );
  }
}

function saveLegacyContact(number, name) {
  legacyContacts = [
    { name: name || number, number },
    ...legacyContacts.filter((c) => c.number !== number),
  ].slice(0, CONTACTS_MAX);
  localStorage.setItem(CONTACTS_KEY, JSON.stringify(legacyContacts));
}

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
    log(`contact save failed (${err.message}); kept locally`, "error");
    announce(t("contactSaveFailed")(err.message), "error");
    saveLegacyContact(clean, name);
    renderContacts();
  }
}

// The shell's data-save-contact affordance (history/voicemail rows)
// bridges here: the island owns /api/contacts writes and their
// feedback, the shell only forwards the gesture.
document.addEventListener("wp:save-contact", (event) => {
  const detail = event.detail || {};
  if (detail.number) saveContact(detail.number, detail.name || "");
});

function removeLegacyContact(number) {
  legacyContacts = legacyContacts.filter((c) => c.number !== number);
  localStorage.setItem(CONTACTS_KEY, JSON.stringify(legacyContacts));
  renderContacts();
}

async function removeServerContact(contact) {
  try {
    const res = await authedFetch(
      `/api/contacts?id=${encodeURIComponent(contact.id)}`,
      { method: "DELETE" },
    );
    if (!res.ok) throw new Error(`HTTP ${res.status}`);
    personalContacts = personalContacts.filter((c) => c.id !== contact.id);
    renderContacts();
  } catch (err) {
    log(`contact removal failed: ${err.message}`, "error");
  }
}

export function renderContacts() {
  if (!phoneApiEnabled && sharedContacts.length === 0) return;
  els.contactsWrap.hidden = false;
  // Server rows first; legacy rows render only for numbers the server
  // does not know yet (pending or failed migration).
  const serverNumbers = new Set(personalContacts.map((c) => c.number));
  const personal = [
    ...personalContacts,
    ...legacyContacts.filter((c) => !serverNumbers.has(c.number)),
  ];
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

let voicemailPoll = null;

export async function refreshVoicemail() {
  if (!phoneApiEnabled || !authHeaderValue()) return;
  const extension = getCredentials().extension;
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
      ...messages.map((msg) => {
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
        play.textContent = "▶";
        play.addEventListener("click", () => {
          new Audio(msg.audio_url).play().catch((err) => {
            log(`playback: ${err.message}`, "warn");
            announce(t("vmPlayFailed")(err.message), "warn");
          });
        });
        const del = document.createElement("button");
        del.className = "ghost small";
        del.textContent = "🗑";
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
            log(`voicemail delete failed: ${err.message}`, "error");
            announce(t("vmDeleteFailed")(err.message), "error");
          }
        });
        li.append(caller, len, when, play, del);
        return li;
      }),
    );
  } catch (err) {
    els.vmStatus.textContent = t("vmAuthFailed");
    log(`voicemail unavailable: ${err.message}`, "error");
  }
}

export function scheduleVoicemailRefresh() {
  if (!phoneApiEnabled) return;
  clearTimeout(voicemailPoll);
  voicemailPoll = setTimeout(refreshVoicemail, 1500);
}

export function cancelVoicemailRefresh() {
  clearTimeout(voicemailPoll);
}
