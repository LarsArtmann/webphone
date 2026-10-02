// shell.js under node:test — the error surfacing the Go asset tripwires
// can only grep for. htmx swaps NOTHING on error responses, so before the
// 3c handler a dead tab session made every tab click and form submit fail
// silently. These specs drive the real document-level listeners the way
// htmx fires them: responseError carries detail.xhr, sendError carries
// none.
import test from "node:test";
import assert from "node:assert/strict";

import { installBrowserGlobals } from "./helpers.mjs";

const doc = installBrowserGlobals();
await import("../shell.js");
// shell.js is CommonJS under node (query-string re-imports return the
// cached module, never a fresh evaluation), so the specs below drive
// the exported seam directly.
const shellApi = (await import("../shell.js")).default;

const toasts = () => doc.getElementById("toasts");
const fakeXhr = (status, headers = {}) => ({
  status,
  getResponseHeader: (name) => headers[name] ?? null,
});

test("a 401 from a dead tab session toasts what happened, without reloading", () => {
  doc.dispatch("htmx:responseError", { detail: { xhr: fakeXhr(401) } });
  const toast = toasts().children.at(-1);
  assert.equal(toast.className, "toast toast-error");
  assert.match(toast.textContent, /session ended/i);
  assert.match(toast.textContent, /calls keep working/i);
  assert.match(toast.textContent, /reload/i);
});

test("server-authored feedback (HX-Trigger) is never double-toasted", () => {
  const before = toasts().children.length;
  doc.dispatch("htmx:responseError", {
    detail: { xhr: fakeXhr(422, { "HX-Trigger": '{"showMessage":{}}' }) },
  });
  assert.equal(toasts().children.length, before);
});

test("the throttle collapses an error storm into one toast", () => {
  const before = toasts().children.length;
  for (let i = 0; i < 5; i++) {
    doc.dispatch("htmx:responseError", { detail: { xhr: fakeXhr(401) } });
  }
  doc.dispatch("htmx:responseError", { detail: { xhr: fakeXhr(502) } });
  doc.dispatch("htmx:sendError", {});
  assert.equal(toasts().children.length, before);
});

test("after the throttle window the next failure toasts again", () => {
  const realNow = Date.now;
  const base = realNow();
  try {
    Date.now = () => base + 10_000;
    doc.dispatch("htmx:sendError", {});
    const toast = toasts().children.at(-1);
    assert.equal(toast.className, "toast toast-error");
    assert.match(toast.textContent, /network request failed/i);
  } finally {
    Date.now = realNow;
  }
});

test("429 toasts the client-correctable feedback (slow down)", () => {
  const realNow = Date.now;
  const base = realNow();
  try {
    Date.now = () => base + 30_000;
    doc.dispatch("htmx:responseError", { detail: { xhr: fakeXhr(429) } });
    const toast = toasts().children.at(-1);
    assert.match(toast.textContent, /too many requests/i);
    assert.match(toast.textContent, /wait a moment/i);
  } finally {
    Date.now = realNow;
  }
});

test("statuses without server feedback get an honest generic toast", () => {
  const realNow = Date.now;
  const base = realNow();
  try {
    Date.now = () => base + 60_000;
    doc.dispatch("htmx:responseError", { detail: { xhr: fakeXhr(500) } });
    assert.match(toasts().children.at(-1).textContent, /HTTP 500/);
  } finally {
    Date.now = realNow;
  }
});

