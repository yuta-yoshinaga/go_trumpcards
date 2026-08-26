import { describe, expect, it } from 'vitest';
import type { Card, CaribbeanDrawResponse } from '../../types/card';
import { CaribbeanDrawPhase } from '../../types/phases';
import { getCaribbeanDrawHint } from './caribbeandrawHint';

const card = (design: Card['design'], value: number): Card => ({ design, value });

function makeState(overrides: Partial<CaribbeanDrawResponse> = {}): CaribbeanDrawResponse {
  return {
    playerHand: [card('SPADE', 2), card('HEART', 5), card('DIAMOND', 8), card('CLOVER', 10), card('SPADE', 12)],
    dealerHand: [],
    phase: CaribbeanDrawPhase.ACTION,
    chips: 1000,
    anteBet: 100,
    jackpotBet: 0,
    playBet: 0,
    result: 0,
    antePayout: 0,
    playPayout: 0,
    jackpotPayout: 0,
    totalPayout: 0,
    dealerQualified: false,
    drawCost: 0,
    playerHandRank: 0,
    dealerHandRank: 0,
    message: '',
    ...overrides,
  };
}

describe('getCaribbeanDrawHint', () => {
  it('returns null outside the draw and action phases', () => {
    expect(getCaribbeanDrawHint(makeState({ phase: CaribbeanDrawPhase.BET }))).toBeNull();
    expect(getCaribbeanDrawHint(makeState({ phase: CaribbeanDrawPhase.END }))).toBeNull();
  });

  describe('draw phase', () => {
    it('recommends standing pat on a made hand', () => {
      const hint = getCaribbeanDrawHint(makeState({ phase: CaribbeanDrawPhase.DRAW, playerHandRank: 1 }));
      expect(hint?.targetAction).toBe('stand');
      expect(hint?.reason).toBe('hint.standPat');
      expect(hint?.confidence).toBe('strong');
    });

    it('recommends drawing without a made hand', () => {
      const hint = getCaribbeanDrawHint(makeState({ phase: CaribbeanDrawPhase.DRAW, playerHandRank: 0 }));
      expect(hint?.targetAction).toBe('draw');
      expect(hint?.reason).toBe('hint.drawWeak');
      expect(hint?.confidence).toBe('moderate');
    });

    it('does not fall back to the action-phase advice', () => {
      // A-K high is a *call* reason. Reaching it during the draw would mean the
      // phase check had let the action branch run one phase early.
      const hint = getCaribbeanDrawHint(
        makeState({
          phase: CaribbeanDrawPhase.DRAW,
          playerHand: [card('SPADE', 1), card('HEART', 13), card('DIAMOND', 8), card('CLOVER', 5), card('SPADE', 3)],
          playerHandRank: 0,
        }),
      );
      expect(hint?.reason).toBe('hint.drawWeak');
      expect(hint?.targetAction).not.toBe('play');
    });
  });

  it('recommends play with pair or better (strong)', () => {
    const hint = getCaribbeanDrawHint(makeState({ playerHandRank: 1 }));
    expect(hint?.targetAction).toBe('play');
    expect(hint?.confidence).toBe('strong');
    expect(hint?.reason).toBe('hint.pairOrBetter');
  });

  it('recommends play with Ace-King high (moderate)', () => {
    const state = makeState({
      playerHand: [card('SPADE', 1), card('HEART', 13), card('DIAMOND', 8), card('CLOVER', 5), card('SPADE', 3)],
      playerHandRank: 0,
    });
    const hint = getCaribbeanDrawHint(state);
    expect(hint?.targetAction).toBe('play');
    expect(hint?.confidence).toBe('moderate');
    expect(hint?.reason).toBe('hint.aceKingHigh');
  });

  it('recommends fold with weak high-card hand', () => {
    const hint = getCaribbeanDrawHint(makeState({ playerHandRank: 0 }));
    expect(hint?.targetAction).toBe('fold');
    expect(hint?.reason).toBe('hint.weakHand');
  });

  it('returns null when player hand is empty', () => {
    expect(getCaribbeanDrawHint(makeState({ playerHand: [] }))).toBeNull();
    expect(getCaribbeanDrawHint(makeState({ phase: CaribbeanDrawPhase.DRAW, playerHand: [] }))).toBeNull();
  });
});
