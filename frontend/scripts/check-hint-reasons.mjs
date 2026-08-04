#!/usr/bin/env bun
// Guard that every hint `Reason` the Go domain emits has a ja AND en
// translation, for the games whose hint adapter passes the reason straight
// through as `hint.${hint.reason}`.
//
// This is the same vertical shape as check-message-codes.mjs, on a different
// axis. Nothing covered it: the horizontal ja-vs-en parity check passes when a
// key is missing from *both* locales, the Go tests assert which reason a hint
// carries and never that it renders, and `t()` returns the key itself on a
// miss — so the failure surfaces as `hint.strong_hand` printed at the player,
// not as an error anywhere.
//
// Five games had already shipped that way when this was written: Bridge
// (#4601), Bid Whist (#4636), and Écarté / Teen Patti / Three Card Brag
// (#4649). In each case the missing keys were the ones describing the *reason*
// — the judgement the hint exists to explain — while the action names beside
// them were translated, which is what made it easy to miss by eye.
//
// Two traps this guard is written around, both of which produced a confidently
// wrong answer when the check was first run by hand:
//
//   1. Hint files are camelCase (`teenPattiHint.ts`) while the i18n namespace
//      is the lowercased name (`teenpatti`). Lowercasing only one side hides
//      whole games.
//   2. Locales carry hint text two ways — nested (`{"hint": {"x": …}}`) and
//      flat (`{"hint.x": …}`). A by-hand version of this check read only the
//      nested shape and called 14 flat-shaped games untranslated. That was a
//      different bug: those games emit no `Reason` literal at all, and the
//      guard below skips a game with no reasons before it ever opens a locale.
//      Measured across all 23 games that do emit reasons, every reason resolves
//      through the nested shape and none through the flat one — so this reads
//      only the nested shape. A game that later files a reason under a flat key
//      fails loudly here instead of being waved through.

import { readdir, readFile } from 'node:fs/promises';
import { join, relative } from 'node:path';
import { fileURLToPath } from 'node:url';
import { assertFloor } from './lib/floor.mjs';

const FRONTEND = fileURLToPath(new URL('..', import.meta.url));
const REPO = join(FRONTEND, '..');
const HINTS = join(FRONTEND, 'src/utils/hints');
const DOMAIN = join(REPO, 'internal/domain');
const LOCALES = join(FRONTEND, 'src/i18n/locales');

/**
 * Reasons that are built by concatenation rather than written as a literal, so
 * the literal fragment in the Go source is not a key. Each needs a note saying
 * where the whole key comes from.
 */
const NOT_A_KEY = new Map([['bet_', 'Mus.go:992 — `"bet_" + musActionName(action)`; the joined keys are present']]);

/** Adapters that forward the server reason verbatim. */
async function passthroughGames() {
  const games = [];
  for (const name of await readdir(HINTS)) {
    if (!name.endsWith('Hint.ts') || name.endsWith('.test.ts')) continue;
    const src = await readFile(join(HINTS, name), 'utf8');
    if (!/reason: `hint\.\$\{/.test(src)) continue;
    games.push(name.slice(0, -'Hint.ts'.length).toLowerCase());
  }
  return games;
}

/**
 * Reason literals in a domain file.
 *
 * Two shapes count. `Reason: "x"` is the common one. The second is a helper that
 * computes the key and returns it — `func (g *X) playHintReason(...) string { ...
 * return "lead_high" }` — and it has to be read too: Ganjifa built its keys that
 * way, so the `Reason:` pattern alone found nothing, the game was skipped as
 * "emits no reasons", and three of its six hints shipped as raw keys while two
 * translated keys were never emitted. A guard that silently skips is worse than
 * no guard, because the passing line reads like coverage.
 */
function reasonsIn(src) {
  const found = new Set([...src.matchAll(/Reason:\s*"([a-z0-9_]+)"/g)].map((m) => m[1]));
  for (const fn of src.matchAll(/func [^\n]*[Hh]intReason[^\n]*\bstring\s*\{/g)) {
    const body = balancedBody(src, src.indexOf('{', fn.index + fn[0].length - 1));
    for (const r of body.matchAll(/return\s+"([a-z0-9_]+)"/g)) found.add(r[1]);
  }
  return found;
}

/** Source from an opening brace to its match, so a nested block cannot end the scan early. */
function balancedBody(src, open) {
  let depth = 0;
  for (let i = open; i < src.length; i++) {
    if (src[i] === '{') depth++;
    else if (src[i] === '}' && --depth === 0) return src.slice(open, i + 1);
  }
  return src.slice(open);
}

/** Every hint key a locale defines under its nested `hint` object. */
function definedKeys(json) {
  return new Set(json.hint && typeof json.hint === 'object' && !Array.isArray(json.hint) ? Object.keys(json.hint) : []);
}

const domainFiles = new Map();
for (const name of await readdir(DOMAIN)) {
  if (!name.endsWith('.go') || name.endsWith('_test.go')) continue;
  domainFiles.set(name.slice(0, -'.go'.length).toLowerCase(), join(DOMAIN, name));
}

const problems = [];
let checkedGames = 0;
let checkedKeys = 0;

for (const game of await passthroughGames()) {
  const domainPath = domainFiles.get(game);
  if (!domainPath) {
    problems.push(`${game}: no internal/domain/*.go matches this hint adapter — the name mapping has drifted`);
    continue;
  }
  const reasons = [...reasonsIn(await readFile(domainPath, 'utf8'))].filter((r) => !NOT_A_KEY.has(r));
  if (reasons.length === 0) continue;
  checkedGames += 1;

  for (const lang of ['ja', 'en']) {
    const locPath = join(LOCALES, lang, `${game}.json`);
    let defined;
    try {
      defined = definedKeys(JSON.parse(await readFile(locPath, 'utf8')));
    } catch {
      problems.push(`${game}: ${relative(REPO, locPath)} is missing or unreadable`);
      continue;
    }
    for (const r of reasons) {
      checkedKeys += 1;
      if (!defined.has(r)) {
        problems.push(`${game} (${lang}): hint.${r} has no translation — the player sees the key itself`);
      }
    }
  }
}

// 49 games / 502 lookups today. The passthrough pattern is matched by regex against the hint
// adapters, so drift trims the set rather than emptying it -- and a run down to three games
// still prints an OK line, just a smaller one.
assertFloor('hint-reasons', checkedGames, 30, 'games with passthrough hint reasons');
assertFloor('hint-reasons', checkedKeys, 300, 'reason lookups');

if (problems.length > 0) {
  console.error('Hint reason keys with no translation:\n');
  for (const p of problems) console.error(`  ${p}`);
  console.error('\nAdd the key to both frontend/src/i18n/locales/{ja,en}/<game>.json.');
  process.exit(1);
}

console.log(`hint-reasons: OK (${checkedKeys} reason lookups across ${checkedGames} games).`);
