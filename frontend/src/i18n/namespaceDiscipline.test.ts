import { describe, expect, it } from 'vitest';

/**
 * Vite-side eager glob: every `.ts`/`.tsx` source file under `src/`
 * (excluding test files via the test runner's filter on `.test.tsx?`)
 * imported as a raw string. Lets the discipline check scan the whole
 * frontend surface without pulling in `node:fs` / `__dirname`, which
 * `tsc` rejects under the current types config.
 */
const sources = import.meta.glob<string>('../**/*.{ts,tsx}', {
  query: '?raw',
  import: 'default',
  eager: true,
});

describe('i18n namespace discipline', () => {
  // `tc` is conventionally `useTranslation('common').t` in this codebase,
  // and `t` is bound to a game namespace via `useGamePageSetup(<game>)`.
  // Both `tc('common.foo')` and `tc('common:foo')` therefore double the
  // namespace prefix — i18next interprets the leading `common.` as a
  // nested key path (or the `:` form as a redundant namespace override),
  // and the lookup silently misses against the flat keys we ship. Same
  // story for `t('common.foo')` since `t` is never bound to common.
  //
  // Regression: real bug fixed alongside this test — multiple pages
  // rendered raw keys like `common.loading` / `common.reset` /
  // `common:showActionLog` to users because they called
  // `tc('common.X')` / `t('common.X')` / `t('common:X')`.
  it('no `t`/`tc` call doubles the `common` namespace', () => {
    const violations: string[] = [];
    for (const [path, src] of Object.entries(sources)) {
      if (/\.test\.tsx?$/.test(path)) continue;
      // `(?<!\.)` excludes direct `i18n.t(...)` calls where the
      // namespace is explicit and correct — only component-scope
      // `t()` / `tc()` (which are bound to a namespace already) are
      // the problem.
      const matches = src.match(/(?<!\.)\b(?:t|tc)\(\s*['"`]common[.:][a-z][\w.]*['"`]/gi);
      if (matches) {
        violations.push(`${path}: ${matches.join(', ')}`);
      }
    }
    expect(violations, `These call sites use the wrong namespace prefix:\n${violations.join('\n')}`).toEqual([]);
  });
});
