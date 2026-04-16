import { describe, expect, it } from 'vitest';
import type { RedDogResponse } from '../../types/card';
import { RedDogPhase } from '../../types/phases';
import { getReddogHint } from './reddogHint';

function state(overrides: Partial<RedDogResponse> = {}): RedDogResponse {
  return {
    initialCards: [],
    phase: RedDogPhase.SPREAD_DECISION,
    chips: 1000,
    ante: 100,
    raise: 0,
    spread: 1,
    result: 0,
    totalPayout: 0,
    message: '',
    ...overrides,
  };
}

describe('getReddogHint', () => {
  it('returns null outside of spread decision phase', () => {
    expect(getReddogHint(state({ phase: RedDogPhase.BET }))).toBeNull();
    expect(getReddogHint(state({ phase: RedDogPhase.END }))).toBeNull();
    expect(getReddogHint(state({ phase: RedDogPhase.PAIR_THIRD }))).toBeNull();
  });

  it('recommends raise on large spreads (>=7)', () => {
    const hint = getReddogHint(state({ spread: 7 }));
    expect(hint?.targetAction).toBe('raise');
    expect(hint?.confidence).toBe('strong');
  });

  it('recommends raise on very large spreads', () => {
    expect(getReddogHint(state({ spread: 11 }))?.targetAction).toBe('raise');
  });

  it('recommends stay on small spreads (<7)', () => {
    expect(getReddogHint(state({ spread: 1 }))?.targetAction).toBe('stay');
    expect(getReddogHint(state({ spread: 6 }))?.targetAction).toBe('stay');
  });
});