// 3d. Per-thread draft persistence (plan T21d): composer text survives
// the re-renders that empty it (tab/thread switches), restores only
// into an empty composer, and clears after a successful send.
test("composer drafts persist per thread and restore after re-render", async () => {
  const transcript = doc.getElementById("thread-transcript");
  transcript.dataset.thread = "t-42";
  const composer = {
    value: "",
    closest(selector) {
      return selector === "textarea.wp-compose-body" ? this : null;
    },
    querySelector(selector) {
      return selector === "textarea.wp-compose-body" ? this : null;
    },
  };
  const realQuerySelector = doc.querySelector.bind(doc);
  doc.querySelector = (selector) =>
    selector === "textarea.wp-compose-body" ? composer : realQuerySelector(selector);

  composer.value = "hallo draft";
  doc.dispatch("input", { target: composer });
  await new Promise((resolve) => setTimeout(resolve, 400));
  assert.equal(localStorage.getItem("wp-draft:t-42"), "hallo draft");

  // A swap re-renders the composer empty: the draft comes back.
  composer.value = "";
  doc.dispatch("htmx:afterSwap", {});
  assert.equal(composer.value, "hallo draft");

  // A successful send clears the draft — and an empty swap stays empty.
  const sendForm = {
    matches: (selector) => selector === "form",
    hasAttribute: () => false,
    querySelector: (selector) => (selector === "textarea.wp-compose-body" ? composer : null),
  };
  composer.value = "hallo draft";
  doc.dispatch("htmx:afterRequest", { target: sendForm, detail: { successful: true } });
  assert.equal(localStorage.getItem("wp-draft:t-42"), null);
  composer.value = "";
  doc.dispatch("htmx:afterSwap", {});
  assert.equal(composer.value, "");

  // A FAILED send keeps the draft for the retry.
  composer.value = "still typing";
  doc.dispatch("input", { target: composer });
  await new Promise((resolve) => setTimeout(resolve, 400));
  doc.dispatch("htmx:afterRequest", { target: sendForm, detail: { successful: false } });
  assert.equal(localStorage.getItem("wp-draft:t-42"), "still typing");

  doc.querySelector = realQuerySelector;
  delete transcript.dataset.thread;
});

// 2b2. data-sms opens the Messages tab and prefills the composer's
// recipient once the partial swapped in; data-save-contact only
// bridges the gesture to the island's wp:save-contact event.
test("data-sms prefills the new-message composer after the tab swap", () => {
  const nav = doc.getElementById("nav-messages");
  nav.dataset.tab = "messages";
  let navClicked = false;
  nav.click = () => {
    navClicked = true;
  };
  const composer = {
    value: "",
    focus() {
      composer.focused = true;
    },
    focused: false,
  };
  const realQuerySelector = doc.querySelector.bind(doc);
  doc.querySelector = (selector) =>
    selector === "form.wp-compose-new input[name='to']"
      ? composer
      : selector === "[data-tab='messages']"
        ? nav
        : realQuerySelector(selector);

  doc.dispatch("click", {
    target: { closest: (sel) => (sel === "[data-sms]" ? { getAttribute: () => "+4930" } : null) },
  });
  assert.ok(navClicked, "the messages nav was clicked");

  doc.dispatch("htmx:afterSwap", {});
  assert.equal(composer.value, "+4930");
  assert.ok(composer.focused, "recipient field focused");

  // The prefill listener is one-shot: later swaps do not re-focus.
  composer.focused = false;
  doc.dispatch("htmx:afterSwap", {});
  assert.ok(!composer.focused, "prefill is one-shot");

  doc.querySelector = realQuerySelector;
  delete nav.dataset.tab;
});

test("data-save-contact bridges to the island's wp:save-contact event", () => {
  let seen = null;
  doc.addEventListener("wp:save-contact", (event) => {
    seen = event.detail;
  });
  doc.dispatch("click", {
    target: {
      closest: (sel) =>
        sel === "[data-save-contact]"
          ? { getAttribute: (name) => (name === "data-save-contact" ? "+441632960961" : null) }
          : null,
    },
  });
  assert.equal(seen.number, "+441632960961");
});

// 3b-2. Jump-to-latest: a live push landing while the reader is scrolled
// up must not yank them down NOR go unseen — it counts into the chip.
// Near-bottom pushes keep the pinned-scroll behavior instead.
test("scrolled-away live pushes count into the jump chip; near-bottom pushes pin", async () => {
  const transcript = doc.getElementById("thread-transcript");
  transcript.dataset.page = "0";
  transcript.dataset.thread = "t-7";
  transcript.scrollHeight = 1000;
  transcript.clientHeight = 500;
  transcript.scrollTop = 0; // scrolled 500px away from the bottom

  const wrap = doc.createElement();
  const chip = doc.createElement();
  chip.className = "wp-jump-latest";
  chip.hidden = true;
  wrap.append(transcript, chip);

  doc.dispatch("htmx:sseBeforeMessage", { target: transcript });
  doc.dispatch("htmx:sseMessage", { target: transcript });
  assert.equal(chip.hidden, false, "chip appears for a scrolled-away push");
  assert.equal(chip.textContent, "↓ 1 new");

  transcript.scrollTop = 0; // still away
  doc.dispatch("htmx:sseBeforeMessage", { target: transcript });
  doc.dispatch("htmx:sseMessage", { target: transcript });
  assert.equal(chip.textContent, "↓ 2 new", "pushes accumulate");

  // Scrolling back to the bottom by hand hides and resets the chip.
  transcript.scrollTop = 480;
  doc.dispatch("scroll", { target: transcript });
  assert.equal(chip.hidden, true, "chip hides at the bottom");
  assert.equal(chip.textContent, "");

  // A near-bottom push pins to the newest bubble instead of counting.
  doc.dispatch("htmx:sseBeforeMessage", { target: transcript });
  doc.dispatch("htmx:sseMessage", { target: transcript });
  await new Promise((resolve) => setTimeout(resolve, 0));
  assert.equal(transcript.scrollTop, 1000, "near-bottom push pins the scroll");
  assert.equal(chip.hidden, true, "no chip for a near-bottom push");

  wrap.remove();
  delete transcript.dataset.page;
  delete transcript.dataset.thread;
});

