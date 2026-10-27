# CRM cases for the webphone-module check: the typed
# services.webphone.crm block, every fail-closed assertion, the raw
# settings.crm lane, and the typed/raw conflict. Mirrors the oauth2
# case file (extracted monolith-split pattern).
{
  lib,
  pkgs,
  base,
}:
let
  inherit (base) moduleSet;

  evaluated = extra: lib.evalModules (moduleSet extra);
  failedAssertions = extra: lib.filter (a: !a.assertion) (evaluated extra).config.assertions;

  # One complete CRM config: the typed knobs only (the token rides the
  # credential; there is no shared-map dependency like oauth2's).
  validCRM = {
    crm = {
      enable = true;
      url = "https://crm.example.org";
      tokenFile = "/run/secrets/webphone-crm-token";
    };
  };

  # Each half shape must trip exactly its NAMED assertion.
  halfShapes = {
    missing-url = {
      fragment = "crm.url must be";
      extra = validCRM // {
        crm = {
          enable = true;
          tokenFile = "/run/secrets/webphone-crm-token";
        };
      };
    };
    plain-http-url = {
      fragment = "absolute https URL";
      extra = validCRM // {
        crm = validCRM.crm // {
          url = "http://crm.example.org";
        };
      };
    };
    missing-token-file = {
      fragment = "tokenFile is required";
      extra = validCRM // {
        crm = {
          enable = true;
          url = "https://crm.example.org";
        };
      };
    };
    knob-without-enable = {
      fragment = "set enable = true";
      extra = {
        crm.url = "https://crm.example.org";
      };
    };
    leaked-token-to-store = {
      fragment = "world-readable Nix store";
      extra = validCRM // {
        crm = validCRM.crm // {
          tokenFile = "/nix/store/fake-crm-token";
        };
      };
    };
  };
in
[
  {
    # The happy path: the typed knobs render the app's crm block with
    # the credential path, and the token is routed through
    # LoadCredential (the app reads the unit's copy, never the
    # operator's file in place).
    name = "crm-typed-render";
    path = pkgs.writeText "crm-typed-render" (
      let
        ev = evaluated validCRM;
        c = ev.config.services.webphone;
        unit = ev.config.systemd.services.webphone.serviceConfig;
        ok =
          c.settings.crm.url == "https://crm.example.org"
          && c.settings.crm.token_file == "/run/credentials/webphone.service/webphone-crm-token"
          && lib.elem "webphone-crm-token:/run/secrets/webphone-crm-token" unit.LoadCredential
          && failedAssertions validCRM == [ ];
      in
      if ok then
        "typed crm renders the app block with the credential path and routes the token through LoadCredential"
      else
        throw "webphone-module check: typed crm did not render the expected settings.crm or LoadCredential shape"
    );
  }
  {
    # Every half shape trips its named assertion; the buildable shape
    # trips none. Pins the eval-time mirror of the app's boot validation.
    name = "crm-fail-closed";
    path = pkgs.writeText "crm-fail-closed" (
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
        "every half CRM shape trips its named assertion and the base eval stays clean"
      else
        throw "webphone-module check: crm assertions mis-fire (untriggered: ${toString untriggered}, base failures: ${toString baseFailures})"
    );
  }
  {
    # The raw lane: settings.crm without the typed block must pass
    # through untouched (no LoadCredential entry) — the module must
    # never fight an operator-declared CRM shape.
    name = "crm-raw-lane";
    path = pkgs.writeText "crm-raw-lane" (
      let
        raw = {
          url = "https://crm.example.org";
          token = "sekrit";
        };
        ev = evaluated {
          settings.crm = raw;
        };
        c = ev.config.services.webphone;
        unit = ev.config.systemd.services.webphone.serviceConfig;
      in
      if
        c.settings.crm == raw && !(unit ? LoadCredential) && failedAssertions { settings.crm = raw; } == [ ]
      then
        "raw settings.crm passes through untouched when the typed lane is off"
      else
        throw "webphone-module check: the raw settings.crm lane was altered by the typed crm wiring"
    );
  }
  {
    # Typed and raw for the same key is an eval CONFLICT (one home per
    # fact) — the oauth2 precedent. A silent winner would let the
    # operator believe the raw token is live while the typed credential
    # file is read (or the reverse).
    name = "crm-typed-raw-conflict";
    path = pkgs.writeText "crm-typed-raw-conflict" (
      let
        attempt =
          builtins.tryEval
            (evaluated (
              validCRM
              // {
                settings.crm.url = "https://raw.example.org";
              }
            )).config.services.webphone.settings.crm.url;
      in
      if !attempt.success then
        "typed + raw crm for the same key is an eval conflict"
      else
        throw "webphone-module check: typed and raw crm definitions merged silently (expected a conflict)"
    );
  }
]
