// Audio output picker: route the REMOTE call audio to a chosen speaker
// via setSinkId. Feature-detected end to end — a browser without
// setSinkId (Safari), without mediaDevices enumeration, or with fewer
// than two DISTINCT physical outputs never sees the picker. Ring tones
// stay room-alarms by decision: only the remote audio element is routed.
// All failures log and keep the picker hidden; a dead picker must never
// break a call.

import { t } from "./i18n.js";
import { els, log } from "./ui.js";

const SINK_KEY = "wp-sink";
// Chrome aliases ("default", "communications") point at the system
// default sink, not separate hardware — they don't count as a choice.
const SINK_ALIASES = ["default", "communications"];

let initialized = false;

// physicalOutputs filters the enumeration down to real, addressable
// output devices (aliases excluded so one-speaker machines stay picker-free).
export function physicalOutputs(devices) {
  return devices.filter(
    (device) =>
      device.kind === "audiooutput" &&
      device.deviceId &&
      !SINK_ALIASES.includes(device.deviceId),
  );
}

function storedSink() {
  try {
    return localStorage.getItem(SINK_KEY) || "";
  } catch {
    return "";
  }
}

function persistSink(deviceId) {
  try {
    localStorage.setItem(SINK_KEY, deviceId);
  } catch {
    // Private-mode storage failures degrade to an unpersisted pick.
  }
}

export function initAudioOutput() {
  if (initialized) return;
  initialized = true;

  const audio = els.remoteAudio;
  const wrap = document.getElementById("audio-out-wrap");
  const select = document.getElementById("audio-output");
  if (!audio || !wrap || !select) return;
  if (typeof audio.setSinkId !== "function") return;
  if (
    !navigator.mediaDevices ||
    typeof navigator.mediaDevices.enumerateDevices !== "function"
  ) {
    return;
  }

  const apply = (deviceId) => {
    if (!deviceId) return;
    const verdict = audio.setSinkId(deviceId);
    if (verdict && typeof verdict.then === "function") {
      verdict.catch((err) =>
        log(`audio output switch failed: ${err.message}`, "error"),
      );
    }
  };

  const refresh = () => {
    navigator.mediaDevices
      .enumerateDevices()
      .then((devices) => {
        const outputs = physicalOutputs(devices);
        if (outputs.length < 2) {
          wrap.hidden = true;
          return;
        }
        const previous = select.value;
        select.replaceChildren();
        const fallback = document.createElement("option");
        fallback.value = "";
        fallback.textContent = t("audioDefault");
        select.append(fallback);
        for (const device of outputs) {
          const option = document.createElement("option");
          option.value = device.deviceId;
          option.textContent = device.label || device.deviceId;
          select.append(option);
        }
        // Keep a still-present selection across refreshes (device
        // changes, language switches); otherwise fall back to default.
        const chosen = outputs.some((device) => device.deviceId === previous)
          ? previous
          : storedSink();
        select.value = outputs.some((device) => device.deviceId === chosen)
          ? chosen
          : "";
        wrap.hidden = false;
        if (select.value) apply(select.value);
      })
      .catch(() => {
        wrap.hidden = true;
      });
  };

  select.addEventListener("change", () => {
    persistSink(select.value);
    apply(select.value);
  });
  if (typeof navigator.mediaDevices.addEventListener === "function") {
    navigator.mediaDevices.addEventListener("devicechange", refresh);
  }
  document.addEventListener("wp:lang-changed", refresh);
  refresh();
}
