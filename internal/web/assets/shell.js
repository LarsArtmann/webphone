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
      });
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
      fetch("/messages/" + transcript.dataset.thread + "/read", {
        method: "POST",
        headers: { "X-CSRF-Token": csrfToken() },
        credentials: "same-origin",
      })
        .then(refreshNav)
        .catch(function () {});
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
      if (button) button.textContent = "Theme: " + theme;
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

    // updateSegcount shows "N SMS" only past one segment — silent
    // otherwise, in every language (deliberately language-neutral).
    function updateSegcount(form, body) {
      var counter = form.querySelector(".wp-segcount");
      if (!counter) return;
      var segments = smsSegments(body);
      if (segments > 1) {
        counter.textContent = segments + " SMS";
        counter.hidden = false;
      } else {
        counter.textContent = "";
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
})();
