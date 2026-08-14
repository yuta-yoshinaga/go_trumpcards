import { describe, expect, it } from 'vitest';
import type { IsraeliWhistResponse } from '../../types/card';
import { getIsraeliWhistHint } from './israeliwhistHint';

const base = (hint?: IsraeliWhistResponse['hint']): IsraeliWhistResponse =>
  ({ hint }) as unknown as IsraeliWhistResponse;

describe('getIsraeliWhistHint', () => {
  it('returns null when the server sent no hint', () => {
    expect(getIsraeliWhistHint(base())).toBeNull();
  });

  // **入札は数とスートの両方で1つの意思決定。** 片方だけでは動けない。
  it('carries both the number and the suit for an auction bid', () => {
    expect(getIsraeliWhistHint(base({ reason: 'israeliwhistAuctionBid', value: 7, suit: 3 }))).toEqual({
      targetAction: 'auction-7-3',
      reason: 'hint.israeliwhistAuctionBid',
      confidence: 'moderate',
    });
  });

  it('turns an auction pass into its own action', () => {
    expect(getIsraeliWhistHint(base({ reason: 'israeliwhistAuctionPass', value: 0, suit: 0 }))).toEqual({
      targetAction: 'auction-pass',
      reason: 'hint.israeliwhistAuctionPass',
      confidence: 'moderate',
    });
  });

  it.each([
    ['israeliwhistBid', 4, 'bid-4'],
    ['israeliwhistAvoidRestricted', 2, 'bid-2'],
    ['israeliwhistMeetQuota', 9, 'bid-9'],
  ])('turns %s into %s', (reason, value, targetAction) => {
    expect(getIsraeliWhistHint(base({ reason, value, suit: 0 }))?.targetAction).toBe(targetAction);
  });

  // **ノルマは守らないと宣言そのものが通らない。** 迷う余地がないので strong。
  it('is certain about meeting the quota', () => {
    expect(getIsraeliWhistHint(base({ reason: 'israeliwhistMeetQuota', value: 9, suit: 0 }))?.confidence).toBe(
      'strong',
    );
    expect(getIsraeliWhistHint(base({ reason: 'israeliwhistBid', value: 4, suit: 0 }))?.confidence).toBe('moderate');
  });

  it('accepts card index 0', () => {
    expect(getIsraeliWhistHint(base({ cardIndex: 0, reason: 'israeliwhistWinTrick', value: 0, suit: 0 }))).toEqual({
      targetAction: 'card-0',
      reason: 'hint.israeliwhistWinTrick',
      confidence: 'moderate',
    });
  });

  it('is more confident about ducking once the call is made', () => {
    expect(getIsraeliWhistHint(base({ cardIndex: 2, reason: 'israeliwhistDuck', value: 0, suit: 0 }))?.confidence).toBe(
      'strong',
    );
    expect(
      getIsraeliWhistHint(base({ cardIndex: 2, reason: 'israeliwhistWinTrick', value: 0, suit: 0 }))?.confidence,
    ).toBe('moderate');
  });
});
