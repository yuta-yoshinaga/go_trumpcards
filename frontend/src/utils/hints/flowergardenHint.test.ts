import { describe, expect, it } from 'vitest';
import type { FlowerGardenResponse } from '../../types/card';
import { getFlowergardenHint } from './flowergardenHint';

function makeState(overrides?: Partial<FlowerGardenResponse>): FlowerGardenResponse {
  return {
    tableau: Array.from({ length: 6 }, () => []),
    reserve: Array.from({ length: 16 }, () => null),
    foundation: [[], [], [], []],
    phase: 0,
    moveCount: 0,
    canUndo: false,
    isStalemate: false,
    message: '',
    ...overrides,
  };
}

describe('getFlowergardenHint', () => {
  it('returns null when not in playing phase', () => {
    expect(getFlowergardenHint(makeState({ phase: 1 }))).toBeNull();
  });

  it('returns null in stalemate', () => {
    expect(getFlowergardenHint(makeState({ isStalemate: true }))).toBeNull();
  });

  it('returns null when no server hint', () => {
    expect(getFlowergardenHint(makeState())).toBeNull();
  });

  it('maps foundation hint to strong play action', () => {
    const r = getFlowergardenHint(
      makeState({
        hint: { fromZone: 'tableau', fromCol: 0, cardIndex: 0, toZone: 'foundation', toCol: 1 },
      }),
    );
    expect(r).toEqual({ targetAction: 'play.foundation', reason: 'hintReason.toFoundation', confidence: 'strong' });
  });

  it('maps tableau hint', () => {
    const r = getFlowergardenHint(
      makeState({
        hint: { fromZone: 'reserve', fromCol: 1, cardIndex: 0, toZone: 'tableau', toCol: 4 },
      }),
    );
    expect(r?.targetAction).toBe('play.tableau');
    expect(r?.reason).toBe('hintReason.toTableau');
  });
});
