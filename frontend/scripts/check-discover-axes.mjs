#!/usr/bin/env bun
// Guard that the Discover survey's axis definitions stay wired to the locale
// files and to every game's profile vector.
//
// Usage: check-discover-axes.mjs [srcDir]
//
// check-discover-blurbs.mjs covers the per-game prose. The survey *around* that
// prose is defined as data in `constants/discoverAxes.ts` -- axis labels,
// sub-question prompts and option labels are all `i18nKey` strings resolved at
// render time -- and nothing checked them.
//
// Two failure modes, both silent:
//
//   1. A new axis or option ships without translations. i18next returns the
//      key's last segment on a miss, so the radio renders as "lively" or the
//      heading as "label". It looks like a copy tweak nobody finished, not a
//      missing translation, and it is invisible to the ja/en parity check
//      because a key absent from *both* locales is symmetric.
//   2. A game's profile vector drifts from its axis length. `GameProfile`
//      types mood/skill as fixed 4-tuples so tsc catches those, but the tuple
//      lengths live in gameRoutes.ts while `profileLength` lives in
//      discoverAxes.ts -- nothing ties the two files together, and nothing
//      constrains the *values*. An out-of-range score does not crash; it
//      quietly skews every recommendation that reads that slot.
//
// Both are cheap to assert and neither is observable from a rendered page.

import { readFile } from 'node:fs/promises';
import { join, resolve } from 'node:path';
import { fileURLToPath } from 'node:url';
import { assertFloor } from './lib/floor.mjs';

const FRONTEND = fileURLToPath(new URL('..', import.meta.url));
const SRC = process.argv[2] ? resolve(process.argv[2]) : join(FRONTEND, 'src');
const SCANNING_REPO = !process.argv[2];

const AXES_FILE = join(SRC, 'constants/discoverAxes.ts');
const ROUTES_FILE = join(SRC, 'constants/gameRoutes.ts');
const LOCALES = join(SRC, 'i18n/locales');

/** Every `…I18nKey: '…'` in the axis definitions. */
const I18N_KEY = /(?:i18nKey|labelI18nKey|questionI18nKey):\s*'([^']+)'/g;
/** `mood: { … profileLength: 4` — the axis name and its vector length. */
const AXIS_LENGTH = /^ {2}(\w+):\s*\{[\s\S]*?profileLength:\s*(\d+)/gm;
/** `PROFILE_MAX = 5` */
const PROFILE_MAX_RE = /PROFILE_MAX\s*(?::\s*number)?\s*=\s*(\d+)/;
/**
 * A route's `page` and its `profile` object. The profile is written on one
 * line, so the body is matched non-greedily up to the closing brace rather
 * than to an indented one; `[\s\S]*?` between them stays inside the entry
 * because `page` is the last field before `profile`.
 */
const ROUTE_ENTRY = /page:\s*'([A-Za-z0-9]+)',\s*\n\s*profile:\s*\{([^}]*)\}/g;
/** `mood: [4, 1, 3, 2]` inside a profile block. */
const AXIS_VECTOR = /(\w+):\s*\[([^\]]*)\]/g;

/** Flatten a nested locale object to dotted leaf paths. */
function leafKeys(obj, prefix = '', out = new Set()) {
  for (const [k, v] of Object.entries(obj)) {
    const path = prefix ? `${prefix}.${k}` : k;
    if (v && typeof v === 'object' && !Array.isArray(v)) leafKeys(v, path, out);
    else out.add(path);
  }
  return out;
}

const axesSrc = await readFile(AXES_FILE, 'utf8');
const routesSrc = await readFile(ROUTES_FILE, 'utf8');

const keys = [...axesSrc.matchAll(I18N_KEY)].map((m) => m[1]);
const axisLengths = new Map([...axesSrc.matchAll(AXIS_LENGTH)].map((m) => [m[1], Number(m[2])]));
const profileMax = Number(PROFILE_MAX_RE.exec(axesSrc)?.[1]);

if (!Number.isInteger(profileMax)) {
  console.error('discover-axes: PROFILE_MAX not found in discoverAxes.ts -- the declaration changed.');
  process.exit(1);
}

if (SCANNING_REPO) {
  assertFloor('discover-axes', keys.length, 24, 'axis/question/option i18n keys');
  // Exact, not the usual two-thirds: the axis set is structural (mood, skill,
  // social, theme), not a list that grows. Adding a fifth axis still passes;
  // silently dropping one to three does not, which is the point.
  assertFloor('discover-axes', axisLengths.size, 4, 'axes with a profileLength');
}

/** Read a dotted path out of a nested locale object. */
function getPath(obj, dotted) {
  return dotted.split('.').reduce((o, k) => (o == null ? o : o[k]), obj);
}

const problems = [];

// 1. every data-defined key resolves, in both locales.
for (const lang of ['ja', 'en']) {
  const locale = JSON.parse(await readFile(join(LOCALES, lang, 'discover.json'), 'utf8'));
  const present = leafKeys(locale);
  for (const key of keys) {
    if (!present.has(key)) problems.push(`MISSING KEY   ${lang}/discover.json  ${key}`);
    else if (String(getPath(locale, key)).trim() === '') problems.push(`EMPTY KEY     ${lang}/discover.json  ${key}`);
  }
}

// 2. every game's profile vector matches its axis length and value range.
let routeCount = 0;
for (const [, page, block] of routesSrc.matchAll(ROUTE_ENTRY)) {
  routeCount++;
  const seen = new Set();
  for (const [, axis, body] of block.matchAll(AXIS_VECTOR)) {
    seen.add(axis);
    const expected = axisLengths.get(axis);
    if (expected === undefined) {
      problems.push(`UNKNOWN AXIS  ${page}.profile.${axis}  (no such axis in discoverAxes.ts)`);
      continue;
    }
    const values = [...body.matchAll(/-?\d+/g)].map((m) => Number(m[0]));
    if (values.length !== expected) {
      problems.push(`WRONG LENGTH  ${page}.profile.${axis}  has ${values.length}, profileLength is ${expected}`);
    }
    for (const v of values) {
      if (v < 0 || v > profileMax) {
        problems.push(`OUT OF RANGE  ${page}.profile.${axis}  ${v} not in 0..${profileMax}`);
      }
    }
  }
  for (const axis of axisLengths.keys()) {
    if (!seen.has(axis)) problems.push(`MISSING AXIS  ${page}.profile.${axis}`);
  }
}

if (SCANNING_REPO) assertFloor('discover-axes', routeCount, 200, 'game profiles in gameRoutes.ts');

if (problems.length > 0) {
  console.error('\ndiscover-axes: survey definitions are out of step.\n');
  for (const p of problems) console.error(`  ${p}`);
  console.error(
    '\nA missing key renders as its own last segment (i18next falls back rather than\n' +
      'throwing), and a bad profile slot silently skews recommendations. Neither is\n' +
      'visible from the page.\n',
  );
  process.exit(1);
}

console.log(
  `discover-axes: OK (${keys.length} survey i18n keys resolve in ja + en; ` +
    `${routeCount} profiles match ${axisLengths.size} axis lengths, values within 0..${profileMax}).`,
);
