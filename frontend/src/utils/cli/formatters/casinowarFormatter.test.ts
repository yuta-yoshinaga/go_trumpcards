import { describe, expect, it } from 'vitest';
import type { CasinoWarResponse } from '../../../types/card';
import { CasinoWarPhase } from '../../../types/phases';
import { formatCasinowarState } from './casinowarFormatter';

function state(overrides: Partial<CasinoWarResponse> = {}): CasinoWarResponse {
  return {
    burnCards: [],
    phase: CasinoWarPhase.BET,
    chips: 1000,
    ante: 0,
    warBet: 0,
    result: 0,
    totalPayout: 0,
    message: '',
    ...overrides,
  };
}

describe('formatCasinowarState', () => {
  it('renders the BET phase header', () => {
    const out = formatCasinowarState(state());
    expect(out).toContain('Casino War');
    expect(out).toContain('phase: BET');
  });

  it('renders the initial cards once dealt', () => {
    const out = formatCasinowarState(
      state({
        phase: CasinoWarPhase.INITIAL_DEALT,
        ante: 100,
        playerCard: { design: 'SPADE', value: 13 },
        dealerCard: { design: 'HEART', value: 7 },
      }),
    );
    expect(out).toContain('player:');
    expect(out).toContain('dealer:');
    expect(out).toContain('ante: 100');
  });

  it('renders the war section with burn cards and warBet', () => {
    const out = formatCasinowarState(
      state({
        phase: CasinoWarPhase.WAR_DEALT,
        ante: 100,
        warBet: 100,
        burnCards: [
          { design: 'SPADE', value: 2 },
          { design: 'SPADE', value: 3 },
          { design: 'SPADE', value: 4 },
        ],
        playerWarCard: { design: 'HEART', value: 13 },
        dealerWarCard: { design: 'DIAMOND', value: 5 },
      }),
    );
    expect(out).toContain('Burn:');
    expect(out).toContain('War:');
    expect(out).toContain('warBet: 100');
  });

  it('renders the result on end', () => {
    const out = formatCasinowarState(
      state({
        phase: CasinoWarPhase.END,
        ante: 100,
        result: 1,
        totalPayout: 200,
        message: 'Player wins!',
      }),
    );
    expect(out).toContain('result: WIN');
    expect(out).toContain('payout: 200');
    expect(out).toContain('Player wins!');
  });
});
