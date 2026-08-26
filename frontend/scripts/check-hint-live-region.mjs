#!/usr/bin/env bun
// Guard that every page rendering `t('hintAvailable')` does so inside an
// **always-mounted** live region.
//
// A hint that appears with no `role`/`aria-live` changes the screen silently:
// pressing `h` moves focus nowhere and announces nothing, so a screen-reader
// user has no way to know the hint arrived. That was the state of 67 pages
// when this guard was written (#5955, following #5596 / #5602).
//
// The subtler half is *where* the region goes. A live region that enters the
// DOM already holding its text is usually not announced -- assistive tech
// watches an existing region for changes -- so putting the attributes on the
// `{hint && (<div>…</div>)}` that only exists while a hint is set produces
// markup that passes an "aria-live is present" check and still says nothing.
// This guard therefore rejects the attributes on a conditional element and
// requires them on a wrapper that renders unconditionally.
//
// The walk is deliberately dumb: it reads the JSX as text, finds the element
// that encloses the `hintAvailable` render, and checks the enclosing chain. It
// cannot see a live region provided by a component (FortyThieves builds its
// own `sr-only` announcement element), so a page that announces some other way
// is listed in ANNOUNCES_ELSEWHERE with the element that does the work.

import { readdir, readFile } from 'node:fs/promises';
import { join } from 'node:path';
import { fileURLToPath } from 'node:url';
import { assertFloor } from './lib/floor.mjs';

// The pages directory, overridable by argv so the guard's own test can run it
// against fixtures instead of the real tree.
const PAGES = process.argv[2] ?? fileURLToPath(new URL('../src/pages', import.meta.url));

/** Pages that announce the hint through their own element rather than around `hintAvailable`. */
const ANNOUNCES_ELSEWHERE = new Map([
  ['FortyThievesPage.tsx', 'ft-hint-announcement'],
  ['RankAndFilePage.tsx', 'rf-hint-announcement'],
]);

/**
 * The opening tags enclosing `index`, innermost first.
 *
 * Walks backwards counting closing tags so a sibling that already closed does
 * not read as an ancestor.
 *
 * @param {string} src - The page source.
 * @param {number} index - Offset of the `hintAvailable` render.
 * @returns {string[]} Opening tags, innermost first.
 */
function enclosingTags(src, index) {
  const head = src.slice(0, index);
  const tokens = [...head.matchAll(/<\/?([A-Za-z][\w.]*)\b[^>]*>/g)];
  const open = [];
  for (const tk of tokens) {
    const text = tk[0];
    if (text.startsWith('</')) open.pop();
    else if (!text.endsWith('/>')) open.push(text);
  }
  return open.reverse();
}

/**
 * Whether the JSX between the live-region tag and the hint text is conditional.
 *
 * @param {string} between - Source between the region's tag and the hint render.
 * @returns {boolean} True when a `&&` or ternary gates the text.
 */
function isGatedBelow(between) {
  return /&&\s*\(?\s*$|\?\s*\(?\s*$/.test(between.replace(/\s*$/, '')) || /&&|\?/.test(between);
}

const problems = [];
let checked = 0;

for (const name of (await readdir(PAGES)).sort()) {
  if (!name.endsWith('Page.tsx')) continue;
  const src = await readFile(join(PAGES, name), 'utf8');
  const idx = src.indexOf("t('hintAvailable')");
  if (idx < 0) continue;
  checked += 1;

  const elsewhere = ANNOUNCES_ELSEWHERE.get(name);
  if (elsewhere) {
    if (!src.includes(elsewhere)) {
      problems.push(`${name}: listed as announcing through "${elsewhere}", but no such element exists`);
    }
    continue;
  }

  const chain = enclosingTags(src, idx);
  const region = chain.find((tag) => tag.includes('aria-live'));
  if (!region) {
    problems.push(`${name}: the hint text is in no live region — requesting a hint changes the screen silently`);
    continue;
  }
  if (!/role="status"|role="alert"/.test(region)) {
    problems.push(`${name}: the live region has aria-live but no role="status"`);
    continue;
  }
  // The region must not be the element that only exists while a hint is set.
  const between = src.slice(src.indexOf(region) + region.length, idx);
  const gatedIndex = chain.indexOf(region);
  const innerGate = chain.slice(0, gatedIndex).length > 0 && isGatedBelow(between);
  if (!innerGate) {
    problems.push(
      `${name}: the live region holds the hint text with nothing conditional below it — ` +
        'it enters the DOM already filled, which assistive tech does not announce. ' +
        'Move role/aria-live to the always-mounted wrapper.',
    );
  }
}

// 76 pages render hintAvailable today. A walk that finds a handful has lost the
// directory, not the feature.
assertFloor('hint-live-region', checked, 50, 'pages rendering hintAvailable');

if (problems.length > 0) {
  console.error('Hint text that is not announced:\n');
  for (const p of problems) console.error(`  ${p}`);
  console.error('\nWrap the conditional hint block in an always-mounted <div role="status" aria-live="polite">.');
  process.exit(1);
}

console.log(`hint-live-region: OK (${checked} pages render a hint, each inside a mounted live region).`);
