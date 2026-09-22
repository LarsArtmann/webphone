// SIP connection: user agent, registerer, and the reconnect watchdog.
//
// SIP.js 0.21's userAgent.reconnect() can hang forever after a transport
// loss (observed by the browser E2E reconnect drill even with the server
// reachable again). Every attempt is therefore bounded; a hung one gets
// a full teardown-and-rebuild instead of an eternal "try N" pill.
// A registration lost AFTER it was established gets the same rebuild:
// retrying register() on the dead Registerer never recovers (the 1001
// E2E anomaly: sofia said user_not_registered, the island kept its
// dead Registerer forever).

import { setCredentials, clearCredentials, getCredentials } from "./auth.js";
import { iceServers, sipDomain, websocketUrl } from "./config.js";
import { ringbackStop, ringToneStart, ringToneStop } from "./audio.js";
import { recordHistory } from "./panels.js";
import { titleFlashStart, titleFlashStop, notifyIncoming } from "./notify.js";
import { t } from "./i18n.js";
import { sessions, state } from "./state.js";
import { announce, els, log, setRegStatus } from "./ui.js";

let registerer = null;
let reconnectAttempts = 0;
let reconnectTimer = null;
let reconnectCycleTimer = null;
let stopping = false;
// True while a wedged user agent is being torn down and rebuilt;
// suppresses the teardown's own disconnect/unregistered events.
let resetting = false;

const RECONNECT_ATTEMPT_TIMEOUT_MS = 5000;
// Hard ceiling for one whole reconnect cycle, watched by an
// INDEPENDENT timer: the 2026-09-22 E2E run observed the attempt's
// own withTimeout chain going silent while the page stayed alive
// (pill frozen mid-cycle, sofia 200-OK'd the re-REGISTER, no further
// SIP or pill updates) — the cycle deadline force-rebuilds regardless
// of the wedged chain.
const RECONNECT_CYCLE_DEADLINE_MS = 15000;

export async function connect(extension, password) {
  stopping = false;
  reconnectAttempts = 0;
  setCredentials(extension, password);
  await buildConnection();
}

function scheduleReconnect() {
  if (stopping || reconnectTimer) return;
  reconnectAttempts += 1;
  const delay = Math.min(30, 2 ** reconnectAttempts);
  setRegStatus("status-offline", t("reconnecting")(delay, reconnectAttempts));
  log(`transport lost; reconnect try ${reconnectAttempts} in ${delay}s`);
  reconnectTimer = setTimeout(() => {
    reconnectTimer = null;
    attemptReconnect();
  }, delay * 1000);
}

function withTimeout(promise, ms, label) {
  let timer;
  return Promise.race([
    promise.then(
      (value) => {
        clearTimeout(timer);
        return value;
      },
      (err) => {
        clearTimeout(timer);
        throw err;
      },
    ),
    new Promise((_, reject) => {
      timer = setTimeout(() => reject(new Error(label)), ms);
    }),
  ]);
}

// The independent cycle watchdog: armed when a reconnect attempt
// starts, cleared the moment registration succeeds or a rebuild takes
// over. If the attempt chain wedges without settling, THIS timer still
// fires and rebuilds the agent.
function armReconnectCycleDeadline() {
  clearReconnectCycleDeadline();
  reconnectCycleTimer = setTimeout(() => {
    reconnectCycleTimer = null;
    if (stopping || resetting) return;
    log("reconnect cycle exceeded its deadline — forcing rebuild");
    rebuildConnection("reconnect cycle deadline exceeded").catch((err) => {
      log(`rebuild failed: ${err.message}`);
      scheduleReconnect();
    });
  }, RECONNECT_CYCLE_DEADLINE_MS);
}

function clearReconnectCycleDeadline() {
  if (reconnectCycleTimer) {
    clearTimeout(reconnectCycleTimer);
    reconnectCycleTimer = null;
  }
}

