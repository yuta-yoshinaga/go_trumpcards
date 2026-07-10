import { describe, expect, it } from 'vitest';
import type { StreetsAndAlleysResponse } from '../../types/card';
import { getStreetsandalleysHint } from './streetsandalleysHint';

function makeState(overrides?: Partial<StreetsAndAlleysResponse>): StreetsAndAlleysResponse {
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

describe('getStreetsandalleysHint', () => {
  it('returns null when not in playing phase', () => {
    expect(getStreetsandalleysHint(makeState({ phase: 1 }))).toBeNull();
  });

  it('returns null in stalemate', () => {
    expect(getStreetsandalleysHint(makeState({ isStalemate: true }))).toBeNull();
  });

  it('returns null when no server hint', () => {
    expect(getStreetsandalleysHint(makeState())).toBeNull();
  });

  it('maps foundation hint to strong play action', () => {
    const r = getStreetsandalleysHint(
      makeState({
        hint: { fromCol: 0, cardIndex: 0, toZone: 'foundation', toCol: 1 },
      }),
    );
    expect(r).toEqual({ targetAction: 'play.foundation', reason: 'hintReason.toFoundation', confidence: 'strong' });
  });

  it('maps tableau hint', () => {
    const r = getStreetsandalleysHint(
      makeState({
        hint: { fromCol: 1, cardIndex: 2, toZone: 'tableau', toCol: 4 },
      }),
    );
    expect(r?.targetAction).toBe('play.tableau');
    expect(r?.reason).toBe('hintReason.toTableau');
  });
});
