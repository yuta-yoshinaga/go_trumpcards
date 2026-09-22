#!/usr/bin/env bun
// Guard against player-facing English labels embedded in Go format literals.

import { readdir } from 'node:fs/promises';
import path from 'node:path';
import { fileURLToPath } from 'node:url';
import { assertFloor } from './lib/floor.mjs';

const HERE = path.dirname(fileURLToPath(import.meta.url));
const ROOT = process.argv[2] ? path.resolve(process.argv[2]) : path.resolve(HERE, '..', '..');
const SCAN_ROOT = path.join(ROOT, 'internal', 'adapter', 'presenter');
const SCANNED_FILE_FLOOR = 250;
// biome-ignore lint/complexity/noUselessEscapeInRegex: Keep the guard pattern identical to the documented rule.
const LABEL_PATTERN = /(?:^|[\s(|\[,])([A-Za-z][A-Za-z0-9]{1,})\s*[:=]\s*%[-+ #0-9.]*[vdsqtfxb]/;

async function goFiles(dir) {
  const files = [];
  for (const entry of await readdir(dir, { withFileTypes: true })) {
    const full = path.join(dir, entry.name);
    if (entry.isDirectory()) files.push(...(await goFiles(full)));
    else if (entry.isFile() && entry.name.endsWith('.go') && !entry.name.endsWith('_test.go')) files.push(full);
  }
  return files;
}

function stringLiterals(source) {
  return source.match(/"(?:\\.|[^"\\])*"|`[^`]*`/g) ?? [];
}

let files;
try {
  files = await goFiles(SCAN_ROOT);
} catch (error) {
  if (error?.code !== 'ENOENT') throw error;
  files = [];
}

if (!process.argv[2]) assertFloor('go-literal-labels', files.length, SCANNED_FILE_FLOOR, 'presenter Go files scanned');

const violations = [];
for (const file of files) {
  const source = await Bun.file(file).text();
  for (const literal of stringLiterals(source)) {
    const match = literal.slice(1, -1).match(LABEL_PATTERN);
    if (!match) continue;
    const line = source.slice(0, source.indexOf(literal)).split('\n').length;
    violations.push(`${path.relative(ROOT, file)}:${line}: ${match[0].trim()}`);
  }
}

if (violations.length > 0) {
  console.error(`go-literal-labels: ${violations.length} English label(s) in format literals found:`);
  for (const violation of violations) console.error(`  ${violation}`);
  process.exit(1);
}

console.log(`go-literal-labels: OK (0 English labels; ${files.length} presenter Go files scanned).`);
