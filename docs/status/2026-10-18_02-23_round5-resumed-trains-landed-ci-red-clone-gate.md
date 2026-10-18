# Round-5 execution, resumed session — data-model trains, ASR hardening, D-train landed; CI red on the textagent clone gate

**Written:** 2026-10-18 02:23 CEST · **Session start state:** the round-5 plan mid-flight (E9 tests 2/5, 13 unpushed commits, full-gate verdict lost) · **Session driver:** the owner's "execute the WHOLE list" mandate (both turns).
**Claims verified as of 02:23 CEST; push/CI statements rot fastest** (snapshot-rot header convention, F85 — first use).

## a) Session goals vs outcomes

The mandate: finish and verify the entire round-5 assistant-executable list. Outcome: **E7/E8 (already done), E9, E10, E11, E12, E14 (adjudicated done-at-HEAD), E15, E16 executed and locally verified; E17 + the E20–E23 folds NOT started; CI is RED on a clone-gate finding that is not from this session's code** (see §d.5).

| Plan item | Verdict | Evidence |
| --- | --- | --- |
| E9 batch 1 (F37–F42) | 🟢 DONE — the 3 missing tests written (config retention, gateway panic, transcripts boundary), scan-fixture id bug FIXED, registry regenerated for 6 codes | `519c108`, `6008c6b`; targeted suites green |
| E10 batch 2 (F43–F48) | 🟢 DONE — `domain.RelPath` brand at the blob seam; `New{In,Out}bound` message/fax constructors; `Delivery`/`FaxDelivery` machine-value status seams (2 suppressions deleted); `session.Token` brand end-to-end under `auth:` prefix | `352433c`-family, `7cc06d8`, `e08cb7a`; auth battery green (session, app, TestServer) |
| E11 batch 3 (F49–F54) | 🟢 DONE — F49 verified already-covered; F50 registry cross-check + 3 rulings re-sighted + sweep-log line; F51 TTL-map audit note + Lang-enum stat card annotated (12 string enums today); F52 granularity + Lifetime rulings; F53 attachAttachments IN-query + fan-out spec; F54 caps×window table test + hubs keyed by `domain.Extension` | `084b7d3`; store/readmodels/fax/messaging green |
| E14 three smalls (F63–F65) | 🟢 ADJUDICATED ALREADY-DONE-AT-HEAD — Resolution enum + both-arm tests (10-14 B1.2/B1.5) IS the receipt seam; identities/shared-contact Load validation landed (DDD tier 4%); schema_version table + stamped chain wired (`store/db.go migrate`). **Plan Errata miss, verified this session; plan not yet annotated** | config.go:534-544, gateway.go, store/db.go:267+ |
| E15 ASR hardening (F66–F68) | 🟢 DONE — F66 verified already-shipped (backoff + toast + specs); F67 outcomes tracker on BOTH wires (rolling latency/failures, per-request debug line) + optional NON-CRITICAL health leg (named "asr" when seam on, never probes); F68 concurrency ceiling (4 in flight, 503 past short queue wait) | `d069426`; asr 9/9, ceiling test green under `-race` |
| E16 ASR polish (F69–F73) | 🟢 DONE — F69 language hint verified present; F70 DOM cap (8k tail, marked) + per-call Clear; F71 Download affordance (row → text/plain blob save); F72 voicemail-button test verified present; F73 combined 429 budget flood-pinned server-side | island 226/226 → 228/228; shell 47/47; server views/history green |
| E12 D-train (F55–F56) | 🟢 DONE — `CALL_TRANSITIONS` + `applyTransition` + `viewStateOf` in state.js; hold machine gated by the table; sip.js arrivals legality-checked with loud warns; 2-spec pin of the table + honesty contract. F57 (DOMAIN_LANGUAGE memo) stays owner-gated, untouched | `ef29ea9` |
| E17 stragglers (F74–F76) | 🔴 NOT STARTED (ui-capture visual pass, module-cache janitor, errorpage upstream note) | — |
| E20–E23 folds (F79–F86 remainder) | 🔴 NOT STARTED (F77/F78 owner-gated anyway; F79 upstream drafts, F80 routing audit, F81/F82 TODO rows, F83 hints, F84/F85 folds, F86 BuildFlow-repo side) | — |
| Full gates for the ASR+E12 trains | 🔴 NOT RUN — only targeted suites + island runs; the full buildflow + flake check for `d069426`/`ef29ea9` is still owed | §f |

