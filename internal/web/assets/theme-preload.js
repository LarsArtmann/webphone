// Theme preload: applies the stored theme to <html> BEFORE first paint,
// so a forced light/dark choice never flashes the OS theme. Same-origin
// file — the strict CSP allows no inline scripts at all. shell.js §4 owns
// the toggle cycle and re-applies after load; "auto" needs nothing here
// (the stylesheets' prefers-color-scheme blocks apply while no data-theme
// attribute is set).
(function () {
  var theme = null;
  var welcomeDismissed = false;
  try {
    theme = localStorage.getItem("wp-theme");
    welcomeDismissed = localStorage.getItem("wp-welcome-dismissed") === "1";
  } catch {
    return; // storage unavailable (private mode) — ride the OS theme
  }
  if (theme === "light" || theme === "dark") {
    document.documentElement.setAttribute("data-theme", theme);
  }
  // M18: a dismissed welcome intro collapses before first paint. The
  // class lives on the root ("wp-welcome-dismissed" in app.css) so
  // this head script can set it before the panel's DOM exists.
  if (welcomeDismissed) {
    document.documentElement.classList.add("wp-welcome-dismissed");
  }
})();
