import { describe, expect, it } from 'vitest';
import type { MinibridgeResponse } from '../../types/card';
import { getMinibridgeHint } from './minibridgeHint';

const state = (hint?: MinibridgeResponse['hint']): MinibridgeResponse => ({ hint }) as MinibridgeResponse;

describe('getMinibridgeHint', () => {
  it('returns null without a hint', () => {
    expect(getMinibridgeHint(state())).toBeNull();
  });

  // **契約の助言は札を指さない。**
  it('names the contract before play', () => {
    expect(getMinibridgeHint(state({ reason: 'minibridgeContract', level: 3, suit: 0 }))).toEqual({
      targetAction: 'contract-3-0',
      reason: 'hint.minibridgeContract',
      confidence: 'moderate',
    });
  });

  it('names a card in your own hand', () => {
    expect(getMinibridgeHint(state({ cardIndex: 4, reason: 'minibridgeWinTrick', level: 0, suit: 0 }))).toEqual({
      targetAction: 'card-4',
      reason: 'hint.minibridgeWinTrick',
      confidence: 'strong',
    });
  });

  // **ダミーの札は別の手札を指す。** 同じ index でも掴む場所が違う。
  it('distinguishes a card in the dummy', () => {
    expect(getMinibridgeHint(state({ cardIndex: 4, reason: 'minibridgeDummy', level: 0, suit: 0 }))).toEqual({
      targetAction: 'dummy-4',
      reason: 'hint.minibridgeDummy',
      confidence: 'moderate',
    });
  });

  // cardIndex 0 は「札を指していない」ではない。
  it('treats card index zero as a card', () => {
    expect(
      getMinibridgeHint(state({ cardIndex: 0, reason: 'minibridgeWinTrick', level: 0, suit: 0 }))?.targetAction,
    ).toBe('card-0');
  });
});
