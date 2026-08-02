import { describe, expect, it } from 'vitest';
import type { SetteEMezzoResponse } from '../../types/card';
import { SetteEMezzoPhase } from '../../types/phases';
import { getSetteEMezzoHint } from './settemezzoHint';

type Extra = { totalHalves?: number; hidden?: boolean; hasMatta?: boolean };

function base({
  totalHalves = 8,
  hidden = false,
  hasMatta = false,
  ...overrides
}: Partial<SetteEMezzoResponse> & Extra = {}) {
  return {
    seats: [
      {
        name: 'You',
        isCpu: false,
        hand: {
          cards: [null, null],
          bet: 10,
          totalHalves,
          totalLabel: String(totalHalves / 2),
          mattaHalves: 0,
          hasMatta,
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
    phase: SetteEMezzoPhase.PLAYER_TURN,
    targetHalves: 15,
    canHit: true,
    canStand: true,
    canSetMatta: false,
    message: '',
    ...overrides,
  } as SetteEMezzoResponse;
}

describe('getSetteEMezzoHint', () => {
  it('stays quiet outside the player turn', () => {
    expect(getSetteEMezzoHint(base({ phase: SetteEMezzoPhase.BET }))).toBeNull();
  });

  it('stays quiet when another seat is acting', () => {
    expect(getSetteEMezzoHint(base({ activeSeat: 1 }))).toBeNull();
  });

  it('stays quiet while the hand is hidden', () => {
    expect(getSetteEMezzoHint(base({ hidden: true }))).toBeNull();
  });

  // **ちょうど 7.5 は最強。**バンクが動くのはこの目だけなので、引く理由がない。
  it('stands on an exact seven and a half', () => {
    expect(getSetteEMezzoHint(base({ totalHalves: 15 }))).toEqual({
      targetAction: 'stand',
      reason: 'frontendHint.settemezzoExact',
      confidence: 'strong',
    });
  });

  it('hits while far from the target', () => {
    expect(getSetteEMezzoHint(base({ totalHalves: 8 }))).toEqual({
      targetAction: 'hit',
      reason: 'frontendHint.settemezzoHitLow',
      confidence: 'strong',
    });
  });

  // 境界: 残り 4 半点（＝5.5）までは引く。それ以上は止める。
  it('treats the boundary consistently', () => {
    expect(getSetteEMezzoHint(base({ totalHalves: 11 }))?.targetAction).toBe('hit');
    expect(getSetteEMezzoHint(base({ totalHalves: 12 }))?.targetAction).toBe('stand');
  });

  it('stands when close to the target', () => {
    expect(getSetteEMezzoHint(base({ totalHalves: 13 }))).toEqual({
      targetAction: 'stand',
      reason: 'frontendHint.settemezzoStandClose',
      confidence: 'moderate',
    });
  });

  // **マッタは値を選べる。**7.5 ちょうどに合わせられる局面が最優先。
  it('names the matta while its value is still open', () => {
    expect(getSetteEMezzoHint(base({ canSetMatta: true, hasMatta: true, totalHalves: 8 }))).toEqual({
      targetAction: 'matta',
      reason: 'frontendHint.settemezzoSetMatta',
      confidence: 'strong',
    });
  });

  // **押せない手を勧めない。**
  it('does not tell the player to hit when hitting is closed', () => {
    expect(getSetteEMezzoHint(base({ totalHalves: 8, canHit: false }))).toBeNull();
  });

  it('does not tell the player to stand when standing is closed', () => {
    expect(getSetteEMezzoHint(base({ totalHalves: 13, canStand: false }))).toBeNull();
  });

  // ちょうど 7.5 なのにスタンドが閉じている（マッタ確定前など）。
  it('says nothing on an exact total when standing is closed', () => {
    expect(getSetteEMezzoHint(base({ totalHalves: 15, canStand: false }))).toBeNull();
  });

  it('stays quiet without a dealt hand', () => {
    const s = base();
    s.seats[0].hand = undefined;
    expect(getSetteEMezzoHint(s)).toBeNull();
  });
});
