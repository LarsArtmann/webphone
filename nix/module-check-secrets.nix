# Freeform literal-secret cases for the webphone-module check: the
# inline-secret guard (a secret-shaped key with an inline string value
# renders into the world-readable config JSON). Covers both arms — the
# planted literal fails eval naming the key, and the sanctioned lanes
# (file-form keys, runtime-path values, non-secret passthrough) keep
# evaluating clean.
{
  lib,
  pkgs,
  base,
}:
let
  inherit (base) moduleSet;

  evaluated = extra: lib.evalModules (moduleSet extra);
  failedAssertions = extra: lib.filter (a: !a.assertion) (evaluated extra).config.assertions;

  # The planted literal: the deployment mistake the guard owns. Its
  # message must name the offending attrpath so the operator sees WHICH
  # key to move.
  planted = failedAssertions {
    settings.gateway.webhook_secret = "hunter2-super-secret";
  };

  # The sanctioned lanes must stay clean: a *_file key pointing at a
  # runtime file, and a value that already names a runtime location.
  fileLane = failedAssertions {
    settings.sip.credential_secret_file = "/run/secrets/webphone-sip-credential";
  };
  pathLane = failedAssertions {
    settings.asr.token = "/run/secrets/webphone-asr-token";
  };

  # A settings key that merely CONTAINS a secret-ish word without being
  # one (here: the freeform passthrough class) must keep evaluating.
  passthrough = failedAssertions {
    settings.gateway.mode = "webhook";
    settings.csrf.trusted_origins = [ "https://phone.example.org" ];
  };
in
[
  {
    name = "literal-secret-guard";
    path = pkgs.writeText "literal-secret-guard" (
      if
        lib.length planted == 1
        && lib.hasInfix "gateway.webhook_secret" (lib.head planted).message
        && lib.hasInfix "inline secret values" (lib.head planted).message
        && fileLane == [ ]
        && pathLane == [ ]
        && passthrough == [ ]
      then
        "planted literal secret refused by name; file/path/passthrough lanes stay clean"
      else
        throw
          "webphone-module check: inline-secret guard mis-fires (planted=${toString (lib.length planted)}, file=${toString (lib.length fileLane)}, path=${toString (lib.length pathLane)}, passthrough=${toString (lib.length passthrough)})"
    );
  }
]
