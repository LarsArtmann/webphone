// ICE/media diagnostics for the focused call: live stats (selected
// candidate pair types, RTT, loss, jitter, codec) plus plain-language
// hints — NAT/ICE is this stack's #1 failure mode, so the panel says
// WHAT to suspect instead of only showing numbers.

import { t } from "./i18n.js";
import { sessions, state } from "./state.js";
import { els } from "./ui.js";

function candidateTypeOf(stats, candidateId) {
  const candidate = stats.get(candidateId);
  return candidate ? `${candidate.candidateType || "?"}` : "?";
}

function iceHints(summary) {
  const hints = [];
  const pair = summary.selectedPair ? summary.selectedPairData : null;
  if (summary.localType === "relay" || summary.remoteType === "relay") {
    hints.push(t("iceRelay"));
  } else if (summary.localType === "srflx" || summary.remoteType === "srflx") {
    hints.push(t("iceSrflx"));
  } else if (summary.localType === "host") {
    hints.push(t("iceHost"));
  } else if (
    (pair && pair.state === "failed") ||
    summary.iceState === "failed"
  ) {
    hints.push(t("iceFailed"));
  } else {
    hints.push(t("iceNoMedia"));
  }
  if (summary.packetsLost > 50) hints.push(t("iceLoss")(summary.packetsLost));
  return hints;
}

async function updateIcePanel() {
  if (!state.focusedId) return;
  const entry = sessions.get(state.focusedId);
  const pc =
    entry && entry.session.sessionDescriptionHandler
      ? entry.session.sessionDescriptionHandler.peerConnection
      : null;
  if (!pc) return;
  let stats;
  try {
    stats = await pc.getStats();
  } catch {
    return;
  }
  const summary = {
    iceState: pc.iceConnectionState,
    selectedPair: null,
    selectedPairData: null,
    localType: "",
    remoteType: "",
    rtt: null,
    packetsLost: 0,
    jitter: null,
    codec: "",
    bytesReceived: 0,
  };
  for (const stat of stats.values()) {
    if (stat.type === "transport" && stat.selectedCandidatePairId) {
      summary.selectedPair = stat.selectedCandidatePairId;
    }
    if (
      stat.type === "candidate-pair" &&
      (stat.selected || stat.nominated) &&
      stat.state === "succeeded"
    ) {
      summary.selectedPair = stat.id;
    }
    if (stat.type === "inbound-rtp" && stat.kind === "audio") {
      summary.packetsLost = stat.packetsLost || 0;
      summary.jitter = stat.jitter;
      summary.bytesReceived = stat.bytesReceived || 0;
    }
  }
  if (summary.selectedPair) {
    const pair = stats.get(summary.selectedPair);
    if (pair) {
      summary.selectedPairData = pair;
      summary.rtt = pair.currentRoundTripTime ?? null;
      summary.localType = candidateTypeOf(stats, pair.localCandidateId);
      summary.remoteType = candidateTypeOf(stats, pair.remoteCandidateId);
    }
  }
  for (const stat of stats.values()) {
    if (
      stat.type === "codec" &&
      stat.mimeType &&
      stat.mimeType.startsWith("audio")
    ) {
      summary.codec = stat.mimeType.replace("audio/", "");
      break;
    }
  }
  const lines = [
    `ice: ${summary.iceState}  path: ${summary.localType || "?"} → ${summary.remoteType || "?"}`,
    `codec: ${summary.codec || "?"}  rtt: ${summary.rtt != null ? `${Math.round(summary.rtt * 1000)} ms` : "—"}`,
    `received: ${summary.bytesReceived} B  lost: ${summary.packetsLost}  jitter: ${summary.jitter != null ? `${Math.round(summary.jitter * 1000)} ms` : "—"}`,
  ];
  const hints = iceHints(summary).map((hint) => `· ${hint}`);
  els.icePanel.replaceChildren(
    ...[...lines, ...hints].map((line) => {
      const div = document.createElement("div");
      if (line.startsWith("·")) div.className = "hintline";
      div.textContent = line;
      return div;
    }),
  );
}

let iceTimer = null;

export function startIcePanel() {
  if (iceTimer) return;
  updateIcePanel();
  iceTimer = setInterval(updateIcePanel, 2000);
}

export function stopIcePanel() {
  if (!iceTimer) return;
  clearInterval(iceTimer);
  iceTimer = null;
  els.icePanel.replaceChildren();
}

// calls.js announces session-table changes as an event because calls and
// ice must never import each other (module-graph invariant, enforced by
// the arch test); main.js imports this module for the listener.
document.addEventListener("wp:calls-changed", () => {
  els.iceWrap.hidden = sessions.size === 0;
  if (sessions.size > 0) startIcePanel();
  else stopIcePanel();
});
