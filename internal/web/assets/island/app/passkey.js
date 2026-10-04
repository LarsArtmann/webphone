// Passkey (WebAuthn) login: the email-first front door for deployments
// with the identity layer enabled. The order is INVERTED versus the
// extension login: the finish endpoint mints the server session (the
// SIP directory password is sourced and verified server-side, fail
// closed), and only then does the island REGISTER with the returned
// credentials — so a stale password file can never produce a half
// login. The markup contract (#passkey-login-form, #passkey-email,
// #passkey-login-error) is rendered by the shell ONLY when the mode is
// on; initPasskeyLogin no-ops without it, keeping a disabled deployment
// byte-identical.
import { connect, networkOnline, registerWasRejected } from "./connection.js";
import { sipDomain } from "./config.js";
import { t } from "./i18n.js";
import { adoptServerSession, signOutQuiet } from "./session.js";
import { els, log } from "./ui.js";
import { prepareLoginOptions, serializeCredential } from "./webauthn.js";

export function initPasskeyLogin() {
  if (!els.passkeyForm) return; // passkey mode off — no form was rendered
  els.passkeyForm.addEventListener("submit", (event) => {
    event.preventDefault();
    passkeyLogin().catch((err) => {
      // Everything user-visible is handled inside; this is the wiring
      // itself failing, which the operator log deserves raw.
      log(`passkey login crashed (${err.message})`, "error");
    });
  });
}

async function passkeyLogin() {
  els.passkeyError.hidden = true;
  const email = els.passkeyEmail.value.trim();
  if (!email) return;
  const submit = els.passkeyForm.querySelector('button[type="submit"]');
  if (submit) submit.disabled = true;
  try {
    const finish = await runCeremony(email);
    if (!finish) return; // the error is already where the user is looking
    await registerIsland(finish);
  } finally {
    if (submit) submit.disabled = false;
  }
}

// runCeremony drives begin → navigator.credentials.get → finish and
// returns the finish payload ({extension, password, did, display_name,
// numbers}) on success, null on every handled failure (the message is
// already rendered in #passkey-login-error).
async function runCeremony(email) {
  let data;
  try {
    const res = await fetch("/api/auth/passkey/begin", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "X-CSRF-Token": csrfToken(),
      },
      body: JSON.stringify({ email }),
    });
    if (res.status === 401) {
      showLoginError(t("passkeyUnknown"));
      return null;
    }
    if (res.status === 429) {
      showLoginError(t("passkeyThrottled"));
      return null;
    }
    if (!res.ok) {
      showLoginError(t("passkeyBeginFailed")(res.status));
      return null;
    }
    data = await res.json();
  } catch (err) {
    showLoginError(t("passkeyNetFailed"));
    log(`passkey begin failed (${err.message})`, "error");
    return null;
  }

  let credential;
  try {
    credential = await navigator.credentials.get({
      publicKey: prepareLoginOptions(data.options),
    });
  } catch (err) {
    // The user dismissed the prompt or it timed out: not an error to
    // surface — they will see the form again. The log keeps the class.
    log(`passkey prompt dismissed (${err.name})`);
    return null;
  }

  try {
    const res = await fetch(
      `/api/auth/passkey/finish?user_id=${encodeURIComponent(data.session_key)}`,
      {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          "X-CSRF-Token": csrfToken(),
        },
        body: JSON.stringify(serializeCredential(credential, "login")),
      },
    );
    if (res.status === 401) {
      showLoginError(t("passkeyUnknown"));
      return null;
    }
    if (res.status === 429) {
      showLoginError(t("passkeyThrottled"));
      return null;
    }
    if (!res.ok) {
      showLoginError(t("passkeyFinishFailed")(res.status));
      return null;
    }
    return await res.json();
  } catch (err) {
    showLoginError(t("passkeyNetFailed"));
    log(`passkey finish failed (${err.message})`, "error");
    return null;
  }
}

// registerIsland completes the login with the minted session's
// credentials: adopt the rotated CSRF + SSE + panels event, REGISTER
// with the returned SIP credentials, then show the phone with the
// user's real identity. A transport failure is not fatal (the resume
// path's honest-offline posture); a REGISTER rejection means the
// directory changed mid-ceremony — drop the now-stale session quietly
// and say why.
async function registerIsland(data) {
  els.whoami.textContent = whoamiLine(data);
  const adopted = await adoptServerSession({
    did: data.did || "",
    displayName: data.display_name || "",
    numbers: Array.isArray(data.numbers) ? data.numbers : [],
  });
  try {
    await connect(data.extension, data.password);
    log("connected via passkey login");
  } catch (err) {
    if (registerWasRejected()) {
      await signOutQuiet();
      els.whoami.textContent = "";
      showLoginError(t("resumeRejected")(err.message));
      log(
        `passkey credentials rejected; server session dropped (${err.message})`,
        "error",
      );
      return;
    }
    networkOnline();
    log(`passkey session live; SIP unreachable (${err.message})`, "error");
  }
  els.loginView.hidden = true;
  els.phoneView.hidden = false;
}

// whoamiLine renders the signed-in line for an identity-bearing session
// payload: the display name plus the user's numbers when the deployment
// knows them, the plain extension otherwise. Exported for the resume
// path (a passkey session resumed from its cookie carries the same
// fields).
export function whoamiLine(data) {
  const numbers = Array.isArray(data.numbers)
    ? data.numbers.filter(Boolean)
    : [];
  if (data.display_name) {
    const tail = numbers.length
      ? numbers.join(", ")
      : `${data.extension}@${sipDomain}`;
    return `${data.display_name} · ${tail}`;
  }
  if (numbers.length) {
    return `${data.extension}@${sipDomain} · ${numbers.join(", ")}`;
  }
  return `${data.extension}@${sipDomain}`;
}

function showLoginError(message) {
  els.passkeyError.textContent = message;
  els.passkeyError.hidden = false;
}

// The server renders the CSRF token into <meta name="csrf-token">; the
// nosurf double-submit cookie pairs with it (same contract as
// session.js — the meta is shared state, not a private copy).
function csrfToken() {
  const meta = document.querySelector('meta[name="csrf-token"]');
  return meta ? meta.getAttribute("content") : "";
}
