#!/usr/bin/env bun
// Keep solitaire move targets focusable before a source card is selected (#8524).

import { readdir, readFile } from 'node:fs/promises';
import { join, relative, resolve } from 'node:path';
import { fileURLToPath } from 'node:url';
import { assertFloor } from './lib/floor.mjs';

const FRONTEND = fileURLToPath(new URL('..', import.meta.url));
const ROOT = process.argv[2] ? resolve(process.argv[2]) : join(FRONTEND, 'src');
const SCANNING_SRC = !process.argv[2];
const TARGET_DIRS = ['pages', 'components'];
const SKIP = new Set(['node_modules', '.git', 'dist', 'coverage']);
const TEST = /\.test\.[^.]+$/;
const BAD_TERMS = new Set(['!selectedSource', 'selectedSource === null', '!isTarget']);

async function* sourceFiles(dir) {
  let entries;
  try {
    entries = await readdir(dir, { withFileTypes: true });
  } catch (error) {
    if (error.code === 'ENOENT') return;
    throw error;
  }
  for (const entry of entries) {
    if (SKIP.has(entry.name)) continue;
    const full = join(dir, entry.name);
    if (entry.isDirectory()) yield* sourceFiles(full);
    else if (entry.name.endsWith('.tsx') && !TEST.test(entry.name)) yield full;
  }
}

function expressionEnd(source, open) {
  let depth = 1;
  for (let i = open + 1; i < source.length; i++) {
    if (source[i] === '{') depth++;
    else if (source[i] === '}' && --depth === 0) return i;
  }
  return -1;
}

function topLevelOrTerms(expression) {
  const terms = [];
  let start = 0;
  let round = 0;
  let square = 0;
  let curly = 0;
  for (let i = 0; i < expression.length; i++) {
    const ch = expression[i];
    if (ch === '(') round++;
    else if (ch === ')') round--;
    else if (ch === '[') square++;
    else if (ch === ']') square--;
    else if (ch === '{') curly++;
    else if (ch === '}') curly--;
    else if (ch === '|' && expression[i + 1] === '|' && round === 0 && square === 0 && curly === 0) {
      terms.push(expression.slice(start, i).trim());
      start = i + 2;
      i++;
    }
  }
  terms.push(expression.slice(start).trim());
  return terms;
}

const files = [];
for (const dir of TARGET_DIRS) {
  for await (const file of sourceFiles(join(ROOT, dir))) files.push(file);
}
files.sort();

if (files.length === 0) {
  console.error('move-target-focusable: scanned 0 files; check the scan root.');
  process.exit(1);
}
if (SCANNING_SRC) assertFloor('move-target-focusable', files.length, 344, 'source files');

const failures = [];
let expressions = 0;
for (const file of files) {
  const source = await readFile(file, 'utf8');
  const rel = relative(ROOT, file).split('\\').join('/');
  for (const match of source.matchAll(/(?:^|\s)disabled\s*=\s*\{/gm)) {
    const open = match.index + match[0].lastIndexOf('{');
    const end = expressionEnd(source, open);
    if (end < 0) continue;
    expressions++;
    const expression = source.slice(open + 1, end);
    const bad = topLevelOrTerms(expression).some((term) => BAD_TERMS.has(term));
    if (bad) {
      const line = source.slice(0, match.index).split('\n').length;
      failures.push(`${rel}:${line}: disabled={${expression.trim()}}`);
    }
  }
}

if (SCANNING_SRC) assertFloor('move-target-focusable', expressions, 1794, 'disabled expressions');
if (failures.length > 0) {
  for (const failure of failures) console.error(failure);
  console.error(
    'Move the selection term into aria-disabled and point aria-describedby at a label.selectSourceFirst hint, as in #8524.',
  );
  process.exit(1);
}

console.log(`move-target-focusable: OK (${files.length} files scanned, ${expressions} disabled expressions checked).`);
