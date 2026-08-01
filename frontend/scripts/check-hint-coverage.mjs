#!/usr/bin/env bun
// Guard that every registered game has a frontend hint factory.
//
// docs/new-game-checklist.md item 14 asks each new game for a hint factory
// registered in `hintFactories`. Nothing enforced it, and two games in a row
// shipped without one (Literature #257, Guandan #258) before a reviewer
// noticed — see issue #4557. Neither game's absence was visible anywhere:
//
//   - `useGameHint` looks the name up with `hintFactories[gameName]?.(state)`,
//     so a missing factory yields `null` rather than throwing. The page renders
//     with no hint and no error;
//   - `HintGameName` is `keyof typeof hintFactories`, so an unregistered game
//     is simply not in the union. Nothing references it, so tsc is silent;
//   - the checklist is prose. Prose drifts.
//
// This is the same vertical shape as check-discover-blurbs.mjs: route table to
// implementation. A game may be exempted, but only by naming it in ALLOWED
// below with a reason, so the exemption is a decision someone made rather than
// an omission nobody saw.

import { readdir, readFile } from 'node:fs/promises';
import { join, relative } from 'node:path';
import { fileURLToPath } from 'node:url';

const FRONTEND = fileURLToPath(new URL('..', import.meta.url));
const REPO = join(FRONTEND, '..');
const ROUTES = join(FRONTEND, 'src/constants/gameRoutes.ts');
const HOOK = join(FRONTEND, 'src/hooks/useGameHint.ts');
const PAGES = join(FRONTEND, 'src/pages');

/**
 * Games that genuinely need no hint, each with the reason.
 *
 * An entry here is a claim that a hint would be meaningless. It is not a place
 * to park unwritten ones — that is BACKLOG.
 */
const ALLOWED = new Map();

/**
 * Games that predate this guard and still have no hint.
 *
 * This is a RATCHET, not an exemption list: the guard fails if a game outside
 * both maps is missing a hint, so the set can shrink but never grow. Delete a
 * name from here when you write its factory. Filed as issue #4557.
 *
 * The count was 39 when the guard went in, which is the point. The reviewer on
 * PR #4519 saw two consecutive games skip checklist item 14 and read it as a
 * recent slip; running this for the first time showed four of the last five
 * games missing (Six-Bid Solo, Karnöffel, Literature, Guandan) and 39 overall.
 * It was never recent. Prose alone had never held.
 */
const BACKLOG = new Set([
  'acesup',
  'bideuchre',
  'blackhole',
  'boston',
  'bourre',
  'bridge',
  'briscola',
  'chinchon',
  'conquian',
  'cuarenta',
  'cuckoo',
  'doubleklondike',
  'escoba',
  'faro',
  'fivecardstud',
  'guandan',
  'handandfoot',
  'kaiser',
  'kalooki',
  'karnoffel',
  'kemps',
  'kille',
  'klaberjass',
  'labellelucie',
  'literature',
  'mao',
  'niuniu',
  'openfacechinese',
  'piquet',
  'pishti',
  'pontoon',
  'russianbank',
  'scopone',
  'settemezzo',
  'simplesimon',
  'sixbidsolo',
  'spoons',
  'threethirteen',
  'vint',
]);

/**
 * Games whose factory exists but whose page does not call useGameHint yet.
 *
 * A ratchet like BACKLOG: it only shrinks. These were already in this state
 * when the check was added (#4557) -- the factory was written, and then either
 * the page wiring was forgotten or the page predates the hook. Their hint UI
 * shows nothing today.
 */
const UNWIRED_BACKLOG = new Set([
  'bigo',
  'bigohilo',
  'blackjackswitch',
  'carioca',
  'contractrummy',
  'crazypineapple',
  'deuceswild',
  'dragontiger',
  'fourcardpoker',
  'irishpoker',
  'jokerpoker',
  'omaha',
  'omahahilo',
  'pineapple',
  'russianpoker',
  'sevencardstud',
  'sevencardstudhilo',
  'shengji',
  'spanish21',
  'videopoker',
]);

/** Game slugs for every registered route, keyed the way useGameHint is called. */
async function registeredGames() {
  const src = await readFile(ROUTES, 'utf8');
  const names = new Set();
  for (const m of src.matchAll(/^\s*path: '\/([a-z0-9]*)',$/gm)) {
    // BlackJack sits at '/', and useGameHint is called with 'blackjack' there.
    names.add(m[1] === '' ? 'blackjack' : m[1]);
  }
  return names;
}

