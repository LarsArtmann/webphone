// Incoming-call attention: system notifications, tab-title flash. The
// audible half (ring tone) lives in audio.js.

import { sipDomain } from "./config.js";
import { errorMessage, log } from "./ui.js";

// Ask once per device, from the login click (a user gesture — browsers
// refuse permission prompts without one).
export function requestNotifications() {
  if (!("Notification" in window)) return;
  if (Notification.permission === "default") {
    Notification.requestPermission()
      .then((state) => {
        log(`notifications ${state}`);
      })
      .catch((err) => {
        log(`notification permission failed: ${errorMessage(err)}`, "warn");
      });
  } else {
    log(`notifications ${Notification.permission}`);
  }
}

// One notification PER CALLER: the tag carries the caller so two
// simultaneous rings are two shade entries (the stock ntfy/Chrome
// behavior replaces same-tag notifications only).
/** @param {string} from */
export function notifyIncoming(from) {
  if (!("Notification" in window) || Notification.permission !== "granted")
    return;
  try {
    const n = new Notification(`☎ ${from}`, {
      body: `Incoming call on ${sipDomain}`,
      tag: `pbx-incoming-${from}`,
    });
    n.addEventListener("click", () => window.focus());
  } catch (err) {
    log(`notification failed: ${errorMessage(err)}`);
  }
}

const originalTitle = document.title;
/** @type {ReturnType<typeof setInterval> | null} */
let titleFlashTimer = null;

export function titleFlashStart() {
  if (titleFlashTimer) return;
  let on = false;
  titleFlashTimer = setInterval(() => {
    document.title = (on = !on) ? "☎ ☎ ☎" : originalTitle;
  }, 900);
}

export function titleFlashStop() {
  if (!titleFlashTimer) return;
  clearInterval(titleFlashTimer);
  titleFlashTimer = null;
  document.title = originalTitle;
}
