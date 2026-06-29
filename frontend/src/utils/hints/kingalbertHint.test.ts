import { describe, expect, it } from 'vitest';
import type { KingAlbertResponse } from '../../types/card';
import { getKingalbertHint } from './kingalbertHint';

function makeState(overrides?: Partial<KingAlbertResponse>): KingAlbertResponse {
  return {
    tableau: Array.from({ length: 9 }, () => []),
    reserve: Array.from({ length: 7 }, () => null),
    foundation: [[], [], [], []],
    phase: 0,
    moveCount: 0,
    canUndo: false,
    isStalemate: false,
    message: '',
    ...overrides,
  };
}

describe('getKingalbertHint', () => {
  it('returns null when not in playing phase', () => {
    expect(getKingalbertHint(makeState({ phase: 1 }))).toBeNull();
  });

  it('returns null in stalemate', () => {
    expect(getKingalbertHint(makeState({ isStalemate: true }))).toBeNull();
  });

  it('returns null when no server hint', () => {
    expect(getKingalbertHint(makeState())).toBeNull();
  });

  it('maps foundation hint to strong play action', () => {
    const r = getKingalbertHint(
      makeState({
        hint: { fromZone: 'tableau', fromCol: 0, cardIndex: 0, toZone: 'foundation', toCol: 1 },
      }),
    );
    expect(r).toEqual({ targetAction: 'play.foundation', reason: 'hintReason.toFoundation', confidence: 'strong' });
  });

  it('maps tableau hint', () => {
    const r = getKingalbertHint(
      makeState({
        hint: { fromZone: 'reserve', fromCol: 1, cardIndex: 0, toZone: 'tableau', toCol: 4 },
      }),
    );
    expect(r?.targetAction).toBe('play.tableau');
    expect(r?.reason).toBe('hintReason.toTableau');
  });
});
