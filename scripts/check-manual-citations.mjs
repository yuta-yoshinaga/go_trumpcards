#!/usr/bin/env bun
/**
 * Verifies that every external source cited by a game manual still resolves.
 *
 * Deliberately NOT part of `bun run check`: it makes real network requests, so
 * wiring it into the commit gate would make an offline machine or a slow
 * upstream fail a build for reasons unrelated to the change being made. Run it
 * on demand — before a documentation pass, or alongside the annual trademark
 * sweep in TRADEMARKS.md.
 *
 * Why it exists: three cited pagat.com pages were returning 404 (guandan,
 * bideuchre, shengji). A citation that does not resolve is worse than no
 * citation, because this repository leans on those links as the evidence that
 * its manuals summarise the rules rather than translate someone's prose.
 *
 * Usage:
 *   bun scripts/check-manual-citations.mjs
 */

import { readdirSync, readFileSync } from 'node:fs';
import { join } from 'node:path';

/** Repository root — this file lives at scripts/. */
const REPO_ROOT = new URL('../', import.meta.url).pathname;
const MANUAL_DIR = join(REPO_ROOT, 'docs/manual');

/** Hosts we cite as rule sources and therefore care about. */
const SOURCE_HOSTS = /pagat\.com|wikipedia\.org|gambiter\.com/;

/**
 * Sanity floor. Finding no citations at all would mean the walk or the regex
 * broke, which reads identically to "every citation is fine".
 */
const MIN_CITATIONS = 20;

/** Requests in flight at once — enough to be quick, few enough to be polite. */
const CONCURRENCY = 5;

/**
 * Pulls markdown link targets out of a document. The URL pattern allows
 * balanced parentheses because Wikipedia titles contain them —
 * `Boston_(card_game)` was truncated by a naive `[^)]*` and reported as a
 * false 404.
 */
function citationsIn(text) {
  return [...text.matchAll(/\]\((https?:\/\/(?:[^()\s]|\([^()\s]*\))+)\)/g)]
    .map((m) => m[1])
    .filter((u) => SOURCE_HOSTS.test(u));
}

const cited = new Map();
let kinds;
try {
  kinds = readdirSync(MANUAL_DIR);
} catch {
  console.error(`check-manual-citations: cannot read ${MANUAL_DIR} — the layout moved and this check is scanning nothing.`);
  process.exit(1);
}
for (const kind of kinds) {
  const dir = join(MANUAL_DIR, kind);
  let entries;
  try {
    entries = readdirSync(dir);
  } catch {
    continue; // not a directory
  }
  for (const file of entries.filter((f) => f.endsWith('.md'))) {
    for (const url of citationsIn(readFileSync(join(dir, file), 'utf8'))) {
      if (!cited.has(url)) cited.set(url, []);
      cited.get(url).push(`${kind}/${file}`);
    }
  }
}

if (cited.size < MIN_CITATIONS) {
  console.error(
    `check-manual-citations: only ${cited.size} citations found (expected >= ${MIN_CITATIONS}).\n` +
      'The walk or the link pattern is broken, not the manuals uncited.',
  );
  process.exit(1);
}

const urls = [...cited.keys()].sort();
const broken = [];

async function head(url) {
  try {
    const res = await fetch(url, { redirect: 'follow', headers: { 'User-Agent': 'go_trumpcards-citation-check' } });
    return res.status;
  } catch (err) {
    return `ERROR ${String(err).slice(0, 60)}`;
  }
}

for (let i = 0; i < urls.length; i += CONCURRENCY) {
  const batch = urls.slice(i, i + CONCURRENCY);
  const results = await Promise.all(batch.map(head));
  batch.forEach((url, j) => {
    if (results[j] !== 200) broken.push({ url, status: results[j], files: cited.get(url) });
  });
}

if (broken.length > 0) {
  console.error(`check-manual-citations: ${broken.length} of ${urls.length} cited source(s) do not resolve:`);
  for (const b of broken) {
    console.error(`  [${b.status}] ${b.url}`);
    console.error(`          cited by ${b.files.join(', ')}`);
  }
  console.error('\nFind where the page moved and update the link, or cite a stable index page instead.');
  process.exit(1);
}

console.log(`manual-citations: OK (${urls.length} cited sources across ${cited.size} URLs, all resolve).`);
