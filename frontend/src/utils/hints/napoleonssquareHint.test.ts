import { describe, expect, it } from 'vitest';
import type { NapoleonsSquareResponse } from '../../types/card';
import { getNapoleonsSquareHint } from './napoleonssquareHint';

function makeState(overrides?: Partial<NapoleonsSquareResponse>): NapoleonsSquareResponse {
  return {
    tableau: Array.from({ length: 12 }, () => []),
    stockCount: 48,
    waste: [],
    foundation: Array.from({ length: 8 }, () => []),
    phase: 0,
    moveCount: 0,
    canUndo: false,
    isStalemate: false,
    message: '',
    ...overrides,
  };
}

describe('getNapoleonsSquareHint', () => {
  it('returns null when not in playing phase', () => {
    expect(getNapoleonsSquareHint(makeState({ phase: 1 }))).toBeNull();
  });

  it('returns null in stalemate', () => {
    expect(getNapoleonsSquareHint(makeState({ isStalemate: true }))).toBeNull();
  });

  it('returns null when no server hint', () => {
    expect(getNapoleonsSquareHint(makeState())).toBeNull();
  });

  it('maps a foundation hint to strong advice', () => {
    const r = getNapoleonsSquareHint(
      makeState({ hint: { fromZone: 'tableau', fromCol: 2, cardIndex: 3, toZone: 'foundation', toCol: 0 } }),
    );
    expect(r).toEqual({ targetAction: 'play.foundation', reason: 'hintReason.toFoundation', confidence: 'strong' });
  });

  it('maps a waste-to-foundation hint to strong advice too', () => {
    const r = getNapoleonsSquareHint(
      makeState({ hint: { fromZone: 'waste', fromCol: -1, cardIndex: -1, toZone: 'foundation', toCol: 5 } }),
    );
    expect(r?.targetAction).toBe('play.foundation');
  });

  it('maps a tableau hint', () => {
    const r = getNapoleonsSquareHint(
      makeState({ hint: { fromZone: 'tableau', fromCol: 1, cardIndex: 2, toZone: 'tableau', toCol: 4 } }),
    );
    expect(r?.targetAction).toBe('play.tableau');
    expect(r?.reason).toBe('hintReason.toTableau');
    expect(r?.confidence).toBe('strong');
  });

  // Turning a card is always available and says nothing about the position, so
  // it must not read as strongly as a real move.
  it('downgrades a stock hint to moderate', () => {
    const r = getNapoleonsSquareHint(
      makeState({ hint: { fromZone: 'stock', fromCol: -1, cardIndex: -1, toZone: 'waste', toCol: -1 } }),
    );
    expect(r).toEqual({ targetAction: 'play.draw', reason: 'hintReason.draw', confidence: 'moderate' });
  });
});
