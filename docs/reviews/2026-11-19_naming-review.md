# Naming review — 2026-11-19

> SNAPSHOT-ROT HEADER: claims verified as of **2026-11-19 06:49 UTC** against tree `1a4c4786`.
> Tool versions and grep counts drift; re-verify before citing counts downstream.
> READ-ONLY REVIEW — no renames executed (renames are adjudicated centrally; wire
> contracts and pins ride these names).

Scope: Go identifiers across `internal/**` + `cmd/**` (types, functions, fields,
variables), plus the island JS exported module surface
(`internal/web/assets/island/app/`). OUT OF SCOPE per the review contract: DOM
ids / `data-*` attributes (wire contract), error-code stanzas (`<seam>.<op>`),
wire DTO field names, config keys (`koanf` tags), vendored and generated code,
`*_templ.go` regenerated artifacts (findings there target the `.templ` source).

Cross-referenced prior reviews (not re-litigated): the 2026-09-19 naming review's
acceptances (`Deps`, `handlers`, `T(lang,key)`, `sanitizeDialable`, `provider`,
cookie names); the 2026-10-25 data-model review's OPEN ROADMAP (`machineSurface`
bool-last, `push.Transport`, `sso.Finish` value object, `FaxJob.DocumentPath`)
— same items stay on that roadmap, not duplicated here.

## 1. Automated detection

| Tool                                                                                                             | Invocation                 | Result                                                                                                                                                                                                                                                                                                                                                                                                                                       |
| ---------------------------------------------------------------------------------------------------------------- | -------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `naming-smells.sh` (skill script, `--lang go`)                                                                   | full repo                  | **0 hits** in every category (vague type names, Manager/Handler/Processor/Helper/Util, Impl/Abstract/Base, boolean non-questions, Hungarian, number suffixes). Second clean run in a row (2026-09-19 was also 0).                                                                                                                                                                                                                            |
| `golangci-lint` (repo devshell, v2 config: `revive` + `staticcheck`, no repo config exists so both run defaults) | `./internal/... ./cmd/...` | 50 findings, all revive. **Naming-relevant: 9** — 4× `unexported-return` (store threadSummary family), 3× `redefines-builtin-id` (`max` ×2 files, `real` ×1 test), 2× `var-naming` (`extensionIds` ×2 in one test). The other 41 are comment-convention gaps (28 `exported` missing-doc, 13 `unused-parameter`, 3 `dot-imports`, 2 `package-comments`) — style, not naming semantics; BuildFlow does not gate revive, so these are advisory. |

## 2. Findings

Severity per the skill: 🔴 honesty (name lies), 🟠 clarity (name carries no
meaning), 🟡 domain alignment, 🔵 consistency.

