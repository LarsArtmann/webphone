# The services.webphone.crm option surface: the typed front for the
# app's `crm.url` + `crm.token_file` config (the Ledger CRM integration:
# caller-name enrichment and call-journal logging). Split out of
# options.nix so each file stays focused (nix-review monolith
# guideline); imported by nixos-module.nix.
#
# Why a typed front at all: the app fails closed at BOOT on half a CRM
# configuration, and the bearer token must never sit in the
# world-readable config JSON or the Nix store. Both facts surface at
# EVAL time here — an incomplete or leaky CRM config fails
# `nixos-rebuild` with a legible message, never a silent dark seam or a
# published secret.
{ lib, ... }:
{
  options.services.webphone.crm = {
    enable = lib.mkEnableOption ''
      the Ledger CRM integration: caller names enrich the call and
      message surfaces, finished calls journal on the matching contact.
      The bearer token rides a systemd credential (LoadCredential), so
      the app reads a 0400 copy owned by the service user and no
      host-side group grant is needed.

      The integration stays optional and observer-grade by design: a
      dead CRM costs toasts and un-enriched names, never call flow.
    '';

    url = lib.mkOption {
      type = lib.types.nullOr lib.types.str;
      default = null;
      example = "https://crm.example.org";
      description = ''
        The CRM's base URL: an absolute https URL (plain http only for
        a loopback deployment). Required when `enable` is set.
      '';
    };

    tokenFile = lib.mkOption {
      type = lib.types.nullOr lib.types.str;
      default = null;
      example = "/run/secrets/webphone-crm-token";
      description = ''
        HOST path of the CRM bearer token file (single line; trailing
        whitespace trimmed). Required when `enable` is set.

        The module hands this file to systemd as a credential
        (`LoadCredential`): systemd copies it into the unit's private
        credential directory at every unit start (mode 0400, owned by
        the service user), so the app reads the copy and never the host
        path in place — a missing source file fails the UNIT START
        instead of the first enrichment.

        Rotation: systemd snapshots the file at unit start, so after
        rotating the token run `systemctl restart webphone`.

        Declared as a STRING, never a Nix `path`: a path value would be
        copied into the world-readable Nix store and publish the token
        to every user on the machine.
      '';
    };
  };
}
