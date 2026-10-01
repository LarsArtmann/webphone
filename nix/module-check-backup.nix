# backup.* cases for the webphone-module check: the timer/oneshot pair,
# the retention gate, and the destDir assertion. Extracted from
# module-check.nix (monolith split).
{
  lib,
  pkgs,
  base,
}:
let
  inherit (base) moduleSet;
in
[
  {
    # backup.enable must render the timer (OnCalendar + Unit)
    # and the oneshot service.
    name = "backup-timer";
    path = pkgs.writeText "backup-timer" (
      let
        backupEvaluated = lib.evalModules (moduleSet {
          backup.enable = true;
        });
        timer = backupEvaluated.config.systemd.timers.webphone-backup;
        service = backupEvaluated.config.systemd.services.webphone-backup;
      in
      if
        timer.timerConfig.OnCalendar == "*-*-* 04:30:00"
        && timer.timerConfig.Unit == "webphone-backup.service"
        && timer.wantedBy == [ "timers.target" ]
        && service.serviceConfig.Type == "oneshot"
      then
        "backup timer + oneshot rendered"
      else
        throw "webphone-module check: backup.enable did not render the timer/oneshot pair"
    );
  }
  {
    # backup.retentionDays: null (default) must render the
    # plain single-snapshot script; a number must add the
    # dated-history branch (snapshots dir, link-dest basis,
    # bounded prune).
    name = "backup-retention";
    path = pkgs.writeText "backup-retention" (
      let
        plainScript =
          (lib.evalModules (moduleSet {
            backup.enable = true;
          })).config.systemd.services.webphone-backup.script;
        retentionScript =
          (lib.evalModules (moduleSet {
            backup.enable = true;
            backup.retentionDays = 7;
          })).config.systemd.services.webphone-backup.script;
      in
      if
        !lib.hasInfix "snapshots" plainScript
        && lib.hasInfix "snapshots" retentionScript
        && lib.hasInfix "--link-dest" retentionScript
        && lib.hasInfix "-mtime +7" retentionScript
      then
        "backup retention branch renders per option"
      else
        throw "webphone-module check: backup.retentionDays did not gate the history/prune script correctly"
    );
  }
  {
    # backup.destDir is policed like dataDir: an off-/var/lib
    # path must trip exactly the destDir assertion (and the
    # base backup eval must stay free of failed assertions).
    name = "backup-destdir-assertion";
    path = pkgs.writeText "backup-destdir-assertion" (
      let
        failedAssertionsOf =
          extra:
          let
            evaled = lib.evalModules (moduleSet extra);
          in
          lib.filter (a: !a.assertion) evaled.config.assertions;
        bad = failedAssertionsOf {
          backup.enable = true;
          backup.destDir = "/tmp/webphone-backup";
        };
        clean = failedAssertionsOf { backup.enable = true; };
      in
      if lib.length bad == 1 && lib.length clean == 0 then
        "backup.destDir assertion fires exactly off-/var/lib"
      else
        throw "webphone-module check: backup.destDir assertion mis-fires (bad=${toString (lib.length bad)}, clean=${toString (lib.length clean)})"
    );
  }
]
