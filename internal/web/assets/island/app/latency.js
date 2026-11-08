import { authedFetch } from "./auth.js";
import { metricsBeacon } from "./config.js";

/**
 * sendLatencyBeacon posts one fire-and-forget island latency measurement
 * to the server's beacon intake (the measure-first latency design). The
 * server folds it into the webphone_island_latency_seconds histograms.
 *
 * Instrumentation never alerts on itself: a seam-off server (no
 * PBX_CONFIG.metrics, the route 404ing), a flood budget, a network
 * failure — every failure path is silently dropped, never retried, and
 * the promise resolves swallowed. The beacon rides authedFetch because
 * the route is session-gated like every island endpoint.
 *
 * @param {IslandLatencyInterval} interval the measured interval's name
 * @param {Record<string, number>} legsMS the interval's named legs in
 *   milliseconds (the server folds each leg into its own labeled
 *   histogram; unknown leg names are dropped server-side)
 * @param {number} [segmentBytes] the ASR segment's audio size — the
 *   size context every round-trip percentile needs
 * @returns {void}
 */
export function sendLatencyBeacon(interval, legsMS, segmentBytes) {
  if (!metricsBeacon) return;
  /** @type {{ interval: IslandLatencyInterval, legs_ms: Record<string, number>, bytes?: number }} */
  const body = { interval, legs_ms: legsMS };
  if (typeof segmentBytes === "number" && segmentBytes >= 0) {
    body.bytes = segmentBytes;
  }
  authedFetch("/api/metrics/latency", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(body),
    keepalive: true,
  }).catch(() => {});
}
