#!/usr/bin/env bun
// Guard that every messageCode the Go presenters emit has a ja AND en
// translation. See issue #4365.
//
// This is a *vertical* check, backend to locale, and nothing covered it before.
// The existing i18n parity check is horizontal — it compares ja against en, and
// it passes today with 654 keys on each side. A code that Go emits and neither
// locale defines is symmetric, so parity sees nothing wrong.
//
// Nor does anything else. Go tests assert that a presenter returns a particular
// code, never that the code can be rendered. Frontend tests exercise
// GameMessageBox with codes that already exist. And GameMessageBox falls back to
// the backend's raw `message` when a lookup misses, which turns a missing
// translation into silent degradation rather than an error:
//
//   - message == ""  -> displayMessage is empty -> the box returns null and the
//                       player is told nothing at all
//   - message != ""  -> the raw Go literal renders, in whichever language it
//                       happens to be written, in both locales
//
// 193 of 569 codes were in one of those two states when this guard was written.

import { readdir, readFile } from 'node:fs/promises';
import { join, relative } from 'node:path';
import { fileURLToPath } from 'node:url';

const FRONTEND = fileURLToPath(new URL('..', import.meta.url));
const REPO = join(FRONTEND, '..');
const PRESENTER_DIR = join(REPO, 'internal/adapter/presenter');
const LOCALES = join(FRONTEND, 'src/i18n/locales');

/**
 * Emitted codes, as code -> the message literals returned alongside it.
 *
 * Presenters return `(message, messageCode, params)`. An empty message means the
 * frontend has nothing to fall back on, which is the worse of the two failure
 * modes, so the two are reported separately.
 */
async function emittedCodes() {
  const out = new Map();
  const files = (await readdir(PRESENTER_DIR)).filter((f) => f.endsWith('WebPresenter.go'));
  for (const name of files) {
    const text = await readFile(join(PRESENTER_DIR, name), 'utf8');
    for (const m of text.matchAll(/return\s+("(?:[^"\\]|\\.)*"),\s*"([a-zA-Z0-9_.]+)"/g)) {
      const [, literal, code] = m;
      if (!out.has(code)) out.set(code, { literals: new Set(), files: new Set() });
      out.get(code).literals.add(literal);
      out.get(code).files.add(name);
    }
  }
  return out;
}

const emitted = await emittedCodes();
const ja = JSON.parse(await readFile(join(LOCALES, 'ja/common.json'), 'utf8')).messageCode ?? {};
const en = JSON.parse(await readFile(join(LOCALES, 'en/common.json'), 'utf8')).messageCode ?? {};

const missing = [];
for (const [code, info] of emitted) {
  const gaps = [];
  if (!(code in ja)) gaps.push('ja');
  if (!(code in en)) gaps.push('en');
  if (gaps.length === 0) continue;
  const blank = [...info.literals].every((l) => l === '""');
  missing.push({ code, gaps, blank, file: [...info.files][0] });
}

if (missing.length > 0) {
  const blanks = missing.filter((m) => m.blank);
  const literals = missing.filter((m) => !m.blank);
  console.error('\nUntranslated messageCodes (emitted by Go, absent from the locale files):\n');
  if (blanks.length > 0) {
    console.error(`  ${blanks.length} render as an EMPTY message box (no fallback text at all):`);
    for (const m of blanks) console.error(`    ${m.code}  [${m.gaps.join(' + ')}]  ${m.file}`);
  }
  if (literals.length > 0) {
    console.error(`\n  ${literals.length} render the raw Go literal, in both locales:`);
    for (const m of literals) console.error(`    ${m.code}  [${m.gaps.join(' + ')}]  ${m.file}`);
  }
  console.error(
    `\n${missing.length} untranslated code(s) of ${emitted.size} emitted.` +
      ` Add them under "messageCode" in ${relative(REPO, LOCALES)}/{ja,en}/common.json.` +
      ' See issue #4365.',
  );
  process.exit(1);
}

console.log(`message-codes: OK (all ${emitted.size} codes emitted by Go have ja + en translations).`);
