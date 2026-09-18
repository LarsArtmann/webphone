// Incoming-call attention: system notifications, tab-title flash. The
// audible half (ring tone) lives in audio.js.

import { sipDomain } from "./config.js";
import { log } from "./ui.js";

// Ask once per device, from the login click (a user gesture — browsers
// refuse permission prompts without one).
export function requestNotifications() {
  if (!("Notification" in window)) return;
  if (Notification.permission === "default") {
    Notification.requestPermission().then((state) => {
      log(`notifications ${state}`);
    });
  } else {
    log(`notifications ${Notification.permission}`);
  }
}

export function notifyIncoming(from) {
  if (!("Notification" in window) || Notification.permission !== "granted")
    return;
  try {
    const n = new Notification(`☎ ${from}`, {
      body: `Incoming call on ${sipDomain}`,
      tag: "pbx-incoming",
    });
    n.addEventListener("click", () => window.focus());
  } catch (err) {
    log(`notification failed: ${err.message}`);
  }
}

const originalTitle = document.title;
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
