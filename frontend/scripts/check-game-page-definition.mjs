#!/usr/bin/env bun
// Keep new game pages on the ADR-0044 createGamePage path and ratchet existing pages forward.

import { readdir, readFile } from 'node:fs/promises';
import { fileURLToPath } from 'node:url';
import { assertFloor } from './lib/floor.mjs';

const SCRIPT_DIR = fileURLToPath(new URL('.', import.meta.url));
const PAGES_DIR = process.env.GAME_PAGE_DEFINITION_PAGES ?? `${SCRIPT_DIR}../src/pages`;
const BASELINE_PATH = process.env.GAME_PAGE_DEFINITION_BASELINE ?? `${SCRIPT_DIR}game-page-definition-baseline.txt`;

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
const baseline = new Set(entries);
let failed = false;
for (const entry of entries) {
  if (!pages.includes(entry)) {
    console.error(`game-page-definition: ${entry}: stale baseline entry`);
    failed = true;
  }
}
let migrated = 0;
for (const page of pages) {
  const source = await readFile(`${PAGES_DIR}/${page}`, 'utf8');
  if (baseline.has(page)) {
    if (source.includes('createGamePage(')) {
      console.error(`game-page-definition: ${page}: migrated — remove it from the baseline`);
      failed = true;
    }
  } else if (!source.includes('createGamePage(')) {
    console.error(
      `game-page-definition: ${page}: new page must be built with createGamePage (ADR-0044); clone from a migrated page (AcesUpPage / HeartsPage) or migrate the clone source first. Do not add it to the baseline`,
    );
    failed = true;
  } else {
    migrated += 1;
  }
}
if (failed) process.exit(1);
if (process.env.GAME_PAGE_DEFINITION_PAGES) {
  assertFloor('game-page-definition', pages.length, 1, 'fixture game pages');
} else {
  assertFloor('game-page-definition', pages.length, 380, 'game pages');
}
assertFloor('game-page-definition', entries.length, 1, 'baseline pages');
console.log(`game-page-definition: OK (${migrated} migrated, ${entries.length} in baseline, ${pages.length} pages)`);
