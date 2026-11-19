# Full code review, part 2 of 4 — telephony/messaging core — 2026-11-19

> SNAPSHOT-ROT HEADER: claims verified as of **2026-11-19 08:08 UTC** against tree `2f317894`.
> Line numbers and grep counts drift; re-verify before citing downstream.
> READ-ONLY REVIEW — no fixes applied (fixes are adjudicated centrally afterward).

Scope (PART 2 of 4): `internal/domain/**`, `internal/messaging/**`, `internal/fax/**`,
`internal/gateway/**`, `internal/pbx/**`, `internal/sipcred/**`, `internal/session/**`,
`internal/blob/**`, `internal/retention/**`, `internal/httpclient/**`, `internal/vcard/**`.
**75 files visited in full** (34 code + 41 test), ~9.9k LOC. All 11 scoped packages re-run
green inside `nix develop` (`go test -count=1`, GOEXPERIMENT leak neutralized):
domain, messaging, fax, gateway, pbx, sipcred, session, blob, retention, httpclient, vcard — `ok` each.

Cross-referenced, not re-litigated: the machine-tables doctrine (deliveryMoves/faxMoves/
AppendToThread/Notifier.Project — all four remain the ONE homes; no branch-by-branch
legality leaked into services), the typed-error-products pattern (ForwardError/formError/
ErrInvalidSend/ErrInvalidFax all present and family-pinned), owner-scoping (`*Any` seams
exist only in messaging for the admin surface, matching the documented contract), the
sessions-stay-SQL D2 ruling, byte-pinned gateway golden tests, and the dedup registry's
baselined COMPOSITION_mixin facts. Part 1 (server) owns `internal/server/notifier.go`'s
`Project` switch; this part verified the events it consumes are the closed set.

## 1. Per-package verdicts

| Package                               | Verdict                                                                                                                                                                                                                                                                                                                                              |
| ------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `internal/domain` (21 code + 12 test) | **Strong.** The zero-internal-imports kernel holds: machine tables are data, constructors make illegal rows unbuildable, branded ids/Phone/Extension/Address/RelPath close the type doors, `sanitizeDialable` matches the island regex with pinned tests both sides. Findings M2, L1, L2 below (catalog gap, prefix-blind parseID, dead predicates). |
| `internal/messaging`                  | **Strong.** Sender registry, self-send fast path, compensating blob rollback, race probe accepting both legal first-sight interleavings — exemplary. No new findings beyond the shared L4/L6 below. Prior F1 (orphaned blobs) RESOLVED, see §3.                                                                                                      |
| `internal/fax`                        | **Good with one real gap.** Lifecycle via faxMoves, receipt-Resolution contract (no type sniffing), panic-contained detached archiver, compensating unspool. M1 (Resend/retention shared-path) is the package's one open defect.                                                                                                                     |
| `internal/gateway`                    | **Strong.** Byte-pinned form goldens (field order + full header block + honest Content-Type), typed ForwardError/formError, embedded-Telnyx-status family rule with a war-story comment. L6 buffering nit.                                                                                                                                           |
| `internal/pbx`                        | **Strong.** `do()` chokepoint owns disabled-policy + auth + family split; nil-receiver safety pinned. L7 (second inline path/query join) nit.                                                                                                                                                                                                        |
| `internal/sipcred`                    | **Clean.** 24 lines, cross-repo wire vector pinned — nothing to add.                                                                                                                                                                                                                                                                                 |
| `internal/session`                    | **Strong.** Sliding-TTL decision table exhaustively pinned (throttle, cap, dead rows), Mem/SQLite parity via shared makeSession + seam pin, cookie Max-Age parity tested. No findings.                                                                                                                                                               |
| `internal/blob`                       | **Strong.** RelPath brand is the escape guard; atomic write pinned incl. no `.tmp` residue. L4 ordering nit.                                                                                                                                                                                                                                         |
| `internal/retention`                  | **Good.** Deterministic sweepOnce seam, corrupt-path containment, disabled-default pin. M1's unlink side lives here/its sweeper seam.                                                                                                                                                                                                                |
| `internal/httpclient`                 | **Clean.** The one-home isolated-client rule with the flake rationale in the package comment.                                                                                                                                                                                                                                                        |
| `internal/vcard`                      | **Good.** Tolerant parser (unfolding, quoted params, escapes) with round-trip pins. L5 (Apple group-prefix shape) gap.                                                                                                                                                                                                                               |

