import { describe, expect, it } from 'vitest';
import type { TerraceResponse } from '../../types/card';
import { getTerraceHint } from './terraceHint';

function makeState(overrides?: Partial<TerraceResponse>): TerraceResponse {
  return {
    reserve: [],
    tableau: Array.from({ length: 9 }, () => []),
    foundation: Array.from({ length: 8 }, () => []),
    stockCount: 84,
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

describe('getTerraceHint', () => {
  it('returns null when not in playing phase', () => {
    expect(getTerraceHint(makeState({ phase: 1 }))).toBeNull();
  });

  it('returns null in stalemate', () => {
    expect(getTerraceHint(makeState({ isStalemate: true }))).toBeNull();
  });

  it('returns null when no server hint', () => {
    expect(getTerraceHint(makeState())).toBeNull();
  });

  // Nothing reaches a foundation until the rank is fixed, so this outranks
  // everything -- including a stalemate flag and a missing server hint.
  it('asks for the base rank first', () => {
    expect(getTerraceHint(makeState({ awaitingBaseRank: true, baseRank: 0 }))).toEqual({
      targetAction: 'play.chooseBase',
      reason: 'hintReason.chooseBase',
      confidence: 'strong',
    });
  });

  it('asks for the base rank even when the board looks stuck', () => {
    const r = getTerraceHint(makeState({ awaitingBaseRank: true, baseRank: 0, isStalemate: true }));
    expect(r?.reason).toBe('hintReason.chooseBase');
  });

  // The terrace only ever reaches a foundation and never refills, so spending it
  // reads differently from an ordinary foundation move.
  it('gives a terrace move its own reason', () => {
    const r = getTerraceHint(makeState({ hint: { fromZone: 'reserve', fromIdx: -1, toZone: 'foundation', toIdx: 0 } }));
    expect(r).toEqual({ targetAction: 'play.foundation', reason: 'hintReason.fromTerrace', confidence: 'strong' });
  });

  it('maps an ordinary foundation hint', () => {
    const r = getTerraceHint(makeState({ hint: { fromZone: 'tableau', fromIdx: 2, toZone: 'foundation', toIdx: 0 } }));
    expect(r).toEqual({ targetAction: 'play.foundation', reason: 'hintReason.toFoundation', confidence: 'strong' });
  });

  it('maps a tableau hint', () => {
    const r = getTerraceHint(makeState({ hint: { fromZone: 'tableau', fromIdx: 1, toZone: 'tableau', toIdx: 4 } }));
    expect(r).toEqual({ targetAction: 'play.tableau', reason: 'hintReason.toTableau', confidence: 'strong' });
  });

  it('downgrades a draw hint to moderate', () => {
    const r = getTerraceHint(makeState({ hint: { fromZone: 'stock', fromIdx: -1, toZone: 'waste', toIdx: -1 } }));
    expect(r).toEqual({ targetAction: 'play.draw', reason: 'hintReason.draw', confidence: 'moderate' });
  });
});
