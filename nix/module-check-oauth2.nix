# OIDC/SSO cases for the webphone-module check: the typed
# services.webphone.oauth2 block, every fail-closed assertion, the raw
# settings.auth.oauth2 lane (the consuming stack's), the external-front
# lane, and the typed/raw conflict. Extracted from module-check.nix
# (monolith split).
{
  lib,
  pkgs,
  base,
}:
let
  inherit (base) moduleSet;

  evaluated = extra: lib.evalModules (moduleSet extra);
  failedAssertions = extra: lib.filter (a: !a.assertion) (evaluated extra).config.assertions;

  # One bootable SSO config: typed knobs plus the shared identity map the
  # app folds into passkey mode (users + one password file per extension).
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

  # Each half shape must trip exactly its NAMED assertion, so an operator
  # reads the fix in the eval error instead of discovering it at boot.
  halfShapes = {
    missing-issuer = {
      fragment = "oauth2.issuer";
      extra = validOAuth2 // {
        oauth2 = {
          enable = true;
          clientSecretFile = "/run/secrets/webphone-oidc";
        };
      };
    };
    missing-secret = {
      fragment = "clientSecretFile";
      extra = validOAuth2 // {
        oauth2 = {
          enable = true;
          issuer = "https://auth.example.org";
        };
      };
    };
    scopes-without-openid = {
      fragment = "openid";
      extra = validOAuth2 // {
        oauth2 = validOAuth2.oauth2 // {
          scopes = [ "email" ];
        };
      };
    };
    plain-http-issuer = {
      fragment = "bare https URL";
      extra = validOAuth2 // {
        oauth2 = validOAuth2.oauth2 // {
          issuer = "http://auth.example.org";
        };
      };
    };
    missing-users = {
      fragment = "settings.auth.users";
      extra = validOAuth2 // {
        settings = { };
      };
    };
    missing-password-file = {
      fragment = "extension_password_files";
      extra = validOAuth2 // {
        settings = {
          auth.users."lars@example.org".extensions = [ "1000" ];
        };
      };
    };
    knob-without-enable = {
      fragment = "enable = true";
      extra = {
        oauth2.issuer = "https://auth.example.org";
      };
    };
    no-derivable-callback = {
      fragment = "callback cannot be derived";
      extra = validOAuth2 // {
        caddy = {
          enable = false;
          hostName = "phone.example.org";
        };
      };
    };
  };
