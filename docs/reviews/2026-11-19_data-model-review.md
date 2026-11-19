# Data-model review 2026-11-19: the vocabularies that arrived after the doctrine

> Snapshot-rot header — **claims verified as of 2026-11-19 ~08:40 CET**, every cited
> declaration read in source today. Sixth series entry by content: 2026-09-19 →
> 2026-09-30 → 2026-10-14 → 2026-10-25 (review 4) → 2026-11-11 → **this**.
> Companions: `2026-11-11_data-model-review.html` (the journal-world swap review whose
> roadmap this report also closes), `2026-11-18_line-ownership-data-model.html` (the
> authmap model doc), `2026-11-19_naming-review.md` (same-day naming pass).
> Scope: the full domain-model walk (domain → cqrs facts/deciders → readmodels gates →
> config structs → server DTOs → views), the review-4 open roadmap adjudicated, and the
> post-review-4 fresh material (authmap.Line ownership, the session wire's typed
> `lines[]`, the note-autosave decider, catalog export types, the beacon leg vocabulary).
> Nothing implemented — proposals only, per the series contract.

## 0. Current state

| Measure                                 | Value                                    |
| --------------------------------------- | ---------------------------------------- |
| Review-4 open roadmap items adjudicated | 5 (+2 carried by the 11-11 entry)        |
| Landed and verified in source           | 6 of 7 (incl. both 11-11 carry-ons)      |
| Closed by later ruling (declined)       | 1 — `sso.Finish`                         |
| New findings cataloged                  | 8 (1 medium, 4 small, 3 low)             |
| Baseline-ruled premises re-verified     | all 9 hold (3 edge doors + 6 Deps seams) |

