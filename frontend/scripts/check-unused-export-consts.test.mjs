import { describe, expect, it } from 'vitest';
import { findUnusedExportConsts } from './check-unused-export-consts.mjs';

describe('findUnusedExportConsts', () => {
  it('reports a declaration with no references', () => {
    expect(findUnusedExportConsts([['constants.ts', 'export const UNUSED = 1;']])).toEqual([
      { name: 'UNUSED', file: 'constants.ts' },
    ]);
  });

  it('reports a type-annotated declaration with no references', () => {
    expect(findUnusedExportConsts([['constants.ts', 'export const TYPED: number = 1;']])).toEqual([
      { name: 'TYPED', file: 'constants.ts' },
    ]);
  });

  it('keeps a declaration referenced in another source file', () => {
    expect(
      findUnusedExportConsts([
        ['constants.ts', 'export const USED = 1;'],
        ['page.tsx', 'if (phase === USED) {}'],
      ]),
    ).toEqual([]);
  });
});
