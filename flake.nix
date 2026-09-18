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
      flake.nixosModules.default = import ./package/nixos-module.nix;

      perSystem =
        {
          config,
          pkgs,
          self',
          ...
        }:
        {
          packages = {
            default = self'.packages.webphone;

            # One Go binary: templ shell + embedded island assets + SQLite.
            # GOEXPERIMENT=jsonv2 is required by templ-components (encoding/
            # json/v2) until Go 1.27 ships it stable.
            webphone = pkgs.buildGoModule {
              pname = "webphone";
              version = "2.0.0";

              src = pkgs.lib.cleanSource self;

              vendorHash = "sha256-WlFJ83w9VX+alrCHZRC7nVPFU/l9yoUB1JGNrTGU8R0=";

              proxyVendor = true;

              env.GOEXPERIMENT = "jsonv2";

              subPackages = [ "cmd/webphone" ];

              ldflags = [
                "-s"
                "-w"
              ];

              doCheck = true;

              meta = {
                description = "Self-hosted unified-communications web app: calls, SMS/MMS threads, fax, voicemail";
                homepage = "https://github.com/LarsArtmann/webphone";
                license = pkgs.lib.licenses.mit;
                mainProgram = "webphone";
                platforms = pkgs.lib.platforms.linux;
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
                evaluated = pkgs.lib.evalModules {
                  modules = [
                    { _module.args.pkgs = pkgs; }
                    # Minimal stand-ins for the NixOS nginx/systemd/users
                    # modules: the webphone module writes services.nginx,
                    # systemd.services, and users.{users,groups} config.
                    {
                      options = {
                        services.nginx = {
                          enable = pkgs.lib.mkOption {
                            type = pkgs.lib.types.bool;
                            default = false;
                          };
                          recommendedProxySettings = pkgs.lib.mkOption {
                            type = pkgs.lib.types.bool;
                            default = false;
                          };
                          virtualHosts = pkgs.lib.mkOption { type = pkgs.lib.types.attrsOf pkgs.lib.types.anything; };
                        };
                        systemd.services = pkgs.lib.mkOption { type = pkgs.lib.types.attrsOf pkgs.lib.types.anything; };
                        users.users = pkgs.lib.mkOption { type = pkgs.lib.types.attrsOf pkgs.lib.types.anything; };
                        users.groups = pkgs.lib.mkOption { type = pkgs.lib.types.attrsOf pkgs.lib.types.anything; };
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
              in
              pkgs.linkFarm "webphone-module-check" [
                {
                  name = "webphone-config.json";
                  path = (pkgs.formats.json { }).generate "webphone-config.json" cfg.settings;
                }
                {
                  name = "listen-port";
                  path = pkgs.writeText "listen-port" (pkgs.lib.last (pkgs.lib.splitString ":" cfg.settings.addr));
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
          };

          devShells.default = pkgs.mkShellNoCC {
            packages = with pkgs; [
              config.treefmt.build.wrapper
              go
              templ
              golangci-lint
              esbuild
              jq
              nil
            ];
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
