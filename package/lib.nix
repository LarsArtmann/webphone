# Shared URL validation for the typed option seams (oauth2.nix, crm.nix):
# the eval-time mirrors of the app's boot-time URL checks. One home, two
# consumers (nix-review split, 2026-10-27).
{ lib }:
{
  # A bare absolute https URL with an optional path component — no query
  # string, no fragment (issuer/client URLs are matched or resolved
  # verbatim by the app).
  bareHTTPS = url: lib.match "https://[^/?#]+(/[^?#]*)?" url != null;
}
