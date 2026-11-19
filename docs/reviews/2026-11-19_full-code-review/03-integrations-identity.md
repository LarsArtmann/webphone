# Full code review 2026-11-19, part 3 of 4: integrations + identity layer

> Snapshot-rot header — **claims verified as of 2026-11-19 ~15:30 CET**; every cited
> line was read in source today, and the scoped suite was run green today
> (`go test -count=1` over all 14 packages, `go vet` clean, inside `nix develop -c`).
> Companions: `01-*` (part 1), part 2 (domain/store/vcard), part 4 (server/web) in this
> directory; same-day `2026-11-19_data-model-review.md` (F1–F8) and
> `2026-11-19_naming-review.md` cross-referenced, not re-discovered.
> Repo discipline was READ-ONLY: findings only, no code changed, nothing committed.

## Scope + method

All 14 packages in the integrations + identity layer, **68 Go files visited**
(30 code, 38 test; ~15.5 k LOC incl. tests), each read file-by-file per the
full-code-review skill, judged against AGENTS.md doctrine (CRM data mesh,
paperless seam, ASR, textagent chain, push seam, matrix/signal v1 fences +
chatloop backoff, SSO/passkey, authmap.Line 2026-11-18, admin read-only,
config unknown-key WARN). TTL cache constants, the 15 s-budget /
`context.WithoutCancel` fan pattern, and WARN-not-fail boot postures were
treated as pinned design, not findings.

## Per-file verdicts

| Package              | Files                                                           | Verdict                                                                                                                                                                                                                                                                                                                                                                                                                                                                                               |
| -------------------- | --------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `internal/config`    | config.go, unknown.go + 7 test files                            | **Strong.** Fail-closed seam validation is exemplary (both-or-neither, at-most-one secret source, shape-before-read). unknown-key WARN is schema-derived with dynamic-map exemption and value-free output. One doc-comment misattachment (F-A), one test-hermeticity gap (F-E). ASR provider enum rawness already filed as data-model F4.                                                                                                                                                             |
| `internal/crm`       | client, resolver, context, transport, vocabulary + 6 test files | **Exemplary.** Data-mesh doctrine holds end to end: `Person{Ref,Name}` identity, context strip consumes context only (identity absent from the wire DTO by design, forward-compat fixture + producer snapshot pinned), vocabulary bridge one table both directions, single-flight + TTL caches with failures never cached, authenticated health leg bypasses the cache, per-client CA pool scoped to this seam. Snapshot test (`context_snapshot_test.go`) is the model consumer contract. Nits only. |
| `internal/paperless` | paperless.go + 3 test files                                     | **Clean.** Fire-and-forget archiver, lazy metadata ensure (failed ensure not cached), duplicate = inert success, nil-interface-not-nil-pointer return.                                                                                                                                                                                                                                                                                                                                                |
| `internal/asr`       | asr, client, google, transport, outcomes + 5 test files         | **Strong.** Two provider kinds behind one `Provider` seam, wiring structs, local-outcomes health leg (no probe traffic), shared round-trip classifier. Gaps: `LocalityOf` untested (F-C), no per-roundtrip debug log on the Google wire (F-I3), oversized-read posture inconsistent with crm (F-H).                                                                                                                                                                                                   |
| `internal/textagent` | client.go + 2 test files                                        | **Strong.** draft-id path guard, verdict pass-through with agent phrasing, mode/reset, `DraftFromTranscript` rides `/draft`. One real byte-boundary bug in the transcript cap (F-B).                                                                                                                                                                                                                                                                                                                  |
| `internal/push`      | push, fcm, ntfy, fixtures + 4 test files                        | **Exemplary stdlib-only seam.** JWT-bearer grant + `messages:send` in two POSTs, poisoned-token invalidation on 401, `WithoutCancel` + 15 s budget + panic containment + prune-on-`ErrUnregistered` all in ONE home (`Fan`), wire fixtures rendered from the same pure builders the senders marshal. Gaps: lock-across-mint (F-D), panic containment untested (F-F).                                                                                                                                  |
| `internal/matrix`    | client, receiver, sync + 2 test files                           | **Strong.** Thin CS-API client, DM map + m.direct maintenance, txn-id idempotent sends, read-receipts→delivery machine, self-heal arms (poisoned cursor, oversized answer) both self-limiting and test-pinned. Doc-comment duplication (F-A).                                                                                                                                                                                                                                                         |
| `internal/signal`    | client, receiver + 2 test files                                 | **Strong.** JSON-RPC demuxer with correct single-conn locking, panic-contained read loop, reconnect-on-next-call, attachment fence arms from config with per-attachment caps. One real escaping bug in `dataURI` (F-G).                                                                                                                                                                                                                                                                               |
| `internal/chatloop`  | chatloop.go + test                                              | **Clean.** Rejections stop loops loudly, Infrastructure keeps retrying — the documented narrowing; jitter ladder pinned 50× per attempt.                                                                                                                                                                                                                                                                                                                                                              |
| `internal/sso`       | sso.go + test                                                   | **Strong.** Lazy discovery, per-exchange secret read, nonce/PKCE/S256, group allowlist fail-closed both ways, single-flight discovery with the safe buffered-channel pattern, claims audit line carries groups only in the journal. Well tested incl. concurrency.                                                                                                                                                                                                                                    |
| `internal/userauth`  | userauth.go, enroll.go + test                                   | **Strong.** Strict `identitymodel.UserID` at the door, authmap alias for the mapping, enroll tokens hashed-at-rest + burned-on-use + no token-state oracle, projection-failure WARN hook.                                                                                                                                                                                                                                                                                                             |
| `internal/authmap`   | authmap.go + test                                               | **Exemplary.** The 2026-11-18 `Line{Extension,DID}` model as documented: line-less users and duplicate lines unrepresentable, deterministic shared-line resolution, per-request admin property, passwords read-per-login never cached. Stringly join already filed as data-model F7.                                                                                                                                                                                                                  |
| `internal/admin`     | admin.go + test                                                 | **Clean.** Read-only inventory, nil-safe fail-closed, deterministic sorts, identities normalization re-asserted at construction.                                                                                                                                                                                                                                                                                                                                                                      |
| `internal/reports`   | reports.go + test                                               | **Clean.** Pure projection, deterministic total orderings, Undated counted not guessed, NoAnswer causes in FreeSWITCH's own vocabulary.                                                                                                                                                                                                                                                                                                                                                               |

