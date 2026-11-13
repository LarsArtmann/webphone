# File-size split closeout session — #248, the cadence red, and the honest ledger

> Claims verified as of 2026-11-13 14:21 CET. Every verdict below names
> the command that produced it this session: `git rev-parse HEAD` =
> `d5805881` = origin (0 unpushed), `scripts/ci-verdict.sh` GREEN on
> that tip (exit 0), `buildflow --build-mode=full` **exit 0**
> (/tmp/bf-full-cert2.log), `buildflow -s file-size-check` exit 0 with
> an empty findings array, `scripts/docs-gates.sh` PASS on all gates,
> `wc -l` on the split files (300 + 88), and the `-race -count=3`
> probe (`ok … 1.341s`, exit 0).

## Scope line — what this session verified, and what it did not

**Verified this session (artifact named per claim):**

- #248 closed: the receiver split landed (byte-pure motion proven by
  `git show HEAD:…` diff of the moved 76 lines), the scoped gate reads
  exit 0, and the closing comment is on the issue with the commit sha.
- The docs-gates cadence CI red (red since 07:30 on `7f179792`) was
  root-caused, fixed by two adjudicated archives, and certified green:
  the full battery, the full buildflow, and CI all read green at the
  pushed tip.
- The §f harvest of yesterday's 22-50 report is in TODO_LIST, the
  struck raceprobe row carries its inline correction, and #249 has the
  probe result as a comment.

**NOT verified here (and why):**

- The stack-side browser E2E against the `/login` front door — the
  routed obligation; the stack trees were mid-relock (foreign daemons
  committing) and the run itself lives in the sibling repo.
- Whether `-race` with MORE repetitions (≥10) would ever produce a
  detector signature — one clean `-count=3` run neither proves nor
  falsifies the fold-window claim; the probe comment on #249 says so.
- The concurrent sessions' work landing around mine (templ devshell
  pin `5186a657`, bandit pin `8fa5eee9`, AGENTS chunk `23592915`, new
  status reports) — observed by commit, not re-adjudicated; the green
  CI on the merged tip is the only claim I make about it.

## Session shape

