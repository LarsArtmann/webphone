// Dial typeahead: ranked suggestions from the shared contacts the
// server already ships in window.PBX_CONFIG — zero round-trips, fully
// keyboard-driven, and invisible whenever anything is missing (no
// contacts, no dial form, no JS). Pure client-side over config data.

import { sharedContacts } from "./config.js";
import {
  cooldownFor,
  cooldownRemainder,
  refreshDialSlots,
} from "./dialgate.js";
import { t } from "./i18n.js";
import { els } from "./ui.js";

const MAX_SUGGESTIONS = 6;

/** @type {HTMLUListElement | null} */
let list = null;
/** @type {PBXSharedContact[]} */
let entries = [];
let activeIndex = -1;
let initialized = false;

// Per-extension personal contacts (the server store, fed by panels.js
// once the session exists). A number present in both worlds yields ONE
// suggestion — the personal name wins, it is the one the user edits.
/** @type {PBXSharedContact[]} */
let personalContacts = [];

/** @param {unknown} contacts */
export function setExtraContacts(contacts) {
  personalContacts = Array.isArray(contacts)
    ? /** @type {PBXSharedContact[]} */ (
        contacts.filter((contact) => contact && contact.number)
      )
    : [];
}

/** @returns {PBXSharedContact[]} */
function suggestionSource() {
  const personalNumbers = new Set(personalContacts.map((c) => c.number));
  return [
    ...personalContacts,
    ...sharedContacts.filter((c) => !personalNumbers.has(c.number)),
  ];
}

// rankContacts scores a contact against the query: a name that starts
// with it beats a name that merely contains it, which beats a number
// that contains it. Ties break by name, then number — stable enough
// that the list never jumps around while typing.
/**
 * @param {PBXSharedContact[]} contacts
 * @param {unknown} query
 * @returns {PBXSharedContact[]}
 */
export function rankContacts(contacts, query) {
  const q = String(query || "")
    .trim()
    .toLowerCase();
  if (!q) return [];
  /** @type {Array<{contact: PBXSharedContact, score: number, name: string, number: string}>} */
  const scored = [];
  for (const contact of contacts) {
    const name = String(contact.name || "");
    const number = String(contact.number || "");
    if (!name && !number) continue;
    const lowerName = name.toLowerCase();
    const lowerNumber = number.toLowerCase();
    let score = -1;
    if (lowerName.startsWith(q)) score = 0;
    else if (lowerName.includes(q)) score = 1;
    else if (lowerNumber.includes(q)) score = 2;
    if (score < 0) continue;
    scored.push({ contact, score, name: lowerName, number: lowerNumber });
  }
  scored.sort(
    (a, b) =>
      a.score - b.score ||
      a.name.localeCompare(b.name) ||
      a.number.localeCompare(b.number),
  );
  return scored.slice(0, MAX_SUGGESTIONS).map((entry) => entry.contact);
}

function hide() {
  entries = [];
  activeIndex = -1;
  if (list) {
    list.replaceChildren();
    list.hidden = true;
  }
}

/** @param {number} next */
function setActive(next) {
  if (!list || entries.length === 0) return;
  activeIndex = (next + entries.length) % entries.length;
  [...list.children].forEach((option, i) => {
    option.classList.toggle("active", i === activeIndex);
    option.setAttribute("aria-selected", i === activeIndex ? "true" : "false");
  });
}

/** @param {PBXSharedContact} contact */
function choose(contact) {
  els.dest.value = contact.number;
  hide();
  els.dest.focus();
}

/** @param {PBXSharedContact[]} matches */
function render(matches) {
  if (!list) return;
  // Bound once: a module-level `let` does not stay narrowed inside the
  // forEach callback below.
  const box = list;
  box.replaceChildren();
  entries = matches;
  activeIndex = matches.length > 0 ? 0 : -1;
  matches.forEach((contact, i) => {
    const option = document.createElement("li");
    option.setAttribute("role", "option");
    option.dataset.index = String(i);
    option.setAttribute("aria-selected", i === 0 ? "true" : "false");
    option.classList.toggle("active", i === 0);
    const name = document.createElement("span");
    name.className = "typeahead-name";
    name.textContent = contact.name || contact.number;
    option.append(name);
    if (contact.name && contact.number) {
      const number = document.createElement("span");
      number.className = "typeahead-number";
      number.textContent = contact.number;
      option.append(number);
    }
    // Cooldown state lives IN the row (issue #7): a contact inside its
    // window says so where the number is picked, not only after the
    // gate denies the dial. The cache is server truth, refreshed
    // debounced; a failed/absent fetch decorates nothing.
    const slot = cooldownFor(contact.number);
    if (slot) {
      const cooldown = document.createElement("span");
      cooldown.className = "typeahead-cooldown";
      cooldown.textContent = t("cooldownRow")(cooldownRemainder(slot));
      option.append(cooldown);
    }
    box.append(option);
  });
  box.hidden = matches.length === 0;
  void refreshDialSlots();
}

export function initDialTypeahead() {
  if (initialized) return;
  if (!els.dest || !els.dialForm) return;
  if (sharedContacts.length === 0 && personalContacts.length === 0) return;
  initialized = true;

  list = document.createElement("ul");
  list.id = "dial-suggest";
  list.className = "wp-typeahead";
  list.setAttribute("role", "listbox");
  list.setAttribute("aria-label", "Contact suggestions");
  list.hidden = true;
  els.dialForm.append(list);

  els.dest.addEventListener("input", () => {
    render(rankContacts(suggestionSource(), els.dest.value));
  });

  els.dest.addEventListener("keydown", (event) => {
    if (!list || list.hidden) return;
    if (event.key === "ArrowDown") {
      event.preventDefault();
      setActive(activeIndex + 1);
    } else if (event.key === "ArrowUp") {
      event.preventDefault();
      setActive(activeIndex - 1);
    } else if (event.key === "Enter") {
      if (activeIndex >= 0 && entries[activeIndex]) {
        // Enter picks the active suggestion; the NEXT Enter dials.
        event.preventDefault();
        choose(entries[activeIndex]);
      }
    } else if (event.key === "Escape") {
      event.preventDefault();
      hide();
    }
  });

  // mousedown, not click: it wins the race against the input's blur and
  // keep the focus in the field (choose() refocuses anyway).
  list.addEventListener("mousedown", (event) => {
    const target = /** @type {Element | null} */ (event.target);
    const option =
      target && target.closest ? target.closest("li[data-index]") : null;
    if (!option) return;
    event.preventDefault();
    const row = /** @type {HTMLElement} */ (option);
    const contact = entries[Number(row.dataset.index)];
    if (contact) choose(contact);
  });

  document.addEventListener("click", (event) => {
    if (!list || list.hidden) return;
    // A click on (or inside) the dial form keeps the list open; anything
    // else is "clicked away". event.target may be a text node, hence the
    // closest guard rather than a bare call.
    const target = /** @type {Element | null} */ (event.target);
    if (target && target.closest && target.closest("#dial-form")) return;
    hide();
  });
}

// Language switches re-label the island in place; the list's aria-label
// is the only localized string this module owns.
/** @param {string} label */
export function relabelTypeahead(label) {
  if (list) list.setAttribute("aria-label", label);
}
