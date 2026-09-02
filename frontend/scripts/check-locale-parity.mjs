#!/usr/bin/env bun
/**
 * Guard: every ja locale file has an en twin with the same keys, and vice versa.
 *
 * Without this, a missing en key is not a blank string — `fallbackLng: 'ja'`
 * means an English player gets the Japanese sentence dropped into an otherwise
 * English screen. Nothing else in `bun run check` compares the two trees, and a
 * PR shipped exactly that gap (#6549).
 *
 * i18next plural suffixes (`key_one` / `key_other` / …) are folded onto their
 * base key, because a language needs only the forms its plural rule uses: ja has
 * one form, en has two.
 */
import { readdirSync, readFileSync } from 'node:fs';
import { join } from 'node:path';
import { assertFloor } from './lib/floor.mjs';

// argv[2] lets the guard's own test point it at a fixture tree, so the test can
// spawn the real script rather than re-implement its logic.
const FIXTURE_DIR = process.argv[2];
const LOCALES = FIXTURE_DIR ?? new URL('../src/i18n/locales/', import.meta.url).pathname;
const PLURAL_SUFFIX = /_(zero|one|two|few|many|other)$/;

/** Flatten a translation tree to its leaf key paths, folding plural suffixes. */
function leafKeys(node, prefix = '', out = new Set()) {
  if (node !== null && typeof node === 'object' && !Array.isArray(node)) {
    for (const [k, v] of Object.entries(node)) {
      leafKeys(v, prefix ? `${prefix}.${k}` : k, out);
    }
    return out;
  }
  out.add(prefix.replace(PLURAL_SUFFIX, ''));
  return out;
}

const read = (lang, file) => JSON.parse(readFileSync(join(LOCALES, lang, file), 'utf8'));
const list = (lang) => readdirSync(join(LOCALES, lang)).filter((f) => f.endsWith('.json'));

const ja = list('ja');
const en = list('en');
const problems = [];

for (const file of [...new Set([...ja, ...en])].sort()) {
  if (!ja.includes(file)) {
    problems.push(`${file}: exists in en but not ja`);
    continue;
  }
  if (!en.includes(file)) {
    problems.push(`${file}: exists in ja but not en`);
    continue;
  }
  const kja = leafKeys(read('ja', file));
  const ken = leafKeys(read('en', file));
  const jaOnly = [...kja].filter((k) => !ken.has(k)).sort();
  const enOnly = [...ken].filter((k) => !kja.has(k)).sort();
  if (jaOnly.length > 0) problems.push(`${file}: missing from en -> ${jaOnly.join(', ')}`);
  if (enOnly.length > 0) problems.push(`${file}: missing from ja -> ${enOnly.join(', ')}`);
}

if (problems.length > 0) {
  console.error('locale-parity: ja and en disagree.\n');
  for (const p of problems) console.error(`  ${p}`);
  console.error('\nfallbackLng is "ja", so a key only ja has renders Japanese text inside the English UI.');
  process.exit(1);
}

if (FIXTURE_DIR) {
  // A fixture tree holds a handful of files; the real floor would reject it.
  assertFloor('locale-parity', ja.length, 1, 'ja locale files compared');
} else {
  assertFloor('locale-parity', ja.length, 250, 'ja locale files compared');
}
console.log(`locale-parity: OK (${ja.length} locale file pairs, keys match).`);
