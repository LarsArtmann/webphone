// The ended card with the agent seam OFF (M12): the post-call row is
// Close only — no summarize affordance exists, and nothing about the
// card leaks the seam. No PBX_CONFIG at all: the strictest seam-off boot.
import test from "node:test";
import assert from "node:assert/strict";

import { installSummarySeam, settleEnded } from "./calls-summary-seam.mjs";

const { placeCall, sessions, state } = await installSummarySeam({
  id: "call-noagent",
});

test("ended card with the seam off: Close only, no summarize row", async () => {
  state.userAgent = {};
  await placeCall("+4930123456");
  const entry = sessions.get("call-noagent");
  await settleEnded(entry);

  assert.equal(sessions.has("call-noagent"), false);
  assert.ok(entry.dom.parent, "the ended card stays visible");
  assert.ok(entry.dom.classList.contains("wp-ended-card"));
  assert.equal(
    entry.dom.querySelector(".summarize-btn"),
    null,
    "the summarize affordance does not exist without the seam",
  );
  assert.ok(entry.dom.querySelector(".close-ended-btn"), "Close remains");
});
