# The services.webphone.passkey lane: the typed front for PASSKEY-ONLY
# deployments (plan T11, 2026-11-15). The OIDC lane (oauth2.nix) already
# derives the WebAuthn rp_id/rp_origins from `caddy.hostName` — but its
# derivation is keyed on `cfg.oauth2.enable`, so an operator running
# passkeys WITHOUT SSO had to hand-write
# `settings.auth.passkey.{rp_id,rp_origins}` and could not get the
# derived-rp convenience at all. This lane gives that deployment the
# same typed door: `enable` turns the identity layer on (the shared
# `settings.auth.users` map + one SIP-password file per mapped
# extension — both asserted), and behind the generated vhost the rp
# follows from `caddy.hostName` exactly as the OIDC lane derives it.
#
# mkDefault semantics deliberately mirror oauth2.nix: an explicitly
# declared `settings.auth.passkey.*` value wins, so an operator pin (or
# an external front's hand-declared rp) stays authoritative. The raw
# lane stays untouched when the typed lane is off.
{
  config,
  lib,
  ...
}:
let
  cfg = config.services.webphone;

  # Behind the generated vhost the rp follows from caddy.hostName —
  # keyed on the TYPED lane only (reading the merged settings.auth.users
  # from inside the settings definition would recurse; the oauth2
  # precedent).
  deriveWebAuthnRP = cfg.caddy.enable && cfg.passkey.enable;

  # The identity map is shared by both passwordless modes and lives in
  # the app's ONE home for it (settings.auth.*, folded into the passkey
  # shape by the app). Read the merged value so the assertions reflect
  # what the binary will actually see.
  passkeyUsers = cfg.settings.auth.users or { };
  passkeyPasswordFiles =
    (cfg.settings.auth.extension_password_files or { })
    // (cfg.settings.auth.passkey.extension_password_files or { });
  passkeyMappedExtensions = lib.unique (
    lib.concatMap (user: user.extensions or [ ]) (lib.attrValues passkeyUsers)
  );
  passkeyMissingPasswordFiles = lib.filter (
    ext: (passkeyPasswordFiles.${ext} or "") == ""
  ) passkeyMappedExtensions;
in
{
  options.services.webphone.passkey = {
    enable = lib.mkEnableOption ''
      passkey (WebAuthn) sign-in as a STANDALONE mode: the login page
      renders the passkey ceremony (token-gated enrollment at /enroll;
      the extension + password form stays as the break-glass lifeline).
      This is the typed door for deployments that run passkeys WITHOUT
      the OIDC lane — behind the generated vhost the WebAuthn
      rp_id/rp_origins derive from `caddy.hostName`, exactly as the
      oauth2 lane derives them for SSO deployments.

      Requires the shared mapping (`settings.auth.users`) and one
      SIP-password file per mapped extension
      (`settings.auth.extension_password_files`) — both asserted, the
      same eval-time mirrors the oauth2 lane carries.
    '';
  };

  config = lib.mkIf cfg.enable {
    # Eval-time mirrors of the app's fail-closed passkey validation —
    # the same assertions the OIDC lane carries, keyed on this lane.
    assertions =
      lib.optionals cfg.passkey.enable [
        {
          assertion = passkeyUsers != { };
          message = ''
            services.webphone.passkey.enable requires settings.auth.users (the
            shared email→extension mapping) — without it every passkey login
            is unmapped and denied.
          '';
        }
        {
          assertion =
            (cfg.settings.auth.passkey.rp_id or "") != ""
            && (cfg.settings.auth.passkey.rp_origins or [ ]) != [ ];
          message = ''
            services.webphone.passkey.enable turns on the identity layer —
            settings.auth.passkey.rp_id and rp_origins must resolve. Enable
            caddy to derive them from caddy.hostName, or set both explicitly
            behind an external front.
          '';
        }
        {
          assertion = passkeyMissingPasswordFiles == [ ];
          message = ''
            services.webphone.passkey: every mapped extension needs a SIP
            directory password file in settings.auth.extension_password_files
            (missing: ${lib.concatStringsSep ", " passkeyMissingPasswordFiles}).
          '';
        }
      ];

    services.webphone.settings.auth = lib.mkIf cfg.passkey.enable (
      lib.mkIf deriveWebAuthnRP {
        # mkDefault: an explicitly declared settings.auth.passkey.*
        # value wins, so an operator pin stays authoritative (the
        # oauth2 precedent).
        passkey = {
          rp_id = lib.mkDefault cfg.caddy.hostName;
          rp_origins = lib.mkDefault [ "https://${cfg.caddy.hostName}" ];
        };
      }
    );
  };
}
