// WebAuthn wire coercion: the server half of the ceremony is
// go-webauthn's protocol.URLEncodedBase64 — RFC 4648 §5 base64url,
// unpadded — while the browser half (navigator.credentials) speaks
// ArrayBuffer. Every field crossing that boundary is converted here, in
// the one module both ceremony flows (login, enrollment) share. Pure
// functions, no DOM: the node:test suite pins the round-trips.

const base64url = {
  // decode loosens the padding requirement deliberately: go-webauthn
  // trims "=" before decoding, so a padded spelling (a proxy, a manual
  // re-encode) still decodes instead of hard-failing the ceremony.
  decode(value) {
    if (typeof value !== "string") {
      throw new Error("webauthn: expected base64url string, got " + typeof value);
    }
    const normalized = value.replace(/=+$/, "").replace(/-/g, "+").replace(/_/g, "/");
    const binary = atob(normalized);
    const out = new Uint8Array(binary.length);
    for (let i = 0; i < binary.length; i += 1) out[i] = binary.charCodeAt(i);
    return out;
  },
  encode(buffer) {
    const view = buffer instanceof Uint8Array ? buffer : new Uint8Array(buffer);
    let binary = "";
    for (const byte of view) binary += String.fromCharCode(byte);
    return btoa(binary).replace(/\+/g, "-").replace(/\//g, "_").replace(/=+$/, "");
  },
};

export function base64urlToBytes(value) {
  return base64url.decode(value);
}

export function bytesToBase64url(buffer) {
  return base64url.encode(buffer);
}

const toBytes = (value) =>
  value instanceof ArrayBuffer || value instanceof Uint8Array
    ? value
    : base64url.decode(value);

// prepareLoginOptions turns the server's PublicKeyCredentialRequestOptions
// JSON into the dictionary navigator.credentials.get accepts: challenge
// and each allowCredentials[].id become byte views. Fields the server
// omitted stay omitted — the browser fills its own defaults.
export function prepareLoginOptions(options) {
  const prepared = { ...options, challenge: toBytes(options.challenge) };
  if (Array.isArray(options.allowCredentials)) {
    prepared.allowCredentials = options.allowCredentials.map((entry) => ({
      ...entry,
      id: toBytes(entry.id),
    }));
  }
  return prepared;
}

// prepareRegistrationOptions turns the server's
// PublicKeyCredentialCreationOptions JSON into the dictionary
// navigator.credentials.create accepts: challenge, user.id and each
// excludeCredentials[].id become byte views.
export function prepareRegistrationOptions(options) {
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
export function serializeCredential(credential, kind) {
  const response = {};
  const source = credential.response;
  for (const key of ["clientDataJSON", "authenticatorData", "signature", "attestationObject"]) {
    if (source[key] != null) response[key] = base64url.encode(source[key]);
  }
  if (source.userHandle != null) response.userHandle = base64url.encode(source.userHandle);
  if (Array.isArray(source.transports)) response.transports = [...source.transports];
  return {
    id: credential.id,
    rawId: base64url.encode(credential.rawId),
    type: credential.type,
    response,
  };
}