test("clicking the jump chip returns to the newest bubble and resets", () => {
  const transcript = doc.getElementById("thread-transcript");
  transcript.dataset.page = "0";
  transcript.dataset.thread = "t-8";
  transcript.scrollHeight = 1000;
  transcript.clientHeight = 500;
  transcript.scrollTop = 0;

  const wrap = doc.createElement();
  const chip = doc.createElement();
  chip.className = "wp-jump-latest";
  chip.hidden = true;
  wrap.append(transcript, chip);

  doc.dispatch("htmx:sseBeforeMessage", { target: transcript });
  doc.dispatch("htmx:sseMessage", { target: transcript });
  assert.equal(chip.hidden, false);

  chip.closest = (selector) => (selector === ".wp-jump-latest" ? chip : null);
  doc.dispatch("click", { target: chip });
  assert.equal(transcript.scrollTop, 1000, "click jumps to the bottom");
  assert.equal(chip.hidden, true, "click resets the chip");

  wrap.remove();
  delete transcript.dataset.page;
  delete transcript.dataset.thread;
});

// 2d. Missed-call presence: island wp:call-missed events count into the
// header badge; opening the History tab clears it (a REJECT never
// dispatches — that lives in the island tests).
test("missed calls badge in the header and clear when History opens", () => {
  const actions = doc.createElement();
  actions.className = "wp-header-actions";
  const realQuerySelector = doc.querySelector.bind(doc);
  doc.querySelector = (selector) =>
    selector === ".wp-header-actions" ? actions : realQuerySelector(selector);

  doc.dispatch("wp:call-missed", {});
  const badge = actions.children.find((el) => el.id === "missed-badge");
  assert.ok(badge, "badge appears on the first missed call");
  assert.equal(badge.textContent, "missed · 1");
  assert.equal(actions.children[0], badge, "badge leads the header actions");

  doc.dispatch("wp:call-missed", {});
  assert.equal(badge.textContent, "missed · 2", "counts accumulate on one node");

  const historyNav = { closest: (sel) => (sel === "[data-tab='history']" ? historyNav : null) };
  doc.dispatch("click", { target: historyNav });
  assert.equal(
    actions.children.some((el) => el.id === "missed-badge"),
    false,
    "opening History clears the badge",
  );

  doc.querySelector = realQuerySelector;
});

// 3b-3. Thread-search guard: a live threads push must not stomp an
// active search; it goes through once the box is empty again. The next
// keystroke's debounced fetch re-renders the list, so nothing is lost.
test("live threads pushes are cancelled while a search query is active", () => {
  const list = doc.createElement();
  list.className = "wp-thread-list";
  // The registry auto-creates the search input the guard reads.
  const input = doc.getElementById("wp-thread-search-input");
  input.value = "launch code";

  let prevented = 0;
  doc.dispatch("htmx:sseBeforeMessage", {
    target: list,
    preventDefault: () => (prevented += 1),
  });
  assert.equal(prevented, 1, "push cancelled while the query is set");

  input.value = "   ";
  doc.dispatch("htmx:sseBeforeMessage", {
    target: list,
    preventDefault: () => (prevented += 1),
  });
  assert.equal(prevented, 1, "whitespace-only counts as an empty box");

  input.value = "";
  doc.dispatch("htmx:sseBeforeMessage", {
    target: list,
    preventDefault: () => (prevented += 1),
  });
  assert.equal(prevented, 1, "empty box lets the push through");

  // An unrelated target (no wp-thread-list class) is never touched by
  // the guard — a transcript id here would trip the 3b paging guard
  // instead, so this element carries a neutral id.
  const other = doc.createElement();
  other.id = "some-other-region";
  doc.dispatch("htmx:sseBeforeMessage", {
    target: other,
    preventDefault: () => (prevented += 1),
  });
  assert.equal(prevented, 1, "the guard only guards the thread list");
});

