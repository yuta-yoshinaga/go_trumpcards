import { describe, expect, it } from 'vitest';
import type { HasenpfefferResponse } from '../../types/card';
import { getHasenpfefferHint } from './hasenpfefferHint';

const base = (hint?: HasenpfefferResponse['hint']): HasenpfefferResponse =>
  ({ hint }) as unknown as HasenpfefferResponse;

describe('getHasenpfefferHint', () => {
  it('returns null when the server sent no hint', () => {
    expect(getHasenpfefferHint(base())).toBeNull();
  });

  // **競りの助言は札ではなく額を指す。**
  it('names the number when it recommends a bid', () => {
    expect(getHasenpfefferHint(base({ reason: 'hasenpfefferBid', value: 4, suit: 0 }))).toEqual({
      targetAction: 'bid-4',
      reason: 'hint.hasenpfefferBid',
      confidence: 'moderate',
    });
  });

  it('turns a pass into its own action', () => {
    expect(getHasenpfefferHint(base({ reason: 'hasenpfefferPass', value: 0, suit: 0 }))).toEqual({
      targetAction: 'pass',
      reason: 'hint.hasenpfefferPass',
      confidence: 'moderate',
    });
  });

  // **親が降りられない場面は選択肢が無い。** 迷う余地がないので強く出す。
  it('is strong when the dealer cannot pass', () => {
    expect(getHasenpfefferHint(base({ reason: 'hasenpfefferMustBid', value: 3, suit: 0 }))).toEqual({
      targetAction: 'bid-3',
      reason: 'hint.hasenpfefferMustBid',
      confidence: 'strong',
    });
  });

  it('accepts card index 0', () => {
    expect(getHasenpfefferHint(base({ cardIndex: 0, reason: 'hasenpfefferDiscard', value: 0, suit: 3 }))).toEqual({
      targetAction: 'card-0',
      reason: 'hint.hasenpfefferDiscard',
      confidence: 'moderate',
    });
  });

  it('is strong only when going for a trick', () => {
    expect(
      getHasenpfefferHint(base({ cardIndex: 4, reason: 'hasenpfefferWinTrick', value: 0, suit: 0 }))?.confidence,
    ).toBe('strong');
    expect(
      getHasenpfefferHint(base({ cardIndex: 4, reason: 'hasenpfefferFeedPartner', value: 0, suit: 0 }))?.confidence,
    ).toBe('moderate');
  });
});
