import { describe, expect, it } from 'vitest';
import { scoreTableClass } from './scoreTableStyles';

describe('scoreTableClass', () => {
  it('returns the compact fixed table classes on large desktop', () => {
    expect(scoreTableClass(true, 'w-full text-sm text-ds-text-muted min-w-[360px]')).toBe(
      'w-full table-fixed text-xs text-ds-text-muted [&_th]:px-1 [&_td]:px-1 [&_th]:tabular-nums [&_td]:tabular-nums [&_tbody_th:first-child]:truncate [&_tbody_td:first-child]:truncate [&_thead_th:first-child]:w-1/4 [&_tbody_th:first-child]:w-1/4 [&_tbody_td:first-child]:w-1/4',
    );
  });

  it('preserves the mobile table classes exactly', () => {
    const mobileClass = 'w-full text-sm text-ds-text-muted min-w-[340px]';
    expect(scoreTableClass(false, mobileClass)).toBe(mobileClass);
  });

  it('preserves the score section spacing on large desktop', () => {
    expect(scoreTableClass(true, 'w-full text-sm text-ds-text-muted min-w-[360px] mt-1')).toBe(
      'w-full table-fixed text-xs text-ds-text-muted [&_th]:px-1 [&_td]:px-1 [&_th]:tabular-nums [&_td]:tabular-nums [&_tbody_th:first-child]:truncate [&_tbody_td:first-child]:truncate [&_thead_th:first-child]:w-1/4 [&_tbody_th:first-child]:w-1/4 [&_tbody_td:first-child]:w-1/4 mt-1',
    );
  });
});
