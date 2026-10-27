# Issue: LarsArtmann/cqrs-htmx

Title: `AuditLog projection: bound the in-memory entries (cap + eviction)`

## Why / ## What breaks

The AuditLog projection grows without bound for the lifetime of the
process: every audited user event appends to an in-memory slice that is
never trimmed. A long-running embedder accumulates entries until
restart — webphone (usermgmt v4.16.0 behind a passkey login) left the
projection UNWIRED for exactly this reason
(docs/research/2026-10-25_cqrs-htmx-deep-dive.html, N/A row 22).

Evidence: `usermgmt/audit_log.go:97` —
`a.entries = append(a.entries, entry)` under the mutex; no cap, no
eviction, no persistence anywhere in the file. Verified at HEAD,
2026-10-25.

## Root cause (source)

The projection predates the durable auth-audit journal (auth_audit.go,
#23): the journal now owns the durable record (Auth-typed streams in
the event store), so the in-memory slice is no longer the only record
and can be bounded without losing durability.

## Design / ## Automation steps (repo)

- [ ] Cap the entries slice (ring buffer or trim-on-append), newest
      retained; default sized so a compliance glance at the last hours
      survives
- [ ] Expose cap + dropped-count on the projection status if cheap
- [ ] README: one line — the projection is a bounded recent-window,
      the journal is the durable record

## To verify

- [ ] Test projects >cap events and asserts the cap plus
      newest-retention
- [ ] webphone re-audit flips the AuditLog row from N/A to adoptable

Out of scope: persisting AuditLog to SQL — the journal already is the
durable record.

💘 Generated with Crush
