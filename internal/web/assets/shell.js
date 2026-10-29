// Shell glue: small behaviors the server-rendered tabs need.
// CSP-strict: a plain external script, no inline handlers.

(function () {
  "use strict";

  // The element an event came from. DOM events type their target as a bare
  // EventTarget (and a document-dispatched event has no element at all), so
  // every delegated handler here would otherwise cast at each call site —
  // and a text-node or document target would throw on .closest() instead of
  // falling through the "not my element" test. An event with no element
  // answers the document element, which matches none of this file's
  // selectors.
  /**
   * The element an event came from, as an HTMLElement: every event wired in
   * this file is dispatched from (or bubbles through) HTML markup — the
   * inline SVG icons live inside HTML buttons, so a path target still
   * resolves through closest().
   * @param {Event} event
   * @returns {HTMLElement}
   */
  function eventElement(event) {
    var target = /** @type {HTMLElement | null} */ (event.target);
    // Duck check rather than `instanceof Element`: the node test stubs are
    // plain objects (and some are forms that answer matches() but not
    // closest()), so instanceof would throw there and take every delegated
    // handler down with it.
    if (
      target &&
      (hasMethod(target, "closest") || hasMethod(target, "matches"))
    ) {
      return target;
    }
    return document.documentElement;
  }

  /**
   * The custom-event detail htmx and the island attach. Deliberately `any`:
   * each listener below narrows the fields it reads, and htmx's own detail
   * shapes are untyped, so a single honest hole beats twenty casts.
   * @param {Event} event
   * @returns {any}
   */
  function eventDetail(event) {
    return /** @type {CustomEvent} */ (event).detail;
  }

  /**
   * closest() with the HTMLElement result the shell's handlers read.
   * Element.closest returns the bare Element base, which carries neither
   * dataset nor value nor disabled — and every selector this file searches
   * matches a real HTML element (form/button/a/input/textarea/li).
   * @param {Element} el
   * @param {string} selector
   * @returns {HTMLElement | null}
   */
  function closestHTMLElement(el, selector) {
    return /** @type {HTMLElement | null} */ (el.closest(selector));
  }

  /**
   * One of the voicemail panel's media elements by id, typed as the audio
   * element it is (the panel renders <audio> under these ids).
   * @param {string} id
   * @returns {HTMLAudioElement | null}
   */
  function audioElement(id) {
    return /** @type {HTMLAudioElement | null} */ (document.getElementById(id));
  }

  /**
   * One of the voicemail panel's wave canvases by id. Twin of audioElement:
   * the <canvas> is server-rendered under that id, so the lookup is the only
   * untyped step.
   * @param {string} id
   * @returns {HTMLCanvasElement | null}
   */
  function canvasElement(id) {
    return /** @type {HTMLCanvasElement | null} */ (
      document.getElementById(id)
    );
  }

  /**
   * The scrubber canvas a pointer/key event landed on (the panel marks it
   * with .wp-vm-wave), typed as the canvas it is.
   * @param {Element} el
   * @param {string} selector
   * @returns {HTMLCanvasElement | null}
   */
  function canvasIn(el, selector) {
    return /** @type {HTMLCanvasElement | null} */ (
      closestHTMLElement(el, selector)
    );
  }

  /**
   * The form an element belongs to, typed. Every delegated handler here acts
   * on markup templ renders as a real <form>.
   * @param {Element} el
   * @returns {HTMLFormElement | null}
   */
  function closestForm(el) {
    return /** @type {HTMLFormElement | null} */ (
      closestHTMLElement(el, "form")
    );
  }

  /**
   * The textarea a selector resolves to inside a scope (the composer body,
   * a call's note box). Templ renders each of them as <textarea>.
   * @param {ParentNode} scope
   * @param {string} selector
   * @returns {HTMLTextAreaElement | null}
   */
  function textareaIn(scope, selector) {
    return /** @type {HTMLTextAreaElement | null} */ (
      scope.querySelector(selector)
    );
  }

  // Inline SVG icon markup (the JS-created twins of the server's
  // templ-components icons): plain strings assigned via innerHTML —
  // sized by the shared .wp-icon CSS class, tinted by currentColor.
  // Platform emoji rendered differently per OS and read as decoration.
  /**
   * The agent-mode slice the island forwards on wp:session-opened (the
   * session identity's agent fields). Shell-local on purpose: only this
   * file's pill renders it, so types/island.d.ts does not own it.
   * @typedef {{mode: string, source: string, defaultMode: string, modes: string[]}} AgentModeState
   */

  /**
   * Whether an element-like value actually answers a DOM method. A real
   * HTMLElement always does, so `if (el.closest)` reads as dead code to the
   * checker — but the node:test stubs are plain objects (some answer
   * matches() and not closest()), so these guards are load-bearing at
   * runtime. One documented home beats four casts.
   * @param {unknown} el
   * @param {string} method
   * @returns {boolean}
   */
  function hasMethod(el, method) {
    return (
      Boolean(el) &&
      typeof (/** @type {Record<string, unknown>} */ (el)[method]) ===
        "function"
    );
  }

  /**
   * The human message of a thrown value. Mirrors the island's errorMessage
   * (ui.js) so a toast quotes the same text the event log does; a rejection
   * can be an Error, a string, a plain object, or null, and every one of
   * those must still render something readable.
   * @param {unknown} err
   * @returns {string}
   */
  function errorMessage(err) {
    if (err instanceof Error) return err.message;
    if (typeof err === "string") return err;
    return String(err);
  }

  /**
   * @param {string} path
   * @returns {string}
   */
  var svgIconHtml = function (path) {
    return (
      '<svg class="wp-icon" viewBox="0 0 24 24" fill="currentColor" aria-hidden="true">' +
      '<path d="' +
      path +
      '"></path></svg>'
    );
  };
  var WP_PLAY_SVG = svgIconHtml("M8 5.14v13.72L19 12 8 5.14z");
  var WP_PAUSE_SVG = svgIconHtml("M7 5h4v14H7zM13 5h4v14h-4z");
  var WP_ATTACH_SVG = svgIconHtml(
    "M16.5 6v11.5a4 4 0 0 1-8 0V5a2.5 2.5 0 0 1 5 0v10.5a1 1 0 0 1-2 0V6H10v9.5a2.5 2.5 0 0 0 5 0V5a4 4 0 0 0-8 0v12.5a5.5 5.5 0 0 0 11 0V6h-1.5z",
  );
  // Load-error boundary (plan T19): if ANY shell wiring throws at load,
  // leave a breadcrumb in the island's event log (it may still exist
  // even when the island itself failed) and rethrow to the console —
  // never a silent half-wired shell.
  try {
    // 1. Nav active state across HTMX partial swaps: the server marks the
    //    active link on full renders; after a partial swap only the clicked
    //    link knows. htmx:afterRequest fires on the link that issued it.
    document.addEventListener("htmx:afterRequest", function (event) {
      var element = eventElement(event);
      if (!element || !element.hasAttribute("data-tab")) return;
      var nav = closestHTMLElement(element, ".wp-nav");
      if (!nav) return;
      nav.querySelectorAll(".wp-nav-link").forEach(function (link) {
        link.classList.toggle("wp-active", link === element);
        // aria-current mirrors wp-active with the same definite values
        // the server renders ("page" / "false") — screen readers keep
        // the active-tab announcement in step with the visual state.
        link.setAttribute("aria-current", link === element ? "page" : "false");
      });
      // 1a. Remember the last-active tab (E3): a plain "/" reload can
      //     restore it. Deep links never touch this path — only real
      //     tab navigations do.
      var tab = element.getAttribute("data-tab");
      if (tab) {
        try {
          localStorage.setItem("wp-last-tab", tab);
        } catch {
          // Private-mode storage: the tab memory is best-effort.
        }
      }
    });

    // 1b. Restore the remembered tab on a plain "/" load (E3): only
    //     when the shell renders real tab content — never over a deep
    //     link, the sign-in hint, or a stored value that is not one of
    //     the rendered tabs (garbage must not 404 the boot). The
    //     address bar follows the content via replaceState.
    // Named + called immediately so the node:test specs can drive the
    // same code path directly (shell.js is CommonJS under node — a
    // query-string re-import returns the cached module, never a fresh
    // evaluation).
    var restoreLastTab = function () {
      var path = window.location && window.location.pathname;
      if (path !== "/") return;
      var stored;
      try {
        stored = localStorage.getItem("wp-last-tab");
      } catch {
        return;
      }
      if (!stored || stored === "messages") return;
      var link = document.querySelector(
        '#wp-nav .wp-nav-link[data-tab="' + stored + '"]',
      );
      var active = document.querySelector("#wp-nav .wp-nav-link.wp-active");
      if (!link || (active && active === link)) return;
      var content = document.getElementById("tab-content");
      if (!content || !window.htmx || content.querySelector(".wp-welcome"))
        return;
      window.htmx.ajax("GET", "/partials/" + stored, {
        target: "#tab-content",
        swap: "innerHTML",
      });
      window.htmx.ajax("GET", "/partials/nav?active=" + stored, {
        target: "#wp-nav",
        swap: "morph:innerHTML",
      });
      if (window.history && window.history.replaceState) {
        window.history.replaceState(null, "", "/" + stored);
      }
    };
    restoreLastTab();

    // 1c. Welcome dismissal (M18): collapsing the sign-in intro is a
    //     per-browser choice. theme-preload.js applied the root class
    //     before first paint at load; here we only own the click —
    //     persist the flag and flip the class so the collapse takes
    //     effect immediately, without a reload.
    document.addEventListener("click", function (event) {
      var button = /** @type {HTMLButtonElement | null} */ (
        closestHTMLElement(eventElement(event), ".wp-welcome-dismiss")
      );
      if (!button) return;
      try {
        localStorage.setItem("wp-welcome-dismissed", "1");
      } catch {
        // Private-mode storage: the dismissal still applies to this page.
      }
      document.documentElement.classList.add("wp-welcome-dismissed");
    });

    // 2. data-dial buttons (contacts, history, voicemail, threads): push
    //    the number into the island's dial form and submit it — the same
    //    gesture as the island's redial. When the island is signed out
    //    (#phone-view hidden) a submit would dead-end inside the hidden
    //    form, so we say so and point at the login field instead. The
    //    toast mirrors the island's announce() markup (shared CSS, no
    //    island import — shell and island never load each other's code).
    /**
     * @param {string} message
     * @param {IslandToastKind} [kind]
     */
    var shellToast = function (message, kind) {
      var host = document.getElementById("toasts");
      if (!host) return;
      // Identical-consecutive dedup (island announce() parity): a repeated
      // identical failure must not stack look-alike toasts. The 8s error
      // throttle above collapses storms; this covers non-throttled paths.
      if (
        host.children.length > 0 &&
        host.children[host.children.length - 1].textContent === message
      ) {
        return;
      }
      var toast = document.createElement("div");
      toast.className = "toast toast-" + kind;
      toast.textContent = message;
      // Keyboard parity with the island's announce(): focusable, dismiss
      // with Enter/Space/Escape. Never steals focus on creation.
      toast.tabIndex = 0;
      var dismiss = function () {
        toast.remove();
      };
      toast.addEventListener("click", dismiss);
      toast.addEventListener("keydown", function (event) {
        if (
          event.key === "Enter" ||
          event.key === " " ||
          event.key === "Escape"
        ) {
          event.preventDefault();
          dismiss();
        }
      });
      host.append(toast);
      while (host.children.length > 4) host.firstChild?.remove();
      setTimeout(dismiss, 6000);
    };
    document.addEventListener("click", function (event) {
      var button = closestHTMLElement(eventElement(event), "[data-dial]");
      if (!button) return;
      var dest = /** @type {HTMLInputElement | null} */ (
        document.getElementById("dest")
      );
      var form = /** @type {HTMLFormElement | null} */ (
        document.getElementById("dial-form")
      );
      var phoneView = document.getElementById("phone-view");
      if (!dest || !form) return;
      if (phoneView && phoneView.hidden) {
        var ext = document.getElementById("ext");
        if (ext) ext.focus();
        shellToast(
          "Phone is signed out — sign in on the phone panel to call.",
          "warn",
        );
        return;
      }
      dest.value = button.getAttribute("data-dial") || "";
      form.requestSubmit();
    });

    // 2b2. data-sms (history/voicemail rows → the Messages composer
    //      prefilled with the number) and data-save-contact (☆ → the
    //      island's contact save via the wp:save-contact event; the
    //      island owns /api/contacts writes). The prefill rides a
    //      one-shot afterSwap: the composer only exists once the tab
    //      partial has swapped in.
    document.addEventListener("click", function (event) {
      var sms = closestHTMLElement(eventElement(event), "[data-sms]");
      if (sms) {
        var number = sms.getAttribute("data-sms");
        var nav = document.querySelector("[data-tab='messages']");
        if (nav && hasMethod(nav, "click")) {
          /** @type {HTMLElement} */ (nav).click();
        }
        var prefill = function () {
          document.removeEventListener("htmx:afterSwap", prefill);
          var to = /** @type {HTMLInputElement | null} */ (
            document.querySelector("form.wp-compose-new input[name='to']")
          );
          if (to) {
            to.value = number || "";
            to.focus();
          }
        };
        document.addEventListener("htmx:afterSwap", prefill);
        return;
      }
      var save = closestHTMLElement(eventElement(event), "[data-save-contact]");
      if (save) {
        document.dispatchEvent(
          new CustomEvent("wp:save-contact", {
            detail: {
              number: save.getAttribute("data-save-contact"),
              name: save.getAttribute("data-name") || "",
            },
          }),
        );
      }
    });

    // 2b4. data-delete-transcript (History's transcript section → the
    //      end-user erasure affordance, the GDPR leg next to the
    //      operator's age-sweep): confirm with the SERVER's localized
    //      prompt (data-confirm), DELETE the call's segments
    //      owner-scoped, then drop the row — an emptied section goes
    //      with its last row. A failed delete keeps the row and says
    //      so; erasure must never LOOK done when it is not.
    document.addEventListener("click", function (event) {
      var btn = /** @type {HTMLButtonElement | null} */ (
        closestHTMLElement(eventElement(event), "[data-delete-transcript]")
      );
      if (!btn) return;
      // Bound once: a `var` is not narrowed inside the promise closures.
      var button = btn;
      var callId = btn.getAttribute("data-delete-transcript") || "";
      var confirmText =
        btn.getAttribute("data-confirm") || "Delete this transcript?";
      if (hasMethod(window, "confirm") && !window.confirm(confirmText)) return;
      fetch("/api/transcripts?call=" + encodeURIComponent(callId), {
        method: "DELETE",
        credentials: "same-origin",
        headers: { "X-CSRF-Token": csrfToken() },
      })
        .then(function (res) {
          if (!res.ok) throw new Error("HTTP " + res.status);
          var row = closestHTMLElement(button, ".wp-transcript-call");
          if (row) {
            var section = closestHTMLElement(row, ".wp-transcripts");
            row.remove();
            if (section && !section.querySelector(".wp-transcript-call")) {
              section.remove();
            }
          }
          shellToast("Transcript deleted.", "ok");
        })
        .catch(function () {
          shellToast(
            "Could not delete the transcript — nothing was removed.",
            "warn",
          );
        });
    });

    // 2b5. Copy affordances (M16): [data-copy-transcript] copies its
    //      History row's transcript; [data-copy-target] copies the text
    //      of a referenced node (the chip the transcribe flow plants
    //      next to voicemail/attachment results). One delegated
    //      handler, clipboard API only (secure context — the stack
    //      fronts TLS), honest toasts either way. English (D3).
    document.addEventListener("click", function (event) {
      var btn =
        closestHTMLElement(eventElement(event), "[data-copy-transcript]") ||
        closestHTMLElement(eventElement(event), "[data-copy-target]");
      if (!btn) return;
      var text = "";
      if (btn.hasAttribute("data-copy-transcript")) {
        var row = closestHTMLElement(btn, ".wp-transcript-call");
        var node = row && row.querySelector(".wp-transcript-text");
        text = node ? (node.textContent || "").trim() : "";
      } else {
        var target = document.getElementById(
          btn.getAttribute("data-copy-target") || "",
        );
        text =
          target && !target.hidden ? (target.textContent || "").trim() : "";
      }
      if (!text) {
        shellToast("Nothing to copy yet.", "warn");
        return;
      }
      var clip = navigator.clipboard;
      if (!clip || !clip.writeText) {
        shellToast("Could not copy — select the text manually.", "warn");
        return;
      }
      clip.writeText(text).then(
        function () {
          shellToast("Copied to clipboard.", "ok");
        },
        function () {
          shellToast("Could not copy — select the text manually.", "warn");
        },
      );
    });

    // 2b6. Transcript download (#40): [data-download-transcript] saves
    //      its History row's transcript as one text/plain file named
    //      after the call id. Blob + object URL + one synthetic click —
    //      no server round trip (the rendered row is the truth), works
    //      when island scripts fail. English toasts (D3).
    document.addEventListener("click", function (event) {
      var btn = /** @type {HTMLButtonElement | null} */ (
        closestHTMLElement(eventElement(event), "[data-download-transcript]")
      );
      if (!btn) return;
      var row = closestHTMLElement(btn, ".wp-transcript-call");
      var node = row && row.querySelector(".wp-transcript-text");
      var text = node ? (node.textContent || "").trim() : "";
      if (!text) {
        shellToast("Nothing to download yet.", "warn");
        return;
      }
      var callId = btn.getAttribute("data-download-transcript") || "call";
      var blob = new Blob([text], { type: "text/plain" });
      var url = URL.createObjectURL(blob);
      var link = document.createElement("a");
      link.href = url;
      link.download = "transcript-" + callId + ".txt";
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);
      URL.revokeObjectURL(url);
      shellToast("Transcript downloaded.", "ok");
    });

    // 2b7. Post-call notes (M16): [data-note-save] PUTs the row's
    //      textarea to /api/calls/{id}/note; [data-note-draft] asks the
    //      summarize chain for a draft and inserts its text for a HUMAN
    //      edit (the AI never writes the note itself — the save is
    //      always the owner's keystroke). The server's failure lines
    //      are already human-phrased; the toasts relay them verbatim.
    //      English toasts (D3).
    /**
     * @param {string} callId
     * @returns {HTMLTextAreaElement | null}
     */
    var noteBodyFor = function (callId) {
      return textareaIn(document, '[data-note-body="' + callId + '"]');
    };
    document.addEventListener("click", function (event) {
      var btn = /** @type {HTMLButtonElement | null} */ (
        closestHTMLElement(eventElement(event), "[data-note-save]")
      );
      if (!btn) return;
      // Bound once: the fetch callbacks below run in a later frame, where
      // the checker can no longer narrow the mutable `btn`.
      var saveButton = /** @type {HTMLButtonElement} */ (btn);
      var callId = btn.getAttribute("data-note-save") || "";
      var box = noteBodyFor(callId);
      if (!box) return;
      saveButton.disabled = true;
      fetch("/api/calls/" + encodeURIComponent(callId) + "/note", {
        method: "PUT",
        headers: {
          "Content-Type": "application/json",
          "X-CSRF-Token": csrfToken(),
        },
        credentials: "same-origin",
        body: JSON.stringify({ body: box.value }),
      })
        .then(function (res) {
          if (res.status === 204) {
            shellToast("Note saved.", "ok");
            return;
          }
          return res.text().then(function (detail) {
            shellToast(
              detail || "Could not save the note (HTTP " + res.status + ").",
              "warn",
            );
          });
        })
        .catch(function () {
          shellToast(
            "Could not save the note — check your connection.",
            "error",
          );
        })
        .then(function () {
          saveButton.disabled = false;
        });
    });
    document.addEventListener("click", function (event) {
      var btn = /** @type {HTMLButtonElement | null} */ (
        closestHTMLElement(eventElement(event), "[data-note-draft]")
      );
      if (!btn) return;
      var draftButton = /** @type {HTMLButtonElement} */ (btn);
      var callId = btn.getAttribute("data-note-draft") || "";
      var box = noteBodyFor(callId);
      if (!box) return;
      var note = box;
      draftButton.disabled = true;
      fetch("/api/calls/" + encodeURIComponent(callId) + "/summary-draft", {
        method: "POST",
        headers: { "X-CSRF-Token": csrfToken() },
        credentials: "same-origin",
      })
        .then(function (res) {
          if (res.status === 200) {
            return res.json().then(function (draft) {
              if (draft.duplicate) {
                shellToast("Already drafted — see Messages.", "info");
                return;
              }
              note.value = draft.text || "";
              shellToast("Draft inserted — edit before saving.", "ok");
            });
          }
          return res.text().then(function (detail) {
            shellToast(
              detail || "Could not draft (HTTP " + res.status + ").",
              "warn",
            );
          });
        })
        .catch(function () {
          shellToast(
            "Could not reach the AI agent — check your connection.",
            "error",
          );
        })
        .then(function () {
          draftButton.disabled = false;
        });
    });

    // 2c. Live-call presence: the island dispatches wp:calls-changed after
    //     every call render; the shell mirrors the live call count into
    //     the header so every tab shows the phone is busy. Cards are
    //     removed on teardown (calls.js), so counting them counts live
    //     calls (ringing included). English by the shell.js precedent
    //     (theme toggle) — presence is operator glanceable state.
    /** @type {HTMLElement | null} */
    var callBadge = null;
    var updateCallBadge = function () {
      var calls = document.getElementById("calls");
      var actions = document.querySelector(".wp-header-actions");
      if (!calls || !actions) return;
      var count = calls.querySelectorAll(".call-card").length;
      if (count > 0) {
        if (!callBadge) {
          callBadge = document.createElement("span");
          callBadge.id = "call-badge";
          callBadge.className = "wp-call-badge";
          actions.prepend(callBadge);
        }
        callBadge.textContent = count === 1 ? "on call" : "on call · " + count;
      } else if (callBadge) {
        callBadge.remove();
        callBadge = null;
      }
    };
    document.addEventListener("wp:calls-changed", updateCallBadge);
    updateCallBadge();

    // 2d. Missed-call presence: the island dispatches wp:call-missed when
    //     an inbound call ended without being answered (the caller gave
    //     up, or an accepted call died before media). A deliberate user
    //     REJECT is a seen call and never counts. The badge survives
    //     until the History tab — the surface where missed calls are
    //     reviewed — is opened. English shell copy (D3); client-only
    //     state: the PBX CDR carries no missed flag, and the badge is
    //     a session-scoped glance signal by nature.
    /** @type {HTMLElement | null} */
    var missedBadge = null;
    var missedCount = 0;
    var renderMissedBadge = function () {
      var actions = document.querySelector(".wp-header-actions");
      if (!actions) return;
      if (missedCount > 0) {
        if (!missedBadge) {
          missedBadge = document.createElement("span");
          missedBadge.id = "missed-badge";
          missedBadge.className = "wp-missed-badge";
          actions.prepend(missedBadge);
        }
        missedBadge.textContent = "missed · " + missedCount;
      } else if (missedBadge) {
        missedBadge.remove();
        missedBadge = null;
      }
    };
    document.addEventListener("wp:call-missed", function () {
      missedCount += 1;
      renderMissedBadge();
    });
    document.addEventListener("click", function (event) {
      var link = closestHTMLElement(
        eventElement(event),
        "[data-tab='history']",
      );
      if (!link) return;
      missedCount = 0;
      renderMissedBadge();
    });

    // 2d-2. Agent-mode pill (M13): the textagent's draft autonomy,
    //     glanceable next to the call badge. The island forwards the
    //     session identity's agent fields on wp:session-opened; no mode
    //     field = no pill (the seam is off, or the boot probe could not
    //     see the agent — hiding beats lying). Click cycles through the
    //     agent's modes; a runtime pin grows a reset chip (the Messages
    //     panel select stays the ONE settings home — the pill is
    //     visibility plus one gesture, never a second config surface).
    //     English shell copy (D3); announcements ride the polite live
    //     region.
    /** @type {HTMLButtonElement | null} */
    var agentPill = null;
    /** @type {HTMLButtonElement | null} */
    var agentReset = null;
    /** @type {AgentModeState | null} */
    var agentModeState = null;
    /** @param {string} text */
    var announceLive = function (text) {
      var live = document.getElementById("wp-live");
      if (!live) return;
      live.textContent = "";
      live.textContent = text;
    };
    var renderAgentPill = function () {
      var actions = document.querySelector(".wp-header-actions");
      if (!actions) return;
      if (!agentModeState || !agentModeState.mode) {
        if (agentPill) {
          agentPill.remove();
          agentPill = null;
        }
        if (agentReset) {
          agentReset.remove();
          agentReset = null;
        }
        return;
      }
      if (!agentPill) {
        agentPill = document.createElement("button");
        agentPill.id = "agent-mode-pill";
        agentPill.className = "wp-agent-pill";
        agentPill.type = "button";
        agentPill.addEventListener("click", cycleAgentMode);
        actions.appendChild(agentPill);
      }
      var pinned = agentModeState.source === "runtime";
      agentPill.classList.toggle("pinned", pinned);
      agentPill.textContent =
        "AI drafts: " + agentModeState.mode + (pinned ? " · pinned" : "");
      if (pinned) {
        if (!agentReset) {
          agentReset = document.createElement("button");
          agentReset.id = "agent-mode-reset";
          agentReset.className = "wp-agent-reset";
          agentReset.type = "button";
          agentReset.addEventListener("click", function () {
            postAgentMode("reset");
          });
          actions.appendChild(agentReset);
        }
        agentReset.textContent = "reset";
      } else if (agentReset) {
        agentReset.remove();
        agentReset = null;
      }
    };
    /** @param {string} value */
    var postAgentMode = function (value) {
      var state = /** @type {AgentModeState} */ (agentModeState);
      if (!state) return;
      fetch("/messages/drafts/mode", {
        method: "POST",
        headers: { "X-CSRF-Token": csrfToken() },
        credentials: "same-origin",
        body: new URLSearchParams({ mode: value }).toString(),
      })
        .then(function (res) {
          if (!res.ok) throw new Error("HTTP " + res.status);
          if (value === "reset") {
            state.mode = state.defaultMode;
            state.source = "config";
            announceLive("AI drafts back to " + state.defaultMode + ".");
          } else {
            state.mode = value;
            state.source = value === state.defaultMode ? "config" : "runtime";
            announceLive("AI drafts set to " + value + ".");
          }
          renderAgentPill();
        })
        .catch(function () {
          announceLive("Could not change the AI drafts mode.");
        });
    };
    var cycleAgentMode = function () {
      if (
        !agentModeState ||
        !agentModeState.modes ||
        !agentModeState.modes.length
      )
        return;
      var modes = agentModeState.modes;
      var at = modes.indexOf(agentModeState.mode);
      postAgentMode(modes[(at + 1 + modes.length) % modes.length]);
    };
    /** @param {Event} event */
    var agentModeOpened = function (event) {
      /** @type {{agentMode?: string, agentModeSource?: string, agentDefaultMode?: string, agentModes?: string[]}} */
      var detail = /** @type {CustomEvent} */ (event).detail || {};
      if (!detail.agentMode) {
        agentModeState = null;
      } else {
        agentModeState = {
          mode: detail.agentMode,
          source: detail.agentModeSource || "config",
          defaultMode: detail.agentDefaultMode || detail.agentMode,
          modes: Array.isArray(detail.agentModes) ? detail.agentModes : [],
        };
      }
      renderAgentPill();
    };
    document.addEventListener("wp:session-opened", agentModeOpened);

    // 2e. Reply snippets (M21): a chip or picker button carries the body
    //     itself (data-snippet). The fill REPLACES the reply textarea's
    //     content (the data-sms prefill precedent — a snippet is a whole
    //     answer, not a fragment) and focuses the box for a quick edit
    //     before send. The buttons live INSIDE the reply form, so
    //     closest() finds the compose box; a fill from the picker also
    //     closes the disclosure.
    document.addEventListener("click", function (event) {
      var snippet = closestHTMLElement(eventElement(event), "[data-snippet]");
      if (!snippet) return;
      var form = closestHTMLElement(snippet, "form.wp-compose");
      var body = /** @type {HTMLTextAreaElement | null} */ (
        form ? form.querySelector("textarea.wp-compose-body") : null
      );
      if (!body) return;
      body.value = snippet.getAttribute("data-snippet") || "";
      body.focus();
      var details = closestHTMLElement(snippet, "details");
      var disclosure = /** @type {HTMLDetailsElement | null} */ (details);
      if (disclosure) disclosure.open = false;
    });

    // 2f. Attachment lightbox (M21.1): an image attachment link carries
    //     data-lightbox; clicking it opens the full-size blob in a
    //     singleton native <dialog> instead of navigating away (the
    //     download stays available via the browser's image context
    //     menu). ESC and backdrop clicks close it natively — CSP-clean,
    //     zero keyboard JS. English shell copy (D3).
    /** @type {HTMLDialogElement | null} */
    var lightbox = null;
    /** @type {HTMLImageElement | null} */
    var lightboxImg = null;
    document.addEventListener("click", function (event) {
      var link = closestHTMLElement(eventElement(event), "a[data-lightbox]");
      if (!link) return;
      event.preventDefault();
      if (!lightbox) {
        lightbox = document.createElement("dialog");
        lightbox.id = "wp-lightbox";
        lightbox.className = "wp-lightbox";
        lightboxImg = document.createElement("img");
        lightbox.append(lightboxImg);
        var close = document.createElement("button");
        close.type = "button";
        close.className = "wp-mini wp-lightbox-close";
        close.textContent = "✕ Close";
        close.setAttribute("aria-label", "Close");
        close.addEventListener("click", function () {
          if (lightbox) lightbox.close();
        });
        lightbox.append(close);
        // A click on the dialog itself (the padding around the image)
        // reads as "backdrop" — close. The image and the button swallow
        // their own clicks.
        lightbox.addEventListener("click", function (e) {
          if (lightbox && e.target === lightbox) lightbox.close();
        });
        document.body.append(lightbox);
      }
      // Bound once: the module-level `var`s are not narrowed inside the
      // listener closures above, and both are assigned by now.
      var dialog = /** @type {HTMLDialogElement | null} */ (lightbox);
      var img = /** @type {HTMLImageElement | null} */ (lightboxImg);
      if (!dialog || !img) return;
      img.src = link.getAttribute("href") || "";
      img.alt = link.getAttribute("data-lightbox") || "";
      dialog.showModal();
    });

    // 2b. data-reload buttons (error panel): full reload, same as the old
    //      inline onclick but CSP-safe via this delegated listener.
    document.addEventListener("click", function (event) {
      if (!closestHTMLElement(eventElement(event), "[data-reload]")) return;
      location.reload();
    });

    // 2b-2. data-transcribe-src buttons (Voicemail + Messages tabs): fetch
    //     the audio at the given same-origin URL, POST it to the ASR seam
    //     (/api/transcribe), and render the text into the target node. One
    //     delegated handler covers every audio stream the tabs render —
    //     voicemail rows and MMS audio attachments alike. Shell copy stays
    //     English (D3); the button LABEL is localized by the server.
    //     A 429 from the seam's flood budget is retried ONCE after the
    //     Retry-After delay (bounded to 10 s — a hostile header must not
    //     freeze the tab).
    /**
     * @param {Blob} blob
     * @param {string} name
     * @param {boolean} retried429
     * @returns {Promise<Response>}
     */
    var postTranscribe = function (blob, name, retried429) {
      return fetch("/api/transcribe?filename=" + encodeURIComponent(name), {
        method: "POST",
        credentials: "same-origin",
        headers: {
          "X-CSRF-Token": csrfToken(),
          "Content-Type": blob.type || "application/octet-stream",
        },
        body: blob,
      }).then(function (res) {
        if (res.status === 429 && !retried429) {
          var seconds = parseInt(res.headers.get("Retry-After") || "", 10);
          if (!(seconds > 0) || seconds > 10) seconds = 1;
          return new Promise(function (resolve) {
            setTimeout(function () {
              resolve(postTranscribe(blob, name, true));
            }, seconds * 1000);
          });
        }
        return res;
      });
    };
    /** @param {Element} btn @returns {Promise<void>} */
    var runTranscribe = function (btn) {
      var src = btn.getAttribute("data-transcribe-src");
      var targetId = btn.getAttribute("data-transcribe-target");
      var found = targetId ? document.getElementById(targetId) : null;
      if (!src || !found) return Promise.resolve();
      var source = src;
      var target = found;
      /** @param {string} text */
      var render = function (text) {
        target.hidden = false;
        target.textContent = text;
      };
      // A real result plants a copy chip next to the text (once per
      // target per DOM lifetime — a morph swap rebuilds the row, the
      // next transcription re-plants). The chip only references the
      // target; the copy handler in 2b5 owns the clipboard write.
      var plantCopyChip = function () {
        // The chip is a convenience: a node that cannot take a sibling
        // (detached stub, exotic parent) must never fail the render it
        // accompanies.
        try {
          if (!target.id) return;
          if (document.querySelector('[data-copy-target="' + target.id + '"]'))
            return;
          var chip = document.createElement("button");
          chip.type = "button";
          chip.className = "wp-mini";
          chip.setAttribute("data-copy-target", target.id);
          chip.textContent = "Copy";
          target.insertAdjacentElement("afterend", chip);
        } catch {
          // A detached target only loses its copy chip; the copy command
          // path stays intact.
        }
      };
      render("Transcribing…");
      return fetch(source, { credentials: "same-origin" })
        .then(function (res) {
          if (!res.ok) throw new Error("HTTP " + res.status);
          return res.blob();
        })
        .then(function (blob) {
          var name = source.split("/").pop() || "audio";
          return postTranscribe(blob, name, false);
        })
        .then(function (res) {
          if (!res.ok) throw new Error("HTTP " + res.status);
          return res.json();
        })
        .then(function (data) {
          var text = (data.text || "").trim();
          render(text || "No speech detected");
          if (text) plantCopyChip();
        })
        .catch(function (err) {
          render("Transcription failed (" + errorMessage(err) + ")");
        });
    };
    document.addEventListener("click", function (event) {
      var btn = closestHTMLElement(
        eventElement(event),
        "[data-transcribe-src]",
      );
      if (!btn) return;
      runTranscribe(btn);
    });

    // 2b-3. Auto-start (owner decision 2026-10-07): with the ASR seam on,
    //     every rendered audio stream transcribes itself — the button
    //     stays as the manual re-run. Once per src+target per page life:
    //     morph re-renders of the same row never re-POST, and a target
    //     that already carries text (a prior run preserved by the morph)
    //     is left alone. Debounced so a burst of swaps coalesces, and
    //     SERIALIZED through one promise chain: a dozen rendered clips
    //     fire one at a time instead of stampeding the flood budget into
    //     429s (manual clicks stay immediate — a human paces themselves).
    var transcribeAutoSeen = new Set();
    var transcribeAutoQueue = Promise.resolve();
    var autoTranscribe = function () {
      if (!(window.PBX_CONFIG && window.PBX_CONFIG.asr)) return;
      var buttons = document.querySelectorAll("[data-transcribe-src]");
      Array.prototype.forEach.call(buttons, function (btn) {
        var key =
          btn.getAttribute("data-transcribe-src") +
          "\u2192" +
          btn.getAttribute("data-transcribe-target");
        if (transcribeAutoSeen.has(key)) return;
        transcribeAutoSeen.add(key);
        var targetId = btn.getAttribute("data-transcribe-target");
        var target = targetId ? document.getElementById(targetId) : null;
        if (!target) return;
        if (target.textContent && target.textContent.trim() !== "") return;
        transcribeAutoQueue = transcribeAutoQueue.then(function () {
          return runTranscribe(btn);
        });
      });
    };
    /** @type {ReturnType<typeof setTimeout> | null} */
    var autoTranscribeTimer = null;
    var scheduleAutoTranscribe = function () {
      if (autoTranscribeTimer) clearTimeout(autoTranscribeTimer);
      autoTranscribeTimer = setTimeout(function () {
        autoTranscribeTimer = null;
        autoTranscribe();
      }, 300);
    };
    document.addEventListener("htmx:afterSwap", scheduleAutoTranscribe);
    scheduleAutoTranscribe();

    // 3. Keep the transcript pinned to the newest message after renders.
    var scrollTranscript = function () {
      var transcript = document.getElementById("thread-transcript");
      if (transcript) transcript.scrollTop = transcript.scrollHeight;
    };
    document.addEventListener("htmx:afterSwap", scrollTranscript);
    scrollTranscript();

    // 3b. Live-transcript polish for SSE "thread" pushes (the sse
    //     extension swaps #thread-transcript's bubbles directly):
    //     - while older pages are open (data-page != "0") the push is
    //       cancelled, so paging state survives the live event;
    //     - on the newest page the push marks the thread read — a live
    //       swap never re-GETs the partial, so only the client can clear
    //       the unread badge for the conversation on screen.
    /** @returns {string} */
    var csrfToken = function () {
      var meta = document.querySelector('meta[name="csrf-token"]');
      return (meta && meta.getAttribute("content")) || "";
    };
    var refreshNav = function () {
      if (!window.htmx) return;
      var active = /** @type {HTMLElement | null} */ (
        document.querySelector("#wp-nav .wp-nav-link.wp-active")
      );
      var source = active
        ? "/partials/nav?active=" + active.dataset.tab
        : "/partials/nav";
      window.htmx.ajax("GET", source, {
        target: "#wp-nav",
        swap: "morph:innerHTML",
      });
      // E8: a morph re-render can leave the ACTIVE tab outside the
      // strip's scroll window on narrow screens; "nearest" scrolls
      // only when it is genuinely out of view.
      if (active && active.scrollIntoView) {
        active.scrollIntoView({ block: "nearest", inline: "nearest" });
      }
    };
    document.addEventListener("htmx:sseBeforeMessage", function (event) {
      var transcript = eventElement(event);
      if (!transcript || transcript.id !== "thread-transcript") return;
      if (transcript.dataset.page !== "0") event.preventDefault();
    });
    document.addEventListener("htmx:sseMessage", function (event) {
      var transcript = eventElement(event);
      if (!transcript || transcript.id !== "thread-transcript") return;
      if (transcript.dataset.page !== "0" || !transcript.dataset.thread) return;
      // Politely announce the arrival (the live region is SR-only, so
      // sighted readers only see the bubble itself).
      var live = document.getElementById("wp-live");
      if (live) {
        live.textContent = "";
        live.textContent = "New message";
      }
      fetch("/messages/" + transcript.dataset.thread + "/read", {
        method: "POST",
        headers: { "X-CSRF-Token": csrfToken() },
        credentials: "same-origin",
      })
        .then(refreshNav)
        .catch(function () {});
    });

    // 3b-2. Jump-to-latest chip: a live push that lands while the reader
    //     has scrolled up must neither yank them to the bottom nor go
    //     unseen. The near-bottom decision happens BEFORE the swap
    //     (sseBeforeMessage); after it, a near-bottom reader keeps the
    //     pinned-scroll behavior, a scrolled-away reader gets a counter
    //     chip that scrolls back on click. Returning to the bottom by
    //     hand hides and resets it. Text is language-neutral ("↓ N new"),
    //     matching the segcounter precedent.
    var NEAR_BOTTOM_PX = 80;
    var pendingNew = 0;
    /** @type {HTMLElement | null} */
    var jumpChip = null;
    /** @param {HTMLElement} el @returns {boolean} */
    var nearBottom = function (el) {
      return el.scrollHeight - el.scrollTop - el.clientHeight <= NEAR_BOTTOM_PX;
    };
    var resetJumpChip = function () {
      pendingNew = 0;
      if (jumpChip) {
        jumpChip.hidden = true;
        jumpChip.textContent = "";
      }
    };
    document.addEventListener(
      "scroll",
      function (event) {
        var transcript = document.getElementById("thread-transcript");
        if (!transcript || eventElement(event) !== transcript) return;
        if (nearBottom(transcript)) resetJumpChip();
      },
      true,
    );
    document.addEventListener("htmx:sseBeforeMessage", function (event) {
      var transcript = eventElement(event);
      if (!transcript || transcript.id !== "thread-transcript") return;
      transcript.dataset.wasNearBottom = nearBottom(transcript) ? "1" : "0";
    });
    document.addEventListener("htmx:sseMessage", function (event) {
      var transcript = eventElement(event);
      if (!transcript || transcript.id !== "thread-transcript") return;
      if (transcript.dataset.page !== "0" || !transcript.dataset.thread) return;
      if (transcript.dataset.wasNearBottom === "0") {
        pendingNew += 1;
        var wrap = transcript.parentElement;
        var chip = /** @type {HTMLElement | null} */ (
          wrap && wrap.querySelector(".wp-jump-latest")
        );
        if (chip) {
          jumpChip = chip;
          chip.textContent = "↓ " + pendingNew + " new";
          chip.hidden = false;
        }
        return;
      }
      // The swap lands right after this event: pin to the newest bubble
      // on the next tick, once the new bubble is in the DOM.
      setTimeout(function () {
        var live = document.getElementById("thread-transcript");
        if (live) live.scrollTop = live.scrollHeight;
      }, 0);
    });
    document.addEventListener("click", function (event) {
      var chip = closestHTMLElement(eventElement(event), ".wp-jump-latest");
      if (!chip) return;
      var transcript = document.getElementById("thread-transcript");
      if (transcript) transcript.scrollTop = transcript.scrollHeight;
      resetJumpChip();
    });
    // A full swap renders a fresh hidden chip — drop the stale counter.
    document.addEventListener("htmx:afterSwap", function () {
      if (document.getElementById("thread-transcript")) {
        pendingNew = 0;
        jumpChip = null;
      }
    });

    // 3b-3. Thread-search guard: a live "threads" push re-renders the
    //     UNFILTERED list, so while the reader has a query in the search
    //     box that swap would stomp the filtered view and lie about the
    //     matches. Cancel the push until the box is empty again — the
    //     next keystroke's debounced fetch re-renders the list anyway.
    document.addEventListener("htmx:sseBeforeMessage", function (event) {
      var list = eventElement(event);
      if (
        !list ||
        !list.classList ||
        !list.classList.contains("wp-thread-list")
      )
        return;
      var input = /** @type {HTMLInputElement | null} */ (
        document.getElementById("wp-thread-search-input")
      );
      if (input && input.value.trim() !== "") event.preventDefault();
    });
    // The island's language switch re-labels itself and re-fetches the open
    // tab; the nav is shell territory, so it asks via this event.
    document.addEventListener("wp:lang-changed", refreshNav);
    // A freshly summarized call became a draft in Messages (island
    // summarize affordance); the nav badges are shell territory too.
    document.addEventListener("wp:drafts-changed", refreshNav);

    // 3c. Error surfacing for htmx requests: htmx swaps NOTHING on error
    //     responses (responseHandling defaults: 4xx/5xx → error, no swap),
    //     so a dead tab session (the session store is in-memory and dies
    //     with every server restart) used to make every tab click and
    //     form submit fail silently. Toast instead. Never reload: the SIP
    //     registration is independent of the server session, so calls
    //     survive and the user reloads when convenient. Responses that
    //     carry HX-Trigger (validation, panel errors) already toast through
    //     the island's showMessage listener, so skip those to avoid double
    //     feedback. An expired session 401s every later request (SSE nav
    //     refreshes included), so the throttle collapses the storm into
    //     one toast instead of hiding its own message.
    var lastErrorToast = 0;
    // The shell's clock seam (type-gate plan item 19): every time read in
    // this file goes through ONE replaceable holder, so specs drive time
    // through the exported seam instead of monkey-patching Date.now or
    // installing a window global. Call-time indirection keeps the holder
    // swappable after this script loaded.
    var clock = {
      now: function () {
        return Date.now();
      },
    };
    /** @param {string} message */
    var showThrottledError = function (message) {
      var nowMs = clock.now();
      if (nowMs - lastErrorToast < 8000) return;
      lastErrorToast = nowMs;
      shellToast(message, "error");
    };
    document.addEventListener("htmx:responseError", function (event) {
      var xhr = eventDetail(event) && eventDetail(event).xhr;
      if (!xhr) return;
      if (xhr.getResponseHeader("HX-Trigger")) return;
      if (xhr.status === 401) {
        showThrottledError(
          "Tab session ended; calls keep working. Reload to sign back in.",
        );
        return;
      }
      if (xhr.status === 429) {
        // Rate limiting is client-correctable, so the toast says what to
        // DO. The server's limiter is generic middleware with no toast
        // header, so client text owns this feedback (plan T03 decision).
        showThrottledError(
          "Too many requests — wait a moment, then try again.",
        );
        return;
      }
      showThrottledError("The request failed (HTTP " + xhr.status + ").");
    });
    document.addEventListener("htmx:sendError", function () {
      showThrottledError("Network request failed; check your connection.");
    });

    // 3d. Per-thread draft persistence (plan T21d): composer text
    //     survives tab and thread switches — the two paths that
    //     re-render the composer empty. Live SSE pushes morph-preserve
    //     the node, so drafts already survive those. Saved debounced on
    //     input, restored ONLY into an empty composer (never clobbers
    //     fresh typing), cleared after a successful send. Best-effort:
    //     storage failures (private mode, quota) are silently ignored.
    var DRAFT_MAX = 4000;
    /** @param {string} threadId @returns {string} */
    var draftKey = function (threadId) {
      return "wp-draft:" + threadId;
    };
    var currentDraftId = function () {
      var transcript = document.getElementById("thread-transcript");
      return transcript && transcript.dataset.thread
        ? transcript.dataset.thread
        : "new";
    };
    /** @param {string} draftId @param {string} text */
    var saveDraft = function (draftId, text) {
      try {
        if (text && text.length <= DRAFT_MAX) {
          localStorage.setItem(draftKey(draftId), text);
        } else {
          localStorage.removeItem(draftKey(draftId));
        }
      } catch {
        /* storage unavailable — drafts are best-effort */
      }
    };
    /** @type {ReturnType<typeof setTimeout> | null} */
    var draftTimer = null;
    document.addEventListener("input", function (event) {
      var area = /** @type {HTMLTextAreaElement} */ (eventElement(event));
      if (
        !area.closest ||
        !closestHTMLElement(area, "textarea.wp-compose-body")
      )
        return;
      if (draftTimer) clearTimeout(draftTimer);
      draftTimer = setTimeout(function () {
        saveDraft(currentDraftId(), area.value);
      }, 300);
    });
    document.addEventListener("htmx:afterSwap", function () {
      var area = /** @type {HTMLTextAreaElement | null} */ (
        document.querySelector("textarea.wp-compose-body")
      );
      if (!area || area.value !== "") return;
      var draft = null;
      try {
        draft = localStorage.getItem(draftKey(currentDraftId()));
      } catch {
        return;
      }
      if (draft) area.value = draft;
    });
    document.addEventListener("htmx:afterRequest", function (event) {
      var form = eventElement(event);
      if (!form || !form.matches || !form.matches("form")) return;
      if (!eventDetail(event) || !eventDetail(event).successful) return;
      if (!form.querySelector("textarea.wp-compose-body")) return;
      saveDraft(currentDraftId(), "");
    });

    // 3e. Optimistic send: the reply composer appends a pending bubble
    //     the instant the form submits, so a send reads as instant
    //     instead of waiting on the POST round-trip. The server's
    //     response (or an SSE push carrying the sent message) swaps the
    //     transcript and replaces the node with the real bubble; on
    //     failure the bubble flips to a failed state and the typed text
    //     is restored, so a retry never loses the draft. The
    //     new-conversation composer has no transcript to append to and
    //     is skipped. Shell copy stays English (D3) — the pending state
    //     is transient by design.
    var optPending = new WeakMap();
    document.addEventListener("htmx:beforeRequest", function (event) {
      var form =
        (eventDetail(event) && eventDetail(event).elt) || eventElement(event);
      if (!form || !form.matches || !form.matches("form.wp-compose")) return;
      if (form.classList.contains("wp-compose-new")) return;
      var transcript = document.getElementById("thread-transcript");
      if (!transcript) return;
      var area = form.querySelector("textarea.wp-compose-body");
      var body = area ? area.value : "";
      var fileInput = form.querySelector('input[type="file"]');
      var names = [];
      if (fileInput && fileInput.files) {
        for (var i = 0; i < fileInput.files.length; i++)
          names.push(fileInput.files[i].name);
      }
      if (!body.trim() && names.length === 0) return;
      var bubble = document.createElement("div");
      bubble.className = "wp-bubble wp-out wp-opt";
      if (body) {
        var text = document.createElement("p");
        text.className = "wp-bubble-body";
        text.textContent = body;
        bubble.append(text);
      }
      names.forEach(function (name) {
        var chip = document.createElement("span");
        chip.className = "wp-attachment";
        chip.innerHTML = WP_ATTACH_SVG;
        chip.append(document.createTextNode(" " + name));
        bubble.append(chip);
      });
      var meta = document.createElement("span");
      meta.className = "wp-bubble-meta";
      var clock = document.createElement("span");
      var stamp = new Date();
      /** @param {number} n @returns {string} */
      var pad = function (n) {
        return (n < 10 ? "0" : "") + n;
      };
      clock.textContent = pad(stamp.getHours()) + ":" + pad(stamp.getMinutes());
      var status = document.createElement("span");
      status.className = "wp-status wp-status-queued";
      status.textContent = "sending";
      meta.append(clock, status);
      bubble.append(meta);
      transcript.append(bubble);
      transcript.scrollTop = transcript.scrollHeight;
      optPending.set(form, { bubble: bubble, body: body });
    });
    /** @param {Event} event */
    var rollbackOptimistic = function (event) {
      /** @type {{elt?: Element}} */
      var detail = /** @type {CustomEvent} */ (event).detail || {};
      var form = detail.elt || eventElement(event);
      if (!form || !form.matches || !form.matches("form.wp-compose")) return;
      var record = optPending.get(form);
      if (!record) return;
      optPending.delete(form);
      var status = record.bubble.querySelector(".wp-status");
      if (status) {
        status.className = "wp-status wp-status-failed";
        status.textContent = "failed";
      }
      record.bubble.classList.remove("wp-opt");
      record.bubble.classList.add("wp-opt-failed");
      var area = /** @type {HTMLTextAreaElement | null} */ (
        form.querySelector("textarea.wp-compose-body")
      );
      if (area && area.value === "" && record.body) area.value = record.body;
      // M17 J3/J4: the failure surface carries its own recovery. Retry
      // re-submits the reply composer (the draft was restored above, so
      // the exact text goes out again); Dismiss drops the failed bubble
      // while the draft stays in the composer for manual editing. Both
      // look in the CURRENT document for the form — a tab swap since the
      // failure would have replaced this bubble along with the thread.
      var actions = document.createElement("span");
      actions.className = "wp-opt-actions";
      var retry = document.createElement("button");
      retry.type = "button";
      retry.className = "wp-mini";
      retry.textContent = "Retry";
      retry.addEventListener("click", function () {
        record.bubble.remove();
        // Scoped to the messages panel: the contacts add/import forms are
        // also .wp-compose, so a document-wide query can rebuild the wrong
        // composer (or none) when both are open.
        var current = /** @type {HTMLFormElement | null} */ (
          document.querySelector(
            "#messages-panel form.wp-compose:not(.wp-compose-new)",
          )
        );
        if (current) current.requestSubmit();
      });
      var dismiss = document.createElement("button");
      dismiss.type = "button";
      dismiss.className = "wp-mini";
      dismiss.textContent = "Dismiss";
      dismiss.addEventListener("click", function () {
        record.bubble.remove();
      });
      actions.append(retry, dismiss);
      record.bubble.append(actions);
    };
    document.addEventListener("htmx:responseError", rollbackOptimistic);
    document.addEventListener("htmx:sendError", rollbackOptimistic);
    document.addEventListener("htmx:afterRequest", function (event) {
      var form =
        (eventDetail(event) && eventDetail(event).elt) || eventElement(event);
      if (!form || !form.matches || !form.matches("form.wp-compose")) return;
      if (eventDetail(event) && eventDetail(event).successful)
        optPending.delete(form);
    });

    // 3f. Focus lands on the new panel's heading after a navigation
    //     swap (tab click, thread open, back link): keyboard and
    //     screen-reader users start reading at the top of what just
    //     arrived instead of wherever focus happened to be. Typing-
    //     driven swaps (search, composer, drafts) deliberately keep
    //     focus where the user is.
    document.addEventListener("htmx:afterSwap", function (event) {
      var elt = eventElement(event);
      if (!hasMethod(elt, "matches")) return;
      var navigating =
        elt.hasAttribute && elt.hasAttribute("data-tab")
          ? true
          : Boolean(
              hasMethod(elt, "closest") &&
              closestHTMLElement(elt, ".wp-thread-rowwrap, .wp-back"),
            );
      if (!navigating) return;
      var heading = /** @type {HTMLElement | null} */ (
        document.querySelector("#tab-content h2")
      );
      if (!heading || !hasMethod(heading, "focus")) return;
      heading.setAttribute("tabindex", "-1");
      heading.focus({ preventScroll: true });
    });

    // 3g. Relative-time tick: thread-list stamps ("3m") go stale while
    //     the page sits open. Every 30s the client recomputes the short
    //     forms from the element's data-when epoch — deliberately the
    //     same language-neutral vocabulary the server renders
    //     (now / Nm / Nh). Stamps older than a day already carry an
    //     absolute date and are left untouched.
    var tickRelative = function () {
      var nodes = document.querySelectorAll("[data-when]");
      for (var i = 0; i < nodes.length; i++) {
        var node = /** @type {HTMLElement} */ (nodes[i]);
        var seconds =
          Math.floor(clock.now() / 1000) - Number(node.dataset.when || 0);
        if (!(seconds >= 0)) continue;
        if (seconds < 60) node.textContent = "now";
        else if (seconds < 3600)
          node.textContent = Math.floor(seconds / 60) + "m";
        else if (seconds < 86400)
          node.textContent = Math.floor(seconds / 3600) + "h";
      }
    };
    setInterval(tickRelative, 30000);

    // 3h. Tab skeleton (F1): navigation swaps (tab links, thread rows,
    //     back) reveal the shimmer placeholder while the partial is in
    //     flight. Typing-driven fetches (search, composer) stay quiet —
    //     morph keeps those surfaces alive and a flash there would be
    //     noise, not signal.
    /** @type {HTMLElement | null} */
    var skeleton = null;
    /** @param {Element} elt @returns {boolean} */
    var navigatingSwap = function (elt) {
      if (!hasMethod(elt, "matches")) return false;
      return Boolean(
        (elt.hasAttribute && elt.hasAttribute("data-tab")) ||
        (hasMethod(elt, "closest") &&
          closestHTMLElement(elt, ".wp-thread-rowwrap, .wp-back")),
      );
    };
    document.addEventListener("htmx:beforeRequest", function (event) {
      if (!navigatingSwap(eventElement(event))) return;
      skeleton = skeleton || document.getElementById("wp-tab-skeleton");
      if (skeleton) skeleton.hidden = false;
    });
    var hideSkeleton = function () {
      if (skeleton) skeleton.hidden = true;
    };
    document.addEventListener("htmx:afterRequest", hideSkeleton);
    document.addEventListener("htmx:responseError", hideSkeleton);
    document.addEventListener("htmx:sendError", hideSkeleton);

    // 3i. Voicemail player chrome (M13 C1–C3/C9/C10): each voicemail row
    //     ships a controls-free <audio> plus custom chrome (play button,
    //     waveform canvas, speed toggle, time readout). This section is
    //     the chrome's engine — delegated at document level so tab swaps
    //     never need re-wiring; per-audio listeners attach lazily on
    //     first interaction. The audio element is the single source of
    //     truth (idiomorph keeps it alive across panel morphs by id), and
    //     when waveform decoding is impossible (no WebAudio/fetch) the
    //     honest fallback hands playback back to the browser's native
    //     controls instead of a dead scrubber.
    // uuid -> peak array (48 bars). CAPPED: the panel can hold hundreds of
    // rows over a long-lived session, and each entry is a decoded waveform
    // the page never releases. Oldest-first eviction — a re-decode is cheap
    // (the audio is already cached by the browser), an unbounded Map is a
    // slow leak.
    var vmPeaksCache = new Map();
    var VM_PEAKS_MAX = 48;
    /** @param {string} uuid @param {number[]} peaks */
    var vmCachePeaks = function (uuid, peaks) {
      vmPeaksCache.delete(uuid); // re-insert so the Map stays LRU-ordered
      vmPeaksCache.set(uuid, peaks);
      while (vmPeaksCache.size > VM_PEAKS_MAX) {
        var oldest = vmPeaksCache.keys().next();
        if (oldest.done) break;
        vmPeaksCache.delete(oldest.value);
      }
    };
    var vmWired = new WeakSet(); // audio elements with listeners attached
    /** @type {HTMLAudioElement | null} */
    var vmActiveAudio = null; // single-active-player invariant
    /** @type {AudioContext | null} */
    var vmAudioContext = null; // browsers cap contexts — one per page, lazy
    var vmFallbackToasted = false;

    var VM_BARS = 48;
    var VM_SPEED_LADDER = [1, 1.5, 2, 0.5];

    // vmPeaks reduces decoded PCM (−1..1) to one normalized max-abs peak
    // per bar. Pure and unit-tested: empty input renders silence, and a
    // bar whose sample slice is empty (more bars than samples) inherits
    // the nearest valued bar so the wave starts where the audio does.
    /**
     * @param {Float32Array} samples
     * @param {number} bars
     * @returns {number[]}
     */
    var vmPeaks = function (samples, bars) {
      var peaks = new Array(bars).fill(0);
      if (!samples || samples.length === 0) return peaks;
      var step = samples.length / bars;
      /** @type {Array<number | null>} */
      var filled = new Array(bars).fill(null);
      for (var b = 0; b < bars; b++) {
        var start = Math.floor(b * step);
        var end = Math.min(samples.length, Math.floor((b + 1) * step));
        if (end <= start) continue;
        var max = 0;
        for (var i = start; i < end; i++) {
          var v = Math.abs(samples[i]);
          if (v > max) max = v;
        }
        filled[b] = max;
      }
      // stretch empty bars outward from their nearest valued neighbor
      var last = null;
      for (var k = 0; k < bars; k++) {
        if (filled[k] !== null) last = filled[k];
        else if (last !== null) filled[k] = last;
      }
      var next = null;
      for (var k2 = bars - 1; k2 >= 0; k2--) {
        if (filled[k2] !== null) next = filled[k2];
        else if (next !== null) filled[k2] = next;
      }
      for (var m = 0; m < bars; m++) {
        if (filled[m] !== null) peaks[m] = filled[m];
      }
      return peaks;
    };

    // vmClock mirrors the server's m:ss shape (vmClock in voicemail.templ)
    // so the running readout never disagrees with the initial render.
    /** @param {number} seconds @returns {string} */
    var vmClock = function (seconds) {
      if (!isFinite(seconds) || seconds < 0) seconds = 0;
      var s = Math.floor(seconds);
      var mm = Math.floor(s / 60);
      var ss = s % 60;
      return mm + ":" + (ss < 10 ? "0" : "") + ss;
    };

    /** @param {Element | null} el @returns {Element | null} */
    var vmRowOf = function (el) {
      if (!el) return null;
      return hasMethod(el, "closest")
        ? closestHTMLElement(el, ".wp-vm-row")
        : null;
    };

    /**
     * @param {HTMLElement} canvas
     * @param {string} name
     * @param {string} fallback
     * @returns {string}
     */
    var vmToken = function (canvas, name, fallback) {
      if (typeof getComputedStyle !== "function") return fallback;
      var value = getComputedStyle(canvas).getPropertyValue(name);
      return value && value.trim() ? value.trim() : fallback;
    };

    /**
     * @param {HTMLCanvasElement} canvas
     * @param {number[]} peaks
     * @param {number} progress
     */
    var vmDraw = function (canvas, peaks, progress) {
      if (!canvas || !canvas.getContext) return;
      var ctx = canvas.getContext("2d");
      if (!ctx) return;
      var w = canvas.width || 300;
      var h = canvas.height || 28;
      ctx.clearRect(0, 0, w, h);
      var bars = peaks.length;
      if (!bars) return;
      var gap = 1;
      var bw = Math.max(1, Math.floor((w - gap * (bars - 1)) / bars));
      var playedColor = vmToken(canvas, "--accent-strong", "#43c79f");
      var restColor = vmToken(canvas, "--muted", "#8fa1aa");
      for (var i = 0; i < bars; i++) {
        var bh = Math.max(2, Math.round(peaks[i] * (h - 2)));
        var x = i * (bw + gap);
        var y = (h - bh) / 2;
        ctx.fillStyle = i / bars <= progress ? playedColor : restColor;
        ctx.fillRect(x, y, bw, bh);
      }
    };

    // vmClearUnread (C9): the row's unread styling drops on first play,
    // and the nav badge count follows (the server re-renders the real
    // counts on the next nudge — this is the same optimistic shape the
    // mark-read path uses, grounded in the user's own play action).
    /** @param {Element | null} row */
    var vmClearUnread = function (row) {
      if (!row || !row.classList || !row.classList.contains("wp-unread")) {
        return;
      }
      row.classList.remove("wp-unread");
      var badge = document.getElementById("wp-nav-vm-badge");
      if (!badge) return;
      var count = parseInt(badge.textContent, 10);
      if (isNaN(count) || count <= 1) {
        badge.hidden = true;
      } else {
        badge.textContent = String(count - 1);
      }
    };

    /**
     * @param {string} uuid
     * @param {HTMLAudioElement} audio
     * @param {boolean} playing
     */
    var vmSetPlaying = function (uuid, audio, playing) {
      var row = vmRowOf(audio);
      if (row && row.classList) {
        if (playing) row.classList.add("wp-playing");
        else row.classList.remove("wp-playing");
      }
      var btn = document.getElementById("vm-play-" + uuid);
      if (btn) {
        btn.innerHTML = playing ? WP_PAUSE_SVG : WP_PLAY_SVG;
        var label = btn.getAttribute(
          playing ? "data-label-pause" : "data-label-play",
        );
        if (label) btn.setAttribute("aria-label", label);
      }
      if (playing) vmClearUnread(row);
    };

    /** @param {string} uuid @param {HTMLAudioElement} audio */
    var vmPaint = function (uuid, audio) {
      var duration = isFinite(audio.duration) ? audio.duration : 0;
      var current = isFinite(audio.currentTime) ? audio.currentTime : 0;
      var progress = duration > 0 ? current / duration : 0;
      var time = document.getElementById("vm-time-" + uuid);
      if (time) {
        time.textContent = vmClock(current) + " / " + vmClock(duration);
      }
      var canvas = canvasElement("vm-wave-" + uuid);
      if (canvas) {
        canvas.setAttribute("aria-valuenow", String(Math.floor(current)));
        var peaks = vmPeaksCache.get(uuid);
        if (peaks) vmDraw(canvas, peaks, progress);
      }
    };

    /** @param {HTMLAudioElement} audio @param {string} uuid */
    var vmWire = function (audio, uuid) {
      if (vmWired.has(audio)) return;
      vmWired.add(audio);
      audio.addEventListener("play", function () {
        vmSetPlaying(uuid, audio, true);
      });
      audio.addEventListener("pause", function () {
        vmSetPlaying(uuid, audio, false);
      });
      audio.addEventListener("ended", function () {
        vmSetPlaying(uuid, audio, false);
        vmPaint(uuid, audio);
      });
      audio.addEventListener("timeupdate", function () {
        vmPaint(uuid, audio);
        // Self-heal (C3): a panel morph re-renders the row from server
        // truth and can wipe the playing class mid-playback; the audio
        // still playing IS the truth, so the ~4 Hz repaint re-asserts it.
        if (!audio.paused) vmSetPlaying(uuid, audio, true);
      });
    };

    /** @param {string} uuid @param {HTMLAudioElement | null} audio */
    var vmFallback = function (uuid, audio) {
      // Honest degradation (C1): without WebAudio there is no waveform;
      // the browser's native controls take over so playback still works.
      if (audio) audio.controls = true;
      ["vm-play-", "vm-speed-", "vm-wave-"].forEach(function (prefix) {
        var el = document.getElementById(prefix + uuid);
        if (el) el.hidden = true;
      });
      if (!vmFallbackToasted) {
        vmFallbackToasted = true;
        shellToast(
          "voicemail waveform unavailable — using built-in player controls",
          "warn",
        );
      }
    };

    /**
     * Decode the row's audio once and cache its waveform peaks.
     * @param {string} uuid
     * @param {HTMLAudioElement} audio
     * @returns {Promise<number[] | undefined>}
     */
    var vmBuildWave = function (uuid, audio) {
      if (vmPeaksCache.has(uuid)) {
        return Promise.resolve(vmPeaksCache.get(uuid));
      }
      var src = audio.getAttribute("src");
      var Ctor = window.AudioContext || window.webkitAudioContext;
      if (typeof fetch !== "function" || !src || !Ctor) {
        return Promise.reject(new Error("webaudio unavailable"));
      }
      // Bound once: the promise callbacks below run in a later frame, where
      // the checker can no longer narrow the module-level `vmAudioContext`.
      var context = vmAudioContext;
      if (!context) {
        context = new Ctor();
        vmAudioContext = context;
      }
      return fetch(src)
        .then(function (response) {
          if (!response.ok)
            throw new Error("audio fetch HTTP " + response.status);
          return response.arrayBuffer();
        })
        .then(function (buffer) {
          // Bound once: the promise callbacks below run in a later frame,
          // where the checker can no longer narrow `context`.
          var decoder = /** @type {AudioContext} */ (context);
          return new Promise(function (resolve, reject) {
            // callback form: the oldest widest-compatible signature
            decoder.decodeAudioData(buffer, resolve, reject);
          });
        })
        .then(function (decoded) {
          var peaks = vmPeaks(decoded.getChannelData(0), VM_BARS);
          vmCachePeaks(uuid, peaks);
          return peaks;
        });
    };

    /** @param {HTMLElement} btn */
    var vmTogglePlay = function (btn) {
      var uuid = btn.getAttribute("data-vm-play") || "";
      var audio = audioElement("vm-audio-" + uuid);
      if (!audio || !hasMethod(audio, "play")) return;
      vmWire(audio, uuid);
      if (!audio.paused) {
        audio.pause();
        return;
      }
      // single active player: starting one pauses the previous row
      if (
        vmActiveAudio &&
        vmActiveAudio !== audio &&
        hasMethod(vmActiveAudio, "pause")
      ) {
        vmActiveAudio.pause();
      }
      vmActiveAudio = audio;
      // Bound once: the promise callbacks below run in a later frame, where
      // the checker can no longer narrow the nullable lookup.
      var player = /** @type {HTMLAudioElement} */ (audio);
      var started = audio.play();
      if (started && hasMethod(started, "catch")) {
        started.catch(function () {});
      }
      vmBuildWave(uuid, audio)
        .then(function (peaks) {
          vmPaint(uuid, player);
          var canvas = canvasElement("vm-wave-" + uuid);
          if (canvas) vmDraw(canvas, peaks || [], 0);
        })
        .catch(function () {
          vmFallback(uuid, player);
        });
    };

    /** @param {HTMLElement} btn */
    var vmCycleSpeed = function (btn) {
      var uuid = btn.getAttribute("data-vm-speed") || "";
      var audio = audioElement("vm-audio-" + uuid);
      if (!audio) return;
      var current = Number(audio.playbackRate) || 1;
      var idx = VM_SPEED_LADDER.indexOf(current);
      var next =
        idx >= 0 ? VM_SPEED_LADDER[(idx + 1) % VM_SPEED_LADDER.length] : 1;
      audio.playbackRate = next;
      btn.textContent = next + "×";
    };

    document.addEventListener("click", function (event) {
      var play = /** @type {HTMLButtonElement | null} */ (
        closestHTMLElement(eventElement(event), ".wp-vm-play")
      );
      if (play) {
        vmTogglePlay(play);
        return;
      }
      var speed = /** @type {HTMLButtonElement | null} */ (
        closestHTMLElement(eventElement(event), ".wp-vm-speed")
      );
      if (speed) vmCycleSpeed(speed);
    });

    /**
     * @param {HTMLCanvasElement} canvas
     * @param {HTMLAudioElement} audio
     * @param {number} clientX
     */
    var vmSeek = function (canvas, audio, clientX) {
      if (!canvas.getBoundingClientRect) return;
      var rect = canvas.getBoundingClientRect();
      if (!rect || !rect.width) return;
      var ratio = (clientX - rect.left) / rect.width;
      if (ratio < 0) ratio = 0;
      if (ratio > 1) ratio = 1;
      if (isFinite(audio.duration) && audio.duration > 0) {
        audio.currentTime = ratio * audio.duration;
        vmPaint(canvas.id.slice("vm-wave-".length), audio);
      }
    };

    document.addEventListener("pointerdown", function (event) {
      var canvas = canvasIn(eventElement(event), ".wp-vm-wave");
      if (!canvas) return;
      var uuid = canvas.id.slice("vm-wave-".length);
      var audio = audioElement("vm-audio-" + uuid);
      if (!audio) return;
      vmWire(audio, uuid);
      vmSeek(canvas, audio, event.clientX);
      // drag-to-scrub continues while the pointer is down
      var scrubber = /** @type {HTMLCanvasElement} */ (canvas);
      var scrubAudio = /** @type {HTMLAudioElement} */ (audio);
      /** @param {PointerEvent} moveEvent */
      var move = function (moveEvent) {
        vmSeek(scrubber, scrubAudio, moveEvent.clientX);
      };
      var up = function () {
        document.removeEventListener("pointermove", move);
        document.removeEventListener("pointerup", up);
      };
      document.addEventListener("pointermove", move);
      document.addEventListener("pointerup", up);
    });

    // keyboard seek on the scrubber (it carries role=slider + tabindex)
    document.addEventListener("keydown", function (event) {
      var canvas = canvasIn(eventElement(event), ".wp-vm-wave");
      if (!canvas) return;
      if (event.key !== "ArrowLeft" && event.key !== "ArrowRight") return;
      var uuid = canvas.id.slice("vm-wave-".length);
      var audio = audioElement("vm-audio-" + uuid);
      if (!audio) return;
      var delta = event.key === "ArrowRight" ? 5 : -5;
      audio.currentTime = Math.max(0, (audio.currentTime || 0) + delta);
      vmPaint(uuid, audio);
      event.preventDefault();
    });

    // 4. Manual theme override: cycles auto (prefers-color-scheme) →
    //    light → dark, persisted in localStorage. data-theme on <html>
    //    beats both stylesheets' media queries via attribute specificity.
    var THEME_KEY = "wp-theme";
    var THEMES = ["auto", "light", "dark"];
    var storedTheme = null;
    try {
      storedTheme = localStorage.getItem(THEME_KEY);
    } catch {
      /* storage unavailable (private mode) — fall through to auto */
    }
    var themeIndex = storedTheme ? THEMES.indexOf(storedTheme) : -1;
    if (themeIndex < 0) themeIndex = 0;

    var applyTheme = function () {
      var theme = THEMES[themeIndex];
      if (theme === "auto") {
        document.documentElement.removeAttribute("data-theme");
      } else {
        document.documentElement.setAttribute("data-theme", theme);
      }
      var button = document.getElementById("theme-toggle");
      if (button) {
        button.textContent = "Theme: " + theme;
        // G10: the accessible name carries the current state too, so a
        // screen reader hears the theme change without re-reading the
        // visible label.
        button.setAttribute("aria-label", "Color theme: " + theme);
      }
    };
    applyTheme();

    var toggle = document.getElementById("theme-toggle");
    if (toggle) {
      toggle.addEventListener("click", function () {
        themeIndex = (themeIndex + 1) % THEMES.length;
        try {
          localStorage.setItem(THEME_KEY, THEMES[themeIndex]);
        } catch {
          /* best-effort persistence only */
        }
        applyTheme();
      });
    }

    // 4b. Collapsible phone panel: the island's 320–400px column is the
    //     product while you are calling and dead weight while you read a
    //     thread, a fax list or settings. The class lands on <html> (not
    //     .wp-root) because theme-preload.js applies the stored state
    //     before first paint — the grid's column count changes, so a late
    //     apply would flash the wide grid and snap.
    //
    //     Collapsing HIDES the column but never tears the island down: the
    //     SIP registration, the call cards and every island module keep
    //     running. CSS force-re-expands the column while a call card is
    //     live (`:has(.call-card)`) — hiding the phone mid-call is never
    //     what the user meant, and a hidden RTCPeerConnection would look
    //     exactly like a dropped call.
    var ISLAND_KEY = "wp-island-collapsed";
    var islandToggle = document.getElementById("wp-island-toggle");

    /** @param {boolean} collapsed */
    var applyIslandCollapsed = function (collapsed) {
      if (collapsed) {
        document.documentElement.classList.add("wp-island-collapsed");
      } else {
        document.documentElement.classList.remove("wp-island-collapsed");
      }
      if (islandToggle) {
        // aria-expanded is the state channel: the button's accessible NAME
        // stays constant ("Toggle the phone panel"), which is what makes
        // the control legible in a screen-reader control list.
        islandToggle.setAttribute("aria-expanded", collapsed ? "false" : "true");
      }
    };

    // Re-assert the pre-paint class so aria-expanded matches what is on
    // screen (theme-preload.js cannot touch the button — it has not been
    // parsed yet at that point).
    applyIslandCollapsed(
      document.documentElement.classList.contains("wp-island-collapsed"),
    );

    if (islandToggle) {
      islandToggle.addEventListener("click", function () {
        var collapsed = !document.documentElement.classList.contains(
          "wp-island-collapsed",
        );
        try {
          localStorage.setItem(ISLAND_KEY, collapsed ? "1" : "0");
        } catch {
          /* best-effort persistence only */
        }
        applyIslandCollapsed(collapsed);
      });
    }

    // 5. Command palette (Ctrl/Cmd-K) and shortcut help (?): one
    //    lazily-built overlay, two modes. The palette lists the tabs
    //    plus a few shell actions and filters as you type; the help
    //    mode lists every binding that actually exists (island keys
    //    included) instead of the ones we wish existed. Overlay markup
    //    is JS-created like the toasts — the served DOM contract stays
    //    untouched, and every string is shell English (D3).
    var HELP_ITEMS = [
      ["A", "Answer the incoming call"],
      ["H", "Hang up the focused call"],
      ["M", "Mute / unmute the focused call"],
      ["P", "Hold / resume the focused call"],
      ["Esc", "Reject incoming, or hang up"],
      ["Enter", "Send the message you are typing"],
      ["Shift+Enter", "New line in the composer"],
      ["Ctrl/Cmd+K", "Command palette"],
      ["?", "This help"],
    ];
    /** @type {HTMLElement | null} */
    var overlay = null;
    /** @type {HTMLInputElement | null} */
    var overlayInput = null;
    /** @type {HTMLElement | null} */
    var overlayList = null;
    /** @type {Array<{label: string, run?: () => void}>} */
    var overlayItems = [];
    // The unfiltered row set for the CURRENT open. Built once per open rather
    // than per keystroke: collectCommands() walks the nav and allocates a
    // closure per link, and renderOverlay runs on every input event.
    /** @type {Array<{label: string, run?: () => void}>} */
    var overlaySource = [];
    var overlayIndex = 0;
    /** @type {HTMLElement | null} */
    var overlayReturnFocus = null;

    var collectCommands = function () {
      var items = [];
      var links = document.querySelectorAll("#wp-nav .wp-nav-link");
      for (var i = 0; i < links.length; i++) {
        (function (link) {
          items.push({
            label: "Go to " + link.textContent.trim(),
            run: function () {
              link.click();
            },
          });
        })(/** @type {HTMLElement} */ (links[i]));
      }
      items.push({
        label: "Call a number",
        run: function () {
          var dest = document.getElementById("dest");
          if (dest) dest.focus();
        },
      });
      items.push({
        label: "New message",
        run: function () {
          var messages = /** @type {HTMLElement | null} */ (
            document.querySelector("[data-tab='messages']")
          );
          if (messages) messages.click();
        },
      });
      items.push({
        label: "Cycle color theme",
        run: function () {
          var themeButton = document.getElementById("theme-toggle");
          if (themeButton) themeButton.click();
        },
      });
      return items;
    };

    var renderOverlay = function () {
      var list = overlayList;
      var input = overlayInput;
      if (!list || !input) return;
      list.replaceChildren();
      overlayItems = [];
      overlayIndex = 0;
      var query = input.value.trim().toLowerCase();
      var source = overlaySource;
      for (var i = 0; i < source.length; i++) {
        if (query && source[i].label.toLowerCase().indexOf(query) < 0) continue;
        overlayItems.push(source[i]);
      }
      for (var j = 0; j < overlayItems.length; j++) {
        var li = document.createElement("li");
        li.className = "wp-overlay-item" + (j === 0 ? " wp-selected" : "");
        li.textContent = overlayItems[j].label;
        (function (index) {
          li.addEventListener("click", function () {
            overlayIndex = index;
            runOverlayItem();
          });
        })(j);
        list.append(li);
      }
    };

    /** @param {number} delta */
    var selectOverlay = function (delta) {
      var list = overlayList;
      if (!list || overlayItems.length === 0) return;
      overlayIndex =
        (overlayIndex + delta + overlayItems.length) % overlayItems.length;
      var rows = list.children;
      for (var i = 0; i < rows.length; i++) {
        rows[i].className =
          "wp-overlay-item" + (i === overlayIndex ? " wp-selected" : "");
      }
    };

    var runOverlayItem = function () {
      var item = overlayItems[overlayIndex];
      if (!item) return;
      closeOverlay();
      if (item.run) item.run();
    };

    var buildOverlay = function () {
      var panel = document.createElement("div");
      panel.id = "wp-palette";
      panel.className = "wp-overlay";
      panel.hidden = true;
      panel.setAttribute("role", "dialog");
      panel.setAttribute("aria-label", "Command palette");
      var input = document.createElement("input");
      input.type = "text";
      input.className = "wp-overlay-input";
      input.setAttribute("aria-label", "Search commands");
      input.placeholder = "Type a command…";
      var list = document.createElement("ul");
      list.className = "wp-overlay-list";
      panel.append(input, list);
      document.body.append(panel);
      // Published before the listeners attach: renderOverlay and the
      // keydown handler read the module-level homes, and openOverlay binds
      // its own non-null locals from them.
      overlay = panel;
      overlayInput = input;
      overlayList = list;
      input.addEventListener("input", renderOverlay);
      input.addEventListener("keydown", function (event) {
        if (event.key === "ArrowDown") {
          selectOverlay(1);
          event.preventDefault();
        } else if (event.key === "ArrowUp") {
          selectOverlay(-1);
          event.preventDefault();
        } else if (event.key === "Enter") {
          runOverlayItem();
          event.preventDefault();
        } else if (event.key === "Escape") {
          closeOverlay();
          event.preventDefault();
        }
      });
      panel.addEventListener("click", function (event) {
        if (eventElement(event) === panel) closeOverlay();
      });
    };

    /** @param {string} mode */
    var openOverlay = function (mode) {
      if (!overlay) buildOverlay();
      var panel = overlay;
      var input = overlayInput;
      if (!panel || !input) return;
      overlaySource =
        mode === "commands"
          ? collectCommands()
          : HELP_ITEMS.map(function (pair) {
              return { label: pair[0] + " — " + pair[1] };
            });
      overlayReturnFocus = /** @type {HTMLElement | null} */ (
        document.activeElement
      );
      panel.hidden = false;
      input.value = "";
      input.placeholder =
        mode === "commands" ? "Type a command…" : "Shortcuts — Esc closes";
      renderOverlay();
      input.focus();
    };

    var closeOverlay = function () {
      if (overlay) overlay.hidden = true;
      if (overlayReturnFocus) overlayReturnFocus.focus();
      overlayReturnFocus = null;
    };

    document.addEventListener("keydown", function (event) {
      if (
        (event.ctrlKey || event.metaKey) &&
        (event.key === "k" || event.key === "K")
      ) {
        event.preventDefault();
        if (overlay && !overlay.hidden) closeOverlay();
        else openOverlay("commands");
        return;
      }
      if (event.key === "?" && !event.ctrlKey && !event.metaKey) {
        var target = eventElement(event);
        if (
          hasMethod(target, "closest") &&
          closestHTMLElement(
            target,
            "input, textarea, select, [contenteditable='true']",
          )
        )
          return;
        event.preventDefault();
        openOverlay("help");
      }
    });

    // ----------------------------------------------------------------
    // Composer UX: Enter sends / Shift+Enter breaks a line in the
    // composer textareas, an SMS segment counter that only speaks past
    // one segment, and attachment chips with remove. Everything is
    // delegated at document level, so swapped-in composers pick the
    // behaviors up without re-wiring, and each affordance degrades to
    // the plain form when the pieces are missing.
    // ----------------------------------------------------------------

    // GSM 03.38: basic charset plus the extension table (its characters
    // cost TWO units). Anything outside the sets forces UCS-2, which
    // carries 70 chars per segment (67 once concatenated).
    var GSM7_BASIC =
      "@£$¥èéùìòÇ\nØø\rÅåΔ_ΦΓΛΩΠΨΣΘΞÆæßÉ !\"#¤%&'()*+,-./0123456789:;<=>?" +
      "¡ABCDEFGHIJKLMNOPQRSTUVWXYZÄÖÑÜ§¿abcdefghijklmnopqrstuvwxyzäöñüà";
    var GSM7_EXT = "^{}\\[~]€|\f";
    var GSM7_SET = new Set(GSM7_BASIC);
    var GSM7_EXT_SET = new Set(GSM7_EXT);

    /** @param {string} text @returns {number} */
    function ucs2Segments(text) {
      var chars = Array.from(text).length;
      return chars <= 70 ? 1 : Math.ceil(chars / 67);
    }

    // smsSegments returns the billable segment count for a body. CRLF
    // is normalized first: the composer textarea submits \r\n. Code
    // points (not UTF-16 units) count, so emoji cost one UCS-2 char.
    /** @param {unknown} body @returns {number} */
    function smsSegments(body) {
      var text = String(body == null ? "" : body).replace(/\r\n/g, "\n");
      var units = 0;
      for (var ch of text) {
        if (GSM7_EXT_SET.has(ch)) units += 2;
        else if (!GSM7_SET.has(ch)) return ucs2Segments(text);
        else units += 1;
      }
      return units <= 160 ? 1 : Math.ceil(units / 153);
    }

    // The server's hard cap, mirrored client-side: messaging's
    // MaxBodyLength (1600) counts BYTES (Go len over UTF-8), so the
    // countdown measures encoded bytes, not characters.
    var MAX_BODY_BYTES = 1600;
    // Show the countdown only inside this window of the cap — silent
    // while the body is comfortably small.
    var LIMIT_WINDOW_BYTES = 200;

    /** @param {unknown} text @returns {number} */
    function bodyBytes(text) {
      return new TextEncoder().encode(String(text == null ? "" : text)).length;
    }

    // updateSegcount shows "N SMS" past one segment, and near/over the
    // byte cap appends "bytes/1600" — silent otherwise, in every
    // language (deliberately language-neutral). Over-cap turns the
    // counter loud; the submit stays enabled so the server's own 422
    // banner remains the enforcing verdict (the counter warns, it does
    // not block).
    /** @param {HTMLFormElement} form @param {string} body */
    function updateSegcount(form, body) {
      var counter = /** @type {HTMLElement | null} */ (
        form.querySelector(".wp-segcount")
      );
      if (!counter) return;
      var segments = smsSegments(body);
      var bytes = bodyBytes(body);
      var nearLimit = bytes > MAX_BODY_BYTES - LIMIT_WINDOW_BYTES;
      var over = bytes > MAX_BODY_BYTES;
      if (segments > 1 || nearLimit) {
        var text = segments > 1 ? segments + " SMS" : "";
        if (nearLimit) {
          text += (text ? " · " : "") + bytes + "/" + MAX_BODY_BYTES;
        }
        counter.textContent = text;
        counter.classList.toggle("wp-segcount-over", over);
        counter.hidden = false;
      } else {
        counter.textContent = "";
        counter.classList.remove("wp-segcount-over");
        counter.hidden = true;
      }
    }

    document.addEventListener("keydown", function (event) {
      var area = eventElement(event);
      if (
        !area ||
        !area.closest ||
        !closestHTMLElement(area, "textarea.wp-compose-body")
      )
        return;
      if (event.key !== "Enter" || event.shiftKey || event.isComposing) return;
      var form = closestForm(area);
      if (!form || !form.requestSubmit) return; // degrade: Enter inserts a newline
      event.preventDefault();
      form.requestSubmit();
    });

    document.addEventListener("input", function (event) {
      var area = /** @type {HTMLTextAreaElement} */ (eventElement(event));
      if (
        !area ||
        !area.closest ||
        !closestHTMLElement(area, "textarea.wp-compose-body")
      )
        return;
      var form = closestForm(area);
      if (form) updateSegcount(form, area.value);
    });

    // Attachment chips: names + remove for the composer file inputs.
    // Remove rebuilds the FileList via DataTransfer (the only mutable
    // route); without DataTransfer the native input stays the only UI.
    /** @param {HTMLInputElement} input @param {HTMLElement} box */
    function renderAttachChips(input, box) {
      var files = input.files || [];
      box.replaceChildren();
      for (var i = 0; i < files.length; i++) {
        var chip = document.createElement("span");
        chip.className = "wp-chip";
        chip.textContent = files[i].name;
        var remove = document.createElement("button");
        remove.type = "button";
        remove.className = "wp-chip-x";
        remove.setAttribute("aria-label", "remove");
        remove.dataset.index = String(i);
        chip.append(remove);
        box.append(chip);
      }
      box.hidden = files.length === 0;
    }

    document.addEventListener("change", function (event) {
      var input = /** @type {HTMLInputElement} */ (eventElement(event));
      if (!input || !input.closest || !closestHTMLElement(input, ".wp-compose"))
        return;
      if (String(input.type || "") !== "file") return;
      if (typeof DataTransfer === "undefined") return;
      var form = closestForm(input);
      var box = /** @type {HTMLElement | null} */ (
        form && form.querySelector(".wp-attach")
      );
      if (box) renderAttachChips(input, box);
    });

    document.addEventListener("click", function (event) {
      var remove = closestHTMLElement(eventElement(event), ".wp-chip-x");
      if (!remove) return;
      var form = closestForm(remove);
      var input = /** @type {HTMLInputElement | null} */ (
        form && form.querySelector('input[type="file"]')
      );
      var box = /** @type {HTMLElement | null} */ (
        form && form.querySelector(".wp-attach")
      );
      if (!input || !box || typeof DataTransfer === "undefined") return;
      var index = Number(remove.dataset.index);
      var transfer = new DataTransfer();
      var files = input.files || [];
      for (var i = 0; i < files.length; i++) {
        if (i !== index) transfer.items.add(files[i]);
      }
      input.files = transfer.files;
      renderAttachChips(input, box);
    });
    // Test seam only: on the wire this file is a classic script where
    // `module` does not exist, so the guard is a no-op in the browser. It
    // sits INSIDE the try on purpose: the five seams exist only once the
    // wiring above completed, and a load that threw must publish nothing.
    if (typeof module !== "undefined" && module.exports) {
      module.exports.restoreLastTab = restoreLastTab;
      module.exports.refreshNav = refreshNav;
      module.exports.vmPeaks = vmPeaks;
      module.exports.vmClock = vmClock;
      module.exports.agentModeOpened = agentModeOpened;
      module.exports.errorMessage = errorMessage;
      module.exports.clock = clock;
    }
  } catch (err) {
    var logList = document.getElementById("log");
    if (logList) {
      var entry = document.createElement("li");
      entry.textContent = "shell load failed: " + errorMessage(err);
      entry.dataset.level = "error";
      logList.prepend(entry);
    }
    if (window.console && console.error)
      console.error("webphone: shell load failed", err);
  }
})();
