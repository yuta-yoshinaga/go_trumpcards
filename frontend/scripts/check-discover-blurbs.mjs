#!/usr/bin/env bun
// Guard that every registered game has a Discover blurb in ja AND en.
//
// discover.json's `blurb` and `stretch_blurb` are flat maps keyed by game name,
// and nothing kept them in step with gameRoutes.ts. When Braid became game 230
// both maps still held 229 keys, and no check anywhere noticed:
//
//   - the i18n parity check is *horizontal*, ja against en. A game missing from
//     both locales is symmetric, so parity passes;
//   - the registry/doc guards are backend-side and never read discover.json;
//   - DiscoverPage renders `t(`blurb.${game}`)`, and i18next returns the key's
//     own last segment on a miss rather than throwing, so the card renders with
//     the bare game name where its description should be. It looks like a
//     styling bug, not a missing translation.
//
// This is a *vertical* check, route table to locale, matching
// check-message-codes.mjs.

import { readFile } from 'node:fs/promises';
import { join, relative } from 'node:path';
import { fileURLToPath } from 'node:url';

const FRONTEND = fileURLToPath(new URL('..', import.meta.url));
const REPO = join(FRONTEND, '..');
const ROUTES = join(FRONTEND, 'src/constants/gameRoutes.ts');
const LOCALES = join(FRONTEND, 'src/i18n/locales');
const SECTIONS = ['blurb', 'stretch_blurb'];

/**
 * Blurb slugs for every registered game.
 *
 * The slug is `page.toLowerCase()`, NOT the route path -- RecommendationCard and
 * StretchPickCard both derive it that way. The two differ more often than it
 * looks: BlackJack sits at `/` rather than `/blackjack`, and Pig's Tail is
 * `PigsTail` at `/pigtail`, so keying on the path reports both as broken when
 * they are fine.
 */
async function registeredGames() {
  const src = await readFile(ROUTES, 'utf8');
  const names = new Set();
  for (const m of src.matchAll(/^\s*page: '([A-Za-z0-9]+)',$/gm)) names.add(m[1].toLowerCase());
  return names;
}

const games = await registeredGames();
if (games.size === 0) {
  console.error(`discover-blurbs: found no game pages in ${relative(REPO, ROUTES)} — the regex has drifted.`);
  process.exit(1);
}

const gaps = [];
const empties = [];
for (const lang of ['ja', 'en']) {
  const file = join(LOCALES, lang, 'discover.json');
  const data = JSON.parse(await readFile(file, 'utf8'));
  for (const section of SECTIONS) {
    const map = data[section];
    if (!map) {
      console.error(`discover-blurbs: ${lang}/discover.json has no "${section}" section.`);
      process.exit(1);
    }
    for (const game of games) {
      if (!(game in map)) gaps.push({ lang, section, game });
      else if (String(map[game]).trim() === '') empties.push({ lang, section, game });
    }
    // A key with no matching route is dead weight and usually a rename.
    for (const key of Object.keys(map)) {
      if (!games.has(key)) gaps.push({ lang, section, game: key, stale: true });
    }
  }
}

if (gaps.length > 0 || empties.length > 0) {
  console.error('\nDiscover blurb gaps:\n');
  for (const g of gaps.filter((g) => !g.stale)) {
    console.error(`  MISSING  ${g.lang}/discover.json  ${g.section}.${g.game}`);
  }
  for (const g of gaps.filter((g) => g.stale)) {
    console.error(`  STALE    ${g.lang}/discover.json  ${g.section}.${g.game}  (no such game)`);
  }
  for (const e of empties) {
    console.error(`  EMPTY    ${e.lang}/discover.json  ${e.section}.${e.game}`);
  }
  console.error(
    `\n${gaps.length + empties.length} problem(s) across ${games.size} registered games.` +
      ' A missing blurb renders as the bare game name on the Discover page, not as an error.',
  );
  process.exit(1);
}

console.log(`discover-blurbs: OK (all ${games.size} games have ja + en blurb and stretch_blurb).`);