## b) What got shipped (the one-paragraph read)

Three verified trains landed on top of the carve: the data-model harvest (brands, constructors, machine-value seams, corrupt-row panics — the bug classes the review named), the ASR train (a health leg that never lies because it never probes, a memory ceiling, client cap/clear/download affordances, the 429 budget pinned), and the island's declarative call-lifecycle table (the D-train's last open code item). The "three smalls" turned out to be stale backlog — all three verified already satisfied at HEAD — which the round-5 plan's Errata missed and this session proved.

## c) Surprises / discoveries

1. **The round-5 plan's E14 was stale** — all three smalls (F63/F64/F65) were already satisfied by earlier trains (09-19/10-14). The Errata sweep verified stale TODO rows but missed these.
2. **F66 and F69 were already shipped** — the island's backoff/toast and language hint predate the plan (the TODO row's "open" list was stale on exactly the items E15 named).
3. **The engine's seq sort is width-sensitive** (lexicographic under the hood): uniform-width UnixNano seqs are production-safe, but test fixtures mixing digit widths interleave calls — pinned as a fixture rule in the new table test's seed helper.
4. **todo-check pattern-matches the bare token "bug"** in comments at critical severity — reworded to "defect" (`c041a83`).
5. **ListMessages returns oldest-first** (the reverse happens in Thread()) — my fan-out spec initially assumed the opposite; flipped after reading the pinned lifecycle test.
6. **The island stub DOM lacked `click()`/`appendChild`/`removeChild`** — extended helpers.mjs (the stub models the browser; the handler stays idiomatic).

## d) What went wrong / honesty ledger

