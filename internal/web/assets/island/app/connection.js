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
import { micMediaStreamFactory, releaseWarmMic, warmMic } from "./mic.js";
import { instrumentSessionDescriptionHandler } from "./pcsetup.js";
import { recordHistory } from "./panels.js";
import { titleFlashStart, titleFlashStop, notifyIncoming } from "./notify.js";
import { t } from "./i18n.js";
import { sessions, state } from "./state.js";
import {
  announce,
  els,
  errorMessage,
  hideIncomingBanner,
  log,
  renderIncomingSlots,
  setOfflineBanner,
  setRegStatus,
} from "./ui.js";

/** @type {SipRegisterer | null} */
let registerer = null;
let reconnectAttempts = 0;
/** @type {ReturnType<typeof setTimeout> | null} */
let reconnectTimer = null;
/** @type {ReturnType<typeof setTimeout> | null} */
let reconnectCycleTimer = null;
let stopping = false;
// True while a wedged user agent is being torn down and rebuilt;
// suppresses the teardown's own disconnect/unregistered events.
let resetting = false;
// The server rejected the REGISTER itself (bad credentials, account
// disabled) — as opposed to a transport that never carried it. The
// resume path keys on this: only a genuine rejection means the stored
// session's credentials are stale; a dead network must NOT log the
// user out of the whole app.
let registerRejected = false;

// Whether the last connect() ended in a REGISTER the server refused.
export function registerWasRejected() {
  return registerRejected;
}

const RECONNECT_ATTEMPT_TIMEOUT_MS = 5000;
// Hard ceiling for one whole reconnect cycle, watched by an
// INDEPENDENT timer: the 2026-09-22 E2E run observed the attempt's
// own withTimeout chain going silent while the page stayed alive
// (pill frozen mid-cycle, sofia 200-OK'd the re-REGISTER, no further
// SIP or pill updates) — the cycle deadline force-rebuilds regardless
// of the wedged chain.
const RECONNECT_CYCLE_DEADLINE_MS = 15000;

// Every offline-truth pill flips the banner on: the pill carries the
// precise reason, the banner is the loud "you cannot call right now".
/** @param {string} text */
function setOfflinePill(text) {
  setRegStatus("status-offline", text);
  setOfflineBanner(true);
}

// The browser's network came back: recover ONLY if the transport is
// actually down and no retry is pending. It never claims registered —
// that verdict stays with the REGISTER (the pill flips there).
export function networkOnline() {
  if (stopping || resetting) return;
  if (!state.userAgent || state.userAgent.isConnected()) return;
  if (!reconnectTimer) scheduleReconnect();
}

/** @param {string} extension @param {string} password @returns {Promise<void>} */
export async function connect(extension, password) {
  stopping = false;
  reconnectAttempts = 0;
  registerRejected = false;
  setCredentials(extension, password);
  // Honest until proven otherwise: between login and the REGISTER's
  // 200-OK the phone cannot call — the banner says so.
  setOfflineBanner(true);
  await buildConnection();
}

function scheduleReconnect() {
  if (stopping || reconnectTimer) return;
  reconnectAttempts += 1;
  const delay = Math.min(30, 2 ** reconnectAttempts);
  setOfflinePill(t("reconnecting")(delay, reconnectAttempts));
  log(`transport lost; reconnect try ${reconnectAttempts} in ${delay}s`);
  reconnectTimer = setTimeout(() => {
    reconnectTimer = null;
    attemptReconnect();
  }, delay * 1000);
}

/**
 * withTimeout bounds one promise. The loser of the race is cleared, so a
 * settled attempt never leaves a timer armed (a leaked one would reject an
 * already-resolved promise — harmless, but it keeps the event loop awake).
 * @template T
 * @param {Promise<T>} promise
 * @param {number} ms
 * @param {string} label
 * @returns {Promise<T>}
 */
