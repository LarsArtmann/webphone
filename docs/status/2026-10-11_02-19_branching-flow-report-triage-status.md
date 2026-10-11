# Status: branching-flow report triage (owner-pasted, 2026-10-11 02:19)

Owner pasted the full `branching-flow all . --format markdown` output
(72 rows across 9 non-empty families). Per the dedup-registry protocol,
every flagged site was re-read at HEAD and attributed against the
standing rulings before anything was touched: the registry (10-05 dupe
+ mixin entries), the `.buildflow.yml` branching-flow skip rationale
(2026-09-18/09-22 triage classes), and the AGENTS.md composition /
strong-id / wire-DTO doctrine. One finding was fixed; the other 71 are
attributed. No re-litigated ruling was overturned.

## Verdict summary

| Family            | Rows | Fixed | Attributed | Notes                                                       |
| ----------------- | ---- | ----- | ---------- | ----------------------------------------------------------- |
| Duplicate types   | 29   | 0     | 29         | 2 `actionable` groups both carry standing rulings            |
| Phantom types     | 14   | 1     | 13         | `server.EventName` introduced; rest are the non-identity class |
| Panic conditions  | 8    | 0     | 8          | 6 verified false positives, 2 standing EXPLICIT_PANIC class  |
| Strong ID         | 4    | 0     | 4          | CallID = opaque island token, pass-through class             |
| Composition anti  | 2    | 0     | 2          | Both are documented contracts (koanf config, Deps DI surface) |
| Composition mixin | 12   | 0     | 12         | Standing 10-05 doctrine + `.buildflow.yml` BY-DESIGN class    |
| samber/do         | 3    | 0     | 3          | `App.Start` InvokeNamed is the AGENTS-documented pattern      |

## The one fix: `server.EventName`

`Publish(extension, eventName, html)` carried a transpose hazard with a
silent failure mode (an event/payload swap broadcasts garbage event
names — tabs never update, no error anywhere). Unlike the other phantom
rows, the event vocabulary was ALREADY a closed set of five named
constants, and the repo's own design language types closed string sets
(`domain.OutboundStatus`, `domain.Direction`). Typing the constants
completes that pattern: the swap no longer compiles (the constant does
not fit the `html string` slot), wire names are unchanged, call sites
needed zero churn. Changed: `server.EventName` + typed constants +
`ExtensionHubs.Publish`/`Notifier.publish` signatures, two test
comparison conversions. Verified: full `go test ./internal/server/`
green, build clean.

## Family verdicts

### Duplicate types (29 rows)

- `vcard.Card` / `domain.SharedContact` / `server.apiSharedContact`
  (`actionable`): REJECTED 2026-10-05 in the registry — parse/domain/
  wire boundary layers sharing a Name+Number coincidence; consolidating
  welds an external-format parser type to the domain and the JSON DTO.
  Verified intact at HEAD (contacts_api.go `apiSharedContact`, the
  json-tag wire contract comment on `domain.SharedContact`).
- `config.CRM` / `config.Paperless` (`actionable`): the OPEN owner call
  (10-05, resurfaced by a SECOND detector then) — now a THIRD sighting
  (this run). Still an owner call: they are independent seams whose
  both-or-neither validation is per-seam; a shared auth struct would
  touch the koanf/env contract for zero behavioral gain. NOT resolved
  unilaterally. Logged in the registry sweep log.
- Brand-marker structs (9), store wrapper structs with one `db` field
  (5), `FaxWebhook`/`Webhook` embedding, `ErrInvalidFax`/`ErrInvalidSend`
  (parallel-seam errorfamily rejections), the CRM/ASR wire response
  shapes (different external providers), `Extension`/`Phone` (distinct
  validated domain identities), `contextKey`: all tool-marked false
  positives; each re-read and confirmed (the `Webhook`/`FaxWebhook`
  embedding IS the 09-18 mixin extraction the tool still flags).

### Phantom types (14 rows)

Policy class per `.buildflow.yml`: "wants phantom types for display
fields … maximalist churn with zero safety gain for non-identity data".
Accepted under that standing rationale (each site re-read):

- `baseURL/token(/model)` constructor pairs (asr, crm, paperless):
  single caller each (composition root), config-validated both-or-
  neither at boot, failure mode is immediate and visible.
