import { describe, expect, it } from 'vitest';
import type { BigBenResponse } from '../../types/card';
import { getBigBenHint } from './bigbenHint';

function makeState(overrides?: Partial<BigBenResponse>): BigBenResponse {
  return {
    tableau: Array.from({ length: 8 }, () => []),
    foundation: Array.from({ length: 12 }, (_, i) => ({ cards: [], targetRank: i + 1, complete: false })),
    phase: 0,
    moveCount: 0,
    stockCount: 52,
    canUndo: false,
    isStalemate: false,
    message: '',
    ...overrides,
  };
}

describe('getBigBenHint', () => {
  it('returns null when not in playing phase', () => {
    expect(getBigBenHint(makeState({ phase: 1 }))).toBeNull();
  });

  it('returns null in stalemate', () => {
    expect(getBigBenHint(makeState({ isStalemate: true }))).toBeNull();
  });

  it('returns null when no server hint', () => {
    expect(getBigBenHint(makeState())).toBeNull();
  });

  it('maps a clock-face hint to strong advice', () => {
    expect(
      getBigBenHint(makeState({ hint: { fromZone: 'tableau', fromCol: 2, toZone: 'foundation', toIdx: 7 } })),
    ).toEqual({
      targetAction: 'play.foundation',
      reason: 'hintReason.toFoundation',
      confidence: 'strong',
    });
  });

  it('maps a tableau hint', () => {
    const r = getBigBenHint(makeState({ hint: { fromZone: 'tableau', fromCol: 1, toZone: 'tableau', toIdx: 4 } }));
    expect(r?.targetAction).toBe('play.tableau');
    expect(r?.reason).toBe('hintReason.toTableau');
  });
});
