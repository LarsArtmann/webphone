// Tailwind-coexistence probe (throwaway, plan M7): after DOM ready,
// compute the styles the A/B verdict needs and write them as JSON into
// #probe-out so a headless `chromium --dump-dom` capture can diff the
// tw=1 vs tw=0 states without any browser automation tooling.
(function () {
  "use strict";

  var PROPS = [
    "display",
    "position",
    "margin",
    "padding",
    "font-size",
    "line-height",
    "color",
    "background-color",
    "border-radius",
    "border-top-width",
    "list-style-type",
    "text-decoration",
    "width",
    "height",
  ];

  function probe(el) {
    if (!el) return null;
    var cs = getComputedStyle(el);
    var out = {};
    PROPS.forEach(function (prop) {
      out[prop] = cs.getPropertyValue(prop);
    });
    return out;
  }

  function firstChild(id) {
    var host = document.getElementById(id);
    return host ? host.firstElementChild : null;
  }

  function badgeSpan() {
    var host = document.getElementById("lib-badge");
    if (!host) return null;
    return host.querySelector("span.absolute") || null;
  }

  function collect() {
    return {
      webphone: {
        panel: probe(document.getElementById("wp-panel")),
        panelHeadH2: probe(document.querySelector("#wp-panel-head h2")),
        listItem: probe(document.querySelector("#wp-list li")),
        error: probe(document.querySelector(".wp-error")),
        panelSub: probe(document.querySelector(".wp-panel-sub")),
      },
      library: {
        emptyState: probe(firstChild("lib-empty")),
        emptyStateTitle: probe(document.querySelector("#lib-empty h3, #lib-empty h2")),
        emptyStateIcon: probe(document.querySelector("#lib-empty svg")),
        relativeTime: probe(document.querySelector("#lib-time time")),
        badgeHost: probe(firstChild("lib-badge")),
        badgePill: probe(badgeSpan()),
      },
    };
  }

  function ready() {
    var out = document.getElementById("probe-out");
    if (!out) return;
    out.textContent = "SPIKE_PROBE_JSON:" + JSON.stringify(collect(), null, 0);
  }

  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", ready);
  } else {
    ready();
  }
})();
