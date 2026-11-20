# The sandboxed backup/restore drill: no KVM gate (loopback-only, runs in
# the sandbox); the script boots a live webphone, snapshots, restores, and
# reads a byte-identical attachment back out of the restored store.
{
  self,
  ...
}:
{
  perSystem =
    {
      pkgs,
      self',
      ...
    }:
    {
      checks = {
        # data dir, reboot in a fresh location, pull a byte-identical
        # attachment back out of the restored store. Loopback-only, so
        # it runs in the sandbox (no KVM gate).
        webphone-backup-drill =
          pkgs.runCommand "webphone-backup-drill-check"
            {
              nativeBuildInputs = [ pkgs.python3 ];
              meta.description = "end-to-end backup/restore drill over a live webphone instance";
              meta.timeout = 300;
            }
            ''
              export WP_BIN=${self'.packages.webphone}/bin/webphone
              python3 ${self}/scripts/webphone-backup-drill.py
              echo "drill passed" > $out
            '';
      };
    };
}
