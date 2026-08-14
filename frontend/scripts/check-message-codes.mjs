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
import { assertFloor } from './lib/floor.mjs';

const FRONTEND = fileURLToPath(new URL('..', import.meta.url));
const REPO = join(FRONTEND, '..');
// The two roots are overridable by argv so the guard's own test can run it against a
// fixture. A guard that can only ever be pointed at the real tree cannot be shown to
// fail on bad input, and one that is never shown to fail is indistinguishable from one
// that passes everything.
const PRESENTER_DIR = process.argv[2] ?? join(REPO, 'internal/adapter/presenter');
const LOCALES = process.argv[3] ?? join(FRONTEND, 'src/i18n/locales');
const REAL_FLOOR = 1050;
const FIXTURE_FLOOR = 1;

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
    // The assignment form, which the `return` scan above cannot see at all.
    //
    // Newer presenters build an output struct and set the field:
    // `resObj.MessageCode = "colorado.gameClear"`. Those codes were never
    // collected, so the guard reported full coverage of the codes it happened
    // to recognise while the assignment form went unchecked entirely — the
    // same "confidently silent" failure the splitArgs comment above describes.
    //
    // Which of the two failure modes an untranslated one lands in cannot be read
    // off this line: `Message` is a separate statement, and several branches
    // never assign it (ColoradoWebPresenter's playing/stalemate/hint paths leave
    // a zero-valued struct). Rather than guess by proximity, they get their own
    // bucket in the report.
    for (const m of text.matchAll(/\.MessageCode\s*=\s*"([a-zA-Z0-9_.]+)"/g)) {
      const code = m[1];
      if (!out.has(code)) out.set(code, { literals: new Set(), files: new Set() });
      out.get(code).literals.add('<assigned>');
      out.get(code).files.add(name);
    }
  }
  return out;
}

const emitted = await emittedCodes();
// 1585 codes today, parsed out of the presenter sources by regex. A parser change that stops
// recognising most `return` forms — or the `.MessageCode =` assignments, which account for
// roughly half — would leave a subset of codes, all of them translated, and this guard would
// announce full coverage of the codes it could still see. The floor is what makes that visible:
// dropping the assignment scan alone takes the count back to ~774, which must not clear it.
// A fixture run (argv-supplied roots) has a handful of codes by design, so it floors at
// FIXTURE_FLOOR; the real run keeps the number that matters. Both are plain literals so
// the floor can be read off the source — check-guard-floors rejects a computed one.
if (process.argv[2]) {
  assertFloor('message-codes', emitted.size, FIXTURE_FLOOR, 'codes emitted from the fixture');
} else {
  assertFloor('message-codes', emitted.size, REAL_FLOOR, `codes emitted from ${relative(REPO, PRESENTER_DIR)}`);
}

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
  const lits = [...info.literals];
  const assigned = lits.length > 0 && lits.every((l) => l === '<assigned>');
  const blank = !assigned && lits.every((l) => l === '""');
  missing.push({ code, gaps, blank, assigned, file: [...info.files][0] });
}

if (missing.length > 0) {
  const blanks = missing.filter((m) => m.blank);
  const assigned = missing.filter((m) => m.assigned);
  const literals = missing.filter((m) => !m.blank && !m.assigned);
  console.error('\nUntranslated messageCodes (emitted by Go, absent from the locale files):\n');
  if (blanks.length > 0) {
    console.error(`  ${blanks.length} render as an EMPTY message box (no fallback text at all):`);
    for (const m of blanks) console.error(`    ${m.code}  [${m.gaps.join(' + ')}]  ${m.file}`);
  }
  if (literals.length > 0) {
    console.error(`\n  ${literals.length} render the raw Go literal, in both locales:`);
    for (const m of literals) console.error(`    ${m.code}  [${m.gaps.join(' + ')}]  ${m.file}`);
  }
  if (assigned.length > 0) {
    console.error(
      `\n  ${assigned.length} set via \`resObj.MessageCode =\`; each renders the Go literal that` +
        ' branch assigned, or an empty box if it assigned none:',
    );
    for (const m of assigned) console.error(`    ${m.code}  [${m.gaps.join(' + ')}]  ${m.file}`);
  }
  console.error(
    `\n${missing.length} untranslated code(s) of ${emitted.size} emitted.` +
      ` Add them under "messageCode" in ${relative(REPO, LOCALES)}/{ja,en}/common.json.` +
      ' See issue #4365.',
  );
  process.exit(1);
}

console.log(`message-codes: OK (all ${emitted.size} codes emitted by Go have ja + en translations).`);
