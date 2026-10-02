// Pre-call device check (M25 A10): one button in the advanced
// diagnostics area proves the audio path BEFORE the first call — the
// mic permission, a live stream, and the speaker count. The stream is
// released immediately: the check BORROWS the devices, it never holds
// them (the mic indicator must go dark again). #log copy is English —
// operator diagnostics, the same policy as every #log line.
import { $, log } from "./ui.js";

export async function runDeviceCheck() {
  log("device check: starting");
  const media = navigator.mediaDevices;
  if (!media || typeof media.getUserMedia !== "function") {
    log("device check: media devices unavailable in this browser", "error");
    return;
  }
  try {
    const stream = await media.getUserMedia({ audio: true, video: false });
    const tracks = stream.getTracks();
    const label = tracks[0] && tracks[0].label ? ` (${tracks[0].label})` : "";
    for (const track of tracks) track.stop();
    log(`device check: microphone OK${label}`);
  } catch (err) {
    const reason = err && err.name ? err.name : "error";
    log(`device check: microphone FAILED — ${reason}`, "error");
    return;
  }
  if (typeof media.enumerateDevices !== "function") return;
  try {
    const devices = await media.enumerateDevices();
    const outputs = devices.filter((device) => device.kind === "audiooutput");
    log(`device check: ${outputs.length} audio output(s) visible`);
  } catch {
    log("device check: could not enumerate outputs", "warn");
  }
}

// initDeviceCheck wires the button: the check runs disabled-while-
// pending so a double click cannot stack two acquisitions.
export function initDeviceCheck() {
  const button = $("wp-devtest-btn");
  if (!button) return;
  button.addEventListener("click", () => {
    button.disabled = true;
    runDeviceCheck().finally(() => {
      button.disabled = false;
    });
  });
}
