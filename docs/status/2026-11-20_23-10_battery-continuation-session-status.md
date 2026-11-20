# Status — the battery continuation session: safe-wave execution to green, the placeholder-pin heal, cadence drain (2026-11-20 23:1x)

**Snapshot-rot header: every claim below was verified as of 2026-11-20
~23:15 local.** Push/CI statements rot fastest; the CI verdict on the
heal commit was still PENDING at write time (§b) — poll
`scripts/ci-verdict.sh` before trusting.

Resuming the 2026-11-19/20 skill-battery session per its handoff
report (`2026-11-20_15-25_skill-battery-session-final-status.md`):
owner-gated work stayed gated (§D10, the deepening pick, the bump
trains, the release fold); everything else in the harvest waves was
executed READ→implement→test→verify, one unit per commit.

## a) FULLY DONE

1. **Final verification of the battery handoff (todo 13)** — the
   overnight trains' pushes carry a GREEN CI verdict (`6498c8ca`,
   `scripts/ci-verdict.sh` exit 0), the full Go suite answers 37/37
   packages green ×2 (`env -u GOEXPERIMENT go test -count=1 ./...`),
   and the catalog render drift my session left behind (the 1.1.0 +
   `review.saved` tree uncommitted at handoff) landed via the daemon
   byte-identical to a fresh re-render.
2. **Wave 1 — the fcr small-findings batch is CLOSED**: the favicon
   green split brain healed (the tab icon painted Tailwind-stock teal
   `#0f766e` against the brand `#17a467`; now accent + dark
   `--on-accent` ink, pinned by `TestAccentColorIsTheOneBrandGreen`
   across favicon.svg + manifest + app.css); the Fan panic-containment
   pin landed (`TestFanContainsAPanickingTransport`,
   mutation-verified: removing the recover crashes the test process);
   the T18 401/403 ask landed with the resource-scope nuance the row's
   wording missed — 401 stays credential-classified (an ntfy token
   rotation must not mass-prune), 403 now prunes (ntfy's ACL refuses
   THIS topic = the device identity; FCM's two documented device-scoped
   403 codes `SENDER_ID_MISMATCH` + `THIRD_PARTY_PERMISSION_DENIED`,
   verified against the Firebase error-code reference before
   encoding); the other three batch items (FCM token-mint single-flight,
   blob brand-before-write, the Seq retention-clock comment) were
   verified ALREADY LANDED by the overnight trains and struck with
   evidence.
3. **Wave 2 — the data-model #6 remainder is CLOSED**:
   `push.Transport` now threads through `PushDevices.Upsert`,
   `SavePushDevice`, the projection row and the app adapter (the API
   door's ParseTransport verdict no longer dies at the gate; stored and
   wire bytes unchanged; the adapter's parse stays as the pre-door-row
   guard); `SetDisposition` answers Set's `(changed, error)` verdict
   (pre-read skips echo dispatches, the decider stays the race
   authority, Ginkgo specs strengthened with
   `Expect(changed).To(BeFalse())`); `LoginPageExtra(showPrimary,
   showSSO bool)` became `LoginPageExtra(LoginLanes{PasskeyOn, SSOOn})`
   — callers state facts, the section derives presentation, served
   bytes unchanged; the overnight trains' `asr.Locality` classifier and
   `config.ASRProvider` enum — landed WITHOUT pins — got their pins
   (`TestLocalityOfClassifiesTheTrustTable` 16 cases incl. the
   172.16–31 range edges; `TestASRProviderVocabularyIsClosed`).
4. **Wave 3 — the naming adjudication train is CLOSED**: the three
   remaining renames landed (`views.VoicemailItem` →
   `views.VoicemailMessage` pairing with `views.VoicemailSummary`;
   `authmap.Map` → `authmap.Ownership` — DOMAIN_LANGUAGE's ownership
   term, fields `ownership` in userauth/sso; the twin `Shell` wiring
   interfaces → `auth.Server` / `platform.Server` so the carved seams
   stop stealing the ratified page-frame word). No wire bytes anywhere;
   full suite green after all three.
5. **Wave 4 — the mechanical tails are CLOSED**: `nix/vm-tests.nix`
   (535 lines) split into `nix/vm-tests/{webphone-backup,
   webphone-error-drill,webphone-oauth2,backup-drill}.nix` + a
   `default.nix` aggregator (flake.nix keeps the single import; the
   drill keeps the top-level `self` its script path interpolates; all
   four checks + webphone-module eval at the new paths); the erraudit
   JS `.catch` class closed by moving the ten intent comments INSIDE
   the catch bodies (`erraudit lint ./internal/web/assets/...` → 0
   findings; island tests 369/369; type gate green); the two NixOS
   module eval goldens moved to `nix/testdata/` (every path literal
   updated, checks eval); README gained the `## Install` section; the
   structure-linter tail went 4 findings → 1 known false positive (the
   linter flags the `openapi_golden_test.go` FILE NAME; its fixture
   already lives in testdata/).
6. **The dedup gate repair** — the disposition symmetry made
   `Set`/`SetDisposition` literal twins, so the
   lock/pre-read/echo-skip/dispatch/await sequence extracted into
   `Notes.upsertSlot` (closures carry what differs); the
   `pbx.VoicemailSummary` ↔ `views.VoicemailSummary` twin is the
   accepted seam→view conversion pair (registry row); baseline
   re-recorded with the CI-pinned v0.7.2 (114 → 112 entries, set-diff
   +1 adjudicated / −3 dissolved), `check` = 0 new.
