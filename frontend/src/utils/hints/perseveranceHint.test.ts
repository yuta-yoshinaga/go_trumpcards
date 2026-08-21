import { describe, expect, it } from 'vitest';
import type { PerseveranceResponse } from '../../types/card';
import { getPerseveranceHint } from './perseveranceHint';

function makeState(overrides?: Partial<PerseveranceResponse>): PerseveranceResponse {
  return {
    tableau: Array.from({ length: 13 }, () => []),
    foundation: [[], [], [], []],
    phase: 0,
    moveCount: 0,
    canUndo: false,
    isStalemate: false,
    redealsLeft: 2,
    message: '',
    ...overrides,
  };
}

describe('getPerseveranceHint', () => {
  it('returns null when not in playing phase', () => {
    expect(getPerseveranceHint(makeState({ phase: 1 }))).toBeNull();
  });

  it('returns null in stalemate', () => {
    expect(getPerseveranceHint(makeState({ isStalemate: true }))).toBeNull();
  });

  it('returns null when no server hint', () => {
    expect(getPerseveranceHint(makeState())).toBeNull();
  });

  it('maps foundation hint to strong play action', () => {
    const r = getPerseveranceHint(
      makeState({
        hint: { fromCol: 0, cardIndex: 0, toZone: 'foundation', toCol: 1 },
      }),
    );
    expect(r).toEqual({ targetAction: 'play.foundation', reason: 'hintReason.toFoundation', confidence: 'strong' });
  });

  it('maps tableau hint', () => {
    const r = getPerseveranceHint(
      makeState({
        hint: { fromCol: 1, cardIndex: 2, toZone: 'tableau', toCol: 4 },
      }),
    );
    expect(r?.targetAction).toBe('play.tableau');
    expect(r?.reason).toBe('hintReason.toTableau');
  });
});
