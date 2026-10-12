// The voicemail panel's ASR-off arm (#37): with the seam disabled, rows
// render WITHOUT the transcribe button, the transcript node, or any
// /api/transcribe traffic. Lives in its own file because config.js reads
// window.PBX_CONFIG at import time — the gate is only honestly off in a
// fresh process (the on-arm lives in voicemail-transcribe.test.mjs).
import test from "node:test";
import assert from "node:assert/strict";

import { installBrowserGlobals } from "./helpers.mjs";

const doc = installBrowserGlobals();
globalThis.window.PBX_CONFIG = { phoneApi: true };

let fetchCalls = [];
globalThis.fetch = async (url) => {
  fetchCalls.push(String(url));
  if (String(url).includes("/summary")) {
    return {
      ok: true,
      status: 200,
      headers: new Headers(),
      json: async () => ({ new: 1 }),
    };
  }
  return {
    ok: true,
    status: 200,
    headers: new Headers(),
    json: async () => ({
      messages: [
        {
          uuid: "u-1",
          cid_number: "1002",
          seconds: 12,
          created: Date.now() / 1000,
          read: true,
          audio_url: "/phone-api/voicemail/1000/audio/u-1",
        },
      ],
    }),
  };
};

const { setCredentials } = await import("../island/app/auth.js");
setCredentials("1000", "pw");

const panels = await import("../island/app/panels.js?test=vm-asr-off");

test("asr off: voicemail rows carry no transcribe affordance", async () => {
  await panels.refreshVoicemail();

  const row = doc.getElementById("vm-list").children[0];
  assert.ok(row, "voicemail row still renders with the seam off");
  assert.equal(
    row.children.find((c) => c.className === "vm-transcript"),
    undefined,
    "no transcript node without the seam",
  );
  assert.equal(
    row.children.filter((c) => c.textContent && c.className.includes("ghost")).length,
    0,
    "no labeled ghost button without the seam",
  );
  assert.equal(
    fetchCalls.filter((u) => u.includes("/api/transcribe")).length,
    0,
    "no transcription traffic without the seam",
  );
});
