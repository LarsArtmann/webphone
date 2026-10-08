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

// M17 J7: navigating between threads never cross-contaminates drafts —
// each thread restores its own, and a thread without one starts empty.
test("drafts survive thread navigation without cross-contamination", async () => {
  const transcript = doc.getElementById("thread-transcript");
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

  // Type in thread A, then navigate to thread B.
  transcript.dataset.thread = "t-a";
  composer.value = "draft for A";
  doc.dispatch("input", { target: composer });
  await new Promise((resolve) => setTimeout(resolve, 400));

  transcript.dataset.thread = "t-b";
  composer.value = "";
  doc.dispatch("htmx:afterSwap", {});
  assert.equal(composer.value, "", "thread B starts clean");

  // Type in B, navigate back to A: A's draft returns, B's stays stored.
  composer.value = "draft for B";
  doc.dispatch("input", { target: composer });
  await new Promise((resolve) => setTimeout(resolve, 400));

  transcript.dataset.thread = "t-a";
  composer.value = "";
  doc.dispatch("htmx:afterSwap", {});
  assert.equal(composer.value, "draft for A");
  assert.equal(localStorage.getItem("wp-draft:t-b"), "draft for B");

  doc.querySelector = realQuerySelector;
  delete transcript.dataset.thread;
  localStorage.removeItem("wp-draft:t-a");
  localStorage.removeItem("wp-draft:t-b");
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

// M17 J3/J4: the failed bubble carries its own recovery — Retry
// re-submits the reply composer, Dismiss drops the failed bubble while
// the draft stays in the composer.
test("the failed bubble offers Retry and Dismiss affordances", () => {
  const transcript = doc.getElementById("thread-transcript");
  const { form, area } = makeComposeForm();
  doc.dispatch("htmx:beforeRequest", { target: form });
  area.value = "";
  doc.dispatch("htmx:responseError", { target: form });
  const bubble = transcript.children.at(-1);
  const actions = bubble.querySelector(".wp-opt-actions");
  assert.ok(actions, "the failed bubble grows an actions row");
  const [retry, dismiss] = actions.children;
  assert.equal(retry.textContent, "Retry");
  assert.equal(dismiss.textContent, "Dismiss");

  // Retry: the failed bubble goes away and the CURRENT reply composer
  // (the draft was restored into it) submits again.
  const realQuerySelector = doc.querySelector.bind(doc);
  const submits = [];
  doc.querySelector = (selector) =>
    selector === "form.wp-compose:not(.wp-compose-new)"
      ? { requestSubmit: () => submits.push(1) }
      : realQuerySelector(selector);
  for (const fn of retry.listeners.click ?? []) fn();
  doc.querySelector = realQuerySelector;
  assert.equal(submits.length, 1, "retry re-submits the composer");
  assert.equal(bubble.parent, null, "the failed bubble is gone");

  // Dismiss: same removal, no submit — the draft stays in the composer.
  doc.dispatch("htmx:beforeRequest", { target: form });
  area.value = "";
  doc.dispatch("htmx:responseError", { target: form });
  const second = transcript.children.at(-1);
  const dismissAgain = second.querySelector(".wp-opt-actions").children[1];
  for (const fn of dismissAgain.listeners.click ?? []) fn();
  assert.equal(second.parent, null);
  assert.equal(submits.length, 1, "dismiss never submits");
  assert.equal(area.value, "optimistic hello", "the draft survives dismiss");
});

test("dismissing the welcome intro persists the choice and collapses now", () => {
  localStorage.removeItem("wp-welcome-dismissed");
  doc.documentElement.classList.remove("wp-welcome-dismissed");
  const button = doc.getElementById("wp-welcome-dismiss");
  button.className = "wp-mini wp-welcome-dismiss";
  doc.dispatch("click", { target: button });
  assert.equal(localStorage.getItem("wp-welcome-dismissed"), "1");
  assert.ok(
    doc.documentElement.classList.contains("wp-welcome-dismissed"),
    "the collapse applies without a reload",
  );

  // Unrelated clicks never mint the flag.
  localStorage.removeItem("wp-welcome-dismissed");
  doc.documentElement.classList.remove("wp-welcome-dismissed");
  const stranger = doc.getElementById("wp-welcome-stranger");
  doc.dispatch("click", { target: stranger });
  assert.equal(localStorage.getItem("wp-welcome-dismissed"), null);
  assert.ok(!doc.documentElement.classList.contains("wp-welcome-dismissed"));
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
    selector === "#wp-nav .wp-nav-link.wp-active" ? link : realQuerySelector.call(doc, selector);

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

// Builds one player row from REGISTRY instances: the shell looks every
// node up by getElementById, and only registry hits resolve to the same
// object the test holds.
const makePlayerRow = (uuid, { unread = true } = {}) => {
  const row = doc.createElement();
  row.id = "vm-" + uuid;
  row.className = "wp-row wp-vm-row" + (unread ? " wp-unread" : "");
  const audio = doc.getElementById("vm-audio-" + uuid);
  audio.setAttribute("src", "/phone-api/x");
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
  const play = doc.getElementById("vm-play-" + uuid);
  play.className = "wp-mini wp-vm-play";
  play.textContent = "▶";
  play.hidden = false;
  play.setAttribute("data-vm-play", uuid);
  play.setAttribute("data-label-play", "Play");
  play.setAttribute("data-label-pause", "Pause");
  const speed = doc.getElementById("vm-speed-" + uuid);
  speed.className = "wp-mini wp-vm-speed";
  speed.textContent = "1×";
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

test("the play button drives the audio and falls back honestly", async () => {
  const { row, audio, play } = makePlayerRow("u1");
  // No fetch in node: the waveform build must fail → native controls.
  doc.dispatch("click", { target: play });
  assert.deepEqual(audio.playCalls, ["play"], "the click starts the audio");
  // The fallback rides vmBuildWave's rejection — a microtask later.
  await new Promise((resolve) => setTimeout(resolve, 0));
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

// 2e. Reply snippets (M21): a chip fills the compose box — REPLACE, the
// data-sms prefill precedent — closes an open picker, and leaves
// strangers untouched.
test("a snippet chip replaces the reply text and focuses the composer", () => {
  const composer = doc.createElement("");
  composer.selector = "textarea.wp-compose-body";
  composer.value = "half-typed answer";
  const details = doc.createElement("");
  details.selector = "details";
  details.open = true;
  const chip = doc.createElement("");
  chip.selector = "[data-snippet]";
  chip.attrs["data-snippet"] = "Thanks, on it!";
  const form = doc.createElement("");
  form.selector = "form.wp-compose";
  form.append(composer, details);
  details.append(chip);

  doc.dispatch("click", { target: chip });
  assert.equal(composer.value, "Thanks, on it!");
  assert.ok(composer.focused, "the composer takes focus for a quick edit");
  assert.equal(details.open, false, "the picker closes after the fill");

  // A click with no snippet context changes nothing.
  const stranger = doc.createElement("");
  stranger.selector = "[data-dial]";
  doc.dispatch("click", { target: stranger });
  assert.equal(composer.value, "Thanks, on it!");
});

test("a snippet outside any compose form is a no-op", () => {
  const orphan = doc.createElement("");
  orphan.selector = "[data-snippet]";
  orphan.attrs["data-snippet"] = "nowhere to land";
  doc.dispatch("click", { target: orphan });
});

// 2f. Attachment lightbox (M21.1): the click opens the singleton dialog
// instead of navigating; a second open reuses it; the close button shuts
// it (ESC and the backdrop are native dialog behavior).
test("an image attachment opens the singleton lightbox, not a navigation", () => {
  const link = doc.createElement("");
  link.selector = "a[data-lightbox]";
  link.attrs.href = "/attachments/Att:first";
  link.attrs["data-lightbox"] = "pic.png";
  let prevented = false;
  doc.dispatch("click", {
    target: link,
    preventDefault() {
      prevented = true;
    },
  });
  assert.ok(prevented, "the download/navigation default is suppressed");

  const dialog = doc.body.children.find((child) => child.id === "wp-lightbox");
  assert.ok(dialog, "the dialog exists");
  assert.ok(dialog.opened, "showModal ran");
  const image = dialog.children.find((child) => child.src === "/attachments/Att:first");
  assert.ok(image, "the full-size blob src landed on the dialog image");
  assert.equal(image.alt, "pic.png");

  // The second open reuses the singleton and swaps the src.
  const second = doc.createElement("");
  second.selector = "a[data-lightbox]";
  second.attrs.href = "/attachments/Att:second";
  second.attrs["data-lightbox"] = "other.png";
  doc.dispatch("click", { target: second, preventDefault() {} });
  assert.equal(
    doc.body.children.filter((child) => child.id === "wp-lightbox").length,
    1,
    "one dialog, not one per image",
  );
  const imageAfter = dialog.children.find((child) => child.src === "/attachments/Att:second");
  assert.ok(imageAfter, "the singleton's image src follows the click");
  assert.equal(imageAfter.alt, "other.png");

  // The close button shuts it; a plain image click does not.
  const closeButton = dialog.children.find(
    (child) => child.type === "button" && child.textContent === "✕ Close",
  );
  closeButton.listeners.click.forEach((fn) => fn());
  assert.equal(dialog.opened, false);
});

// --- data-transcribe-src (server-rendered tabs) ------------------------------
// The delegated handler + the auto-start pass behind every server-rendered
// transcription surface (voicemail rows, MMS audio attachments).

const settle = async (rounds = 10) => {
  while (rounds--) await new Promise((resolve) => setImmediate(resolve));
};

test("a data-transcribe-src click fetches the audio, POSTs the seam, renders the text", async () => {
  const calls = [];
  globalThis.fetch = async (url, options = {}) => {
    calls.push({ url, options });
    if (String(url).startsWith("/vm-audio/")) {
      return { ok: true, status: 200, blob: async () => ({ type: "audio/wav" }) };
    }
    return { ok: true, status: 200, json: async () => ({ text: "  hi from voicemail  " }) };
  };
  const btn = doc.getElementById("vm-btn-spec");
  btn.selector = "[data-transcribe-src]";
  btn.setAttribute("data-transcribe-src", "/vm-audio/1001/abc.wav");
  btn.setAttribute("data-transcribe-target", "vm-transcript-spec");

  doc.dispatch("click", { target: btn });
  await settle();

  assert.equal(calls.length, 2, "one audio fetch, one seam POST");
  assert.equal(calls[0].url, "/vm-audio/1001/abc.wav");
  assert.match(calls[1].url, /^\/api\/transcribe\?/);
  assert.match(calls[1].url, /filename=abc\.wav/);
  assert.equal(calls[1].options.method, "POST");
  assert.equal(calls[1].options.headers["Content-Type"], "audio/wav");
  const target = doc.getElementById("vm-transcript-spec");
  assert.equal(target.textContent, "hi from voicemail");
  assert.equal(target.hidden, false);
});

test("auto-start transcribes rendered audio once per src+target, only with the seam on", async () => {
  const calls = [];
  globalThis.fetch = async (url, options = {}) => {
    calls.push({ url, options });
    if (String(url).startsWith("/vm-audio/")) {
      return { ok: true, status: 200, blob: async () => ({ type: "audio/wav" }) };
    }
    return { ok: true, status: 200, json: async () => ({ text: "auto text" }) };
  };
  const btn = doc.getElementById("vm-btn-auto");
  btn.selector = "[data-transcribe-src]";
  btn.setAttribute("data-transcribe-src", "/vm-audio/1001/auto.wav");
  btn.setAttribute("data-transcribe-target", "vm-transcript-auto");
  const target = doc.getElementById("vm-transcript-auto");
  const originalQuerySelectorAll = doc.querySelectorAll;
  doc.querySelectorAll = (selector) =>
    selector === "[data-transcribe-src]" ? [btn] : originalQuerySelectorAll(selector);
  const waitDebounce = () => new Promise((resolve) => setTimeout(resolve, 400));

  // Seam off (no PBX_CONFIG.asr): renders never POST on their own.
  doc.dispatch("htmx:afterSwap", {});
  await waitDebounce();
  assert.equal(calls.length, 0, "auto-start is gated on the ASR flag");

  // Seam on: the swap alone drives one transcription.
  globalThis.window.PBX_CONFIG = { asr: true };
  doc.dispatch("htmx:afterSwap", {});
  await waitDebounce();
  const seamPosts = () => calls.filter((c) => String(c.url).startsWith("/api/transcribe"));
  assert.equal(seamPosts().length, 1);
  assert.equal(target.textContent, "auto text");

  // A morph re-render of the SAME row never re-POSTs.
  doc.dispatch("htmx:afterSwap", {});
  await waitDebounce();
  assert.equal(seamPosts().length, 1);

  // A target that already carries text (a prior run preserved by the
  // morph) is never re-run — simulate by resetting the seen key through
  // a NEW src (a different message).
  btn.setAttribute("data-transcribe-src", "/vm-audio/1001/other.wav");
  target.textContent = "already transcribed";
  doc.dispatch("htmx:afterSwap", {});
  await waitDebounce();
  assert.equal(seamPosts().length, 1, "filled targets are left alone");

  delete globalThis.window.PBX_CONFIG;
  doc.querySelectorAll = originalQuerySelectorAll;
});

// M8 F31: a dozen rendered clips must not stampede the seam's flood
// budget — auto-runs ride ONE promise chain, so the second clip's first
// fetch only happens after the first run fully settles.
test("auto-runs serialize: the next clip starts only after the previous run settles", async () => {
  const calls = [];
  let releaseFirst;
  globalThis.fetch = async (url, options = {}) => {
    calls.push(String(url));
    if (String(url).startsWith("/vm-audio/2001/ser-a")) {
      return { ok: true, status: 200, blob: async () => ({ type: "audio/wav" }) };
    }
    if (String(url).startsWith("/vm-audio/2001/ser-b")) {
      return { ok: true, status: 200, blob: async () => ({ type: "audio/wav" }) };
    }
    if (String(url).includes("filename=ser-a.wav")) {
      // Gated: the first run stays in flight until the test releases it.
      return new Promise((resolve) => {
        releaseFirst = () =>
          resolve({ ok: true, status: 200, json: async () => ({ text: "first done" }) });
      });
    }
    return { ok: true, status: 200, json: async () => ({ text: "second done" }) };
  };

  const btnA = doc.getElementById("vm-btn-ser-a");
  const btnB = doc.getElementById("vm-btn-ser-b");
  for (const [btn, name] of [
    [btnA, "ser-a"],
    [btnB, "ser-b"],
  ]) {
    btn.selector = "[data-transcribe-src]";
    btn.setAttribute("data-transcribe-src", `/vm-audio/2001/${name}.wav`);
    btn.setAttribute("data-transcribe-target", `vm-transcript-${name}`);
  }
  const originalQuerySelectorAll = doc.querySelectorAll;
  doc.querySelectorAll = (selector) =>
    selector === "[data-transcribe-src]" ? [btnA, btnB] : originalQuerySelectorAll(selector);
  globalThis.window.PBX_CONFIG = { asr: true };

  doc.dispatch("htmx:afterSwap", {});
  await new Promise((resolve) => setTimeout(resolve, 400));
  const firstSeamAt = calls.findIndex((url) => url.includes("filename=ser-a.wav"));
  assert.ok(firstSeamAt >= 0, "the first clip's seam POST started");
  await settle();
  assert.equal(
    calls.filter((url) => url.startsWith("/vm-audio/2001/ser-b")).length,
    0,
    "the second clip has not even fetched its audio yet",
  );

  releaseFirst();
  await settle(20);
  const secondAudioAt = calls.findIndex((url) => url.startsWith("/vm-audio/2001/ser-b"));
  assert.ok(secondAudioAt > firstSeamAt, "the second run starts only after the first settles");
  assert.equal(doc.getElementById("vm-transcript-ser-a").textContent, "first done");
  assert.equal(doc.getElementById("vm-transcript-ser-b").textContent, "second done");

  delete globalThis.window.PBX_CONFIG;
  doc.querySelectorAll = originalQuerySelectorAll;
});

// M8 F31: a 429 from the flood budget is retried ONCE after the
// Retry-After delay — and a second 429 stops there (the bound keeps a
// hostile header from freezing the tab in a retry loop).
test("a 429 is retried once after Retry-After; a second 429 fails honestly", async () => {
  const calls = [];
  let ok429Attempts = 0;
  let fail429Attempts = 0;
  globalThis.fetch = async (url) => {
    calls.push(String(url));
    if (String(url).startsWith("/vm-audio/2001/")) {
      return { ok: true, status: 200, blob: async () => ({ type: "audio/wav" }) };
    }
    if (String(url).includes("filename=retry-ok.wav")) {
      ok429Attempts += 1;
      if (ok429Attempts === 1) {
        return { ok: false, status: 429, headers: new Headers({ "Retry-After": "1" }) };
      }
      return { ok: true, status: 200, json: async () => ({ text: "healed text" }) };
    }
    fail429Attempts += 1;
    return { ok: false, status: 429, headers: new Headers({ "Retry-After": "1" }) };
  };

  const okBtn = doc.getElementById("vm-btn-retry-ok");
  const failBtn = doc.getElementById("vm-btn-retry-fail");
  for (const [btn, name] of [
    [okBtn, "retry-ok"],
    [failBtn, "retry-fail"],
  ]) {
    btn.selector = "[data-transcribe-src]";
    btn.setAttribute("data-transcribe-src", `/vm-audio/2001/${name}.wav`);
    btn.setAttribute("data-transcribe-target", `vm-transcript-${name}`);
  }

  doc.dispatch("click", { target: okBtn });
  doc.dispatch("click", { target: failBtn });
  await settle();
  assert.equal(ok429Attempts, 1, "first attempt answered 429, retry still pending");
  assert.equal(fail429Attempts, 1, "no retry before the Retry-After delay");

  await new Promise((resolve) => setTimeout(resolve, 1300));
  await settle();
  assert.equal(ok429Attempts, 2, "exactly one retry after the delay");
  assert.equal(doc.getElementById("vm-transcript-retry-ok").textContent, "healed text");
  assert.equal(fail429Attempts, 2, "the second 429 also gets exactly one retry");
  assert.match(
    doc.getElementById("vm-transcript-retry-fail").textContent,
    /HTTP 429/,
    "a second 429 surfaces as a failure, never a loop",
  );
});

// M11: the History transcript delete affordance. Confirm-gated (the
// server's localized prompt), owner-scoped DELETE, honest removal — a
// failed delete keeps the row and says so.
test("data-delete-transcript confirms, DELETEs, and removes the row only on success", async () => {
  const calls = [];
  globalThis.fetch = async (url, options = {}) => {
    calls.push({ url: String(url), options });
    return { ok: true, status: 204 };
  };
  const confirmations = [];
  globalThis.window.confirm = (text) => {
    confirmations.push(text);
    return confirmations.length === 1 ? false : true;
  };

  const section = doc.createElement();
  section.className = "wp-transcripts";
  const row = doc.createElement();
  row.className = "wp-transcript-call";
  const btn = doc.createElement();
  btn.selector = "[data-delete-transcript]";
  btn.setAttribute("data-delete-transcript", "call-x");
  btn.setAttribute("data-confirm", "Wirklich löschen?");
  row.append(btn);
  section.append(row);

  // A declined confirm never reaches the network.
  doc.dispatch("click", { target: btn });
  await settle();
  assert.deepEqual(confirmations, ["Wirklich löschen?"], "the server's prompt is what the user sees");
  assert.equal(calls.length, 0, "declined confirm sends nothing");
  assert.equal(section.children.length, 1, "the row stays");

  // A failed delete keeps the row and toasts the truth.
  globalThis.fetch = async (url, options = {}) => {
    calls.push({ url: String(url), options });
    return { ok: false, status: 500 };
  };
  doc.dispatch("click", { target: btn });
  await settle();
  assert.equal(calls.length, 1);
  assert.equal(calls[0].url, "/api/transcripts?call=call-x");
  assert.equal(calls[0].options.method, "DELETE");
  assert.equal(calls[0].options.headers["X-CSRF-Token"], "", "the CSRF header rides the delete");
  assert.equal(section.children.length, 1, "a failed delete keeps the row");
  assert.match(
    toasts().children.at(-1).textContent,
    /could not delete/i,
    "the failure says so",
  );

  // Success: the row goes, and the emptied section goes with it.
  globalThis.fetch = async (url, options = {}) => {
    calls.push({ url: String(url), options });
    return { ok: true, status: 204 };
  };
  doc.dispatch("click", { target: btn });
  await settle();
  assert.equal(calls.length, 2);
  assert.equal(section.children.length, 0, "the row is gone");
  assert.equal(section.parent, null, "the emptied section removed itself");
  assert.match(toasts().children.at(-1).textContent, /transcript deleted/i);

  delete globalThis.window.confirm;
});

// M16: copy affordances. History rows carry data-copy-transcript; the
// on-demand chips (planted by the transcribe flow) carry data-copy-target.
test("data-copy-transcript copies the row's transcript and toasts honestly", async () => {
  const copied = [];
  globalThis.navigator.clipboard = {
    writeText: (text) => {
      copied.push(text);
      return Promise.resolve();
    },
  };
  const row = doc.createElement();
  row.className = "wp-transcript-call";
  const text = doc.createElement();
  text.className = "wp-transcript-text";
  text.textContent = "guten tag wie gehts";
  const btn = doc.createElement();
  btn.selector = "[data-copy-transcript]";
  btn.setAttribute("data-copy-transcript", "call-x");
  row.append(text, btn);

  doc.dispatch("click", { target: btn });
  await settle();
  assert.deepEqual(copied, ["guten tag wie gehts"]);
  assert.match(toasts().children.at(-1).textContent, /copied to clipboard/i);

  // An empty row says so instead of copying nothing.
  text.textContent = "";
  doc.dispatch("click", { target: btn });
  await settle();
  assert.equal(copied.length, 1);
  assert.match(toasts().children.at(-1).textContent, /nothing to copy yet/i);

  // A clipboard failure keeps the text and says so.
  globalThis.navigator.clipboard = {
    writeText: () => Promise.reject(new Error("denied")),
  };
  text.textContent = "real words";
  doc.dispatch("click", { target: btn });
  await settle();
  assert.match(toasts().children.at(-1).textContent, /could not copy/i);

  delete globalThis.navigator.clipboard;
});

test("a real transcription plants a copy chip; empty results plant nothing", async () => {
  const copied = [];
  globalThis.navigator.clipboard = {
    writeText: (text) => {
      copied.push(text);
      return Promise.resolve();
    },
  };
  let reply = { text: "" };
  globalThis.fetch = async (url) => {
    if (String(url).startsWith("/vm-audio/")) {
      return { ok: true, status: 200, blob: async () => ({ type: "audio/wav" }) };
    }
    return { ok: true, status: 200, json: async () => reply };
  };
  const holder = doc.createElement();
  const target = doc.getElementById("vm-transcript-chip-spec");
  holder.append(target);
  doc.body.append(holder);
  const btn = doc.getElementById("vm-btn-chip-spec");
  btn.selector = "[data-transcribe-src]";
  btn.setAttribute("data-transcribe-src", "/vm-audio/1001/chip.wav");
  btn.setAttribute("data-transcribe-target", "vm-transcript-chip-spec");
  // The chip guard probes document level; route the attribute selector
  // into the stub tree.
  const realQS = doc.querySelector;
  doc.querySelector = (selector) => {
    const match = /^\[data-copy-target="(.+)"\]$/.exec(selector);
    if (!match) return realQS(selector);
    const walk = (node) => {
      for (const kid of node.children) {
        if (kid.getAttribute && kid.getAttribute("data-copy-target") === match[1]) {
          return kid;
        }
        const found = walk(kid);
        if (found) return found;
      }
      return null;
    };
    return walk(doc.body);
  };
  const chips = () => holder.children.filter((c) => c.hasAttribute("data-copy-target"));

  // "No speech detected" is honest feedback, not text: no chip.
  reply = { text: "" };
  doc.dispatch("click", { target: btn });
  await settle();
  assert.equal(target.textContent, "No speech detected");
  assert.equal(chips().length, 0, "empty result plants no copy chip");

  // Real text plants exactly one chip, once per DOM lifetime.
  reply = { text: "  hi from voicemail  " };
  doc.dispatch("click", { target: btn });
  await settle();
  assert.equal(target.textContent, "hi from voicemail");
  assert.equal(chips().length, 1, "a chip follows the text");
  assert.equal(chips()[0].textContent, "Copy");

  // Clicking the chip copies the target's text.
  const chip = chips()[0];
  chip.selector = "[data-copy-target]";
  doc.dispatch("click", { target: chip });
  await settle();
  assert.deepEqual(copied, ["hi from voicemail"]);
  assert.match(toasts().children.at(-1).textContent, /copied to clipboard/i);

  doc.querySelector = realQS;
  delete globalThis.navigator.clipboard;
});
