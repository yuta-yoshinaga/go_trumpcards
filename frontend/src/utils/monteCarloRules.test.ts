import { describe, expect, it } from 'vitest';
import type { MonteCarloResponse } from '../types/card';
import {
  canMonteCarloDeal,
  canMonteCarloUndo,
  isMonteCarloAdjacent,
  listMonteCarloRemovablePairs,
} from './monteCarloRules';

const state: MonteCarloResponse = {
  board: Array.from({ length: 5 }, (_, row) =>
    Array.from({ length: 5 }, (_, col) => ({
      card: row === 0 && col < 2 ? { design: col === 0 ? ('SPADE' as const) : ('HEART' as const), value: 7 } : null,
    })),
  ),
  phase: 0,
  stockCount: 1,
  removedCount: 0,
  dealCount: 0,
  removablePairs: 1,
  canUndo: true,
  isStalemate: false,
  message: '',
  messageCode: 'montecarlo.playing',
};

describe('Monte Carlo shared rules', () => {
  it('recognizes adjacent and non-adjacent positions', () => {
    expect(isMonteCarloAdjacent({ row: 0, col: 0 }, { row: 1, col: 1 })).toBe(true);
    expect(isMonteCarloAdjacent({ row: 0, col: 0 }, { row: 0, col: 0 })).toBe(false);
    expect(isMonteCarloAdjacent({ row: 0, col: 0 }, { row: 0, col: 2 })).toBe(false);
  });

  it('lists each removable pair once using flat cell indices', () => {
    expect(listMonteCarloRemovablePairs(state)).toEqual([[0, 1]]);
    expect(
      listMonteCarloRemovablePairs({ ...state, board: state.board.map((row) => row.map(() => ({ card: null }))) }),
    ).toEqual([]);
  });

  it('reports deal and undo availability only in the playing phase', () => {
    expect(canMonteCarloDeal(state)).toBe(true);
    expect(canMonteCarloDeal({ ...state, stockCount: 0 })).toBe(false);
    expect(canMonteCarloDeal({ ...state, phase: 1 })).toBe(false);
    expect(canMonteCarloDeal(undefined)).toBe(false);
    expect(canMonteCarloUndo(state)).toBe(true);
    expect(canMonteCarloUndo({ ...state, canUndo: false })).toBe(false);
    expect(canMonteCarloUndo({ ...state, phase: 2 })).toBe(false);
    expect(canMonteCarloUndo(undefined)).toBe(false);
  });
});
