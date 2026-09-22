// Theme preload: applies the stored theme to <html> BEFORE first paint,
// so a forced light/dark choice never flashes the OS theme. Same-origin
// file — the strict CSP allows no inline scripts at all. shell.js §4 owns
// the toggle cycle and re-applies after load; "auto" needs nothing here
// (the stylesheets' prefers-color-scheme blocks apply while no data-theme
// attribute is set).
(function () {
  var theme = null;
  try {
    theme = localStorage.getItem("wp-theme");
  } catch (err) {
    return; // storage unavailable (private mode) — ride the OS theme
  }
  if (theme === "light" || theme === "dark") {
    document.documentElement.setAttribute("data-theme", theme);
  }
})();
