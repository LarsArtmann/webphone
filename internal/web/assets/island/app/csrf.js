// The CSRF token's ONE home. The server renders it into
// <meta name="csrf-token"> and the nosurf double-submit cookie pairs
// with it; every module that POSTs (island session/passkey surfaces and
// the standalone enroll page) imports this reader — the meta is shared
// state, not a private copy per feature.
export function csrfToken() {
  const meta = document.querySelector('meta[name="csrf-token"]');
  return meta ? meta.getAttribute("content") : "";
}
