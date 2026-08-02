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
    'BlackHole',
    'showLegalHint という自前のフラグでのみ読む。ヒントボタンが立て、次の手で降ろす。' +
      ' #4483 の常時表示とは別の経路。',
  ],
  ['LaBelleLucie', 'showHint という自前のフラグでのみ読む。ヒントボタンが立て、次の手で降ろす。'],
  ['RussianBank', 'showHint という自前のフラグでのみ読む。ヒントボタンが立て、次の手で降ろす。'],
  [
    'Bura',
    'Bura の hint は #4492 (ゲーム追加時) から Output に乗っていて、手札のハイライトは' +
      ' 元からの仕様。#4483 が持ち込んだ常時表示ではない。',
  ],
  [
    'Speed',
    'Speed の hint は #1055 (ゲーム追加時) から Output に乗っていて、リアルタイム進行の' +
      ' UI の一部。#4483 が持ち込んだ常時表示ではないので、門番を通すと元からある機能が消える。',
  ],
]);

/**
 * Whether a page's test exercises the gate in **both** directions.
 *
 * A test that only asserts "nothing shows when not requested" passes with
 * `isRequestedHint` stubbed to `false` — it is precisely the assertion that
 * holds when the feature is broken. 16 of the banner pages were in that state
 * and got the positive case added.
 *
 * The two surfaces need different assertions, so both spellings count:
 *
 *   - **banner** pages render a line of text and use the paired
 *     `renders no hint banner …` / `renders the hint banner once …`;
 *   - **highlight** pages (the solitaires — Accordion, Yukon, …) put a ring or
 *     an aria-label on a card instead, so their tests are named for what they
 *     look at. Requiring the banner wording there would be asking them to test
 *     something they do not render.
 *
 * So a page counts as covered when its test mentions `hintAvailable` or
 * `hintRequested` at all: that message code is what `isRequestedHint` reads, and
 * a test that sets it is necessarily driving the true branch.
 */
function testsBothSides(src) {
  const negative = src.includes('renders no hint banner when the hint was not requested');
  const positive = src.includes('hintAvailable') || src.includes('hintRequested');
  return positive && (negative || !src.includes('renders no hint banner'));
}

const oneSided = [];
const offenders = [];
for (const file of await readdir(PAGES)) {
  if (!file.endsWith('Page.tsx')) continue;
  const game = file.replace(/Page\.tsx$/, '');
  if (ALLOWED.has(game)) continue;
  const src = await readFile(join(PAGES, file), 'utf8');
  // Two surfaces leak an unasked hint, and the second was missed at first:
  //
  //   1. a JSX banner — `{state.hint && …}` / `{state.hint ? … : …}`;
  //   2. a derived highlight — `const hintTo = state.hint && …`, which ends up
  //      as a ring or an aria-label on a card rather than a line of text.
  //
  // Checking only the banner let Yukon and Russian Solitaire keep announcing
  // the suggested card in its aria-label to a player who never asked.
  const rendersHint = /\{state\.hint(\?)?[.\s]/.test(src);
  // **`[\s\S]` で改行をまたぐ。**`[^\n]*` にすると、Biome が長い三項を
  // `isRequestedHint(state)` の後で折り返した瞬間に見落とす。今そう書いている
  // ページは無いが、この門番の役目は「次のを驚きで捕まえる」ことなので、
  // 折り返しに強い形にしておく (#4608 のレビュー指摘)。
  // 宣言 1 つ分に収めるため、次の `;` までで止める。
  const derivesHint = /^\s*const \w+ =[^;]*state\.hint/m.test(src);
  if (!rendersHint && !derivesHint) continue;
  if (!src.includes('isRequestedHint')) {
    offenders.push(game);
    continue;
  }

  // 門番は付いている。**そのテストが両側を踏んでいるか**も見る。
  let test;
  try {
    test = await readFile(join(PAGES, `${game}Page.test.tsx`), 'utf8');
  } catch {
    oneSided.push(`${game} (Page.test.tsx が無い)`);
    continue;
  }
  if (!testsBothSides(test)) oneSided.push(game);
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

if (oneSided.length > 0) {
  console.error('\nGated, but only one side is tested:\n');
  for (const game of oneSided.sort()) console.error(`  ONE-SIDED  ${game}`);
  console.error(
    `\nA test that only asserts "no banner when not requested" passes with\n` +
      `isRequestedHint stubbed to false — it holds precisely when the feature is\n` +
      `broken. Add the other half:\n\n` +
      `  it('renders the hint banner once the hint was requested', async () => {\n` +
      `    mockExec.mockResolvedValue({ ...state, hint: {…}, messageCode: '<game>.hintRequested' });\n` +
      `    renderWithProviders(<XPage />);\n` +
      `    expect(await screen.findByText(/\\(\\[0\\]\\)/)).toBeInTheDocument();\n` +
      `  });\n`,
  );
  process.exit(1);
}

const gated = (await readdir(PAGES)).filter((f) => f.endsWith('Page.tsx')).length;
console.log(`hint-gate: OK (both sides tested on every gated page; ${gated} pages checked).`);
