#!/usr/bin/env bun
// Keep new game pages on the ADR-0044 createGamePage path and ratchet existing pages forward.

import { readdir, readFile } from 'node:fs/promises';
import { fileURLToPath } from 'node:url';
import { assertFloor } from './lib/floor.mjs';

const SCRIPT_DIR = fileURLToPath(new URL('.', import.meta.url));
const PAGES_DIR = process.env.GAME_PAGE_DEFINITION_PAGES ?? `${SCRIPT_DIR}../src/pages`;
const BASELINE_PATH = process.env.GAME_PAGE_DEFINITION_BASELINE ?? `${SCRIPT_DIR}game-page-definition-baseline.txt`;
const ROUTES_PATH = process.env.GAME_PAGE_DEFINITION_ROUTES ?? `${SCRIPT_DIR}../src/constants/gameRoutes.ts`;

let baselineText;
try {
  baselineText = await readFile(BASELINE_PATH, 'utf8');
} catch {
  console.error(`game-page-definition: could not read baseline at ${BASELINE_PATH}`);
  process.exit(1);
}
const entries = baselineText.split(/\r?\n/).filter((line) => line.length > 0 && !line.startsWith('#'));
const sorted = [...entries].sort();
if (entries.some((entry, index) => entry !== sorted[index]) || new Set(entries).size !== entries.length) {
  console.error('game-page-definition: baseline not sorted or contains duplicate lines');
  process.exit(1);
}

let pages;
try {
  pages = (await readdir(PAGES_DIR)).filter((file) => file.endsWith('Page.tsx')).sort();
} catch {
  console.error(`game-page-definition: could not read pages directory ${PAGES_DIR}`);
  process.exit(1);
}
let routesSource;
try {
  routesSource = await readFile(ROUTES_PATH, 'utf8');
} catch {
  console.error(`game-page-definition: could not read routes at ${ROUTES_PATH}`);
  process.exit(1);
}
const routedPages = new Set([...routesSource.matchAll(/\bpage\s*:\s*['"]([^'"]+)['"]/g)].map((match) => match[1]));
const baseline = new Set(entries);
// The rule the docs state: the page's export is built by the factory. A bare
// `createGamePage(` substring would also match a comment or an unused call.
const isMigrated = (source) => /^export const \w+Page = createGamePage\(/m.test(source);
let failed = false;
const wrappers = new Map();
for (const page of pages) {
  const source = await readFile(`${PAGES_DIR}/${page}`, 'utf8');
  const importMatch = source.match(/^import\s+(?:[\s\S]*?)\s+from\s+['"]\.\/([\w]+Page)['"];?\s*$/m);
  if (importMatch) wrappers.set(page, importMatch[1]);
  if (importMatch && /\bGamePageShell\b/.test(source)) {
    console.error(
      `game-page-definition: ${page}: wrapper of ${importMatch[1]} imports GamePageShell — account for it as a real page`,
    );
    failed = true;
  }
}
const nonRoutedCount = pages.filter((page) => !routedPages.has(page.replace(/Page\.tsx$/, ''))).length;
const wrapperCount = wrappers.size;
for (const entry of entries) {
  if (!pages.includes(entry)) {
    console.error(`game-page-definition: ${entry}: stale baseline entry`);
    failed = true;
  }
  const wrapperTarget = wrappers.get(entry);
  if (wrapperTarget || !routedPages.has(entry.replace(/Page\.tsx$/, ''))) {
    const reason = wrapperTarget ? `variant wrapper of ${wrapperTarget}` : 'not a game page';
    console.error(`game-page-definition: ${entry}: ${reason} — remove from baseline`);
    failed = true;
  }
}
let migrated = 0;
for (const page of pages) {
  const source = await readFile(`${PAGES_DIR}/${page}`, 'utf8');
  if (wrappers.has(page) || !routedPages.has(page.replace(/Page\.tsx$/, ''))) continue;
  if (baseline.has(page)) {
    if (isMigrated(source)) {
      console.error(`game-page-definition: ${page}: migrated — remove it from the baseline`);
      failed = true;
    }
  } else if (!isMigrated(source)) {
    console.error(
      `game-page-definition: ${page}: new page must be built with createGamePage (ADR-0044); clone from a migrated page (AcesUpPage / HeartsPage) or migrate the clone source first. Do not add it to the baseline`,
    );
    failed = true;
  } else {
    migrated += 1;
  }
}
if (failed) process.exit(1);
// The fixture override is a test seam only; CI never sets it, so the real floor applies there.
if (process.env.GAME_PAGE_DEFINITION_PAGES) {
  assertFloor('game-page-definition', pages.length, 1, 'fixture game pages');
} else {
  assertFloor('game-page-definition', pages.length, 380, 'game pages');
  assertFloor('game-page-definition', nonRoutedCount, 4, 'non-routed pages');
  assertFloor('game-page-definition', wrapperCount, 8, 'variant wrappers');
}
assertFloor('game-page-definition', entries.length, 1, 'baseline pages');
console.log(
  `game-page-definition: OK (${migrated} migrated, ${entries.length} in baseline, ${pages.length} pages, ${nonRoutedCount} non-routed, ${wrapperCount} wrappers)`,
);
