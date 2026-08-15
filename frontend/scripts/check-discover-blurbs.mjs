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
//
// Presence is necessary but not sufficient, so the text itself is checked too.
// The realistic way a blurb goes wrong is not deletion but **copy-paste**: a new
// game is added by duplicating a neighbouring entry and the prose is never
// rewritten, so two unrelated games carry the same description. That renders
// perfectly, reads plausibly, and is wrong -- no presence check can see it.
// The same applies to a ja entry pasted into en, which looks translated until
// you read it.

import { readFile } from 'node:fs/promises';
import { join, relative, resolve } from 'node:path';
import { fileURLToPath } from 'node:url';
import { assertFloor } from './lib/floor.mjs';

const FRONTEND = fileURLToPath(new URL('..', import.meta.url));
const REPO = join(FRONTEND, '..');
// An optional srcDir lets the self-test point the guard at a fixture tree.
// Without it the floors below would have to be met by a two-game fixture.
const SRC = process.argv[2] ? resolve(process.argv[2]) : join(FRONTEND, 'src');
const SCANNING_REPO = !process.argv[2];
const ROUTES = join(SRC, 'constants/gameRoutes.ts');
const LOCALES = join(SRC, 'i18n/locales');
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
// 264 games are registered today. A regex that has drifted usually still matches *some*
// entries, so `> 0` is not the interesting boundary -- checking 12 games and declaring all
// blurbs present is the failure this floor exists to catch.
if (SCANNING_REPO) assertFloor('discover-blurbs', games.size, 200, `game pages in ${relative(REPO, ROUTES)}`);

const gaps = [];
const empties = [];
/** lang -> section -> game -> text, for the wording checks below. */
const text = { ja: {}, en: {} };
for (const lang of ['ja', 'en']) {
  const file = join(LOCALES, lang, 'discover.json');
  const data = JSON.parse(await readFile(file, 'utf8'));
  for (const section of SECTIONS) {
    const map = data[section];
    if (!map) {
      console.error(`discover-blurbs: ${lang}/discover.json has no "${section}" section.`);
      process.exit(1);
    }
    text[lang][section] = {};
    for (const game of games) {
      if (!(game in map)) gaps.push({ lang, section, game });
      else if (String(map[game]).trim() === '') empties.push({ lang, section, game });
      else text[lang][section][game] = String(map[game]).trim();
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

// --- wording checks -------------------------------------------------------
//
// Only run once presence is established, so a missing blurb is reported as
// missing rather than as a wording fault.

/** Placeholder text that should never ship. */
const PLACEHOLDER = /\b(TODO|FIXME|WIP|Lorem ipsum|xxx)\b/i;
/**
 * Shortest believable blurb. The real ones run far longer; this only has to
 * catch a stub such as "準備中" or "TBD" that survived review.
 */
const MIN_CHARS = 8;

const wording = [];
for (const section of SECTIONS) {
  for (const lang of ['ja', 'en']) {
    const entries = Object.entries(text[lang][section]);

    // Two games sharing one description -- the copy-paste failure.
    const byText = new Map();
    for (const [game, body] of entries) {
      const seen = byText.get(body);
      if (seen) wording.push({ kind: 'DUPLICATE', lang, section, game, other: seen, body });
      else byText.set(body, game);
    }

    for (const [game, body] of entries) {
      if (PLACEHOLDER.test(body)) wording.push({ kind: 'PLACEHOLDER', lang, section, game, body });
      else if ([...body].length < MIN_CHARS) wording.push({ kind: 'TOO SHORT', lang, section, game, body });
    }
  }

  // A ja entry pasted verbatim into en (or the reverse) is untranslated. Games
  // whose blurb is legitimately identical in both languages do not exist here:
  // every one is prose, not a bare proper noun.
  for (const game of Object.keys(text.ja[section])) {
    const ja = text.ja[section][game];
    const en = text.en[section][game];
    if (en !== undefined && ja === en) {
      wording.push({ kind: 'UNTRANSLATED', lang: 'ja=en', section, game, body: ja });
    }
  }
}

if (wording.length > 0) {
  console.error('\nDiscover blurb wording problems:\n');
  for (const w of wording) {
    const where = `${w.lang}/discover.json  ${w.section}.${w.game}`;
    const detail = w.kind === 'DUPLICATE' ? `  (identical to ${w.section}.${w.other})` : '';
    console.error(`  ${w.kind.padEnd(12)} ${where}${detail}\n${' '.repeat(16)}"${w.body.slice(0, 70)}"`);
  }
  console.error(
    `\n${wording.length} problem(s). A duplicated or untranslated blurb renders perfectly` +
      ' and reads plausibly, so nothing else in the pipeline will catch it.',
  );
  process.exit(1);
}

console.log(
  `discover-blurbs: OK (all ${games.size} games have ja + en blurb and stretch_blurb;` +
    ` ${games.size * SECTIONS.length * 2} entries checked for duplicates, placeholders and untranslated text).`,
);
