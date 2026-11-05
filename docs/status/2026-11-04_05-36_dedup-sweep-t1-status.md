# Dedup sweep — 2026-11-04 05:36 (t1 gate + v0.9.1 t3 cross-check)

## What ran

1. **Owner-pasted report**: host `art-dupl v0.9.1-0.20261017225421` at `-t 3 --sarif`
   (`~/.crush/shell-output/output-3792907013.log`). Per protocol, a paste is
   evidence of ITS run, not of HEAD — everything below was re-verified at HEAD.
2. **Ratified re-run**: CI-pinned `art-dupl@v0.7.2` (built from the tag into
   `/tmp/artdupl072`) at the gate's exact flags:
   `check . -t 1 --baseline-path .art-dupl-baseline.json` →
   **"Found total 0 clone groups" vs the 86-entry baseline (recorded 2026-10-29).**

## Verdict

**Zero harmful duplication, zero new groups.** No code changed; no baseline
re-record (nothing to bless). Registry upkeep only.

## v0.9.1 `-t 3` report: 29 distinct groups, all attributed

The two detectors speak different group universes (the documented version-skew
trap, 10-15 ruling). Groups in the paste, ranked by fragment size:

| Fragment size                                                                        | Shape                                                                                                                                                                              | Ruling                                                                                                                                                       |
| ------------------------------------------------------------------------------------ | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| 43-tok ×43 (`f49b5c35…`)                                                             | Lone struct-field declarations (`email string`, `raw string`, …) across 43 sites                                                                                                   | ACCEPT — field-shape similarity is architectural ROLE, not duplication (the 10-05 mixin doctrine class)                                                      |
| 18-tok ×9 (`4e7ec54f…`)                                                              | `lang Lang) string` i18n-helper signature tails                                                                                                                                    | ACCEPT — signature-shape noise                                                                                                                               |
| 14-tok ×7 (`33cf19fd…`)                                                              | `… string) string` signature tails                                                                                                                                                 | ACCEPT — signature-shape noise                                                                                                                               |
| 12-tok ×3 (`abccd11e…`)                                                              | Multi-error join `Error()` body ×3 (`bootJoinError` / `ShutdownError` / `notFoundError`)                                                                                           | ACCEPT — see below; registry row added                                                                                                                       |
| 12-tok ×12 (`4202ab8d…`)                                                             | `lang Lang` parameter declarations                                                                                                                                                 | ACCEPT — signature-shape noise                                                                                                                               |
| 8-tok ×2 (`0e73909c…`)                                                               | matrix/receiver.go:165+181 — two `chatloop.Backoff` retry arms (cursor load vs sync)                                                                                               | ACCEPT — call sites of the 10-23 extraction; call sites ARE the dedup                                                                                        |
| 8-tok ×2 (`5f0d2d55…`)                                                               | callnotes.go:112+142 mutex-guarded store-op prologues                                                                                                                              | ACCEPT — standing 10-27 evening ruling                                                                                                                       |
| 6-tok pairs (`d773c476`, `3b9c7093`, `a2cb19cb`, `f3c57ca5`, `b2eace62`, `9eddfa7d`) | `providerRoundTrip` call pair; `mustMarshalRing`/`mustMarshalDraft`; fcm Send/SendDraft + Fanout Send/SendDraft prologues; Notes.Get/Reviews.Get; idempotency Seen/Record prologue | ACCEPT — every one a standing registry row (10-18, 10-27, 10-15, 09-18)                                                                                      |
| 6-tok ×2 (`5468c76e…`)                                                               | matrix:184 + signal:73 — `chatloop.Backoff` call + backoff reset across receivers                                                                                                  | ACCEPT — 10-23 chatloop ruling: receivers keep wire-specific code                                                                                            |
| 4-tok ×4 (`ae65c9c4…`)                                                               | history.templ `cdrMissed`/`cdrDirGlyph`/`cdrTarget`/`CDRDialTarget` — four domain predicates sharing a `Context == "public"` conditional                                           | ACCEPT — different logic behind a similar shape                                                                                                              |
| 4-tok ×4 (`c127597b…`)                                                               | fax.templ `faxSubtitleKey`/`faxUnavailableKey`/`faxEmptyKey` + settings.templ `settingsFaxKey` — capability→copy mappers over the SAME `FaxCapability` flags                       | ACCEPT — each mapper owns a different copy vocabulary; this is the 10-30 fax-decision doctrine (one flag conversion, per-surface copy), not duplicated logic |
| 4-tok ×2 (`f3d618d9`, `38cb8981`, `8b639a8e`)                                        | `preview` vs `vmCaller`; `attachAttachments` guard vs `firstUnreadIndex` guard; `formatFor` vs `flagLabel`                                                                         | ACCEPT — coincidental guard/signature shapes across unrelated seams                                                                                          |
| 3-tok trios (`0b836f4e`, `146a8713`, `0d08b48c`)                                     | signature/guard coincidences (fax service+test+templ; latency guard vs test poller vs `flagURL`; `agentDetail` vs `FaxCapability.Any` vs `settingsItems`)                          | ACCEPT — coincidence class                                                                                                                                   |
| 2-tok tails (7 groups)                                                               | Field/signature/type-name declarations (`t time.Time`, `FaxEvent`, `*textagent.Reviews`, `[]domain.Snippet`, mapper signature lines)                                               | ACCEPT — pure declaration noise                                                                                                                              |

## The one real logic clone: multi-error join ×3

`cmd/webphone/main.go:238` (`bootJoinError`), `internal/app/app.go:640`
(`ShutdownError`), `internal/readmodels/notfound.go:16` (`notFoundError`)
carry the identical 12-line `Error()` body. This was **deliberate**: the
2026-10-27 erraudit-max-zero train hand-rolled all three typed aggregates to
replace `errors.Join` (erraudit's `generic_return` rule + family
classification need typed products with `Unwrap() []error`), and its status
doc named the extraction trigger — a shared format helper becomes a
dedup-registry entry "IF a third appears" beyond these three.

Trigger check today: `grep WriteByte('\n')` + `Unwrap() []error` over
`internal/` + `cmd/` (non-test) → exactly the same three files. **Not fired.**
The ruling now lives as a standing registry row (the point-in-time status doc
was the only previous home).

## Registry upkeep (docs-only, no code)

- Merged the duplicated rail-row ruling: the standing-rulings row (line 63)
  and the broken 3-line table fragment ("Rail text block vs thread row text
  block", previously below the extracted table) are now ONE row carrying both
  fragments and the 17:4x hash-move note.
- Added the multi-error-join ruling row with its grep-able trigger.

## Detector-class note (for the next sweep)

The v0.9.x universe keys heavily on declaration/signature spans (lone field
lines, `lang Lang` params, `… string) string` tails) — 24 of today's 29
groups. This is the short-span matching class the 10-29 ruling flagged for
upstream filing if it lands in the PINNED detector's CI view. It has not
(the pinned gate is clean); today's sighting is recorded as context only.