- `direction, number` (crm.LogCall client+resolver): the number travels
  VERBATIM by domain intent ("the CRM owns the whole matching
  contract") — typing it as `domain.Phone` would sanitize and CHANGE
  the wire; direction is the CRM's lowercase wire string.
- `rest, rawQuery` (pbx.ResolvePath): the documented join-path-and-
  query-separately home; param names carry the contract, single caller.
- `subDir, ext` (blob.Save): literal args at both call sites, failure
  immediately visible in dev.
- `providerRef, errorText` / `providerRef, failureKind, failureDetail`
  (store update methods): adjacent strings, but the failure vocabulary
  is documented at `domain.Message.FailureKind` and the full ProviderRef
  type would touch webhook JSON DTOs, the partial unique index, and the
  byte-pinned golden tests — churn the standing policy declines for
  non-identity data.
- `page, limit` (ListMessagesPage): mis-order degrades VISIBLE (short
  window), `page < 0` guarded; few call sites.
- collisions (`rejected/transport -> key` i18n keys, `needle -> query`
  in likeEscape): plumbing idioms, nothing to type.

FIXED: `eventName, html` (sse.Publish) — see above.

### Panic conditions (8 rows)

False positives (guard verified in source):

- `transcripts.go:154` — `groups[idx]`: idx only ever comes from the
  map entry written as `len(groups)` immediately before the append, or
  is freshly computed the same way; `idx < len(groups)` is an
  invariant the analyzer cannot see through the map.
- `transcripts.go:168` (×2) — `reverseLines` two-pointer swap under
  `i < j`, classic bounds-safe shape.
- `vcard.go:34` — `*current` deref is behind the `current == nil`
  early return at the top of the same closure.
- `bootreport.go:149-150` — `copyRows[c]` reads are behind the
  `c >= 0 && int(c) < len(copyRows)` condition in the same statement.

Accepted (standing EXPLICIT_PANIC class, `.buildflow.yml`: "go:embed
guarantees the subtree; the panic guards a compile-time impossibility"):
`assets.go:27` (assets glob) and `server/assets.go:21` (fs.Sub). Both
now carry a one-line rationale comment at the site so the next reader
(or the next triage) does not re-derive it.

### Strong ID (4 rows: CallID)

REJECTED, two grounds:

1. Standing class: the call id is an opaque ISLAND-owned token passed
   through verbatim (same class as the 09-18 pbx-UUID ruling — "opaque
   foreign identifiers passed through verbatim"). It stays inside the
   transcript feature, never joins CDR rows (AGENTS: "CDR rows carry no
   uuid, so NEVER join the two lists"), and the handler door ALREADY
   validates it (transcripts_api: trim + empty + 128-cap, pinned by
   tests).
2. Wire-DTO doctrine hazard: the value round-trips raw through the DOM
   delete affordance, the export zip (`callId`), and EXISTING raw rows
   in `call_transcripts.call_id`. A branded `id.ID` renders
   `Brand:value` via `String()` (history attribute, store writes), so
   old rows would become undeleteable unless every edge converted —
   exactly the "JSON wire DTOs stay raw" trap.

### Composition anti-patterns (2 rows) + mixins (12 rows)

`config.Config` / `server.Deps` large-struct (low confidence, warning):
both are deliberate contracts — the koanf/`__`-env config surface and
the documented DI surface with per-field nil-optional seam docs.

Mixins: all REJECTED under the 10-05 registry doctrine (field-shape
similarity is architectural ROLE, not duplication) plus the
`.buildflow.yml` BY-DESIGN class:

- Client trio (asr/crm/pbx): the documented chokepoint idiom, parallel
  by design ("the same chokepoint shape as pbx.Client.do" is in-code).
- config.Passkey→PasskeyRuntime, config.PasskeyUser→MappedUser:
  consolidating would make userauth import config — AGENTS FORBIDS
  ("userauth imports ONLY config-free types").
- CallTranscriptSegment/CallTranscript/exportTranscript,
  InboundMessage/OutboundMessage: domain/DTO boundary layers, same
  class as the vcard trio REJECT.
- fax.Service/messaging.Service: documented intentional symmetry.
- idemStore/unreadCache, SweepResult/Counts, GoogleConfig/config.ASR:
  coincidental field shapes behind different domains.

### samber/do (3 rows)

`App.Start` resolving the named critical pair + dashboard via
`InvokeNamed`+`wrapf` is the AGENTS-documented composition pattern
(composition-root-only resolution, fail-fast miswires). Accepted as
doctrine; nothing to change.

## Verification

- `go build ./...` clean; `go test -count=1 ./internal/server/
  ./internal/app/` green after the change.
- Full suite + `nix fmt` + buildflow run recorded in the session log;
  branching-flow itself is buildflow-SKIPPED (documented noise policy),
  so this doc + the registry sweep line are the record of record.
