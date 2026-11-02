# Session status slice 2: OpenAPI-spec answer + verification self-review

_Snapshot: claims verified as of 2026-11-02 05:29 CET. Covers the work since the 05:19 report: one question ("Do we have an OpenAPI spec and how does it currently work?"), answered from source. No code changed; tree carries only the two status reports (daemon owns commits)._

**Correction issued by this report:** the chat answer said "24 operations" — the reconciled count is **23**. Both grep heuristics counted comment lines (`apidocs.go`'s modeling-rule comments literally contain `catalog.Command[` / `MsgOperation`). Details in (d1).

---

## a) FULLY DONE

1. **OpenAPI question answered from SOURCE this time, not AGENTS.md** — the round-1 improvement (verify before verdict) was actually applied. Verified in-session:
   - `apiCatalog()` declares the API as a go-cqrs-lite catalog; schemas derive from the handler DTOs; `description:` tags feed field docs; `cookieAuth` security scheme.
   - Serving: `openapiHandler` (apidocs.go:414) on the open mux at server.go:670 (`GET /openapi.json`), `application/schema+json`, `no-cache`, `sync.OnceValues` memoized, honest-500 on render failure (never stale bytes).
   - Modeling rules read verbatim from the file header (Command[T] for POSTs, Query[struct{}] for GETs, query-kind+method-override for body-less DELETEs, Retry-After in 429 description text at catalog v4.9.0).
   - Three drift pins read: `TestOpenAPIEndpoint` (served media type), `TestOpenAPISeamSurfacesMatchHandlers` (honest 404s + Retry-After clauses + status inventories), `TestAPIContractPathsAreMounted` (contract↔mux probe via 404-shape discriminator).
   - Split rule: `/hooks/*` + `/events` live in `docs/contracts/asyncapi.yaml` (wirecontracts.go, `webphone contracts` CLI); `RenderOpenAPIContract()` feeds the `webphone catalog` export.
2. **Post-answer reconciliation (this report's recon):** operation count reconciled to 23; route mount located (server.go:670); `json.Deterministic` grep came back EMPTY locally (see d3).

## b) PARTIALLY DONE

1. **The chat answer's evidence is structural, not current-green**: I read the pins but did not RUN them (`go test ./internal/server -run 'TestOpenAPI|TestAPIContract'`). Existence verified; present-tense pass not evidenced this session. Effort S.
2. **docs/contracts/README.md unread** — I cited the split rule from code comments + AGENTS; the canonical consumer-facing statement is one view away. Effort S.

## c) NOT STARTED

1. **HARVEST of report 1's section f** — still blocked on owner answers (its three questions stand, see g).
2. **`json.Deterministic` upstream check** — not in webphone code; settling where it actually lives (go-cqrs-lite catalog exporter) means reading the library. Out of scope until someone leans on the freshness-diff claim.

## d) TOTALLY FUCKED UP

No code damage, no wrong architecture guidance — but three honesty defects, all caught and corrected in this slice:

1. **Quoted a wrong number with confidence: "24 operations".** Both greps counted the file's own modeling-rule COMMENTS (they contain the literal tokens). 25 matched instantiations = 23 real + 2 comments; 24 `MsgOperation` = 23 real + 1 comment. The reconciled truth is **23 operations**. Root cause: grep-count-as-fact without looking at the matches. This is the same family as round 1's "analogical cost claim" — a number I did not earn.
2. **Abandoned search masquerading as not-found:** the route-mount grep was piped through `head -5`, the cutoff swallowed the server.go hit (alphabetically after middleware_test.go), and I moved on, compensating via the test pin instead of finishing the search. The mount was one uncut grep away (found it in this slice: server.go:670).
3. **Inherited parenthetical, attached to the wrong artifact:** "`json.Deterministic` makes it freshness-diffable" — stated inside the asyncapi.yaml sentence, implying local code. Grep: ZERO hits in wirecontracts.go / docs/contracts/. The claim is AGENTS doctrine about the go-cqrs-lite catalog marshals, not something webphone's wirecontracts sets. Round-1 improvement (provenance labels) was applied to the load-bearing claims but leaked on a parenthetical.

## e) WHAT WE SHOULD IMPROVE

1. **Never quote a grep count without eyeballing the matches** — comments match tokens. Print the matches or reconcile two independent counts before stating a number.
2. **A truncated search is an unfinished search** — when `head` cuts output, rerun wider or say "search incomplete"; the cutoff must not become negative evidence.
3. **Provenance labels reach parentheticals** — inherited facts slip in as asides; every unverified-by-me claim gets tagged, not just the thesis-level ones. (Repeat offender: round 1 d2 → round 2 d3. Per the 2+ report rule this practice is now a written check, not a habit.)
4. **Run the pins when explaining a mechanism** — a 10-second `go test -run` converts "this is how it works" into "this is how it works, green as of 05:29".

## f) Things to get done next

Grounded in this slice only (no padding; the real backlog stays in TODO_LIST.md, still unread by design):

| # | Task                                                                                                                                                           | Impact       | Effort | Category      |
| - | -------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------ | ------ | ------------- |
| 1 | Correct the record where the wrong count matters: none persisted outside chat — if any durable doc quotes "24 operations", fix to 23 (grep docs/ for it first) | Low          | S      | Documentation |
| 2 | Run `nix develop -c go test -count=1 ./internal/server -run 'TestOpenAPI\|TestAPIContract'` for a current-green receipt of the seam                            | Low          | S      | Quality       |
| 3 | Read docs/contracts/README.md; confirm the split-rule statement matches what I told the owner                                                                  | Low          | S      | Documentation |
| 4 | Locate `json.Deterministic` in go-cqrs-lite catalog; correct AGENTS.md's phrasing if it misattributes the mechanism to webphone                                | Low          | S      | Documentation |
| 5 | Report 1's f1–f6 remain open (drift audit, CUE ruling note, schema-export spike, HARVEST) — owner questions pending                                            | per report 1 | —      | —             |

## g) Questions I cannot answer myself

1. **Pattern read:** two consecutive questions are contract/schema-shaped (CUE, then OpenAPI). Is there one underlying goal — e.g., a typed client for the Android shell or the stack, or contract-locked consumption? It decides whether report 1's schema-export spike graduates from "parked" to "scheduled".
2. **Persistence:** should the CUE ruling (and now the corrected OpenAPI count, if anything durable) land as owner decision notes / HARVEST into TODO_LIST.md, or stay session ephemera? (Carried from report 1, now blocking two reports' section f.)
3. **Product direction:** is `/openapi.json` documentation-only by design (EventCatalog hub links it), or do you want a generated consumer someday (TS client would collide with the island's verbatim-JS doctrine; the Android shell is the natural fit)? I cannot read intent from code.

---

_HARVEST: both reports' section f await the owner's g-answers; do not entomb._
