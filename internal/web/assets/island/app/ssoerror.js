// The ?sso_error card message. Read-then-strip: the sentence renders
// from the query parameter ONCE, then the parameter leaves the address
// bar (history.replaceState) so a reload or a shared link never
// re-shows a stale failure. Without JS the card element stays hidden —
// the strip is address-bar hygiene, never a render dependency.
// The server's ?sso_error=<code> vocabulary (ssoErrorCode in sso_api.go)
// as message keys. A static map, not string concatenation: the key list is
// then checkable at compile time, and an unknown code falls back to the
// generic sentence exactly as the old suffix-join did at runtime.
/** @type {Record<string, IslandMessageKey>} */
const SSO_ERROR_KEYS = {
  unmapped: "ssoErrorUnmapped",
  directory: "ssoErrorDirectory",
  unverified: "ssoErrorUnverified",
  groups: "ssoErrorGroups",
  stale: "ssoErrorStale",
  config: "ssoErrorConfig",
  unavailable: "ssoErrorUnavailable",
  failed: "ssoErrorFailed",
};

/**
 * @param {(key: IslandMessageKey) => any} t
 */
export function showSSOError(t) {
  const params = new URLSearchParams(location.search ?? "");
  const code = params.get("sso_error");
  if (!code) return;
  const el = document.getElementById("sso-error");
  if (el) {
    const key = SSO_ERROR_KEYS[code] ?? "ssoErrorFailed";
    const message = t(key);
    el.textContent =
      typeof message === "string" ? message : t("ssoErrorFailed");
    el.hidden = false;
  }
  params.delete("sso_error");
  const rest = params.toString();
  history.replaceState(
    {},
    "",
    location.pathname + (rest ? "?" + rest : "") + (location.hash ?? ""),
  );
}