// 3e. Optimistic send: the reply composer appends a pending bubble on
// submit; a failure flips it to the failed state and restores the draft.
const makeComposeForm = () => {
  const area = {
    selector: "textarea.wp-compose-body",
    value: "optimistic hello",
  };
  const files = { selector: 'input[type="file"]', files: [] };
  const form = {
    selector: "form.wp-compose",
    className: "wp-compose",
    dataset: {},
    matches: (sel) => sel === "form.wp-compose",
    classList: {
      contains: (name) => form.className.split(" ").includes(name),
    },
    querySelector(sel) {
      if (sel === "textarea.wp-compose-body") return area;
      if (sel === 'input[type="file"]') return files;
      return null;
    },
  };
  return { form, area, files };
};

test("the reply composer appends a pending bubble the moment it submits", () => {
  const transcript = doc.getElementById("thread-transcript");
  const before = transcript.children.length;
  const { form } = makeComposeForm();
  doc.dispatch("htmx:beforeRequest", { target: form });
  assert.equal(transcript.children.length, before + 1);
  const bubble = transcript.children.at(-1);
  assert.match(bubble.className, /\bwp-bubble\b/);
  assert.match(bubble.className, /\bwp-out\b/);
  assert.match(bubble.className, /\bwp-opt\b/);
  assert.match(bubble.querySelector(".wp-bubble-body").textContent, /optimistic hello/);
  assert.match(bubble.querySelector(".wp-status").className, /wp-status-queued/);
  assert.match(bubble.querySelector(".wp-status").textContent, /sending/);
});

test("attachment-only sends render the optimistic attachment chips", () => {
  const transcript = doc.getElementById("thread-transcript");
  const before = transcript.children.length;
  const { form, area, files } = makeComposeForm();
  area.value = "";
  files.files = [{ name: "cat.png" }, { name: "notes.pdf" }];
  doc.dispatch("htmx:beforeRequest", { target: form });
  assert.equal(transcript.children.length, before + 1);
  const bubble = transcript.children.at(-1);
  const chips = bubble.children.filter((kid) =>
    String(kid.className).split(" ").includes("wp-attachment"),
  );
  assert.equal(chips.length, 2);
  assert.match(chips[0].textContent, /cat\.png/);
  assert.match(bubble.querySelector(".wp-status").textContent, /sending/);
});

test("a failed send flips the optimistic bubble to failed and restores the draft", () => {
  const transcript = doc.getElementById("thread-transcript");
  const before = transcript.children.length;
  const { form, area } = makeComposeForm();
  doc.dispatch("htmx:beforeRequest", { target: form });
  assert.equal(transcript.children.length, before + 1);
  area.value = ""; // htmx may already have reset the form when errors fire
  doc.dispatch("htmx:responseError", { target: form });
  const bubble = transcript.children.at(-1);
  assert.match(bubble.className, /\bwp-opt-failed\b/);
  assert.match(bubble.querySelector(".wp-status").className, /wp-status-failed/);
  assert.match(bubble.querySelector(".wp-status").textContent, /failed/);
  assert.equal(area.value, "optimistic hello", "the draft comes back");
  // A validation failure (HX-Trigger) still rolls the bubble back via
  // the same listener, while the 3c toast contract stays untouched.
  doc.dispatch("htmx:responseError", {
    target: form,
    detail: { xhr: fakeXhr(422, { "HX-Trigger": '{"showMessage":{}}' }) },
  });
  assert.equal(area.value, "optimistic hello");
});

