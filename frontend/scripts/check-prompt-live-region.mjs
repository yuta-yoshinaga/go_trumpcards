#!/usr/bin/env bun
// Guard that every bid/discard prompt renders inside an **always-mounted**
// live region.
//
// The prompt is text that appears when the phase changes. Without
// `role="status"` / `aria-live` nothing reaches a screen reader, so the turn
// passes to the player silently. Twelve pages were in that state when this
// guard was written (#6880, following the same shape #6484 established for
// Calabresella).
//
// The subtle half is the same one `check-hint-live-region.mjs` documents: a
// region that enters the DOM already holding its text is generally not
// announced, because assistive tech watches an *existing* region for changes.
// Putting the attributes on the `{canBid && (<div>…</div>)}` that only exists
// during the bid phase therefore passes an "aria-live is present" check while
// still announcing nothing. This guard rejects that and requires the region to
// be mounted unconditionally, with only its contents swapping.
//
// It also rejects a region that merely *sits beside* the prompt: the prompt has
// to be a descendant, which is what `toContainElement` asserts in the page
// tests.

import { readdir, readFile } from 'node:fs/promises';
import { join } from 'node:path';
import { fileURLToPath } from 'node:url';
import { assertFloor } from './lib/floor.mjs';

const PAGES = process.argv[2] ?? fileURLToPath(new URL('../src/pages', import.meta.url));

// **Match on the shape, never on a list of kinds.** Tysiac's talon prompt was
// left outside its region by the very commit that added the region, and went
// unnoticed because a draft of this guard only looked at `-bid-`/`-discard-`
// names. A kind list has to be extended by the same person who forgets; every
// `*prompt` testid is in scope instead (#7042 review).
const PROMPT_RE = /data-testid="([a-z0-9]+-[a-z0-9-]*prompt)"/g;

/** Attributes of the element carrying a given `data-testid`. */
const OWN_TAG_RE = (tid) => new RegExp(`<(?:div|span)\\b([^>]*?\\bdata-testid="${tid}"[^>]*?)>`);

/** The permanent wrapper each page must provide. `span` for prompts that sit in a flex row. */
const REGION_RE = /<(div|span)\s+data-testid="([a-z0-9]+-prompt-live)"([^>]*)>/;

/**
 * Whether the element opened at `at` is gated by a conditional.
 *
 * @param {string} src - Page source.
 * @param {number} at - Offset of the opening tag.
 * @returns {boolean} True when `&&` or `?` immediately precedes it.
 */
function isConditionallyMounted(src, at) {
  return /(?:&&|\?)\s*\(?\s*$/.test(src.slice(Math.max(0, at - 80), at));
}

/**
 * Offset just past the element opened at `open`, matching nested tags.
 *
 * **A self-closing region ends at itself.** Written as
 * `<div data-testid="x-prompt-live" role="status" aria-live="polite" />` the
 * open tag contributes no depth, so the general loop below would walk on and
 * return the close of the *next* sibling -- reporting a prompt that merely
 * follows the region as being inside it.
 *
 * @param {string} src - Page source.
 * @param {number} open - Offset of the region's opening tag.
 * @param {string} tag - Element name of the region (`div` or `span`).
 * @returns {number} Offset of the matching close, or -1.
 */
function closeOf(src, open, tag) {
  const tags = [...src.slice(open).matchAll(new RegExp(`<\\/?${tag}\\b[^>]*>`, 'g'))];
  if (tags.length > 0 && tags[0].index === 0 && tags[0][0].endsWith('/>')) {
    return open + tags[0][0].length;
  }
  let depth = 0;
  for (const t of tags) {
    if (t[0].startsWith('</')) {
      depth -= 1;
      if (depth === 0) return open + t.index + t[0].length;
    } else if (!t[0].endsWith('/>')) depth += 1;
  }
  return -1;
}

const problems = [];
let pages = 0;
let prompts = 0;

for (const name of (await readdir(PAGES)).sort()) {
  if (!name.endsWith('Page.tsx')) continue;
  const src = await readFile(join(PAGES, name), 'utf8');
  const found = [...src.matchAll(PROMPT_RE)];
  if (found.length === 0) continue;
  pages += 1;
  prompts += found.length;

  // **A prompt can be its own region.** MonteCarlo/FourteenOut render the
  // prompt unconditionally with `role="status"` on the prompt element itself
  // and merely swap its text, which is exactly the behaviour this guard wants.
  // Demanding a second wrapper around it would be cargo cult.
  const selfLive = new Set();
  for (const m of found) {
    const own = OWN_TAG_RE(m[1]).exec(src);
    if (own === null) continue;
    if (!/role="status"/.test(own[1]) || !/aria-live=/.test(own[1])) continue;
    if (isConditionallyMounted(src, own.index)) continue;
    selfLive.add(m[1]);
  }
  const outstanding = found.filter((m) => !selfLive.has(m[1]));
  if (outstanding.length === 0) continue;

  const region = REGION_RE.exec(src);
  if (region === null) {
    problems.push(`${name}: ${outstanding.map((m) => m[1]).join(', ')} has no *-prompt-live region`);
    continue;
  }
  const attrs = region[3];
  if (!/role="status"/.test(attrs) || !/aria-live=/.test(attrs)) {
    problems.push(`${name}: ${region[2]} needs both role="status" and aria-live`);
    continue;
  }
  if (isConditionallyMounted(src, region.index)) {
    problems.push(
      `${name}: ${region[2]} is mounted conditionally. A region that appears together with its text is not announced; mount it always and swap the contents.`,
    );
    continue;
  }
  const end = closeOf(src, region.index, region[1]);
  for (const m of outstanding) {
    if (m.index < region.index || m.index > end) {
      problems.push(`${name}: ${m[1]} is outside ${region[2]}, not inside it`);
    }
  }
}

assertFloor('prompt-live-region', pages, 24, 'pages with a phase-change prompt');
assertFloor('prompt-live-region', prompts, 34, 'phase-change prompts scanned');

if (problems.length > 0) {
  console.error('prompt-live-region: NG');
  for (const p of problems) console.error(`  ${p}`);
  process.exit(1);
}
console.log(`prompt-live-region: OK (${prompts} prompts across ${pages} pages).`);
