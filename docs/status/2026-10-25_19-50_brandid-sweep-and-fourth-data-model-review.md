# Status: brandid-lint BD003 sweep + fourth data-model review train

Snapshot header (push/CI claims rot fastest): claims verified as of 2026-10-25 ~19:50 CET. Session: single Crush session, webphone repo, base commit c05f16b.

## a) FULLY DONE

1. **brandid-lint `-rules all` sweep: 20 findings to 0 (exit 0, re-verified twice).**
   - Root-caused the two BD002s from linter source (`~/projects/go-branded-id/linter/primitives.go:334`): a directive only covers when its SINGLE physical line is the last line of its comment group, immediately above the target line or the enclosing declaration. Both existing directives (`internal/domain/push.go:20`, `internal/server/api/push_api.go:34`) violated this: the first sat above the wrong type (`PushDeviceRegistered` instead of `PushDeviceKey`), the second was a multi-line directive whose prefixed line lost `LastInGroup`.
   - 17 BD003 sites suppressed with single-line, doctrine-consistent directives (`<category>: <reason> (strong-id ... ruling)`), zero mechanical brands minted (AGENTS.md BD003 ruling honored).
   - Files (12): config.go, domain/push.go, matrix/client.go, matrix/receiver.go, push/fcm.go, push/push.go, readmodels/push.go, server/api/push_api.go, server/api/transcripts_api.go, server/hooks/ring.go, signal/client.go, sso/sso.go. Daemon-committed as 8041d27; diff verified 100% comment lines.
   - gofmt: sso.go struct-alignment fix applied; `nix fmt` (treefmt) green, 0 changed repo-wide.
   - Tests: all 9 touched packages pass (`go test -count=1`): config, domain, matrix, push, readmodels, server/api, server/hooks, signal, sso.
2. **Fourth data-model review written and committed**: `docs/reviews/2026-10-25_data-model-review.html` (from the html-report-kit template, tag-balanced, no placeholders, scrollspy intact). All 18 branching-flow type-safety findings triaged; 10-14 roadmap audited 6/7 landed with per-item source evidence; 4 adopted designs (push.Transport, sso.Finish, FaxJob.DocumentPath as RelPath, asr/matrix wiring structs + machineSurface bool-last); edge-door/seam-interface class baseline-ruled with reasons.
3. **Memory updated**: AGENTS.md BD003 sentence refreshed to 2026-10-25 with the single-line-directive lesson; new type-safety-triage bullet with the open roadmap and baseline rulings.

## b) PARTIALLY DONE

1. **10-14 roadmap item 7** (prior review): asr.Provider is now an interface, but provider selection is still the string compare `cfg.ASR.Provider == "whispercpp"` (app.go:202); transcript-remote and hub-map-keys sub-items unverified/carried. Logged as the new review's carry-forward.
2. **The new roadmap** (Transport/Finish/RelPath/wiring structs/bool-last): DESIGNED and ruled in the review, NOT implemented (auth surfaces + schema-adjacent changes deserve their own green-lit trains; the review is the decision artifact).

## c) NOT STARTED

1. Implementing any of the 7 roadmap items of the 2026-10-25 review.
2. Recording the edge-door/seam-interface class in the branching-flow baseline (`.buildflow-baseline.json` regeneration must use the CI-pinned detector per AGENTS.md).
3. CI verification of the 3+ unpushed daemon commits (see d2/f).

## d) TOTALLY FUCKED UP (honesty section)

1. **Nothing destroyed.** All edits are comment-only or docs. But two process stumbles, both caught and fixed:
   - The first template splice left the old placeholder body in front of the report content (found via placeholder grep, removed, re-verified).
   - The daemon committed mid-edit twice (8041d27, 26847bc, 8b65720), which tripped the edit tool's staleness guard once; resolved by re-reading before editing.
