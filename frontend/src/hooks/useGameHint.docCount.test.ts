import { describe, expect, it } from 'vitest';
import { splitRegistrations, stubbedFactoryNames } from '../utils/hintStubs';

/**
 * Raw-source imports via Vite's `?raw` glob — mirrors the approach in
 * `src/i18n/namespaceDiscipline.test.ts` to scan source/doc text without
 * pulling in `node:fs` / `__dirname` (which `tsc` rejects under the current
 * types config).
 */
const hintSource = Object.values(
  import.meta.glob<string>('./useGameHint.ts', {
    query: '?raw',
    import: 'default',
    eager: true,
  }),
)[0];

const hintFactorySources = import.meta.glob<string>('../utils/hints/*Hint.ts', {
  query: '?raw',
  import: 'default',
  eager: true,
});

const claudeDoc = Object.values(
  import.meta.glob<string>('../../CLAUDE.md', {
    query: '?raw',
    import: 'default',
    eager: true,
  }),
)[0];

/**
 * Count the entries in the `hintFactories` registry object. The block is flat
 * (one `name: (s) => …,` per line), delimited by `const hintFactories = {`
 * and `} satisfies Record<string, HintFn>;`.
 */
function countHintFactories(src: string): number {
  const start = src.indexOf('const hintFactories = {');
  const end = src.indexOf('} satisfies Record<string, HintFn>;', start);
  if (start === -1 || end === -1) {
    throw new Error('hintFactories block not found in useGameHint.ts — update useGameHint.docCount.test.ts');
  }
  const block = src.slice(start, end);
  // Each entry begins a line with `  <name>: (` — the start of its arrow fn.
  const matches = block.match(/^\s+[A-Za-z0-9]+:\s*\(/gm);
  return matches ? matches.length : 0;
}

/** Count registrations that can actually produce a hint. */
function countImplemented(src: string): number {
  const start = src.indexOf('const hintFactories = {');
  const end = src.indexOf('} satisfies Record<string, HintFn>;', start);
  // 判定は `utils/hintStubs.ts` に 1 つだけ置く。こことビルドガードで
  // 同じ正規表現を持ち合うと、片方だけ直したときに黙って食い違う (#4602 のレビュー指摘)。
  const stubs = stubbedFactoryNames(hintFactorySources);
  return splitRegistrations(src.slice(start, end), stubs).hinted.size;
}

describe('frontend/CLAUDE.md hint-count claim', () => {
  // The `useGameHint` row in frontend/CLAUDE.md hard-codes the number of
  // registered hint factories ("currently N"). It drifts whenever a hint is
  // added (see issue #2474, where it still said 124 after the registry reached
  // 125). This test fails CI the moment the doc and the registry diverge.
  it('matches the actual hintFactories entry count', () => {
    const actual = countHintFactories(hintSource);
    const m = claudeDoc.match(/hintFactories` \(currently (\d+) registered/);
    expect(
      m,
      'could not find the "currently N" hint count in frontend/CLAUDE.md — update the regex if the wording moved',
    ).not.toBeNull();
    const documented = Number(m?.[1]);
    expect(
      documented,
      `frontend/CLAUDE.md says ${documented} hint factories, but useGameHint.ts has ${actual} — update the doc`,
    ).toBe(actual);
  });

  // **登録数と実装数は別の数字。**登録だけを数えていたので、常に null を返す
  // 22 件が「対応済み」に入り、doc の数字が実態より 22 多かった (#4602)。
  it('matches the number of registrations that can actually return a hint', () => {
    const actual = countImplemented(hintSource);
    const m = claudeDoc.match(/of which (\d+) actually return a hint/);
    expect(m, 'could not find the implemented-hint count in frontend/CLAUDE.md').not.toBeNull();
    expect(
      Number(m?.[1]),
      `frontend/CLAUDE.md says ${m?.[1]} implemented, but useGameHint.ts has ${actual} — update the doc`,
    ).toBe(actual);
  });
});
