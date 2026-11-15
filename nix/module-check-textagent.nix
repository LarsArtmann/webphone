# Textagent cases for the webphone-module check: the typed
# services.webphone.textagent block, every fail-closed assertion, the
# raw settings.textagent lane, and the typed/raw conflict. Mirrors the
# crm case file (the typed-front pattern).
{
  lib,
  pkgs,
  base,
}:
let
  inherit (base) moduleSet;

  evaluated = extra: lib.evalModules (moduleSet extra);
  failedAssertions = extra: lib.filter (a: !a.assertion) (evaluated extra).config.assertions;

  # One complete textagent config: the typed knobs only (the secret
  # rides the credential; loopback http is the normal review-API shape).
  validTextagent = {
    textagent = {
      enable = true;
      url = "http://127.0.0.1:8071";
      secretFile = "/run/secrets/webphone-textagent-secret";
    };
  };

  # Each half shape must trip exactly its NAMED assertion.
  halfShapes = {
    missing-url = {
      fragment = "must be an absolute http(s)";
      extra = validTextagent // {
        textagent = {
          enable = true;
          secretFile = "/run/secrets/webphone-textagent-secret";
        };
      };
    };
    relative-url = {
      fragment = "must be an absolute http(s)";
      extra = validTextagent // {
        textagent = validTextagent.textagent // {
          url = "review-api:8071";
        };
      };
    };
    missing-secret-file = {
      fragment = "secretFile is required";
      extra = validTextagent // {
        textagent = {
          enable = true;
          url = "http://127.0.0.1:8071";
        };
      };
    };
    knob-without-enable = {
      fragment = "set enable = true";
      extra = {
        textagent.url = "http://127.0.0.1:8071";
      };
    };
    leaked-secret-to-store = {
      fragment = "world-readable Nix store";
      extra = validTextagent // {
        textagent = validTextagent.textagent // {
          secretFile = "/nix/store/fake-textagent-secret";
        };
      };
    };
  };
in
[
  {
    # The happy path: the typed knobs render the app's textagent block
    # with the credential path, and the secret routes through
    # LoadCredential. Zero failed assertions doubles as the
    # module-check-secrets pass for this lane: the rendered
    # textagent.secret_file is the sanctioned file-form value, so the
    # inline-secret guard stays silent.
    name = "textagent-typed-render";
    path = pkgs.writeText "textagent-typed-render" (
      let
        ev = evaluated validTextagent;
        c = ev.config.services.webphone;
        unit = ev.config.systemd.services.webphone.serviceConfig;
        ok =
          c.settings.textagent.url == "http://127.0.0.1:8071"
          && c.settings.textagent.secret_file == "/run/credentials/webphone.service/webphone-textagent-secret"
          && lib.elem "webphone-textagent-secret:/run/secrets/webphone-textagent-secret" (
            unit.LoadCredential or [ ]
          )
          && failedAssertions validTextagent == [ ];
      in
      if ok then
        "typed textagent renders the app block with the credential path and routes the secret through LoadCredential"
      else
        throw "webphone-module check: typed textagent did not render the expected settings.textagent or LoadCredential shape"
    );
  }
  {
    # Every half shape trips its named assertion; the buildable shape
    # trips none. Pins the eval-time mirror of the app's boot validation.
    name = "textagent-fail-closed";
    path = pkgs.writeText "textagent-fail-closed" (
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
        "every half textagent shape trips its named assertion and the base eval stays clean"
      else
        throw "webphone-module check: textagent assertions mis-fire (untriggered: ${toString untriggered}, base failures: ${toString baseFailures})"
    );
  }
  {
    # The raw lane: settings.textagent without the typed block must pass
    # through untouched (no LoadCredential entry) — the module must
    # never fight an operator-declared seam shape. A RAW secret_file
    # stays sanctioned (file-form), so the eval stays clean.
    name = "textagent-raw-lane";
    path = pkgs.writeText "textagent-raw-lane" (
      let
        raw = {
          url = "http://127.0.0.1:8071";
          secret_file = "/run/secrets/webphone-textagent-secret";
        };
        ev = evaluated {
          settings.textagent = raw;
        };
        c = ev.config.services.webphone;
        unit = ev.config.systemd.services.webphone.serviceConfig;
        ok =
          c.settings.textagent == raw
          && !(lib.any (entry: lib.hasInfix "textagent" entry) (unit.LoadCredential or [ ]))
          && failedAssertions { settings.textagent = raw; } == [ ];
      in
      if ok then
        "raw settings.textagent passes through untouched when the typed lane is off"
      else
        throw "webphone-module check: the raw settings.textagent lane was altered by the typed textagent wiring"
    );
  }
  {
    # Typed and raw for the same key is an eval CONFLICT (one home per
    # fact) — the crm/oauth2 precedent. A silent winner would let the
    # operator believe the raw secret is live while the credential
    # copy is read (or the reverse).
    name = "textagent-typed-raw-conflict";
    path = pkgs.writeText "textagent-typed-raw-conflict" (
      let
        attempt =
          builtins.tryEval
            (evaluated (
              validTextagent
              // {
                settings.textagent.url = "http://raw.example.org:8071";
              }
            )).config.services.webphone.settings.textagent.url;
      in
      if !attempt.success then
        "typed + raw settings.textagent.url is an eval conflict (one home per fact)"
      else
        throw "webphone-module check: the raw settings.textagent.url silently won over the typed lane (got ${attempt.value})"
    );
  }
]
