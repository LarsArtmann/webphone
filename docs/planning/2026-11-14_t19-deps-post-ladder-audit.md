# server.Deps post-ladder audit (T19, 2026-11-14)

> Ruling: PIN `server.Deps` as DELIBERATE. The accessor-collapse design
> (2026-11-08) is dead — its target was a handle wrapping the readmodels
> engine, and the engine retired 2026-11-11 (A43) — while its own trigger
> ("EXECUTE only when the eighth read model lands") has since fired twice
> over with no harm. This memo is the ONE home for the ruling; the
> TODO_LIST row (705-area, item 5 of the 2026-11-05 deferred findings)
> cites here when struck.

## 1. The field inventory (34 fields — server.go:196-334)

| Group                              | Fields                                                                                  | Count |
| ---------------------------------- | --------------------------------------------------------------------------------------- | ----- |
| Read-model gates (readmodels.*)    | Messages, Faxes, Contacts, Snippets, Transcripts, Reviews, Dials, Notes, AgentAudit, PushDevices | 10 |
| Identity/session seams             | Sessions, UserAuth, SSO, Admin                                                           | 4     |
| Optional integration seams         | CRM, ASR, ASRLocality, TextAgent, Push, Dedup, Trace, TraceMetrics                        | 8     |
| Always-wired service seams         | Messaging, Fax, PhoneAPI, Hubs, Notifier                                                 | 5     |
| Config/probe seams                 | Config, Probe, Dashboard, DB, BlobRoot                                                   | 5     |
| Derived data / test overrides      | Shared, ProjectionStatus                                                                 | 2     |

## 2. The count vs the seven-field smell

The 2026-11-08 design doc counted SEVEN metaengine-backed fields
(Contacts, Snippets, Transcripts, Reviews, Dials, Notes, AgentAudit;
Messages/Faxes were excluded as pre-engine store seams). Today the struct
carries TEN readmodels.* gates: the seven survivors plus Messages/Faxes
re-typed to journal-backed gates (SUPERB Phases 5-6) and PushDevices
(Phase 3). Numerically the smell worsened 7 -> 10 and the row's trigger
fired unexecuted. Substantively the surface IMPROVED: every gate has ONE
implementation home (internal/cqrs), readmodels is contracts-only
(arch-pinned), and the per-addition cost that motivated the collapse (a
second backend lane to wire) no longer exists.

## 3. Per-gate home + wiring + nil posture

| Gate            | One home (internal/cqrs)        | Compile-time pin      | Nil posture (test comps)                          |
| --------------- | ------------------------------- | --------------------- | ------------------------------------------------- |
| ThreadGate      | Threads (threadsstore.go:37)    | `var _` yes           | required — no nil arm                             |
| FaxGate         | Faxes (faxesstore.go:53)        | `var _` yes           | required — no nil arm                             |
| ContactGate     | Contacts (contacts.go:268)      | `var _` yes           | read lanes nil-tolerant; WRITE lanes unguarded (drift) |
| SnippetGate     | Snippets (snippets.go)          | test-only (drift)     | panels List checked; save/delete unguarded (drift) |
| TranscriptGate  | Transcripts (transcriptsstore.go:34) | `var _` yes      | 503 (api/transcripts_api.go:53,129)               |
| ReviewsGate     | Reviews (callreviews.go:198)    | `var _` yes           | 503 (api/review_api.go:157,200)                   |
| DialGate        | DialWindows (dialwindows.go)    | test-only (drift)     | 503 (api/dialgate_api.go:76,102)                  |
| NotesGate       | Notes (callnotes.go:283)        | `var _` yes           | 503 (api/note_api.go:51,83,133)                   |
| AgentAuditGate  | AgentAudit (agentaudit.go:238)  | `var _` yes           | silent degrade to journal line (auth_audit.go:45, panels.go:642) |
| PushDeviceGate  | PushDevices (pushdevices.go:216)| `var _` yes           | 404-gated (api/push_api.go:63) + panels check     |

All ten resolve must-invoke from the CqrsService accessors in the
composition root (internal/app/app.go, the server.Deps literal) — never
nil in prod; internal/arch/wiring_parity_test.go fails CI by name if a
Deps field misses its wiring literal (the M16 Notes 503 class).

## 4. Ruling: DELIBERATE — pin, do not collapse

- The collapse's premise died with the engine: `readmodels.Handle` was
  to wrap the retired readmodels.db store. Collapsing now would only
  blur the per-seam fail-closed boundaries that are policy (503 vs 404
  vs silent degrade) — the same reason the 2026-11-08 doc excluded the
  optional seams from its own scope.
- Each field earns its place: 10 gates x distinct consumer surfaces;
  4 identity seams x distinct login modes; 8 optional seams x distinct
  config gates; DB/BlobRoot carry the "readiness probes only, nothing
  else may use them" comment.
- Hygiene residue (small, not a refactor): (a) SnippetGate/DialGate
  lack the `var _` pin their eight siblings have — add two one-liners
  in internal/cqrs; (b) the Contacts/Snippets doc comments claim "nil
  degrades", but the write routes (actions.go:315,394,417;
  api/contacts_api.go:61,131,153) call unguarded — a true-nil test
  composition would panic there; fix the comments or guard the writes;
  (c) accessor-name drift in internal/app/cqrs.go (Dials() vs *Svc())
  — cosmetic.

## 5. Evidence

server.go:196-334 + 438-482 (Deps + gate fan-out); internal/app/app.go
server.Deps literal (all ten must-invoke); internal/app/cqrs.go:109-184
(accessor family); internal/readmodels/doc.go (contracts-only doctrine);
internal/arch/arch_test.go:92-129 (engine-retired pins);
internal/arch/wiring_parity_test.go (field<->wiring parity);
docs/planning/2026-11-08_server-deps-accessor-collapse-design.md (the
dead design, trigger in "When it pays"); docs/planning/
2026-11-09_storage-ladder-retirement.md (the retirement, sections 1+4).