| #  | Sev | Identifier                                                                               | Location                                                                                                        | The lie / domain mismatch                                                                                                                                                                                                                                                                                                                                                      | Proposed name                                                                                                          | Blast radius of rename                                                                                                                                                                                              |
| -- | --- | ---------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ | ---------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 1  | 🔴  | `FindThread`                                                                             | `internal/cqrs/threadsstore.go:146`, `internal/store/messages.go:432`, gate `internal/readmodels/threads.go:93` | "Find" promises a pure query. The **store** impl CREATES (persists an empty row "when absent"); the **cqrs** impl mints a fresh id for unknown addresses. One interface name, two contradictory side-effect contracts — and neither is "find". Both doc comments say "resolve".                                                                                                | `ResolveThreadID` (verb already used by the implementations' own docs and the identity seam's "resolution" vocabulary) | The `ThreadsGate`/`Messages` interfaces (2 sites), both implementations, `messaging/service.go:132,207` call sites, plus their tests. No wire bytes (internal only).                                                |
| 2  | 🟡  | `views.VoicemailItem`                                                                    | `internal/web/views/types.go:23`                                                                                | Split-brain: the same concept (one stored voicemail row) is `pbx.VoicemailMessage` at the seam and `views.VoicemailItem` one converter later. DOMAIN_LANGUAGE says "**voicemail summary / messages**"; the summary half keeps its name through the same converter, only the message half is renamed. "Item" is the checklist's vague noun.                                     | `views.VoicemailMessage` (pairs with the unchanged `views.VoicemailSummary`)                                           | `views/types.go`, `voicemail_templ.go` signature + `vmCaller` (via `.templ` + regen), `server/panels.go` (`voicemailItems` converter, `crmNumbers` closure), view tests. No wire bytes (view shape, rendered only). |
| 3  | 🟡  | `authmap.Map`                                                                            | `internal/authmap/authmap.go:92`                                                                                | "Map" names the data structure, not the domain concept. DOMAIN_LANGUAGE: "**ownership** — the config join of humans to lines"; this type IS that join. All four consumers spell the field `mapped *authmap.Map` — `authmap.Map` says nothing twice.                                                                                                                            | `authmap.Ownership` (fields `ownership` / keep `app.mappedUsers`, the ratified one-home join)                          | `authmap.go` + `admin/admin.go`, `sso/sso.go`, `userauth/userauth.go`, `app/app.go` (5 files, field renames included) + package tests. No wire bytes.                                                               |
| 4  | 🟡  | `Shell` interface ×2                                                                     | `internal/server/auth/auth.go:88`, `internal/server/platform/platform.go:70`                                    | Collides with the RATIFIED domain word: DOMAIN_LANGUAGE defines **shell** as "the page frame around both: nav, badges, toasts, error banner". These are server-side wiring seams ("the narrow wiring seam the carved handlers ride on"), and there are two of them under one name — split-brain plus word theft. A newcomer reading `platform.Shell` expects shell.js's frame. | `auth.Server` / `platform.Server` (consumer-defined "server seam" idiom), or a shared `wiring.Seam` if the twins merge | 2 interface declarations, the `*handlers` implementations in `internal/server`, every carved handler's field/type reference, test stubs. In-package + server, no wire bytes.                                        |
| 5  | 🟠  | `MastBody`, `MastDir`, `MastChan`, `MastAttach`                                          | `internal/cqrs/threadsstate.go:28-31`                                                                           | Cryptic abbreviation that is not universal (comment defines them as "last-by-created-at" facts) — and the SAME fields are spelled `LastBody/LastDirection/LastChannel/LastAttachment` in `domain.ThreadSummaryImage`, `store`, `readmodels`, and `views` (4 sites). Two vocabularies for one concept inside one pipeline.                                                      | `LastBody`, `LastDir`, `LastChan`, `LastAttach` (state struct is unexported; short forms fine)                         | `internal/cqrs/threadsstate.go` only (4 fields, ~5 sites: `summaryImage`, apply/adopt twins at :219/:233) + cqrs replay tests. Smallest blast radius in this list.                                                  |
| 6  | 🔵  | `CDROff bool`                                                                            | `internal/web/views/admin_templ.go:51` (source `.templ`)                                                        | Negative-polarity boolean — "Off" reads as a verb phrase about the reader, and every consumer must negate mentally (`if !CDROff`).                                                                                                                                                                                                                                             | `CDRsAvailable` (or `PhoneAPIOn` to mirror the seam's own config word)                                                 | Admin `.templ` props + `internal/server` admin handler fill + admin view tests.                                                                                                                                     |
| 7  | 🔵  | `ChatMatrix`, `ChatSignal`, `LiveContext` (bool fields)                                  | `internal/web/views/settings_templ.go:71`, `layout_templ.go:122`                                                | Noun-shaped booleans don't read as yes/no questions ("is the chat matrix?"). Documented, so mild — but the question form is the house rule (`HasMore`, `OverBudget`).                                                                                                                                                                                                          | `MatrixOn`/`SignalOn` (or `HasMatrixLane`…), `ShowLiveContext`                                                         | Settings + layout `.templ` props, server prop-fill sites, view tests. Same class as the roadmap's `machineSurface bool-last`.                                                                                       |
| 8  | 🔵  | `pauseLiveTranscription(id, paused)`                                                     | `internal/web/assets/island/app/transcribe.js:448`                                                              | The name says pause; the parameter can UN-pause (`calls.js:156` passes the hold state both ways). Command-Query lie in miniature: a toggle wearing a one-way verb.                                                                                                                                                                                                             | `setLiveTranscriptionPaused(id, paused)`                                                                               | `transcribe.js`, one call site `calls.js:156`, island test specs. No DOM/wire impact (function name only).                                                                                                          |
| 9  | 🔵  | bare `bool` params `asr`, `textAgent`, `sso`, `primary`                                  | `views.VoicemailRow`, `NewMessageForm`, `Transcript(At)`, `Bubble`, `LoginPageExtra` (`.templ` sources)         | Boolean-trap at templ call sites (`@Bubble(m, l, true)` — true of what?). Param renames to question form cost nothing at call sites.                                                                                                                                                                                                                                           | `asrEnabled`, `textAgentOn`, `ssoOn`, `primaryMethod`                                                                  | `.templ` signatures + `templ generate` regen; zero caller-visible change. Already on the 2026-10-25 roadmap as a class (`machineSurface bool-last`) — this enumerates the view-side instances.                      |
| 10 | 🔵  | `max` (local)                                                                            | `internal/chatloop/chatloop.go:55`, `internal/reports/reports.go:325,329`                                       | Redefines the Go 1.21+ builtin `max` (revive `redefines-builtin-id`); readers can't tell a local from the builtin at the use site.                                                                                                                                                                                                                                             | `peak` / `top` per site                                                                                                | 3 local variables, zero API surface. (`docpin_test.go`'s `real` is the test-file sibling.)                                                                                                                          |
| 11 | 🔵  | `ListThreads`/`ListArchivedThreads`/`SearchThreads`/`ListAllThreads` → `[]threadSummary` | `internal/store/messages.go:212,220,259,522`                                                                    | Exported methods return an unexported element type — callers outside `store` can hold the slice but never spell it (revive `unexported-return` ×4). The ladder is FROZEN, so exporting `ThreadSummary` may instead collide with `readmodels.ThreadSummary` (the ratified view home) — ruling needed, not a mechanical rename.                                                  | Either export as `store.ThreadRow` (row-shape word, avoids the collision) or accept + document                         | `store/messages.go` + every cross-package caller of the four methods (incl. the admin `ListAllThreads` seam) + arch note that `readmodels.ThreadSummary` stays the view home.                                       |
| 12 | 🔵  | `extensionIds`                                                                           | `internal/web/views/passkey_view_test.go:20`                                                                    | Non-universal abbreviation of IDs (revive `var-naming` ×2).                                                                                                                                                                                                                                                                                                                    | `extensionIDs`                                                                                                         | Single test file.                                                                                                                                                                                                   |

