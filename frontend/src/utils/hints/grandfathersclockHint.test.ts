import { describe, expect, it } from 'vitest';
import type { GrandfathersClockResponse } from '../../types/card';
import { getGrandfathersClockHint } from './grandfathersclockHint';

function makeState(overrides?: Partial<GrandfathersClockResponse>): GrandfathersClockResponse {
  return {
    tableau: Array.from({ length: 8 }, () => []),
    foundation: Array.from({ length: 12 }, (_, i) => ({ cards: [], targetRank: i + 1, complete: false })),
    phase: 0,
    moveCount: 0,
    canUndo: false,
    isStalemate: false,
    message: '',
    ...overrides,
  };
}

describe('getGrandfathersClockHint', () => {
  it('returns null when not in playing phase', () => {
    expect(getGrandfathersClockHint(makeState({ phase: 1 }))).toBeNull();
  });

  it('returns null in stalemate', () => {
    expect(getGrandfathersClockHint(makeState({ isStalemate: true }))).toBeNull();
  });

  it('returns null when no server hint', () => {
    expect(getGrandfathersClockHint(makeState())).toBeNull();
  });

  it('maps a clock-face hint to strong advice', () => {
    expect(getGrandfathersClockHint(makeState({ hint: { fromCol: 2, toZone: 'foundation', toIdx: 7 } }))).toEqual({
      targetAction: 'play.foundation',
      reason: 'hintReason.toFoundation',
      confidence: 'strong',
    });
  });

  it('maps a tableau hint', () => {
    const r = getGrandfathersClockHint(makeState({ hint: { fromCol: 1, toZone: 'tableau', toIdx: 4 } }));
    expect(r?.targetAction).toBe('play.tableau');
    expect(r?.reason).toBe('hintReason.toTableau');
  });
});
