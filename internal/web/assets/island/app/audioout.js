// Audio output picker: route the call audio to a chosen speaker via
// setSinkId. Feature-detected end to end — a browser without setSinkId
// (Safari), without mediaDevices enumeration, or with fewer than two
// DISTINCT physical outputs never sees the picker.
//
// The choice carries BOTH halves of "call audio": the remote media
// element AND every locally generated tone (ringback, the incoming-call
// ring) through AudioContext.setSinkId. Owner decision 2026-11-06 (issue
// #10) replaced the earlier "rings stay room alarms" ruling: on a
// multi-output desk the device the operator wears is the phone, and a
// ring that lands on a different speaker is a missed call. Where the
// WebAudio half is unsupported the ring still sounds on the default
// output — the picker never claims a routing it could not perform.
// All failures log and keep the picker honest; a dead picker must never
// break a call.

import { setToneSink } from "./audio.js";
import { t } from "./i18n.js";
import { $, els, errorMessage, log } from "./ui.js";

const SINK_KEY = "wp-sink";
// Chrome aliases ("default", "communications") point at the system
// default sink, not separate hardware — they don't count as a choice.
const SINK_ALIASES = ["default", "communications"];

let initialized = false;

// physicalOutputs filters the enumeration down to real, addressable
// output devices (aliases excluded so one-speaker machines stay picker-free).
/**
 * @param {MediaDeviceInfo[]} devices
 * @returns {MediaDeviceInfo[]}
 */
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
    // Private-mode storage: degrade to the browser's default output.
    return "";
  }
}

/** @param {string} deviceId */
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
  const wrap = $("audio-out-wrap");
  const select = $("audio-output");
  if (!audio || !wrap || !select) return;
  if (typeof audio.setSinkId !== "function") return;
  if (
    !navigator.mediaDevices ||
    typeof navigator.mediaDevices.enumerateDevices !== "function"
  ) {
    return;
  }

  /** @param {string} deviceId */
  const apply = (deviceId) => {
    // Both halves follow the same choice, and "" (the browser default) is
    // a REAL pick: choosing Default after a device must move the audio
    // back, not silently stay on the last speaker (the pre-2026-11-06
    // picker skipped this and reset nothing).
    setToneSink(deviceId);
    const verdict = audio.setSinkId(deviceId);
    if (verdict && typeof verdict.then === "function") {
      verdict.catch((err) =>
        log(`audio output switch failed: ${errorMessage(err)}`, "error"),
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
        apply(select.value);
      })
      .catch(() => {
        // Device enumeration failure: hide the picker; audio still plays
        // on the browser default.
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
