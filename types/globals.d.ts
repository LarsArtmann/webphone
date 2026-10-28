// The window.PBX_CONFIG wire contract, declared once for the island.
//
// The keys mirror the server's writer EXACTLY (internal/server/configjs.go,
// the anonymous struct behind `window.PBX_CONFIG = …`). A server-side rename
// or an island-side typo therefore becomes a compile error instead of a
// silent `undefined` at runtime — the drift class this gate exists for.
// Everything is optional: a config-less boot degrades to same-host defaults.

/** One ICE server entry, mirroring config.ICEServer's JSON tags. */
interface PBXICEServer {
  urls: string[];
  username?: string;
  credential?: string;
}

/** One shared contact, mirroring domain.SharedContact's JSON tags. */
interface PBXSharedContact {
  name: string;
  number: string;
}

/** Where the ASR provider runs: a loopback/private host or an external one. */
type PBXASRLocality = "" | "local" | "remote";

interface PBXConfig {
  sipDomain?: string;
  websocketPath?: string;
  iceServers?: PBXICEServer[];
  phoneApi?: boolean;
  crm?: boolean;
  asr?: boolean;
  agent?: boolean;
  asrLocality?: PBXASRLocality;
  sipCredential?: boolean;
  contacts?: PBXSharedContact[];
}

interface Window {
  PBX_CONFIG?: PBXConfig;
  /** Older WebKit/Chromium spelling of AudioContext (progressive fallback). */
  webkitAudioContext?: typeof AudioContext;
  /** htmx 2.x, loaded as a classic script before the island modules. */
  htmx?: {
    trigger(target: Element | string, eventName: string): void;
    ajax(
      verb: string,
      path: string,
      context?: {
        target?: Element | string;
        swap?: string;
        select?: string;
        values?: Record<string, string>;
        headers?: Record<string, string>;
      },
    ): Promise<unknown>;
    process(element: Element): void;
  };
  /** The E2E diagnostics handle: call id → the live peer connection. */
  __pcs?: Map<string, RTCPeerConnection>;
  /**
   * Injectable clock for the shell's node:test specs (shell.js reads it at
   * call time, so a test can install a fake after the script loaded). `now`
   * is REQUIRED: an object without it is not a clock, and declaring it
   * optional forced a duck check that could not narrow the call.
   */
  __wpClock?: { now: () => number };
}

/**
 * The CommonJS seam the shell's node:test specs use to reach helpers that are
 * otherwise closure-private (`typeof module !== "undefined"` guards it, so the
 * browser never sees it). Declared separately from Window because this is a
 * module-scope binding, not a page global. Before this declaration the seam
 * compiled only by accident — the `typeof` guard tolerates an unknown name —
 * so a typo inside `module.exports` would have been invisible to the gate.
 */
declare var module: { exports: Record<string, unknown> } | undefined;
