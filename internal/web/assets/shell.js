// Shell glue: small behaviors the server-rendered tabs need.
// CSP-strict: a plain external script, no inline handlers.

(function () {
  "use strict";

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

  // 2. data-dial buttons (contacts): push the number into the island's
  //    dial form and submit it — the same gesture as the island's redial.
  document.addEventListener("click", function (event) {
    var button = event.target.closest("[data-dial]");
    if (!button) return;
    var dest = document.getElementById("dest");
    var form = document.getElementById("dial-form");
    if (!dest || !form) return;
    dest.value = button.getAttribute("data-dial");
    form.requestSubmit();
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
    window.htmx.ajax("GET", source, { target: "#wp-nav", swap: "innerHTML" });
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
})();
