#!/usr/bin/env bun
// Guard that a game page never renders the server hint unasked.
//
// #4483 made `Output()` carry the passive hint so the board tooltip could read
// `state.hint`, which had been permanently undefined. The CLI formatters were
// given `isRequestedHint()` so they only print `HINT:` for an explicit hint
// command. The Web GUI pages were not: they had always read `state.hint`
// directly, and nothing in the change touched them.
//
// The result was that 42 game pages showed a "ヒントあり: …" banner on every
// turn, to players who had never pressed the hint button (#4605). Nothing
// failed — the pages rendered exactly what they were written to render, and the
// one page test that covered the banner asserted the unasked case as correct.
//
// So this checks the pairing directly: a page that reads `state.hint` in JSX
// must also gate on `isRequestedHint`. It is deliberately narrow — it does not
// try to prove the gate is on the right expression, only that the page knows
// the distinction exists.

import { readdir, readFile } from 'node:fs/promises';
import { join, relative } from 'node:path';
import { fileURLToPath } from 'node:url';

const FRONTEND = fileURLToPath(new URL('..', import.meta.url));
const REPO = join(FRONTEND, '..');
const PAGES = join(FRONTEND, 'src/pages');

/**
 * Pages that read `state.hint` for something other than a banner.
 *
 * An entry is a claim that the read cannot leak an unasked hint — for example
 * it feeds a value the player already sees. Add the reason, not just the name.
 */
const ALLOWED = new Map([
  [
    'Speed',
    'Speed の hint は #1055 (ゲーム追加時) から Output に乗っていて、リアルタイム進行の' +
      ' UI の一部。#4483 が持ち込んだ常時表示ではないので、門番を通すと元からある機能が消える。',
  ],
]);

const offenders = [];
for (const file of await readdir(PAGES)) {
  if (!file.endsWith('Page.tsx')) continue;
  const game = file.replace(/Page\.tsx$/, '');
  if (ALLOWED.has(game)) continue;
  const src = await readFile(join(PAGES, file), 'utf8');
  // Only JSX reads matter: `{state.hint && …}` / `{state.hint?.x && …}` /
  // `{state.hint ? … : …}`. A read inside a hook argument is not a render.
  if (!/\{state\.hint(\?)?[.\s]/.test(src)) continue;
  if (src.includes('isRequestedHint')) continue;
  offenders.push(game);
}

if (offenders.length > 0) {
  console.error('\nUnasked hint banners:\n');
  for (const game of offenders.sort()) {
    console.error(`  UNGATED  ${game}Page.tsx  (renders state.hint without isRequestedHint)`);
  }
  console.error(
    `\nSince #4483 every response carries state.hint, so rendering it directly shows\n` +
      `the hint to players who never asked (#4605). Gate the render:\n\n` +
      `  {state.hint && isRequestedHint(state) && (\n\n` +
      `importing isRequestedHint from ${relative(REPO, join(FRONTEND, 'src/utils/hintRequest.ts'))}.\n`,
  );
  process.exit(1);
}

const gated = (await readdir(PAGES)).filter((f) => f.endsWith('Page.tsx')).length;
console.log(`hint-gate: OK (no page renders state.hint unasked; ${gated} pages checked).`);