test("the new-conversation composer and non-compose forms never get a bubble", () => {
  const transcript = doc.getElementById("thread-transcript");
  const before = transcript.children.length;
  const fresh = makeComposeForm();
  fresh.form.className = "wp-compose wp-compose-new";
  doc.dispatch("htmx:beforeRequest", { target: fresh.form });
  assert.equal(transcript.children.length, before);
  doc.dispatch("htmx:beforeRequest", {
    target: { selector: "form.wp-other" },
  });
  assert.equal(transcript.children.length, before);
  const empty = makeComposeForm();
  empty.area.value = "   ";
  doc.dispatch("htmx:beforeRequest", { target: empty.form });
  assert.equal(transcript.children.length, before, "no empty-body bubbles");
});

// 5. Command palette: Ctrl/Cmd-K opens the overlay over collected
// commands, filtering narrows it, Enter runs the selection, Escape
// closes; "?" opens the shortcut help without touching the island.
// The overlay wires its input listeners on the ELEMENT (not document),
// so element events are fired through the stub's listener registry.
const realQuerySelectorAll = doc.querySelectorAll.bind(doc);
const fireOn = (el, type, props = {}) =>
  (el.listeners[type] ?? []).forEach((fn) =>
    fn({ target: el, preventDefault: () => {}, ...props }),
  );
const findOverlay = () => doc.body.children.find((el) => el.id === "wp-palette");

test("Ctrl+K opens the palette, filters, runs, and closes", () => {
  const clicked = [];
  const link = {
    id: "nav-messages",
    textContent: "Messages",
    closest: () => null,
    hasAttribute: (name) => name === "data-tab",
    getAttribute: () => "messages",
    className: "wp-nav-link wp-active",
    click: () => clicked.push("messages"),
    addEventListener() {},
  };
  doc.querySelectorAll = () => [link];
  try {
    doc.dispatch("keydown", {
      key: "k",
      ctrlKey: true,
      preventDefault: () => {},
    });
    const overlay = findOverlay();
    assert.ok(overlay, "overlay created");
    assert.equal(overlay.hidden, false);
    const input = overlay.children[0];
    const list = overlay.children[1];
    assert.ok(list.children.length >= 4, "tabs plus shell actions listed");
    assert.match(list.children[0].textContent, /Go to Messages/);

    input.value = "theme";
    fireOn(input, "input");
    assert.equal(list.children.length, 1);
    assert.match(list.children[0].textContent, /theme/i);

    input.value = "";
    fireOn(input, "input");
    fireOn(input, "keydown", { key: "Enter" });
    assert.deepEqual(clicked, ["messages"], "Enter runs the selected row");
    assert.equal(overlay.hidden, true, "overlay closes after running");
  } finally {
    doc.querySelectorAll = realQuerySelectorAll;
  }
});

test("Escape closes the palette without running anything", () => {
  doc.querySelectorAll = () => [];
  try {
    doc.dispatch("keydown", {
      key: "k",
      metaKey: true,
      preventDefault: () => {},
    });
    const overlay = findOverlay();
    assert.equal(overlay.hidden, false);
    const input = overlay.children[0];
    fireOn(input, "keydown", { key: "Escape" });
    assert.equal(overlay.hidden, true);
  } finally {
    doc.querySelectorAll = realQuerySelectorAll;
  }
});

test("? opens the shortcut help outside typing fields", () => {
  doc.dispatch("keydown", {
    key: "?",
    target: { closest: () => null },
    preventDefault: () => {},
  });
  const overlay = findOverlay();
  assert.equal(overlay.hidden, false);
  const list = overlay.children[1];
  assert.ok(list.children.length >= 8, "every real binding is listed");
  assert.match(list.children[0].textContent, /Answer/);

  const input = overlay.children[0];
  fireOn(input, "keydown", { key: "Escape" });
  assert.equal(overlay.hidden, true);

  // Typing "?" inside a field must not hijack the key.
  doc.dispatch("keydown", {
    key: "?",
    target: {
      closest: (sel) => (sel === "input, textarea, select, [contenteditable='true']" ? {} : null),
    },
    preventDefault: () => {
      throw new Error("preventDefault should not fire while typing");
    },
  });
  assert.equal(overlay.hidden, true, "stays closed");
});