2. **Pre-existing red, NOT this session's and NOT touched**: `internal/server/passkey_api_test.go` and `internal/userauth/userauth_test.go` fail to compile (pkStub/stubWebAuthn vs the new `FinishLogin` signature). This is the documented cqrs-htmx round-3 identity-train stub adaptation, outstanding since e27093d per AGENTS.md. It does not affect the packages this session touched.
3. **Push state**: 3+ daemon commits sit unpushed (8041d27 19:35 → 8b65720 19:49). The manual-push bar (60 min from the FIRST unpushed commit's timestamp) was NOT met at snapshot time; `scripts/push-stall-watch.sh` is the arbiter. No CI verdict exists yet for these commits; check `gh run list` after the push lands.

## e) WHAT WE SHOULD IMPROVE

1. Verify after each batch, not after the whole batch (the brandid re-run came two hours after the first edit).
2. The directive-placement contract deserves a test: a tiny repo drift-test asserting every `//brandid-lint:ignore(BD003)` line is the last line of its comment group directly above an ID declaration would have caught both BD002s at write time.
3. Copying the report template: splice by content anchors, not by end-of-file marker (the placeholder-body repeat).
4. The branching-flow type-safety family has no per-repo baseline story yet; the review's item 6 is the fix, but the baseline IDs are host-absolute (upstream #13), which makes scripted regeneration brittle.

## f) NEXT (prioritized; the review's roadmap is items 1-7)

1. Implement `push.Transport` (defined type, typed consts, ParseTransport, re-typed Fanout registry/API, push family tests).
2. Implement `sso.Finish` value object (auth: prefix + session-behavior suite).
3. Type `FaxJob.DocumentPath` as domain.RelPath (scan door converts; golden rows re-verified).
4. asr.Wire + matrix.Wire structs; app wiring updated.
5. machineSurface bool-last reorder.
6. Record the branching-flow baseline class (edge doors + six Deps seams) with per-site reasons.
7. Finish 10-14 item 7: ASR.Provider parsed enum at the next config touch; transcript remote + hub map keys.
8. Fix the identity-train stubs (pkStub/stubWebAuthn FinishLogin signatures) + vendorHash re-pin + push + CI-watch (unblocks the full suite).
9. After the daemon push lands: `gh run list` to green, per the check-CI-after-every-push rule.
10. If push stalls > 60 min from 19:35 (8041d27): invoke the manual-push bar, state the anchor, push, watch CI.
11. Drift-test for directive placement (e.2) as an arch_test addition.
12. Re-run `buildflow` full at next train start (AGENTS.md doctrine) to re-baseline tier timing.
13. The duplication/structure families (mixin suggestions, DUPLICATE_TYPE trio config.CRM/Paperless/PushNtfy, vcard.Card) belong to a dedup-registry sweep, not this review; schedule one.
14. Consider wiring `cqrshtmx.ProjectionReadinessCheck` + `ServiceConfig.OnProjectionFailed` (round-3 audit gap 1, already pinned).
15. Wire `ExtensionHubs.Shutdown` so restarts stop burning the 30s drain budget (gap 3).
16. Bounded AuditLog projection decision (stale second reason in AGENTS.md since usermgmt v4.15.0).
17. webphone-android P1/P2 continuation per the 2026-10-23 owner GO (outside this repo).
18. Quarterly art-dupl baseline audit (due per the 2026-10-15 pin note; next window open).
19. Re-measure branching-flow tiers 1+2 (next due 2026-11-05).
20. matrix_test.go S1038 (fmt.Fprint(fmt.Sprintf)); signal/client.go unused ctx param; signal_test.go scanner.Err() + unused dialer — lint-hygiene sweep candidates.

## g) QUESTIONS ONLY THE OWNER CAN ANSWER

1. **Identity-train stubs**: is a concurrent session actively adapting pkStub/stubWebAuthn to the new `FinishLogin` signature, or should this session's successor take it (it blocks `go test ./internal/server` and `./internal/userauth`, and therefore full-buildflow)?
2. **Roadmap green-light**: implement review items 1-5 as one train now, or hold the review as the decision artifact until the next natural touch of each seam (push/auth/fax trains)?
3. **Detector policy**: should the edge-door/seam-interface class be baselined silently (buildflow baseline) or surfaced as an AGENTS.md ruling that branching-flow findings of this class are expected noise until the tool learns test-side implementers?
