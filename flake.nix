{
  description = "Webphone: self-hosted unified-communications web app (calls, SMS/MMS, fax, voicemail) in one Go binary";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";

    flake-parts = {
      url = "github:hercules-ci/flake-parts";
      inputs.nixpkgs-lib.follows = "nixpkgs";
    };

    treefmt-nix = {
      url = "github:numtide/treefmt-nix";
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };

  outputs =
    inputs@{ self, flake-parts, ... }:
    flake-parts.lib.mkFlake { inherit inputs; } {
      systems = [
        "x86_64-linux"
        "aarch64-linux"
      ];

      imports = [ inputs.treefmt-nix.flakeModule ];

      # The NixOS module ships from this repo so the binary and its
      # deployment shape stay in sync; the consuming telephony stack may
      # import it or keep its own reverse proxy (see package/nixos-module.nix).
      flake.nixosModules = {
        default = import ./package/nixos-module.nix;
        # Alias for consumers that prefer an explicit name.
        webphone = import ./package/nixos-module.nix;
      };

      perSystem =
        {
          config,
          lib,
          pkgs,
          self',
          ...
        }:
        {
          packages = {
            default = self'.packages.webphone;

            # One Go binary: templ shell + embedded island assets + SQLite.
            # GOEXPERIMENT=jsonv2 is required by templ-components (encoding/
            # json/v2) until Go 1.27 ships it stable. webphoneVersion is the
            # single source: the package version AND the /version ldflags
            # injection — keep it in lockstep with the git tag at release.
            webphone =
              let
                webphoneVersion = "2.0.0";
              in
              pkgs.buildGoModule {
                pname = "webphone";
                version = webphoneVersion;

                src = lib.fileset.toSource {
                  root = ./.;
                  # Exactly the build inputs: sources under cmd/ and
                  # internal/ (island assets + committed *_templ.go live
                  # there) plus the module definition files. Everything
                  # else (docs, scripts, package/) never reaches Go.
                  fileset = lib.fileset.unions [
                    ./cmd
                    ./go.mod
                    ./go.sum
                    ./internal
                  ];
                };

                vendorHash = "sha256-WlFJ83w9VX+alrCHZRC7nVPFU/l9yoUB1JGNrTGU8R0=";

                proxyVendor = true;

                env.GOEXPERIMENT = "jsonv2";

                subPackages = [ "cmd/webphone" ];

                ldflags = [
                  "-s"
                  "-w"
                  # /version reports the released version (v-prefixed, like
                  # the git tag) instead of Go's "(devel)".
                  "-X github.com/larsartmann/webphone/internal/server.buildVersion=v${webphoneVersion}"
                ];

                doCheck = true;

                meta = {
                  description = "Self-hosted unified-communications web app: calls, SMS/MMS threads, fax, voicemail";
                  homepage = "https://github.com/LarsArtmann/webphone";
                  license = lib.licenses.mit;
                  mainProgram = "webphone";
                  platforms = lib.platforms.linux;
                  maintainers = [
                    {
                      name = "Lars Artmann";
                      github = "LarsArtmann";
                    }
                  ];
                };
              };
          };

          checks = {
            webphone = self'.packages.webphone;
            format = config.treefmt.build.check self;

            # Evaluate the NixOS module with a minimal config and build
            # the artifacts it would generate — catches option/syntax
            # breakage without a full NixOS evaluation.
            webphone-module =
              let
                evaluated = lib.evalModules {
                  modules = [
                    { _module.args.pkgs = pkgs; }
                    # Minimal stand-ins for the NixOS nginx/systemd/users
                    # modules: the webphone module writes services.nginx,
                    # systemd.services, and users.{users,groups} config.
                    {
                      options = {
                        services.nginx = {
                          enable = lib.mkOption {
                            type = lib.types.bool;
                            default = false;
                          };
                          recommendedProxySettings = lib.mkOption {
                            type = lib.types.bool;
                            default = false;
                          };
                          virtualHosts = lib.mkOption { type = lib.types.attrsOf lib.types.anything; };
                        };
                        systemd.services = lib.mkOption { type = lib.types.attrsOf lib.types.anything; };
                        users.users = lib.mkOption { type = lib.types.attrsOf lib.types.anything; };
                        users.groups = lib.mkOption { type = lib.types.attrsOf lib.types.anything; };
                        # NixOS's modules.nix normally provides this.
                        assertions = lib.mkOption {
                          type = lib.types.listOf lib.types.anything;
                          default = [ ];
                        };
                      };
                    }
                    (import ./package/nixos-module.nix)
                    {
                      services.webphone = {
                        enable = true;
                        package = self'.packages.webphone;
                        nginx.enable = true;
                        nginx.hostName = "phone.example.org";
                        settings.sip_domain = "pbx.example.org";
                      };
                    }
                  ];
                };
                cfg = evaluated.config.services.webphone;
                vhost = evaluated.config.services.nginx.virtualHosts."phone.example.org";
                locationNames = lib.attrNames vhost.locations;
                # Every location the DOM/SSE contract rides on must be
                # proxied by the module's own vhost.
                missingLocations = lib.filter (loc: !lib.elem loc locationNames) [
                  "/"
                  cfg.settings.websocket_path
                  "/events"
                ];
                unitPresent = evaluated.config.systemd.services ? "webphone";
              in
              pkgs.linkFarm "webphone-module-check" [
                {
                  name = "webphone-config.json";
                  path = (pkgs.formats.json { }).generate "webphone-config.json" cfg.settings;
                }
                {
                  name = "listen-port";
                  path = pkgs.writeText "listen-port" (lib.last (lib.splitString ":" cfg.settings.addr));
                }
                {
                  name = "vhost-locations";
                  path = pkgs.writeText "vhost-locations" (
                    if missingLocations == [ ] then
                      lib.concatStringsSep "\n" locationNames
                    else
                      throw "webphone-module check: vhost locations missing: ${toString missingLocations}"
                  );
                }
                {
                  name = "systemd-unit";
                  path = pkgs.writeText "systemd-unit" (
                    if unitPresent then
                      "systemd.services.webphone present"
                    else
                      throw "webphone-module check: systemd.services.webphone missing"
                  );
                }
              ];

            statix =
              pkgs.runCommand "statix-check"
                {
                  nativeBuildInputs = [ pkgs.statix ];
                  meta.description = "statix lint over the Nix files";
                }
                ''
                  cd ${self}
                  statix check -o errfmt . 2>&1 | tee $out
                '';
            deadnix =
              pkgs.runCommand "deadnix-check"
                {
                  nativeBuildInputs = [ pkgs.deadnix ];
                  meta.description = "deadnix scan over the Nix files";
                }
                ''
                  cd ${self}
                  deadnix --fail --no-lambda-pattern-names . 2>&1 | tee $out
                '';

            # no-undef over the SIP island modules: a call to an undefined
            # identifier used to surface only as a silent browser
            # ReferenceError (the accept/reject bug class) — this gate is
            # the local tripwire. The island is excluded from BuildFlow,
            # so the gate lives here with its config beside the sources.
            island-lint =
              pkgs.runCommand "island-lint-check"
                {
                  nativeBuildInputs = [ pkgs.oxlint ];
                  meta.description = "oxlint no-undef over the SIP island modules";
                }
                ''
                  cd ${self}
                  if oxlint -c internal/web/assets/island/oxlint.json internal/web/assets/island/app/ internal/web/assets/shell.js; then
                    echo "no-undef clean over:" >$out
                    ls internal/web/assets/island/app/ >>$out
                  else
                    echo "island no-undef gate FAILED" >&2
                    exit 1
                  fi
                '';
          };

          devShells.default = pkgs.mkShellNoCC {
            packages = [
              config.treefmt.build.wrapper
              pkgs.go
              pkgs.templ
              pkgs.golangci-lint
              pkgs.esbuild
              pkgs.jq
              pkgs.nil
              pkgs.oxlint
              pkgs.vulnix
              pkgs.go-licenses
            ];
            env = {
              # templ-components needs encoding/json/v2 until Go 1.27 ships
              # it stable; local keeps the shell's go instead of downloading
              # a toolchain behind the user's back.
              GOEXPERIMENT = "jsonv2";
              GOTOOLCHAIN = "local";
            };
          };

          # Runtime-closure vulnerability scan: `nix run .#vulnix` (repo
          # root, network required for the advisory DB). Scans ONLY the
          # runtime closure — the build closure (bootstrap toolchains,
          # gcc, zlib) never deploys and drowns the signal. Networked by
          # nature, so it is an app, not a flake check.
          apps.vulnix =
            let
              script = pkgs.writeShellApplication {
                name = "webphone-vulnix";
                runtimeInputs = [
                  pkgs.nix
                  pkgs.vulnix
                ];
                text = ''
                  out=$(nix build --no-link --print-out-paths .#webphone)
                  echo "webphone-vulnix: scanning the runtime closure of $out ($(nix-store -qR "$out" | wc -l) derivations)"
                  # --closure: runtime dependencies ONLY. Without it vulnix
                  # closes over BUILD inputs (bootstrap toolchains, gcc,
                  # binutils) and drowns the signal. Caveat: NVD cannot see
                  # distro patch suffixes, so range-matched advisories that
                  # nixpkgs has already fixed may appear (e.g. glibc
                  # CVE-2026-5450 printed against 2.42-84; the fix shipped
                  # in 2.42-67). Verify each finding against the nixpkgs
                  # patch level before acting on it.
                  if vulnix --closure "$out"; then
                    echo "webphone-vulnix: no known advisories in the runtime closure"
                  else
                    echo "webphone-vulnix: findings above — triage each against the nixpkgs patched version" >&2
                    exit 1
                  fi
                '';
              };
            in
            {
              type = "app";
              program = lib.getExe script;
            };

          treefmt = {
            projectRootFile = "flake.nix";
            programs = {
              nixfmt.enable = true;
              gofmt.enable = true;
              prettier = {
                enable = true;
                # No *.html here: the only HTML in the repo is historical
                # docs/status + docs/planning snapshots (point-in-time,
                # never reformatted — and prettier cannot parse one).
                includes = [
                  "*.css"
                  "internal/web/assets/shell.js"
                  "internal/web/assets/island/**/*.js"
                ];
              };
            };
          };
        };
    };
}
