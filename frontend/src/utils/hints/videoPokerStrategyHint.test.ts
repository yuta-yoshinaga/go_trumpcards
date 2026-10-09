import { describe, expect, it } from 'vitest';
import type { VideoPokerResponse } from '../../types/card';
import { VideoPokerPhase } from '../../types/phases';
import { getVideoPokerStrategyHint } from './videoPokerStrategyHint';

const state = (overrides: Partial<VideoPokerResponse> = {}): VideoPokerResponse => ({
  hand: [],
  phase: VideoPokerPhase.DRAW,
  chips: 0,
  betAmount: 0,
  result: 0,
  payout: 0,
  handRank: 0,
  handName: '',
  heldIndices: [],
  recommendedHold: [],
  recommendedHoldRule: '',
  variantName: 'videopoker',
  hands: 0,
  winRate: 0,
  net: 0,
  message: '',
  ...overrides,
});

describe('getVideoPokerStrategyHint', () => {
  it('returns null outside DRAW and when the rule key is empty', () => {
    expect(
      getVideoPokerStrategyHint(state({ phase: VideoPokerPhase.BET, recommendedHoldRule: 'keepPair' })),
    ).toBeNull();
    expect(getVideoPokerStrategyHint(state())).toBeNull();
  });
  it('returns draw-all advice for an empty recommendation', () => {
    expect(getVideoPokerStrategyHint(state({ recommendedHoldRule: 'drawAll' }))).toEqual({
      targetAction: 'draw-all',
      reason: 'hint.strategy.drawAll',
      confidence: 'moderate',
    });
  });
  it.each(['videopoker', 'deuceswild', 'jokerpoker'])('passes through server holds for %s', (variantName) => {
    expect(
      getVideoPokerStrategyHint(state({ variantName, recommendedHold: [0, 3], recommendedHoldRule: 'royalDraw4' })),
    ).toEqual({
      targetAction: 'hold:0,3',
      reason: 'hint.strategy.royalDraw4',
      confidence: 'strong',
    });
  });
  it('returns the same server recommendation for all three variants', () => {
    const hand = state({ recommendedHold: [0, 1], recommendedHoldRule: 'keepPair' });
    const hints = ['videopoker', 'deuceswild', 'jokerpoker'].map((variantName) =>
      getVideoPokerStrategyHint({ ...hand, variantName }),
    );
    expect(hints[0]).toEqual(hints[1]);
    expect(hints[1]).toEqual(hints[2]);
  });
});
