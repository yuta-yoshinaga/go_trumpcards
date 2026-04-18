#!/usr/bin/env bun
// Guard against design-token regressions. Fails the build if source code under
// frontend/src uses forbidden raw Tailwind palette classes or opacity-suffixed
// `text-white/N`. See DESIGN.md and issue #1411.

import { readdir, readFile } from 'node:fs/promises';
import { join, relative } from 'node:path';

const ROOT = new URL('..', import.meta.url).pathname;
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

console.log(`design-tokens: OK (${files.length} source files scanned, test files skipped).`);