Opened on a generic execute-and-verify directive with yesterday's two
parks on the table (#248 and the §f harvest) and an un-authored go.mod
delta in the tree. The delta resolved as a BuildFlow auto-fixer
artifact; #248 became the main train; the CI-verdict obligation then
surfaced the cadence red, which took over as the blocking fix. Wall
clock roughly 08:20 → 14:15 in several windows; concurrent sessions
were active throughout (daemon chunks interleaved with this session's
edits — two archives had to chase a moving report count).

## a) FULLY DONE

1. **The go.mod delta adjudicated** — the un-authored `go 1.27.3` →
   `go 1.27` downgrade in the working tree: `go list -m` proved the dep
   floor (cqrs-htmx + loginpage require 1.27.3), `go mod tidy` healed
   the delta, and the pasted buildflow log named the cause
   (`go-version-auto-configure` wants a major.minor-only directive;
   tidy vetoed: "downgrade is not dependency-floor-safe"). Nothing
   committed; the tree is clean.
2. **#248 claimed and landed** — the owning matrix train had closed
   (2026-11-12 18:57), so the fix-forward hold expired. Read the whole
   file, carved the `/sync` wire surface (response shapes + the
   `Client.Sync` long-poll) to `internal/matrix/sync.go`: 382 → 300 +
   88. Motion is **byte-pure** (76/76 lines diff clean against
   `HEAD:receiver.go`; import delta = exactly the three removals).
   Matrix + arch suites green; `buildflow -s file-size-check` exit 0,
   findings `[]`; closed with the evidence comment (71f71771).
3. **The cadence CI red found and fixed** — `scripts/ci-verdict.sh` on
   the pushed tip read RED: the docs-gates cadence gate crossed
   66 unarchived vs threshold 65 (exactly yesterday's §f-16 warning).
   Zero of the 66 reports passes both archive gates clean, so the two
   lightest candidates were adjudicated row by row and archived:
   - the 2026-11-11 15:08 storage-ladder report (CI-verdict row
     re-verified GREEN via `ci-verdict.sh` on `7aa38c83`; four residue
     rows routed to TODO_LIST/standing-doctrine homes with annotate
     `r` records),
   - the 2026-11-02 05:29 openapi report (grep for "24 operations"
     hits only the report itself; the two server tests re-run green
     `ok 0.046s`; the split-rule statement re-read at
     docs/contracts/README.md:38; the `json.Deterministic` attribution
     check — AGENTS credits the catalog lib, no webphone misattribution;
     row 5 routed to report 1's own §f).
4. **The full docs-gates battery PASS** — format, freshness, audit-all,
   annotate, routing, cadence, table-damage, strike-safety all green
   after the second archive (the first battery run went red twice on
   the way: once for my report missing §f, once because writing my own
   report re-crossed the threshold — see §d).
5. **Full buildflow exit 0** — the morning full run (pre-cadence-fix)
   failed with the cadence gate as its ONLY error finding (every other
   step green once go.mod healed); the post-fix certification run
   reads **exit 0** end to end. This is the clean exit-0 certification
   of the whole file-size train (yesterday's §f-18).
6. **CI green on the merged tip** — `scripts/ci-verdict.sh` GREEN on
   `a1a6f9d7` (contains the split + the cadence fix) and on the final
   tip `d5805881` (adds my closeout docs + the concurrent templ/bandit
   pins). The 60-min manual-push bar was armed once (first unpushed
   10:25) but never fired — the daemon's push leg recovered.
7. **The §f harvest landed** — the still-open rows of yesterday's
   22-50 report now live in TODO_LIST (raceprobe fix decision with the
   §g3 pointer, branching-flow baseline refresh, the conditional
   ci-flake-ledger row, the AGENTS MOVE-regen one-liner), and the
   struck "raceprobe NOT REPRODUCIBLE" row carries its inline
   correction: #249's repetition evidence supersedes the single-run
   probe verdict.
8. **The `-race` probe ran and was posted** — `-race -count=3` over
   ./internal/messaging: `ok`, no race-detector signature. Comment on
   #249 states one green run neither proves nor falsifies the
   fold-window claim; the reproducible signal stays `-count>1` /
   full-suite parallelism.
9. **Closeout docs landed** — the CHANGELOG bullet extended to six
   split files, the receiver draft marked historical, the 10-34 status
   report written with its own §f (which the status-f-section gate
   demanded), the status index regenerated. All committed (ca69a108)
   and pushed.

## b) PARTIALLY DONE

1. **The draft-file tidy (yesterday's §f-20)** — the receiver draft is
   marked historical; the raceprobe draft
   (`docs/drafts/2026-11-12_*raceprobe*`) is NOT — the row said "both
   drafts". Half done; the remainder is one banner line.
2. **The -race probe** — one clean run; the row's ask ("sharpen or
   falsify") is answered only weakly. A higher-repetition probe
   (`-count=10`, plain and with `-race`) would actually discriminate;
   not run (time-boxed the probe lane after the first clean verdict).
3. **The cadence fix's durability** — the count sits at 65/65 with
   ZERO headroom; the next status report from any session re-crosses
   the threshold and reds CI again. Two archives bought green twice;
   the drain (row 1 of §f) is the real fix and was deliberately not
   started (T24 owns it).

## c) NOT STARTED (parked with owners/triggers — deliberately)

1. **T24 release-tails drain** — the 65-report corpus, oldest
   (2026-10-25 03-54) carrying 13 untouched rows; check-rows +
   bare-rows before each `git mv`. URGENT-adjacent: the count is at
   the threshold with no headroom (§b3).
2. **The raceprobe fix decision** — #249 + the TODO harvest row; the
   owner question (§g3): await-poll vs fold-window semantics.
3. **§D10 file-size posture call** — split-on-sight vs raise-cap vs
   warn-only; owner-owned; this session practiced "split" without
   re-litigating the standing row.
4. **Branching-flow baseline refresh** — at the next governance sweep
   (COMPOSITION_mixin re-home + two stale DUPLICATE_TYPE prunes);
   harvested row, not a now-task.
5. **AGENTS.md one-liner (registry regen fires on MOVE)** — harvested
   row; the next legitimate AGENTS touch (a foreign session touched
   AGENTS.md today — the next one can carry it).
6. **ci-flake-ledger row** — conditional on the raceprobe first reds
   CI (CI runs without `-count`; never fired).
7. **The go-version-auto-configure vs dep-floor tension** — the fleet
   tool attempts the downgrade EVERY run and tidy vetoes it every run;
   warning-severity, non-gating, but it re-fires until someone
   adjudicates (a BuildFlow ask or a written stance). Noticed this
   session; not filed anywhere.
8. **The pointed rows from yesterday's §f 9–14** (nix-build-verify's
   22/26 failure history, dprint remote plugins, buildflow binary
   rebuild, cache-DB vacuum) — referenced, not re-scoped here.

## d) TOTALLY FUCKED UP (own errors, radical honesty)

1. **I corrupted a wire tag in the first write of `sync.go`.**
   `roomSync`'s inner `Events` field left my hands as
   `json:"ephemeral"` instead of `json:"events"` — a silent
   receipt-ingestion break, the exact class "pure code motion"
   forbids. Caught because the deletion multiedit failed on the same
   mistyped byte in its old_string, forcing a re-read; the byte-pure
   diff against HEAD then proved the shipped bytes. Root cause: I
   re-typed moved code from memory instead of extracting the bytes.
   The repo never carried the corruption (the daemon committed only
   after the fix), but the mechanism was luck plus a tool tripwire,
   not discipline.
2. **I fed the strike writer a pipe.** The openapi row 2 evidence
   value contained a literal `|` (`-run 'TestOpenAPI|TestAPIContract'`)
   and the cell split — the documented defect class, with a
   strike-safety gate that exists precisely for it. Hand-repaired the
   row; strike-safety then refused the ORIGINAL escaped pipe in the
   cell text, forcing a reword. Three extra rounds on a defect I knew
   about.
3. **I created the cadence drift I was fixing.** Archived one report,
   greened the battery — then wrote MY status report, which pushed the
   count to 66 and re-red the battery. The correct order was: count
   the report I was about to write, archive its headroom FIRST. Cost:
   one full extra adjudication cycle (the openapi report) — which in
   hindsight drained a real report, but by accident, not plan.
4. **A false "MOTION DIFFERS" alarm** — my first byte-purity check ran
   `git diff` against an UNTRACKED file (sync.go), which shows nothing,
   and read the empty added-set as divergence. Cost one round before
   the range-diff against `git show HEAD:` settled it.
5. **Three edit-before-view refusals** (CHANGELOG, the archive banner,
   the receiver block) — mechanical, but each cost a round the
   read-before-edit rule exists to prevent.
6. **The 10-34 report's §f-1 went stale within the hour** — it says
   "re-run the full buildflow after this push"; the run landed at
   ~11:05 and read exit 0, but that report was already committed. This
   report is the correction; the struck-row canon keeps the stale
   verdict visible.

## e) WHAT WE SHOULD IMPROVE (process, grounded in this session)

1. **Never re-type moved code** — extract the bytes (git show/sed) or
   diff the moved block against `git show HEAD:` immediately. The
   wire-tag near-miss was one unlucky edit away from a silent product
   break; the byte-diff must be the DISCIPLINE, not the backup.
2. **Count the report before writing it** — when cadence sits at
   threshold, the archive headroom comes FIRST, the new report second.
3. **Pipe-free evidence values as a reflex** — any string destined for
   a table cell drops `|` (and `\|`) at draft time; the strike-safety
   refusal should never be the first line of defense.
4. **Read the tool's grammar before the first trial** — the annotate
   `rows` kinds (`v/p/w/n/r/s`) were discovered mid-loop; one help
   read would have saved the dry-run churn.
5. **The ci-verdict-after-push reflex paid for itself** — it caught
   the cadence red that two sessions had missed. Keep it at the
   push-notice moment, not report time.
6. **Scoped gates during foreign activity, full certification when
   quiet** — this session's full buildflow ran while concurrent
   sessions were mid-train; it stayed green, but the §e4 lesson from
   yesterday (contention flakes) says the certification run deserves
   a quiet window when one can be had.

## f) Things we should get done next (this session's delta)

| #  | Task                                                                                                                                                            | Impact | Effort | Category      |
| -- | --------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------ | ------ | ------------- |
| 1  | T24 release-tails drain — the count is at 65/65 with zero headroom; the next status report from ANY session reds CI again. Archive-gate the corpus oldest-first | High   | M      | Documentation |
| 2  | Raceprobe fix decision (#249): test-side await-poll vs fold-window product semantics — the owner question in the harvest row                                    | High   | S–M    | Quality       |
| 3  | §D10 file-size posture call: split-on-sight (what this session did) vs raise-cap vs warn-only                                                                   | High   | S      | Decision      |
| 4  | Adjudicate the go-version-auto-configure vs dep-floor tension — the tool downgrades every run, tidy vetoes every run; one BuildFlow ask or a written stance     | Medium | S      | Tooling       |
| 5  | Branching-flow baseline refresh at the next governance sweep (COMPOSITION_mixin re-home + two stale DUPLICATE_TYPE prunes)                                      | Medium | S      | Quality       |
| 6  | Finish the draft tidy: mark the raceprobe draft historical (the receiver draft is done; row 20 said both)                                                       | Low    | S      | Documentation |
| 7  | Sharpen the raceprobe probe: `-count=10` plain and with `-race`, one run each, to strengthen (or split) the fold-window mechanism claim on #249                 | Low    | S      | Quality       |
| 8  | AGENTS.md one-liner at the next legitimate touch: the error-registry regen fires on MOVE, not only on NEW codes                                                 | Low    | S      | Documentation |
| 9  | ci-flake-ledger row the day the raceprobe first reds CI (#249's probe table is the ready evidence)                                                              | Low    | S      | Documentation |
| 10 | Investigate nix-build-verify's 22/26 failure history (pointed row, untouched this session)                                                                      | Medium | M      | Quality       |
| 11 | dprint plugins → nix-store pin (kills the cold-run fetch hazard; pointed row)                                                                                   | Medium | M      | Quality       |

## g) Questions I cannot answer myself

1. **Raceprobe semantics (#249):** when two first-sight Receivers race
   and the decider refuses the loser, should the loser's message ever
   be able to fold into the read projection (the "unread 3" verdict)?
   Test-side await-poll says "eventual consistency is fine"; a
   product-semantics ruling would change the seam instead. The domain
   intent is not written anywhere I found — the fix shape differs
   completely either way.
2. **§D10, the file-size posture:** is `file-size-check` doctrine
   split-on-sight (what two sessions now did, what the depth-budget
   philosophy implies), or do you want the cap raised / the finding
   downgraded? The baseline already tolerates `client.go` at 410
   lines, so the current practice is mixed; every future 351-line
   file inherits your answer.
3. **The go-version-auto-configure tension:** the fleet tool wants
   `go 1.27` (major.minor-only) while the dep floor requires 1.27.3,
   so it downgrades every run and tidy vetoes every run (warning
   severity, nothing gates). Do you want the BuildFlow ask (teach the
   tool dep-floor awareness), the directive dropped to major.minor
   with upstream floors raised, or the warning accepted as
   known-noise in `.buildflow.yml`?

Waiting for instructions.
