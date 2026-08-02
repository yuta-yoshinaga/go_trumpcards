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
    throw new Error('hintFactories block not found in useGameHint.ts — update useGameHint.docClaim.test.ts');
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
  // **数字はもう doc に書かない。**書いてあった "currently N" はヒントが 1 つ増える
  // たびに変わり、並行 PR がすべて同じ 1 行で競合していた。1 セッションで 15 本以上を
  // 解消するはめになり、一度は取り直しを忘れて develop を赤くしている (#4651)。
  // その周辺を機械的に解消した結果、テーブル行が 1 行に潰れて表が壊れたことも
  // あった —— `.md` は tsc も Biome も見ないので、誰も気づかなかった (#4656)。
  //
  // 正確な数は `check-hint-coverage.mjs` が毎回出しており、ドリフトすれば
  // exit 1 する。doc に同じ数を二重に書く必要は無い (#4652)。
  //
  // 残すのは「書き戻されないこと」の見張りだけ。数字を戻した人は、その行が
  // また全 PR を直列化することを知らないまま戻すので、ここで止める。
  it('does not hard-code a hint count that would conflict on every parallel PR', () => {
    const row = claudeDoc.split('\n').find((l) => l.includes('`useGameHint`'));
    expect(row, 'the useGameHint row vanished from frontend/CLAUDE.md').toBeDefined();
    expect(
      row,
      'a hint count is back in frontend/CLAUDE.md — that line conflicts on every parallel PR (#4652). ' +
        'Point at `bun scripts/check-hint-coverage.mjs` instead; it prints the counts and fails on drift.',
    ).not.toMatch(/currently \d+|\d+ registered|\d+ actually return/);
  });

  // 数えられること自体は残す。ガードとこのテストが同じ判定を持ち合わないよう、
  // スタブ判定は `utils/hintStubs.ts` に 1 つだけ置いてある (#4602 のレビュー指摘)。
  it('can still tell an implemented factory from a stub', () => {
    const total = countHintFactories(hintSource);
    const implemented = countImplemented(hintSource);
    expect(total).toBeGreaterThan(0);
    expect(implemented).toBeLessThanOrEqual(total);
  });
});
