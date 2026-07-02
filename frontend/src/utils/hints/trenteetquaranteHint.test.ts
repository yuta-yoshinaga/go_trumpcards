import { describe, expect, it } from 'vitest';
import type { TrenteEtQuaranteResponse } from '../../types/card';
import { TrenteEtQuaranteBetType, TrenteEtQuarantePhase } from '../../types/phases';
import { getTrenteEtQuaranteHint } from './trenteetquaranteHint';

function state(overrides: Partial<TrenteEtQuaranteResponse> = {}): TrenteEtQuaranteResponse {
  return {
    phase: TrenteEtQuarantePhase.BET,
    roundNumber: 0,
    chips: 1000,
    stake: 0,
    currentBet: TrenteEtQuaranteBetType.ROUGE,
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

describe('getTrenteEtQuaranteHint', () => {
  it('returns null outside the bet phase', () => {
    expect(getTrenteEtQuaranteHint(state({ phase: TrenteEtQuarantePhase.RESULT }))).toBeNull();
  });

  it('returns null when Noir is already selected', () => {
    expect(getTrenteEtQuaranteHint(state({ currentBet: TrenteEtQuaranteBetType.NOIR }))).toBeNull();
  });

  it('recommends Noir during the bet phase when another bet is selected', () => {
    const hint = getTrenteEtQuaranteHint(state());
    expect(hint?.targetAction).toBe('bet');
    expect(hint?.confidence).toBe('moderate');
    expect(hint?.reason).toBe('hint.evenMoney');
  });
});
