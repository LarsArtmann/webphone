# Draft upstream ask: batch identity lookup (`GET /api/contacts/by-phone?numbers=`)

> DRAFT 2026-11-15 (T20/F95) — owner tick to file. Not sent. Written in the
> github-voice register; adjust before filing. Target repo: the CRM (Kith,
> ~/projects/crm), machine API.

## Body

**Batch identity lookup on the machine API — `GET /api/contacts/by-phone?numbers=`**

Today the phone's list surfaces (Messages rail, drafts panel, History) resolve
caller names one number at a time: N rows cost N `GET /api/contacts/by-phone`
requests per render. At list density that is the only N+1 in the integration,
and the resolver's caches (positive 6h / negative 5min) only hide the repeat
cost, not the first-render fan-out.

**Ask:** a batch form of the existing lookup —
`GET /api/contacts/by-phone?numbers=<dialable>,<dialable>,…` — answering the
same `Person` shape keyed by the REQUESTED number (including `null` for
numbers that match nothing, so callers do not guess at absent keys):

```json
{ "+4989123456": { "id": "…", "name": "Ada Lovelace" }, "+4917612345678": null }
```

**Constraints carried over from the single-number contract:** bearer auth
unchanged; matching stays CRM-owned (digits-normalized, same as today);
unknown numbers are `null`, never an error; response stays under the phone's
4 KiB read cap per answer (cap the batch at ~50 numbers, reject larger with
422 and the count in the message).

**Consumer side:** webphone's `Resolver.People` gains a batch path behind the
same cache semantics; the single-number call sites stay until the batch lands
(no flag day). One phone at a production-scale list (~20-40 rows) is the
entire demand — this is a latency ask, not a throughput one.

Reopen trigger if you'd rather not: none needed — the N+1 stays bounded by
the caches and the 1024/256 caps; it is annoying, not load-bearing.
