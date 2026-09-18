// The panels below the call area: call history (local + server CDR),
// contacts (shared config + personal localStorage) and voicemail (phone
// API). All server access rides the extension's own SIP credentials via
// authedFetch — there is no second login.

import { phoneApiEnabled, sharedContacts } from "./config.js";
import { authedFetch, authHeaderValue, getCredentials } from "./auth.js";
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
    saveContact(number, number);
    renderContacts();
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

// --- contacts ---------------------------------------------------------------

function readPersonalContacts() {
  try {
    const raw = JSON.parse(localStorage.getItem(CONTACTS_KEY) || "[]");
    return Array.isArray(raw) ? raw : [];
  } catch {
    return [];
  }
}

export function saveContact(number, name) {
  const clean = String(number).replace(/[^\d+*#]/g, "");
  if (!clean) return;
  const list = [
    { name, number: clean },
    ...readPersonalContacts().filter((c) => c.number !== clean),
  ].slice(0, CONTACTS_MAX);
  localStorage.setItem(CONTACTS_KEY, JSON.stringify(list));
}

function removeContact(number) {
  const rest = readPersonalContacts().filter((c) => c.number !== number);
  localStorage.setItem(CONTACTS_KEY, JSON.stringify(rest));
}

export function renderContacts() {
  if (!phoneApiEnabled && sharedContacts.length === 0) return;
  els.contactsWrap.hidden = false;
  const personal = readPersonalContacts();
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
        removeContact(contact.number);
        renderContacts();
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
