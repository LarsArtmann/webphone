# The services.webphone.oauth2 seam, end to end: the typed option front
# for the app's `auth.oauth2` config (OpenID Connect login; Pocket ID in
# the telephony stack), its eval-time assertions, and the settings +
# LoadCredential wiring. Extracted from nixos-module.nix (nix-review
# monolith split, 2026-10-27) so the seam lives in ONE file; imported by
# nixos-module.nix.
#
# Why a typed front at all: the app fails closed at BOOT on half an OIDC
# configuration, and enabling OIDC turns on the shared email→extension
# mapping, which the app derives passkey mode from (and then demands
# rp_id/rp_origins/password files for). Both facts surface at EVAL time
# here instead — an incomplete SSO config fails `nixos-rebuild` with a
# legible message, never a reboot into a dead login card.
{
  config,
  lib,
  ...
}:
let
  cfg = config.services.webphone;

  inherit (import ./lib.nix { inherit lib; }) bareHTTPS;

  # The secret is handed to systemd as a credential; systemd exposes
  # credentials at $CREDENTIALS_DIRECTORY, which is the unit's
  # /run/credentials path. The app reads THIS copy, never the
  # operator's host file, so no group grant is required.
  oauth2CredentialName = "webphone-oidc-client-secret";
  oauth2SecretPath = "/run/credentials/webphone.service/${oauth2CredentialName}";

  # null = derive from the vhost; eval fails closed when neither is
  # available (an external front must declare the registration).
  oauth2RedirectUrl =
    if cfg.oauth2.redirectUrl != null then
      cfg.oauth2.redirectUrl
    else if cfg.caddy.enable then
      "https://${cfg.caddy.hostName}/api/auth/oauth2/callback"
    else
      null;

  # The identity map is shared by both passwordless modes and lives in
  # the app's ONE home for it (settings.auth.*, folded into the passkey
  # shape by the app). This module reads the same merged value so its
  # assertions reflect what the binary will actually see.
  oauth2Users = cfg.settings.auth.users or { };
  oauth2PasswordFiles =
    (cfg.settings.auth.extension_password_files or { })
    // (cfg.settings.auth.passkey.extension_password_files or { });
  oauth2MappedExtensions = lib.unique (
    lib.concatMap (user: user.extensions or [ ]) (lib.attrValues oauth2Users)
  );
  oauth2MissingPasswordFiles = lib.filter (
    ext: (oauth2PasswordFiles.${ext} or "") == ""
  ) oauth2MappedExtensions;

  # A typed OIDC config derives passkey mode in the app (the shared
  # users map feeds both modes), which then DEMANDS rp_id/rp_origins.
  # Behind the generated vhost they follow from caddy.hostName; a raw
  # value still wins (mkDefault), so an operator pin stays authoritative.
  # Deliberately keyed on the TYPED lane only: reading the merged
  # settings.auth.users from inside the settings definition would recurse.
  deriveWebAuthnRP = cfg.caddy.enable && cfg.oauth2.enable;
