// The drafts seam under node:test (F11/M29): the 30s panel poll must
// not fire while an edit <details> is open (the swap would eat the
// typed change), and the debounced composer draft files under the
// thread it was typed in even when the thread switches mid-timer.
import test from "node:test";
import assert from "node:assert/strict";

import { installBrowserGlobals } from "./helpers.mjs";

const doc = installBrowserGlobals();
await import("../shell.js");

test("the drafts poll is dropped while an edit is open", () => {
  const panel = doc.getElementById("wp-drafts");
  panel.matches = (s) => s === "#wp-drafts" || s === panel.selector;
  const edit = doc.createElement();
  edit.className = "wp-draft-edit";
  edit.selector = ".wp-draft-edit[open]";
  panel.append(edit);

  let dropped = 0;
  doc.dispatch("htmx:beforeRequest", {
    detail: { elt: panel },
    preventDefault() {
      dropped += 1;
    },
  });
  assert.equal(dropped, 1, "the poll is canceled while the edit holds the panel");

  edit.remove();
  dropped = 0;
  doc.dispatch("htmx:beforeRequest", {
    detail: { elt: panel },
    preventDefault() {
      dropped += 1;
    },
  });
  assert.equal(dropped, 0, "the poll flows once the edit closes");
});

test("a debounced draft files under the thread it was typed in", async () => {
  const transcript = doc.getElementById("thread-transcript");
  transcript.dataset.thread = "t1";
  const area = doc.createElement();
  area.selector = "textarea.wp-compose-body";
  area.value = "draft for t1";
  doc.dispatch("input", { target: area });

  transcript.dataset.thread = "t2";
  await new Promise((resolve) => setTimeout(resolve, 350));

  assert.equal(
    localStorage.getItem("wp-draft:t1"),
    "draft for t1",
    "the draft stays under t1 despite the mid-timer switch",
  );
  assert.equal(localStorage.getItem("wp-draft:t2"), null, "t2 never receives t1's text");
});
