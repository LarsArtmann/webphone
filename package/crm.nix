# The services.webphone.crm seam, end to end: the typed option front for
# the app's `crm.url` + `crm.token_file` config (the Ledger CRM
# integration: caller-name enrichment and call-journal logging), its
# eval-time assertions, and the settings + LoadCredential wiring.
# Extracted from nixos-module.nix (nix-review monolith split, 2026-10-27)
# so the seam lives in ONE file; imported by nixos-module.nix.
#
# Why a typed front at all: the app fails closed at BOOT on half a CRM
# configuration, and the bearer token must never sit in the
# world-readable config JSON or the Nix store. Both facts surface at
# EVAL time here — an incomplete or leaky CRM config fails
# `nixos-rebuild` with a legible message, never a silent dark seam or a
# published secret.
{
  config,
  lib,
  ...
}:
let
  cfg = config.services.webphone;

  inherit (import ./lib.nix { inherit lib; }) bareHTTPS;

  # Same credential doctrine as oauth2.nix: systemd copies the
  # operator's file into the unit's private credential directory at
  # start, and the app reads that copy.
  crmCredentialName = "webphone-crm-token";
  crmSecretPath = "/run/credentials/webphone.service/${crmCredentialName}";
in
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
        credential directory at every unit start (mode 0400, owned by the
        service user), so the app reads the copy and never the host
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

  config = lib.mkIf cfg.enable {
    # Eval-time mirror of the app's fail-closed CRM validation: an
    # enabled-but-half CRM config must fail `nixos-rebuild`, never boot
    # a dark seam the operator believes is on.
    assertions =
      lib.optionals cfg.crm.enable [
        {
          assertion = cfg.crm.url != null && bareHTTPS cfg.crm.url;
          message = ''
            services.webphone.crm.url must be an absolute https URL (got ${
              if cfg.crm.url == null then "unset" else cfg.crm.url
            }): the machine API answers 401 without TLS in any real
            deployment, and a relative URL cannot resolve.
          '';
        }
        {
          assertion = cfg.crm.tokenFile != null;
          message = ''
            services.webphone.crm.tokenFile is required when enable = true —
            the bearer token never lives in the world-readable config JSON or
            the Nix store; point it at the runtime file.
          '';
        }
      ]
      ++ lib.optionals (!cfg.crm.enable) [
        {
          assertion = cfg.crm.url == null && cfg.crm.tokenFile == null;
          message = ''
            services.webphone.crm: set enable = true — url or tokenFile
            without enable would silently do nothing.
          '';
        }
      ];

    # PLAIN definitions, deliberately: a raw settings.crm.* of the same
    # key makes the eval fail with a conflict (one home per fact), and
    # when the typed lane is OFF this block is absent entirely — the
    # raw lane stays untouched.
    services.webphone.settings.crm = lib.mkIf cfg.crm.enable {
      url = cfg.crm.url;
      token_file = crmSecretPath;
    };

    systemd.services.webphone.serviceConfig.LoadCredential = lib.mkIf cfg.crm.enable [
      "${crmCredentialName}:${cfg.crm.tokenFile}"
    ];
  };
}
