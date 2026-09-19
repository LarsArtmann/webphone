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
    // The login response invalidated the old CSRF token (fixation
    // defense); adopt the fresh one before anything else POSTs.
    try {
      await adoptFreshCsrfToken();
    } catch (err) {
      console.warn(
        "webphone: CSRF token rotation failed (" + err.message + "), reloading",
      );
      // The fresh server session survives a reload (cookie); the served
      // page then carries a matching token again. No call is lost: no
      // call can exist before the REGISTER that just succeeded.
      window.location.reload();
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

// --- SSE live indicator ----------------------------------------------------
// Live tab updates deserve their own visibility, separate from SIP
// registration: the feed can drop while calls keep working. htmx's SSE
// extension fires htmx:sseOpen/-Close/-Error on the sse-connect element
// and they bubble, so one document-level listener covers both the
// server-rendered feed and the post-login dynamic attach. The pill is
// created here (JS-only) so the served markup — and with it the DOM
// contract and the upstream E2E — is untouched.
export function initSseLiveIndicator() {
  if (document.getElementById("wp-sse-live")) return;
  const pill = document.createElement("div");
  pill.id = "wp-sse-live";
  pill.title = "live tab updates";
  pill.setAttribute("aria-hidden", "true");
  document.addEventListener("htmx:sseOpen", () => {
    pill.dataset.live = "1";
  });
  document.addEventListener("htmx:sseClose", () => {
    delete pill.dataset.live;
  });
  document.addEventListener("htmx:sseError", () => {
    delete pill.dataset.live;
  });
  document.body.append(pill);
}

// The server renders the CSRF token into <meta name="csrf-token">; the
// nosurf double-submit cookie pairs with it.
function csrfToken() {
  const meta = document.querySelector('meta[name="csrf-token"]');
  return meta ? meta.getAttribute("content") : "";
}

// Login rotates the CSRF token (fixation defense): the login response
// deletes the cookie, so this fetches the fresh masked token from the
// CSRF middleware (the GET regenerates the deleted cookie) and updates
// every token consumer. They all read live: the meta tag here and in
// auth.js, and htmx re-reads the body's hx-headers per request, so
// updating those two spots re-arms every later POST.
async function adoptFreshCsrfToken() {
  const res = await fetch("/api/csrf");
  if (!res.ok) throw new Error("HTTP " + res.status);
  const data = await res.json();
  if (!data.token) throw new Error("empty token");
  const meta = document.querySelector('meta[name="csrf-token"]');
  if (meta) meta.setAttribute("content", data.token);
  if (document.body.hasAttribute("hx-headers")) {
    document.body.setAttribute(
      "hx-headers",
      JSON.stringify({ "X-CSRF-Token": data.token }),
    );
  }
}
