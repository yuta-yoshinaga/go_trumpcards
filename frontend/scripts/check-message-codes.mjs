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
/**
 * Split a return statement's arguments at top-level commas, respecting nesting
 * and string literals.
 *
 * A regex cannot do this. The first version of this guard matched the message
 * argument as a quoted literal, so it never even looked at
 * `return fmt.Sprintf("…%d…", n), "code", params`. Widening it to `[^,\n]+`
 * did not help either, because a Sprintf contains its own commas. Eleven codes
 * fell through that hole, and since an unmatched return is simply not collected,
 * the guard printed `OK (all 569 codes…)` while they stayed broken — the worst
 * thing a guard can do, which is to be confidently silent.
 */
function splitArgs(text, start) {
  const args = [];
  let depth = 0;
  let quote = null;
  let cur = '';
  for (let i = start; i < text.length; i += 1) {
    const c = text[i];
    if (quote) {
      cur += c;
      if (c === '\\') {
        cur += text[i + 1] ?? '';
        i += 1;
      } else if (c === quote) quote = null;
      continue;
    }
    if (c === '"' || c === '`') {
      quote = c;
      cur += c;
      continue;
    }
    if ('([{'.includes(c)) depth += 1;
    else if (')]}'.includes(c)) {
      if (depth === 0) break; // the enclosing func's closing brace
      depth -= 1;
    } else if (c === '\n' && depth === 0) {
      break;
    } else if (c === ',' && depth === 0) {
      args.push(cur.trim());
      cur = '';
      continue;
    }
    cur += c;
  }
  if (cur.trim()) args.push(cur.trim());
  return args;
}

async function emittedCodes() {
  const out = new Map();
  const files = (await readdir(PRESENTER_DIR)).filter((f) => f.endsWith('WebPresenter.go'));
  for (const name of files) {
    const text = await readFile(join(PRESENTER_DIR, name), 'utf8');
    for (const m of text.matchAll(/\breturn\s+/g)) {
      const args = splitArgs(text, m.index + m[0].length);
      // Presenters return (message, messageCode, params). Only the shape where
      // the second argument is a quoted string is a messageCode emission.
      if (args.length < 2) continue;
      const codeArg = args[1];
      if (!/^"[a-zA-Z0-9_.]+"$/.test(codeArg)) continue;
      const code = codeArg.slice(1, -1);
      if (!out.has(code)) out.set(code, { literals: new Set(), files: new Set() });
      // A computed message (Sprintf, a variable) cannot be empty, so it renders
      // as a fallback when untranslated — the same failure as a plain literal.
      out.get(code).literals.add(args[0].startsWith('"') ? args[0] : '<computed>');
      out.get(code).files.add(name);
    }
  }
  return out;
}

const emitted = await emittedCodes();
const ja = JSON.parse(await readFile(join(LOCALES, 'ja/common.json'), 'utf8')).messageCode ?? {};
const en = JSON.parse(await readFile(join(LOCALES, 'en/common.json'), 'utf8')).messageCode ?? {};

/**
 * Codes whose message deliberately *is* the payload, so there is nothing to
 * translate. `return lastErr.Error(), "error", nil` in 8 presenters carries the
 * Go error text itself; GameMessageBox's fallback to the raw message is the
 * intended behaviour there, not a defect.
 */
const PASSTHROUGH = new Set(['error']);

const missing = [];
for (const [code, info] of emitted) {
  if (PASSTHROUGH.has(code)) continue;
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
