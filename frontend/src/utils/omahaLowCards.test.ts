import { describe, expect, it } from 'vitest';
import type { Card } from '../types/card';
import { lowCardIndexSets } from './omahaLowCards';

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
