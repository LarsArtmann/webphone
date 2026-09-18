{
  description = "Webphone: standalone SIP.js WebRTC softphone UI, packaged as a static site";

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
            webphone = pkgs.callPackage ./package { };
          };

          checks = {
            webphone = self'.packages.webphone;
            format = config.treefmt.build.check self;
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
              esbuild
              jq
              curl
              nil
            ];
          };

          treefmt = {
            projectRootFile = "flake.nix";
            programs = {
              nixfmt.enable = true;
              prettier = {
                enable = true;
                includes = [
                  "src/*.html"
                  "src/*.css"
                  "src/app/*.js"
                ];
              };
            };
          };
        };
    };
}
