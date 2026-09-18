// SIP connection: user agent, registerer, and the reconnect watchdog.
//
// SIP.js 0.21's userAgent.reconnect() can hang forever after a transport
// loss (observed by the browser E2E reconnect drill even with the server
// reachable again). Every attempt is therefore bounded; a hung one gets
// a full teardown-and-rebuild instead of an eternal "try N" pill.

import { setCredentials, clearCredentials, getCredentials } from "./auth.js";
import { iceServers, sipDomain, websocketUrl } from "./config.js";
import { ringbackStop, ringToneStart, ringToneStop } from "./audio.js";
import { recordHistory } from "./panels.js";
import { titleFlashStart, titleFlashStop, notifyIncoming } from "./notify.js";
import { t } from "./i18n.js";
import { sessions, state } from "./state.js";
import { announce, els, log, setRegStatus } from "./ui.js";

let userAgent = null;
let registerer = null;
let reconnectAttempts = 0;
let reconnectTimer = null;
let stopping = false;
// True while a wedged user agent is being torn down and rebuilt;
// suppresses the teardown's own disconnect/unregistered events.
let resetting = false;

const RECONNECT_ATTEMPT_TIMEOUT_MS = 5000;

export function getUserAgent() {
  return userAgent;
}

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

// Tear the wedged agent down and build a fresh one (the page-reload
// recovery path without losing the UI state).
async function rebuildConnection(reason) {
  log(`reconnect watchdog: ${reason} — rebuilding connection`);
  resetting = true;
  const old = userAgent;
  userAgent = null;
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
  try {
    await withTimeout(
      (async () => {
        await userAgent.reconnect();
        await registerer.register();
      })(),
      RECONNECT_ATTEMPT_TIMEOUT_MS,
      "reconnect timed out",
    );
    reconnectAttempts = 0;
    log("transport reconnected; re-registered");
  } catch (err) {
    log(`reconnect failed: ${err.message}`);
    if (err.message.includes("timed out")) {
      // A hung attempt leaves the agent unusable — rebuild it.
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

  userAgent = new SIP.UserAgent({
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

  await userAgent.start();

  registerer = new SIP.Registerer(userAgent);
  registerer.stateChange.addListener((regState) => {
    log(`registration ${regState}`);
    if (regState === SIP.RegistererState.Registered) {
      const wasReconnecting = reconnectAttempts > 0;
      reconnectAttempts = 0;
      setRegStatus("status-registered", t("registered"));
      // Reconnect polish: after a transport recovery, say what the user
      // still has instead of silently resuming.
      if (wasReconnecting && sessions.size > 0) {
        log(t("reconnectPreserved")(sessions.size));
      }
    } else if (regState === SIP.RegistererState.Unregistered) {
      // Deliberate logout or a rebuild's teardown sets its own pill;
      // anything else means the server rejected the REGISTER (wrong
      // credentials after a reconnect, account disabled) — say so
      // instead of "offline".
      if (!stopping && !resetting) {
        setRegStatus("status-offline", t("regRejected"));
      }
    } else {
      setRegStatus("status-offline", regState.toLowerCase());
    }
  });
  await registerer.register();
}

export async function disconnect() {
  stopping = true;
  clearCredentials();
  if (reconnectTimer) clearTimeout(reconnectTimer);
  reconnectTimer = null;
  ringbackStop();
}
