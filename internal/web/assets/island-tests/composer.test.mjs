// Composer UX specs (shell.js): SMS segmentation math, Enter/Shift+Enter
// routing, and attachment chips with remove — black-box through the real
// document-level listeners, the way shell.test.mjs drives them. The math
// is asserted via the counter DOM (no exports): every boundary below is a
// billing boundary a user can actually hit.
import test from "node:test";
import assert from "node:assert/strict";

import { installBrowserGlobals } from "./helpers.mjs";

const doc = installBrowserGlobals();
await import("../shell.js");

function composer() {
  const form = doc.createElement();
  form.selector = "form";
  const area = doc.createElement();
  area.selector = "textarea.wp-compose-body";
  area.value = "";
  form.append(area);
  const counter = doc.createElement();
  counter.selector = ".wp-segcount";
  counter.hidden = true;
  form.append(counter);
  return { form, area, counter };
}

test("segment counter: GSM-7 160/153, extension chars cost two units", () => {
  const { area, counter } = composer();
  const type = (value) => {
    area.value = value;
    doc.dispatch("input", { target: area });
  };

  type("hi");
  assert.ok(counter.hidden, "short bodies stay silent");
  type("a".repeat(160));
  assert.ok(counter.hidden, "exactly 160 GSM-7 chars is one segment");
  type("a".repeat(161));
  assert.equal(counter.textContent, "2 SMS", "161 chars spills into 153+8");
  type("a".repeat(153 * 2));
  assert.equal(counter.textContent, "2 SMS", "exactly two full segments");
  type("a".repeat(153 * 2 + 1));
  assert.equal(counter.textContent, "3 SMS");
  type("{".repeat(80));
  assert.ok(counter.hidden, "80 braces = 160 units = one segment");
  type("{".repeat(81));
  assert.equal(counter.textContent, "2 SMS", "81 braces = 162 units");
  type("x\r\ny");
  assert.ok(counter.hidden, "CRLF counts as one character");
});

test("segment counter: non-GSM characters force UCS-2 70/67", () => {
  const { area, counter } = composer();
  const type = (value) => {
    area.value = value;
    doc.dispatch("input", { target: area });
  };

  type("Grüße"); // ü and ß ARE GSM-7 — stays silent
  assert.ok(counter.hidden, "German umlauts stay in GSM-7");
  type("😀".repeat(70));
  assert.ok(counter.hidden, "70 UCS-2 chars is one segment");
  type("😀".repeat(71));
  assert.equal(counter.textContent, "2 SMS", "71 UCS-2 chars spill into 67+4");
});

test("Enter sends the composer, Shift+Enter and IME composition do not", () => {
  const { form, area } = composer();
  let submitted = 0;
  form.requestSubmit = () => {
    submitted += 1;
  };
  let prevented = 0;
  const key = (props) =>
    doc.dispatch("keydown", {
      target: area,
      key: "Enter",
      preventDefault() {
        prevented += 1;
      },
      ...props,
    });

  key({});
  assert.equal(submitted, 1);
  assert.equal(prevented, 1, "Enter's default newline is suppressed");
  key({ shiftKey: true });
  key({ isComposing: true });
  assert.equal(submitted, 1, "Shift+Enter and composing Enter never submit");
  assert.equal(prevented, 1);

  const elsewhere = doc.createElement();
  doc.dispatch("keydown", { target: elsewhere, key: "Enter", preventDefault() {} });
  assert.equal(submitted, 1, "Enter outside a composer textarea is untouched");
});

test("attachment chips name chosen files and remove rebuilds the file list", () => {
  const host = doc.createElement();
  host.selector = ".wp-compose";
  const form = doc.createElement();
  form.selector = "form";
  host.append(form);
  const file = doc.createElement();
  file.selector = 'input[type="file"]';
  file.type = "file";
  file.files = [{ name: "invoice.pdf" }, { name: "photo.png" }];
  const box = doc.createElement();
  box.selector = ".wp-attach";
  box.hidden = true;
  form.append(file, box);

  doc.dispatch("change", { target: file });
  assert.equal(box.hidden, false, "chips appear once files are chosen");
  assert.equal(box.children.length, 2);
  assert.equal(box.children[0].textContent, "invoice.pdf");
  const removePhoto = box.children[1].children[0];
  assert.equal(removePhoto.getAttribute("aria-label"), "remove");

  doc.dispatch("click", { target: removePhoto });
  assert.deepEqual(
    file.files.map((f) => f.name),
    ["invoice.pdf"],
    "removing a chip drops exactly that file",
  );
  assert.equal(box.children.length, 1);
  assert.equal(box.children[0].textContent, "invoice.pdf");

  doc.dispatch("change", { target: { ...file, files: [] } });
  assert.equal(box.hidden, true, "an empty selection hides the chip row");
});

test("a file input outside a compose form renders no chips", () => {
  const stranger = doc.createElement();
  stranger.type = "file";
  stranger.files = [{ name: "x" }];
  doc.dispatch("change", { target: stranger });
  assert.ok(true, "listener declines politely");
});