// 3h. Tab skeleton (F1): a NAVIGATING swap (tab link, thread row, back
// link) reveals the shimmer while the partial is in flight; a
// typing-driven fetch (search, composer) deliberately stays quiet —
// morph keeps those surfaces alive and a flash there would be noise.
test("the tab skeleton reveals during a navigating swap and hides after", () => {
  const skeleton = doc.getElementById("wp-tab-skeleton");
  const tabLink = {
    matches: () => false,
    hasAttribute: (attr) => attr === "data-tab",
    closest: () => null,
  };
  skeleton.hidden = true;
  doc.dispatch("htmx:beforeRequest", { target: tabLink, detail: {} });
  assert.equal(skeleton.hidden, false, "navigating swap reveals the skeleton");
  doc.dispatch("htmx:afterRequest", { target: tabLink, detail: {} });
  assert.equal(skeleton.hidden, true, "the settled swap hides it again");
});

test("typing-driven fetches never flash the skeleton", () => {
  const skeleton = doc.getElementById("wp-tab-skeleton");
  const search = {
    matches: () => false,
    hasAttribute: () => false,
    closest: () => null,
  };
  skeleton.hidden = true;
  doc.dispatch("htmx:beforeRequest", { target: search, detail: {} });
  assert.equal(skeleton.hidden, true, "search/composer stays quiet");
});

// 1. Nav active state after a partial swap (shell §1): the server marks
// the active link on full renders, but after a partial swap only the
// clicked link knows — the shell mirrors wp-active onto aria-current so
// screen readers keep the active-tab announcement in step.
test("a partial swap moves aria-current to the clicked nav link", () => {
  const clicked = doc.createElement();
  clicked.className = "wp-nav-link";
  clicked.setAttribute("data-tab", "messages");
  const other = doc.createElement();
  other.className = "wp-nav-link";
  const nav = doc.createElement();
  nav.querySelectorAll = (selector) => (selector === ".wp-nav-link" ? [clicked, other] : []);
  clicked.closest = (selector) => (selector === ".wp-nav" ? nav : null);

  doc.dispatch("htmx:afterRequest", { target: clicked });
  assert.equal(clicked.getAttribute("aria-current"), "page");
  assert.equal(other.getAttribute("aria-current"), "false");
  assert.ok(clicked.classList.contains("wp-active"));
  assert.ok(!other.classList.contains("wp-active"));
});

// 3e (edge). The successful server swap owns the bubble: once the send
// settles OK the pending record is dropped, so a LATER unrelated error
// must not resurrect a rollback of the already-sent bubble (the morph
// swap replaced it with the real one).
test("a settled send is never rolled back by a later error", () => {
  const transcript = doc.getElementById("thread-transcript");
  const before = transcript.children.length;
  const { form } = makeComposeForm();
  // The nav-active listener (shell §1) fires on every htmx:afterRequest;
  // give the form the DOM method it probes.
  form.hasAttribute = () => false;
  doc.dispatch("htmx:beforeRequest", { target: form });
  assert.equal(transcript.children.length, before + 1);
  const bubble = transcript.children.at(-1);

  doc.dispatch("htmx:afterRequest", { target: form, detail: { successful: true } });
  doc.dispatch("htmx:responseError", { target: form });
  assert.match(bubble.className, /\bwp-opt\b/, "still the pending bubble");
  assert.doesNotMatch(bubble.className, /wp-opt-failed/, "no rollback after success");
  assert.match(bubble.querySelector(".wp-status").textContent, /sending/);
});

test("a tab navigation stores the last-active tab (E3)", () => {
  localStorage.removeItem("wp-last-tab");
  const nav = doc.createElement();
  nav.className = "wp-nav";
  nav.querySelectorAll = (selector) => (selector === ".wp-nav-link" ? [link] : []);
  const link = doc.createElement();
  link.setAttribute("data-tab", "voicemail");
  link.className = "wp-nav-link";
  nav.append(link);
  doc.dispatch("htmx:afterRequest", { target: link });
  assert.equal(localStorage.getItem("wp-last-tab"), "voicemail");
  localStorage.removeItem("wp-last-tab");
});