// Tear the wedged agent down and build a fresh one (the page-reload
// recovery path without losing the UI state).
async function rebuildConnection(reason) {
  if (resetting) {
    log(`reconnect watchdog: ${reason} — rebuild already in progress`);
    return;
  }
  log(`reconnect watchdog: ${reason} — rebuilding connection`);
  resetting = true;
  // The transient rebuilding pill: the watchdog just decided to tear
  // the agent down — the user sees WORK happening before the fresh
  // REGISTER lands (or fails) and flips the pill for real.
  setRegStatus("status-offline", t("regRebuilding"));
  // A rebuild supersedes every pending recovery rhythm: the reconnect
  // timer AND the cycle deadline belong to the fresh agent from here.
  if (reconnectTimer) {
    clearTimeout(reconnectTimer);
    reconnectTimer = null;
  }
  clearReconnectCycleDeadline();
  const old = state.userAgent;
  state.userAgent = null;
  registerer = null;
  try {
    await withTimeout(
      old ? old.stop() : Promise.resolve(),
      3000,
      "stop timed out",
    );
  } catch (err) {
    log(`old agent stop: ${err.message}`);
  }
  try {
    await withTimeout(
      buildConnection(),
      RECONNECT_ATTEMPT_TIMEOUT_MS,
      "rebuild timed out",
    );
  } finally {
    resetting = false;
  }
}

async function attemptReconnect() {
  armReconnectCycleDeadline();
  try {
    await withTimeout(
      (async () => {
        await state.userAgent.reconnect();
        await registerer.register();
      })(),
      RECONNECT_ATTEMPT_TIMEOUT_MS,
      "reconnect timed out",
    );
    reconnectAttempts = 0;
    clearReconnectCycleDeadline();
    // sip.js fires NO stateChange when the Registerer never left
    // Registered (a transport loss does not demote it), so the
    // listener cannot refresh the pill here — without this explicit
    // set it shows the last backoff state forever while the phone is
    // fully re-registered (the 2026-09-22 E2E runs read that stale
    // pill as "stuck" and fell back to reloads).
    setRegStatus("status-registered", t("registered"));
    if (sessions.size > 0) {
      log(t("reconnectPreserved")(sessions.size));
    }
    log("transport reconnected; re-registered");
  } catch (err) {
    log(`reconnect failed: ${err.message}`);
    // A hung attempt or a Terminated registerer leaves the agent
    // unusable (retrying register() on a Terminated registerer throws
    // forever); rebuild it.
    const unusable =
      err.message.includes("timed out") ||
      registerer?.state === SIP.RegistererState.Terminated;
    if (unusable) {
      try {
        await rebuildConnection(err.message);
        reconnectAttempts = 0;
        return;
      } catch (resetErr) {
        log(`rebuild failed: ${resetErr.message}`);
      }
    }
    scheduleReconnect();
  }
}

