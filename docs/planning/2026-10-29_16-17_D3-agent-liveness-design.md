# D3 agent liveness — design input (2026-10-29 16:17)

**Status:** DESIGN ONLY — owner-GO + security-review gated (the standing D3
posture). Nothing here is wired. Source: the ai-phone steal-list L6 row
(TODO_LIST) + the ai-phone relay-mode prior art, mapped onto webphone's
seams. Task M18 / F75–F77 of the Pareto plan.

## 1. The gap in webphone today

The textagent seam answers "is the agent REACHABLE" (the bounded 3 s
`fillAgentMode` probe in `session_api.go`, fail-open; the 60 s HTTP timeout
in `textagent.Client`) and "did the ask FAIL" (the `agent.summarize` 502
toast). Neither answers **"accepted then silent"**: the agent took the
request, never answered, and the human waits forever with no surface and no
recovery. That is the class the D3 lane must design for.

## 2. The liveness ladder (adapted from ai-phone relay mode)

ai-phone's numbers, kept as the starting calibration (F75):

| t (since accept) | action | webphone adaptation |
|------|--------|---------------------|
| 1.5 s | first filler ("one moment") | the DRAFT lane already shows a pending pill; the CALL lane would speak it |
| 10 s | second filler (status honesty) | same voice line, distinct text so the caller hears progress |
| 25 s | polite goodbye + BYE | the ladder TERMINATES the call rather than hanging up silently — the anti-goal is an open channel with a dead brain |
| 30 s + 10 s | (hangup op) two-stage wait, then named failure | maps to the agent-session teardown below |

Two webphone deltas, stated openly:

1. **t0 differs.** ai-phone's rule — "t0 is the last speech, the silence is
   INSIDE the budget" — transfers verbatim: the ladder clock restarts on
   every agent utterance, not on call start. A ladder driven off call-start
   would fire mid-conversation.
2. **BYE authority.** In webphone the BROWSER island owns the SIP dialog;
   a server-side answering agent (the D3 end-state) terminates calls at the
   STACK (FreeSWITCH executes transfers and hangs up server-side — the M19
   transfer spike's finding). The ladder's goodbye+BYE is therefore a
   stack-side action with a webphone-issued instruction, never an island
   action: the island may be the stalled party.

## 3. Stalled-client slot release (F76)

ai-phone frees the listener slot on socket EOF so the next call falls back
to autonomous mode. webphone's equivalent resources and their release
triggers:

- **Agent call slot** (one concurrent agent call per extension, the D3 v1
  shape): released on SIP BYE (fact from the stack) AND on a supervisor
  watchdog — a slot whose ladder reached the goodbye stage is dead even if
  the stack never reports BYE. The watchdog is the lesson from the island's
  own connection.js: 0.x reconnects hang; recovery paths set state
  explicitly.
- **Draft queue slots** (the existing textagent review lane): already
  bounded; the missing piece is the deadline — a draft accepted-but-silent
  past the ladder's goodbye threshold should surface as `ReviewFailed` (the
  state exists, `domain.ReviewState`) instead of pending forever. This
  doubles as the F89 review-state SSE nudge's first real consumer.

**Named exit codes** (the ai-phone hangup contract, adapted): every
teardown path names itself — `agent.ladder_filler` (progress markers),
`agent.ladder_goodbye` (the 25 s termination), `agent.slot_watchdog`
(the supervisor killed it), `agent.draft_deadline` (draft lane). Codes
register in the same commit per the registry-freshness rule.

## 4. webphone seam mapping (F77)

| ai-phone concept | webphone seam today | D3 v1 change |
|------------------|---------------------|--------------|
| relay liveness probe | `fillAgentMode` 3 s bounded probe (fail-open) | reuse as the heartbeat primitive; a failed probe must FLIP the pill (today it silently drops the pill) |
| filler scheduler | none | a ladder table in the textagent client context, mirroring the machine-table doctrine (rows, not branches) |
| slot release on EOF | none (draft lane is stateless HTTP) | the D3 supervisor owns slots; draft lane gains the deadline above |
| hangup exit codes | `textagent.http` + `agent.summarize` | the four codes above; family pins in `textagent/family_test.go` |
| BYE authority | island-owned SIP | stack-side BYE on webphone instruction (M19 spike's server-side execution finding) |

## 5. What this doc deliberately does NOT decide

- Whether D3 answers calls at all (owner-GO + security-review — the
  standing gate; an autonomous voice agent on a phone line is a product
  decision, not an engineering one).
- The ladder's exact seconds for the DRAFT lane (25 s goodbye does not
  translate; a draft deadline wants minutes, not seconds — proposed
  default: 10 min, surfaced as `ReviewFailed` + the F89 nudge).
- Voice synthesis/filler audio (out of scope; the stack owns media).

*Prior art: oxnw/ai-phone relay mode (the steal-list deep dive,
`docs/research/2026-10-29_ai-phone-deep-dive.html`).*
