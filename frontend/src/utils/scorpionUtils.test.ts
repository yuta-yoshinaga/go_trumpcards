import { describe, expect, it } from 'vitest';
import type { Card, KlondikeTableauCard } from '../types/card';
import { scorpionLegalTargets } from './scorpionUtils';

const up = (design: Card['design'], value: number): KlondikeTableauCard => ({ card: { design, value }, faceUp: true });
const down = (): KlondikeTableauCard => ({ card: { design: 'SPADE', value: 5 }, faceUp: false });

describe('scorpionLegalTargets', () => {
  it('flags columns whose top card is same-suit and one rank higher', () => {
    const tableau: KlondikeTableauCard[][] = [
      [up('SPADE', 7)], // source col 0: moving a ♠7
      [up('SPADE', 8)], // ♠8 → legal (one higher, same suit)
      [up('HEART', 8)], // ♥8 → wrong suit
      [up('SPADE', 9)], // ♠9 → two higher
    ];
    const targets = scorpionLegalTargets(tableau, { col: 0, cardIndex: 0 });
    expect([...targets]).toEqual([1]);
  });

  it('ignores empty columns, the source column, and face-down tops', () => {
    const tableau: KlondikeTableauCard[][] = [
      [up('CLOVER', 4)], // source
      [], // empty
      [down()], // face-down top
      [up('CLOVER', 5)], // legal
    ];
    expect([...scorpionLegalTargets(tableau, { col: 0, cardIndex: 0 })]).toEqual([3]);
  });

  it('returns no targets when the source card is missing', () => {
    expect(scorpionLegalTargets([[]], { col: 0, cardIndex: 0 }).size).toBe(0);
  });
});
