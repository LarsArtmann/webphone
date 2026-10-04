# Status: tooling-stragglers run — v6-sweep items executed, registry shipped, honest misses logged

**Session:** 2026-10-04, ~06:50–07:32 CEST · **Scope of this report:** this
session's work only (per operator instruction — no repo-wide re-audit).
**Mandate:** work the TODO_LIST assistant-legal items end to end
(research → execute → verify), then self-review.

**Baseline → end:** TODO tooling row: 3 open assistant items → 0 (row now
owner-remainder only); smoke mypy 8 → 0; error-code registry:
absent → 104-code generated table, freshness-pinned by a new arch test;
one latent fail-open walk bug found and fixed. Every gate that ran: green
(the only buildflow findings are the documented gomod-check vendor FP).

---

## a) FULLY DONE (verified this session)

1. **`scripts/build-health-css.sh` trailing `nix fmt`** (TODO v6-sweep
   item; t21-t22 f1). Appended a formatter step SCOPED to the artifact
   (a repo-wide `nix fmt` would trample concurrent in-flight edits — the
   AGENTS concurrent-sessions rule). LIVE-VERIFIED: ran the full script —
   the raw rebuild produced a file prettier had to change ("formatted 1
   files (1 changed)"), landing byte-identical with HEAD afterwards. The
   exact "rebuild re-breaks the `format` gate" failure mode is now
   impossible from this script.
2. **`scripts/webphone-smoke.py` mypy 8 → 0** (02:14 f7, twice-carried).
   Reproduced the exact 8 findings with nixpkgs mypy 2.1.0 first, then
   fixed at ROOT CAUSE — the annotations were the lie, not the code:
   `hook` returns a 3-tuple (declared 2) which also poisoned four
   unpack/arg-type findings; `sse_events`/`wait_for` receive
   `list[tuple[str, str]]` (declared `list[str]`) which poisoned the
   append + str-unpack + 2 arg-type findings; the `stop_set` bool
   defeated mypy narrowing (inlined the condition); plus one
   `check=False` consistency fix (the file's own convention, silences
   the default-ruleset PLW1510). Verified: mypy clean, `ruff format
   --check` clean, `ruff check` clean, `py_compile` ok, `--help` ok, and
   the FULL LIVE SUITE re-run green — **47 + 4 + 8 checks, 0 failed**,
   fresh binary, restart + boot-failure scenarios included.
3. **Error-code registry + "error families are total" story**
   (family-adoption f7/f24/f35, TODO v6-sweep item). New
   `internal/arch/errorregistry_test.go`:
   `TestErrorCodeRegistryIsFresh` walks every non-test Go file, extracts
   every `errorfamily` code literal (all constructor shapes incl. the
   two runtime-split sites `pbx.http`/`crm.http`), validates the
   `<seam>.<op>` shape (P5), and compares against a GENERATED table in
   `docs/error-contract.md` — **104 unique codes**. A rename now fails
   CI instead of silently breaking journal
   `grep '[family:code]'` recipes. `-update` regenerates the block;
   NEGATIVE-TESTED in both directions (a sed-inserted bogus doc row
   failed with "documented code has no source constructor: …"; restore →
   green). `docs/error-contract.md` gained the family-doctrine section
   (six families, P1–P7 pointer, `[family:code]` vocabulary,
   runtime-split explanation, regen command).
4. **Latent fail-open bug fixed: `filepath.SkipAll` from a DIRECTORY
   callback terminates the ENTIRE WalkDir — and the error is swallowed
   at the top.** My first registry version inherited the pattern from
   the sibling arch test and the walk died silently at `.git` ("walked
   0 go files", nil error). Fixed in BOTH files
   (`SkipAll` → `SkipDir`): `internal/arch/arch_test.go` carried the
   same bug un-triggered since its birth (its `.git`/`vendor` skip list
   would have vacuously passed a compliance walk had those dirs ever
   appeared under `internal/`). Lesson recorded in docs/lessons.md.
5. **Bookkeeping** (docs-health conventions): TODO_LIST tooling row
   pruned to its owner-remainder with the three DONEs moved into
   Evidence (dated 2026-10-04); CHANGELOG `[Unreleased]` entries added
   (registry + tooling batch); lessons.md entry (SkipAll/SkipDir trap +
   generated-doc-beats-hand-curated + "the annotations lied").
6. **Gates, all green:** `go test -count=1 ./...` (19 packages, 0 FAIL);
   live smoke (see #2); `nix fmt` clean on all touched Go files;
   `buildflow` — the ONLY remaining findings are the documented
   gomod-check vendor-consistency false positives (54; AGENTS: do NOT
   hand-edit vendor markers), with 46 steps "not applicable" (the
   documented JS/TS+Python noise). Auto-commit daemon absorbed the
   script/doc commits mid-session (observed, expected).

## b) PARTIALLY DONE

1. **TODO "Tooling hygiene batch" row** — my three v6-sweep items are
   done and moved to Evidence; the row REMAINS for its owner legs
   (markdownlint posture, codespell policy for archived snapshots,
   buildflow-claim reconcile follow-ups, BuildFlow-binary freshness).
   Not mine to close.
2. **Boot-contract tail, item (2)** — the stack ops-runbook patch
   (`docs/planning/2026-10-02_11-05_boot-contract-stack-runbook-patch.md`)
   was NOT applied by this session: it belongs to the tri-repo ritual
   (webphone pushed first, clean stack tree, relock) and rides the
   cross-repo train. `docs/error-contract.md` § "Cross-repo sync" still
   honestly names it as pending. Owner/stack dispatch.
3. **Registry polish** — the extractor works and is pinned, but it is a
   REGEX over source, not `go/ast`. It fails LOUD on shapes it cannot
   parse (by design), which is safe today (137-ish constructor sites,
   zero surprises), but an AST-walk is the strictly-correct tool.
   Deferred, not forgotten — see (f).
4. **The negative test is manual.** I proved the drift guard bites via a
   live sed-insert/restore, but I did NOT commit an automated negative
   case (e.g. a unit test over `parseRegistryRows`/synthetic drift).
   The freshness direction is automated; the stale-row direction is
   currently guarded only by the main comparison (which IS automated —
   a stale doc row fails the plain `go test`), so the risk is small,
   but the explicit synthetic-case test is still owed.

## c) NOT STARTED (in TODO_LIST, untouched this session, with the blocker)

- **v2.8.0 deploy tail** — OWNER terminal (pbx-artmann AGENTS forbids
  assistant ssh/deploy). Nothing for me to do.
- **SMS bridge 422 journal leg** — OWNER (ssh forbidden); triage pack
  ready since 2026-10-01.
- **D3 retry-loop owner call** — owner decision; module deliberately
  unchanged.
- **markdownlint posture / codespell archived-snapshot policy** — owner
  calls (briefing row 18).
- **OWNER-calls batch sitting** — briefing ready; not my sitting.
- **Release announcements posting** — owner picks channels/wording.
- **Standing watches** — not due: erraudit re-measure + boot-surface
  re-grade next 2026-10-22; sip.js 0.22 + templ-components 1.20.x +
  quarterly items next 2026-12-20.
- **internal/server carve** — trigger-based (next file added to
  internal/server); trigger not hit today (still 19 files).
- **Mic pre-warm live ritual** — owner (live call on the stack).
- **Island-honesty stack browser E2E** — blocked upstream (FreeSWITCH
  `mod_enum` build break; repair first), then relock to `cc98c2e`+.
- **samber/do follow-ups** (stack `/health` exposure policy, v2.9.0
  fold) — owner calls.
- **Visual-harness shot disposition + vision-CLI cross-check** — owner.
- **Cross-repo obligations** (paperless module option + smoke arm,
  WebTransport verdict doc, deploy.md secret column, demo-call recipe,
  MOH/recordings/CDR check, `ftypqt` sniff fix, E2E MMS-outbound,
  pbx-artmann FEATURES:87 text) — no assistant ssh; verified handovers
  only.

## d) TOTALLY FUCKED UP (this session's honest misses)

1. **I shipped a test that silently walked NOTHING on its first
   repo-wide run — because I copied a bug I didn't question.** The
   `SkipAll`-from-directory pattern was sitting in the reviewed,
   long-standing arch test and I reused it wholesale. "The sibling
   worked" is not verification: the sibling had simply never met a
   `.git` dir. It took three debug cycles (instrumented walk logging)
   to find. The save was that I had a `len(entries) == 0` tripwire —
   without it the registry would have "passed" as an empty table
   against a doc full of markers-missing failures downstream. Lesson
   recorded; the meta-lesson (question copied invariants, add
   emptiness tripwires to any extraction) is in lessons.md.
2. **An unexplained smoke-count delta left standing.** AGENTS says
   "48-check live smoke"; my green run printed "47 passed". My edits
   demonstrably added/removed no checks, so the delta is almost
   certainly the two version-enrichment checks being conditional on
   commit metadata that a bare `go build` binary lacks — but I did NOT
   verify that hypothesis, and I did NOT fix the AGENTS/script count
   mismatch. A 1-check discrepancy in a load-bearing gate description
   should not survive a session that touched that very script. Owed in
   (f).
3. **Two first-draft bugs that self-review caught only pre-run:** the
   `parseRegistryRows` field-split would have poisoned map keys with
   trailing spaces (caught by re-reading my own diff before running),
   and the doc path was package-relative-wrong (`../docs` vs
   `../../docs`, caught by the run). Neither escaped, but both were
   sloppy drafting, and the parsing one would have produced a silently
   WRONG freshness map (all keys trailing-space-mangled → everything
   "missing") — loud, but embarrassing.
4. **Edit/daemon race mid-chain:** one `edit` was rejected with
   "modified since last read" because my own `ruff format` run updated
   the file after my read; I re-read and re-applied (no damage), but
   the AGENTS warning about the adversarial daemon applies to MY OWN
   tooling steps too — format-then-edit ordering bit me once.
5. **Scope honesty about the registry's guarantee:** the test pins
   codes to the DOC — it does NOT pin codes to the WIRE. If a code
   rename is committed together with a regen (`-update`) in the same
   change, CI stays green while journal consumers break. The guard is
   review-and-ritual, not the test alone. I shipped the test without
   wiring a release-time re-gen step (see (f) item 3) — the family
   table in the release gates was f35's full ask and I under-shipped
   it by one notch.
6. **`nix flake check` was not run.** Justified by scope (no nix,
   island, or css changes; buildflow + full go test green), but the
   honest ledger entry is: the arch test inside the sandbox lane is
   UNVERIFIED this session — `nix develop -c go test` is my substitute,
   not the flake's own harness.

**Split-brain scan (session-introduced): none found.** The registry
lives in exactly one doc; AGENTS points at error-contract.md without
duplicating codes; lessons/CHANGELOG/TODO carry the EVENT, not copies of
the table. Two time-bombs flagged, not planted: (i) CHANGELOG now says
"104-code" — a generated count baked into prose that will go stale at
the next regen (acceptable for history, avoid going forward); (ii) the
BEGIN/END marker text embeds the test name — renames surface as a loud
"markers missing" failure (guarded, by design).

**Ghost-system scan: none.** The registry test runs in `go test ./...`,
buildflow's test steps, and (unverified but expected, see d6) the flake
sandbox lane; nothing was added unwired.

## e) WHAT WE SHOULD IMPROVE

1. **Extraction correctness:** replace the regex extractor with a
   `go/ast` walk (ast.Inspect + selector/call matching). Exact arg
   resolution, immune to formatting and nested-call shapes; the shape
   validation stays as a policy check.
2. **Automate the negative:** a committed unit test asserting
   `parseRegistryRows` round-trips generated rows and rejects a
   synthetic stale row — the manual sed proof should not be the only
   evidence.
3. **Close the regen loophole:** add the registry re-gen check to the
   release ritual (release-runbook step or release.sh grep — a one-line
   `go test ./internal/arch -run TestErrorCodeRegistryIsFresh` before
   tagging), so a mid-train rename cannot ride a green `-update` into a
   tag without a human look.
4. **Smoke-count truth:** make the AGENTS smoke bullet cite the suite's
   own summary line (or pin the check inventory in the script) so the
   47/48 class of drift cannot recur.
5. **Match local tooling to the gate:** the PLW1510 finding came from
   nixpkgs ruff's default ruleset, not necessarily buildflow's — a
   committed minimal `ruff.toml`/`mypy.ini` for `scripts/*.py` would
   make local runs reproduce gate verdicts exactly.
6. **Ordering discipline around the daemon:** format FIRST, then read,
   then edit — my own formatter invalidated my own read buffer once.

## f) Next things to get done (brainstorm, impact-sorted; most beyond the first block are ROUTING fuel for docs-health HARVEST, not commitments)

| #  | Thing                                                                                                                                | Effort | Route                                   |
| -- | ------------------------------------------------------------------------------------------------------------------------------------ | ------ | --------------------------------------- |
| 1  | Explain + fix the 47-vs-48 smoke count (verify the enrichment-conditional hypothesis; pin AGENTS to the suite's summary)             | S      | TODO tooling row                        |
| 2  | go/ast-based registry extractor (replace regex; keep shape validation)                                                               | S      | TODO tooling row                        |
| 3  | Registry re-gen into the release ritual (release.sh/runbook line) — close the `-update` loophole                                     | S      | TODO tooling row                        |
| 4  | Automated negative test for `parseRegistryRows` + synthetic drift                                                                    | S      | TODO tooling row                        |
| 5  | Sweep ALL repo WalkDir callers for the SkipAll-from-directory pattern (the two arch sites are fixed; are there others?)              | S      | TODO tooling row                        |
| 6  | Commit `ruff.toml`/`mypy.ini` for scripts/ so local runs match gate rulesets                                                         | S      | TODO tooling row                        |
| 7  | Boot-contract stack runbook patch application (tri-repo ritual; patch text ready)                                                    | M      | TODO boot-contract row (stack dispatch) |
| 8  | v2.8.0 deploy tail (stack lock bump → stack gates incl. owed browser E2E → aarch64 → pbx relock → deploy → `--expect-version` smoke) | M      | OWNER                                   |
| 9  | SMS-bridge journal leg (unit status → grep → creds → restart → test SMS); record root cause                                          | S      | OWNER                                   |
| 10 | Stack FreeSWITCH `mod_enum` build repair → THEN the T11–T19-owed stack browser E2E → relock `cc98c2e`+                               | M      | OWNER/stack                             |
| 11 | markdownlint posture decision (house-style config vs recorded detect-only)                                                           | S      | OWNER (briefing 18)                     |
| 12 | erraudit tier re-measure 2026-10-22 + boot-surface re-grade + context_loss scan-site sweep (three legs, one sitting)                 | S      | TODO watches row                        |
| 13 | OWNER-calls batch sitting (28-row briefing; now incl. registry-is-release-artifact question)                                         | S      | OWNER                                   |
| 14 | Stack `/health` exposure policy + v2.9.0 fold decision                                                                               | S      | OWNER                                   |
| 15 | Render-diff recipe parking in AGENTS/lessons (row claims script committed+verified; recipe parking never verified)                   | S      | TODO tooling row                        |
| 16 | Quarterly standing-watch sweep 2026-12-20 (sip.js 0.22, templ-components 1.20.x, E2E wall-time budget)                               | S      | TODO watches row                        |
| 17 | internal/server carve trigger check (next file added → carve `server/api` + `server/hooks`)                                          | M      | TODO carve row (trigger-based)          |
| 18 | Mic pre-warm live ritual (accept→speak, indicator timing, warm release)                                                              | S      | OWNER                                   |
| 19 | Visual-harness shot disposition (per-release vs per-train) + vision-CLI cross-check decision                                         | S      | OWNER                                   |
| 20 | Cross-repo: `services.webphone.paperless` module option + smoke arm                                                                  | M      | TODO cross-repo row                     |
| 21 | Cross-repo: WebTransport not-adopted verdict doc                                                                                     | S      | TODO cross-repo row                     |
| 22 | Cross-repo: telephony `deploy.md` secret PATH column                                                                                 | S      | TODO cross-repo row                     |
| 23 | Cross-repo: ops-runbook demo-call recipe (`originate user/1000 &playback(local_stream://moh)` + password path)                       | S      | TODO cross-repo row                     |
| 24 | Cross-repo: MOH audibility + `/recordings/` + CDR check                                                                              | S      | TODO cross-repo row                     |
| 25 | Cross-repo: `ftypqt`→`video/quicktime` sniff fix (stack bridge)                                                                      | S      | TODO cross-repo row                     |
| 26 | Cross-repo: stack E2E MMS-outbound coverage                                                                                          | M      | TODO cross-repo row                     |
| 27 | Cross-repo: pbx-artmann FEATURES:87 stale sniff text                                                                                 | S      | TODO cross-repo row                     |
| 28 | aarch64 ELF-byte verify at the next final gate (owed at each cross-build close-out)                                                  | S      | rides next release                      |
| 29 | Release announcements: owner picks channel + disclosure posture (drafts A/B/C ready since 2026-09-30)                                | S      | OWNER                                   |
| 30 | codespell policy for archived status snapshots (exclude `docs/status/**` or fix the words)                                           | S      | OWNER                                   |
| 31 | Reconcile AGENTS buildflow-full claim vs observed skips (if the 2026-09-26 skips re-appear)                                          | S      | TODO tooling row                        |
| 32 | BuildFlow binary freshness re-check (doctor) at next gate run                                                                        | S      | TODO tooling row                        |
| 33 | Registry table: human-facing anchor links from AGENTS erraudit bullet to the registry section                                        | S      | ROADMAP                                 |
| 34 | Consider `/metrics` error-code label now that the closed code set is published (f18 revisit)                                         | M      | ROADMAP                                 |
| 35 | Consider `errorfamily` family column in the stack runbook's error-contract mirror (f7's cross-repo half)                             | S      | rides stack runbook patch               |
| 36 | Drop generated counts ("104-code") from CHANGELOG-style prose going forward (process note for docs-health)                           | S      | process                                 |
| 37 | Smoke suite: make version-enrichment checks skip-counted when the binary lacks commit metadata (stable totals)                       | S      | TODO tooling row                        |
| 38 | Island oxlint globals watch: next new browser global → `island-lint` config (standing)                                               | S      | standing watch                          |
| 39 | Push-lag threshold policy (when is a silent daemon push stall BROKEN)                                                                | S      | OWNER (briefing)                        |
| 40 | AGENTS restructure permission (the 377-line cap pressure)                                                                            | S      | OWNER (briefing)                        |
| 41 | Existing-prod-data chmod/re-backup for the UMask tightening                                                                          | S      | OWNER                                   |
| 42 | `destDir` nesting legality ruling (backup module)                                                                                    | S      | OWNER                                   |
| 43 | Registry: doc the "deliberate rename" checklist inline (regen command + journal-consumer notice)                                     | S      | TODO tooling row                        |
| 44 | Add registry freshness to `nix flake check` expectations explicitly (verify the sandbox lane runs it today)                          | S      | TODO tooling row                        |
| 45 | `scripts/*.sh` shellcheck posture (buildflow runs it? verify; wire or record skip)                                                   | S      | TODO tooling row                        |
| 46 | buildflow timing-regression verdicts: document "cold cache" as expected noise or add warm-up step                                    | S      | process                                 |
| 47 | Self-send 422 sign-in banner browser check (live proof still open since 2.6.0)                                                       | S      | OWNER (rides deploy)                    |
| 48 | Sniff-fallback lifespan ruling (keep forever vs delete-after-deploy-confirmed)                                                       | S      | OWNER (briefing)                        |
| 49 | `gateway.attachment_limit` knob vs bridge-422-teaches design final call                                                              | S      | OWNER (briefing)                        |
| 50 | Lesson to crush-config `references/lessons.md` (cross-project: SkipAll/SkipDir walk trap) — commit-only, other repo                  | S      | crush-config repo                       |

## g) Questions I can NOT figure out myself

1. **Registry guarantee level:** is the error-code registry a CI artifact
   (doc must stay green on every train — current behavior) or a RELEASE
   artifact (regen + human eyeball required at tag time, allowing
   deliberate mid-train renames to ride `-update`)? This decides whether
   item 3 lands in release.sh or stays a convention.
2. **Smoke-count canonicalization:** should AGENTS stop quoting a
   literal check count ("48-check") and instead require the suite's own
   summary line to be pasted into evidence — accepting that the count
   drifts with conditional checks — or do you want a literal,
   gate-enforced inventory? (I can implement either; the intent is
   yours.)
3. **Local-tooling parity:** do you want committed `ruff.toml`/`mypy.ini`
   pins for `scripts/*.py` mirroring BuildFlow's ruleset (locals exactly
   reproduce gates, but a second config to maintain), or should locals
   stay tool-default and buildflow remain the only arbiter?

---

**Verification ledger (commands as run):** `nix shell nixpkgs#mypy -c mypy
scripts/webphone-smoke.py` (8 findings → 0) · `nix shell nixpkgs#ruff -c
ruff format --check / ruff check scripts/webphone-smoke.py` (clean) ·
`python3 scripts/webphone-smoke.py --help` + full live run (47+4+8, 0
failed) · `bash scripts/build-health-css.sh` (rebuild + fmt + tree clean)
· `nix develop -c go test -count=1 ./...` (all ok) · `nix develop -c go
test ./internal/arch -run TestErrorCodeRegistryIsFresh [-update]` (ok;
negative test failed as designed, restored) · `nix fmt` on touched Go
files (clean) · `nix develop -c buildflow` (only gomod-check 54, the
documented FP).

**Format note:** the status-report skill's canonical output is a styled
HTML dashboard; your instruction named this `.md` path explicitly, so
this report is Markdown by your override (flagged, not propagated into
the skill). The brutal-self-review skill's separate HTML report at
`docs/reviews/` was likewise folded into sections (d)/(e) here per your
single-report instruction.

**Per the skill contract: WAITING FOR FURTHER INSTRUCTIONS.**
