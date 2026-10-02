// Shell glue: small behaviors the server-rendered tabs need.
// CSP-strict: a plain external script, no inline handlers.

(function () {
  "use strict";
  // Load-error boundary (plan T19): if ANY shell wiring throws at load,
  // leave a breadcrumb in the island's event log (it may still exist
  // even when the island itself failed) and rethrow to the console —
  // never a silent half-wired shell.
  try {
    // 1. Nav active state across HTMX partial swaps: the server marks the
    //    active link on full renders; after a partial swap only the clicked
    //    link knows. htmx:afterRequest fires on the link that issued it.
    document.addEventListener("htmx:afterRequest", function (event) {
      var element = event.target;
      if (!element || !element.hasAttribute("data-tab")) return;
      var nav = element.closest(".wp-nav");
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
        } catch (err) {}
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
      } catch (err) {
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
      var button = event.target.closest(".wp-welcome-dismiss");
      if (!button) return;
      try {
        localStorage.setItem("wp-welcome-dismissed", "1");
      } catch (err) {}
      document.documentElement.classList.add("wp-welcome-dismissed");
    });

    // 2. data-dial buttons (contacts, history, voicemail, threads): push
    //    the number into the island's dial form and submit it — the same
    //    gesture as the island's redial. When the island is signed out
    //    (#phone-view hidden) a submit would dead-end inside the hidden
    //    form, so we say so and point at the login field instead. The
    //    toast mirrors the island's announce() markup (shared CSS, no
    //    island import — shell and island never load each other's code).
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
      while (host.children.length > 4) host.firstChild.remove();
      setTimeout(dismiss, 6000);
    };
    document.addEventListener("click", function (event) {
      var button = event.target.closest("[data-dial]");
      if (!button) return;
      var dest = document.getElementById("dest");
      var form = document.getElementById("dial-form");
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
      dest.value = button.getAttribute("data-dial");
      form.requestSubmit();
    });

    // 2b2. data-sms (history/voicemail rows → the Messages composer
    //      prefilled with the number) and data-save-contact (☆ → the
    //      island's contact save via the wp:save-contact event; the
    //      island owns /api/contacts writes). The prefill rides a
    //      one-shot afterSwap: the composer only exists once the tab
    //      partial has swapped in.
    document.addEventListener("click", function (event) {
      var sms = event.target.closest("[data-sms]");
      if (sms) {
        var number = sms.getAttribute("data-sms");
        var nav = document.querySelector("[data-tab='messages']");
        if (nav) nav.click();
        var prefill = function () {
          document.removeEventListener("htmx:afterSwap", prefill);
          var to = document.querySelector(
            "form.wp-compose-new input[name='to']",
          );
          if (to) {
            to.value = number;
            to.focus();
          }
        };
        document.addEventListener("htmx:afterSwap", prefill);
        return;
      }
      var save = event.target.closest("[data-save-contact]");
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

    // 2c. Live-call presence: the island dispatches wp:calls-changed after
    //     every call render; the shell mirrors the live call count into
    //     the header so every tab shows the phone is busy. Cards are
    //     removed on teardown (calls.js), so counting them counts live
    //     calls (ringing included). English by the shell.js precedent
    //     (theme toggle) — presence is operator glanceable state.
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
      var link =
        event.target && event.target.closest
          ? event.target.closest("[data-tab='history']")
          : null;
      if (!link) return;
      missedCount = 0;
      renderMissedBadge();
    });

    // 2b. data-reload buttons (error panel): full reload, same as the old
    //      inline onclick but CSP-safe via this delegated listener.
    document.addEventListener("click", function (event) {
      if (!event.target.closest("[data-reload]")) return;
      location.reload();
    });

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
    var csrfToken = function () {
      var meta = document.querySelector('meta[name="csrf-token"]');
      return meta ? meta.getAttribute("content") : "";
    };
    var refreshNav = function () {
      if (!window.htmx) return;
      var active = document.querySelector("#wp-nav .wp-nav-link.wp-active");
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
      var transcript = event.target;
      if (!transcript || transcript.id !== "thread-transcript") return;
      if (transcript.dataset.page !== "0") event.preventDefault();
    });
    document.addEventListener("htmx:sseMessage", function (event) {
      var transcript = event.target;
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
    var jumpChip = null;
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
        if (!transcript || event.target !== transcript) return;
        if (nearBottom(transcript)) resetJumpChip();
      },
      true,
    );
    document.addEventListener("htmx:sseBeforeMessage", function (event) {
      var transcript = event.target;
      if (!transcript || transcript.id !== "thread-transcript") return;
      transcript.dataset.wasNearBottom = nearBottom(transcript) ? "1" : "0";
    });
    document.addEventListener("htmx:sseMessage", function (event) {
      var transcript = event.target;
      if (!transcript || transcript.id !== "thread-transcript") return;
      if (transcript.dataset.page !== "0" || !transcript.dataset.thread) return;
      if (transcript.dataset.wasNearBottom === "0") {
        pendingNew += 1;
        var wrap = transcript.parentElement;
        var chip = wrap && wrap.querySelector(".wp-jump-latest");
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
      var chip =
        event.target && event.target.closest
          ? event.target.closest(".wp-jump-latest")
          : null;
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
      var list = event.target;
      if (
        !list ||
        !list.classList ||
        !list.classList.contains("wp-thread-list")
      )
        return;
      var input = document.getElementById("wp-thread-search-input");
      if (input && input.value.trim() !== "") event.preventDefault();
    });
    // The island's language switch re-labels itself and re-fetches the open
    // tab; the nav is shell territory, so it asks via this event.
    document.addEventListener("wp:lang-changed", refreshNav);

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
    // Injectable clock (plan T19): node:test passes a fake via
    // window.__wpClock.now; the browser default is Date.now. Resolved at
    // CALL time so a test can install the fake after this script loads.
    var now = function () {
      if (window.__wpClock && typeof window.__wpClock.now === "function") {
        return window.__wpClock.now();
      }
      return Date.now();
    };
    var showThrottledError = function (message) {
      var nowMs = now();
      if (nowMs - lastErrorToast < 8000) return;
      lastErrorToast = nowMs;
      shellToast(message, "error");
    };
    document.addEventListener("htmx:responseError", function (event) {
      var xhr = event.detail && event.detail.xhr;
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
    var draftKey = function (threadId) {
      return "wp-draft:" + threadId;
    };
    var currentDraftId = function () {
      var transcript = document.getElementById("thread-transcript");
      return transcript && transcript.dataset.thread
        ? transcript.dataset.thread
        : "new";
    };
    var saveDraft = function (draftId, text) {
      try {
        if (text && text.length <= DRAFT_MAX) {
          localStorage.setItem(draftKey(draftId), text);
        } else {
          localStorage.removeItem(draftKey(draftId));
        }
      } catch (err) {
        /* storage unavailable — drafts are best-effort */
      }
    };
    var draftTimer = null;
    document.addEventListener("input", function (event) {
      var area = event.target;
      if (!area.closest || !area.closest("textarea.wp-compose-body")) return;
      if (draftTimer) clearTimeout(draftTimer);
      draftTimer = setTimeout(function () {
        saveDraft(currentDraftId(), area.value);
      }, 300);
    });
    document.addEventListener("htmx:afterSwap", function () {
      var area = document.querySelector("textarea.wp-compose-body");
      if (!area || area.value !== "") return;
      var draft = null;
      try {
        draft = localStorage.getItem(draftKey(currentDraftId()));
      } catch (err) {
        return;
      }
      if (draft) area.value = draft;
    });
    document.addEventListener("htmx:afterRequest", function (event) {
      var form = event.target;
      if (!form || !form.matches || !form.matches("form")) return;
      if (!event.detail || !event.detail.successful) return;
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
      var form = (event.detail && event.detail.elt) || event.target;
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
        chip.textContent = "📎 " + name;
        bubble.append(chip);
      });
      var meta = document.createElement("span");
      meta.className = "wp-bubble-meta";
      var clock = document.createElement("span");
      var stamp = new Date();
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
    var rollbackOptimistic = function (event) {
      var form = (event.detail && event.detail.elt) || event.target;
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
      var area = form.querySelector("textarea.wp-compose-body");
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
        var current = document.querySelector(
          "form.wp-compose:not(.wp-compose-new)",
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
      var form = (event.detail && event.detail.elt) || event.target;
      if (!form || !form.matches || !form.matches("form.wp-compose")) return;
      if (event.detail && event.detail.successful) optPending.delete(form);
    });

    // 3f. Focus lands on the new panel's heading after a navigation
    //     swap (tab click, thread open, back link): keyboard and
    //     screen-reader users start reading at the top of what just
    //     arrived instead of wherever focus happened to be. Typing-
    //     driven swaps (search, composer, drafts) deliberately keep
    //     focus where the user is.
    document.addEventListener("htmx:afterSwap", function (event) {
      var elt = event.target;
      if (!elt || !elt.matches) return;
      var navigating =
        elt.hasAttribute && elt.hasAttribute("data-tab")
          ? true
          : Boolean(elt.closest && elt.closest(".wp-thread-rowwrap, .wp-back"));
      if (!navigating) return;
      var heading = document.querySelector("#tab-content h2");
      if (!heading || typeof heading.focus !== "function") return;
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
        var seconds =
          Math.floor(Date.now() / 1000) - Number(nodes[i].dataset.when || 0);
        if (!(seconds >= 0)) continue;
        if (seconds < 60) nodes[i].textContent = "now";
        else if (seconds < 3600)
          nodes[i].textContent = Math.floor(seconds / 60) + "m";
        else if (seconds < 86400)
          nodes[i].textContent = Math.floor(seconds / 3600) + "h";
      }
    };
    setInterval(tickRelative, 30000);

    // 3h. Tab skeleton (F1): navigation swaps (tab links, thread rows,
    //     back) reveal the shimmer placeholder while the partial is in
    //     flight. Typing-driven fetches (search, composer) stay quiet —
    //     morph keeps those surfaces alive and a flash there would be
    //     noise, not signal.
    var skeleton = null;
    var navigatingSwap = function (elt) {
      if (!elt || !elt.matches) return false;
      return Boolean(
        (elt.hasAttribute && elt.hasAttribute("data-tab")) ||
        (elt.closest && elt.closest(".wp-thread-rowwrap, .wp-back")),
      );
    };
    document.addEventListener("htmx:beforeRequest", function (event) {
      if (!navigatingSwap(event.target)) return;
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
    var vmPeaksCache = new Map(); // uuid -> peak array (48 bars)
    var vmWired = new WeakSet(); // audio elements with listeners attached
    var vmActiveAudio = null; // single-active-player invariant
    var vmAudioContext = null; // browsers cap contexts — one per page, lazy
    var vmFallbackToasted = false;

    var VM_BARS = 48;
    var VM_SPEED_LADDER = [1, 1.5, 2, 0.5];

    // vmPeaks reduces decoded PCM (−1..1) to one normalized max-abs peak
    // per bar. Pure and unit-tested: empty input renders silence, and a
    // bar whose sample slice is empty (more bars than samples) inherits
    // the nearest valued bar so the wave starts where the audio does.
    var vmPeaks = function (samples, bars) {
      var peaks = new Array(bars).fill(0);
      if (!samples || samples.length === 0) return peaks;
      var step = samples.length / bars;
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
    var vmClock = function (seconds) {
      if (!isFinite(seconds) || seconds < 0) seconds = 0;
      var s = Math.floor(seconds);
      var mm = Math.floor(s / 60);
      var ss = s % 60;
      return mm + ":" + (ss < 10 ? "0" : "") + ss;
    };

    var vmRowOf = function (el) {
      return el && el.closest ? el.closest(".wp-vm-row") : null;
    };

    var vmToken = function (canvas, name, fallback) {
      if (typeof getComputedStyle !== "function") return fallback;
      var value = getComputedStyle(canvas).getPropertyValue(name);
      return value && value.trim() ? value.trim() : fallback;
    };

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

    var vmSetPlaying = function (uuid, audio, playing) {
      var row = vmRowOf(audio);
      if (row && row.classList) {
        if (playing) row.classList.add("wp-playing");
        else row.classList.remove("wp-playing");
      }
      var btn = document.getElementById("vm-play-" + uuid);
      if (btn) {
        btn.textContent = playing ? "⏸" : "▶";
        var label = btn.getAttribute(
          playing ? "data-label-pause" : "data-label-play",
        );
        if (label) btn.setAttribute("aria-label", label);
      }
      if (playing) vmClearUnread(row);
    };

    var vmPaint = function (uuid, audio) {
      var duration = isFinite(audio.duration) ? audio.duration : 0;
      var current = isFinite(audio.currentTime) ? audio.currentTime : 0;
      var progress = duration > 0 ? current / duration : 0;
      var time = document.getElementById("vm-time-" + uuid);
      if (time) {
        time.textContent = vmClock(current) + " / " + vmClock(duration);
      }
      var canvas = document.getElementById("vm-wave-" + uuid);
      if (canvas) {
        canvas.setAttribute("aria-valuenow", String(Math.floor(current)));
        var peaks = vmPeaksCache.get(uuid);
        if (peaks) vmDraw(canvas, peaks, progress);
      }
    };

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

    var vmBuildWave = function (uuid, audio) {
      if (vmPeaksCache.has(uuid)) {
        return Promise.resolve(vmPeaksCache.get(uuid));
      }
      var src = audio.getAttribute ? audio.getAttribute("src") : null;
      if (
        typeof fetch !== "function" ||
        !src ||
        !(window.AudioContext || window.webkitAudioContext)
      ) {
        return Promise.reject(new Error("webaudio unavailable"));
      }
      if (!vmAudioContext) {
        var Ctor = window.AudioContext || window.webkitAudioContext;
        vmAudioContext = new Ctor();
      }
      return fetch(src)
        .then(function (response) {
          if (!response.ok)
            throw new Error("audio fetch HTTP " + response.status);
          return response.arrayBuffer();
        })
        .then(function (buffer) {
          return new Promise(function (resolve, reject) {
            // callback form: the oldest widest-compatible signature
            vmAudioContext.decodeAudioData(buffer, resolve, reject);
          });
        })
        .then(function (decoded) {
          var peaks = vmPeaks(decoded.getChannelData(0), VM_BARS);
          vmPeaksCache.set(uuid, peaks);
          return peaks;
        });
    };

    var vmTogglePlay = function (btn) {
      var uuid = btn.getAttribute("data-vm-play");
      var audio = document.getElementById("vm-audio-" + uuid);
      if (!audio || typeof audio.play !== "function") return;
      vmWire(audio, uuid);
      if (audio.paused === false) {
        audio.pause();
        return;
      }
      // single active player: starting one pauses the previous row
      if (
        vmActiveAudio &&
        vmActiveAudio !== audio &&
        typeof vmActiveAudio.pause === "function"
      ) {
        vmActiveAudio.pause();
      }
      vmActiveAudio = audio;
      var started = audio.play();
      if (started && typeof started.catch === "function") {
        started.catch(function () {});
      }
      vmBuildWave(uuid, audio)
        .then(function (peaks) {
          vmPaint(uuid, audio);
          vmDraw(document.getElementById("vm-wave-" + uuid), peaks, 0);
        })
        .catch(function () {
          vmFallback(uuid, audio);
        });
    };

    var vmCycleSpeed = function (btn) {
      var uuid = btn.getAttribute("data-vm-speed");
      var audio = document.getElementById("vm-audio-" + uuid);
      if (!audio) return;
      var current = Number(audio.playbackRate) || 1;
      var idx = VM_SPEED_LADDER.indexOf(current);
      var next =
        idx >= 0 ? VM_SPEED_LADDER[(idx + 1) % VM_SPEED_LADDER.length] : 1;
      audio.playbackRate = next;
      btn.textContent = next + "×";
    };

    document.addEventListener("click", function (event) {
      if (!event.target || !event.target.closest) return;
      var play = event.target.closest(".wp-vm-play");
      if (play) {
        vmTogglePlay(play);
        return;
      }
      var speed = event.target.closest(".wp-vm-speed");
      if (speed) vmCycleSpeed(speed);
    });

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
      if (!event.target || !event.target.closest) return;
      var canvas = event.target.closest(".wp-vm-wave");
      if (!canvas) return;
      var uuid = canvas.id.slice("vm-wave-".length);
      var audio = document.getElementById("vm-audio-" + uuid);
      if (!audio) return;
      vmWire(audio, uuid);
      vmSeek(canvas, audio, event.clientX);
      // drag-to-scrub continues while the pointer is down
      var move = function (moveEvent) {
        vmSeek(canvas, audio, moveEvent.clientX);
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
      if (!event.target || !event.target.closest) return;
      var canvas = event.target.closest(".wp-vm-wave");
      if (!canvas) return;
      if (event.key !== "ArrowLeft" && event.key !== "ArrowRight") return;
      var uuid = canvas.id.slice("vm-wave-".length);
      var audio = document.getElementById("vm-audio-" + uuid);
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
    } catch (err) {
      /* storage unavailable (private mode) — fall through to auto */
    }
    var themeIndex = THEMES.indexOf(storedTheme);
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
        } catch (err) {
          /* best-effort persistence only */
        }
        applyTheme();
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
    var overlay = null;
    var overlayInput = null;
    var overlayList = null;
    var overlayMode = "commands";
    var overlayItems = [];
    var overlayIndex = 0;
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
        })(links[i]);
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
          var messages = document.querySelector("[data-tab='messages']");
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
      overlayList.replaceChildren();
      overlayItems = [];
      overlayIndex = 0;
      var query = overlayInput.value.trim().toLowerCase();
      var source =
        overlayMode === "commands"
          ? collectCommands()
          : HELP_ITEMS.map(function (pair) {
              return { label: pair[0] + " — " + pair[1] };
            });
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
        overlayList.append(li);
      }
    };

    var selectOverlay = function (delta) {
      if (overlayItems.length === 0) return;
      overlayIndex =
        (overlayIndex + delta + overlayItems.length) % overlayItems.length;
      var rows = overlayList.children;
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
      overlay = document.createElement("div");
      overlay.id = "wp-palette";
      overlay.className = "wp-overlay";
      overlay.hidden = true;
      overlay.setAttribute("role", "dialog");
      overlay.setAttribute("aria-label", "Command palette");
      overlayInput = document.createElement("input");
      overlayInput.type = "text";
      overlayInput.className = "wp-overlay-input";
      overlayInput.setAttribute("aria-label", "Search commands");
      overlayInput.placeholder = "Type a command…";
      overlayList = document.createElement("ul");
      overlayList.className = "wp-overlay-list";
      overlay.append(overlayInput, overlayList);
      document.body.append(overlay);
      overlayInput.addEventListener("input", renderOverlay);
      overlayInput.addEventListener("keydown", function (event) {
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
      overlay.addEventListener("click", function (event) {
        if (event.target === overlay) closeOverlay();
      });
    };

    var openOverlay = function (mode) {
      if (!overlay) buildOverlay();
      overlayMode = mode;
      overlayReturnFocus = document.activeElement;
      overlay.hidden = false;
      overlayInput.value = "";
      overlayInput.placeholder =
        mode === "commands" ? "Type a command…" : "Shortcuts — Esc closes";
      renderOverlay();
      overlayInput.focus();
    };

    var closeOverlay = function () {
      if (overlay) overlay.hidden = true;
      if (overlayReturnFocus && overlayReturnFocus.focus)
        overlayReturnFocus.focus();
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
        var target = event.target;
        if (
          target &&
          target.closest &&
          target.closest("input, textarea, select, [contenteditable='true']")
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

    function ucs2Segments(text) {
      var chars = Array.from(text).length;
      return chars <= 70 ? 1 : Math.ceil(chars / 67);
    }

    // smsSegments returns the billable segment count for a body. CRLF
    // is normalized first: the composer textarea submits \r\n. Code
    // points (not UTF-16 units) count, so emoji cost one UCS-2 char.
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

    function bodyBytes(text) {
      return new TextEncoder().encode(String(text == null ? "" : text)).length;
    }

    // updateSegcount shows "N SMS" past one segment, and near/over the
    // byte cap appends "bytes/1600" — silent otherwise, in every
    // language (deliberately language-neutral). Over-cap turns the
    // counter loud; the submit stays enabled so the server's own 422
    // banner remains the enforcing verdict (the counter warns, it does
    // not block).
    function updateSegcount(form, body) {
      var counter = form.querySelector(".wp-segcount");
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
      var area = event.target;
      if (!area || !area.closest || !area.closest("textarea.wp-compose-body"))
        return;
      if (event.key !== "Enter" || event.shiftKey || event.isComposing) return;
      var form = area.closest("form");
      if (!form || !form.requestSubmit) return; // degrade: Enter inserts a newline
      event.preventDefault();
      form.requestSubmit();
    });

    document.addEventListener("input", function (event) {
      var area = event.target;
      if (!area || !area.closest || !area.closest("textarea.wp-compose-body"))
        return;
      var form = area.closest("form");
      if (form) updateSegcount(form, area.value);
    });

    // Attachment chips: names + remove for the composer file inputs.
    // Remove rebuilds the FileList via DataTransfer (the only mutable
    // route); without DataTransfer the native input stays the only UI.
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
      var input = event.target;
      if (!input || !input.closest || !input.closest(".wp-compose")) return;
      if (String(input.type || "") !== "file") return;
      if (typeof DataTransfer === "undefined") return;
      var form = input.closest("form");
      var box = form && form.querySelector(".wp-attach");
      if (box) renderAttachChips(input, box);
    });

    document.addEventListener("click", function (event) {
      var remove =
        event.target && event.target.closest
          ? event.target.closest(".wp-chip-x")
          : null;
      if (!remove) return;
      var form = remove.closest("form");
      var input = form && form.querySelector('input[type="file"]');
      var box = form && form.querySelector(".wp-attach");
      if (!input || !box || typeof DataTransfer === "undefined") return;
      var index = Number(remove.dataset.index);
      var transfer = new DataTransfer();
      for (var i = 0; i < input.files.length; i++) {
        if (i !== index) transfer.items.add(input.files[i]);
      }
      input.files = transfer.files;
      renderAttachChips(input, box);
    });
  } catch (err) {
    var logList = document.getElementById("log");
    if (logList) {
      var entry = document.createElement("li");
      entry.textContent = "shell load failed: " + (err && err.message);
      entry.dataset.level = "error";
      logList.prepend(entry);
    }
    if (window.console && console.error)
      console.error("webphone: shell load failed", err);
  }

  // Test seam only: on the wire this file is a classic script where
  // `module` does not exist, so the guard is a no-op in the browser.
  if (typeof module !== "undefined" && module.exports) {
    module.exports.restoreLastTab = restoreLastTab;
    module.exports.refreshNav = refreshNav;
    module.exports.vmPeaks = vmPeaks;
    module.exports.vmClock = vmClock;
  }
})();
