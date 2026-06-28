import { describe, expect, it } from 'vitest';
import type { Card, KlondikeTableauCard } from '../types/card';
import { waspLegalTargets } from './waspUtils';

const up = (design: Card['design'], value: number): KlondikeTableauCard => ({ card: { design, value }, faceUp: true });

describe('waspLegalTargets', () => {
  it('flags same-suit, one-rank-higher top cards, ignoring wrong suit/rank', () => {
    const tableau: KlondikeTableauCard[][] = [
      [up('SPADE', 7)], // source col 0: ♠7
      [up('SPADE', 8)], // ♠8 → legal
      [up('HEART', 8)], // wrong suit
      [up('SPADE', 9)], // two higher
    ];
    expect([...waspLegalTargets(tableau, { col: 0, cardIndex: 0 })]).toEqual([1]);
  });

  it('ignores empty/source columns and undefined source fields', () => {
    const tableau: KlondikeTableauCard[][] = [[up('CLOVER', 4)], [], [up('CLOVER', 5)]];
    expect([...waspLegalTargets(tableau, { col: 0, cardIndex: 0 })]).toEqual([2]);
    expect(waspLegalTargets(tableau, { col: undefined, cardIndex: undefined }).size).toBe(0);
  });
});