async function buildConnection() {
  const credentials = getCredentials();
  if (!credentials) throw new Error("not signed in");
  const { extension, password } = credentials;
  const uri = SIP.UserAgent.makeURI(`sip:${extension}@${sipDomain}`);
  if (!uri) throw new Error(`invalid extension "${extension}"`);

  state.userAgent = new SIP.UserAgent({
    uri,
    authorizationUsername: extension,
    authorizationPassword: password,
    transportOptions: { server: websocketUrl },
    sessionDescriptionHandlerFactory:
      SIP.Web.defaultSessionDescriptionHandlerFactory(),
    sessionDescriptionHandlerFactoryOptions: {
      peerConnectionConfiguration: { iceServers },
    },
    // Built-in logger sends SIP-stack warnings (transport failures,
    // malformed responses) to the browser console; without it the
    // console is silent exactly when the connection misbehaves.
    logBuiltinEnabled: true,
    logLevel: "warn",
    delegate: {
      onDisconnect: (error) => {
        if (stopping || resetting) return;
        // Surface WHY the transport died — the pill is the first place
        // a user looks when audio goes quiet; "offline" alone hides
        // certificate/TLS vs network failures.
        const reason = error
          ? `offline: ${error.message || error}`
          : t("offline");
        setRegStatus("status-offline", reason);
        if (error) scheduleReconnect();
      },
      onInvite: (invitation) => {
        if (state.incomingSession) {
          invitation.reject();
          log("rejected second incoming call", "warn");
          announce(t("rejectedSecond"), "warn");
          return;
        }
        state.incomingSession = invitation;
        const from = (invitation.remoteIdentity &&
          invitation.remoteIdentity.uri) || {
          user: "unknown",
        };
        els.incomingFrom.textContent = from.user || "unknown";
        els.incoming.hidden = false;
        // Incoming-call UX: system notification, audible ring, tab flash.
        notifyIncoming(from.user || "unknown");
        ringToneStart();
        titleFlashStart();
        invitation.stateChange.addListener((state2) => {
          if (state2 === SIP.SessionState.Terminated && !els.incoming.hidden) {
            els.incoming.hidden = true;
            state.incomingSession = null;
            ringToneStop();
            titleFlashStop();
            // The far end gave up before the user answered: say so and
            // keep the attempt in Recent calls (dir in, no duration).
            const missed = from.user || "unknown";
            log(`missed call from ${missed}`, "warn");
            announce(t("missedCall")(missed), "warn");
            recordHistory({
              dir: "in",
              target: missed,
              at: Date.now(),
              dur: 0,
            });
          }
        });
        log(`incoming call from ${from.user}`);
      },
    },
  });

  await state.userAgent.start();

  registerer = new SIP.Registerer(state.userAgent);
  // Per-registerer "ever reached Registered" flag: the wedge detector
  // below keys on it so a bogus-credentials LOGIN keeps its pill while
  // a lost-after-established registration triggers a rebuild.
  let wasRegistered = false;
  // A registration that HAD succeeded is gone (the server rejected the
  // re-REGISTER after a transport reconnect, dropped the contact, or the
  // Registerer terminated). Retrying on the same Registerer can wedge
  // forever, so rebuild the whole agent like the watchdog does; if even
  // the rebuild fails, fall back to the backoff loop. While the
  // transport is DOWN the reconnect backoff owns recovery instead (a
  // rebuild would build into a dead network and reset the backoff
  // rhythm) — the rebuild is for a registration lost while the
  // transport is UP.
  const registrationLost = () => {
    if (state.userAgent && !state.userAgent.isConnected()) {
      log("registration lost during transport outage; backoff owns it");
      return;
    }
    rebuildConnection("registration lost after it was established").catch(
      (err) => {
        log(`rebuild failed: ${err.message}`);
        scheduleReconnect();
      },
    );
  };
  registerer.stateChange.addListener((regState) => {
    log(`registration ${regState}`);
    if (regState === SIP.RegistererState.Registered) {
      wasRegistered = true;
      clearReconnectCycleDeadline();
      const wasReconnecting = reconnectAttempts > 0;
      reconnectAttempts = 0;
      setRegStatus("status-registered", t("registered"));
      // Reconnect polish: after a transport recovery, say what the user
      // still has instead of silently resuming.
      if (wasReconnecting && sessions.size > 0) {
        log(t("reconnectPreserved")(sessions.size));
      }
      return;
    }
    // Deliberate logout or a rebuild's teardown sets its own pill.
    if (stopping || resetting) return;
    if (regState === SIP.RegistererState.Unregistered) {
      if (wasRegistered) {
        registrationLost();
        return;
      }
      // The REGISTER never succeeded (wrong credentials, account
      // disabled): say so instead of "offline".
      setRegStatus("status-offline", t("regRejected"));
      return;
    }
    if (regState === SIP.RegistererState.Terminated && wasRegistered) {
      registrationLost();
      return;
    }
    setRegStatus("status-offline", regState.toLowerCase());
  });
  await registerer.register();
}

export async function disconnect() {
  stopping = true;
  clearCredentials();
  if (reconnectTimer) clearTimeout(reconnectTimer);
  reconnectTimer = null;
  clearReconnectCycleDeadline();
  ringbackStop();
  // Null the handles like the original single-file app did on logout:
  // a later placeCall must see "not connected", not a stopped agent.
  state.userAgent = null;
  registerer = null;
}
