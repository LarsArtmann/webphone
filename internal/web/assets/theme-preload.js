// Pre-paint shell state: applies the stored theme AND the collapsed phone
// island to <html> BEFORE first paint, so a forced layout never flashes the
// default one. Same-origin file — the strict CSP allows no inline scripts at
// all. shell.js §4 owns the theme toggle cycle and re-applies after load;
// "auto" needs nothing here (the stylesheets' prefers-color-scheme blocks
// apply while no data-theme attribute is set).
(function () {
  var theme = null;
  var welcomeDismissed = false;
  var islandCollapsed = false;
  try {
    theme = localStorage.getItem("wp-theme");
    welcomeDismissed = localStorage.getItem("wp-welcome-dismissed") === "1";
    islandCollapsed = localStorage.getItem("wp-island-collapsed") === "1";
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
  // The collapsible island (P2): same reasoning one layout level up —
  // collapsing changes the grid's column count, so a late class would
  // flash the wide grid and then snap. shell.js §4b owns the toggle and
  // keeps this class in step; the class always lives on <html>.
  if (islandCollapsed) {
    document.documentElement.classList.add("wp-island-collapsed");
  }
})();
