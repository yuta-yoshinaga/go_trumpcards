import { describe, expect, it } from 'vitest';
import type { Card } from '../types/card';
import type { FortyAndEightTableauCard } from '../types/games/fortyandeight';
import { fortyAndEightTableauTargets } from './fortyAndEightTableauTargets';

const card = (design: Card['design'], value: number): Card => ({ design, value });
const tcard = (design: Card['design'], value: number, faceUp = true): FortyAndEightTableauCard => ({
  card: card(design, value),
  faceUp,
});

describe('fortyAndEightTableauTargets', () => {
  it('returns an empty set when no card is selected', () => {
    const tableau: FortyAndEightTableauCard[][] = [[tcard('SPADE', 10)], []];
    expect(fortyAndEightTableauTargets(null, tableau).size).toBe(0);
    expect(fortyAndEightTableauTargets(undefined, tableau).size).toBe(0);
  });

  it('marks empty columns as eligible for any card', () => {
    const tableau: FortyAndEightTableauCard[][] = [[], [tcard('HEART', 5)], []];
    const targets = fortyAndEightTableauTargets(card('SPADE', 8), tableau);
    expect(targets.has(0)).toBe(true);
    expect(targets.has(1)).toBe(false);
    expect(targets.has(2)).toBe(true);
  });

  it('accepts a same-suit card exactly one rank below the column top card', () => {
    const tableau: FortyAndEightTableauCard[][] = [[tcard('SPADE', 10), tcard('SPADE', 9)], [tcard('HEART', 7)]];
    // SPADE 8 on SPADE 9 (top of column 0) -> legal target
    const targets = fortyAndEightTableauTargets(card('SPADE', 8), tableau);
    expect(targets.has(0)).toBe(true);
    expect(targets.has(1)).toBe(false);
  });

  it('does not accept a card with a different suit even if rank is value - 1', () => {
    // Top of col 0 is SPADE 9. CLOVER 8 is value - 1, but wrong suit.
    // HEART 8 is value - 1, opposite color, but Forty and Eight requires exact suit match.
    const tableau: FortyAndEightTableauCard[][] = [[tcard('SPADE', 9)]];
    expect(fortyAndEightTableauTargets(card('CLOVER', 8), tableau).has(0)).toBe(false);
    expect(fortyAndEightTableauTargets(card('HEART', 8), tableau).has(0)).toBe(false);
    expect(fortyAndEightTableauTargets(card('DIAMOND', 8), tableau).has(0)).toBe(false);
  });

  it('does not accept a card with value + 1 (direction is opposite of foundation)', () => {
    // Top of col 0 is SPADE 9. SPADE 10 is value + 1 (which foundation would accept, but tableau rejects).
    const tableau: FortyAndEightTableauCard[][] = [[tcard('SPADE', 9)]];
    expect(fortyAndEightTableauTargets(card('SPADE', 10), tableau).has(0)).toBe(false);
  });

  it('does not accept a card when rank difference is not -1', () => {
    const tableau: FortyAndEightTableauCard[][] = [[tcard('SPADE', 9)]];
    // Same suit, but same rank (9), rank - 2 (7), or rank - 3 (6)
    expect(fortyAndEightTableauTargets(card('SPADE', 9), tableau).has(0)).toBe(false);
    expect(fortyAndEightTableauTargets(card('SPADE', 7), tableau).has(0)).toBe(false);
    expect(fortyAndEightTableauTargets(card('SPADE', 1), tableau).has(0)).toBe(false);
  });
});
