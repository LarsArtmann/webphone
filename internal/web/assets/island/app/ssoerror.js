// The ?sso_error card message. Read-then-strip: the sentence renders
// from the query parameter ONCE, then the parameter leaves the address
// bar (history.replaceState) so a reload or a shared link never
// re-shows a stale failure. Without JS the card element stays hidden —
// the strip is address-bar hygiene, never a render dependency.
export function showSSOError(t) {
  const params = new URLSearchParams(location.search ?? "");
  const code = params.get("sso_error");
  if (!code) return;
  const el = document.getElementById("sso-error");
  if (el) {
    const key = "ssoError" + code.charAt(0).toUpperCase() + code.slice(1);
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
