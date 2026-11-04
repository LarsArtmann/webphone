# go-retry — library utilization audit

**Verdict:** partial fit — webphone's one retry-shaped code (the chat reconnect loops) hand-rolls backoff without jitter or retryable classification; the library's bounded-attempts model does not match the unbounded reconnect loops, but its delay computation does. **Adoption: n/a (indirect-only); optional small swap, not scheduled.**
**Pin:** v0.9.0 (go.mod:103, `// indirect`) · **Latest:** v0.9.0 · **Currency:** current · **Status:** indirect-only — pulled by go-paperless (a direct dep) and go-cqrs-lite/middleware
**Audited:** 2026-11-04 · **Prior reports:** none (first audit).

## Status

| | |
|---|---|
| Module | `github.com/larsartmann/go-retry` |
| In webphone | indirect only; **zero imports** |

## What the library offers

`Do(ctx, Config, AttemptFunc, opts...)` / `DoWithValue[T]`; `Config{MaxAttempts=3, InitialDelay=100ms, MaxDelay=5s, Multiplier=2.0, IsRetryable=errorfamily.IsRetryable, DelayFunc, OnRetry, OnExhausted}`; `WithJitter(JitterAdditive)` (exponential + additive jitter), `WithDelayFunc` (e.g. honor Retry-After); exported `ComputeDelay`/`Backoff` for DIY loops; sentinels `ErrExhausted`/`ErrCanceled`/`ErrDeadlineExceeded` (Infrastructure family, unwrap to last attempt). Retryable predicate rides go-error-family by default — the two libraries compose.

## Webphone's own answer (the overlap)

- `internal/chatloop/chatloop.go`: `Sleep(ctx, d)` (:33), `NextBackoff(d, max)` (:45 — double, cap, **no jitter**), `Backoff(ctx, log, msg, err, wait, max)` (:65 — log + wait + double). Consumers: the signal receiver loop (internal/signal/receiver.go:64-77) and the matrix `/sync` loop (internal/matrix/client.go).
- Everything else is deliberately not retried: push transports are fire-and-forget (a lost ring is healed by the next ring), `pbx.Client` rides request timeouts without a loop, `asr` returns Transient for the island to retry at the UX level.
- Note `failsafe-go` v0.9.8 also sits in the graph (indirect via cqrs-htmx/usermgmt) — two retry libraries transitively present, zero directly used.

## Fit analysis

| Aspect | go-retry | webphone loop | Fit |
|---|---|---|---|
| Termination model | bounded attempts (`MaxAttempts`) | unbounded reconnect until ctx cancel | ✗ — a reconnect loop cannot exhaust; `Do` is the wrong shape |
| Delay computation | `ComputeDelay`: exp + **additive jitter**, Retry-After aware | fixed double with cap; **thundering-herd prone** on shared infra (all sessions reconnect simultaneously after a server restart) | ✓ — the real gap |
| Retryable classification | `errorfamily.IsRetryable` predicate | logs every error, retries unconditionally | ✓ — free once the errorfamily seams already classify |
| OnRetry/OnExhausted hooks | yes | inline logging | neutral |

## Ruling

**Small, optional, and honest:** swap `NextBackoff`'s body for `retry.ComputeDelay` (or copy the additive-jitter idea) and gate the loop on `errorfamily.IsRetryable(err)` so Infrastructure errors retry while Rejections stop the loop with a loud log. Full `retry.Do` adoption fits nothing current. Because the module is already in the graph (indirect), the bump cost is a go.mod promote + tidy — the vendorHash ritual is the main tax.

**Refresh trigger:** a second retry-shaped loop (e.g. ASR segment uploads); observable thundering-herd reconnects after server restarts; the chat-network v2 train touching receivers.

## How to verify

```console
rg -n "func NextBackoff|func Backoff" internal/chatloop/chatloop.go
rg -l "larsartmann/go-retry" --type go | wc -l            # expect 0
rg -n "go-retry" go.mod                                   # indirect pin
```
