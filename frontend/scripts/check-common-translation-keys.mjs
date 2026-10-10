#!/usr/bin/env bun
/** Ensure literal tc() keys used by pages and components exist in both common locales. */
import { readdirSync, readFileSync } from 'node:fs';
import { join, relative } from 'node:path';
import { assertFloor } from './lib/floor.mjs';

const SRC = process.argv[2] ?? new URL('../src/', import.meta.url).pathname;
const LOCALES = join(SRC, 'i18n', 'locales');
const roots = ['pages', 'components'];
const files = roots.flatMap((root) => walk(join(SRC, root))).filter((file) => !/\.test\.tsx?$/.test(file));
const translations = Object.fromEntries(
  ['ja', 'en'].map((lang) => [lang, JSON.parse(readFileSync(join(LOCALES, lang, 'common.json'), 'utf8'))]),
);
const allowlist = new Map();

function walk(dir) {
  return readdirSync(dir, { withFileTypes: true }).flatMap((entry) => {
    const path = join(dir, entry.name);
    return entry.isDirectory() ? walk(path) : entry.isFile() && /\.tsx?$/.test(entry.name) ? [path] : [];
  });
}

function hasKey(tree, key) {
  if (Object.hasOwn(tree, key)) return true;
  return (
    key.split('.').reduce((node, part) => (node && typeof node === 'object' ? node[part] : undefined), tree) !==
    undefined
  );
}

assertFloor('common-translation-keys', files.length, 100, 'page and component source files scanned');
const problems = [];
for (const file of files) {
  const source = readFileSync(file, 'utf8');
  const pattern = /\btc\(\s*(['"])([^'"\n]+)\1\s*\)/g;
  for (const match of source.matchAll(pattern)) {
    const key = match[2];
    for (const lang of ['ja', 'en']) {
      if (!hasKey(translations[lang], key) && !allowlist.has(`${relative(SRC, file)}:${key}`)) {
        problems.push(`${relative(SRC, file)}: tc('${key}') missing from ${lang}/common.json`);
      }
    }
  }
}
if (problems.length) {
  console.error('common-translation-keys: unresolved tc() keys:');
  for (const problem of problems) console.error(`  ${problem}`);
  process.exit(1);
}
console.log(`common-translation-keys: OK (${files.length} page and component source files scanned).`);
