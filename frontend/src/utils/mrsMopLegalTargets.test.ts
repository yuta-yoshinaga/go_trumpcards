import { describe, expect, it } from 'vitest';
import type { Card } from '../types/card';
import type { MrsMopTableauCard } from '../types/games/mrsMop';
import { mrsMopLegalTargets } from './mrsMopLegalTargets';

const card = (design: Card['design'], value: number): MrsMopTableauCard => ({ card: { design, value }, faceUp: true });

describe('mrsMopLegalTargets', () => {
  it('returns empty columns and columns whose top is one rank higher, excluding source', () => {
    const tableau = [[card('SPADE', 7)], [card('HEART', 8)], [], [card('CLOVER', 6)]];
    expect(mrsMopLegalTargets(tableau, 0, 0)).toEqual([1, 2]);
  });

  it('requires a valid movable run and a selected source', () => {
    expect(mrsMopLegalTargets([[card('SPADE', 7)], []], undefined, undefined)).toEqual([]);
    expect(mrsMopLegalTargets([[card('SPADE', 7), card('HEART', 6)], []], 0, 0)).toEqual([]);
  });
});
