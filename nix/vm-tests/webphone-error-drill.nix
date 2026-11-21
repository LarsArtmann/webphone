# The KVM-gated error-contract drill VM test: the designed boot failure and the
_: {
  perSystem =
    {
      pkgs,
      self',
      ...
    }:
    {
      checks = {
        # Error-contract DRILL (SUPERB error-agents M14-M15, 2026-10-26):
        # the failure SHAPE the product owes, asserted in the REAL
        # module/unit/journald context the smoke cannot see. Scenario A:
        # an unwritable data dir under a regular file must produce the
        # designed boot failure — five-part operator block on the
        # journal, exit taxonomy 1, and an honest systemd restart loop.
        # Plus the two runtime shape asserts: /healthz degrades to 503
        # NAMING the failing leg when the blob store becomes unwritable
        # (and recovers when unwedged), and /hooks/* fail CLOSED (503)
        # with no gateway secret configured. KVM-gated like the backup
        # VM test — consumer CI (the fast-checks list) never runs it.
        # The gateway-lane and surface-shape scenarios of the plan stay
        # smoke-owned (scripts/webphone-smoke.py runs them per train on
        # the same binary; one home per fact).
        webphone-error-drill = pkgs.testers.runNixOSTest {
          name = "webphone-error-drill";

          requiredFeatures.kvm = true;

          nodes.machine =
            { pkgs, ... }:
            {
              imports = [ ../../package/nixos-module.nix ];
              services.webphone = {
                enable = true;
                package = self'.packages.webphone;
              };
              environment.systemPackages = [ pkgs.curl ];
              system.stateVersion = "26.05";
            };

          testScript = ''
            # The healthy module boots first: the runtime-shape asserts
            # need a live, listening service.
            machine.wait_for_unit("webphone.service")
            machine.wait_for_open_port(8080)

            # Health is honest while healthy.
            machine.succeed("curl -sf http://127.0.0.1:8080/healthz")

            # /hooks/* without a configured secret fail CLOSED.
            machine.succeed(
              "test \"$(curl -s -o /dev/null -w '%{http_code}' -X POST "
              "http://127.0.0.1:8080/hooks/message -d '{}')\" = 503"
            )

            # Wedge the blob store under the running service: the
            # healthz leg must name its failure, not lie with a 200.
            machine.succeed("chmod 000 /var/lib/webphone/files")
            machine.wait_until_succeeds(
              "test \"$(curl -s -o /dev/null -w '%{http_code}' http://127.0.0.1:8080/healthz)\" = 503",
              timeout=30,
            )
            machine.succeed(
              "curl -s http://127.0.0.1:8080/healthz | grep -q 'blob-dir'"
            )
            machine.succeed("chmod 750 /var/lib/webphone/files")
            machine.wait_until_succeeds(
              "curl -sf http://127.0.0.1:8080/healthz",
              timeout=30,
            )

            # Scenario A: the designed boot failure under systemd —
            # DATA_DIR parked under a regular file (ENOTDIR in the
            # storage phase), Restart=always so the loop is real.
            machine.succeed("touch /tmp/blocked")
            machine.succeed(
              "systemd-run --unit=webphone-drill -p Restart=always -p RestartSec=1 "
              "--setenv=WEBPHONE_ADDR=127.0.0.1:18099 "
              "--setenv=WEBPHONE_DATA_DIR=/tmp/blocked/data "
              "${self'.packages.webphone}/bin/webphone"
            )
            machine.wait_until_succeeds(
              "journalctl -u webphone-drill --no-pager | grep -q 'boot failed (class=data-dir'",
              timeout=60,
            )
            # Exit taxonomy FIRST, and RETRIED: Restart=always resets
            # ExecMainStatus to 0 the moment a new cycle starts, so a
            # single-shot read races the 1s restart loop (the 2026-10-29
            # flake: five journalctl greps in between were enough to lose
            # the window). Every failed cycle re-exposes ExecMainStatus=1,
            # so retrying observes one instead of hoping to win a race.
            machine.wait_until_succeeds(
              "systemctl show webphone-drill -p ExecMainStatus "
              "| grep -q 'ExecMainStatus=1'",
              timeout=30,
            )
            for marker in ("WHAT:", "REASSURE:", "WHY:", "FIX:", "ESCAPE:"):
                machine.succeed(
                  f"journalctl -u webphone-drill --no-pager | grep -q '{marker}'"
                )
            machine.wait_until_succeeds(
              "test \"$(systemctl show webphone-drill -p NRestarts --value)\" -gt 1",
              timeout=30,
            )
            machine.execute("systemctl stop webphone-drill")
          '';
        };
      };
    };
}
