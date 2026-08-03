import { describe, expect, it } from 'vitest';
import type { DuchessResponse } from '../../types/card';
import { getDuchessHint } from './duchessHint';

function makeState(overrides?: Partial<DuchessResponse>): DuchessResponse {
  return {
    reserve: Array.from({ length: 4 }, () => []),
    tableau: Array.from({ length: 4 }, () => []),
    foundation: Array.from({ length: 4 }, () => []),
    stockCount: 35,
    waste: [],
    baseRank: 5,
    awaitingBaseRank: false,
    phase: 0,
    moveCount: 0,
    canUndo: false,
    isStalemate: false,
    message: '',
    ...overrides,
  };
}

describe('getDuchessHint', () => {
  it('returns null when not in playing phase', () => {
    expect(getDuchessHint(makeState({ phase: 1 }))).toBeNull();
  });

  it('returns null in stalemate', () => {
    expect(getDuchessHint(makeState({ isStalemate: true }))).toBeNull();
  });

  it('returns null when no server hint', () => {
    expect(getDuchessHint(makeState())).toBeNull();
  });

  // Nothing else is legal before the base rank is set, so this outranks
  // everything — including a stalemate flag and a missing server hint.
  it('asks for the base rank first', () => {
    expect(getDuchessHint(makeState({ awaitingBaseRank: true, baseRank: 0 }))).toEqual({
      targetAction: 'play.chooseBase',
      reason: 'hintReason.chooseBase',
      confidence: 'strong',
    });
  });

  it('asks for the base rank even when the board looks stuck', () => {
    const r = getDuchessHint(makeState({ awaitingBaseRank: true, baseRank: 0, isStalemate: true }));
    expect(r?.reason).toBe('hintReason.chooseBase');
  });

  it('maps a foundation hint', () => {
    const r = getDuchessHint(
      makeState({ hint: { fromZone: 'tableau', fromIdx: 2, cardIndex: 0, toZone: 'foundation', toIdx: 0 } }),
    );
    expect(r).toEqual({ targetAction: 'play.foundation', reason: 'hintReason.toFoundation', confidence: 'strong' });
  });

  it('maps a tableau hint', () => {
    const r = getDuchessHint(
      makeState({ hint: { fromZone: 'tableau', fromIdx: 1, cardIndex: 2, toZone: 'tableau', toIdx: 3 } }),
    );
    expect(r).toEqual({ targetAction: 'play.tableau', reason: 'hintReason.toTableau', confidence: 'strong' });
  });

  // Emptying the reserve is what unlocks the empty columns, so a reserve move
  // gets its own reason rather than the generic tableau one.
  it('gives a reserve move its own reason', () => {
    const r = getDuchessHint(
      makeState({ hint: { fromZone: 'reserve', fromIdx: 1, cardIndex: -1, toZone: 'tableau', toIdx: 2 } }),
    );
    expect(r).toEqual({ targetAction: 'play.tableau', reason: 'hintReason.fromReserve', confidence: 'strong' });
  });

  it('prefers the foundation reason over the reserve one', () => {
    const r = getDuchessHint(
      makeState({ hint: { fromZone: 'reserve', fromIdx: 1, cardIndex: -1, toZone: 'foundation', toIdx: 0 } }),
    );
    expect(r?.reason).toBe('hintReason.toFoundation');
  });

  // Drawing is always available and says nothing about the position.
  it('downgrades a draw hint to moderate', () => {
    const r = getDuchessHint(
      makeState({ hint: { fromZone: 'stock', fromIdx: -1, cardIndex: -1, toZone: 'waste', toIdx: -1 } }),
    );
    expect(r).toEqual({ targetAction: 'play.draw', reason: 'hintReason.draw', confidence: 'moderate' });
  });
});
