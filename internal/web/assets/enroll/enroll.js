// The standalone passkey-enrollment flow (served at /assets/enroll/;
// the page is views.EnrollPage — a one-purpose surface with NO island
// runtime). Flow: verify AND burn the one-time token → begin the
// registration ceremony → navigator.credentials.create → finish → the
// account can passkey-log in. Every failure class answers ONE message
// (no token-state oracle server-side, no status-splitting client-side);
// a cancelled prompt keeps the token's fate honest: verify burned it,
// so retry needs a fresh link — the copy says so.
import { t } from "../island/app/i18n.js";
import { csrfToken } from "../island/app/csrf.js";
import { prepareRegistrationOptions, serializeCredential } from "./webauthn.js";

/**
 * The enroll page's DOM handle lookup: every id is served markup
 * (views/enroll.templ), so the result is non-null by contract and a typo'd
 * id is a compile error instead of a silent `null`.
 * @template {keyof IslandEnrollDomIds} K
 * @param {K} id
 * @returns {IslandEnrollDomIds[K]}
 */
const $ = (id) => /** @type {IslandEnrollDomIds[K]} */ (document.getElementById(id));

const form = $("enroll-form");
const tokenInput = $("enroll-token");
const nameInput = $("enroll-credential-name");
const statusEl = $("enroll-status");
const errorEl = $("enroll-error");

// The CLI-minted link carries the token in the query; prefill it so the
// user never retypes what the URL already knows.
const params = new URLSearchParams(location.search || "");
const urlToken = params.get("token");
if (urlToken) tokenInput.value = urlToken;

form.addEventListener("submit", (event) => {
  event.preventDefault();
  enroll().catch((err) => {
    showError(t("enrollNetFailed"));
    console.error("webphone enroll:", err);
  });
});

async function enroll() {
  hideError();
  const submit = /** @type {HTMLButtonElement | null} */ (
    form.querySelector('button[type="submit"]')
  );
  if (submit) submit.disabled = true;
  try {
    const token = tokenInput.value.trim();
    if (!token) {
      showError(t("enrollTokenMissing"));
      return;
    }

    /** @type {IslandPostJSONAnswer & {body: IslandEnrollVerifyBody}} */
    const verified = await postJSON("/api/auth/passkey/enroll/verify", {
      token,
    });
    if (!verified.ok) {
      showError(enrollError(verified.status));
      return;
    }
    setStatus(t("enrollVerified")(verified.body.email));

    /** @type {IslandPostJSONAnswer & {body: IslandEnrollBeginBody}} */
    const begun = await postJSON("/api/auth/passkey/enroll/begin", {
      user_id: verified.body.user_id,
    });
    if (!begun.ok) {
      showError(enrollError(begun.status));
      return;
    }

    // The wire dictionary becomes the browser's creation options here; the
    // cast is the boundary, not an escape hatch: the server's JSON IS the
    // PublicKeyCredentialCreationOptions the spec names, minus the base64
    // decoding prepareRegistrationOptions just did.
    const publicKey = /** @type {PublicKeyCredentialCreationOptions} */ (
      /** @type {unknown} */ (prepareRegistrationOptions(begun.body.options))
    );
    /** @type {Credential | null} */
    let credential;
    try {
      credential = await navigator.credentials.create({ publicKey });
    } catch {
      showError(t("enrollDismissed"));
      return;
    }
    // A null answer means the user agent declined without throwing — same
    // fate as a cancelled prompt: the token is spent, ask for a new link.
    if (!credential) {
      showError(t("enrollDismissed"));
      return;
    }

    const credentialName = (nameInput.value || "").trim();
    const query =
      `?user_id=${encodeURIComponent(begun.body.session_key)}` +
      `&credential_name=${encodeURIComponent(credentialName)}`;
    const finished = await postJSON(
      "/api/auth/passkey/enroll/finish" + query,
      serializeCredential(/** @type {PublicKeyCredential} */ (credential), "registration"),
    );
    if (!finished.ok) {
      showError(enrollError(finished.status));
      return;
    }
    form.hidden = true;
    setStatus(t("enrollSuccess"));
  } finally {
    if (submit) submit.disabled = false;
  }
}

// Status 0 is postJSON's network-failure sentinel — the honest copy for
// it is the network message, never "HTTP 0". Every other status rides
// the one anti-oracle failure message.
/** @param {number} status @returns {string} */
const enrollError = (status) => (status === 0 ? t("enrollNetFailed") : t("enrollFailed")(status));

// postJSON is the one fetch shape this page uses: JSON body, CSRF
// header, parsed JSON answer. Never throws — network failures come back
// as {ok: false, status: 0} and map to the network message.
/**
 * @param {string} url
 * @param {unknown} body
 * @returns {Promise<IslandPostJSONAnswer>}
 */
async function postJSON(url, body) {
  try {
    const res = await fetch(url, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "X-CSRF-Token": csrfToken(),
      },
      body: JSON.stringify(body),
    });
    /** @type {any} */
    let parsed = {};
    try {
      parsed = await res.json();
    } catch {
      // a bodyless error is still a status verdict
    }
    return { ok: res.ok, status: res.status, body: parsed };
  } catch {
    return { ok: false, status: 0, body: {} };
  }
}

/** @param {string} text */
function setStatus(text) {
  statusEl.textContent = text;
  statusEl.hidden = false;
}

/** @param {string} text */
function showError(text) {
  errorEl.textContent = text;
  errorEl.hidden = false;
}

function hideError() {
  errorEl.hidden = true;
}
