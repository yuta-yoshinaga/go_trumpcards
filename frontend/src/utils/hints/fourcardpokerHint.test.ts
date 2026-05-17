import { describe, expect, it } from 'vitest';
import type { FourCardPokerResponse } from '../../types/card';
import { getFourCardPokerHint } from './fourcardpokerHint';

function baseState(overrides: Partial<FourCardPokerResponse> = {}): FourCardPokerResponse {
  return {
    playerHand: [],
    dealerHand: [],
    playerBest: [],
    dealerBest: [],
    phase: 1,
    chips: 1000,
    anteBet: 0,
    acesUpBet: 0,
    playBet: 0,
    playMultiplier: 0,
    result: 0,
    antePayout: 0,
    playPayout: 0,
    anteBonusPayout: 0,
    acesUpPayout: 0,
    totalPayout: 0,
    playerHandRank: 0,
    dealerHandRank: 0,
    message: '',
    ...overrides,
  };
}

describe('getFourCardPokerHint', () => {
  it('returns null for the bet phase', () => {
    expect(getFourCardPokerHint(baseState({ phase: 1 }))).toBeNull();
  });

  it('returns null for the action phase', () => {
    expect(getFourCardPokerHint(baseState({ phase: 2 }))).toBeNull();
  });

  it('returns null for the end phase', () => {
    expect(getFourCardPokerHint(baseState({ phase: 3 }))).toBeNull();
  });
});
