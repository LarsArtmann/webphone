# Capability-seam audit findings (plan T31) + retention note (T32)

**Date:** 2026-10-30 · **Scope:** findings only — no silent fixes (plan
§8). The fax train made THE fax seam declare capability honestly; this
audit asks the same question of the other optional seams: can the UI
tell whether the feature actually works?

## Paperless — INVISIBLE SEAM (the one real gap)

**Question:** can the UI tell whether inbound-fax archiving works?

**Answer: NO.** `paperless.url`+`paperless.token` build the fax
archiver, but there is NO Settings row for the seam and NO per-job
archive marker on fax rows. The archive leg is fire-and-forget by
doctrine (failures log, the blob store stays the storage truth), so a
deployment that believes it archives to Paperless learns otherwise only
from the journal. This is the closest living relative of the fax lie:
a configured-but-broken integration with zero UI truth.

**Suggested shape (NOT built):** one conditional Settings row (the
missing-only pattern this app already uses for conditional rows) plus —
optionally, only if the owner wants per-fact truth — an `archived`/
`archive-failed` marker on the fax row's status timeline. The Settings
row alone answers "is it wired"; the marker answers "did THIS fax
land". Owner call; recorded, not built.

## CRM — honest

The Settings row distinguishes connected / auth-failed / setup / unset
via the bounded probe (`crmDetail`), the context strip's absence is a
state (never an empty shell), name enrichment degrades to raw numbers
with a debug line, and call-logging outages toast (502). Transport
failures are never cached. Verdict: no gap worth a row.

## ASR — honest

The Settings row names provider kind·model AND the processing-location
sentence (`asrDetailWithLocality`, M17): local vs remote is stated
where caller audio goes — a trust fact. Affordances are gated by both
config and view props (the Notifier carries the boot flag, so pushes
match full renders). Verdict: no gap.

## Push — honest by absence

The Settings row renders only when a transport is wired, naming the
live transports; off = no row and the device API 404s, which the
native shells (webphone-android) read as "this server cannot ring
closed apps". The PWA itself has no closed state to ring. Verdict: the
surface that matters belongs to the consuming shell (P1), not this
app. No gap here.

## Retention under both fax lanes off (T32)

The retention sweep deletes fax jobs WITH their documents after
`retention_days` (default 0 = forever). Both lanes off stops NEW rows
only; stored documents stay owner-controlled under the same policy as
before the capability train. No orphans, no leak, no change needed.
