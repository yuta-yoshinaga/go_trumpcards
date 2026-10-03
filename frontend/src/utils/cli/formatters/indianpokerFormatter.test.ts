import { describe, expect, it } from 'vitest';
import type { IndianPokerResponse } from '../../../types/card';
import { formatIndianpokerState } from './indianpokerFormatter';

describe('formatIndianpokerState', () => {
  it('formats positive round winnings with a plus sign', () => {
    const state: IndianPokerResponse = {
      players: [],
      pot: 0,
      sidePots: [],
      dealerIdx: 0,
      currentTurn: 0,
      estimatedStrength: 50,
      phase: 3,
      gameEndFlag: false,
      lastBet: 0,
      minRaise: 0,
      bettingLimit: 2,
      raiseCount: 0,
      maxBetAmount: 0,
      roundResults: [{ playerIdx: 0, card: null, cardRank: 0, wonAmount: 3 }],
      cpuActions: [],
      handCount: 1,
      ante: 10,
      message: '',
    };

    expect(formatIndianpokerState(state)).toContain('CPU 0: ? +3');
  });
});
