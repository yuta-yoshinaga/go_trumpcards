import { describe, expect, it } from 'vitest';
import type { GoStopBreakdown } from '../types/card';
import { computeNearYaku } from './gostopYaku';

/** Builds a breakdown with all category counts zeroed, overriding only the counts under test. */
function bd(overrides: Partial<GoStopBreakdown>): GoStopBreakdown {
  return {
    gwang: 0,
    godori: 0,
    tti: 0,
    yeol: 0,
    pi: 0,
    base: 0,
    goCount: 0,
    goMult: 1,
    goScore: 0,
    brightCount: 0,
    ribbonCount: 0,
    animalCount: 0,
    piCount: 0,
    ...overrides,
  };
}

describe('computeNearYaku', () => {
  it('returns nothing for a null breakdown', () => {
    expect(computeNearYaku(null)).toEqual([]);
  });

  it('flags a hand one card away from samgwang (3 brights)', () => {
    const near = computeNearYaku(bd({ brightCount: 2 }));
    expect(near).toEqual([{ category: 'gwang', target: 'samgwang', current: 2, remaining: 1 }]);
  });

  it('advances the gwang target to sagwang once 3 brights are held', () => {
    const near = computeNearYaku(bd({ brightCount: 3 }));
    expect(near).toEqual([{ category: 'gwang', target: 'sagwang', current: 3, remaining: 1 }]);
  });

  it('advances the gwang target to ogwang once 4 brights are held', () => {
    const near = computeNearYaku(bd({ brightCount: 4 }));
    expect(near).toEqual([{ category: 'gwang', target: 'ogwang', current: 4, remaining: 1 }]);
  });

  it('emits no gwang entry when all five brights are already captured', () => {
    expect(computeNearYaku(bd({ brightCount: 5 }))).toEqual([]);
  });

  it('flags tti, yeol and pi one card from their thresholds', () => {
    const near = computeNearYaku(bd({ ribbonCount: 4, animalCount: 4, piCount: 9 }));
    expect(near).toEqual([
      { category: 'tti', target: 'tti', current: 4, remaining: 1 },
      { category: 'yeol', target: 'yeol', current: 4, remaining: 1 },
      { category: 'pi', target: 'pi', current: 9, remaining: 1 },
    ]);
  });

  it('includes a two-card-away yaku but drops one farther than the preview window', () => {
    // brightCount 1 -> 2 away from samgwang (within window); ribbonCount 2 -> 3 away (excluded).
    const near = computeNearYaku(bd({ brightCount: 1, ribbonCount: 2 }));
    expect(near).toEqual([{ category: 'gwang', target: 'samgwang', current: 1, remaining: 2 }]);
  });

  it('shows nothing when every category is far from a threshold', () => {
    expect(computeNearYaku(bd({ brightCount: 0, ribbonCount: 0, animalCount: 0, piCount: 0 }))).toEqual([]);
  });

  it('does not flag categories already at or past their threshold', () => {
    expect(computeNearYaku(bd({ ribbonCount: 5, animalCount: 6, piCount: 12 }))).toEqual([]);
  });
});
