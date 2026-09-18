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
      headers: {
        "Content-Type": "application/json",
        "X-CSRF-Token": csrfToken(),
      },
      body: JSON.stringify({ extension, password }),
    });
    if (!res.ok) {
      console.warn(
        "webphone: server session not created (HTTP " + res.status + ")",
      );
      return;
    }
    connectLiveUpdates();
  } catch (err) {
    console.warn("webphone: server session not created (" + err.message + ")");
  }
}

export async function destroySession() {
  try {
    await fetch("/api/session", {
      method: "DELETE",
      headers: { "X-CSRF-Token": csrfToken() },
      keepalive: true,
    });
  } catch {
    // signing out of the server session is best-effort; the cookie dies
    // with the tab session anyway
  }
  // Reload: the page was rendered signed-in (SSE feed connected, tabs
  // unlocked); after logout the server must render the login-card shell
  // again. Safe by construction — logout ends every call first, so no
  // live call state is lost.
  window.location.reload();
}

// The server only renders sse-connect for already-signed-in pages; this
// session was created client-side after the island's REGISTER succeeded.
// Attach the SSE feed dynamically so live tab updates work from this
// login on — without a reload, which would drop the in-memory password
// and any call in progress.
function connectLiveUpdates() {
  const root = document.querySelector(".wp-root");
  if (!root || root.hasAttribute("sse-connect")) return;
  root.setAttribute("hx-ext", "sse");
  root.setAttribute("sse-connect", "/events");
  if (window.htmx) window.htmx.process(root);
}

// The server renders the CSRF token into <meta name="csrf-token">; the
// nosurf double-submit cookie pairs with it.
function csrfToken() {
  const meta = document.querySelector('meta[name="csrf-token"]');
  return meta ? meta.getAttribute("content") : "";
}