The headline: **four of review 4's five open items landed** (push.Transport, FaxJob
RelPath, asr/matrix wiring structs, machineSurface bool-last), the fifth was **declined
by the 11-11 entry** (sso.Finish, site-ruled), and the fresh material since — authmap,
the session wire, the note decider, catalog mirrors, beacon legs — is doctrine-conformant
with a handful of vocabulary residuals listed below. The 10-14 thesis ("vocabularies that
live in comments and string comparisons instead of types") has exactly **one live
instance** in the tree (`asr.LocalityOf`), one standing carry-forward (ASR provider), and
one adoption gap where a landed brand stops at a gate.

## 1. Series cross-reference

### 1.1 The review-4 open roadmap, adjudicated

| # | Review-4 item                                    | Verdict                   | Evidence verified today                                                                                                                                                                                                                                                                                                                                                                                                                                              |
| - | ------------------------------------------------ | ------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 1 | `push.Transport` defined type + `ParseTransport` | **Landed**                | `internal/push/push.go:46` (`type Transport string`), `:56` (`ParseTransport`), registry `senders map[Transport]Sender` (`:163`); the registration API parses at its door (`server/api/push_api.go:124`); `TestParseTransportVocabulary` pins the closed set. Residual: the brand stops at the device gate (finding F2).                                                                                                                                             |
| 2 | `sso.Finish` value object into `Complete`        | **Superseded — declined** | The 2026-11-11 entry closed this as a site-ruled decline ("foreign protocol tokens; two guards already reject mixing; pay only when a third field joins"). Source carries the ruling: `internal/sso/sso.go:158` — `Complete(ctx, code, verifier, nonce string)` with `//nolint:branching-flow // OAuth2 protocol tokens (code/verifier/nonce), not domain concepts`. **Not re-proposed.** Reopen trigger (11-11's own): a third protocol field joining the callback. |
| 3 | `FaxJob.DocumentPath` as `domain.RelPath`        | **Landed**                | `internal/domain/fax.go:83` — `DocumentPath RelPath`; both lifecycle constructors (`:112`, `:128`) carry it. The journal facts (`FaxReceipted.DocumentPath string`, `fax.go:281`) stay raw per the engine-row doctrine — correct split, no action.                                                                                                                                                                                                                   |
| 4 | asr/matrix wiring structs                        | **Landed, evolved**       | `matrix.New(cfg ClientConfig)` (`matrix/client.go:68-74`) and `asr.NewClient/NewWhisperCppClient(cfg ClientConfig)` (`asr/client.go:50,77`); the asr package has since split (client/google/outcomes/transport) and grown `NewGoogleClient(cfg GoogleConfig)` (`google.go:96`) — the struct door absorbed the growth, which is the design paying off.                                                                                                                |
| 5 | `machineSurface` bool-last                       | **Landed**                | `server/platform/internal_directory.go:37-39` — `(w, r, missing, enabled bool)`, bool last; the nolint comment names the 2026-10-25 roadmap item as closed.                                                                                                                                                                                                                                                                                                          |

### 1.2 The 11-11 entry's own roadmap, adjudicated

| Item                                                                       | Verdict                   | Evidence                                                                                                                                                                                                                                                                                                                                                                                                                                                                       |
| -------------------------------------------------------------------------- | ------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| `push.Transport` / `FaxJob.DocumentPath` / wiring structs / machineSurface | Landed                    | Same rows as above (the 11-11 entry re-carried review 4's list; all landed after it).                                                                                                                                                                                                                                                                                                                                                                                          |
| ThreadSummary re-home (out of the store package)                           | **Landed**                | `readmodels/threads.go:66` — `ThreadSummary` now lives in readmodels, re-parsed from the raw `ThreadSummaryView` at the gate ("wire stays raw by doctrine, the brands come back here").                                                                                                                                                                                                                                                                                        |
| ThreadFlag re-home                                                         | **Settled in readmodels** | `readmodels/threads.go:77` — `type ThreadFlag string` with `FlagPinned/FlagArchived/FlagMuted`. It is a _typed_ enum already; its home beside `ThreadGate.SetThreadFlag` (the only consumer) is defensible layering: the domain row (`domain.Thread`) carries the three axes as bools, the gate carries the command vocabulary. Moving it to domain would split it from its consumer. Adjudicate: no action; revisit only if a second consumer of the flag vocabulary appears. |
| Close P6 (sso value object)                                                | **Closed by decision**    | The decline above is this item's landing.                                                                                                                                                                                                                                                                                                                                                                                                                                      |

### 1.3 Baseline-ruled premises re-verified (never re-litigated — premise check only)

All still hold, three of them relocated by the `internal/server` carve into subpackages:

- **Edge doors**: `ParseToken` at the scan door (`session/sqlite.go:181`, reasoned nolint); BearerGate's secret/unavailable pair (`server/hooks/webhooks.go`, nolint cites the 2026-10-25 baseline); matrix `Upload(name, mimeType, data)` (`matrix/client.go:199`, foreign-vocabulary ruling).
- **Six single-implementer Deps seams**: `TraceRecorder` (`server/platform/trace.go:20`), `DedupWindow` (`server/hooks/idempotency.go:18`), `SSOClient` (`server/auth/auth.go:64`), fax `Store` (`fax/service.go:57`), matrix `Cursor` (`matrix/receiver.go:48`), api `Server` (`server/api/api.go:45`). All intact with their test-side second implementers; none gained a second production implementer (nothing to re-open).

### 1.4 Standing carry-forward from review 4, item 7

**ASR provider selection — still open, and the stated trigger FIRED without landing.**
Review 4 said "becomes a parsed enum at the next config touch." The Google train WAS that
touch: `config.ASR` grew `Project`/`Location`/`Language` and a validation arm that
enumerates the whole vocabulary in prose (`config.go:971-972` — `"not one of \"openai\",
\"google\", \"whispercpp\""`), yet `Provider string` (`config.go:433`) is still compared
raw at `config.go:462` (`a.Provider == "google"`) and `app/app.go:248`
(`cfg.ASR.Provider == "whispercpp"`). The vocabulary now lives in four places. Folded
into finding F4.

## 2. Inventory (the walk)

- **`internal/domain`** (~3.5k lines): the ID wall unchanged and strong — branded
  `Extension`/`Phone` structs with parse doors, `go-branded-id` generics for
  Thread/Message/Attachment/Fax ids, `Network`+`Address` (network owns the validation
  rule), `FailureKind`/`Verdict`/`Disposition`/`ReviewState` closed enums with
  `Parse*`/`Must*` doors, machine tables (`deliveryMoves`, `faxMoves`, `AppendToThread`)
  as the one legality home, the SSE event set as a closed interface union
  (`events.go`), `RelPath` in use, `DeliveryEvent` as an unexported-field closed union
  with constructors as the only door. The catalog export mirrors
  (`Catalog*Schema`, `catalog.go`) are a forced reflection split, pinned by
  `TestCatalogSchemaParity` (`domain/catalog_test.go:154`).
- **`internal/cqrs`** (the ONE journal world since 2026-11-09/11): commands, deciders,
  facts, projections for threads/faxes/notes/reviews/contacts/snippets/dial/push/
  enroll/agentaudit. Facts are deliberate primitive mirrors (journal codec ruling);
  `DeliveryStep` is the ruled wire form of `DeliveryEvent`, rebuilt through the same
  constructors (`threadfacts.go:145-167`) — no second legality door.
- **`internal/readmodels`**: gates + view types only (purity arch test forbids engines
  forever). Views raw-at-rest per the 11-11 ruling; gates re-parse into brands
  (`ThreadSummary` from `ThreadSummaryView` is the model citizen).
- **`internal/config`**: 26 type decls; both-or-neither pairs validated at load;
  identities keys validated as normalized extensions round-trip (`config.go:917-925`).
- **`internal/server`** DTOs: `SessionIdentity` wire carries typed
  `lines[{extension,did}]` with BD003 directives at the fields (`auth/session_identity.go:96-131`);
  the openapi/asyncapi render from handler DTOs; machine tables own legality.
- **Views**: `FaxCapability` plain bools = the documented convert-once shape (config
  keeps the pointer-bool opt-in-off discipline); `LoginPageExtra` is the one
  flagged signature (F5).

## 3. Findings

Severity grades follow the series' convention (adopt / bounded / decline-with-reason).
Every finding is graded against the ratified doctrine: brands only where a repo-owned
identity or closed vocabulary exists; wire DTOs, journal facts and engine views stay raw
with written reasons; foreign tokens never branded mechanically (BD003).

### F1 · Adopt (small) · P4 — `asr.Locality`: a closed vocabulary with no constants at all

**Where**: `internal/asr/asr.go:65-90` (`LocalityOf` returns bare `"local"`/`"remote"`
literals — the comment is the vocabulary's only home) → `app/app.go:576-581` →
`server/server.go:274` (`Deps.ASRLocality string`) → `server/platform/configjs.go:73`
(wire) → `web/views/settings.templ:491` (`props.ASRLocality == "local"` — a string
compare at the leaf of a four-hop seam).

This is the exact 10-14 thesis instance, and the last one of its size in the tree: a
two-value, repo-owned trust label minted by one function, threaded raw through three
package seams, and string-compared at the render. `push.Transport` was precisely this
shape and landed cleanly.

```go
// asr — the trust label the Settings chip and caption strip render.
type Locality string

const (
    LocalityLocal  Locality = "local"
    LocalityRemote Locality = "remote"
)

func LocalityOf(rawURL string) Locality // unchanged logic, typed answers
```

Migration: `LocalityOf` + `Deps.ASRLocality`/`platform` deps/`settings.templ` props
re-typed; the config.js wire field stays `string` (JSON DTO ruling) with the conversion
at the render site (`string(props.ASRLocality)` or a `MarshalJSON`-free `String()`
method). One train, asr family tests extended. Declined alternative: a `bool isLocal` —
the empty/off state ("seam disabled") must stay distinct from both values; three states
are the model.

### F2 · Adopt (small) · P4 — `push.Transport` stops at the device gate

**Where**: `internal/cqrs/pushdevices.go:228` — `Upsert(ctx, owner domain.Extension,
deviceID, transport, token string)` (a three-string tail) and the command
`SavePushDevice{Owner, DeviceID, Transport, Token string}` (`:39-46`).

The landed brand flows API → `ParseTransport` → Fanout registry, then dies at this line:
the gate re-accepts raw strings and the command carries them on. The nolint at the gate
says why it is raw _today_ — "verbatim port of the retired engine store's Upsert
signature (parity rule)" — but the parity rule itself prescribes the second,
separately-verified refactor, and the port landed 2026-11-11. This is that change:

```go
func (s *PushDevices) Upsert(ctx context.Context, owner domain.Extension,
    deviceID string, transport push.Transport, token string) error
```

Migration: `server/api/push_api.go` passes the already-parsed value (it currently
stringifies it away); the decider converts to the raw fact string at the journal edge
(`fact.Transport` stays `string` — engine-row doctrine); push family + gate tests
re-verified. DeviceID and Token stay raw (client-generated foreign key and
write-only credential — BD003 class, reasoned).

### F3 · Adopt (small) · P2-adjacent semantics — the note decider's echo drop is body-lane-only

**Where**: `internal/cqrs/callnotes.go:145` (`decideSaveNote` drops a same-body save over
an existing slot — pinned by `TestNoteSameBodySaveIsNotARevision`) vs `:162`
(`decideSaveDisposition` journals unconditionally). The gate mirrors it:
`readmodels.NotesGate.Set` returns the `changed bool` the handler's SSE nudge gates on,
`SetDisposition` (`notesgate.go:77`) returns bare `error` — it _cannot_ be echo-aware
even if the handler wanted it.

Consequence: a disposition re-click (or a second tab re-PUTting the same verdict)
journals a no-change event, bumps `UpdatedAt` — which the History panel renders as a
recency/order signal — and fires the SSE nudge for other tabs over an unchanged slot.
The asymmetry is undocumented: nothing in either file says disposition echoes are
wanted. Two honest fixes; pick one:

```go
// Option A (symmetry — the autosave doctrine generalized):
if _, exists := s.Slots[cmd.CallID]; exists && prev.Disposition == cmd.Disposition {
    return nil, nil // an unchanged verdict is not a revision
}
// plus SetDisposition(...) (bool, error) at the gate, same contract as Set.

// Option B (documented asymmetry): a written reason at decideSaveDisposition
// naming the operator-click cadence as too sparse to matter.
```

Option A is one guard + one signature + a test symmetric to the body-lane pin; it also
deletes a spurious-journal-entry class. Recommend A.

### F4 · Adopt (small) · P4 — ASR provider enum: the trigger fired, the enum didn't

**Where**: `config/config.go:433` (`Provider string`), `:462` (`Google()` compare),
`:971-972` (validation enumerating the set in prose), `app/app.go:248`
(`== "whispercpp"`).

```go
// config — the ASR wire-selection vocabulary, closed since the google train.
type ASRProvider string

const (
    ASROpenAI     ASRProvider = "openai"     // "" aliases this (wire default)
    ASRWhisperCpp ASRProvider = "whispercpp"
    ASRGoogle     ASRProvider = "google"
)
```

Migration: the field stays string-tagged on the config wire (koanf), `Load` parses into
the enum once (the validation at `:971` becomes the parser's error), `Google()` and the
`app.go` selection compare typed constants. json/koanf tags unchanged, config file
bytes unchanged. This is review-4 item 7 landing with its trigger already behind it.

### F5 · Adopt (trivial) · FLAG_PARAM — `LoginPageExtra(primary, sso bool)`

**Where**: `web/views/login_page.templ:27`, called once at `server/auth/auth.go:174` —
`views.LoginPageExtra(!passkeyOn, deps.SSO != nil)`.

Two positional bools at the auth front door's only consumer section — the exact class
`machineSurface` was (and the same fix shape the 10-25 entry ratified for it). A
transposed call compiles and swaps which sign-in method is the disclosure's primary.

```go
type LoginExtra struct { Primary, SSO bool }
templ LoginPageExtra(x LoginExtra)
// call site: views.LoginPageExtra(views.LoginExtra{Primary: !passkeyOn, SSO: deps.SSO != nil})
```

Migration: one call site, view regen, `passkey_view_test.go` shape pins re-checked.

### F6 · Bounded (low) · documentation-grade — `islandLatencyLegs` is a nested stringly map

**Where**: `server/platform/islandlatency.go:19` —
`map[string]map[string]bool{"dial_to_audio": {...}, "asr_round_trip": {...}}`.

The design is right (server-owned cardinality, wire legs outside the set dropped — the
beacon doctrine held), and the map _is_ the parser at the door. The residual: the
interval vocabulary has no named constants, and `RecordBeacon(interval string, ...)` +
`islandSeriesKey{interval, leg string}` carry it raw across the api→platform seam. A
defined key type names the set without changing behavior:

```go
type latencyInterval string
const (
    intervalDialToAudio  latencyInterval = "dial_to_audio"
    intervalASRRoundTrip latencyInterval = "asr_round_trip"
)
var islandLatencyLegs = map[latencyInterval]map[string]bool{...}
```

Low value beyond readability; do it only when the vocabulary next grows (a third
interval), at which point the typed key also stops a typo'd literal from silently
dropping beacons forever. Legs stay per-interval raw strings (they are interval-scoped,
so a global leg enum would be the wrong model — worth writing down here so nobody mints
one).

### F7 · Bounded (low) · config-doctrine completion — the authmap feeds are stringly-keyed across a validated boundary

**Where**: `config.go:917-925` validates every `identities` key with
`domain.ParseExtension` (round-trip normalized) **and then discards the parse**;
`app.mappedUsers` (`app.go:608-623`) re-joins on raw strings (`identities[raw]`);
`ExtensionPasswordFiles map[string]string` rides raw into `authmap.New`, which keys its
internal `passwordFiles` by `line.Extension.String()`.

Guards exist (bad keys refuse boot), so this is not a silent-miss defect today — but the
join's correctness is string-equality discipline plus a validation step in another
package, where the repo's own pattern (parse once at the door, carry the brand) would
make it structural:

```go
// config.Load mints, the join consumes brands:
identities      map[domain.Extension]string // DID verbatim (11-18 ruling: display-verbatim)
passwordFiles   map[domain.Extension]string
```

Migration rides the next config-struct touch (not worth its own train): `Load` builds
the typed maps after validation, `mappedUsers` and `authmap.New` take them, the
`panic`-on-unparseable in `mappedUsers` (`app.go:616`) becomes dead code and deletes.
`authmap.Line.DID` stays `string` — ruled display-verbatim by the 11-18 model doc;
canonicalizing to `Phone` would change rendered bytes, which is a product decision, not
a type decision.

### F8 · Bounded (low) · P4 residual — `ReviewsGate.Set` parameter train

**Where**: `readmodels/reviewsgate.go` — `Set(ctx, owner, callID string, state
domain.ReviewState, draftID, text, language string, confidence int) error`.

Four strings in sequence; `text`/`language` are adjacent content strings whose swap
compiles and journals wrong (language is rendered as a label). Under the 11-11 ruling
(gates re-parse into brands; content stays content) the honest fix is not brands but a
payload value object mirroring the domain fact it becomes:

```go
type ReviewDraft struct {
    State      domain.ReviewState
    DraftID    string // foreign (agent-side id) — stays raw, BD003 class
    Text       string
    Language   string
    Confidence int
}
Set(ctx context.Context, owner domain.Extension, callID string, d ReviewDraft) error
```

One implementation + one handler call site; do it opportunistically with the next
reviews-seam touch.

## 4. Verified deliberate (added to the standing baseline with reasons)

- **`server.Deps`, 34 fields** — the composition-root wiring seam, parity-pinned by
  `TestAppWiresEveryServerDep`. A god struct by size, a manifest by role; splitting it
  would scatter the one wiring surface the test audits.
- **Journal facts duplicate key + flat fields** (`MessageAppended` carries
  `Key.ThreadID`, `Summary.ThreadID` and `ThreadID`; `CallNoteSaved` carries `Key` +
  `Owner` + `CallID`) — engine-row doctrine: the metaengine key extractor reads `Key`,
  the folds read flat fields, and each fact is built at exactly one site per decider
  from the same variables. The invariant is constructor-discipline-enforced and the
  wire is frozen; embedding would change journal bytes. Keep; note in the baseline so
  detectors stop re-finding it.
- **`Catalog*Schema` mirrors** (`domain/catalog.go`) — forced by the schema reflector's
  blindness to unexported branded state; `TestCatalogSchemaParity` fails the build on
  drift. The mirror _is_ the pinned pattern for reflector-fed exports.
- **`SessionIdentity` / `SessionLine` wire rawness** — BD003 directives at both decls;
  the active-line correlation rule ("the entry whose extension equals the top-level
  extension — the client never correlates parallel lists") is the 11-18 doc's own
  design, landed as specified.
- **textagent two-string methods** (`Approve(id, text)`, `DraftFor(to, hint)`,
  `DraftFromTranscript(to, transcript)`) — each carries a site reason ("a swap fails
  the server's validation loudly"); the ruled loud-failure class, same family as
  `signal.New`/`ntfy` in review 4's decision log.
- **`asr.Provider` interface + `ClientConfig`/`GoogleConfig`** — the wiring-struct door
  (review-4 item 4) absorbed the Google provider without a third string-train; the
  interface stays the seam, config stays out of the service (userauth doctrine).

## 5. Decision log

| Decision                                            | Rationale                                                                                                                                                                     | Rejected alternative                                                                                          |
| --------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------- |
| `Locality` typed in `asr`, not `domain`             | A provider-trust label local to the ASR seam; domain holds cross-seam telephony concepts. Wire field stays string.                                                            | `domain.Locality`: drags a Settings-chip detail into the domain package for zero safety.                      |
| Transport through the gate, facts stay raw          | The gate is in-process (brands allowed, per the 11-11 gates-re-parse ruling); the journal fact is wire (raw, per the engine-row ruling). The parity rule's own second change. | Branding `PushDeviceRegistered.Transport`: changes journal semantics for no consumer gain.                    |
| Disposition echo drop (Option A) over a doc comment | The autosave doctrine ("the journal must record CHANGES, not echoes") is lane-agnostic; symmetry deletes a spurious-nudge class for one guard.                                | Option B (documented asymmetry): leaves a test-unpinned behavioral wrinkle on a save path two surfaces share. |
| ASR enum parsed in `config.Load`, wire untouched    | The vocabulary is repo-owned and boot-rejected today; the parser is the validation, moved one step earlier. koanf/json tags unchanged — no config-file migration.             | A `Config` method layer (`Provider.IsGoogle()` family): more surface, still stringly at the store boundary.   |
| `LoginExtra` struct, not bool-last reorder          | Two flags have no order; the machineSurface fix (reorder) worked because there was exactly one flag.                                                                          | Reorder: leaves the same trap armed for the next flag.                                                        |
| DID stays verbatim string (again)                   | Re-affirms the 11-18 ruling: display-verbatim owner formatting; canonicalization is a product decision.                                                                       | `domain.Phone`: changes rendered bytes on every configured deployment.                                        |
| Interval-scoped leg vocabularies stay maps          | Legs belong to their interval; a global leg enum is the wrong model (a leg name may repeat with different meaning per interval).                                              | `type latencyLeg string` global set: invents cross-interval identity that does not exist.                     |

## 6. Migration roadmap (refresh)

1. **`asr.Locality` defined type + constants** (S) — closes F1; asr family tests + the
   settings chip pin.
2. **`push.Transport` through `PushDevices.Upsert` + `SavePushDevice`** (S) — closes F2;
   the parity rule's owed second change; push gate + api tests.
3. **Disposition echo symmetry** (XS) — closes F3; decider guard + `SetDisposition`
   changed-bool + the disposition-lane twin of `TestNoteSameBodySaveIsNotARevision`.
4. **`ASRProvider` enum at `config.Load`** (S) — closes F4 / review-4 item 7; the
   trigger already fired.
5. **`LoginExtra` struct at the login view seam** (XS) — closes F5; view regen + shape
   pins.
6. **`ReviewsGate.Set` payload struct** (XS) — closes F8; rides the next reviews touch.
7. **authmap typed feed maps** (S, opportunistic) — closes F7 at the next config-struct
   train; deletes the post-validation panic class.
8. **`islandLatencyLegs` typed interval key** (XS, deferred) — F6; only when a third
   interval joins the vocabulary.

Closed this entry: `sso.Finish` (declined 2026-11-11, site-ruled — reopen only when a
third protocol field joins the callback), ThreadSummary re-home (landed), ThreadFlag
re-home (adjudicated settled in readmodels beside its only consumer).

Standing invitations kept open from earlier entries: transcript remote + hub map keys
(the 10-14 item 7 remainder — untouched by this walk, no new pressure).

---

_Generated 2026-11-19 · sixth series entry · read-only review: no source files changed ·
repo discipline: no commits made (auto-commit daemon owns the tree) · claims verified as
of ~08:40 CET: every cited declaration read in source today._
