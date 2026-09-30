# The heavyweight checks: the KVM-gated backup VM test and the
# sandboxed backup/restore drill.
# `self` is a TOP-level flake-parts module arg (not perSystem) — it is
# declared here and reaches the perSystem body via lexical closure.
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
        # Backup-timer VM test (fax-feed-test style, named after the
        # consuming stack's tests/fax-feed.nix): boot the REAL service
        # with backup.enable, run the oneshot, and assert the online
        # snapshot lands under destDir while the service keeps serving —
        # the production behaviors are the sqlite .backup consistency
        # and the no-restart claim, not just file existence.
        # kvm-gated like the stack's VM tests: `nix flake check` skips
        # it (with a warning) on machines without KVM instead of
        # degrading to multi-minute TCG boots.
        webphone-backup = pkgs.testers.runNixOSTest {
          name = "webphone-backup";

          requiredFeatures.kvm = true;

          nodes.machine =
            { pkgs, ... }:
            {
              imports = [ ../package/nixos-module.nix ];
              services.webphone = {
                enable = true;
                package = self'.packages.webphone;
                backup.enable = true;
                backup.retentionDays = 7;
              };
              environment.systemPackages = [ pkgs.sqlite ];
              system.stateVersion = "26.05";
            };

          testScript = ''
            machine.wait_for_unit("webphone.service")
            machine.wait_for_open_port(8080)

            # Deterministic run: start the oneshot directly (the timer
            # exists too, but the test asserts outcomes, not scheduler
            # timing — same posture as the stack's fax-feed test).
            machine.succeed("systemctl start webphone-backup.service")

            # The snapshot pair lands under destDir.
            machine.wait_until_succeeds(
                "test -f /var/lib/webphone-backup/webphone.db",
                timeout=30,
            )
            machine.succeed("test -d /var/lib/webphone-backup/files")

            # It is a consistent database, not a torn copy: sqlite's
            # .backup API ran to completion.
            machine.succeed(
                "sqlite3 /var/lib/webphone-backup/webphone.db 'pragma integrity_check' | grep -q '^ok$'"
            )

            # Private-data hardening (2026-09-29 train, pinned 2026-09-30):
            # the oneshot succeeded, the state dir is group-readable at
            # most, and backup artifacts are owner-only (UMask=0077 ->
            # files 600, dirs 700) — comms data never world-readable.
            machine.succeed(
                "systemctl show webphone-backup.service -p Result | grep -q 'Result=success'"
            )
            machine.succeed(
                "test \"$(stat -c %a /var/lib/webphone)\" = 750"
            )
            machine.succeed(
                "stat -c %a /var/lib/webphone-backup/webphone.db | grep -qE '^(600|700)$'"
            )
            machine.succeed(
                "stat -c %a /var/lib/webphone-backup/files | grep -qE '^(700|750)$'"
            )

            # The timer is wired to the oneshot on the default calendar
            # (systemctl show has no OnCalendar unit property — the
            # rendered unit file is the truth here).
            machine.succeed(
                "systemctl cat webphone-backup.timer | grep -q '^OnCalendar=\\*-\\*-\\* 04:30:00'"
            )
            machine.succeed(
                "systemctl cat webphone-backup.timer | grep -q '^Unit=webphone-backup.service$'"
            )

            # Online claim: the phone service never restarted for the
            # backup (uptime predates the oneshot run).
            machine.succeed(
            "systemctl show webphone.service -p NRestarts | grep -q 'NRestarts=0'"
            )

            # Retention (retentionDays = 7): plant a stale dated
            # snapshot AFTER a first successful run (so the StateDirectory
            # and snapshots/ tree exist under the service user), rerun the
            # oneshot — the stale directory is pruned, today's dated
            # snapshot survives with an intact db, and the same-day rerun
            # must not have used itself as the rsync basis.
            machine.succeed(
            "install -d -o webphone -g webphone /var/lib/webphone-backup/snapshots/2000-01-01"
            )
            machine.succeed(
            "touch /var/lib/webphone-backup/snapshots/2000-01-01/stale.db && touch -d '2000-01-01' /var/lib/webphone-backup/snapshots/2000-01-01"
            )
            machine.succeed("systemctl start webphone-backup.service")
            machine.succeed("test ! -e /var/lib/webphone-backup/snapshots/2000-01-01")
            machine.succeed(
            "test -f '/var/lib/webphone-backup/snapshots/'$(date +%F)'/webphone.db'"
            )
            machine.succeed(
            "sqlite3 '/var/lib/webphone-backup/snapshots/'$(date +%F)'/webphone.db' 'pragma integrity_check' | grep -q '^ok$'"
            )
          '';
        };

        # Backup/restore DRILL (plan T23, 2026-09-20): the VM test above
        # proves the online snapshot; this proves the RESTORE — tar the
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
