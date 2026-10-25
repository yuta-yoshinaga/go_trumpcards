import { describe, expect, it } from 'vitest';
import type { SimpleSimonResponse } from '../types/card';
import { SimpleSimonPhase } from '../types/phases';
import {
  canSimpleSimonPickUp,
  canSimpleSimonUndo,
  isSimpleSimonLegalDestination,
  isSimpleSimonRunStart,
} from './simpleSimonRules';

const state = {
  columns: [
    [
      { design: 'SPADE', value: 9 },
      { design: 'HEART', value: 8 },
    ],
    [{ design: 'DIAMOND', value: 9 }],
    [],
  ],
  runStarts: [[1], [0], []],
  phase: SimpleSimonPhase.PLAYING,
  canUndo: false,
} as unknown as SimpleSimonResponse;

describe('Simple Simon rules', () => {
  it('uses runStarts to decide which cards can be picked up', () => {
    expect(canSimpleSimonPickUp(state, 0, 1)).toBe(true);
    expect(canSimpleSimonPickUp(state, 0, 0)).toBe(false);
    expect(canSimpleSimonPickUp(null, 0, 1)).toBe(false);
  });
  it('allows legal rank destinations and empty columns, but not the source column', () => {
    expect(isSimpleSimonLegalDestination(state, 0, 1, 1)).toBe(true);
    expect(isSimpleSimonLegalDestination(state, 0, 1, 2)).toBe(true);
    expect(isSimpleSimonLegalDestination(state, 0, 1, 0)).toBe(false);
    expect(isSimpleSimonLegalDestination(state, 0, 0, 1)).toBe(false);
    expect(isSimpleSimonLegalDestination(null, 0, 1, 2)).toBe(false);
  });
  it('reports undo availability from state', () => {
    expect(canSimpleSimonUndo(state)).toBe(false);
    expect(canSimpleSimonUndo({ ...state, canUndo: true })).toBe(true);
    expect(canSimpleSimonUndo(undefined)).toBe(false);
  });

  it('treats a run start as one regardless of phase, while pick-up needs play', () => {
    const ended = { ...state, phase: SimpleSimonPhase.GAME_OVER };
    const col = ended.runStarts.findIndex((starts) => starts.length > 0);
    const idx = ended.runStarts[col][0];
    expect(isSimpleSimonRunStart(ended, col, idx)).toBe(true);
    expect(canSimpleSimonPickUp(ended, col, idx)).toBe(false);
  });
});
