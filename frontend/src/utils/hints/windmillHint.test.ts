import { describe, expect, it } from 'vitest';
import type { WindmillResponse } from '../../types/card';
import { getWindmillHint } from './windmillHint';

function makeState(overrides?: Partial<WindmillResponse>): WindmillResponse {
  return {
    sails: Array.from({ length: 8 }, () => null),
    center: [],
    corners: Array.from({ length: 4 }, () => []),
    stockCount: 95,
    waste: [],
    transferBlocked: false,
    phase: 0,
    moveCount: 0,
    canUndo: false,
    isStalemate: false,
    message: '',
    ...overrides,
  };
}

describe('getWindmillHint', () => {
  it('returns null when not in playing phase', () => {
    expect(getWindmillHint(makeState({ phase: 1 }))).toBeNull();
  });

  it('returns null in stalemate', () => {
    expect(getWindmillHint(makeState({ isStalemate: true }))).toBeNull();
  });

  it('returns null when no server hint', () => {
    expect(getWindmillHint(makeState())).toBeNull();
  });

  it('maps a centre hint', () => {
    const r = getWindmillHint(makeState({ hint: { fromZone: 'sail', fromIdx: 2, toZone: 'center', toIdx: -1 } }));
    expect(r).toEqual({ targetAction: 'play.center', reason: 'hintReason.toCenter', confidence: 'strong' });
  });

  it('maps a corner hint', () => {
    const r = getWindmillHint(makeState({ hint: { fromZone: 'waste', fromIdx: -1, toZone: 'corner', toIdx: 1 } }));
    expect(r).toEqual({ targetAction: 'play.corner', reason: 'hintReason.toCorner', confidence: 'strong' });
  });

  // Pulling a card back dismantles a finished corner, so it is called out
  // separately rather than passed off as a routine centre move.
  it('gives the corner pull-back its own reason and a softer confidence', () => {
    const r = getWindmillHint(makeState({ hint: { fromZone: 'corner', fromIdx: 0, toZone: 'center', toIdx: -1 } }));
    expect(r).toEqual({ targetAction: 'play.pullBack', reason: 'hintReason.pullBack', confidence: 'moderate' });
  });

  // Drawing is always available and says nothing about the position.
  it('downgrades a draw hint to moderate', () => {
    const r = getWindmillHint(makeState({ hint: { fromZone: 'stock', fromIdx: -1, toZone: 'waste', toIdx: -1 } }));
    expect(r).toEqual({ targetAction: 'play.draw', reason: 'hintReason.draw', confidence: 'moderate' });
  });
});
