// Keyboard shortcuts and media keys. Deliberately inert while typing in
// any field, and never with modifier keys (those belong to the browser).

import {
  answerIncoming,
  cycleFocus,
  hangupFocused,
  rejectIncoming,
  toggleHoldFocused,
  toggleMuteFocused,
} from "./calls.js";
import { state } from "./state.js";
import { log } from "./ui.js";

/**
 * @param {KeyboardEvent} event
 * @returns {boolean}
 */
function typingTarget(event) {
  const target = event.target;
  if (!(target instanceof Element)) return true;
  return Boolean(
    target.closest("input, textarea, select, [contenteditable='true']"),
  );
}

export function initShortcuts() {
  document.addEventListener("keydown", (event) => {
    if (event.ctrlKey || event.metaKey || event.altKey) return;
    if (typingTarget(event)) return;

    switch (event.key) {
      case "a":
      case "A":
        answerIncoming();
        break;
      case "h":
      case "H":
        hangupFocused();
        break;
      case "m":
      case "M":
        toggleMuteFocused();
        break;
      case "p":
      case "P":
        toggleHoldFocused();
        break;
      case "[":
        cycleFocus(-1);
        break;
      case "]":
        cycleFocus(1);
        break;
      case "Escape":
        // Esc clears the NEWEST ring first (the ringing queue's tail);
        // only with an empty queue does it hang up the focused call —
        // a second caller ringing must never kill the live conversation.
        if (state.incomingSessions.length) rejectIncoming();
        else hangupFocused();
        break;
      case "MediaPlayPause":
        if (!answerIncoming()) toggleMuteFocused();
        break;
      case "MediaStop":
        hangupFocused();
        break;
      default:
        return;
    }
    event.preventDefault();
  });
  log(
    "keyboard shortcuts: A answer oldest ring · H hangup · M mute · P hold · [ ] switch calls · Esc reject newest ring/hangup",
  );
}
