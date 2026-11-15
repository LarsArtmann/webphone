# server.Deps post-ladder audit (T19, 2026-11-14)
> PIN `server.Deps` as DELIBERATE. The 2026-11-08 accessor-collapse design is
> dead: it planned a handle around the readmodels engine, which the 2026-11-11
> ladder retirement (A43) deleted, while its own trigger ("EXECUTE only when
> the eighth read model lands") fired twice over with no harm. This memo is
> the ONE home for the ruling; the TODO_LIST row (705-area, item 5) cites
> here when struck.

## 1. Field inventory (34 fields, server.go:196-334)

| Group | Fields | N |
| --- | --- | --- |
| Read-model gates (readmodels.*) | Messages, Faxes, Contacts, Snippets, Transcripts, Reviews, Dials, Notes, AgentAudit, PushDevices | 10 |
| Identity/session seams | Sessions, UserAuth, SSO, Admin | 4 |
| Optional integration seams | CRM, ASR, ASRLocality, TextAgent, Push, Dedup, Trace, TraceMetrics | 8 |
| Always-wired service seams | Messaging, Fax, PhoneAPI, Hubs, Notifier | 5 |
| Config/probe seams | Config, Probe, Dashboard, DB, BlobRoot | 5 |
| Derived data / test overrides | Shared, ProjectionStatus | 2 |

## 2. Count vs the seven-field smell

The 2026-11-08 doc counted SEVEN metaengine-backed fields (Contacts, Snippets,
Transcripts, Reviews, Dials, Notes, AgentAudit; Messages/Faxes excluded as
pre-engine store seams). The struct now carries TEN readmodels.* gates: the
seven survivors, Messages/Faxes re-typed journal-backed (SUPERB Phases 5-6),
and PushDevices (Phase 3). Numerically the smell worsened 7 -> 10 and the
trigger fired unexecuted; substantively the surface IMPROVED — one
implementation home per gate (internal/cqrs), contracts-only readmodels
(arch-pinned), and no second backend lane left to wire per addition.

## 3. Per-gate home and nil posture (all ten must-invoke in internal/app; never nil in prod)

| Gate | One home (internal/cqrs) | Nil posture |
| --- | --- | --- |
| ThreadGate | Threads (threadsstore.go:37 pin) | required — no nil arm |
| FaxGate | Faxes (faxesstore.go:53 pin) | required — no nil arm |
| ContactGate | Contacts (contacts.go:268 pin) | reads nil-tolerant; writes unguarded (drift) |
| SnippetGate | Snippets (snippets.go; test-only pin) | panels checked; save/delete unguarded (drift) |
| TranscriptGate | Transcripts (transcriptsstore.go:34 pin) | 503 (api/transcripts_api.go:53,129) |
| ReviewsGate | Reviews (callreviews.go:198 pin) | 503 (api/review_api.go:157,200) |
| DialGate | DialWindows (dialwindows.go; test-only pin) | 503 (api/dialgate_api.go:76,102) |
| NotesGate | Notes (callnotes.go:283 pin) | 503 (api/note_api.go:51,83,133) |
| AgentAuditGate | AgentAudit (agentaudit.go:238 pin) | silent degrade (auth_audit.go:45, panels.go:642) |
| PushDeviceGate | PushDevices (pushdevices.go:216 pin) | 404-gated (api/push_api.go:63) |

Eight gates carry the compile-time `var _` pin (SnippetGate/DialGate are
test-only — the one gap); internal/arch/wiring_parity_test.go fails CI by name
if a Deps field misses its wiring literal (the M16 Notes class).

## 4. Ruling: DELIBERATE — pin, do not collapse
- The collapse's premise died with the engine: the planned readmodels handle
  wrapped the retired readmodels.db store. Collapsing now would only blur the
  per-seam fail-closed boundaries that are policy (503 vs 404 vs silent
  degrade) — the same reason the 2026-11-08 doc excluded the optional seams.
- Each field earns its place: 10 gates x distinct consumer surfaces; 4 identity
  seams x distinct login modes; 8 optional seams x distinct config gates;
  DB/BlobRoot carry the "readiness probes only" comment.
- Hygiene residue (small, not a refactor): (a) two missing `var _` pins
  (Snippets, Dials); (b) the Contacts/Snippets doc comments claim nil-degrades
  while the write routes (actions.go:315,394,417; api/contacts_api.go:61,131,153)
  call unguarded — fix the comments or guard the writes; (c) accessor-name
  drift in internal/app/cqrs.go (Dials() vs *Svc()) — cosmetic.

## 5. Evidence
server.go:196-334,438-482; internal/app/app.go (the Deps literal);
internal/app/cqrs.go:109-184 (accessor family); internal/readmodels/doc.go;
internal/arch/arch_test.go:92-129; internal/arch/wiring_parity_test.go;
docs/planning/2026-11-08_server-deps-accessor-collapse-design.md;
docs/planning/2026-11-09_storage-ladder-retirement.md (sections 1, 4).
