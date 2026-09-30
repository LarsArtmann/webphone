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
    inputs@{ flake-parts, ... }:
    let
      # The release version — the package version AND the /version
      # ldflags injection in one place. release.sh seds THIS binding
      # (grep "webphoneVersion = " flake.nix), so keep it in this file,
      # in this exact shape, in lockstep with the git tag at release.
      webphoneVersion = "2.8.0";
    in
    flake-parts.lib.mkFlake { inherit inputs; } {
      systems = [
        "x86_64-linux"
        "aarch64-linux"
      ];

      imports = [
        inputs.treefmt-nix.flakeModule
        # The release version rides into the packages module as a
        # module arg (visible at the module's top-level function).
        { _module.args.webphoneVersion = webphoneVersion; }
        ./nix/packages.nix
        ./nix/apps.nix
        ./nix/checks.nix
        ./nix/module-check.nix
        ./nix/vm-tests.nix
        ./nix/devshell.nix
        ./nix/treefmt.nix
      ];

      # The NixOS module ships from this repo so the binary and its
      # deployment shape stay in sync; the consuming telephony stack may
      # import it or keep its own reverse proxy (see package/nixos-module.nix).
      flake.nixosModules = {
        default = import ./package/nixos-module.nix;
        # Alias for consumers that prefer an explicit name.
        webphone = import ./package/nixos-module.nix;
      };
    };
}