test("a plain / load restores the remembered tab (E3)", () => {
  const calls = [];
  globalThis.window.htmx = {
    ajax: (verb, url) => calls.push(url),
  };
  globalThis.window.location = { pathname: "/" };

  // Signed-in shell: real tab content (no welcome hint), a stored tab,
  // and a rendered nav link for it.
  const content = doc.getElementById("tab-content");
  const panel = doc.createElement();
  panel.className = "wp-panel";
  content.append(panel);
  const nav = doc.getElementById("wp-nav");
  const link = doc.createElement();
  link.setAttribute("data-tab", "history");
  link.className = "wp-nav-link";
  nav.append(link);
  // The stub document's querySelector is a null sink; the restore path
  // reads the nav through it, so route the selector it asks for.
  const realQuerySelector = doc.querySelector;
  doc.querySelector = (selector) =>
    selector === '#wp-nav .wp-nav-link[data-tab="history"]'
      ? link
      : realQuerySelector.call(doc, selector);
  localStorage.setItem("wp-last-tab", "history");
  const replaced = [];
  globalThis.window.history = {
    replaceState: (...args) => replaced.push(args[2]),
  };

  shellApi.restoreLastTab();

  assert.deepEqual(
    calls,
    ["/partials/history", "/partials/nav?active=history"],
    "the stored tab is fetched, content + nav",
  );
  assert.deepEqual(replaced, ["/history"], "the address follows the content");

  doc.querySelector = realQuerySelector;
  localStorage.removeItem("wp-last-tab");
  delete globalThis.window.history;
  delete globalThis.window.htmx;
});

test("restore never overrides a deep link, garbage, or the sign-in hint (E3)", () => {
  const calls = [];
  globalThis.window.htmx = { ajax: (verb, url) => calls.push(url) };
  globalThis.window.location = { pathname: "/messages" };
  localStorage.setItem("wp-last-tab", "history");

  shellApi.restoreLastTab();
  assert.deepEqual(calls, [], "a deep link keeps its own tab");

  globalThis.window.location = { pathname: "/" };
  localStorage.setItem("wp-last-tab", "not-a-tab");
  shellApi.restoreLastTab();
  assert.deepEqual(calls, [], "a stored value no tab renders is ignored");

  // With the link rendered again, the welcome hint is the guard that
  // must refuse the swap (not a missing link).
  localStorage.setItem("wp-last-tab", "history");
  const nav = doc.getElementById("wp-nav");
  const link = doc.createElement();
  link.setAttribute("data-tab", "history");
  link.className = "wp-nav-link";
  nav.append(link);
  const realQuerySelector = doc.querySelector;
  doc.querySelector = (selector) =>
    selector === '#wp-nav .wp-nav-link[data-tab="history"]'
      ? link
      : realQuerySelector.call(doc, selector);
  const content = doc.getElementById("tab-content");
  const welcome = doc.createElement();
  welcome.className = "wp-welcome";
  content.append(welcome);
  shellApi.restoreLastTab();
  assert.deepEqual(calls, [], "the sign-in hint is never replaced");

  welcome.remove();
  doc.querySelector = realQuerySelector;
  localStorage.removeItem("wp-last-tab");
  delete globalThis.window.htmx;
});

test("refreshNav scrolls the active tab back into view (E8)", () => {
  const calls = [];
  globalThis.window.htmx = { ajax: (verb, url) => calls.push(url) };
  const link = doc.createElement();
  link.className = "wp-nav-link wp-active";
  // refreshNav reads the tab from dataset (disconnected from
  // attributes in the stub), so seed it there.
  link.dataset.tab = "history";
  const scrolled = [];
  link.scrollIntoView = (opts) => scrolled.push(opts);
  const realQuerySelector = doc.querySelector;
  doc.querySelector = (selector) =>
    selector === "#wp-nav .wp-nav-link.wp-active"
      ? link
      : realQuerySelector.call(doc, selector);

  shellApi.refreshNav();

  assert.deepEqual(
    calls,
    ["/partials/nav?active=history"],
    "the active tab survives the nav re-fetch",
  );
  assert.deepEqual(
    scrolled,
    [{ block: "nearest", inline: "nearest" }],
    "nearest keeps in-view tabs untouched",
  );

  // No scrollIntoView on the node (old browsers): the guard skips the
  // scroll without breaking the re-fetch.
  delete link.scrollIntoView;
  shellApi.refreshNav();
  assert.equal(calls.length, 2, "the re-fetch itself is unaffected");

  doc.querySelector = realQuerySelector;
  delete globalThis.window.htmx;
});

// --- voicemail player (M13 C1–C3/C9) ------------------------------------

const makeAudioStub = (id, src) => {
  const audio = doc.createElement();
  audio.id = id;
  audio.setAttribute("src", src);
  audio.paused = true;
  audio.controls = false;
  const played = [];
  audio.play = () => {
    played.push("play");
    audio.paused = false;
    return Promise.resolve();
  };
  audio.pause = () => {
    played.push("pause");
    audio.paused = true;
  };
  audio.playCalls = played;
  return audio;
};

