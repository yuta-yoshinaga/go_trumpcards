#!/usr/bin/env bun
import { readdir, readFile } from 'node:fs/promises';
import { dirname, join, relative } from 'node:path';
import { fileURLToPath } from 'node:url';
import { assertFloor } from './lib/floor.mjs';

const SCRIPT_DIR = dirname(fileURLToPath(import.meta.url));
const SRC = join(SCRIPT_DIR, '../src');
const CEILING = 74; // Exact count: lower it (never raise it) when vacuous assertions are fixed.
const VACUOUS = /(?<!\.not\.)\btoHaveTextContent\(\s*(['"]{2})\s*\)/g;

async function* testFiles(dir) {
  for (const entry of await readdir(dir, { withFileTypes: true })) {
    const path = join(dir, entry.name);
    if (entry.isDirectory()) yield* testFiles(path);
    else if (/\.test\.tsx?$/.test(entry.name)) yield path;
  }
}

export function countVacuousAssertions(source) {
  return [...source.matchAll(VACUOUS)].length;
}

function vacuousOccurrences(source) {
  return [...source.matchAll(VACUOUS)].map((match) => match.index);
}

const violations = [];
let files = 0;
for await (const file of testFiles(SRC)) {
  files++;
  const source = await readFile(file, 'utf8');
  for (const index of vacuousOccurrences(source)) {
    violations.push(`${relative(SRC, file).split('\\').join('/')}:${source.slice(0, index).split('\n').length}`);
  }
}
assertFloor('vacuous-text-content', files, 1300, 'test files scanned');
if (violations.length > CEILING) {
  console.error(`vacuous-text-content: ${violations.length} occurrences exceeds ceiling ${CEILING}:`);
  for (const violation of violations) console.error(`  ${violation}`);
  process.exit(1);
}
if (violations.length < CEILING) {
  console.error(
    `vacuous-text-content: ${violations.length} occurrences is below the ceiling ${CEILING}; lower CEILING to ${violations.length} so the slack cannot be refilled.`,
  );
  process.exit(1);
}
console.log(`vacuous-text-content: OK (${violations.length}/${CEILING} occurrences; ${files} test files scanned).`);
