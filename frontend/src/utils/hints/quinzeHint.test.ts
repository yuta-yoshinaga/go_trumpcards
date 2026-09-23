import { describe, expect, it } from 'vitest';
import type { QuinzeResponse } from '../../types/card';
import { QuinzePhase } from '../../types/games/quinze';
import { getQuinzeHint } from './quinzeHint';

type Extra = { totalPoints?: number; hidden?: boolean };

function base({ totalPoints = 8, hidden = false, ...overrides }: Partial<QuinzeResponse> & Extra = {}) {
  return {
    seats: [
      {
        name: 'You',
        isCpu: false,
        hand: {
          cards: [null, null],
          bet: 10,
          totalPoints,
          totalLabel: String(totalPoints / 2),
          stood: false,
          payout: 0,
          hidden,
        },
      },
    ],
    bankerIdx: 1,
    isHumanBanker: false,
    chips: 100,
    activeSeat: 0,
    nextBanker: -1,
    lastResult: '',
    phase: QuinzePhase.PLAYER_TURN,
    targetPoints: 15,
    canHit: true,
    canStand: true,

    message: '',
    ...overrides,
  } as QuinzeResponse;
}

describe('getQuinzeHint', () => {
  it('stays quiet outside the player turn', () => {
    expect(getQuinzeHint(base({ phase: QuinzePhase.BET }))).toBeNull();
  });

  it('stays quiet when another seat is acting', () => {
    expect(getQuinzeHint(base({ activeSeat: 1 }))).toBeNull();
  });

  it('stays quiet while the hand is hidden', () => {
    expect(getQuinzeHint(base({ hidden: true }))).toBeNull();
  });

  // **ちょうど 15 は最強。**バンクが動くのはこの目だけなので、引く理由がない。
  it('stands on an exact fifteen', () => {
    expect(getQuinzeHint(base({ totalPoints: 15 }))).toEqual({
      targetAction: 'stand',
      reason: 'frontendHint.quinzeExact',
      confidence: 'strong',
    });
  });

  it('hits while far from the target', () => {
    expect(getQuinzeHint(base({ totalPoints: 8 }))).toEqual({
      targetAction: 'hit',
      reason: 'frontendHint.quinzeHitLow',
      confidence: 'strong',
    });
  });

  // 境界: 残り 4 半点（＝5.5）までは引く。それ以上は止める。
  it('treats the boundary consistently', () => {
    expect(getQuinzeHint(base({ totalPoints: 11 }))?.targetAction).toBe('hit');
    expect(getQuinzeHint(base({ totalPoints: 12 }))?.targetAction).toBe('stand');
  });

  it('stands when close to the target', () => {
    expect(getQuinzeHint(base({ totalPoints: 13 }))).toEqual({
      targetAction: 'stand',
      reason: 'frontendHint.quinzeStandClose',
      confidence: 'moderate',
    });
  });

  // **押せない手を勧めない。**
  it('does not tell the player to hit when hitting is closed', () => {
    expect(getQuinzeHint(base({ totalPoints: 8, canHit: false }))).toBeNull();
  });

  it('does not tell the player to stand when standing is closed', () => {
    expect(getQuinzeHint(base({ totalPoints: 13, canStand: false }))).toBeNull();
  });

  // ちょうど15でもスタンドが閉じていれば助言しない。
  it('says nothing on an exact total when standing is closed', () => {
    expect(getQuinzeHint(base({ totalPoints: 15, canStand: false }))).toBeNull();
  });

  it('stays quiet without a dealt hand', () => {
    const s = base();
    s.seats[0].hand = undefined;
    expect(getQuinzeHint(s)).toBeNull();
  });
});