in
[
  {
    # The happy path: typed knobs render the app's auth.oauth2 block, the
    # callback and WebAuthn rp follow the vhost, and the secret is routed
    # through LoadCredential (the app reads the unit's copy, never the
    # operator's file in place).
    name = "oauth2-typed-render";
    path = pkgs.writeText "oauth2-typed-render" (
      let
        ev = evaluated validOAuth2;
        c = ev.config.services.webphone;
        unit = ev.config.systemd.services.webphone.serviceConfig;
        ok =
          c.settings.auth.oauth2.issuer == "https://auth.example.org"
          && c.settings.auth.oauth2.client_id == "webphone"
          &&
            c.settings.auth.oauth2.client_secret_file
            == "/run/credentials/webphone.service/webphone-oidc-client-secret"
          && c.settings.auth.oauth2.redirect_url == "https://phone.example.org/api/auth/oauth2/callback"
          &&
            c.settings.auth.oauth2.scopes == [
              "openid"
              "email"
              "profile"
            ]
          && c.settings.auth.oauth2.allowed_groups == [ ]
          && c.settings.auth.passkey.rp_id == "phone.example.org"
          && c.settings.auth.passkey.rp_origins == [ "https://phone.example.org" ]
          &&
            unit.LoadCredential == [
              "webphone-oidc-client-secret:/run/secrets/webphone-oidc"
            ]
          && failedAssertions validOAuth2 == [ ];
      in
      if ok then
        "typed oauth2 renders the app block, derives the callback + WebAuthn rp, and routes the secret through LoadCredential"
      else
        throw "webphone-module check: typed oauth2 did not render the expected settings/auth or LoadCredential shape"
    );
  }
  {
    # Every half shape trips its named assertion; the buildable shape
    # trips none. Pins the eval-time mirror of the app's boot validation.
    name = "oauth2-fail-closed";
    path = pkgs.writeText "oauth2-fail-closed" (
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
        "every half SSO shape trips its named assertion and the base eval stays clean"
      else
        throw "webphone-module check: oauth2 assertions mis-fire (untriggered: ${toString untriggered}, base failures: ${toString baseFailures})"
    );
  }
  {
    # The stack's lane: raw settings.auth.oauth2 without the typed block
    # must pass through untouched (no LoadCredential, no derived rp) —
    # the module must never fight an operator-declared OIDC client shape.
    name = "oauth2-raw-lane";
    path = pkgs.writeText "oauth2-raw-lane" (
      let
        raw = {
          issuer = "https://auth.example.org";
          client_id = "webphone";
          client_secret_file = "/run/secrets/webphone-oidc";
          redirect_url = "https://phone.example.org/api/auth/oauth2/callback";
          scopes = [ "openid" ];
        };
        ev = evaluated {
          settings.auth.oauth2 = raw;
        };
        c = ev.config.services.webphone;
        unit = ev.config.systemd.services.webphone.serviceConfig;
      in
      if
        c.settings.auth.oauth2 == raw
        && !(unit ? LoadCredential)
        && !(c.settings.auth ? passkey)
        && failedAssertions { settings.auth.oauth2 = raw; } == [ ]
      then
        "raw settings.auth.oauth2 passes through untouched when the typed lane is off"
      else
        throw "webphone-module check: the raw settings.auth.oauth2 lane was altered by the typed oauth2 wiring"
    );
  }
  {
    # Behind an external front there is no vhost to derive from: explicit
    # redirectUrl + rp values must satisfy every assertion and render
    # verbatim, and no rp default may appear.
    name = "oauth2-external-front";
    path = pkgs.writeText "oauth2-external-front" (
      let
        ev = evaluated (
          lib.recursiveUpdate validOAuth2 {
            caddy = {
              enable = false;
              hostName = "phone.example.org";
            };
            oauth2.redirectUrl = "https://alt.example.org/api/auth/oauth2/callback";
            settings.auth.passkey = {
              rp_id = "alt.example.org";
              rp_origins = [ "https://alt.example.org" ];
            };
          }
        );
        c = ev.config.services.webphone;
      in
      if
        c.settings.auth.oauth2.redirect_url == "https://alt.example.org/api/auth/oauth2/callback"
        && c.settings.auth.passkey.rp_id == "alt.example.org"
        && c.settings.auth.passkey.rp_origins == [ "https://alt.example.org" ]
        &&
          failedAssertions (
            lib.recursiveUpdate validOAuth2 {
              caddy = {
                enable = false;
                hostName = "phone.example.org";
              };
              oauth2.redirectUrl = "https://alt.example.org/api/auth/oauth2/callback";
              settings.auth.passkey = {
                rp_id = "alt.example.org";
                rp_origins = [ "https://alt.example.org" ];
              };
            }
          ) == [ ]
      then
        "external-front oauth2 needs explicit callback + rp values and renders verbatim"
      else
        throw "webphone-module check: external-front oauth2 wiring did not render the explicit callback/rp values"
    );
  }
  {
    # Declaring the OIDC client shape in BOTH lanes must fail the eval
    # (plain definitions conflict) — one home per fact, never a silent
    # winner between typed options and raw settings.
    name = "oauth2-typed-raw-conflict";
    path = pkgs.writeText "oauth2-typed-raw-conflict" (
      let
        attempt = builtins.tryEval (evaluated (
            lib.recursiveUpdate validOAuth2 {
              settings.auth.oauth2.issuer = "https://raw.example.org";
            }
          )).config.services.webphone.settings.auth.oauth2.issuer;
      in
      if !attempt.success then
        "typed + raw oauth2 for the same key is an eval conflict"
      else
        throw "webphone-module check: typed and raw oauth2 definitions merged silently (expected a conflict)"
    );
  }
  {
    # Secrets must never ride the Nix store: a /nix/store string (or a
    # path VALUE, which the freeform JSON type accepts and the renderer
    # copies) trips the leak guard, while a runtime-file string stays
    # clean.
    name = "oauth2-secret-store-guard";
    path = pkgs.writeText "oauth2-secret-store-guard" (
      let
        trips =
          extra: lib.any (a: !a.assertion && lib.hasInfix "/nix/store" a.message) (failedAssertions extra);
        storeString = trips (
          lib.recursiveUpdate validOAuth2 {
            settings.auth.extension_password_files."1000" = "/nix/store/xxx-secret";
          }
        );
        pathValue = trips (
          lib.recursiveUpdate validOAuth2 {
            settings.auth.extension_password_files."1000" = ./module-check-oauth2.nix;
          }
        );
        clean = failedAssertions validOAuth2 == [ ];
      in
      if storeString && pathValue && clean then
        "the store-path secret guard trips on both a /nix/store string and a path value, and stays quiet otherwise"
      else
        throw "webphone-module check: the store-path secret guard mis-fires (store string: ${toString storeString}, path value: ${toString pathValue}, clean: ${toString clean})"
    );
  }
]
