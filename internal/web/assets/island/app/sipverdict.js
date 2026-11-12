// sipverdict owns the re-INVITE verdict discipline: sip.js resolves
// session.invite() when the re-INVITE is SENT, and the verdict arrives
// only through the request delegate's callbacks — so any caller that
// awaits invite() bare settles its state on send, showing a state the
// network has not confirmed (the island honesty contract's oldest gap:
// the hold/resume toggle). This module is the one place that waits.
//
// Interface: one function — send this re-INVITE, get one of four
// settled verdicts. The sip.js talk (delegates, send-failure promise
// rejections) absorbs here; callers consume settled verdicts only.
//
// Timeout: a re-INVITE that gets no answer at all would otherwise hang
// the pending state forever; the verdict lane resolves "timeout" after
// verdictTimeout and the caller keeps its prior settled state (the
// same failed-toggle contract the optimistic send rides).

/** Milliseconds to wait for a verdict before giving the caller "timeout". */
export const verdictTimeout = 30_000;

/**
 * Send one re-INVITE and resolve with its settled verdict.
 *
 * @param {{ invite: Function }} session the established call's session.
 * @returns {Promise<"accepted"|"rejected"|"send-failed"|"timeout">}
 */
export function inviteVerdict(session) {
  return new Promise((resolve) => {
    let settled = false;
    const settle =
      /** @param {"accepted"|"rejected"|"send-failed"|"timeout"} verdict */ (
        verdict,
      ) => {
        if (settled) return;
        settled = true;
        resolve(verdict);
      };
    const timer = setTimeout(() => settle("timeout"), verdictTimeout);
    const done =
      /** @param {"accepted"|"rejected"|"send-failed"|"timeout"} verdict */ (
        verdict,
      ) => {
        clearTimeout(timer);
        settle(verdict);
      };
    session
      .invite({
        requestDelegate: {
          onAccept: () => done("accepted"),
          onReject: () => done("rejected"),
        },
      })
      .catch(() => {
        // The send itself failed (dialog gone, session ended under
        // us) — a distinct verdict from a refused re-INVITE, same
        // caller outcome: keep the prior settled state.
        done("send-failed");
      });
  });
}
