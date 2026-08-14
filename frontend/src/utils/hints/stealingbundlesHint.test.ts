import { describe, expect, it } from 'vitest';
import type { StealingBundlesResponse } from '../../types/card';
import { getStealingBundlesHint } from './stealingbundlesHint';

const state = (hint?: StealingBundlesResponse['hint']): StealingBundlesResponse =>
  ({ hint }) as StealingBundlesResponse;

describe('getStealingBundlesHint', () => {
  it('returns null without a hint', () => {
    expect(getStealingBundlesHint(state())).toBeNull();
  });

  // **束の略奪は枚数がそのまま差になる。** そこだけ確度が高い。
  it('is certain when a whole bundle is on offer', () => {
    expect(getStealingBundlesHint(state({ cardIndex: 1, victimIdx: 2, reason: 'stealingbundlesSteal' }))).toEqual({
      targetAction: 'card-1',
      reason: 'hint.stealingbundlesSteal',
      confidence: 'strong',
    });
  });

  it.each(['stealingbundlesTake', 'stealingbundlesTrail'])('is only moderate for %s', (reason) => {
    expect(getStealingBundlesHint(state({ cardIndex: 2, victimIdx: -1, reason }))).toEqual({
      targetAction: 'card-2',
      reason: `hint.${reason}`,
      confidence: 'moderate',
    });
  });

  // cardIndex 0 は「札を指していない」ではない。
  it('treats card index zero as a card', () => {
    expect(
      getStealingBundlesHint(state({ cardIndex: 0, victimIdx: -1, reason: 'stealingbundlesTake' }))?.targetAction,
    ).toBe('card-0');
  });

  it('returns null when the hint names no card', () => {
    expect(getStealingBundlesHint(state({ victimIdx: -1, reason: 'stealingbundlesTake' }))).toBeNull();
  });
});
