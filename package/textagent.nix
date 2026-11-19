# The services.webphone.textagent seam, end to end: the typed option
# front for the app's `textagent.url` + `textagent.secret_file` config
# (the post-call AI chain: summarize a call's transcript into an SMS
# follow-up draft in the review queue), its eval-time assertions, and
# the settings + LoadCredential wiring. Mirrors crm.nix (the typed-front
# pattern, 2026-11-15 — the summarize chain made the seam first-class).
#
# Why a typed front: the app fails closed at BOOT on half a textagent
# config, and the review API's bearer secret must never sit in the
# world-readable config JSON or the Nix store. Both facts surface at
# EVAL time here — an incomplete or leaky textagent config fails
# `nixos-rebuild` with a legible message, never a silent dark seam.
#
# Unlike the CRM lane, plain http is FINE here: the review API rides
# the consuming stack on loopback (http://127.0.0.1:8071) — the URL
# check mirrors the app exactly (absolute http(s), host required), not
# the CRM https-only rule.
{
  config,
  lib,
  ...
}:
let
  cfg = config.services.webphone;

  # Same credential doctrine as crm.nix: systemd copies the operator's
  # file into the unit's private credential directory at start, and the
  # app reads that copy.
  textagentCredentialName = "webphone-textagent-secret";
  textagentSecretPath = "/run/credentials/webphone.service/${textagentCredentialName}";

  # The app's own URL shape: absolute http(s), non-empty host, optional
  # path — no query, no fragment. builtins.match anchors implicitly.
  bareHTTPish = url: builtins.match "https?://[^/?#]+(/[^?#]*)?" (toString url) != null;
in
{
  options.services.webphone.textagent = {
    enable = lib.mkEnableOption ''
      the post-call AI chain: finished calls summarize into an SMS
      follow-up draft in the review queue (the agent pill, the ended
      card, History's review artifacts). The review API's bearer secret
      rides a systemd credential (LoadCredential), so the app reads a
      0440 copy owned by the service user and no host-side group grant
      is needed.

      The seam stays optional by design: with it off the agent surfaces
      render their unconfigured states (the `wp-review-unconfigured`
      chip, no ended-card Summarize button) — never a broken affordance.
    '';

    url = lib.mkOption {
      type = lib.types.nullOr lib.types.str;
      default = null;
      example = "http://127.0.0.1:8071";
      description = ''
        The review API's base URL: an absolute http(s) URL (plain http
        is the NORMAL shape here — the API rides the consuming stack on
        loopback). Required when `enable` is set.
      '';
    };

    secretFile = lib.mkOption {
      type = lib.types.nullOr lib.types.str;
      default = null;
      example = "/run/secrets/webphone-textagent-secret";
      description = ''
        HOST path of the review API's bearer secret file (single line;
        trailing whitespace trimmed). Required when `enable` is set.

        The module hands this file to systemd as a credential
        (`LoadCredential`): systemd copies it into the unit's private
        credential directory at every unit start (mode 0440, owned by
        the service user), so the app reads the copy and never the host
        path in place — a missing source file fails the UNIT START
        instead of the first summarize.

        Rotation: systemd snapshots the file at unit start, so after
        rotating the secret run `systemctl restart webphone`.

        Declared as a STRING, never a Nix `path`: a path value would be
        copied into the world-readable Nix store and publish the secret
        to every user on the machine.
      '';
    };
  };

  config = lib.mkIf cfg.enable {
    # Eval-time mirror of the app's fail-closed textagent validation: an
    # enabled-but-half config must fail `nixos-rebuild`, never boot a
    # dark seam the operator believes is on.
    assertions =
      lib.optionals cfg.textagent.enable [
        {
          assertion = cfg.textagent.url != null && bareHTTPish cfg.textagent.url;
          message = ''
            services.webphone.textagent.url must be an absolute http(s)
            URL (got ${
              if cfg.textagent.url == null then "unset" else cfg.textagent.url
            }): the review API rides the consuming stack (typically
            http://127.0.0.1:8071), and a relative URL cannot resolve.
          '';
        }
        {
          assertion = cfg.textagent.secretFile != null;
          message = ''
            services.webphone.textagent.secretFile is required when
            enable = true — the review API rejects anonymous callers,
            and the secret never lives in the world-readable config
            JSON or the Nix store; point it at the runtime file.
          '';
        }
      ]
      ++ lib.optionals (!cfg.textagent.enable) [
        {
          assertion = cfg.textagent.url == null && cfg.textagent.secretFile == null;
          message = ''
            services.webphone.textagent: set enable = true — url or
            secretFile without enable would silently do nothing.
          '';
        }
      ];

    # PLAIN definitions, deliberately: a raw settings.textagent.* of the
    # same key makes the eval fail with a conflict (one home per fact),
    # and when the typed lane is OFF this block is absent entirely — the
    # raw lane stays untouched.
    services.webphone.settings.textagent = lib.mkIf cfg.textagent.enable {
      url = cfg.textagent.url;
      secret_file = textagentSecretPath;
    };

    systemd.services.webphone.serviceConfig.LoadCredential = lib.mkIf cfg.textagent.enable [
      "${textagentCredentialName}:${cfg.textagent.secretFile}"
    ];
  };
}
