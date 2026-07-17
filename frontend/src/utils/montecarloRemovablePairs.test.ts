import { describe, expect, it } from 'vitest';
import type { Card, CardDesign, MonteCarloBoardCell } from '../types/card';
import { countRemovablePairs } from './montecarloRemovablePairs';

const card = (design: CardDesign, value: number): Card => ({ design, value });

function emptyBoard(): MonteCarloBoardCell[][] {
  return Array.from({ length: 5 }, () => Array.from({ length: 5 }, () => ({ card: null }) as MonteCarloBoardCell));
}

describe('countRemovablePairs', () => {
  it('returns 0 for an empty board', () => {
    expect(countRemovablePairs(emptyBoard())).toBe(0);
  });

  it('counts a horizontal adjacent same-rank pair once', () => {
    const b = emptyBoard();
    b[0][0] = { card: card('SPADE', 7) };
    b[0][1] = { card: card('HEART', 7) };
    expect(countRemovablePairs(b)).toBe(1);
  });

  it('counts a diagonal adjacent same-rank pair (8-way adjacency)', () => {
    const b = emptyBoard();
    b[1][1] = { card: card('SPADE', 9) };
    b[2][2] = { card: card('DIAMOND', 9) };
    expect(countRemovablePairs(b)).toBe(1);
  });

  it('does not count non-adjacent same-rank cards', () => {
    const b = emptyBoard();
    b[0][0] = { card: card('SPADE', 5) };
    b[0][2] = { card: card('HEART', 5) };
    expect(countRemovablePairs(b)).toBe(0);
  });

  it('does not count adjacent cards of different ranks', () => {
    const b = emptyBoard();
    b[0][0] = { card: card('SPADE', 5) };
    b[0][1] = { card: card('HEART', 6) };
    expect(countRemovablePairs(b)).toBe(0);
  });

  it('counts every distinct pair around a shared cell without double counting', () => {
    // Three 7s in an L: (0,0), (0,1), (1,0). Pairs: (0,0)-(0,1), (0,0)-(1,0), (0,1)-(1,0) diagonal.
    const b = emptyBoard();
    b[0][0] = { card: card('SPADE', 7) };
    b[0][1] = { card: card('HEART', 7) };
    b[1][0] = { card: card('CLOVER', 7) };
    expect(countRemovablePairs(b)).toBe(3);
  });

  it('counts a down-left diagonal pair', () => {
    const b = emptyBoard();
    b[0][2] = { card: card('SPADE', 3) };
    b[1][1] = { card: card('HEART', 3) };
    expect(countRemovablePairs(b)).toBe(1);
  });
});
