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
