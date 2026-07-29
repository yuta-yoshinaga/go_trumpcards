import { describe, expect, it } from 'vitest';
import type { CongressResponse } from '../../types/card';
import { getCongressHint } from './congressHint';

function makeState(overrides?: Partial<CongressResponse>): CongressResponse {
  return {
    tableau: Array.from({ length: 8 }, () => []),
    foundation: Array.from({ length: 8 }, () => []),
    stockCount: 96,
    waste: [],
    phase: 0,
    moveCount: 0,
    canUndo: false,
    isStalemate: false,
    message: '',
    ...overrides,
  };
}

describe('getCongressHint', () => {
  it('returns null when not in playing phase', () => {
    expect(getCongressHint(makeState({ phase: 1 }))).toBeNull();
  });

  it('returns null in stalemate', () => {
    expect(getCongressHint(makeState({ isStalemate: true }))).toBeNull();
  });

  it('returns null when no server hint', () => {
    expect(getCongressHint(makeState())).toBeNull();
  });

  it('maps a foundation hint', () => {
    const r = getCongressHint(makeState({ hint: { fromZone: 'tableau', fromIdx: 2, toZone: 'foundation', toIdx: 0 } }));
    expect(r).toEqual({ targetAction: 'play.foundation', reason: 'hintReason.toFoundation', confidence: 'strong' });
  });

  it('maps a tableau hint', () => {
    const r = getCongressHint(makeState({ hint: { fromZone: 'tableau', fromIdx: 1, toZone: 'tableau', toIdx: 4 } }));
    expect(r).toEqual({ targetAction: 'play.tableau', reason: 'hintReason.toTableau', confidence: 'strong' });
  });

  // Filling a gap straight from the stock spends a stock card without turning
  // it, which with a single pass is a real decision -- so it reads differently
  // from an ordinary draw.
  it('gives a stock gap-fill its own reason', () => {
    const r = getCongressHint(makeState({ hint: { fromZone: 'stock', fromIdx: -1, toZone: 'tableau', toIdx: 3 } }));
    expect(r).toEqual({ targetAction: 'play.fillGap', reason: 'hintReason.fillGap', confidence: 'strong' });
  });

  it('downgrades an ordinary draw to moderate', () => {
    const r = getCongressHint(makeState({ hint: { fromZone: 'stock', fromIdx: -1, toZone: 'waste', toIdx: -1 } }));
    expect(r).toEqual({ targetAction: 'play.draw', reason: 'hintReason.draw', confidence: 'moderate' });
  });

  it('prefers the foundation reason over the waste source', () => {
    const r = getCongressHint(makeState({ hint: { fromZone: 'waste', fromIdx: -1, toZone: 'foundation', toIdx: 1 } }));
    expect(r?.reason).toBe('hintReason.toFoundation');
  });
});