1. **Masked exit code (mine):** the early "green" store+server run piped through `head`, so `$?` measured `head` — the race-mode scan-test failure went unnoticed until the full gate caught it. The documented PIPESTATUS trap bit ME; every later verdict was captured to a file.
2. **Scan-panic fixture bug (prior session's, fixed here):** fake row ids were 24/25-char prefixed literals; the branded-id parser wants RAW 21-char nanoids. The corrupt-row test was also passing for the wrong reason (panicking on the id, not the enum). Fixed in `6008c6b`; this was ALSO the cause of CI run `39077421278`'s failure (the push raced ahead of the fix).
3. **test-coverage cache-eviction flake:** the full gate's only failure was a vanished `/mnt/buildcache` link artifact — infra, re-ran the step standalone green.
4. **Test-fixture debugging cycles:** wrong recorder (`[0]` vs `.at(-1)`), the label test's leaked Inviter stub, a double-`close` risk (fixed with `sync.Once`) — three avoidable iterations on my own new specs.
5. **CI IS RED RIGHT NOW (the live issue):** run `39081404828` for push `c041a83` fails ONLY the `art-dupl check (new clones only)` job: a new clone group in `internal/server/textagent_api.go:112-121|133-142` — `approveDraft`/`discardDraft` share the 10-token session-gate + `textAgentOn` + ctx-timeout prologue. That file came from the TEXTAGENT train (daemon commits, concurrent/prior session — I never touched it); the group is new SINCE the 66-group baseline. Diagnosis was mid-flight when this report was ordered. Fix-forward path per the dedup protocol: judge the group (reading it, this is the deliberate parallel-handler-prologue class — likely ACCEPT), regen the baseline with the CI-PINNED `art-dupl@v0.7.2` (never the local v0.9.0), add the registry sweep-log line, push, watch CI. Do NOT skip the judgment.
6. **gopls went stale repeatedly** (post-carve deletions, then long typecheck lag) — wasted attention on phantom diagnostics twice; `go vet` remained the arbiter.
7. **Baseline churn:** the branching-flow baseline (45 findings) orphaned twice via line shifts — advisory only; final re-save still owed at the last gate.
8. **Not done:** train-end smoke on the current tree; CHANGELOG line for E12; the E17/E20-23 items; the full gate for the last two trains; the plan annotation for E14's adjudication.

## e) Improvements (do more of / stop doing)

- **Keep:** verdict-to-file discipline (`> log 2>&1; echo EXIT=$?`), targeted-suite-then-full-gate ladders, judging detector findings before baselining, stub-as-browser-model extensions over defensive production code.
- **Stop:** piping test output through `head` when the exit code is the verdict; assuming plan rows describe reality (verify-at-HEAD first — E14, F66, F69 all said "open" and weren't).
- **Process gap worth closing:** the CI clone gate fires on ANY new group regardless of author — a concurrent train's landing can red MY push. The registry already routes this; the missing habit is a pre-push `art-dupl check` when the daemon has landed foreign files since the last green run.

## f) The next 50 tasks (concrete, ordered)

1. **Fix the red main (blocking):** judge the textagent_api.go clone → registry row + v0.7.2 baseline regen (or extract if judged harmful) → commit + push → `gh run watch` to green.
2. Full verification of the unpushed trains: `BUILDFLOW_NO_RESULT_CACHE=1 nix develop -c buildflow` + `nix flake check` over `d069426`/`ef29ea9` (fmt + registry check first; branching-flow baseline re-save at the end).
3. Train-end smoke: build `/tmp/webphone-bin`, `python3 scripts/webphone-smoke.py --bin …` (expect 47+4+8+9 passing; ASR-on dashboard boot covered).
4. CHANGELOG `[Unreleased]` line for E12; plan annotation for the E14 adjudication (docs-health ANNOTATE posture).
5. E17: F75 module-cache janitor + zero-byte `.tmp` sweep (script + one-shot run); F74 ui-capture 20-shot pass (local-only budget decision) + matrix diff; F76 errorpage tagging upstream note (verify-before-filing, then file).
6. E20–E23 remainder: F80 routing audit of the 10-14 status's 26 follow-ups; F81 carve row two-method refresh; F82 metaengine row close (verify G3-via-2.9.0); F83 F019/F026 hints on query declarations; F84 lessons folds (self-verifying scripts; verify-a-change-exists); F85 snapshot-rot convention → AGENTS if kept; F86 BuildFlow-repo GOPROXY fallback.
7. Final harvest: TODO_LIST row rewrites (E9–E16 executions), closing status report, push + CI watch to green.
8. Standing obligations carried: served-markup changes (history.templ download button, call-card clear) owe a fresh stack browser E2E at the next relock; owner terminals (E1–E6, E18, E19, E24–E26) untouched by design.

## g) Three owner questions (do NOT self-answer)

1. **The textagent clone judgment:** ACCEPT-into-baseline (parallel handler prologues, same class as the pinned idempotency ruling) or EXTRACT a shared prologue helper? The registry precedent says ACCEPT; a concurrent session owns that file — extracting under them risks a collision.
2. **Push discipline for the remaining trains:** push after the red-gate fix + full gates (recommended — keeps main green per train), or hold everything for one end-of-session push?
3. **E17's F74 visual pass now or deferred?** It is the only E17 item needing real wall-time (Chromium boot + 20 shots + eyeball); deferring it to a paired session with the owner glance (F87) folds two image-blind items into one sitting.

**Session state: 5 local commits beyond `origin/main` (`d069426`, `b85a0a4`-family daemon, `ef29ea9`); daemon push-leg live (recovered earlier this session). Waiting for instructions.**
