#!/usr/bin/env bun
// Guard against design-token regressions. Fails the build if source code under
// frontend/src uses forbidden raw Tailwind palette classes or opacity-suffixed
// `text-white/N`. See DESIGN.md and issue #1411.

import { readdir, readFile } from 'node:fs/promises';
import { join, relative } from 'node:path';
import { fileURLToPath } from 'node:url';

const ROOT = fileURLToPath(new URL('..', import.meta.url));
const SRC_DIR = join(ROOT, 'src');

const PALETTE =
  '(red|green|yellow|amber|blue|orange|purple|pink|emerald|sky|indigo|violet|fuchsia|rose|slate|zinc|gray|neutral|stone|cyan|teal|lime)';
const RULES = [
  {
    pattern: /\btext-white\/\d+\b/g,
    message: 'Use text-ds-text-primary or text-ds-text-muted instead of text-white/N (see DESIGN.md).',
  },
  {
    pattern: new RegExp(String.raw`\b(bg|text|border|ring|shadow|from|to|via|divide)-${PALETTE}-\d{2,3}(/\d+)?\b`, 'g'),
    message: 'Use design-system tokens (bg-ds-*, text-ds-*, border-ds-*, ring-ds-*) instead of raw Tailwind palette.',
  },
];

const INCLUDED_EXT = new Set(['.ts', '.tsx']);
const IGNORED = /\.(test|spec)\.(ts|tsx)$/;

async function walk(dir) {
  const entries = await readdir(dir, { withFileTypes: true });
  const files = [];
  for (const entry of entries) {
    const full = join(dir, entry.name);
    if (entry.isDirectory()) {
      files.push(...(await walk(full)));
    } else if (INCLUDED_EXT.has(entry.name.slice(entry.name.lastIndexOf('.')))) {
      if (!IGNORED.test(entry.name)) files.push(full);
    }
  }
  return files;
}

const files = await walk(SRC_DIR);
const violations = [];

for (const file of files) {
  const text = await readFile(file, 'utf8');
  const lines = text.split('\n');
  for (const { pattern, message } of RULES) {
    for (let i = 0; i < lines.length; i += 1) {
      for (const match of lines[i].matchAll(pattern)) {
        violations.push({
          file: relative(ROOT, file),
          line: i + 1,
          match: match[0],
          message,
        });
      }
    }
  }
}

if (violations.length > 0) {
  console.error('\nDesign-token policy violations:\n');
  for (const v of violations) {
    console.error(`  ${v.file}:${v.line}  [${v.match}]  ${v.message}`);
  }
  console.error(`\n${violations.length} violation(s). See DESIGN.md and issue #1411.`);
  process.exit(1);
}

// Reduced-motion SSoT guard (issue #4315): prefers-reduced-motion must be
// enforced by a single universal block, not a class-name allowlist that
// silently misses arbitrary animate-[…] utilities and future animations.
const indexCss = await readFile(join(SRC_DIR, 'index.css'), 'utf8');
const rmIdx = indexCss.indexOf('@media (prefers-reduced-motion: reduce)');
// Read to the at-rule's own closing brace — the only `}` in column 0 after it,
// since the nested `*` rule closes indented. This used to slice a fixed 400
// characters, which silently failed the moment a comment was added inside the
// block: the declarations it looks for slid past the window and the guard
// reported the SSoT block missing when it was right there.
const rmEnd = rmIdx === -1 ? -1 : indexCss.indexOf('\n}', rmIdx);
const rmBlock = rmIdx === -1 || rmEnd === -1 ? '' : indexCss.slice(rmIdx, rmEnd + 2);
if (!rmBlock.includes('*,') || !rmBlock.includes('animation-duration: 0.01ms')) {
  console.error(
    '\nreduced-motion: index.css must enforce prefers-reduced-motion with a universal\n' +
      '(`*, *::before, *::after`) block that near-zeroes animation/transition durations,\n' +
      'not a per-class allowlist. See DESIGN.md Motion section and issue #4315.',
  );
  process.exit(1);
}

console.log(`design-tokens: OK (${files.length} source files scanned, test files skipped).`);
console.log('reduced-motion: OK (index.css uses the universal prefers-reduced-motion block).');

