import { describe, expect, it } from 'vitest';
import type { Card, OichoKabuResponse } from '../../../types/card';
import { OichoKabuPhase } from '../../../types/phases';
import { formatOichokabuState } from './oichokabuFormatter';

const kabu = (value: number): Card => ({
  design: 'SPADE',
  value,
  glyph: String(value),
  label: String(value),
  color: 'black',
  deck: 'kabu',
});

function state(overrides: Partial<OichoKabuResponse> = {}): OichoKabuResponse {
  return {
    playerHand: [],
    bankerHand: [],
    playerRank: 0,
    bankerRank: 0,
    phase: OichoKabuPhase.BET,
    chips: 1000,
    bet: 0,
    result: 0,
    totalPayout: 0,
    message: '',
    ...overrides,
  };
}

describe('formatOichokabuState', () => {
  it('formats the bet phase', () => {
    const out = formatOichokabuState(state());
    expect(out).toContain('Oicho-Kabu');
    expect(out).toContain('phase: BET');
  });

  it('shows the child hand and rank during the draw phase', () => {
    const out = formatOichokabuState(
      state({ phase: OichoKabuPhase.DRAW, bet: 100, playerHand: [kabu(4), kabu(5)], playerRank: 9 }),
    );
    expect(out).toContain('bet: 100');
    expect(out).toContain('Child:');
    expect(out).toContain('kabu 9');
    expect(out).not.toContain('Banker:');
  });

  it('shows both hands and the result at the end', () => {
    const out = formatOichokabuState(
      state({
        phase: OichoKabuPhase.END,
        bet: 100,
        playerHand: [kabu(4), kabu(5)],
        playerRank: 9,
        bankerHand: [kabu(8), kabu(9)],
        bankerRank: 7,
        result: 1,
        totalPayout: 200,
        message: 'You win!',
      }),
    );
    expect(out).toContain('Banker:');
    expect(out).toContain('result: WIN');
    expect(out).toContain('payout: 200');
    expect(out).toContain('You win!');
  });
});
