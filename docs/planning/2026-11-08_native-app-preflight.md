# Native-app preflight — go/no-go for the first device train (2026-11-08)

**Status:** PREFLIGHT (read-only; nothing mutated in `~/projects/webphone-android`)
· **Mandate:** P3-8 of the post-v9 execution plan — confirm the Android repo
is quiet, evaluate the two KMP deferral triggers, and list what the first
device train needs.

## 1. Is the repo quiet? — YES

- `git status`: **clean** (zero modified/untracked files).
- Last commit `246fa4e` 2026-11-05 19:43 ("chore: bump CI setup-java to v6.2.0
  and refresh nixpkgs flake input") — a CI-hygiene commit, not a code train.
- 2 commits in the last 7 days, both chore/docs class.
- No concurrent session markers (no daemon `chore:` commits newer than a day,
  no mid-edit compile states).
- Open Dependabot PRs exist (4 branches: AGP 9.4.1, firebase-messaging 26.0.1,
  okhttp 5.5.0, gradle-wrapper 9.8.2) — these are the repo's own housekeeping,
  mergeable independently of a device train.

Verdict: **quiet enough** for Gradle/KMP surgery per the 2026-10-28 deferral
wording ("execute when that repo is quiet"). The deferral's OTHER half is not
met — see §2.

## 2. The two KMP deferral triggers — HALF-MET, still deferred

The 2026-10-28 deferral gates the KMP protocol-layer spike on BOTH:
(a) the android repo is quiet — **MET today**; (b) a committed second target
in ROADMAP — **NOT MET** (no `linuxX64`/second-target entry anywhere in the
repo's ROADMAP or Gradle config; the check errored on no match).

**Verdict: the spike stays DEFERRED**, correctly — the gate requires both
triggers, and half a gate is no gate. The spike is also Tier-0 by its own
wording ("may stay unstarted indefinitely"). No action owed; re-evaluate only
when someone actually wants a second target (the Linux WebRTC 3-call audio
smoke remains the weak leg that made full KMP a NO).

## 3. What the first device train needs (the checklist)

From the README honesty markers + TODO §Native-app train (webphone side
closed; both remaining legs live OUTSIDE this repo):

**Shell leg (webphone-android):**
1. SIP UA vs the LIVE WSS gateway (Caddy-fronted `/sip`): REGISTER, INVITE,
   BYE, DTMF (`application/dtmf-relay`, EQUALS form), hold re-INVITE — all
   implemented, all UNTESTED against FreeSWITCH.
2. Incoming INVITE answering + REFER/NOTIFY transfer — implemented, explicitly
   UNTESTED.
3. Real CallStyle notification (v0.1 posts the FSI heads-up only).
4. Origin-picker settings UI (`App.origin` is a constant today).
5. `/api/session` field-name check against the served `sessionIdentity` shape
   (the server contract is pinned in `internal/server`; the Kotlin side should
   assert the same field names).

**Stack leg (nix-international-telephony):**
6. The FS dialplan must POST `webphone:/hooks/ring` (Bearer webhook secret;
   `{owner, from, call_id, state: ring|end}`) on inbound ring + ring-end —
   until it lands, the push seam is inert (fire-and-forget 202).
7. Expose `services.telephony.webphone.push.*` in the stack deploy docs.

**Sequencing note:** the stack leg (6) is the cheaper of the two and unlocks
ring-while-closed E2E the moment the shell registers a push device; the shell
legs (1–2) need the operator on a real device with the stack's WSS reachable
— one training session covers 1, 2, 3 together.

## 4. Nothing was mutated — verification

- `git status --short` in `~/projects/webphone-android` → empty (re-checked
  after the preflight; identical to before).
- No branch created, no checkout, no Gradle invocation, no file read beyond
  `ls`, `git log/status`, `settings.gradle.kts`, and the README head.
