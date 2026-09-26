# Owner-terminal command sheet (2026-09-24, v2.7.0 cycle)

Everything an assistant cannot run itself: pbx-artmann AGENTS forbids
assistant ssh/deploy, and host-root operations are owner-only. Commands
are copy-paste ready; run in order. Item 0 was URGENT — the host's nix
is down until it runs (as of 2026-09-26 it is healed, see the STATUS
note in §0).

## 0. URGENT: restore host nix (binfmt directory vanished; UPDATED 2026-09-25 with the verified root cause)

> STATUS 2026-09-26: healed on this host — `/run/binfmt/aarch64-linux`
> symlink in place (planted 2026-09-25 05:32) and `nix build
> nixpkgs#hello` verified green 2026-09-26; skip to §1. The DURABLE fix
> below stays open: the hand-rolled store-path pin in `nix.conf` rots
> on the next GC of that path or a reboot without tmpfiles rules.

`/run/binfmt` disappeared (~19:10 on 2026-09-24) while the kernel
still has aarch64 emulation registered; EVERY nix build on this host —
`nix develop`, `nix build`, VM tests, release gates — fails with
`getting attributes of path "/run/binfmt": No such file or directory`.
Client-side options (extra-platforms, sandbox) are restricted for
untrusted users, so only root heals it.

Root cause (verified 2026-09-25 ~04:30, cross-checked with the stack
session's forensics): the kernel binfmt entry `aarch64-linux` names
`/run/binfmt/aarch64-linux` as its interpreter, `/etc/nix/nix.conf`
pins `extra-sandbox-paths = /run/binfmt /nix/store/31rksc1wkr54dadg615g1qd0rxk9gnk6-qemu-aarch64-binfmt-P`
— but this generation's tmpfiles.d carries NO binfmt rules, and
`systemd-binfmt.service` "finished OK" at the 22:35 reboot without
creating anything. **A service restart heals nothing.** Plant the
interpreter symlink by hand (paths verified live on this host):

```console
sudo mkdir -p /run/binfmt
sudo ln -s /nix/store/31rksc1wkr54dadg615g1qd0rxk9gnk6-qemu-aarch64-binfmt-P/bin/qemu-aarch64-binfmt-P /run/binfmt/aarch64-linux
ls -la /run/binfmt        # must show aarch64-linux -> the qemu binary
nix build nixpkgs#hello  # sanity: any nix build must work again
```

Durable fix (pick one, host config): switch the host to
`boot.binfmt.emulatedSystems = [ "aarch64-linux" ]` so nixpkgs owns
the tmpfiles rules (survives reboots), or drop aarch64 emulation and
remove `/run/binfmt` from `extra-sandbox-paths`. Rot risk while the
hand-rolled pin stays: the hard store path in nix.conf breaks builds
again whenever a GC collects it.

## 1. Deploy the released chain to prod (M1, TODO row "Deploy the released chain")

prod served v2.5.0 at drafting time; the 2.6.0 chain was DEPLOYED
early 2026-09-25 (prod `/version` = v2.6.0, probed and recorded in
TODO_LIST) — so the deploy-ordering question is settled: go STRAIGHT
to 2.7.0 once tagged, no stepping stone. After v2.7.0 is cut the
chain is READY (webphone tag `v2.7.0` → stack `<STACK_HASH>` →
pbx-artmann re-pin pending, see §2):

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

One more 2.7.0-visible check: type a shared-directory contact's name
into the dial destination — the typeahead must offer it again (the
capitalized-wire bug silently hid every shared contact; fixed by
`e43fea8`).

## 4. Outbound SMS bridge failure on prod (M3, TODO row "Outbound SMS bridge failure on prod")

webphone-side classification is correct since `1d53f44`/`6ac8962`; the
root cause is stack-side. On the pbx host:

```console
journalctl -u telnyx-webhooks.service --since today | grep -Ei 'sms|422|error'
# restart or fix creds per findings, then send a test SMS; record the root cause
# in the TODO_LIST SMS-bridge row + the stack runbook
```

## 5. OWNER-calls batch session (M16, TODO row "OWNER-calls batch session")

One sitting, briefing ready at
`docs/planning/2026-09-22_13-50_owner-calls-briefing.md` (~15 original
+ 3 CRM + the newer g1/g2/g3/art-dupl/webhook/idem additions).
Decisions land back into TODO_LIST.

## 6. Release announcements (M17, TODO row "Post the release announcements")

Drafts for v2.1.0–v2.3.0, v2.5.0, v2.6.0 AND v2.7.0 live at
`docs/announcements/` (v2.7.0 = `2026-09-24_v2-7-0_drafts.md`, `357ffec`).
Owner picks channels, approves wording, decides the disclosure posture.
