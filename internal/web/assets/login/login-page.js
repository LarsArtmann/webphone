// The /login page's extension + password submit (served at
// /assets/login-page.js; the page is the cqrs-htmx loginpage library page
// hosting views.LoginPageExtra). Flow: POST /api/session mints the server
// session (the handler directory-verifies and answers 201), then the
// redirect home hands the browser to the island's boot probe, which signs
// in and REGISTERs — the exact resume path the passkey and SSO lanes
// take. English-only copy: the login page is a server-rendered surface
// with no island runtime (D3 keeps shell copy English).
//
// The CSRF token pairs with the nosurf cookie the page carries; the
// library renders <meta name="csrf-token"> for exactly this reader.

/** The remembered-extension key, shared with the old island form's storage. */
const REMEMBER_KEY = "pbx-extension";

/**
 * The login page's DOM handle lookup: every id is served markup
 * (views/login_page.templ), so the result is non-null by contract and a
 * typo'd id is a compile error instead of a silent `null`.
 * @template {keyof LoginPageDomIds} K
 * @param {K} id
 * @returns {LoginPageDomIds[K]}
 */
const $ = (id) => /** @type {LoginPageDomIds[K]} */ (document.getElementById(id));

const form = $("login-form");
const extInput = $("ext");
const passInput = $("pass");
const rememberInput = $("remember");
const errorEl = $("login-error");

// The remembered extension prefills the form (never the password).
const remembered = localStorage.getItem(REMEMBER_KEY);
if (remembered) {
  extInput.value = remembered;
  rememberInput.checked = true;
}

form.addEventListener("submit", (/** @type {SubmitEvent} */ event) => {
  event.preventDefault();
  signIn().catch(() => {
    showError("Could not reach the sign-in server; check your connection and try again.");
  });
});

async function signIn() {
  hideError();
  const submit = /** @type {HTMLButtonElement | null} */ (
    form.querySelector('button[type="submit"]')
  );
  if (submit) {
    submit.disabled = true;
  }
  const extension = extInput.value.trim();
  const password = passInput.value;
  try {
    if (!extension || !password) {
      showError("Enter your extension and password.");
      return;
    }
    // The user gesture IS this click: ask for notification permission
    // now, the way the old island login did — the island's boot logs
    // the resulting state after the redirect.
    try {
      if (typeof Notification !== "undefined") {
        void Notification.requestPermission();
      }
    } catch {
      // A browser without the API must not block the sign-in.
    }
    const res = await fetch("/api/session", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "X-CSRF-Token": csrfTokenFromMeta(),
      },
      body: JSON.stringify({ extension, password }),
    });
    if (res.status === 201) {
      if (rememberInput.checked) {
        localStorage.setItem(REMEMBER_KEY, extension);
      } else {
        localStorage.removeItem(REMEMBER_KEY);
      }
      window.location.replace("/");
      return;
    }
    if (res.status === 401) {
      showError("Wrong extension or password.");
      return;
    }
    if (res.status === 429) {
      const wait = Number.parseInt(res.headers.get("Retry-After") || "", 10);
      showError(
        Number.isFinite(wait)
          ? `Too many sign-in attempts — try again in ${wait}s.`
          : "Too many sign-in attempts — try again shortly.",
      );
      return;
    }
    showError(`Sign-in failed (HTTP ${res.status}); try again.`);
  } finally {
    if (submit) {
      submit.disabled = false;
    }
  }
}

/** @returns {string} the CSRF token the page meta carries ("" when absent). */
function csrfTokenFromMeta() {
  const meta = document.querySelector('meta[name="csrf-token"]');
  return meta instanceof HTMLMetaElement ? meta.content : "";
}

/** @param {string} message */
function showError(message) {
  errorEl.textContent = message;
  errorEl.hidden = false;
}

function hideError() {
  errorEl.hidden = true;
}
