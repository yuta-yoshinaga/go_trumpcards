import { describe, expect, it } from 'vitest';
import type { AmericanToadResponse } from '../../types/card';
import { getAmericanToadHint } from './americantoadHint';

function makeState(overrides?: Partial<AmericanToadResponse>): AmericanToadResponse {
  return {
    reserve: [],
    tableau: Array.from({ length: 8 }, () => []),
    foundation: Array.from({ length: 8 }, () => []),
    stockCount: 75,
    waste: [],
    baseRank: 5,
    passesUsed: 0,
    canRedeal: false,
    phase: 0,
    moveCount: 0,
    canUndo: false,
    isStalemate: false,
    message: '',
    ...overrides,
  };
}

describe('getAmericanToadHint', () => {
  it('returns null when not in playing phase', () => {
    expect(getAmericanToadHint(makeState({ phase: 1 }))).toBeNull();
  });

  it('returns null in stalemate', () => {
    expect(getAmericanToadHint(makeState({ isStalemate: true }))).toBeNull();
  });

  it('returns null when no server hint', () => {
    expect(getAmericanToadHint(makeState())).toBeNull();
  });

  it('maps a foundation hint', () => {
    const r = getAmericanToadHint(
      makeState({ hint: { fromZone: 'tableau', fromIdx: 2, cardIndex: 0, toZone: 'foundation', toIdx: 0 } }),
    );
    expect(r).toEqual({ targetAction: 'play.foundation', reason: 'hintReason.toFoundation', confidence: 'strong' });
  });

  it('maps a tableau hint', () => {
    const r = getAmericanToadHint(
      makeState({ hint: { fromZone: 'tableau', fromIdx: 1, cardIndex: 2, toZone: 'tableau', toIdx: 4 } }),
    );
    expect(r).toEqual({ targetAction: 'play.tableau', reason: 'hintReason.toTableau', confidence: 'strong' });
  });

  // Emptying the reserve is what unlocks the empty columns, so it reads
  // differently from an ordinary tableau move.
  it('gives a reserve move its own reason', () => {
    const r = getAmericanToadHint(
      makeState({ hint: { fromZone: 'reserve', fromIdx: -1, cardIndex: -1, toZone: 'tableau', toIdx: 2 } }),
    );
    expect(r).toEqual({ targetAction: 'play.tableau', reason: 'hintReason.fromReserve', confidence: 'strong' });
  });

  it('prefers the foundation reason over the reserve one', () => {
    const r = getAmericanToadHint(
      makeState({ hint: { fromZone: 'reserve', fromIdx: -1, cardIndex: -1, toZone: 'foundation', toIdx: 0 } }),
    );
    expect(r?.reason).toBe('hintReason.toFoundation');
  });

  it('downgrades a draw hint to moderate', () => {
    const r = getAmericanToadHint(
      makeState({ hint: { fromZone: 'stock', fromIdx: -1, cardIndex: -1, toZone: 'waste', toIdx: -1 } }),
    );
    expect(r).toEqual({ targetAction: 'play.draw', reason: 'hintReason.draw', confidence: 'moderate' });
  });

  // Spending the one redeal is not the same cheap move as an ordinary turn, so
  // it gets its own warning rather than being folded into "draw".
  it('warns separately when the draw would spend the redeal', () => {
    const r = getAmericanToadHint(
      makeState({
        canRedeal: true,
        stockCount: 0,
        hint: { fromZone: 'stock', fromIdx: -1, cardIndex: -1, toZone: 'waste', toIdx: -1 },
      }),
    );
    expect(r).toEqual({ targetAction: 'play.redeal', reason: 'hintReason.redeal', confidence: 'moderate' });
  });
});
