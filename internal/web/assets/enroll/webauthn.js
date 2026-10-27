// WebAuthn wire coercion: the server half of the ceremony is
// go-webauthn's protocol.URLEncodedBase64 — RFC 4648 §5 base64url,
// unpadded — while the browser half (navigator.credentials) speaks
// ArrayBuffer. Every field crossing that boundary is converted here, in
// the enroll page's own module (the island's login ceremony client now
// lives in the cqrs-htmx/loginpage library page at /login — its JS owns
// its own coercion). Pure functions, no DOM: the node:test suite pins
// the round-trips.

const base64url = {
  // decode loosens the padding requirement deliberately: go-webauthn
  // trims "=" before decoding, so a padded spelling (a proxy, a manual
  // re-encode) still decodes instead of hard-failing the ceremony.
  /**
   * @param {unknown} value
   * @returns {Uint8Array}
   */
  decode(value) {
    if (typeof value !== "string") {
      throw new Error(
        "webauthn: expected base64url string, got " + typeof value,
      );
    }
    const normalized = value
      .replace(/=+$/, "")
      .replace(/-/g, "+")
      .replace(/_/g, "/");
    const binary = atob(normalized);
    const out = new Uint8Array(binary.length);
    for (let i = 0; i < binary.length; i += 1) out[i] = binary.charCodeAt(i);
    return out;
  },
  /**
   * @param {ArrayBuffer | Uint8Array} buffer
   * @returns {string}
   */
  encode(buffer) {
    const view = buffer instanceof Uint8Array ? buffer : new Uint8Array(buffer);
    let binary = "";
    for (const byte of view) binary += String.fromCharCode(byte);
    return btoa(binary)
      .replace(/\+/g, "-")
      .replace(/\//g, "_")
      .replace(/=+$/, "");
  },
};

/** @param {string} value @returns {Uint8Array} */
export function base64urlToBytes(value) {
  return base64url.decode(value);
}

/** @param {ArrayBuffer | Uint8Array} buffer @returns {string} */
export function bytesToBase64url(buffer) {
  return base64url.encode(buffer);
}

/** @param {string | ArrayBuffer | Uint8Array} value @returns {ArrayBuffer | Uint8Array} */
const toBytes = (value) =>
  value instanceof ArrayBuffer || value instanceof Uint8Array
    ? value
    : base64url.decode(value);

// prepareRegistrationOptions turns the server's
// PublicKeyCredentialCreationOptions JSON into the dictionary
// navigator.credentials.create accepts: challenge, user.id and each
// excludeCredentials[].id become byte views.
/**
 * @param {IslandRegistrationOptionsWire} options
 * @returns {Record<string, unknown>}
 */
export function prepareRegistrationOptions(options) {
  /** @type {Record<string, unknown>} */
  const prepared = { ...options, challenge: toBytes(options.challenge) };
  if (options.user) {
    prepared.user = { ...options.user, id: toBytes(options.user.id) };
  }
  if (Array.isArray(options.excludeCredentials)) {
    prepared.excludeCredentials = options.excludeCredentials.map((entry) => ({
      ...entry,
      id: toBytes(entry.id),
    }));
  }
  return prepared;
}

// serializeCredential turns the browser's PublicKeyCredential into the
// JSON body the finish endpoints parse: rawId and every response byte
// field become base64url strings. A null userHandle is OMITTED — the
// server's omitempty only forgives the empty value, and a bogus "" would
// decode as a present-but-empty handle (a discoverable-credential
// mismatch, not a missing one).
/**
 * @param {PublicKeyCredential} credential
 * @param {"login" | "registration"} kind the ceremony the credential came
 *   from — the serialization is identical, so this only documents the call
 * @returns {IslandSerializedCredential}
 */
export function serializeCredential(credential, kind) {
  /** @type {Record<string, string | string[]>} */
  const response = {};
  const source = /** @type {IslandCredentialResponseFields} */ (
    /** @type {unknown} */ (credential.response)
  );
  /** @type {Array<"clientDataJSON" | "authenticatorData" | "signature" | "attestationObject">} */
  const byteFields = [
    "clientDataJSON",
    "authenticatorData",
    "signature",
    "attestationObject",
  ];
  for (const key of byteFields) {
    const bytes = source[key];
    if (bytes != null) response[key] = base64url.encode(bytes);
  }
  if (source.userHandle != null)
    response.userHandle = base64url.encode(source.userHandle);
  if (Array.isArray(source.transports))
    response.transports = [...source.transports];
  return {
    id: credential.id,
    rawId: base64url.encode(credential.rawId),
    type: credential.type,
    response,
  };
}