## 3. Explicit exemptions (checked, not findings)

- **Ratified machine vocabulary** — `FailureKind`/`IsRetryable`, `Verdict`,
  `deliveryMoves`+`DeliveryAdvance`, `faxMoves`+`FaxAdvance`, `AppendToThread`,
  `gateway.Resolution` (immediate/deferred), the closed event set
  (`MessageReceived`, `MessageDeliveryAdvanced`, `ThreadFlagChanged`,
  `FaxAdvanced`): all present in code exactly as DOMAIN_LANGUAGE's memo names
  them. The naming bar holds.
- **One-home helper names** (`listRows`, `pbx.do`, `makeSession`,
  `requireMultipartTo`, `applyStatusWebhook`, `recordCallIdem`,
  `machineSurface`, …) and the `*Gate` readmodel family, the `*Sweeper`
  retention family, the cqrs verb-noun command family (`AppendMessage`,
  `RecordFax`, `ExpungeFax`, `MintEnrollToken`, `AdvanceDelivery` — commands are
  deliberately verbs; the FACTS are the nouns).
- **`*Any`/`ListAll*`** cross-owner seams — named-deliberate per AGENTS (admin
  surface), pinned by tests.
- **Wire/config contracts** — DTO fields, `koanf` config keys
  (`island_beacon`…), `StatusQueued/Sent/Delivered/Failed` wire enums, matrix
  spec words (`roomEvent`, `receiptContent`, `accountDataEvent`), vCard RFC
  words (`Card.property`), Q.850 (`q850CauseOf`), `DeviceRegistry` (the domain
  doc's own word for the push-device store).
- **Test fixtures** — `internal/arch/testdata/erraudit-rule-probes/` names
  (`I`, `S`, `obj`, `joinish`, `tok`, `E`, `NewThing`) are DELIBERATELY bad:
  they probe the erraudit rule. Fixing them would defeat the fixture.
- **`pushRegistryAdapter`** — unexported, composition-root-only, the idiomatic
  Go way to satisfy `push.DeviceRegistry` over a foreign type; "Adapter" is the
  honest description of exactly one hop.
- **Island JS discipline** — `$`/`els` (universal JS idioms, `T` precedent),
  question functions (`isTranscribing`, `networkOnline`, `registerWasRejected`),
  set-pairs (`ringbackStart/Stop`, `titleFlashStart/Stop`), `sanitizeDialable`
  (ratified). Only finding #8 rises above zero.

## 4. Strengths to preserve

- Domain vocabulary wins the type list: `Thread`, `Line`, `Extension`,
  `FaxJob`, `ProviderRef`, `Bubble`, `Verdict`, `Bridge`-adjacent seam names —
  a telephony operator recognizes her system.
- Errors and booleans follow Go convention strictly (`ErrNotFound`,
  `ErrDisabled`, `IsZero`, `HasMore`, `OverBudget`).
- `cmd/webphone` boot verbs are uniformly honest (`propagatef`, `bootJoin`,
  `classifyBootError`, `reportBootPanic`, `recoverBootPanic`).
- Zero vague-verb functions (`handle*`, `process*`, `manage*`, `perform*`) in
  183 non-test source files; zero Manager/Handler/Processor/Util TYPES; the
  HTTP `Handlers` type was ratified in 2026-09-19 and is the net/http idiom.

## 5. Verdict

**The 2026-09-19 baseline holds**: automated smell count 0 (both the skill
script and the prior review), and the manual sweep over ~400 types found
12 findings, of which exactly ONE is honesty-class (`FindThread`'s
find-verb-over-create-contract — also the only one where two implementations
disagree under one name, which is what makes it worth adjudicating first).
Three domain-alignment findings (`VoicemailItem`, `authmap.Map`, the twin
`Shell` seams) are the vocabulary debt of the post-2026-09 growth spurt
(chat/agent/ownership trains); all are internal-only, so none break wire
contracts, but #2–#4 touch 5–10 files each and should ride one deliberate
rename train each. The rest are polish (boolean polarity, builtin shadowing,
param names) and can land opportunistically. No finding requires touching a
DOM id, error code, or serialized field.

Counts by class: honesty 1 (#1) · clarity 2 (#5, #10) · domain alignment 3
(#2 #3 #4) · consistency/boolean naming 6 (#6 #7 #8 #9 #11 #12) · total 12
rows.
