import { describe, expect, it } from 'vitest';
import type { DragonTigerResponse } from '../../types/card';
import { DragonTigerPhase } from '../../types/phases';
import { getDragontigerHint } from './dragontigerHint';

function state(overrides: Partial<DragonTigerResponse> = {}): DragonTigerResponse {
  return {
    phase: DragonTigerPhase.BET,
    chips: 1000,
    betAmount: 0,
    betType: 0,
    result: 0,
    payout: 0,
    history: [],
    message: '',
    ...overrides,
  };
}

describe('getDragontigerHint', () => {
  it('returns null in the bet phase (no actionable hint)', () => {
    expect(getDragontigerHint(state())).toBeNull();
  });

  it('returns null in the end phase', () => {
    expect(getDragontigerHint(state({ phase: DragonTigerPhase.END, result: 1 }))).toBeNull();
  });
});
