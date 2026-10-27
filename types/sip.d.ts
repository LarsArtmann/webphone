// The vendored sip.js surface the island actually uses.
//
// sip.js 0.21.2 ships no types and the pinned bundle is loaded as a
// global (`vendor/sip.min.js`), so this file is the ONE hand-written
// declaration of the CALLS the island makes. It is deliberately narrow:
// a member the island never touches stays undeclared, which is what
// makes a 0.21 → 0.22 upgrade visible as a compile error instead of a
// runtime `undefined is not a function`.
//
// Scope note: option bags are typed `unknown` on purpose. sip.js's
// option surface is large, versioned with the bundle, and not the
// contract this repo owns; the RETURN shapes and the state machine are.

/** A sip.js URI (SIP.UserAgent.makeURI's result). */
interface SipUri {
  user?: string;
  host?: string;
  toString(): string;
}

/** The literal session-state vocabulary sip.js emits. */
type SipSessionStateName =
  | "Initial"
  | "Establishing"
  | "Established"
  | "Terminating"
  | "Terminated";

/** The literal registerer-state vocabulary sip.js emits. */
type SipRegistererStateName = "Initial" | "Registered" | "Unregistered" | "Terminated";

/** A sip.js state machine subscriber. */
interface SipStateChange<StateName> {
  addListener(listener: (state: StateName) => void): void;
  removeListener(listener: (state: StateName) => void): void;
}

/** The description handler sip.js builds around one peer connection. */
interface SipSessionDescriptionHandler {
  peerConnection: RTCPeerConnection;
  /** 0.21 media-hold helpers; optional because a custom handler may omit them. */
  enableReceiverTracks?(enabled: boolean): void;
  enableSenderTracks?(enabled: boolean): void;
}

/** A live SIP dialog (inviter or invitation). */
interface SipSession {
  readonly id: string;
  readonly state: SipSessionStateName;
  readonly stateChange: SipStateChange<SipSessionStateName>;
  readonly sessionDescriptionHandler: SipSessionDescriptionHandler | null;
  readonly remoteIdentity: { uri?: { user?: string } } | undefined;
  invite(options?: unknown): Promise<void>;
  accept(options?: unknown): Promise<void>;
  reject(options?: unknown): Promise<void>;
  bye(): Promise<void>;
  hold(options?: unknown): Promise<void>;
  unhold(options?: unknown): Promise<void>;
  mute(options?: unknown): void;
  unmute(options?: unknown): void;
}

/** An outbound dialog. */
interface SipInviter extends SipSession {}

/** An inbound dialog. */
interface SipInvitation extends SipSession {}

/** One registration transport. */
interface SipRegisterer {
  readonly state: SipRegistererStateName;
  readonly stateChange: SipStateChange<SipRegistererStateName>;
  register(options?: unknown): Promise<void>;
  unregister(options?: unknown): Promise<void>;
  dispose(): Promise<void>;
}

/** The user agent (one per connection). */
interface SipUserAgent {
  start(): Promise<void>;
  stop(): Promise<void>;
  reconnect(): Promise<void>;
  isConnected(): boolean;
}

/** A description-handler factory as sip.js invokes it. */
type SipDescriptionHandlerFactory = (
  session: SipSession,
  options: unknown,
) => SipSessionDescriptionHandler;

declare const SIP: {
  UserAgent: {
    new (options?: unknown): SipUserAgent;
    makeURI(uri: string): SipUri | undefined;
  };
  Inviter: new (userAgent: SipUserAgent, target: string, options?: unknown) => SipInviter;
  Registerer: new (userAgent: SipUserAgent, options?: unknown) => SipRegisterer;
  SessionState: {
    Initial: "Initial";
    Establishing: "Establishing";
    Established: "Established";
    Terminating: "Terminating";
    Terminated: "Terminated";
  };
  RegistererState: {
    Initial: "Initial";
    Registered: "Registered";
    Unregistered: "Unregistered";
    Terminated: "Terminated";
  };
  Web: {
    defaultSessionDescriptionHandlerFactory(
      mediaStreamFactory?: unknown,
    ): SipDescriptionHandlerFactory;
  };
};
