import { describe, expect, it } from 'vitest';
import type { TrenteEtQuaranteResponse } from '../../../types/card';
import { TrenteEtQuaranteBetType, TrenteEtQuarantePhase } from '../../../types/phases';
import { formatTrenteEtQuaranteState } from './trenteetquaranteFormatter';

function state(overrides: Partial<TrenteEtQuaranteResponse> = {}): TrenteEtQuaranteResponse {
  return {
    phase: TrenteEtQuarantePhase.BET,
    roundNumber: 0,
    chips: 1000,
    stake: 0,
    currentBet: TrenteEtQuaranteBetType.NOIR,
    noirRow: [],
    rougeRow: [],
    noirTotal: 0,
    rougeTotal: 0,
    winningRow: -1,
    firstCardRed: false,
    refait: false,
    result: 0,
    payout: 0,
    remainingDeck: 312,
    gameEndFlag: false,
    config: { defaultBet: 0 },
    message: '',
    ...overrides,
  };
}

describe('formatTrenteEtQuaranteState', () => {
  it('renders the BET phase header', () => {
    const out = formatTrenteEtQuaranteState(state());
    expect(out).toContain('Trente et Quarante');
    expect(out).toContain('phase: BET');
  });

  it('renders the stake and bet type once a bet is placed', () => {
    const out = formatTrenteEtQuaranteState(
      state({ phase: TrenteEtQuarantePhase.RESULT, stake: 100, currentBet: TrenteEtQuaranteBetType.ROUGE }),
    );
    expect(out).toContain('stake: 100');
    expect(out).toContain('bet: ROUGE');
  });

  it('renders both rows with totals', () => {
    const out = formatTrenteEtQuaranteState(
      state({
        phase: TrenteEtQuarantePhase.RESULT,
        stake: 100,
        noirRow: [
          { design: 'SPADE', value: 10 },
          { design: 'CLOVER', value: 13 },
          { design: 'SPADE', value: 8 },
        ],
        rougeRow: [
          { design: 'HEART', value: 9 },
          { design: 'DIAMOND', value: 13 },
          { design: 'HEART', value: 10 },
        ],
        noirTotal: 31,
        rougeTotal: 39,
        winningRow: 0,
        result: 1,
        payout: 200,
      }),
    );
    expect(out).toContain('Noir (31):');
    expect(out).toContain('Rouge (39):');
    expect(out).toContain('winner: NOIR');
    expect(out).toContain('result: WIN');
    expect(out).toContain('payout: 200');
  });

  it('renders a refait on a tie at 31', () => {
    const out = formatTrenteEtQuaranteState(
      state({
        phase: TrenteEtQuarantePhase.RESULT,
        stake: 100,
        noirTotal: 31,
        rougeTotal: 31,
        noirRow: [{ design: 'SPADE', value: 1 }],
        rougeRow: [{ design: 'HEART', value: 1 }],
        winningRow: -1,
        refait: true,
        result: -1,
        payout: 50,
      }),
    );
    expect(out).toContain('REFAIT');
    expect(out).toContain('result: LOSE');
    expect(out).toContain('payout: 50');
  });

  it('renders a LOSE result with a Rouge winner', () => {
    const out = formatTrenteEtQuaranteState(
      state({
        phase: TrenteEtQuarantePhase.RESULT,
        stake: 100,
        noirTotal: 38,
        rougeTotal: 32,
        winningRow: 1,
        result: -1,
        payout: 0,
        message: 'You lose.',
      }),
    );
    expect(out).toContain('winner: ROUGE');
    expect(out).toContain('result: LOSE');
    expect(out).toContain('You lose.');
  });

  it('renders UNKNOWN phase fallback', () => {
    const out = formatTrenteEtQuaranteState(state({ phase: 999 as unknown as number }));
    expect(out).toContain('phase: UNKNOWN');
  });
});
