import { describe, expect, it } from 'vitest';
import type { OichoKabuResponse } from '../../types/card';
import { OichoKabuPhase } from '../../types/phases';
import { getOichokabuHint } from './oichokabuHint';

function state(overrides: Partial<OichoKabuResponse> = {}): OichoKabuResponse {
  return {
    playerHand: [],
    bankerHand: [],
    playerRank: 0,
    bankerRank: 0,
    phase: OichoKabuPhase.DRAW,
    chips: 1000,
    bet: 100,
    result: 0,
    totalPayout: 0,
    message: '',
    ...overrides,
  };
}

describe('getOichokabuHint', () => {
  it('returns null outside of the draw phase', () => {
    expect(getOichokabuHint(state({ phase: OichoKabuPhase.BET }))).toBeNull();
    expect(getOichokabuHint(state({ phase: OichoKabuPhase.END }))).toBeNull();
  });

  it('recommends drawing on a low rank', () => {
    const hint = getOichokabuHint(state({ playerRank: 2 }));
    expect(hint?.targetAction).toBe('draw');
    expect(hint?.reason).toBe('hint.drawLow');
    expect(hint?.confidence).toBe('moderate');
  });

  it('recommends standing on a high rank', () => {
    const hint = getOichokabuHint(state({ playerRank: 8 }));
    expect(hint?.targetAction).toBe('stand');
    expect(hint?.reason).toBe('hint.standHigh');
  });
});