const makePlayerRow = (uuid, { unread = true } = {}) => {
  const row = doc.createElement();
  row.id = "vm-" + uuid;
  row.className = "wp-row wp-vm-row" + (unread ? " wp-unread" : "");
  const audio = makeAudioStub("vm-audio-" + uuid, "/phone-api/x");
  const play = doc.createElement();
  play.id = "vm-play-" + uuid;
  play.className = "wp-mini wp-vm-play";
  play.setAttribute("data-vm-play", uuid);
  play.setAttribute("data-label-play", "Play");
  play.setAttribute("data-label-pause", "Pause");
  const speed = doc.createElement();
  speed.id = "vm-speed-" + uuid;
  speed.className = "wp-mini wp-vm-speed";
  speed.setAttribute("data-vm-speed", uuid);
  row.append(audio, play, speed);
  return { row, audio, play, speed };
};

test("vmPeaks reduces PCM to max-abs bars", () => {
  const empty = shellApi.vmPeaks([], 4);
  assert.deepEqual(empty, [0, 0, 0, 0], "silence renders as flat bars");

  const two = shellApi.vmPeaks([0, -1, 0.5, 0.25], 2);
  assert.deepEqual(two, [1, 0.5], "each bar is the max abs of its slice");

  const carried = shellApi.vmPeaks([0.5], 3);
  assert.deepEqual(carried, [0.5, 0.5, 0.5], "bars beyond the samples carry the neighbor");
});

test("vmClock renders the server's m:ss shape", () => {
  assert.equal(shellApi.vmClock(0), "0:00");
  assert.equal(shellApi.vmClock(5.9), "0:05");
  assert.equal(shellApi.vmClock(65), "1:05");
  assert.equal(shellApi.vmClock(-3), "0:00");
  assert.equal(shellApi.vmClock(undefined), "0:00");
  assert.equal(shellApi.vmClock(NaN), "0:00");
});

test("the play button drives the audio and falls back honestly", () => {
  const { row, audio, play } = makePlayerRow("u1");
  // No fetch in node: the waveform build must fail → native controls.
  doc.dispatch("click", { target: play });
  assert.deepEqual(audio.playCalls, ["play"], "the click starts the audio");
  assert.equal(audio.controls, true, "without WebAudio the native UI takes over");
  assert.equal(play.hidden, true, "the custom play button hides in the fallback");
  const toast = toasts().children.at(-1);
  assert.match(toast.textContent, /built-in player controls/);
});

test("playing marks the row, swaps the label, and clears unread + badge", () => {
  const badge = doc.getElementById("wp-nav-vm-badge");
  badge.textContent = "2";
  badge.hidden = false;
  const { row, audio, play } = makePlayerRow("u2");

  doc.dispatch("click", { target: play });
  // The play listener was wired lazily on first interaction.
  const onPlay = audio.listeners.play[0];
  onPlay();
  assert.ok(row.classList.contains("wp-playing"), "C3: the row is marked playing");
  assert.equal(play.textContent, "⏸");
  assert.equal(play.getAttribute("aria-label"), "Pause", "label swaps per the session language");
  assert.ok(!row.classList.contains("wp-unread"), "C9: unread styling drops on play");
  assert.equal(badge.textContent, "1", "C9: the nav badge mirrors the clear");

  const onPause = audio.listeners.pause[0];
  onPause();
  assert.ok(!row.classList.contains("wp-playing"));
  assert.equal(play.textContent, "▶");
  assert.equal(play.getAttribute("aria-label"), "Play");

  badge.hidden = true;
  badge.textContent = "";
});

test("the speed toggle walks the ladder and writes it back", () => {
  const { audio, speed } = makePlayerRow("u3");
  const labels = [speed.textContent];
  for (let i = 0; i < 5; i++) {
    doc.dispatch("click", { target: speed });
    labels.push(speed.textContent);
  }
  assert.deepEqual(
    labels,
    ["1×", "1.5×", "2×", "0.5×", "1×", "1.5×"],
    "the ladder cycles 1 → 1.5 → 2 → 0.5 → 1",
  );
  assert.equal(audio.playbackRate, 1.5, "the rate lands on the audio element");
});
