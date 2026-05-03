import { describe, expect, it } from 'vitest';
import type { CasinoWarResponse } from '../../types/card';
import { CasinoWarPhase } from '../../types/phases';
import { getCasinowarHint } from './casinowarHint';

function state(overrides: Partial<CasinoWarResponse> = {}): CasinoWarResponse {
  return {
    burnCards: [],
    phase: CasinoWarPhase.TIE_DECISION,
    chips: 1000,
    ante: 100,
    warBet: 0,
    result: 0,
    totalPayout: 0,
    message: '',
    ...overrides,
  };
}

describe('getCasinowarHint', () => {
  it('returns null outside of tie decision phase', () => {
    expect(getCasinowarHint(state({ phase: CasinoWarPhase.BET }))).toBeNull();
    expect(getCasinowarHint(state({ phase: CasinoWarPhase.INITIAL_DEALT }))).toBeNull();
    expect(getCasinowarHint(state({ phase: CasinoWarPhase.WAR_DEALT }))).toBeNull();
    expect(getCasinowarHint(state({ phase: CasinoWarPhase.END }))).toBeNull();
  });

  it('recommends war during tie decision', () => {
    const hint = getCasinowarHint(state());
    expect(hint?.targetAction).toBe('war');
    expect(hint?.confidence).toBe('strong');
    expect(hint?.reason).toBe('hint.warEv');
  });
});
