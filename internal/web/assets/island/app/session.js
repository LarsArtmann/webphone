// Server session sync: after the island's SIP REGISTER succeeds, hand the
// same credentials to the webphone server so the tabs and the phone-api
// proxy can act for this extension. The session lives in an HttpOnly
// cookie server-side; the password is never persisted in the browser.
//
// Failures are deliberately non-fatal: the call UI does not depend on the
// server session — if it cannot be created, tabs stay gated and the island
// keeps calling.

export async function createSession(extension, password) {
  try {
    const res = await fetch("/api/session", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ extension, password }),
    });
    if (!res.ok) {
      console.warn("webphone: server session not created (HTTP " + res.status + ")");
    }
  } catch (err) {
    console.warn("webphone: server session not created (" + err.message + ")");
  }
}

export async function destroySession() {
  try {
    await fetch("/api/session", { method: "DELETE" });
  } catch {
    // signing out of the server session is best-effort; the cookie dies
    // with the tab session anyway
  }
}
