# The backup oneshot script for services.webphone.backup: an online
# sqlite .backup plus a blob rsync, with an optional dated-history and
# prune branch when retentionDays is set. Extracted from
# nixos-module.nix so the module wiring and this embedded shell program
# stay separately reviewable (nix-review monolith guideline).
{ lib, cfg }:
let
  dest = cfg.backup.destDir;
  history = cfg.backup.retentionDays != null;
in
''
  sqlite3 ${cfg.dataDir}/webphone.db ".backup '${dest}/webphone.db'"
  rsync -a --delete ${cfg.dataDir}/files/ ${dest}/files/
''
+ (lib.optionalString history ''
  # Dated history: snapshot today's state under snapshots/,
  # hardlinking unchanged blobs against the newest previous
  # snapshot (first run has no basis — plain copy; a same-day
  # rerun never uses itself as the basis).
  today=$(date +%F)
  snap="${dest}/snapshots/$today"
  basis=$(
    find ${dest}/snapshots -mindepth 1 -maxdepth 1 -type d \
      -name '????-??-??' ! -name "$today" 2>/dev/null | sort | tail -1 || true
  )
  mkdir -p "$snap"
  sqlite3 ${cfg.dataDir}/webphone.db ".backup '$snap/webphone.db'"
  if [ -n "$basis" ]; then
    rsync -a --delete --link-dest="$basis" ${cfg.dataDir}/files/ "$snap/files/"
  else
    rsync -a --delete ${cfg.dataDir}/files/ "$snap/files/"
  fi
  # Prune: only the dated directories themselves, never the
  # latest top-level snapshot pair, and never anything that
  # is not a YYYY-MM-DD name.
  find ${dest}/snapshots -mindepth 1 -maxdepth 1 -type d \
    -name '????-??-??' -mtime +${toString cfg.backup.retentionDays} \
    -exec rm -rf -- {} +
'')
