import { describe, expect, it } from 'vitest';

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

describe('frontend/CLAUDE.md hint-count claim', () => {
  // The `useGameHint` row in frontend/CLAUDE.md hard-codes the number of
  // registered hint factories ("currently N"). It drifts whenever a hint is
  // added (see issue #2474, where it still said 124 after the registry reached
  // 125). This test fails CI the moment the doc and the registry diverge.
  it('matches the actual hintFactories entry count', () => {
    const actual = countHintFactories(hintSource);
    const m = claudeDoc.match(/hintFactories`[^)]*currently (\d+)\)/);
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
});
