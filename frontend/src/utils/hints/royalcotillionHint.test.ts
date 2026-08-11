import { describe, expect, it } from 'vitest';
import type { RoyalCotillionResponse } from '../../types/card';
import { getRoyalCotillionHint } from './royalcotillionHint';

function makeState(overrides?: Partial<RoyalCotillionResponse>): RoyalCotillionResponse {
  return {
    tableau: Array.from({ length: 16 }, () => null),
    reserve: Array.from({ length: 4 }, () => []),
    foundationOdd: [true, true, true, true, false, false, false, false],
    foundation: Array.from({ length: 8 }, () => []),
    stockCount: 76,
    waste: [],
    phase: 0,
    moveCount: 0,
    canUndo: false,
    isStalemate: false,
    message: '',
    ...overrides,
  };
}

describe('getRoyalCotillionHint', () => {
  it('returns null when not in playing phase', () => {
    expect(getRoyalCotillionHint(makeState({ phase: 1 }))).toBeNull();
  });

  it('returns null in stalemate', () => {
    expect(getRoyalCotillionHint(makeState({ isStalemate: true }))).toBeNull();
  });

  it('returns null when no server hint', () => {
    expect(getRoyalCotillionHint(makeState())).toBeNull();
  });

  it('maps a foundation hint', () => {
    const r = getRoyalCotillionHint(
      makeState({ hint: { fromZone: 'tableau', fromIdx: 2, toZone: 'foundation', toIdx: 0 } }),
    );
    expect(r).toEqual({ targetAction: 'play.foundation', reason: 'hintReason.toFoundation', confidence: 'strong' });
  });

  it('maps a tableau hint', () => {
    const r = getRoyalCotillionHint(
      makeState({ hint: { fromZone: 'tableau', fromIdx: 1, toZone: 'tableau', toIdx: 4 } }),
    );
    expect(r).toEqual({ targetAction: 'play.tableau', reason: 'hintReason.toTableau', confidence: 'strong' });
  });

  // Filling a gap straight from the stock spends a stock card without turning
  // it, which with a single pass is a real decision -- so it reads differently
  // from an ordinary draw.
  it('gives a stock gap-fill its own reason', () => {
    const r = getRoyalCotillionHint(
      makeState({ hint: { fromZone: 'stock', fromIdx: -1, toZone: 'tableau', toIdx: 3 } }),
    );
    expect(r).toEqual({ targetAction: 'play.fillGap', reason: 'hintReason.fillGap', confidence: 'strong' });
  });

  it('downgrades an ordinary draw to moderate', () => {
    const r = getRoyalCotillionHint(
      makeState({ hint: { fromZone: 'stock', fromIdx: -1, toZone: 'waste', toIdx: -1 } }),
    );
    expect(r).toEqual({ targetAction: 'play.draw', reason: 'hintReason.draw', confidence: 'moderate' });
  });

  it('prefers the foundation reason over the waste source', () => {
    const r = getRoyalCotillionHint(
      makeState({ hint: { fromZone: 'waste', fromIdx: -1, toZone: 'foundation', toIdx: 1 } }),
    );
    expect(r?.reason).toBe('hintReason.toFoundation');
  });
});
