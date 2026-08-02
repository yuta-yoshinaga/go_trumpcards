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

import { readFile } from 'node:fs/promises';
import { join, relative } from 'node:path';
import { fileURLToPath } from 'node:url';

const FRONTEND = fileURLToPath(new URL('..', import.meta.url));
const REPO = join(FRONTEND, '..');
const ROUTES = join(FRONTEND, 'src/constants/gameRoutes.ts');
const HOOK = join(FRONTEND, 'src/hooks/useGameHint.ts');

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
  'bideuchre',
  'boston',
  'briscola',
  'faro',
  'fivecardstud',
  'guandan',
  'handandfoot',
  'kaiser',
  'kille',
  'literature',
  'niuniu',
  'pishti',
  'sixbidsolo',
  'vint',
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

if (missing.length > 0 || stale.length > 0 || redundant.length > 0) {
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
  console.error(
    `\nEvery game needs a hint factory in ${relative(REPO, HOOK)} (checklist item 14).\n` +
      'Write the factory. Do NOT add the game to BACKLOG — that set only shrinks.\n' +
      'If a hint would genuinely be meaningless, add it to ALLOWED with the reason.\n',
  );
  process.exit(1);
}

console.log(
  `hint-coverage: OK (${hinted.size} of ${games.size} games hinted; ` +
    `${ALLOWED.size} need none; ${BACKLOG.size} still owed, see #4557).`,
);