/** Game names registered in the hintFactories map. */
async function hintedGames() {
  const src = await readFile(HOOK, 'utf8');
  const start = src.indexOf('export const hintFactories');
  if (start < 0) return null;
  // The map closes with `} satisfies Record<...>` at column 0. Stopping at the
  // wrong place swept up `hint:` from the hook's return interface below.
  const end = src.indexOf('\n} satisfies', start);
  if (end < 0) return null;
  const body = src.slice(start, end);
  const names = new Set();
  for (const m of body.matchAll(/^\s{2}([a-z0-9]+):/gm)) names.add(m[1]);
  return names;
}

/**
 * Game names whose page actually calls useGameHint.
 *
 * A factory alone shows the user nothing: the page has to call the hook, and
 * the ratchet was tightening on games that never did (#4557). Checking only
 * hintFactories let Aces Up out of BACKLOG while its page still had no hint UI.
 */
async function wiredPages() {
  const names = new Set();
  for (const file of await readdir(PAGES)) {
    if (!file.endsWith('Page.tsx')) continue;
    const src = await readFile(join(PAGES, file), 'utf8');
    const slug = file.slice(0, -'Page.tsx'.length).toLowerCase();
    for (const m of src.matchAll(/useGameHint\(\s*'([a-z0-9]+)'/g)) names.add(m[1]);
    // BlackJack calls getBlackjackHint directly instead of going through the
    // hook, and its hint banner works. Counting only the hook reported it as
    // unwired, which is the false positive this line removes.
    if (/from '\.\.\/utils\/hints\//.test(src)) names.add(slug);
  }
  return names;
}

const games = await registeredGames();
if (games.size === 0) {
  console.error(`hint-coverage: found no routes in ${relative(REPO, ROUTES)} — the regex has drifted.`);
  process.exit(1);
}

const hinted = await hintedGames();
if (hinted === null || hinted.size === 0) {
  console.error(`hint-coverage: found no factories in ${relative(REPO, HOOK)} — the regex has drifted.`);
  process.exit(1);
}

const missing = [];
for (const game of games) {
  if (!hinted.has(game) && !ALLOWED.has(game) && !BACKLOG.has(game)) missing.push(game);
}
// A factory with no matching route is dead weight, usually a rename.
const stale = [...hinted].filter((game) => !games.has(game));
// An exemption for a game that now has a hint is stale too, and misleading.
const redundant = [...ALLOWED.keys(), ...BACKLOG].filter((game) => hinted.has(game));
// A factory nobody calls is invisible to the user, which is what the ratchet
// is supposed to prevent (#4557).
const wired = await wiredPages();
const unwired = [...hinted].filter((game) => games.has(game) && !wired.has(game) && !UNWIRED_BACKLOG.has(game));
// An entry that has since been wired is stale and misleading.
const wiredBacklog = [...UNWIRED_BACKLOG].filter((game) => wired.has(game));

if (missing.length > 0 || stale.length > 0 || redundant.length > 0 || unwired.length > 0 || wiredBacklog.length > 0) {
  console.error('\nHint coverage gaps:\n');
  for (const game of missing.sort()) {
    console.error(`  MISSING    ${game}  (no factory in hintFactories)`);
  }
  for (const game of stale.sort()) {
    console.error(`  STALE      ${game}  (factory registered, but no such route)`);
  }
  for (const game of redundant.sort()) {
    console.error(`  REDUNDANT  ${game}  (listed in ALLOWED/BACKLOG, but a factory exists — drop the entry)`);
  }
  for (const game of unwired.sort()) {
    console.error(`  UNWIRED    ${game}  (factory exists, but its page never calls useGameHint)`);
  }
  for (const game of wiredBacklog.sort()) {
    console.error(
      `  REDUNDANT  ${game}  (listed in UNWIRED_BACKLOG, but its page now calls useGameHint — drop the entry)`,
    );
  }
  console.error(
    `\nEvery game needs a hint factory in ${relative(REPO, HOOK)} (checklist item 14).\n` +
      'Write the factory. Do NOT add the game to BACKLOG — that set only shrinks.\n' +
      'If a hint would genuinely be meaningless, add it to ALLOWED with the reason.\n',
  );
  process.exit(1);
}

console.log(
  `hint-coverage: OK (${hinted.size} of ${games.size} games hinted; ` +
    `${ALLOWED.size} need none; ${BACKLOG.size} still owed, ${UNWIRED_BACKLOG.size} unwired, see #4557).`,
);
