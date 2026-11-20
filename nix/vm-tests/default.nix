# The heavyweight checks, one file per test (the module-check case-file
# pattern): the KVM-gated VM tests plus the sandboxed backup/restore
# drill. flake.nix keeps the single import (this directory).
{
  imports = [
    ./backup-drill.nix
    ./webphone-backup.nix
    ./webphone-error-drill.nix
    ./webphone-oauth2.nix
  ];
}
