import { describe, expect, it } from 'vitest';
import type { GoofspielResponse } from '../../types/card';
import { getGoofspielHint } from './goofspielHint';

const state = (hint?: GoofspielResponse['hint']): GoofspielResponse => ({ hint }) as GoofspielResponse;

describe('getGoofspielHint', () => {
  it('returns null without a hint', () => {
    expect(getGoofspielHint(state())).toBeNull();
  });

  // **どの入札も読み合い。** 確実な助言は存在しません。
  it.each(['goofspielMatch', 'goofspielHighPrize', 'goofspielLowPrize', 'goofspielCarried'])(
    'is never more than moderate for %s',
    (reason) => {
      expect(getGoofspielHint(state({ cardIndex: 2, reason }))).toEqual({
        targetAction: 'card-2',
        reason: `hint.${reason}`,
        confidence: 'moderate',
      });
    },
  );

  it('treats card index zero as a card', () => {
    expect(getGoofspielHint(state({ cardIndex: 0, reason: 'goofspielMatch' }))?.targetAction).toBe('card-0');
  });

  it('returns null when the hint names no card', () => {
    expect(getGoofspielHint(state({ reason: 'goofspielMatch' }))).toBeNull();
  });
});
