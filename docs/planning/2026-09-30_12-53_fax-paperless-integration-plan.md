# Fax → Paperless-ngx archive — optional integration (go-paperless)

**Status:** EXECUTED 2026-10-01 — design landed unchanged on `main`
(config block + test, `fax.Archiver` seam + nil tests, `internal/paperless`
adapter + httptest stub suite + family pin, composition-root wiring,
go-paperless v0.4.2 + vendorHash roundtrip). Gates: full Go suite, smoke
47+4, flake checks incl. the KVM backup VM + drill, vulnix zero-real,
erraudit 0. Named follow-ups below stay open as their own trains.

**Client SDK:** `github.com/larsartmann/go-paperless` (LarsArtmann,
Go 1.27 + `encoding/json/v2` — webphone's 1.27.1 floor imports it
cleanly). The SDK is deliberately CLIENT ONLY ("document sync pipelines
are domain-coupled and stay in the consuming repos") — webphone owns the
pipeline, the SDK provides `Upload`, the `Ensure*` idempotent lookups,
`WaitForTask`, and `TaskOutcome.Duplicate` (content-hash duplicate
refusal). Webphone becomes its THIRD consumer beside InboxClean and
bank-sync.

## Why this fits

Inbound faxes are the one artifact class webphone stores that has no
downstream filing home: `fax.Service.Receive` spools the PDF into the
blob store and the job row is the only index. Paperless-ngx is exactly
the filing home, and the CRM seam (2026-09-22) already proved the
pattern for optional integrations in this repo: OFF by default,
both-or-neither config validation, nil-safe everywhere, downstream
failure never breaks the primary path.

## Scope (v1)

**Inbound faxes only.** `fax.Service.Receive` is the single hook point.
Outbound archiving (post-`transmitted`) can ride the same seam later —
named follow-up below, not v1.

## Design

### Config (internal/config)

- `paperless.url` + `paperless.token`, env `WEBPHONE_PAPERLESS__URL` /
  `WEBPHONE_PAPERLESS__TOKEN` (scalars via env per the `__` convention).
- Both-or-neither, fail-closed `Rejection` — copy the CRM validation
  (`crm.url`/`crm.token` posture, pinned by `TestLoadValidatesCRMConfig`;
  add `TestLoadValidatesPaperlessConfig` beside it), plus the same
  absolute-http(s) URL check.
- New `Config.Paperless` block. Absent = integration off, zero behavior
  change.

### Seam (internal/fax)

Webphone-owned interface; go-paperless types never leak past it:

```go
type Archiver interface {
    ArchiveFax(ctx context.Context, job domain.FaxJob, pdf []byte) error
}
```

- Lives in `internal/fax`; the implementation lives in a new
  `internal/paperless` package and imports `internal/fax` + `internal/domain`.
  Dependency direction: server → paperless → fax/domain; `fax` never
  imports `paperless` (same acyclicity the arch test enforces elsewhere).
- `fax.Service` gains an optional `archiver Archiver` field (nil = off).
  `nil`-safety pinned by test: every existing service test keeps passing
  with nil, and a nil archiver changes no response.

### Adapter (internal/paperless)

- Construction from config: `paperless.New(cfg)` builds the go-paperless
  client plus lazily-resolved metadata IDs (first `ArchiveFax` call runs
  `EnsureTag("fax")` + `EnsureDocumentType("Fax")` + `EnsureCustomField`
  `webphone-fax-id`, cached thereafter). Deliberately NOT at boot: a
  boot-time probe would couple server startup to Paperless reachability.
- Upload per fax: `UploadRequest{Filename, Content, Title, Created,
  TagIDs, DocumentTypeID, CustomFields}` — title `Fax from <remote>
  <YYYY-MM-DD>` (English, operator-greppable like every service reason),
  `Created` = the job's received timestamp, provenance custom field =
  the webphone fax ID (the SDK's own Gmail message-ID precedent). NO
  per-number `EnsureCorrespondent` — that mints a correspondent per
  caller; the remote number stays in the title.
- 20 MiB buffer ceiling is already enforced by `fax.MaxPDFSize`, so the
  adapter's `[]byte` handoff is safe (the SDK buffers ≤25 MB by design).
- Errors classified with go-error-family at origin (per the 2026-09-30
  error convention): `errorfamily.New*/Wrap*` with `<seam>.<op>` codes,
  pinned by `internal/paperless/family_test.go`.

### Wire-up (internal/server)

- `Deps` grows the archiver (nil when config absent); `fax.New` takes it.
- Fire-and-forget AFTER persist + notify: the service reads the spooled
  PDF (`blobs.Open(job.DocumentPath)` → `io.ReadAll`) and archives on a
  detached goroutine (`context.Background()` + its own timeout). WARN on
  failure. A dead Paperless never delays or fails a fax — same posture
  as the island's fire-and-forget `recordCrmCall`.

### Failure semantics (honest)

- Webhook-sync was REJECTED on evidence: `Receive` has no
  provider-ref dedupe, so a slow Paperless delaying the ack invites
  provider retries → duplicate fax rows. Async keeps the ack fast.
- Cost of async: an in-flight upload at SIGTERM is dropped and never
  retried in v1 (no outbox). Accepted for v1 — fax volume is tiny and
  the blob store keeps the original. The named follow-up below closes it.
- Paperless content-hash dedupe (`TaskOutcome.Duplicate`) makes any
  future re-ingest inert; no dedupe state on our side.

## Non-goals (v1)

- No outbound-fax archiving (seam supports it; hook lands later).
- No retry/outbox persistence, no archive-status column on `FaxJob`.
- No Paperless-backed READING in the UI — the fax panel keeps serving
  blobs from the blob store; Paperless stays a downstream copy, never a
  second storage truth (no split brain).
- No per-extension tags or per-number correspondents.

## Named follow-ups (post-v1, own trains)

- Outbound archiving on `transmitted` (direction-flavored title/tag).
- Archive visibility: `archive_status` column + a boot-time sweep that
  re-offers unarchived jobs (turns the SIGTERM limitation into a
  self-healing queue).

## Gates

- Unit: config both-or-neither + URL shape; adapter against an
  `httptest` Paperless stub (upload form fields, Ensure* caching,
  task-poll verdict); nil-archiver zero-delta test.
- `internal/paperless/family_test.go` pin (error convention).
- `nix develop -c go test -count=1 ./...` + buildflow green.
- vendorHash roundtrip in the same breath as the go.mod addition
  (standing post-close reminder). Binary-size impact trivial (small
  SDK); no DOM/UI change, so no stack E2E re-run owed.
- Lands on `main` per the stack input policy; rides the next release.

## Sizing

~400–500 lines incl. tests: config block + test, `internal/paperless`
adapter + family pin, `fax.Service` optional dep + nil test, server
wiring. Half a session.
