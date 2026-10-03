import { describe, expect, it } from 'vitest';
import type { BaccaratResponse } from '../../../types/card';
import { formatBaccaratState } from './baccaratFormatter';

describe('formatBaccaratState', () => {
  it('formats positive side-bet payouts with a plus sign', () => {
    const state: BaccaratResponse = {
      playerHand: [],
      bankerHand: [],
      playerHandValue: 0,
      bankerHandValue: 0,
      phase: 2,
      chips: 100,
      betAmount: 0,
      betType: 0,
      result: 0,
      payout: 0,
      history: [],
      playerPairBet: 0,
      bankerPairBet: 0,
      sideBetResults: [{ betType: 0, resultType: 1, resultName: 'Pair', betAmount: 1, payout: 3 }],
      message: '',
    };

    expect(formatBaccaratState(state)).toContain('Pair: +3');
  });
});
