// Island-wide type declarations: the DOM handle contract and the wire
// shapes the island reads.
//
// This is a SCRIPT-style .d.ts (no imports/exports) on purpose: every
// declaration here is global, so a module never imports a type — it just
// uses it, exactly like the JSDoc annotations that reference it.
//
// Every id in IslandDomIds is served markup: docs/dom-contract.md is the
// list's one home (pinned there by TestServedPageHoldsTheDomContract), and
// TestIslandDomIdsMirrorTheDomContract refuses an id declared here that the
// document does not carry — so a rename cannot leave one of the three
// mirrors behind. The relation is one-directional: the document also lists
// shell-owned ids (wp-live, wp-tab-skeleton) and the markup-only call-btn,
// which the island never looks up. Declaring an id means a typo'd lookup is
// a compile error instead of a silent `null` at runtime.

interface IslandDomIds {
  "reg-status": HTMLElement;
  lang: HTMLSelectElement;
  "login-view": HTMLElement;
  "login-form": HTMLFormElement;
  "login-error": HTMLElement;
  ext: HTMLInputElement;
  pass: HTMLInputElement;
  remember: HTMLInputElement;
  "phone-view": HTMLElement;
  "whoami-ext": HTMLElement;
  logout: HTMLButtonElement;
  "dial-form": HTMLFormElement;
  dest: HTMLInputElement;
  calls: HTMLElement;
  keypad: HTMLElement;
  "history-list": HTMLElement;
  "history-wrap": HTMLDetailsElement;
  "incoming-call": HTMLElement;
  "incoming-from": HTMLElement;
  "incoming-queue": HTMLElement;
  "accept-btn": HTMLButtonElement;
  "reject-btn": HTMLButtonElement;
  "contacts-wrap": HTMLDetailsElement;
  "contacts-list": HTMLElement;
  "vm-wrap": HTMLDetailsElement;
  "vm-badge": HTMLElement;
  "vm-list": HTMLElement;
  "vm-refresh": HTMLButtonElement;
  "vm-status": HTMLElement;
  "ice-wrap": HTMLDetailsElement;
  "ice-panel": HTMLElement;
  log: HTMLElement;
  toasts: HTMLElement;
  "dial-error": HTMLElement;
  "wp-dial-hint": HTMLElement;
  "offline-banner": HTMLElement;
  "remote-audio": HTMLAudioElement;
  "wp-adv-toggle": HTMLButtonElement;
  "wp-advanced": HTMLElement;
  "wp-devtest-btn": HTMLButtonElement;
  "sso-error": HTMLElement;
  "audio-out-wrap": HTMLElement;
  "audio-output": HTMLSelectElement;
}

/** The island's four feedback kinds (ui.js owns the vocabulary). */
type IslandToastKind = "ok" | "error" | "warn" | "info";

/** The view states CALL_TRANSITIONS declares, in one place. */
type IslandCallViewState =
  | "ringing"
  | "active"
  | "holding"
  | "resuming"
  | "held"
  | "ending"
  | "ended";

/** One live call in the island's session table (state.js owns the map). */
interface IslandCallEntry {
  session: SipSession;
  target: string;
  held: boolean;
  /** The in-flight toggle (the honesty contract's pending state). */
  holdPending: "holding" | "resuming" | null;
  /** A toggle requested mid-flight; the settled state decides whether it runs. */
  holdQueued?: boolean;
  muted: boolean;
  established: boolean;
  startedAt: number;
  timer: ReturnType<typeof setInterval> | null;
  dom: HTMLElement;
  volume?: number;
  transferring?: boolean;
  /** The SIP status code that ended a failed outbound INVITE (0 = none). */
  rejectCause?: number;
  endedReasonShown?: boolean;
  announcedState?: IslandCallDockState;
  captionUnsub?: () => void;
  transcriptSaveQueue?: IslandTranscriptSegment[];
  transcriptSaveFlushing?: boolean;
  transcriptSaved?: number;
  transcriptSaveWarned?: boolean;
  transcriptSavedAnnounced?: boolean;
}

/** One persisted live-call transcript segment (POST /api/transcripts). */
interface IslandTranscriptSegment {
  callId: string;
  direction: "in" | "out";
  remote: string;
  startedAt: number;
  text: string;
}

/** One unanswered inbound INVITE in the ringing queue. */
interface IslandIncomingSession {
  invitation: SipInvitation;
  from: string;
}

/** The island's shared mutable state (state.js). */
interface IslandState {
  focusedId: string | null;
  incomingSessions: IslandIncomingSession[];
  userAgent: SipUserAgent | null;
}

/** The island's local call log entry (panels.js renders it). */
interface IslandHistoryEntry {
  dir: "in" | "out";
  target: string;
  at: number;
  dur: number;
}

/** The /api/session answer (server.sessionIdentity) as the island reads it. */
interface IslandSessionIdentity {
  extension: string;
  did?: string;
  display_name?: string;
  numbers?: string[];
  /** Every extension mapped to the signed-in user (mapped logins only);
   * more than one renders the line switcher. The switch endpoint
   * re-validates membership server-side — this list is a suggestion. */
  extensions?: string[];
  password?: string;
  sip_credential?: string;
  agent_mode?: string;
  agent_mode_source?: string;
  agent_default_mode?: string;
  agent_modes?: string[];
}

/** The ringing-stack button handlers ui.js renders (calls.js supplies them). */
interface IslandIncomingSlotActions {
  /** answerIncoming also takes an exact SipInvitation and returns whether it
   * answered, so the stricter signal types are what the registry accepts. */
  answer(target?: SipInvitation | number): void;
  reject(target?: SipInvitation | number): void;
}

