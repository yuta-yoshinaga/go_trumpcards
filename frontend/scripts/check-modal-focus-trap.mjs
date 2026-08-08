#!/usr/bin/env bun
// Guard that anything claiming to be a modal actually traps focus.
//
// `aria-modal="true"` tells assistive tech that everything outside the dialog is
// inert. If focus can still walk out of it, the AT's model and the real focus
// disagree, and the user lands on elements they were told do not exist — worse
// than never having claimed modality at all.
//
// Issue #4312 consolidated eight hand-rolled dialogs onto `common/Modal` +
// `useFocusTrap`, but four kept private copies that then drifted: all four lost
// the "pull focus back if it escaped" branch, and three bound their key handler
// to the dialog element rather than `document`, so once focus did leave, the
// handler stopped firing and the trap was dead for the rest of the dialog's life
// (issues #5182, #5183, #5184). Every one of those had tests that passed.
//
// The rule is therefore structural rather than behavioural: a component that
// declares `aria-modal` must get its focus handling from the shared hook or the
// shared component, not from an effect of its own.
//
// Non-modal panels are deliberately out of scope. `ActionLogPanel` is a landmark
// `role="region"` and must NOT trap (WCAG 2.1.2) — it opts out with
// `useFocusTrap(..., { trap: false })`, and since it never declares `aria-modal`
// this guard does not look at it. So the guard says "claim modality ⇒ trap", not
// "trap everywhere".
//
// Usage: check-modal-focus-trap.mjs [srcDir]

import { readdir, readFile } from 'node:fs/promises';
import { join, relative, resolve } from 'node:path';
import { fileURLToPath } from 'node:url';
import { assertFloor } from './lib/floor.mjs';

const ROOT = fileURLToPath(new URL('..', import.meta.url));
const SRC_DIR = process.argv[2] ? resolve(process.argv[2]) : join(ROOT, 'src');

/** Comments are blanked, not removed, so reported line numbers stay accurate. */
function stripComments(text) {
  return text
    .replace(/\/\*[\s\S]*?\*\//g, (m) => m.replace(/[^\n]/g, ' '))
    .replace(/\/\/[^\n]*/g, (m) => ' '.repeat(m.length));
}

async function walk(dir) {
  const out = [];
  let entries;
  try {
    entries = await readdir(dir, { withFileTypes: true });
  } catch {
    return out;
  }
  for (const entry of entries) {
    const full = join(dir, entry.name);
    if (entry.isDirectory()) out.push(...(await walk(full)));
    else if (entry.name.endsWith('.tsx') && !/\.(test|spec)\.tsx$/.test(entry.name)) out.push(full);
  }
  return out;
}

const files = await walk(SRC_DIR);

// Blanking comments matters here: ActionLogPanel's source explains *why* it has
// no `aria-modal`, and a raw substring search flags it for saying so.
const ARIA_MODAL = /aria-modal\s*=\s*[{"']?\s*(?:true|"true"|'true')/;
const USES_HOOK = /\buseFocusTrap\s*\(/;
const USES_MODAL = /<Modal[\s/>]/;

const violations = [];
let declaringModality = 0;

for (const file of files) {
  const src = stripComments(await readFile(file, 'utf8'));
  if (!ARIA_MODAL.test(src)) continue;
  declaringModality += 1;
  if (USES_HOOK.test(src) || USES_MODAL.test(src)) continue;
  const lines = src.split('\n');
  const line = lines.findIndex((l) => ARIA_MODAL.test(l)) + 1;
  violations.push({ file: relative(ROOT, file), line });
}

if (violations.length > 0) {
  console.error('\nComponents declaring aria-modal without a shared focus trap:\n');
  for (const v of violations) {
    console.error(`  ${v.file}:${v.line}  declares aria-modal but uses neither useFocusTrap nor <Modal>`);
  }
  console.error(
    '\nUse <Modal> (components/common/Modal.tsx) or useFocusTrap (hooks/useFocusTrap.ts).\n' +
      'A private copy loses the escaped-focus recovery and the document-level listener;\n' +
      'that is what issues #5182 / #5183 / #5184 were. For a NON-modal panel, drop\n' +
      'aria-modal and use useFocusTrap(..., { trap: false }) instead.\n',
  );
  process.exit(1);
}

// Three components declare aria-modal today: the shared Modal itself, plus the
// two bespoke overlays (manual, tutorial spotlight) whose layouts do not fit it.
// The count is low because adopting <Modal> *removes* the attribute from the
// caller — it moves into Modal. Floored at 2 so ordinary churn does not trip it,
// but a walk that stops finding .tsx files — or a regex that stops matching the
// attribute — fails loudly instead of reporting a clean tree it never inspected.
assertFloor('modal-focus-trap', declaringModality, 2, 'components declaring aria-modal');
console.log(`modal-focus-trap: OK (${declaringModality} components declare aria-modal; all use the shared trap).`);
