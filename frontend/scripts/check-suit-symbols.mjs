#!/usr/bin/env bun
// Prevent local SUIT_SYMBOLS tables from returning after #8022.

import { readdir, readFile } from 'node:fs/promises';
import { join, relative, resolve } from 'node:path';
import { fileURLToPath } from 'node:url';
import { assertFloor } from './lib/floor.mjs';

const FRONTEND = fileURLToPath(new URL('..', import.meta.url));
const ROOT = process.argv[2] ? resolve(process.argv[2]) : join(FRONTEND, 'src');
const SCANNING_SRC = !process.argv[2];
const SKIP = new Set(['node_modules', '.git', 'dist', 'coverage']);
const SOURCE = /\.tsx?$/;
const TEST = /\.test\.[^.]+$/;
const DECLARATION = /\b(?:const|let)\s+SUIT_SYMBOLS\b/g;

async function* sourceFiles(dir) {
  for (const entry of await readdir(dir, { withFileTypes: true })) {
    if (SKIP.has(entry.name)) continue;
    const full = join(dir, entry.name);
    if (entry.isDirectory()) yield* sourceFiles(full);
    else if (SOURCE.test(entry.name) && !TEST.test(entry.name)) yield full;
  }
}

let files = 0;
const failures = [];
for await (const file of sourceFiles(ROOT)) {
  files++;
  const source = await readFile(file, 'utf8');
  const rel = relative(ROOT, file).split('\\').join('/');
  for (const match of source.matchAll(DECLARATION)) {
    const line = source.slice(0, match.index).split('\n').length;
    failures.push(`${rel}:${line}`);
  }
}

if (files === 0) {
  console.error('suit-symbols: scanned 0 source files; check the scan root.');
  process.exit(1);
}
if (SCANNING_SRC) assertFloor('suit-symbols', files, 2000, 'source files');
if (failures.length > 0) {
  console.error(`suit-symbols: ${failures.length} local definition(s) found:`);
  for (const failure of failures) console.error(`  ${failure}`);
  console.error('Use suitSymbolAt(suit, fallback) / suitSymbol(design) (#8022).');
  process.exit(1);
}

console.log(`suit-symbols: OK (${files} source files scanned).`);
