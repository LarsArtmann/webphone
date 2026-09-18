# TODO_LIST

Short- and mid-term improvement tasks. Done work is deleted, never
struck through (docs-health style: one home per fact, no decay).

| Task                                                                                                                               | Status      | Priority | Effort | Evidence / notes                                                                                     |
| ---------------------------------------------------------------------------------------------------------------------------------- | ----------- | -------- | ------ | ---------------------------------------------------------------------------------------------------- |
| Re-run the consuming stack's full gate (`nix flake check` + `legacyPackages.telephony-browser`) against this input and record the run | 🔴 `TODO`   | High     | M      | Extraction is verified against the upstream suites at wire-up time; keep the proof with the wire-up commit |
| Manual theme toggle (override `prefers-color-scheme`, persist choice)                                                              | 🔴 `TODO`   | Low      | S      | Tokens already centralized in `src/style.css`; add `data-theme` hook + settings row                   |
| sip.js 0.22 evaluation (update.sh exists; API deltas unknown)                                                                       | 🔴 `TODO`   | Low      | M      | Bundle contract strings must survive; run upstream VM suite after bump                                |
| Keyboard shortcuts (answer, hangup, mute, hold) with visible hints                                                                  | 🔴 `TODO`   | Low      | S      | Accessibility win; no DOM-contract impact                                                             |
| Media-key / headset button support via `navigator.mediaDevices` + MediaSession                                                      | 🔴 `TODO`   | Low      | M      | After keyboard shortcuts prove the plumbing                                                          |
