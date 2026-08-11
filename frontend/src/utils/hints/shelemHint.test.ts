import { describe, expect, it } from 'vitest';
import type { ShelemResponse } from '../../types/card';
import { getShelemHint } from './shelemHint';

const base = (hint?: ShelemResponse['hint']): ShelemResponse => ({ hint }) as unknown as ShelemResponse;

describe('getShelemHint', () => {
  it('returns null when the server sent no hint', () => {
    expect(getShelemHint(base())).toBeNull();
  });

  // **競りのヒントは点数を指す。** 札ではない。
  it('names the points when it recommends a bid', () => {
    expect(getShelemHint(base({ reason: 'shelemBid', value: 125, suit: 0 }))).toEqual({
      targetAction: 'bid-125',
      reason: 'hint.shelemBid',
      confidence: 'moderate',
    });
  });

  // **捨て札のヒントはスートを指す。**
  it('names the suit when it recommends a discard', () => {
    expect(getShelemHint(base({ reason: 'shelemDiscard', value: 0, suit: 3 }))).toEqual({
      targetAction: 'trump-3',
      reason: 'hint.shelemDiscard',
      confidence: 'moderate',
    });
  });

  it('turns a pass into its own action', () => {
    expect(getShelemHint(base({ reason: 'shelemPass', value: 0, suit: 0 }))).toEqual({
      targetAction: 'pass',
      reason: 'hint.shelemPass',
      confidence: 'moderate',
    });
  });

  it('accepts card index 0', () => {
    expect(getShelemHint(base({ cardIndex: 0, reason: 'shelemWinTrick', value: 0, suit: 0 }))).toEqual({
      targetAction: 'card-0',
      reason: 'hint.shelemWinTrick',
      confidence: 'moderate',
    });
  });

  // カード点がそのまま契約の達否になるので、点札を乗せる手はほぼ一択。
  it('is more confident about feeding a winning partner', () => {
    expect(getShelemHint(base({ cardIndex: 2, reason: 'shelemFeedPartner', value: 0, suit: 0 }))?.confidence).toBe(
      'strong',
    );
    expect(getShelemHint(base({ cardIndex: 2, reason: 'shelemWinTrick', value: 0, suit: 0 }))?.confidence).toBe(
      'moderate',
    );
  });
});
