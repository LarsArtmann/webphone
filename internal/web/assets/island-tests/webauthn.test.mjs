// webauthn.js under node:test — the base64url wire coercion. The server
// side of the ceremony is go-webauthn's protocol.URLEncodedBase64 (RFC
// 4648 §5, NO padding); the browser side speaks ArrayBuffer. Every bug
// here is an opaque ceremony failure on a fresh passkey login, so the
// round-trips are pinned against vectors the padding and sign-bit paths
// actually exercise (lengths 1..5 mod 4, plus a 0-length value).
import test from "node:test";
import assert from "node:assert/strict";

import {
  base64urlToBytes,
  bytesToBase64url,
  prepareLoginOptions,
  prepareRegistrationOptions,
  serializeCredential,
} from "../island/app/webauthn.js";

const bytes = (...nums) => new Uint8Array(nums);

test("base64url round-trips every length class without padding", () => {
  const vectors = [
    bytes(),
    bytes(0),
    bytes(0xff),
    bytes(0xff, 0x00),
    bytes(0xff, 0x00, 0x5a),
    bytes(0xff, 0x00, 0x5a, 0xc3),
    bytes(0xff, 0x00, 0x5a, 0xc3, 0x7e),
    bytes(1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16),
  ];
  for (const vec of vectors) {
    const encoded = bytesToBase64url(vec);
    assert.ok(!encoded.includes("="), `padded encoding: ${encoded}`);
    assert.ok(!encoded.includes("+") && !encoded.includes("/"), `not URL-safe: ${encoded}`);
    const decoded = base64urlToBytes(encoded);
    assert.deepEqual([...decoded], [...vec]);
  }
});

test("base64url accepts unpadded and padded spellings (server tolerance)", () => {
  // go-webauthn trims padding before decoding; the island tolerates a
  // padded spelling so a manually re-encoded option never hard-fails.
  // 0xff 0xff is "__8" in the URL alphabet and "//8" in the standard
  // one — both must decode to the same bytes.
  assert.deepEqual([...base64urlToBytes("__8")], [0xff, 0xff]);
  assert.deepEqual([...base64urlToBytes("__8==")], [0xff, 0xff]);
  assert.deepEqual([...base64urlToBytes("//8")], [0xff, 0xff]);
  assert.deepEqual([...base64urlToBytes("_-4")], [0xff, 0xee]);
});

test("base64url rejects garbage instead of decoding silently", () => {
  assert.throws(() => base64urlToBytes("not!base64"));
  assert.throws(() => base64urlToBytes(42));
});

test("prepareLoginOptions decodes challenge and allowCredentials ids to bytes", () => {
  const options = prepareLoginOptions({
    challenge: "AAAAAQ",
    rpId: "pbx.example.com",
    timeout: 60000,
    userVerification: "preferred",
    allowCredentials: [
      { id: "AAECAw", type: "public-key", transports: ["internal"] },
      { id: "BAUGBw", type: "public-key" },
    ],
  });
  assert.deepEqual([...options.challenge], [0, 0, 0, 1]);
  assert.equal(options.rpId, "pbx.example.com");
  assert.equal(options.timeout, 60000);
  assert.equal(options.userVerification, "preferred");
  assert.deepEqual([...options.allowCredentials[0].id], [0, 1, 2, 3]);
  assert.deepEqual(options.allowCredentials[0].transports, ["internal"]);
  assert.deepEqual([...options.allowCredentials[1].id], [4, 5, 6, 7]);
  assert.equal(options.allowCredentials[1].transports, undefined);
});

test("prepareLoginOptions tolerates a missing allowCredentials list", () => {
  const options = prepareLoginOptions({ challenge: "AA", rpId: "x" });
  assert.deepEqual([...options.challenge], [0]);
  assert.equal(options.allowCredentials, undefined);
});

test("prepareRegistrationOptions decodes challenge, user.id and excludeCredentials", () => {
  const options = prepareRegistrationOptions({
    challenge: "AAAAAQ",
    rp: { id: "pbx.example.com", name: "WebPhone" },
    user: { id: "AQIDBA", name: "lars@example.com", displayName: "Lars" },
    pubKeyCredParams: [{ type: "public-key", alg: -7 }],
    timeout: 60000,
    attestation: "none",
    authenticatorSelection: { residentKey: "preferred", userVerification: "preferred" },
    excludeCredentials: [{ id: "BAUGBw", type: "public-key" }],
  });
  assert.deepEqual([...options.challenge], [0, 0, 0, 1]);
  assert.deepEqual([...options.user.id], [1, 2, 3, 4]);
  assert.equal(options.user.name, "lars@example.com");
  assert.equal(options.user.displayName, "Lars");
  assert.deepEqual(options.rp, { id: "pbx.example.com", name: "WebPhone" });
  assert.deepEqual(options.pubKeyCredParams, [{ type: "public-key", alg: -7 }]);
  assert.deepEqual([...options.excludeCredentials[0].id], [4, 5, 6, 7]);
  assert.deepEqual(options.authenticatorSelection, {
    residentKey: "preferred",
    userVerification: "preferred",
  });
});

test("serializeCredential encodes a login assertion the server parses", () => {
  const credential = {
    id: "AAECAw",
    rawId: new Uint8Array([0, 1, 2, 3]).buffer,
    type: "public-key",
    response: {
      clientDataJSON: new Uint8Array([1, 2]).buffer,
      authenticatorData: new Uint8Array([3]).buffer,
      signature: new Uint8Array([4, 5]).buffer,
      userHandle: null,
    },
  };
  const body = serializeCredential(credential, "login");
  assert.deepEqual(body, {
    id: "AAECAw",
    rawId: "AAECAw",
    type: "public-key",
    response: {
      clientDataJSON: "AQI",
      authenticatorData: "Aw",
      signature: "BAU",
    },
  });
  // userHandle null MUST be omitted: go-webauthn's omitempty only omits
  // the empty VALUE, and a bogus "" would decode to a present-but-empty
  // handle — a discoverable-credential mismatch, not a missing one.
  assert.equal("userHandle" in body.response, false);
});

test("serializeCredential keeps a present userHandle on login", () => {
  const credential = {
    id: "AQ",
    rawId: new Uint8Array([1]).buffer,
    type: "public-key",
    response: {
      clientDataJSON: new Uint8Array([1]).buffer,
      authenticatorData: new Uint8Array([1]).buffer,
      signature: new Uint8Array([1]).buffer,
      userHandle: new Uint8Array([9, 9]).buffer,
    },
  };
  const body = serializeCredential(credential, "login");
  assert.equal(body.response.userHandle, "CQk");
});

test("serializeCredential encodes a registration attestation the server parses", () => {
  const credential = {
    id: "AAECAw",
    rawId: new Uint8Array([0, 1, 2, 3]).buffer,
    type: "public-key",
    response: {
      clientDataJSON: new Uint8Array([1]).buffer,
      attestationObject: new Uint8Array([7, 7, 7]).buffer,
      transports: ["internal", "hybrid"],
    },
  };
  const body = serializeCredential(credential, "registration");
  assert.equal(body.rawId, "AAECAw");
  assert.deepEqual(body.response, {
    clientDataJSON: "AQ",
    attestationObject: "BwcH",
    transports: ["internal", "hybrid"],
  });
  // The attestation body never carries assertion-only fields.
  assert.equal("signature" in body.response, false);
});
