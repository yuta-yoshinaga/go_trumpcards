import { describe, expect, it } from 'vitest';
import type { DiplomatResponse } from '../../types/card';
import { getDiplomatHint } from './diplomatHint';

function makeState(overrides?: Partial<DiplomatResponse>): DiplomatResponse {
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

describe('getDiplomatHint', () => {
  it('returns null when not in playing phase', () => {
    expect(getDiplomatHint(makeState({ phase: 1 }))).toBeNull();
  });

  it('returns null in stalemate', () => {
    expect(getDiplomatHint(makeState({ isStalemate: true }))).toBeNull();
  });

  it('returns null when no server hint', () => {
    expect(getDiplomatHint(makeState())).toBeNull();
  });

  it('maps a foundation hint', () => {
    const r = getDiplomatHint(makeState({ hint: { fromZone: 'tableau', fromIdx: 2, toZone: 'foundation', toIdx: 0 } }));
    expect(r).toEqual({ targetAction: 'play.foundation', reason: 'hintReason.toFoundation', confidence: 'strong' });
  });

  it('maps a tableau hint', () => {
    const r = getDiplomatHint(makeState({ hint: { fromZone: 'tableau', fromIdx: 1, toZone: 'tableau', toIdx: 4 } }));
    expect(r).toEqual({ targetAction: 'play.tableau', reason: 'hintReason.toTableau', confidence: 'strong' });
  });

  // Diplomat has no fill-a-gap-from-the-stock move, so anything coming from
  // the stock is just a draw.
  it('treats any stock hint as a draw', () => {
    const r = getDiplomatHint(makeState({ hint: { fromZone: 'stock', fromIdx: -1, toZone: 'tableau', toIdx: 3 } }));
    expect(r).toEqual({ targetAction: 'play.draw', reason: 'hintReason.draw', confidence: 'moderate' });
  });

  it('downgrades an ordinary draw to moderate', () => {
    const r = getDiplomatHint(makeState({ hint: { fromZone: 'stock', fromIdx: -1, toZone: 'waste', toIdx: -1 } }));
    expect(r).toEqual({ targetAction: 'play.draw', reason: 'hintReason.draw', confidence: 'moderate' });
  });

  it('prefers the foundation reason over the waste source', () => {
    const r = getDiplomatHint(makeState({ hint: { fromZone: 'waste', fromIdx: -1, toZone: 'foundation', toIdx: 1 } }));
    expect(r?.reason).toBe('hintReason.toFoundation');
  });
});
