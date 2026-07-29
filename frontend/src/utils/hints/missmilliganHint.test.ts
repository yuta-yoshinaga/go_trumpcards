import { describe, expect, it } from 'vitest';
import type { MissMilliganResponse } from '../../types/card';
import { getMissMilliganHint } from './missmilliganHint';

function makeState(overrides?: Partial<MissMilliganResponse>): MissMilliganResponse {
  return {
    tableau: Array.from({ length: 8 }, () => []),
    stockCount: 96,
    foundation: Array.from({ length: 8 }, () => []),
    waived: [],
    canWaive: false,
    phase: 0,
    moveCount: 0,
    canUndo: false,
    isStalemate: false,
    message: '',
    ...overrides,
  };
}

describe('getMissMilliganHint', () => {
  it('returns null when not in playing phase', () => {
    expect(getMissMilliganHint(makeState({ phase: 1 }))).toBeNull();
  });

  it('returns null in stalemate', () => {
    expect(getMissMilliganHint(makeState({ isStalemate: true }))).toBeNull();
  });

  it('returns null when no server hint', () => {
    expect(getMissMilliganHint(makeState())).toBeNull();
  });

  // Holding cards blocks dealing and a second waive, so returning them is its
  // own piece of advice rather than a generic tableau move.
  it('gives returning the waived cards its own reason', () => {
    const r = getMissMilliganHint(
      makeState({ hint: { fromZone: 'waived', fromCol: -1, cardIndex: -1, toZone: 'tableau', toIdx: 3 } }),
    );
    expect(r).toEqual({
      targetAction: 'play.placeWaived',
      reason: 'hintReason.returnWaived',
      confidence: 'strong',
    });
  });

  it('prefers the waived reason even when the destination is a foundation', () => {
    const r = getMissMilliganHint(
      makeState({ hint: { fromZone: 'waived', fromCol: -1, cardIndex: -1, toZone: 'foundation', toIdx: 0 } }),
    );
    expect(r?.reason).toBe('hintReason.returnWaived');
  });

  it('maps a foundation hint', () => {
    const r = getMissMilliganHint(
      makeState({ hint: { fromZone: 'tableau', fromCol: 2, cardIndex: 0, toZone: 'foundation', toIdx: 0 } }),
    );
    expect(r).toEqual({ targetAction: 'play.foundation', reason: 'hintReason.toFoundation', confidence: 'strong' });
  });

  it('maps a tableau hint', () => {
    const r = getMissMilliganHint(
      makeState({ hint: { fromZone: 'tableau', fromCol: 1, cardIndex: 2, toZone: 'tableau', toIdx: 4 } }),
    );
    expect(r?.targetAction).toBe('play.tableau');
  });

  // Dealing is always available and says nothing about the position.
  it('downgrades a deal hint to moderate', () => {
    const r = getMissMilliganHint(
      makeState({ hint: { fromZone: 'stock', fromCol: -1, cardIndex: -1, toZone: 'tableau', toIdx: -1 } }),
    );
    expect(r).toEqual({ targetAction: 'play.deal', reason: 'hintReason.deal', confidence: 'moderate' });
  });
});