// Tap-target guard (issue #4368): DESIGN.md's "Interactive Element Minimum
// Size" rule requires every interactive control to hit a 44x44 CSS px target
// (WCAG 2.5.5 AAA). The shared components already comply -- buttonStyles.ts
// puts min-h-[44px] in its base and SettingsPanel puts it on both the wrapping
// span and the select -- but page-level implementations kept drifting, because
// nothing checked. A ~16px checkbox is invisible to a type checker and to a
// DOM test that only asserts behaviour.
//
// The rule is deliberately asymmetric, matching DESIGN.md:
//   - select / number / text inputs must carry min-h-[44px] themselves, since
//     the control *is* the target.
//   - checkbox / radio keep their small visual dot, so the requirement lands on
//     the wrapping <label> instead, making the whole row tappable. A sibling
//     <label htmlFor> does not count: it leaves the 16px box as its own
//     undersized target.
const TAP_TARGET = 'min-h-[44px]';

/**
 * Opening tag of the element starting at `<` index `start`.
 *
 * Scans for the `>` that actually closes the tag rather than the first one:
 * JSX attributes routinely embed `>` inside braces (`onChange={(e) => ...}`),
 * and stopping at that arrow truncates the tag before its className, which made
 * an earlier version of this guard report compliant controls as violations.
 */
function openingTag(text, start) {
  let depth = 0;
  for (let i = start; i < text.length; i += 1) {
    const ch = text[i];
    if (ch === '{') depth += 1;
    else if (ch === '}') depth -= 1;
    else if (ch === '>' && depth === 0) return text.slice(start, i + 1);
  }
  return text.slice(start);
}

/** Opening tag of the nearest <label> still open at `index`, or null. */
function enclosingLabelTag(text, index) {
  let depth = 0;
  let i = index;
  while (i > 0) {
    const close = text.lastIndexOf('</label>', i);
    const open = text.lastIndexOf('<label', i);
    if (open === -1) return null;
    if (close > open) {
      depth += 1;
      i = close - 1;
      continue;
    }
    if (depth === 0) return openingTag(text, open);
    depth -= 1;
    i = open - 1;
  }
  return null;
}

const SELF_SIZED = [
  { re: /<select\b/g, what: '<select>' },
  { re: /<input\b[^>]*type="number"/g, what: '<input type="number">' },
  { re: /<input\b[^>]*type="text"/g, what: '<input type="text">' },
];
const LABEL_SIZED = [
  { re: /<input\b[^>]*type="checkbox"/g, what: '<input type="checkbox">' },
  { re: /<input\b[^>]*type="radio"/g, what: '<input type="radio">' },
];

/**
 * Blank out comment bodies, preserving length and newlines so byte offsets and
 * reported line numbers still line up. Several doc comments mention `<select>`
 * in prose, and matching those reported violations in files with no JSX at all.
 */
function stripComments(text) {
  return text
    .replace(/\/\*[\s\S]*?\*\//g, (m) => m.replace(/[^\n]/g, ' '))
    .replace(/\/\/[^\n]*/g, (m) => ' '.repeat(m.length));
}

const tapViolations = [];
for (const file of files) {
  const raw = await readFile(file, 'utf8');
  const text = stripComments(raw);
  const lineOf = (idx) => text.slice(0, idx).split('\n').length;

  for (const { re, what } of SELF_SIZED) {
    for (const m of text.matchAll(re)) {
      const start = what === '<select>' ? m.index : text.lastIndexOf('<input', m.index);
      if (!openingTag(text, start).includes(TAP_TARGET)) {
        tapViolations.push({
          file: relative(ROOT, file),
          line: lineOf(start),
          what,
          why: `must carry ${TAP_TARGET} itself`,
        });
      }
    }
  }

  for (const { re, what } of LABEL_SIZED) {
    for (const m of text.matchAll(re)) {
      const start = text.lastIndexOf('<input', m.index);
      if (openingTag(text, start).includes(TAP_TARGET)) continue;
      const label = enclosingLabelTag(text, start);
      if (!label?.includes(TAP_TARGET)) {
        tapViolations.push({
          file: relative(ROOT, file),
          line: lineOf(start),
          what,
          why: label
            ? `wrapping <label> must carry ${TAP_TARGET}`
            : `must be wrapped in a <label> carrying ${TAP_TARGET} (a sibling <label htmlFor> leaves the box itself undersized)`,
        });
      }
    }
  }
}

if (tapViolations.length > 0) {
  console.error('\nTap-target policy violations (DESIGN.md Interactive Element Minimum Size):\n');
  for (const v of tapViolations) {
    console.error(`  ${v.file}:${v.line}  ${v.what}  ${v.why}`);
  }
  console.error(`\n${tapViolations.length} violation(s). See DESIGN.md and issue #4368.`);
  process.exit(1);
}

console.log('tap-targets: OK (every checkbox/radio/select/number/text control hits a 44px target).');
