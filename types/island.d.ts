// Island-wide type declarations: the DOM handle contract and the wire
// shapes the island reads.
//
// This is a SCRIPT-style .d.ts (no imports/exports) on purpose: every
// declaration here is global, so a module never imports a type — it just
// uses it, exactly like the JSDoc annotations that reference it.
//
// Every id in IslandDomIds is served markup pinned by
// TestServedPageHoldsTheDomContract (docs/dom-contract.md is the list's
// one home). Declaring them means a typo'd id or a renamed element is a
// compile error instead of a silent `null` at runtime.

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
  rejectCause?: string;
  endedReasonShown?: boolean;
  announcedState?: IslandCallViewState;
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