/** The pcsetup.js setup-timing marks (one per peer connection). */
interface IslandSetupMarks {
  gatheringStartedAt: number | null;
  gatheringEndedAt: number | null;
  iceConnectedAt: number | null;
}

/** What setupSummary reports to ice.js's rendering. */
interface IslandSetupSummary {
  gatherMs: number | null;
  gatherCapped: boolean;
  gatheringStartedAt: number | null;
  iceConnectedAt: number | null;
}

/**
 * The subset of an RTCStats dictionary the ICE panel reads. The DOM lib
 * hands `RTCStatsReport.get()` back as `any`; declaring the shape keeps
 * a renamed/removed stat field a compile error instead of "?" on the
 * panel (RTCStats dictionaries are a moving target across browsers).
 */
interface IslandRtcStat {
  id?: string;
  type?: string;
  kind?: string;
  state?: string;
  selected?: boolean;
  nominated?: boolean;
  selectedCandidatePairId?: string;
  localCandidateId?: string;
  remoteCandidateId?: string;
  candidateType?: string;
  currentRoundTripTime?: number;
  packetsLost?: number;
  jitter?: number;
  bytesReceived?: number;
  mimeType?: string;
}

/** The ICE panel's one-tick aggregate. */
interface IslandIceSummary {
  iceState: RTCIceConnectionState;
  selectedPair: string | null;
  selectedPairData: IslandRtcStat | null;
  localType: string;
  remoteType: string;
  rtt: number | null;
  packetsLost: number;
  jitter: number | null;
  codec: string;
  bytesReceived: number;
}

/** The WebAuthn ceremony JSON the enroll endpoints exchange (enroll/webauthn.js). */
interface IslandRegistrationOptionsWire {
  challenge: string;
  user?: { id: string; name?: string; displayName?: string };
  excludeCredentials?: Array<{
    id: string;
    type?: string;
    transports?: string[];
  }>;
}

/** The byte-bearing fields of a PublicKeyCredential response. */
interface IslandCredentialResponseFields {
  clientDataJSON?: ArrayBuffer;
  authenticatorData?: ArrayBuffer;
  signature?: ArrayBuffer;
  attestationObject?: ArrayBuffer;
  userHandle?: ArrayBuffer | null;
  transports?: string[];
}

/** One serialized credential as the enroll finish endpoints parse it. */
interface IslandSerializedCredential {
  id: string;
  rawId: string;
  type: string;
  response: Record<string, string | string[]>;
}

// The standalone passkey-enrollment page (served at /assets/enroll/, no
// island runtime). Ids come from views/enroll.templ; the page is
// CONDITIONAL markup (only rendered when auth.passkey.* is enabled), so
// nothing outside its own module may assume these elements exist.
interface IslandEnrollDomIds {
  "enroll-form": HTMLFormElement;
  "enroll-token": HTMLInputElement;
  "enroll-credential-name": HTMLInputElement;
  "enroll-status": HTMLElement;
  "enroll-error": HTMLElement;
}

/** The verify endpoint's answer: the address the token belongs to. */
interface IslandEnrollVerifyBody {
  email: string;
  user_id: string;
}

/** The begin endpoint's answer: the ceremony options plus the session key
 * the finish call must echo back. */
interface IslandEnrollBeginBody {
  options: IslandRegistrationOptionsWire;
  session_key: string;
}

/** One postJSON verdict — never throws; status 0 is the network sentinel. */
interface IslandPostJSONAnswer {
  ok: boolean;
  status: number;
  body: any;
}

/** The live-capture mix: one AudioContext carrying both call legs, whose
 * destination node the silence probe taps. */
interface IslandCapture {
  stream: MediaStream;
  ctx: AudioContext;
  dest: MediaStreamAudioDestinationNode;
}

/** One running live-capture controller per call id. */
interface IslandCaptureController {
  recorder: MediaRecorder;
  ctx: AudioContext;
  stream: MediaStream;
}

/** One CDR row as /phone-api/history answers it (FreeSWITCH field names —
 * the proxy forwards them verbatim, so the island reads them verbatim). */
interface IslandServerHistoryRow {
  context?: string;
  caller_id_number?: string;
  destination_number?: string;
  start?: string;
  billsec?: number;
}

/** One personal contact as the server store renders it (Contacts tab and
 * /api/contacts agree). The legacy localStorage rows are the same shape
 * minus `id`. */
interface IslandContactRow {
  id?: string;
  name: string;
  number: string;
}

/** One voicemail message as /phone-api/voicemail/…/messages answers it. */
interface IslandVoicemailMessage {
  uuid: string;
  read: boolean;
  cid_number?: string;
  cid_name?: string;
  seconds: number;
  /** Epoch SECONDS (the phone API's unit — not milliseconds). */
  created: number;
  audio_url: string;
}

/** The voicemail summary counters (/…/summary). */
interface IslandVoicemailSummary {
  new: number;
}

/** Every island message key. Derived from t()'s own parameter type, so
 * i18n.js's `en` table stays the ONE home for the key list — this alias is
 * only what other modules can name. */
type IslandMessageKey = Parameters<typeof import("../internal/web/assets/island/app/i18n.js").t>[0];

/** The dock chip / card `data-state` vocabulary. It is the CALL_TRANSITIONS
 * vocabulary minus the settled "ended" (an ended call has no live chip). */
type IslandCallDockState = "ringing" | "established" | "holding" | "held" | "resuming" | "ending";