## 2. Findings

### Medium

**M1 — Fax Resend shares a DocumentPath with its original job; retention unlinks it under the live retry.**
`internal/fax/service.go:215` (retry reuses `job.DocumentPath`), `internal/cqrs/faxesstore.go:264-265`
(sweep emits every expired row's `DocumentPath`, no reference check), `internal/retention/retention.go:158-172` (unlink loop).
Why: the original failed fax (created T0) and its resend (created T1 > T0) reference the SAME blob file.
When the retention cutoff crosses T0 but not T1, the sweep expunges the original, returns the shared path, and
retention unlinks the file — the still-live retry row now points at nothing: its `Document()` stream
(`fax/service.go:327`) fails with `blob.open` for a row the fax list still renders, and a further Resend
refuses at the reopen probe (`fax/service.go:208-212` — the probe's comment shows the author knows documents
can vanish; the probe makes the failure honest, not impossible). No test pins the interaction (retention_test
never builds a shared-path pair). Why it matters: a live row losing its document is a data-integrity break in
the only storage truth, even if the blast radius is one row's download for a (T1−T0)-wide window.
Fix (pick one): copy the blob on Resend (each row owns its path — simplest, one extra disk copy per retry),
or make `Faxes.SweepBefore` skip paths still referenced by non-expired rows. Add the retention-test pin either way.

**M2 — The event catalog is missing `ReviewSaved`: 4 of the 5 closed seam events are documented.**
`internal/domain/catalog.go:28-51` (mirror types) and `:69-92` (event rows); `internal/domain/events.go:54-63`
(five `isDomainEvent` implementers); `internal/domain/catalog_test.go:75-80` (wantEventIDs) and `:157-165`
(parity pairs).
Why: `BuildCatalog`'s own doc says it documents "the closed event set", and `webphone catalog` feeds the
federation hub — but `ReviewSaved` (added with issue #8, live since 2026-10-28: emitted at
`internal/server/api/review_api.go:171`, projected at `internal/server/notifier.go:77`) never joined it.
`TestCatalogSchemaParity` cannot catch this: its pairs list is hand-maintained, so the two enumerations
(the Event interface implementers vs the catalog) are a drifting split brain — exactly the shape the
one-home doctrine forbids. Impact is documentation/governance, not runtime.
Fix: add `CatalogReviewSavedSchema{Owner, Call string}` + the `review.saved` event row + the parity pair;
better, derive the parity pairs from an exhaustive implementer list so the next event cannot silently miss.

### Low

**L1 — `parseID` strips any `prefix:` without checking it is THIS brand's prefix.**
`internal/domain/ids.go:254-263`.
Why: `strings.Cut(s, ":")` discards the brand and only length (21) is validated, so `"Fax:<21 nanoid chars>"`
parses happily as a ThreadID — the branded-id system's cross-type confusion guard silently accepts
cross-brand strings in their prefixed form. `TestBrandedIDParsersRejectJunk`'s "wrong brand for type" case
(`Contact:abc`, ids_parse_test.go:78) passes only because the body is 3 chars — the test gives the illusion
of brand checking that the code doesn't do. Fix: when a prefix is present, require it to equal the brand's
`Name()` (the rebrand stays legal only for the raw form).

**L2 — Dead/duplicated zero-predicates on the branded value types.**
`internal/domain/relpath.go:60-63` (`Empty()` AND `IsZero()` with identical bodies — all callers use `Empty()`:
fax_templ/admin_templ), `internal/domain/network.go:135` (`Address.IsZero`), `internal/domain/ids.go:83`
(`Phone.IsZero`), `internal/session/service.go:59` (`Token.IsZero`).
Why: repo-wide grep finds zero callers for these four (only `Extension.IsZero` is live, via `authmap.Line`).
Two names for one fact (`RelPath`) violates the one-home rule; the others are dead API. Fix: delete the
uncalled predicates (or keep exactly one name per type); callers compare through the live one.

**L3 — Inbound fax bytes are uncapped at the service seam while Send caps at 20 MiB.**
`internal/fax/service.go:227-229` (Receive checks only non-empty) vs `:111-114` (Send enforces
`MaxPDFSize` 20 MiB + `%PDF-` signature).
Why: the only bound on an inbound document is the hooks lane's 40 MiB body cap
(`internal/server/hooks/webhooks.go:426`), i.e. a webhook-secret holder can spool roughly double the
send cap per fax, and non-PDF bytes are stored (and later served) without the signature check Send
applies. Deliberate trust of the machine lane is plausible — but nothing at the site says so.
Fix: enforce `MaxPDFSize` (and optionally the signature) in `Receive`, or write the one-line trust
justification at the site per the error-path owner rule.

**L4 — `blob.Save` writes the file BEFORE branding the path it returns.**
`internal/blob/store.go:55-63`.
Why: `atomicwrite.WriteWithPerm` runs at line 55; `ParseRelPath` at 59. A malformed `ext` (the only way
branding fails) would leave the just-written file orphaned with no path in the error to roll back —
the rollbackSaved callers only know paths Save RETURNED. The comment ("a malformed extension fails here
instead of minting a path no later Open would accept") describes the brand refusal but not the residue.
Unreachable today (both call sites pass closed-set literals: `extensionOf`'s allowlist and `".pdf"`),
so this is defense-in-depth, not a live leak. Fix: build+parse the rel path before the atomic write.

**L5 — `vcard.Decode` ignores group-prefixed properties (`item1.TEL:…`) — the shape Apple exports.**
`internal/vcard/vcard.go:120-138`.
Why: `splitProperty` uppercases the head up to the first `;` but never strips a `group.` prefix, so
`item1.TEL;type=CELL:+1…` reads as property name `ITEM1.TEL`, never matches `TEL`, and the card
(a numberless card is skipped, `vcard.go:36`) is dropped entirely. Apple Contacts/Address Book — the
most likely import source for a personal phone app — prefixes TEL/EMAIL/FN with `itemN.` groups.
The pinned fixtures (vcard_test.go) cover quoted params and folding but not this shape.
Fix: in `splitProperty`, `strings.Cut(head, ".")` once before the parameter cut (group names never
contain `;`), and pin an Apple-shaped fixture.

**L6 — Provider forms fully buffer binary payloads into a `strings.Builder`.**
`internal/gateway/webhook.go:172-189` (`providerForm`) and `:209-232` (`messageForm`).
Why: a 5-attachment MMS (5×10 MiB) or a 20 MiB fax PDF is assembled as one string — peak ~50-70 MiB
per send, and binary-in-`string` is a smell where `bytes.Buffer` is the idiomatic zero-copy-adjacent
choice. Correctness is unaffected (Go strings hold arbitrary bytes; `multipart.Writer` takes an
`io.Writer`). Fix opportunistically: swap to `bytes.Buffer`/`bytes.Reader` next time the file is
touched — do not churn the byte-pinned goldens for it alone.

**L7 — The "join path and query separately" rule has two homes in `pbx.Client`.**
`internal/pbx/client.go:68-79` (`UpstreamURL`, the documented named home) vs `:206-213` (`do()`'s inline
`strings.Cut(path, "?")` + re-attach).
Why: AGENTS pins this exact lesson (`url.JoinPath` percent-encodes `?`); `do()` re-implements what
`UpstreamURL` already names instead of calling it (`c.UpstreamURL(pathOnly, query)`). Two sites that
must agree on an escaping subtlety = drift risk on precisely the class the rule exists for.
Fix: `do()` delegates to `UpstreamURL`.

### Info / nits (no action required, recorded for completeness)

- `DeliveryStep.Event()` (`internal/domain/threadfacts.go:158-167`) default-maps ANY unknown `Kind` to a
  verdict event — trusted-journal reasoning is documented at the type; a default refusal would be stronger
  if the journal ever gains an untrusted producer.
- `ContactDeleted`/`SnippetDeleted`/`ThreadDeleted`/`CallNoteDeleted` pin `json:"Key"` while sibling facts
  ride default field names — documented "byte-neutral now, rename-safe later"; consistent as a pattern.
- `Loopback` ProviderRef = `prefix-<UnixNano>` (`internal/gateway/gateway.go:134-148`): two calls in the
  same nanosecond would share a ref; the loopback lane never resolves by ref, so harmless today.
- `MemStore.Get` reads `s.now()` after `RUnlock` (`internal/session/service.go:223`) — benign (clock is a
  func field, no shared state); noted only because the lock discipline is otherwise impeccable.
- `messaging.selfSendRejection`'s `ParsePhone` error branch (`internal/messaging/service.go:418-421`) is
  dead by construction (an SMS `Address` already carries a ≤32-dialable value under the same rule) — the
  inline comment says exactly this; fine as defensive code.
- The `json:"Key"`-tagged deletes vs untagged facts (above) and `rollbackSaved` (messaging) vs `unspool`
  (fax) as two names for one compensating-delete concept are both registry-adjacent accepts — the fax
  service's baselined COMPOSITION_mixin comment owns the symmetry ruling; not re-litigated.

## 3. Prior-series verdicts

| Prior finding                                                                             | Verdict this sweep | Evidence                                                                                                                                                                                                                                                                                                                                                                                                           |
| ----------------------------------------------------------------------------------------- | ------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| 2026-09-23 F1 (re-confirmed open 2026-11-11): orphaned blob files on send/persist failure | **RESOLVED**       | `messaging.rollbackSaved` (`internal/messaging/service.go:244-250`, both Send and Receive paths call it on every pre-persist failure) and `fax.unspool` (`internal/fax/service.go:137-141`); pinned by `TestFailedSendLeavesNoOrphanedBlobs`, `TestFailedReceiveLeavesNoOrphanedBlobs` (messaging/blobcomp_test.go) and the fax twins (fax/blobcomp_test.go) — zero-files-survive assertions over a refusing gate. |

## 4. Split-brain candidates (for the central adjudication)

1. **M2** — catalog event enumeration vs the closed `Event` implementer set (live drift found).
2. **L7** — two path/query-join sites inside `pbx.Client`.
3. **L2** — `Empty()`/`IsZero()` duplicate predicate per branded type.
4. (accepted, recorded) `rollbackSaved`/`unspool` naming pair — baselined seam-parallel accept.

## 5. Strengths worth preserving

- The machine tables really are the one home: `messaging.Send`/`fax.drive`/`UpdateProviderStatus` contain
  zero status literals outside the table's vocabulary, and every inert case is documented + pinned
  (`TestFaxAdvanceInertMoves`, `TestDeliveryAdvanceSendEventsFromSettledStatesAreInert`).
- Constructor discipline (`NewOutboundMessage`/`NewInboundMessage`/`NewOutboundFaxJob`/`NewInboundFaxJob`):
  direction/channel/status are consequences, so the illegal rows are unbuildable — and
  `constructors_test.go` pins exactly that sentence.
- The compensating-delete tests assert on the FILESYSTEM (walk the blob root, count zero), not on mock
  calls — the right grain for the invariant that matters.
- `raceprobe_test.go` is a model concurrency pin: both legal interleavings accepted, the refused lane's
  recoverability re-driven, and the fold-settled read polled instead of raced.
- Error-family posture is total across the scope: every package's family_test.go pins codes AND families
  through the wrap chains the handlers actually add.

## 6. Forward-looking work (for TODO_LIST harvest, not this snapshot)

- M1 fax Resend/retention shared-path (with its missing test pin).
- M2 catalog exhaustiveness (mechanism, not just the one missing row).
- L5 Apple-shaped vCard fixture once the group-prefix strip lands.
