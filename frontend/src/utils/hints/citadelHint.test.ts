import { describe, expect, it } from 'vitest';
import type { CitadelResponse } from '../../types/card';
import { getCitadelHint } from './citadelHint';

function makeState(overrides?: Partial<CitadelResponse>): CitadelResponse {
  return {
    tableau: Array.from({ length: 8 }, () => []),
    foundation: [[], [], [], []],
    phase: 0,
    moveCount: 0,
    canUndo: false,
    isStalemate: false,
    message: '',
    ...overrides,
  };
}

describe('getCitadelHint', () => {
  it('returns null when not in playing phase', () => {
    expect(getCitadelHint(makeState({ phase: 1 }))).toBeNull();
  });

  it('returns null in stalemate', () => {
    expect(getCitadelHint(makeState({ isStalemate: true }))).toBeNull();
  });

  it('returns null when no server hint', () => {
    expect(getCitadelHint(makeState())).toBeNull();
  });

  it('maps foundation hint to strong play action', () => {
    const r = getCitadelHint(
      makeState({
        hint: { fromCol: 0, cardIndex: 0, toZone: 'foundation', toCol: 1 },
      }),
    );
    expect(r).toEqual({ targetAction: 'play.foundation', reason: 'hintReason.toFoundation', confidence: 'strong' });
  });

  it('maps tableau hint', () => {
    const r = getCitadelHint(
      makeState({
        hint: { fromCol: 1, cardIndex: 2, toZone: 'tableau', toCol: 4 },
      }),
    );
    expect(r?.targetAction).toBe('play.tableau');
    expect(r?.reason).toBe('hintReason.toTableau');
  });
});