function withTimeout(promise, ms, label) {
  /** @type {ReturnType<typeof setTimeout>} */
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
      log(`rebuild failed: ${errorMessage(err)}`);
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
/** @param {string} reason @returns {Promise<void>} */
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
  setOfflinePill(t("regRebuilding"));
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
    log(`old agent stop: ${errorMessage(err)}`);
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
    // Bound once: `state.userAgent` is nullable and the await boundary
    // invalidates the narrowing of a live module field.
    const agent = state.userAgent;
    const registered = registerer;
    if (!agent || !registered) throw new Error("not connected");
    await withTimeout(
      (async () => {
        await agent.reconnect();
        await registered.register();
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
    setOfflineBanner(false);
    if (sessions.size > 0) {
      log(t("reconnectPreserved")(sessions.size));
    }
    log("transport reconnected; re-registered");
  } catch (err) {
    log(`reconnect failed: ${errorMessage(err)}`);
    // A hung attempt or a Terminated registerer leaves the agent
    // unusable (retrying register() on a Terminated registerer throws
    // forever); rebuild it.
    const message = errorMessage(err);
    const unusable =
      message.includes("timed out") ||
      registerer?.state === SIP.RegistererState.Terminated;
    if (unusable) {
      try {
        await rebuildConnection(message);
        reconnectAttempts = 0;
        return;
      } catch (resetErr) {
        log(`rebuild failed: ${errorMessage(resetErr)}`);
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

  // The base factory hands the warm mic to the session; the thin wrap
  // additionally instruments each fresh handler's peer connection so
  // ICE setup timings are captured from construction (pcsetup.js) —
  // sip.js invokes the factory as a plain function, so an arrow wrapper
  // is safe, and the base call stays eager (one factory, many sessions).
  const baseDescriptionHandlerFactory =
    SIP.Web.defaultSessionDescriptionHandlerFactory(micMediaStreamFactory);

  state.userAgent = new SIP.UserAgent({
    uri,
    authorizationUsername: extension,
    authorizationPassword: password,
    transportOptions: { server: websocketUrl },
    sessionDescriptionHandlerFactory: (
      /** @type {SipSession} */ session,
      /** @type {unknown} */ options,
    ) => {
      const handler = baseDescriptionHandlerFactory(session, options);
      instrumentSessionDescriptionHandler(handler);
      return handler;
    },
    sessionDescriptionHandlerFactoryOptions: {
      iceGatheringTimeout: 1000,
      peerConnectionConfiguration: { iceServers },
    },
    // Built-in logger sends SIP-stack warnings (transport failures,
    // malformed responses) to the browser console; without it the
    // console is silent exactly when the connection misbehaves.
    logBuiltinEnabled: true,
    logLevel: "warn",
    delegate: {
      onDisconnect: (/** @type {Error | null} */ error) => {
        if (stopping || resetting) return;
        // Surface WHY the transport died — the pill is the first place
        // a user looks when audio goes quiet; "offline" alone hides
        // certificate/TLS vs network failures.
        const reason = error
          ? `offline: ${error.message || error}`
          : t("offline");
        setOfflinePill(reason);
        if (error) scheduleReconnect();
      },
      onInvite: (/** @type {SipInvitation} */ invitation) => {
        warmMic();
        const from = (invitation.remoteIdentity &&
          invitation.remoteIdentity.uri) || {
          user: "unknown",
        };
        const caller = from.user || "unknown";
        // The ringing queue (N-deep): every invite takes a slot answered
        // oldest-first. A second ring never rejects itself.
        state.incomingSessions.push({ invitation, from: caller });
        renderIncomingSlots();
        // A6 incoming focus: a ringing phone moves focus to Accept —
        // unless the user is mid-type somewhere (the assertive
        // announcement already reached them; stealing keystrokes would
        // be hostile).
        const active = document.activeElement;
        const focused = /** @type {HTMLElement | null} */ (active);
        const midType =
          focused &&
          (focused.tagName === "INPUT" ||
            focused.tagName === "TEXTAREA" ||
            focused.isContentEditable);
        if (!midType) els.accept.focus();
        // Assertive announcement (M17 J9): a hidden→visible swap of the
        // banner is invisible to screen readers — the ringing call is
        // the one message allowed to interrupt (role=alert). Per invite:
        // N simultaneous callers are N announcements.
        announce(t("incomingCall")(caller), "info", {
          assertive: true,
        });
        // Incoming-call UX: system notification (per caller, so N callers
        // are N notifications), audible ring, tab flash.
        notifyIncoming(caller);
        ringToneStart();
        titleFlashStart();
        invitation.stateChange.addListener(
          (/** @type {SipSessionStateName} */ state2) => {
            if (state2 !== SIP.SessionState.Terminated) return;
            // An answered (or deliberately rejected) invite left the queue
            // already; a Terminated arrival only means "the caller gave
            // up" while the invite still held its slot — remove exactly
            // that slot, keep the remaining rings ringing.
            const slot = state.incomingSessions.findIndex(
              (entry) => entry.invitation === invitation,
            );
            if (slot === -1) return;
            state.incomingSessions.splice(slot, 1);
            if (state.incomingSessions.length > 0) {
              renderIncomingSlots();
            } else {
              hideIncomingBanner();
              ringToneStop();
              titleFlashStop();
              releaseWarmMic();
            }
            // The far end gave up before the user answered: say so and
            // keep the attempt in Recent calls (dir in, no duration).
            log(`missed call from ${caller}`, "warn");
            announce(t("missedCall")(caller), "warn");
            // shell.js mirrors this into the missed-call header badge; a
            // deliberate REJECT empties the queue slot first and never
            // lands here (the queue-membership check above).
            document.dispatchEvent(
              new CustomEvent("wp:call-missed", { detail: { target: caller } }),
            );
            recordHistory({
              dir: "in",
              target: caller,
              at: Date.now(),
              dur: 0,
            });
          },
        );
        log(`incoming call from ${caller}`);
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
        log(`rebuild failed: ${errorMessage(err)}`);
        scheduleReconnect();
      },
    );
  };
  registerer.stateChange.addListener((/** @type {string} */ regState) => {
    log(`registration ${regState}`);
    if (regState === SIP.RegistererState.Registered) {
      wasRegistered = true;
      registerRejected = false;
      clearReconnectCycleDeadline();
      const wasReconnecting = reconnectAttempts > 0;
      reconnectAttempts = 0;
      setRegStatus("status-registered", t("registered"));
      setOfflineBanner(false);
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
      // disabled): say so instead of "offline", and mark it so the
      // resume path knows the session's credentials are the problem —
      // not the network.
      registerRejected = true;
      setOfflinePill(t("regRejected"));
      return;
    }
    if (regState === SIP.RegistererState.Terminated && wasRegistered) {
      registrationLost();
      return;
    }
    setOfflinePill(regState.toLowerCase());
  });
  await registerer.register();
}

export async function disconnect() {
  stopping = true;
  clearCredentials();
  if (reconnectTimer) clearTimeout(reconnectTimer);
  reconnectTimer = null;
  clearReconnectCycleDeadline();
  setOfflineBanner(false);
  ringbackStop();
  releaseWarmMic();
  // Null the handles like the original single-file app did on logout:
  // a later placeCall must see "not connected", not a stopped agent.
  state.userAgent = null;
  registerer = null;
}
