import { describe, expect, it } from 'vitest';
import type { ClockSolitaireResponse } from '../../types/card';
import { getClocksolitaireHint } from './clocksolitaireHint';

function makeState(overrides?: Partial<ClockSolitaireResponse>): ClockSolitaireResponse {
  return {
    piles: Array.from({ length: 13 }, () => []),
    faceUpCount: Array.from({ length: 13 }, () => 0),
    phase: 0,
    stepCount: 0,
    message: '',
    ...overrides,
  };
}

describe('getClocksolitaireHint', () => {
  it('returns first step hint when no current card', () => {
    const result = getClocksolitaireHint(makeState());
    expect(result).toEqual({ targetAction: 'step', reason: 'hint.firstStep', confidence: 'strong' });
  });

  it('returns continue hint when game in progress', () => {
    const result = getClocksolitaireHint(makeState({ currentCard: { design: 'SPADE', value: 5 } }));
    expect(result).toEqual({ targetAction: 'step', reason: 'hint.continueStep', confidence: 'strong' });
  });

  it('returns null when game cleared', () => {
    expect(getClocksolitaireHint(makeState({ phase: 1 }))).toBeNull();
  });

  it('returns null when game over', () => {
    expect(getClocksolitaireHint(makeState({ phase: 2 }))).toBeNull();
  });
});