in
{
  options.services.webphone.oauth2 = {
    enable = lib.mkEnableOption ''
      OpenID Connect (SSO) login: the login card gains a
      "Continue to Pocket ID" entry that runs the authorization-code +
      PKCE dance against `issuer`; the server exchanges the code,
      verifies the ID token, maps the verified email through the shared
      `settings.auth.users` map and mints the same webphone session the
      passkey and extension logins mint. The extension login stays
      available, so a dead IdP never locks the phone.

      Enabling this also requires the shared mapping
      (`settings.auth.users`) and one SIP-password file per mapped
      extension (`settings.auth.extension_password_files`) — both
      asserted. When `caddy.enable` is set the module derives the
      WebAuthn rp_id/rp_origins from `caddy.hostName` and the callback
      from the same host, so the vhost and the OIDC registration cannot
      drift apart.
    '';

    issuer = lib.mkOption {
      type = lib.types.nullOr lib.types.str;
      default = null;
      example = "https://auth.example.org";
      description = ''
        The provider's OIDC discovery base URL: a bare https URL with no
        query or fragment (the ID token's `iss` claim is matched
        verbatim). Required when `enable` is set; discovery stays lazy
        inside the app, so an unreachable IdP costs a login, never the
        boot.
      '';
    };

    clientId = lib.mkOption {
      type = lib.types.str;
      default = "webphone";
      example = "webphone";
      description = ''
        Registered OIDC client id — the confidential client the IdP
        provisioner creates. Client creation stays an IdP concern; this
        module only consumes the registration.
      '';
    };

    clientSecretFile = lib.mkOption {
      type = lib.types.nullOr lib.types.str;
      default = null;
      example = "/run/secrets/webphone-oidc-client-secret";
      description = ''
        HOST path of the client secret file (single line; trailing
        whitespace trimmed). Required when `enable` is set.

        The module hands this file to systemd as a credential
        (`LoadCredential`): systemd copies it into the unit's private
        credential directory at every unit start (mode 0400, owned by the
        service user), so the app reads the copy and never the host
        path in place — no host-side group grant needed, and a missing
        source file fails the UNIT START instead of the first SSO login.

        Rotation: systemd snapshots the file at unit start, so after
        rotating the secret run `systemctl restart webphone` (the app
        itself re-reads the file per login; the copy is what pins the
        value for the unit's lifetime).

        Declared as a STRING, never a Nix `path`: a path value would be
        copied into the world-readable Nix store and publish the secret
        to every user on the machine.
      '';
    };

    redirectUrl = lib.mkOption {
      type = lib.types.nullOr lib.types.str;
      default = null;
      example = "https://phone.example.org/api/auth/oauth2/callback";
      description = ''
        The registered callback, absolute https; it must match the IdP
        registration verbatim or every login dies at the token exchange
        (invalid_redirect_uri).

        null (the default) derives
        `https://<caddy.hostName>/api/auth/oauth2/callback` when
        `caddy.enable` is set, so the vhost and the callback cannot
        drift. Behind an external front set it explicitly; eval fails
        closed when neither is available.
      '';
    };

    scopes = lib.mkOption {
      type = lib.types.listOf lib.types.str;
      default = [
        "openid"
        "email"
        "profile"
      ];
      description = ''
        OIDC scopes. MUST include "openid" (without it the IdP returns
        no id_token and no login can complete); "email" is what the
        identity mapping consumes.
      '';
    };

    allowedGroups = lib.mkOption {
      type = lib.types.listOf lib.types.str;
      default = [ ];
      example = [ "pbx-phone" ];
      description = ''
        In-app IdP group allowlist. Empty (the default) disables the
        check — admission stays with the users mapping and the IdP-side
        client-group link. Non-empty, every SSO login must carry at
        least one of these groups in its token claims or it is denied
        (fail-closed: a token with no groups claim at all is denied
        too). A second gate on top of the IdP link, not a replacement.
      '';
    };
  };

  config = lib.mkIf cfg.enable {
    # Eval-time mirror of the app's fail-closed OIDC validation: an
    # enabled-but-half SSO config must fail `nixos-rebuild`, never a
    # reboot into a login card whose every attempt dies.
    assertions =
      lib.optionals cfg.oauth2.enable [
        {
          assertion = cfg.oauth2.issuer != null && bareHTTPS cfg.oauth2.issuer;
          message = ''
            services.webphone.oauth2.issuer must be a bare https URL with no
            query or fragment (got `${
              if cfg.oauth2.issuer == null then "unset" else cfg.oauth2.issuer
            }`): the IdP's `iss` claim is matched verbatim.
          '';
        }
        {
          assertion = cfg.oauth2.clientSecretFile != null;
          message = ''
            services.webphone.oauth2.clientSecretFile is required when
            enable = true — the client secret never lives in the world-readable
            config JSON or the Nix store; point it at the runtime file.
          '';
        }
        {
          assertion = lib.elem "openid" cfg.oauth2.scopes;
          message = ''
            services.webphone.oauth2.scopes must include "openid" — without it
            the IdP returns no id_token and no login can complete.
          '';
        }
        {
          assertion = oauth2RedirectUrl != null && bareHTTPS oauth2RedirectUrl;
          message = ''
            services.webphone.oauth2: the callback cannot be derived — enable
            caddy (derives https://<caddy.hostName>/api/auth/oauth2/callback)
            or set redirectUrl to the registered absolute https callback.
          '';
        }
        {
          assertion = oauth2Users != { };
          message = ''
            services.webphone.oauth2.enable requires settings.auth.users (the
            shared email→extension mapping) — without it every SSO login is
            unmapped and denied.
          '';
        }
        {
          assertion =
            (cfg.settings.auth.passkey.rp_id or "") != ""
            && (cfg.settings.auth.passkey.rp_origins or [ ]) != [ ];
          message = ''
            services.webphone.oauth2.enable turns on the shared identity layer,
            which the app derives passkey mode from — settings.auth.passkey.rp_id
            and rp_origins must resolve. Enable caddy to derive them from
            caddy.hostName, or set both explicitly behind an external front.
          '';
        }
        {
          assertion = oauth2MissingPasswordFiles == [ ];
          message = ''
            services.webphone.oauth2: every mapped extension needs a SIP
            directory password file in settings.auth.extension_password_files
            (missing: ${lib.concatStringsSep ", " oauth2MissingPasswordFiles}).
          '';
        }
      ]
      ++ lib.optionals (!cfg.oauth2.enable) [
        {
          assertion =
            cfg.oauth2.issuer == null
            && cfg.oauth2.clientSecretFile == null
            && cfg.oauth2.redirectUrl == null
            && cfg.oauth2.allowedGroups == [ ];
          message = ''
            services.webphone.oauth2: set enable = true — issuer,
            clientSecretFile, redirectUrl or allowedGroups without enable would
            silently do nothing.
          '';
        }
      ];

    services.webphone.settings.auth = lib.mkIf cfg.oauth2.enable (lib.mkMerge [
      (lib.mkIf deriveWebAuthnRP {
        # mkDefault: an explicitly declared settings.auth.passkey.*
        # value wins, so an operator pin stays authoritative.
        passkey = {
          rp_id = lib.mkDefault cfg.caddy.hostName;
          rp_origins = lib.mkDefault [ "https://${cfg.caddy.hostName}" ];
        };
      })
      {
        # PLAIN definitions, deliberately: a raw
        # settings.auth.oauth2.* of the same key makes the eval fail
        # with a conflict (declare the OIDC client shape ONCE) instead
        # of one lane silently winning.
        oauth2 = {
          issuer = cfg.oauth2.issuer;
          client_id = cfg.oauth2.clientId;
          client_secret_file = oauth2SecretPath;
          redirect_url = oauth2RedirectUrl;
          scopes = cfg.oauth2.scopes;
          allowed_groups = cfg.oauth2.allowedGroups;
        };
      }
    ]);

    systemd.services.webphone.serviceConfig.LoadCredential =
      lib.mkIf cfg.oauth2.enable [ "${oauth2CredentialName}:${cfg.oauth2.clientSecretFile}" ];
  };
}
