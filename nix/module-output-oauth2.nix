# The oauth2-enabled full-text rendering of the module's generated
# output: the same three sections as module-output.nix (Caddy vhost
# body, retention=7 backup script, settings JSON), evaluated WITH the
# typed SSO lane on. This is the SINGLE source both the
# `oauth2-enabled-output-golden` check and the golden fixture come from
# (the check diffs the rendered text against
# nix/module-output-oauth2.golden). Regenerate the fixture after an
# intentional change with the module-output.nix recipe, importing THIS
# file instead.
#
# The pinned JSON carries the derivations the lane makes from
# `caddy.hostName` (the harness default phone.example.org): the
# registered callback https://phone.example.org/api/auth/oauth2/callback
# and the passkey rp_id/rp_origins — config drift in those derivations
# fails eval here, not login.
{ lib, base }:
let
  # One bootable SSO config: typed knobs plus the shared identity map
  # the app folds into passkey mode (users + one password file per
  # extension). Mirrors validOAuth2 in module-check-oauth2.nix.
  validOAuth2 = {
    oauth2 = {
      enable = true;
      issuer = "https://auth.example.org";
      clientSecretFile = "/run/secrets/webphone-oidc";
    };
    settings = {
      auth.users."lars@example.org".extensions = [ "1000" ];
      auth.extension_password_files."1000" = "/run/secrets/telephony_ext_1000";
    };
  };

  evaluated = lib.evalModules (base.moduleSet validOAuth2);
  backupScript =
    (lib.evalModules (
      base.moduleSet (
        validOAuth2
        // {
          backup.enable = true;
          backup.retentionDays = 7;
        }
      )
    )).config.systemd.services.webphone-backup.script;
  vhost = evaluated.config.services.caddy.virtualHosts."phone.example.org";
in
''
  == Caddy vhost extraConfig ==
  ${vhost.extraConfig}
  == backup script (retentionDays = 7) ==
  ${backupScript}
  == settings JSON ==
  ${builtins.toJSON evaluated.config.services.webphone.settings}
''
