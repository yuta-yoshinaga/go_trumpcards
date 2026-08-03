import { describe, expect, it } from 'vitest';
import type { BraidResponse } from '../../types/card';
import { getBraidHint } from './braidHint';

function makeState(overrides?: Partial<BraidResponse>): BraidResponse {
  return {
    braid: [],
    fields: Array.from({ length: 4 }, () => null),
    helpers: Array.from({ length: 8 }, () => null),
    foundation: Array.from({ length: 8 }, () => []),
    stockCount: 71,
    waste: [],
    baseRank: 5,
    direction: 1,
    awaitingDirection: false,
    redealsLeft: 2,
    canRedeal: false,
    phase: 0,
    moveCount: 0,
    canUndo: false,
    isStalemate: false,
    message: '',
    ...overrides,
  };
}

describe('getBraidHint', () => {
  it('returns null when not in playing phase', () => {
    expect(getBraidHint(makeState({ phase: 1 }))).toBeNull();
  });

  it('returns null in stalemate', () => {
    expect(getBraidHint(makeState({ isStalemate: true }))).toBeNull();
  });

  it('returns null without a backend hint', () => {
    expect(getBraidHint(makeState())).toBeNull();
  });

  it('asks for the direction first', () => {
    expect(getBraidHint(makeState({ awaitingDirection: true }))).toEqual({
      targetAction: 'play.chooseDirection',
      reason: 'hintReason.chooseDirection',
      confidence: 'strong',
    });
  });

  // Nothing can move until the direction is fixed, so it outranks a stalemate
  // readout -- otherwise the opening position would report "no moves".
  it('the direction outranks stalemate', () => {
    const result = getBraidHint(makeState({ awaitingDirection: true, isStalemate: true }));
    expect(result?.reason).toBe('hintReason.chooseDirection');
  });

  // The four braid fields are the only thing that consumes the braid, so
  // clearing one is worth more than an ordinary foundation move.
  it('gives a braid-field move its own reason', () => {
    const result = getBraidHint(makeState({ hint: { fromZone: 'field', fromIdx: 2, toZone: 'foundation', toIdx: 0 } }));
    expect(result).toEqual({
      targetAction: 'play.foundation',
      reason: 'hintReason.fromBraidField',
      confidence: 'strong',
    });
  });

  it('reports an ordinary foundation move', () => {
    const result = getBraidHint(
      makeState({ hint: { fromZone: 'braid', fromIdx: -1, toZone: 'foundation', toIdx: 1 } }),
    );
    expect(result?.reason).toBe('hintReason.toFoundation');
  });

  it('reports a draw', () => {
    const result = getBraidHint(makeState({ hint: { fromZone: 'stock', fromIdx: -1, toZone: 'waste', toIdx: -1 } }));
    expect(result).toEqual({ targetAction: 'play.draw', reason: 'hintReason.draw', confidence: 'moderate' });
  });

  it('reports parking a card in a helper', () => {
    const result = getBraidHint(makeState({ hint: { fromZone: 'waste', fromIdx: -1, toZone: 'helper', toIdx: 3 } }));
    expect(result).toEqual({ targetAction: 'play.helper', reason: 'hintReason.toHelper', confidence: 'moderate' });
  });
});
