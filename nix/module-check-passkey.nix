# Passkey cases for the webphone-module check: the typed
# services.webphone.passkey lane (the passkey-only deployment's
# derived-rp door, plan T11), its fail-closed assertions, and the raw
# settings.auth.passkey lane. Mirrors the crm/textagent case files.
{
  lib,
  pkgs,
  base,
}:
let
  inherit (base) moduleSet;

  evaluated = extra: lib.evalModules (moduleSet extra);
  failedAssertions = extra: lib.filter (a: !a.assertion) (evaluated extra).config.assertions;

  # One bootable passkey-only config: the typed knob plus the shared
  # identity map and its password file (the harness default vhost
  # phone.example.org drives the rp derivation).
  validPasskey = {
    passkey.enable = true;
    settings = {
      auth.users."lars@example.org".extensions = [ "1000" ];
      auth.extension_password_files."1000" = "/run/secrets/telephony_ext_1000";
    };
  };

  # Each half shape must trip exactly its NAMED assertion. (The lane has
  # ONE knob — enable — so a knobs-without-enable shape does not exist;
  # a raw users map with the lane off is inert, legal config.)
  halfShapes = {
    missing-users = {
      fragment = "requires settings.auth.users";
      extra = {
        passkey.enable = true;
      };
    };
    missing-password-file = {
      fragment = "needs a SIP directory password file";
      extra = {
        passkey.enable = true;
        settings.auth.users."lars@example.org".extensions = [ "1000" ];
      };
    };
  };
in
[
  {
    # The happy path: the typed knob derives the WebAuthn rp from
    # caddy.hostName (the OIDC lane's convenience, now for passkey-only
    # deployments), no LoadCredential entry (a passkey carries no
    # secret), and the eval stays clean.
    name = "passkey-typed-render";
    path = pkgs.writeText "passkey-typed-render" (
      let
        ev = evaluated validPasskey;
        c = ev.config.services.webphone;
        unit = ev.config.systemd.services.webphone.serviceConfig;
        ok =
          c.settings.auth.passkey.rp_id == "phone.example.org"
          && c.settings.auth.passkey.rp_origins == [ "https://phone.example.org" ]
          && !(lib.any (entry: lib.hasInfix "passkey" entry) (unit.LoadCredential or [ ]))
          && failedAssertions validPasskey == [ ];
      in
      if ok then
        "typed passkey derives rp_id/rp_origins from caddy.hostName and needs no credential"
      else
        throw "webphone-module check: typed passkey did not render the expected settings.auth.passkey shape (rp_id: ${
          toString c.settings.auth.passkey.rp_id
        })"
    );
  }
  {
    # Every half shape trips its named assertion; the buildable shape
    # trips none. Mirrors the OIDC lane's fail-closed set.
    name = "passkey-fail-closed";
    path = pkgs.writeText "passkey-fail-closed" (
      let
        trips =
          variant:
          lib.any (message: lib.hasInfix variant.fragment message) (
            map (a: a.message) (failedAssertions variant.extra)
          );
        untriggered = lib.filter (name: !trips halfShapes.${name}) (lib.attrNames halfShapes);
        baseFailures = lib.length (failedAssertions { });
      in
      if untriggered == [ ] && baseFailures == 0 then
        "every half passkey shape trips its named assertion and the base eval stays clean"
      else
        throw "webphone-module check: passkey assertions mis-fire (untriggered: ${toString untriggered}, base failures: ${toString baseFailures})"
    );
  }
  {
    # The raw lane: a hand-declared settings.auth.passkey.* (the
    # external-front shape) passes through untouched when the typed
    # lane is off, and an operator pin BEATS the typed derivation
    # (mkDefault semantics — the oauth2 precedent).
    name = "passkey-raw-lane";
    path = pkgs.writeText "passkey-raw-lane" (
      let
        rawRP = {
          rp_id = "pbx.example.org";
          rp_origins = [ "https://pbx.example.org" ];
        };
        rawEv = evaluated {
          settings.auth.passkey = rawRP;
        };
        pinnedEv = evaluated (
          validPasskey
          // {
            settings = validPasskey.settings // {
              auth.passkey = rawRP;
            };
          }
        );
        ok =
          rawEv.config.services.webphone.settings.auth.passkey == rawRP
          && pinnedEv.config.services.webphone.settings.auth.passkey.rp_id == "pbx.example.org"
          && failedAssertions { settings.auth.passkey = rawRP; } == [ ];
      in
      if ok then
        "raw settings.auth.passkey passes through and an operator pin beats the typed derivation"
      else
        throw "webphone-module check: the raw settings.auth.passkey lane was altered by the typed passkey wiring"
    );
  }
]