7. **The cadence heal** — the day's reports pushed unarchived over the
   66 pin; the oldest report (2026-11-08 06-17) went through the
   ritual: all 18 ledger rows per-row adjudicated (13 done claims
   verified against their in-report evidence; 5 next-things struck with
   landed evidence or routed), both archive gates
   (`check-rows --strict`, `bare-rows`) pass, `git mv` + banner, and
   the one unfolded work item (the `--save-baseline` footgun draft
   fold) rides a new TODO_LIST row. `docs-gates run .` = 5/5 PASS.
8. **The placeholder-pin heal (main was red ~35 minutes)** — a
   concurrent session's hash-refresh dance left the templ source hash,
   templ vendorHash and webphone vendorHash as all-zero placeholders;
   the daemon swept the half-finished worktree onto main mid-dance and
   CI red on every `nix develop`-entering job. The derivation errors
   themselves name the correct values (no upstream dependency change —
   all three match the previously committed pins); all three restored,
   dev shell + `nix build .#webphone` green before push, ledger row
   recorded. Manual-push bar invoked for the 20-commit daemon stall
   (anchor: first unpushed `2026-11-20T20:24:52+01:00`, 144 min >
   60).

## b) PARTIALLY DONE

1. **The CI verdict on the heal push (`6a3368c5`)** was PENDING at
   report time — the red on the predecessor (`5c9d90b5`) is the
   placeholder class the heal fixes; the successor verdict must be read
   through `scripts/ci-verdict.sh 6a3368c5` before this report is
   cited as green.
2. **The README install section** covers the four lanes but the linter's
   one remaining INFO (the test-file-name false positive) needs an
   upstream ask if it is ever to stop re-flagging — recorded in the
   TODO strike, not filed upstream (verify-before-filing: the rule's
   heuristic would need reading in the tool's source first).

## c) NOT STARTED (owner-gated, unchanged)

- The §D10 file-size disposition → the v2.15.0 fold (release.sh's
  buildflow step still carries the 47-finding file-size class).
- The deepening pick (five candidates; top = messaging read surface).
- Bump Trains 1–4 vs the release ordering.

## d) TOTALLY FUCKED UP

1. **The templ regen used the right CLI but the wrong scope twice**:
   the first `sed` batch on `vm-tests.nix` boundaries was off by two
   lines twice (comment blocks start earlier than the def lines), and
   one attempt left the original file DELETED while the splitter had
   produced nothing — recovered from git, then rewritten to find the
   boundaries programmatically (`find_def_end`) instead of counting.
   The lesson stands: carve structurally, never by line arithmetic.
2. **The daemon race ate every code commit's message** — all eight
   units landed, but the daemon's `chore:` sweep beat my `git commit`
   on the code halves six times; the narrative messages only ever
   carried the docs halves. The end state is complete and green, but
   the history reads mismatched; nothing to fix retroactively.

## e) WHAT WE SHOULD IMPROVE

1. **A blank-hash worktree is a main-red worktree.** The daemon
   commits continuously, so a vendorHash recipe's blank window must
   stay under ONE command; anything longer risks the sweep landing the
   placeholder state (today's ~35-min red). The recipe's step one
   should write to a scratch copy or run the whole fill in one
   `nix build --rebuild` round-trip.
2. **Adjudicate dedup groups BEFORE the baseline write, but verify the
   set-diff AFTER** — the +1/−3 composition check caught that the
   naive group count (110 → 112) misread the delta.
3. **Verify-external-claims before encoding provider error strings** —
   the FCM 403 subset rides Google's documented codes
   (`SENDER_ID_MISMATCH`, `THIRD_PARTY_PERMISSION_DENIED`), confirmed
   against the Firebase reference rather than guessed prose.

## f) NEXT (harvest inventory — routed to TODO_LIST homes)

| # | Task                                                                                                                                                                                | Home                                                             |
| - | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------- |
| 1 | Read the CI verdict on `6a3368c5` (and every later push) — the heal's green is the session's last open claim                                                                        | the standing CI-verdict rule (AGENTS)                            |
| 2 | §D10 file-size disposition → the v2.15.0 fold                                                                                                                                       | TODO_LIST § "BuildFlow file-size-check disposition"              |
| 3 | The deepening pick (five candidates)                                                                                                                                                | TODO_LIST § "Deepening brief candidates await the owner pick"    |
| 4 | Bump Trains 1–4 vs the release ordering                                                                                                                                             | TODO_LIST § "Dependency bump trains are ARMED but unexecuted"    |
| 5 | Fold the `--save-baseline` footgun into the docs-gates issue draft, then post (owner-gated)                                                                                         | TODO_LIST § "Fold the `--save-baseline` remainder-write footgun" |
| 6 | The remaining §f-50 items from the 15:25 handoff report (typed feed maps F7 tail, cadence pre-drain automation, depth-pin runway note, .templ depth ratchet, island file splits, …) | TODO_LIST + the 15:25 report's §f                                |

## g) Questions for the owner (carried, unchanged)

1. The §D10 file-size ruling (split-on-sight vs `max_file_size` cap vs
   warn-only) — the release gate.
2. The deepening pick.
3. Bump trains vs the release: which first?
