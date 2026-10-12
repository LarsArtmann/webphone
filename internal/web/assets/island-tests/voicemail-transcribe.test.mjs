// The voicemail panel's transcribe affordance (#37, ASR-on arm): every
// voicemail row grows a transcribe button whose click runs the shared
// transcribeNow flow — audio URL fetch → blob POST → trimmed text into
// the row's .vm-transcript node. Panels render through the same els stub
// registry as the other island suites. The ASR-off arm lives in
// voicemail-noasr.test.mjs: config.js reads PBX_CONFIG at import time,
// so flipping the gate honestly needs a fresh process, not a re-import.
import test from "node:test";
import assert from "node:assert/strict";

import { installBrowserGlobals } from "./helpers.mjs";

const doc = installBrowserGlobals();
globalThis.window.PBX_CONFIG = { phoneApi: true, asr: true };

let fetchCalls = [];
let transcribeText = " hello from the stub ";
const respond = (url) => {
  fetchCalls.push(String(url));
  if (String(url).includes("/api/transcribe")) {
    return {
      ok: true,
      status: 200,
      headers: new Headers(),
      json: async () => ({ text: transcribeText }),
    };
  }
  if (String(url).includes("/audio/")) {
    return {
      ok: true,
      status: 200,
      headers: new Headers(),
      blob: async () => ({ type: "audio/wav" }),
    };
  }
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

// Credentials ride the SHARED no-query auth.js instance — panels.js
// resolves "./auth.js" unqualified, and a ?query import would mint a
// second instance whose credentials panels never sees.
const { setCredentials } = await import("../island/app/auth.js");
setCredentials("1000", "pw");

const panels = await import("../island/app/panels.js?test=vm-asr-on");
const flush = () => new Promise((resolve) => setTimeout(resolve, 20));

test("asr on: voicemail rows carry a transcribe button wired to the seam", async () => {
  globalThis.fetch = respond;
  await panels.refreshVoicemail();

  const row = doc.getElementById("vm-list").children[0];
  assert.ok(row, "voicemail row rendered");
  const transcript = row.children.find((c) => c.className === "vm-transcript");
  assert.ok(transcript, "row carries the transcript target node");
  const button = row.children.find(
    (c) => c.className.includes("ghost") && typeof c.listeners.click === "object" && c.textContent,
  );
  assert.ok(button, "row carries the transcribe button (labeled, click-wired)");

  fetchCalls = [];
  button.listeners.click[0]();
  await flush();

  assert.equal(
    transcript.textContent,
    "hello from the stub",
    "transcribeNow rendered the trimmed provider text into the row",
  );
  assert.ok(
    fetchCalls.some((u) => u.includes("/phone-api/voicemail/1000/audio/u-1")),
    "the row's audio URL was fetched",
  );
  const post = fetchCalls.find((u) => u.includes("/api/transcribe"));
  assert.ok(post, "the transcription POST rode /api/transcribe");
  assert.ok(post.includes("filename=u-1.wav"), `expected the voicemail filename hint, got ${post}`);
});

test("asr on: a second render rebuilds the affordance (nudge-safe)", async () => {
  globalThis.fetch = respond;
  await panels.refreshVoicemail();
  const row = doc.getElementById("vm-list").children[0];
  const button = row.children.find(
    (c) => c.className.includes("ghost") && typeof c.listeners.click === "object" && c.textContent,
  );
  assert.ok(button, "re-render keeps the transcribe button");
});
