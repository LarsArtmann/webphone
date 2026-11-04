# go-retry — library utilization audit

**Verdict:** partial fit, ADOPTED on the fitting half (2026-11-04, closeout T17): the chat reconnect loops now ride `retry.ComputeDelay` (exponential + ≤50% additive jitter, hard-capped, saturation-hardened) instead of the hand-rolled deterministic doubling, and gate on `errorfamily.Classify` so Rejections stop the loop LOUDLY. The bounded-attempts model (`retry.Do`) still fits nothing — a reconnect loop cannot exhaust — and stays declined.
**Pin:** v0.9.0 (go.mod:32, **DIRECT since 2026-11-04**; previously indirect via go-paperless + go-cqrs-lite/middleware) · **Latest:** v0.9.0 · **Currency:** current · **Status:** direct dep
**Audited:** 2026-11-04 · **Prior reports:** none (first audit).

## Status

| | |
|---|---|
| Module | `github.com/larsartmann/go-retry` |
| In webphone | direct: `internal/chatloop` (`nextDelay` via `ComputeDelay`, the Rejection-stop gate in `Backoff`); consumers: matrix + signal receiver loops |

## What the library offers

`Do(ctx, Config, AttemptFunc, opts...)` / `DoWithValue[T]`; `Config{MaxAttempts=3, InitialDelay=100ms, MaxDelay=5s, Multiplier=2.0, IsRetryable=errorfamily.IsRetryable, DelayFunc, OnRetry, OnExhausted}`; `WithJitter(JitterAdditive)` (exponential + additive jitter), `WithDelayFunc` (e.g. honor Retry-After); exported `ComputeDelay`/`Backoff` for DIY loops; sentinels `ErrExhausted`/`ErrCanceled`/`ErrDeadlineExceeded` (Infrastructure family, unwrap to last attempt). Retryable predicate rides go-error-family by default — the two libraries compose.

## Webphone's own answer (the overlap)

- WAS: `internal/chatloop/chatloop.go`: `Sleep(ctx, d)`, `NextBackoff(d, max)` (double, cap, **no jitter**), `Backoff(...)` (log + wait + double), callers carrying a Duration. **LANDED 2026-11-04:** attempt-numbered ladder (`nextDelay(attempt, max)` → `retry.ComputeDelay(1s, max, 2, attempt)`), `Backoff` stops on `errorfamily.Classify(err) == Rejection` with an Error line (a revoked matrix token or a refusing signal daemon ends the loop naming the permanence; every other family retries) — `NextBackoff` deleted (attempt-derivation from a jittered delay is unsound: 1.4s floors to log2 0 and resets the ladder, which is why the callers thread an attempt int and reset it on success). Consumers: signal ping loop (internal/signal/receiver.go), matrix /sync loop (internal/matrix/receiver.go).
- Everything else is deliberately not retried: push transports are fire-and-forget (a lost ring is healed by the next ring), `pbx.Client` rides request timeouts without a loop, `asr` returns Transient for the island to retry at the UX level.
- Note `failsafe-go` v0.9.8 also sits in the graph (indirect via cqrs-htmx/usermgmt) — two retry libraries transitively present, one directly used.

## Fit analysis

| Aspect | go-retry | webphone loop | Fit |
|---|---|---|---|
| Termination model | bounded attempts (`MaxAttempts`) | unbounded reconnect until ctx cancel | ✗ — a reconnect loop cannot exhaust; `Do` is the wrong shape |
| Delay computation | `ComputeDelay`: exp + **additive jitter**, Retry-After aware | fixed double with cap; **thundering-herd prone** on shared infra (all sessions reconnect simultaneously after a server restart) | ✓ — the real gap |
| Retryable classification | `errorfamily.IsRetryable` predicate | logs every error, retries unconditionally | ✓ — free once the errorfamily seams already classify |
| OnRetry/OnExhausted hooks | yes | inline logging | neutral |

## Ruling

**ADOPTED 2026-11-04 (the delay half + the classification half; closeout T17):** `nextDelay` rides `retry.ComputeDelay` (additive jitter kills the thundering-herd lockstep; the hard cap is the library's saturation-tested math, not a hand-rolled branch), and `chatloop.Backoff` gates on `errorfamily.Classify` — **Rejections stop loudly, everything else retries**. That gate is deliberately NARROWER than `errorfamily.IsRetryable` (which also stops Infrastructure): a garbage homeserver response (Infrastructure decode failure) must not kill inbound-message intake until app restart; only the "retrying changes nothing" class (matrix auth/permission per `matrix.classify`, signal daemon RPC refusal) ends the loop, with an Error log as the visible surface. Full `retry.Do` adoption still fits nothing current (bounded attempts vs reconnect-until-ctx). vendorHash needed NO re-pin: the promote changed only go.mod's comment column — the module content was already in graph and go.sum.

**Refresh trigger:** a second retry-shaped loop (e.g. ASR segment uploads) — the natural home for `retry.Do`; an upstream ComputeDelay semantics change (jitter strategy, cap behavior).

## How to verify

```console
rg -n "func NextBackoff" internal/chatloop/chatloop.go        # expect 0 (deleted 2026-11-04)
rg -n "ComputeDelay|Classify" internal/chatloop/chatloop.go   # expect the nextDelay use + the Rejection-stop gate
rg -n "attempt" internal/matrix/receiver.go internal/signal/receiver.go  # attempt-int threading, reset to 1 on success
rg -n "go-retry" go.mod                                        # DIRECT pin v0.9.0 (no // indirect)
```
