// The ended card with the agent seam ON but ASR OFF: no transcript can
// exist for the call, so the summarize affordance must not render —
// today it 404s into an announce after the click (the defect the
// Summarize micro-cluster pins). The agent gate alone is not enough:
// the transcript seam is the producer of the text the agent summarizes.
import test from "node:test";
import assert from "node:assert/strict";

import { installSummarySeam, settleEnded } from "./calls-summary-seam.mjs";

const { placeCall, sessions, state } = await installSummarySeam({
  id: "call-noasr",
  config: { agent: true, sipDomain: "sip.test" },
});

test("ended card with agent on but ASR off: Close only, no summarize row", async () => {
  state.userAgent = {};
  await placeCall("+4930123456");
  const entry = sessions.get("call-noasr");
  await settleEnded(entry);

  assert.equal(sessions.has("call-noasr"), false);
  assert.ok(entry.dom.parent, "the ended card stays visible");
  assert.ok(entry.dom.classList.contains("wp-ended-card"));
  assert.equal(
    entry.dom.querySelector(".summarize-btn"),
    null,
    "no summarize affordance without the transcript seam — the button cannot work",
  );
  assert.ok(entry.dom.querySelector(".close-ended-btn"), "Close remains");
});
