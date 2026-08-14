import { describe, expect, it } from 'vitest';
import type { HoneymoonBridgeResponse } from '../../types/card';
import { getHoneymoonBridgeHint } from './honeymoonbridgeHint';

const state = (hint?: HoneymoonBridgeResponse['hint']): HoneymoonBridgeResponse =>
  ({ hint }) as HoneymoonBridgeResponse;

describe('getHoneymoonBridgeHint', () => {
  it('returns null without a hint', () => {
    expect(getHoneymoonBridgeHint(state())).toBeNull();
  });

  // **競りの助言は札を指さない。** 契約を指す。
  it('names the contract during the auction', () => {
    expect(getHoneymoonBridgeHint(state({ reason: 'honeymoonbridgeBid', level: 3, suit: 0 }))).toEqual({
      targetAction: 'bid-3-0',
      reason: 'hint.honeymoonbridgeBid',
      confidence: 'moderate',
    });
  });

  // レベル 0 は「降りる」。契約ではない。
  it('names a pass as a pass, not a level-zero contract', () => {
    expect(getHoneymoonBridgeHint(state({ reason: 'honeymoonbridgePass', level: 0, suit: 0 }))?.targetAction).toBe(
      'pass',
    );
  });

  it('names a card while playing', () => {
    expect(
      getHoneymoonBridgeHint(state({ cardIndex: 4, reason: 'honeymoonbridgeWinTrick', level: 0, suit: 0 })),
    ).toEqual({
      targetAction: 'card-4',
      reason: 'hint.honeymoonbridgeWinTrick',
      confidence: 'strong',
    });
  });

  // **引き合いは得点にならない。** 取りにいく助言より弱く出す。
  it('is only moderately confident about the draw phase', () => {
    expect(
      getHoneymoonBridgeHint(state({ cardIndex: 0, reason: 'honeymoonbridgeDraw', level: 0, suit: 0 }))?.confidence,
    ).toBe('moderate');
  });

  // cardIndex 0 は「札を指していない」ではない。
  it('treats card index zero as a card', () => {
    expect(
      getHoneymoonBridgeHint(state({ cardIndex: 0, reason: 'honeymoonbridgeWinTrick', level: 0, suit: 0 }))
        ?.targetAction,
    ).toBe('card-0');
  });
});
