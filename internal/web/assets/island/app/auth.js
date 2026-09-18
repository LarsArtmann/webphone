// Credentials for the per-extension phone API (voicemail, CDR history).
// Kept in memory only; never persisted.

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
  const meta = document.querySelector('meta[name="csrf-token"]');
  if (meta) headers.set("X-CSRF-Token", meta.getAttribute("content") || "");
  return fetch(path, { ...options, headers });
}