## Findings

Severity scale: **small** (real defect, cheap fix) · **low** (debt/nit with a
named fix) · **info** (observation, no action demanded). Nothing medium or
high found in this layer.

### F-A · low (3 sites, one class) · duplicated/misplaced doc comments

| Site                                  | What                                                                                                                                                                                                                                                                                                  |
| ------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `internal/config/config.go:1192-1200` | `validatePasskey`'s doc block is physically attached to `validateMatrix` (the two comment blocks are contiguous above `func validateMatrix`); `validatePasskey` itself at `:1249` has none.                                                                                                           |
| `internal/matrix/client.go:56-63`     | "New builds the client. Config validation has already failed closed…" appears twice — once orphaned above the `var _ gateway.MessageGateway` conformance (where `New`'s doc used to live), once folded into `ClientConfig`'s comment; the real `New` at `:74` now documents itself with the leftover. |
| `internal/signal/receiver.go:50-53`   | "// NewReceiver builds the receiver; attachmentsDir is signal-cli's…" is immediately followed by "// NewReceiver builds the receiver." — a truncated duplicate.                                                                                                                                       |

**Why it matters:** godoc renders the wrong prose on the wrong symbol, and the
pattern is copy-paste drift from insertions between decl and comment — the same
class that eats docs silently as the file grows.
**Fix:** move each block to its owning symbol; one commit, zero behavior.

### F-B · small · `internal/textagent/client.go:250-253` — transcript cap slices mid-rune

`trimmed = "… " + trimmed[len(trimmed)-SummarizeHintCap:]` truncates by BYTES.
When the cut lands inside a multibyte rune (German umlauts, emoji — likely in
call transcripts), the hint starts with an invalid UTF-8 sequence, which
`json.Marshal` silently replaces with U+FFFD. Not a protocol break, but the
agent receives a mangled first character on exactly the boundary case, and the
behavior is untested (no test references `DraftFromTranscript`/`SummarizeCap`).
**Fix:** back the slice start up to a rune boundary
(`for !utf8.RuneStart(trimmed[0]) { trimmed = trimmed[1:] }` or decode the last
rune), plus one table test pinning the multibyte boundary and the cap itself.

### F-C · low · `internal/asr/asr.go:65-90` — `LocalityOf` has zero direct tests

A pure trust-classifying function with real branch logic (172.16–31 octet
parse, `.lan`/`.home.arpa` suffixes, dotless hostnames, `::1`) is referenced
only by `internal/app/app.go:580` and pinned nowhere — the sole test touch is
the empty-label default in `server/transcripts_api_test.go:240`. The same-day
data-model review filed the missing _type_ (its F1); the missing _test_ is this
finding. The RFC1918 172.x parse is precisely the code that regresses silently.
**Fix:** one table test in `internal/asr` over the closed input set
(loopback, ::1, 10.x, 192.168.x, 172.15/16/31/32.x, .lan, .home.arpa, dotless,
public, unparseable).

### F-D · low · `internal/push/fcm.go:261-263` — token mint holds the sender mutex across the network

`fetchAccessToken` holds `s.mu` for the whole exchange round trip (up to
`fcmRequestTimeout` = 10 s). Concurrent `Send`s whose cached token is still
VALID also serialize behind the minting caller (the cache check needs the same
lock). Stampede-prevention is the right goal and this achieves it, but the
worst case (mint 10 s + per-device sends × 10 s) can exceed the 15 s
`fanBudget`, after which the fan's ctx cancels the remaining devices. The code
base already has the better shape: `sso.providerFor`'s buffered-channel
single-flight lets valid-cache readers pass and joiners wait on the flight.
**Fix:** adopt that shape (cache check under lock; joiners wait on a buffered
chan; only the leader mints), or at minimum document the serialization as
deliberate next to the mutex.

### F-E · low · `internal/config/family_test.go:48-52,64-70` — Load-driven tests skip the scrub/env discipline

`config_test.go` established `scrubEnv` + `WEBPHONE_CONFIG=absent-file` (the
env provider reads the whole process env and `Load` falls back to
`/etc/webphone/config.json`); `family_test.go` (external package) calls
`config.Load()` with neither. On the actual PBX host — where
`/etc/webphone/config.json` exists — these cases read production config and can
misattribute or flip. CI is clean (no such file), which is why it never fired.
**Fix:** point `WEBPHONE_CONFIG` at an absent temp path (and scrub if the
external package ever grows a scrub helper); `t.Setenv` alone is not enough.

### F-G · small · `internal/signal/client.go:369` — data-URI filename interpolated unescaped

`fmt.Sprintf("data:%s;filename=%s;base64,%s", mime, att.Name, …)` — an
attachment name containing `;`, `,`, or whitespace (user-chosen file names on
the outbound lane) corrupts the parameter list signal-cli parses: the send is
rejected or the filename truncated. **Fix:** percent-escape the filename
(`url.PathEscape`, or strip the delimiter set) at this one site; add a test
with `photo;v2, final.png`.

### F-F · low · `internal/push/push_test.go` — Fan's panic containment has no pin

`push.go:236-241` recovers a panicking transport/registry per the 2026-10-27
owner ruling (goroutine containment), but no test drives a panicking sender or
registry (grep: no `panic(` in any push test). The containment is a deliberate
safety mechanism that a refactor could silently drop.
**Fix:** one test with a sender whose `Send` panics; assert `Settle` completes
and the other device still received its ring.

### F-H · info · oversized-read posture is inconsistent across seams

`crm` distinguishes `readCap+1` → `crm.too_large` (both read sites);
`matrix.do` learned the same lesson and documents why silent truncation is bad
(`client.go:327-338`); `asr.decodeTranscript` (transport.go:24) and
`textagent.do` (client.go:154) still silently truncate at 1 MiB and surface
oversized answers as decode errors. Harmless at current payload sizes; worth
one small train if either seam ever carries bigger bodies.

### F-I · info / nits

1. `internal/crm/context.go:217-219` — the cached `Context` is returned as a
   struct copy, but `NextTask *Task` and `Deals` still alias the cache entry;
   a mutating consumer would corrupt the memo. Read-only consumers today.
2. `internal/matrix/receiver.go:141` — ingest-before-Save means a crash in the
   window replays the batch (at-least-once, duplicates inbound rows). Bounded
   to one batch; matrix event ids exist if a dedup is ever wanted. The comment
   says "loses" — the _duplication_ direction is only implied.
3. `internal/asr/google.go:172-176` — the Google wire's defer records stats
   only; the per-roundtrip `slog.Debug` line the OpenAI client emits
   (`client.go:107`, the "#23 operator-greppable latency signal") is missing.
4. `internal/crm/vocabulary_test.go:13` — test name typo
   `TestCRMOuTcomeCovers…`.
5. `internal/crm/crm_test.go:212-216` — dead `resolver` binding and a
   discarded loop variable (`"probe %d", 1` always prints 1).
6. `internal/config/config.go:627/633/640/647` — `var errX error = …`: the
   explicit interface type is fine but reads as noise; harmless.

### Doctrine watch (not re-litigated, recorded for the operator)

- **ntfy 404 ⇒ `ErrUnregistered`** (`ntfy.go:158-159`): AGENTS.md pins 404 as
  the prune verdict for BOTH transports, but the semantics differ — FCM's 404
  names a dead device token, while ntfy's publish-404 can also mean a
  wrong-shaped base URL (topics are normally created on publish). A
  mispointed `push.ntfy.url` would mass-prune every device row; recovery is
  device-side re-registration. Behavior is doctrine-pinned; the asymmetry is
  worth an operator note in the runbook, not a code change here.

## Cross-references (already filed in `2026-11-19_data-model-review.md`, verified still open in source)

- **F1** `asr.Locality` typed enum (`asr.go:65` → app → server → settings.templ).
- **F4** ASR provider vocabulary in four places (`config.go:433/462/971-972`,
  `app/app.go:248`).
- **F7** authmap feeds stringly-keyed across the validated boundary
  (`config.go:917-925` discards the parse; `authmap.passwordFiles` keyed by
  raw `line.Extension.String()`).
- **F2** `push.Transport` stops at the cqrs device gate (cqrs side — part 2's
  package, this layer's `ParseTransport` is the landed half).

## Split-brain candidates

1. **ASR provider vocabulary** (data-model F4, cross-ref) — the only live
   multi-home vocabulary in this layer.
2. **F-A's doc-comment drift** — not a logic split brain, but the same
   reproduce-without-deleting failure mode at three sites.
3. Nothing else: `resolver` forwarders, `admin`→`authmap` delegation, and the
   `chatloop.Sink`/`ReceiptSink` split are single-home delegations, verified.

## Verdict

This layer is the strongest of the repo's quarters I have reviewed against its
own doctrine: every optional seam is nil-safe and fail-closed at boot, every
outbound client is pool-isolated, error families are registered and pinned per
seam, and the hard mechanisms (push fan policy, matrix self-heal, SSO
single-flight, authmap ownership) each live in exactly one place with the
dangerous cases test-pinned. The genuine defects are two small escaping/byte-
boundary bugs (F-B, F-G) and a handful of low-severity test/doc debts; nothing
medium, nothing high, no security regression. 68 files visited; suite green;
tree left untouched.

_Unfixed findings are forward-looking work — the actionable rows above (F-A…F-G)
belong in TODO_LIST.md at the orchestrator's discretion, not entombed here._
