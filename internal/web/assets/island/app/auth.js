// Credentials for the per-extension phone API (voicemail, CDR history).
// Kept in memory only; never persisted.

import { csrfToken } from "./csrf.js";
import { announce } from "./ui.js";

let credentials = null;

export function setCredentials(extension, password) {
  credentials = { extension, password };
}

export function clearCredentials() {
  credentials = null;
}

export function getCredentials() {
  return credentials;
}

export function authHeaderValue() {
  if (!credentials) return "";
  return `Basic ${btoa(`${credentials.extension}:${credentials.password}`)}`;
}

export async function authedFetch(path, options = {}) {
  const headers = new Headers(options.headers || {});
  if (credentials) headers.set("Authorization", authHeaderValue());
  // The server's CSRF middleware protects state-changing phone-api calls
  // (e.g. voicemail delete); the token pairs with the nosurf cookie.
  // csrf.js owns the meta reader — this is the island's ONE home rule,
  // same as session/passkey/enroll.
  const token = csrfToken();
  if (token) headers.set("X-CSRF-Token", token);
  const res = await fetch(path, { ...options, headers });
  noteThrottled(res, path);
  return res;
}

// Rate-limit visibility: the server answers 429 with a computed
// Retry-After on flood-sensitive surfaces. Silent !res.ok handling used
// to hide this from the user; surface it once per response instead.
// No automatic retry — callers keep their own semantics; the user sees
// why the panel did not refresh and how long the backoff lasts.
export function noteThrottled(res, path) {
  if (res.status !== 429) return;
  const waitSeconds = Number.parseInt(res.headers.get("Retry-After") || "", 10);
  const waitText = Number.isFinite(waitSeconds)
    ? ` — retry in ${waitSeconds}s`
    : "";
  console.warn(`webphone: rate limited on ${path}${waitText}`);
  announce(`Too many requests${waitText}`, "warn");
}
