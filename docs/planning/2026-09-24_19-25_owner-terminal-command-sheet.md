# Owner-terminal command sheet (2026-09-24, v2.7.0 cycle)

Everything an assistant cannot run itself: pbx-artmann AGENTS forbids
assistant ssh/deploy, and host-root operations are owner-only. Commands
are copy-paste ready; run in order. Item 0 is URGENT — the host's nix
is down until it runs.

## 0. URGENT: restore host nix (binfmt symlink vanished)

`/run/binfmt` disappeared (~19:10 on 2026-09-24) while the kernel
still has aarch64 emulation registered; EVERY nix build on this host —
`nix develop`, `nix build`, VM tests, release gates — fails with
`getting attributes of path "/run/binfmt": No such file or directory`.
Client-side options (extra-platforms, sandbox) are restricted for
untrusted users, so only root heals it:

```console
systemctl restart systemd-binfmt.service
ls -la /run/binfmt   # must exist again (symlink → /proc/sys/fs/binfmt_misc)
```

## 1. Deploy the released chain to prod (M1, TODO row "Deploy the released chain")

prod serves v2.5.0 today. After v2.7.0 is cut the chain is
READY (webphone tag `v2.7.0` → stack `<STACK_HASH>` → pbx-artmann
re-pin pending, see §2):

```console
cd ~/projects/pbx-artmann && nixos-rebuild test --flake .#pbx --target-host root@pbx.artmann.tech
# then, if the smoke below is green:
nixos-rebuild switch --flake .#pbx --target-host root@pbx.artmann.tech
```

## 2. pbx-artmann relock ritual (AFTER the v2.7.0 stack bump landed)

Per docs/release-runbook.md — ALWAYS `git rev-parse`, never typed:

```console
cd ~/projects/nix-international-telephony && git rev-parse HEAD   # <- new stack rev
cd ~/projects/pbx-artmann
# swap the rev in flake.nix's telephony URL, then:
nix flake update telephony
nix run .#lock-drift-probe
nix build .#nixosConfigurations.pbx.config.system.build.toplevel
nix build .#nixosConfigurations.pbx-aarch64.config.system.build.toplevel  # NOT --system
# verify the webphone unit's ExecStart store path MOVED vs the last build, then commit
```

## 3. Post-deploy verification (M1 tail, from the webphone repo)

```console
python3 scripts/webphone-smoke.py --base https://pbx.artmann.tech --expect-version 2.7.0
```

Then the live self-send check (needs a real extension session):
sign in, self-send an SMS to the PBX DID — with 2.7.0 the send is
refused LOCALLY (train C): expect the 422 banner
"messages cannot be sent to your own number" AND the failed bubble in
the thread (evidence row). On v2.6.0 the provider's 40310 text was the
expectation; that changed with train C.

## 4. Outbound SMS bridge failure on prod (M3, TODO row 40)

webphone-side classification is correct since `1d53f44`/`6ac8962`; the
root cause is stack-side. On the pbx host:

```console
journalctl -u telnyx-webhooks.service --since today | grep -Ei 'sms|422|error'
# restart or fix creds per findings, then send a test SMS; record the root cause
# in TODO_LIST row 40 + the stack runbook
```

## 5. OWNER-calls batch session (M16, TODO row 42)

One sitting, briefing ready at
`docs/planning/2026-09-22_13-50_owner-calls-briefing.md` (~15 original
+ 3 CRM + the newer g1/g2/g3/art-dupl/webhook/idem additions).
Decisions land back into TODO_LIST.

## 6. Release announcements (M17, TODO row 43)

Drafts for v2.1.0–v2.3.0, v2.5.0, v2.6.0 live at `docs/announcements/`
(v2.7.0 draft to be added post-release). Owner picks channels,
approves wording, decides the disclosure posture.
