import { describe, expect, it } from 'vitest';
import type { PaiGowResponse } from '../../../types/card';
import { formatPaigowState } from './paigowFormatter';

function makeState(overrides?: Partial<PaiGowResponse>): PaiGowResponse {
  return {
    playerCards: [],
    dealerCards: [],
    playerHighHand: [],
    playerLowHand: [],
    dealerHighHand: [],
    dealerLowHand: [],
    phase: 1,
    chips: 1000,
    bet: 0,
    result: 0,
    highHandResult: 0,
    lowHandResult: 0,
    payout: 0,
    commission: 0,
    playerHighRank: 0,
    playerLowRank: 0,
    dealerHighRank: 0,
    dealerLowRank: 0,
    message: '',
    messageCode: '',
    ...overrides,
  };
}

describe('formatPaigowState', () => {
  it('shows bet phase with chips', () => {
    const result = formatPaigowState(makeState());
    expect(result).toContain('Pai Gow Poker');
    expect(result).toContain('chips: 1000');
    expect(result).toContain('BET');
  });

  it('shows player cards in set hands phase', () => {
    const result = formatPaigowState(
      makeState({
        phase: 2,
        bet: 100,
        playerCards: [
          { design: 'SPADE', value: 1 },
          { design: 'HEART', value: 13 },
          { design: 'CLOVER', value: 12 },
          { design: 'DIAMOND', value: 11 },
          { design: 'SPADE', value: 10 },
          { design: 'HEART', value: 5 },
          { design: 'CLOVER', value: 3 },
        ],
      }),
    );
    expect(result).toContain('SET_HANDS');
    expect(result).toContain('Your cards:');
    expect(result).toContain('bet: 100');
  });

  it('shows hands and payout in end phase', () => {
    const result = formatPaigowState(
      makeState({
        phase: 3,
        bet: 100,
        payout: 195,
        commission: 5,
        playerHighHand: [{ design: 'SPADE', value: 1 }],
        playerLowHand: [{ design: 'HEART', value: 5 }],
        dealerHighHand: [{ design: 'DIAMOND', value: 9 }],
        dealerLowHand: [{ design: 'CLOVER', value: 2 }],
        message: 'Player wins!',
      }),
    );
    expect(result).toContain('Player high:');
    expect(result).toContain('Player low:');
    expect(result).toContain('Dealer high:');
    expect(result).toContain('Dealer low:');
    expect(result).toContain('payout: 195');
    expect(result).toContain('commission: 5');
    expect(result).toContain('Player wins!');
  });

  it('shows UNKNOWN for unexpected phase', () => {
    const result = formatPaigowState(makeState({ phase: 99 }));
    expect(result).toContain('UNKNOWN');
  });

  it('hides bet line when bet is 0', () => {
    const result = formatPaigowState(makeState({ bet: 0 }));
    expect(result).not.toContain('bet:');
  });
});
