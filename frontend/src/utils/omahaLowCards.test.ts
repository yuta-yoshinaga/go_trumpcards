import { describe, expect, it } from 'vitest';
import type { Card } from '../types/card';
import { boardLowPossibility, lowCardIndexSets } from './omahaLowCards';

const c = (design: Card['design'], value: number): Card => ({ design, value });

describe('lowCardIndexSets', () => {
  const hole = [c('SPADE', 1), c('HEART', 2), c('DIAMOND', 13), c('CLOVER', 12)];
  const board = [c('SPADE', 3), c('HEART', 4), c('DIAMOND', 7), c('CLOVER', 10), c('SPADE', 9)];

  it('maps a qualifying low hand to hole and board indices', () => {
    // Low hand A,2 (hole 0,1) + 3,4,7 (board 0,1,2)
    const low = [c('SPADE', 1), c('HEART', 2), c('SPADE', 3), c('HEART', 4), c('DIAMOND', 7)];
    const { loHoleSet, loBoardSet } = lowCardIndexSets(low, hole, board);
    expect([...loHoleSet].sort()).toEqual([0, 1]);
    expect([...loBoardSet].sort()).toEqual([0, 1, 2]);
  });

  it('returns empty sets when there is no qualifying low', () => {
    const { loHoleSet, loBoardSet } = lowCardIndexSets(undefined, hole, board);
    expect(loHoleSet.size).toBe(0);
    expect(loBoardSet.size).toBe(0);
  });

  it('ignores cards not present in the hole or board', () => {
    const low = [c('SPADE', 5)]; // not in hole/board
    const { loHoleSet, loBoardSet } = lowCardIndexSets(low, hole, board);
    expect(loHoleSet.size).toBe(0);
    expect(loBoardSet.size).toBe(0);
  });
});

describe('boardLowPossibility', () => {
  it('reports possible on an empty (pre-flop) board', () => {
    const result = boardLowPossibility([]);
    expect(result.status).toBe('possible');
    expect(result.lowRankCount).toBe(0);
    expect(result.needed).toBe(3);
  });

  it('reports live once the board shows 3 distinct low ranks', () => {
    // Flop 2,4,7 — all ≤ 8 and distinct.
    const board = [c('SPADE', 2), c('HEART', 4), c('DIAMOND', 7)];
    const result = boardLowPossibility(board);
    expect(result.status).toBe('live');
    expect(result.lowRankCount).toBe(3);
    expect(result.needed).toBe(0);
  });

  it('counts the ace as a low rank', () => {
    const board = [c('SPADE', 1), c('HEART', 5), c('DIAMOND', 8)];
    const result = boardLowPossibility(board);
    expect(result.status).toBe('live');
    expect(result.lowRankCount).toBe(3);
  });

  it('treats duplicate low ranks as a single distinct rank', () => {
    // Two 4s + one 6 → only 2 distinct low ranks, one board card still to come.
    const board = [c('SPADE', 4), c('HEART', 4), c('DIAMOND', 6), c('CLOVER', 11)];
    const result = boardLowPossibility(board);
    expect(result.status).toBe('possible');
    expect(result.lowRankCount).toBe(2);
    expect(result.needed).toBe(1);
  });

  it('reports impossible on a full board with fewer than 3 distinct low ranks', () => {
    // River: only 5 and 8 are ≤ 8 → 2 distinct low ranks, no cards left.
    const board = [c('SPADE', 5), c('HEART', 8), c('DIAMOND', 10), c('CLOVER', 12), c('SPADE', 13)];
    const result = boardLowPossibility(board);
    expect(result.status).toBe('impossible');
    expect(result.lowRankCount).toBe(2);
    expect(result.needed).toBe(1);
  });

  it('reports impossible on a full board with no low ranks at all', () => {
    const board = [c('SPADE', 9), c('HEART', 10), c('DIAMOND', 11), c('CLOVER', 12), c('SPADE', 13)];
    const result = boardLowPossibility(board);
    expect(result.status).toBe('impossible');
    expect(result.lowRankCount).toBe(0);
    expect(result.needed).toBe(3);
  });
});
