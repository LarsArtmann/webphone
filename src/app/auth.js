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
  return fetch(path, { ...options, headers });
}
