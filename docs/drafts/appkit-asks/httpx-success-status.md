> [!NOTE]
> This filing was drafted by GLM-5.3 via Crush from an AI-run investigation (the webphone go-appkit fleet-frame re-review, owner-directed), not from a direct filing request.
>
> - [ ] MANUALLY REVIEWED by `@Lars Artmann` at `[<date-time>]`

## Problem

`ResultHandler` hardcodes the success status: "A nil error serializes T as JSON with status 200 (204 when T is a nil pointer — there is no payload to speak of)" (`httpx/result.go:40-42`, sha `ed79e4f`). The error side is fully covered — `errorfamily.HTTPStatus` honors an error's own `HTTPStatuser` override (`go-error-family` v0.12.0 `http.go:22-26`), so per-lane error statuses are already expressible. But success statuses other than 200/204 have no path except writing the response yourself, which bypasses the battery's serialization (and its recording-writer semantics) — at which point `ResultHandler` is no longer doing the work.

The concrete case: 202-accepted job-dispatch lanes. webphone's `POST /api/calls/{id}/review` answers 202 by contract and the island checks the exact code before deduping; every "accepted, work started asynchronously" API in the fleet has the same shape. Today those lanes cannot ride the handler battery.

## Ask

A success-status override that keeps the battery's serialization, e.g.:

```go
mux.Handle("POST /api/calls/{id}/review", httpx.ResultHandler(
    func(c *httpx.Context) (*ReviewAccepted, error) {
        return httpx.WithStatus(http.StatusAccepted, review)
    },
))
```

— a small generic wrapper `WithStatus(code int, T T) T` (or an equivalent sentinel type the `writeResult` path unwraps before its nil-pointer 204 check). Default behavior byte-identical when unwrapped; no API break.

## Why it generalizes

202 is the canonical "dispatched, poll or push for the result" answer across the fleet's agent/push/job surfaces (webphone: review dispatch, ring hooks in other services). With the error side already overridable, this is the one missing half of "the battery can express my API's contract".

Cross-links: sibling asks #59 (strict readiness route), #60 (aggregates-only mode) — together they are webphone's remaining framework-adoption walls (evidence: webphone `docs/review/libs/go-appkit.md`, 2026-11-12 re-review).

💘 Generated with Crush
