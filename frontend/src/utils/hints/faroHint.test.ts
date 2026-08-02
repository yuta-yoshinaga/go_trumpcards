import { describe, expect, it } from 'vitest';
import type { Card, FaroResponse } from '../../types/card';
import { FaroPhase } from '../../types/phases';
import { getFaroHint } from './faroHint';

const card = (design: Card['design'], value: number): Card => ({ design, value });

function base(overrides: Partial<FaroResponse> = {}) {
  return {
    phase: FaroPhase.CALL,
    chips: 100,
    bets: [],
    soda: null,
    losingCard: null,
    winningCard: null,
    split: false,
    turnsPlayed: 24,
    turnsTotal: 25,
    remaining: 3,
    callCards: [card('SPADE', 3), card('HEART', 7), card('CLOVER', 11)],
    callOrder: [],
    callWon: false,
    totalPayout: 0,
    gameEndFlag: false,
    message: '',
    ...overrides,
  } as unknown as FaroResponse;
}

describe('getFaroHint', () => {
  it('stays quiet once the game is over', () => {
    expect(getFaroHint(base({ gameEndFlag: true }))).toBeNull();
  });

  // **賭け場には触れない。**めくられた札の履歴が届かないので数えようがない。
  it('says nothing during the betting phase', () => {
    expect(getFaroHint(base({ phase: FaroPhase.BETTING }))).toBeNull();
  });

  it('says nothing during a turn', () => {
    expect(getFaroHint(base({ phase: FaroPhase.TURN }))).toBeNull();
  });

  // **3 枚 6 通りに 4:1。**数え続けていない限り不利。
  it('advises skipping the call', () => {
    expect(getFaroHint(base())).toEqual({
      targetAction: 'skipCall',
      reason: 'frontendHint.faroCallOdds',
      confidence: 'moderate',
    });
  });

  it('stays quiet before the last three cards are known', () => {
    expect(getFaroHint(base({ callCards: [] }))).toBeNull();
  });

  it('stays quiet when fewer than three cards are offered', () => {
    expect(getFaroHint(base({ callCards: [card('SPADE', 3), card('HEART', 7)] }))).toBeNull();
  });
});
