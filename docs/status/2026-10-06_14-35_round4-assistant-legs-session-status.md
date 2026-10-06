# Round-4 assistant legs executed (D18, D19, D25) — session status

**Session:** 2026-10-06 ~14:05–14:35 CEST · execution of the round-4 plan's three
free-floating assistant legs (D18 auth tail, D19 island-side audit, D25 hygiene minus the
D1.4-gated 25.3); every owner/stack/date/trigger-gated leg left untouched, exactly as the plan orders.

## Verdicts

| # | Item | Proof |
| --- | ------ | ----- |
| 1 | **D18.1 renewal Secure re-issue spec**: past-half-life request under https trusted origins (r.TLS nil, plain-HTTP suite server) asserts the re-issued Set-Cookie carries Secure + same token + live MaxAge; also pins the half-life throttle (young session re-issues NOTHING). Boot helper fix rides along: `NewMemStore(cfg.SessionTTL)` — the hardcoded hour made store rows un-expirable and renewal undrivable; production app.go wires the SQLite store from cfg.SessionTTL, so the helper now matches | `62bc7dc`; focused Ginkgo run "Ran 1 of 9 Specs" PASS (~0.7s = the 700ms sleep); suite `ok` |
| 2 | **D18.2 slog secret-leak grep** over server/userauth/gateway: CLEAN. 22 server call sites log errors/status/family/code/extension/entry names only; the 422-arm logs no provider detail (user banner carries it); userauth logs nothing itself — the slog.Default logger flows to vendored usermgmt, whose auth sites log the cookie NAME (middleware.go), bot id/name (api_token_middleware.go), never values; enroll-token errors are static strings (hash at rest); gateway has zero logging; cmd startup logs carry addr/mode/origins/CRM url only | verdict recorded in the TODO_LIST AUTH TAIL note (committed) |
| 3 | **D18.3 codespell over the 10-06 delta** (CHANGELOG, lessons, all four 10-06 status docs, the round-4 plan, then AGENTS+TODO_LIST too): EXIT 0. The only hits were four `keep-alives` — the twice-adjudicated correct-English plural (three hits literally quote the CHANGELOG line). Ruling made durable: `keep-alives` joined `.codespellrc` ignore-words with the adjudication dated in the comment | `.codespellrc` diff; second run zero output |
| 4 | **D18.4 auth-regression convention** in AGENTS' Sessions bullet: auth-surface changes carry an `auth:` commit prefix + run the session-behavior suite (home of the TLS-fronted Secure mint + sliding-renewal specs) before push. AGENTS 129/377 lines | AGENTS diff |
| 5 | **D19 island-side auth audit**: clean overall. Adoption ladder correct (3 backoff retries 250/500ms, reload LAST resort; adopt re-arms meta + body hx-headers). XSS: zero raw innerHTML assignments anywhere in island/shell/enroll — whoami/snippet-fill/lightbox/login-errors all textContent/value/src (img src + CSP same-origin). localStorage: zero credential persistence (theme/lang/tab/drafts/extension/sink/history/contacts only); the legacy pbx-contacts import deletes the key ONLY after every row is server-accepted, else retries next login (upsert-by-number idempotence). No password/token in any log line. FIXED: `authedFetch` had drifted to an inline meta querySelector — now imports the csrf.js reader (the ONE-home rule; session/passkey/enroll already did). ACCEPTED+RECORDED: passkey/enroll finish carry the one-time ceremony session_key as `?user_id=` (burned at finish, worthless without the WebAuthn assertion + CSRF; exposure = proxy access logs only) — body-move is a cheap option inside D30's decision space, not silently changed | auth.js diff; TODO_LIST ISLAND AUDIT note; island tests 186/186; `go test ./internal/web/...` ok |
| 6 | **D25.1 near-aligned-table sweep** over every 10-05/10-06 status+planning doc: exactly ONE mixed-shape table found (the 03-40 report — rows 1–4 hand-aligned wide, rows 5–6 drifted = the churn-bait middle state). Normalized to the daemon-canonical compact shape (single-space cells, header-sized separator). Detector (mixed column widths + extra cell padding) reports ZERO across the corpus | 03-40 diff; detector output "near-aligned tables remaining: 0" |
| 7 | **D25.2 whitespace-drift staged-diff detector**: `scripts/whitespace-drift.sh` — staged `git diff` vs `-w` divergence over formatter-OWNED files only (Go minus _templ.go/vendor, JS/MJS/CSS/Nix); md deliberately EXCLUDED because formatter-unowned means no drift definition (first draft gated md too — dogfooding on my own 25.1 normalization caught the false-positive class). Self-tested in a scratch repo: whitespace-only Go edit EXIT 1 (with the `nix fmt` fix line), substantive change EXIT 0, md-only "unowned/clean". Wired as an AGENTS Commands habit line (after `git add`, before commit) | script + scratch-repo test transcript; AGENTS Commands line |
| 8 | **Session gates**: `nix fmt` 0 changed · full `go test -count=1 ./...` rc=0 · single checks `format` + `island-lint` + `island-js` all rc=0 (the e62fe34-class insurance, run BEFORE push this time) · `buildflow` EXIT 0 | command transcripts in-session |

## Not done (correctly left alone)

D1 THE SITTING + D2–D6 deploy chain (owner/ssh), D7/D8 (gated on D1/D12), D9–D17 owner legs,
D15/D16 stack lane, D20 (2026-11-05), D21 (2026-12-20), D22 trigger-gated, D23 release-gated,
D24 zombie guard, D29/D30 conditional on D1 verdicts. **D25.3** (aligner-vs-accept-churn
convention note) waits on D1.4's coupling-home verdict per its dependency edge.

## Handoff

The six §g questions still ride D1 (CI-bar subset, push-lag threshold — now six datapoints,
daemon-format coupling home, `__Host-` prefix, enroll token binding vs ULID-secrecy, auth audit
cadence). New for the owner from THIS session: the session_key query-param acceptance is
recorded next to the D30 option; `__Host-` evaluation (g.1) unchanged. Everything
assistant-executable in the round-4 plan is now closed — the next assistant legs open only as
D1 verdicts unlock them (D7 paper closes, D29 conditional implementation).
