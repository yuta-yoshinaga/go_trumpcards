import { describe, expect, it } from 'vitest';
import type { Card } from '../types/card';
import {
  ADJUTANT_JOKER_VALUE,
  ADJUTANT_SUIT_BY_DESIGN,
  buildAdjutantCardRows,
  isAdjutantCardInHand,
} from './napoleonAdjutant';

describe('napoleonAdjutant', () => {
  it('maps each design to the domain adjutant-suit code', () => {
    expect(ADJUTANT_SUIT_BY_DESIGN).toEqual({ JOKER: 0, SPADE: 1, CLOVER: 2, HEART: 3, DIAMOND: 4 });
  });

  describe('buildAdjutantCardRows', () => {
    const rows = buildAdjutantCardRows();

    it('builds four 13-card suit rows plus a single joker row', () => {
      expect(rows).toHaveLength(5);
      for (const row of rows.slice(0, 4)) expect(row).toHaveLength(13);
      expect(rows[4]).toHaveLength(1);
    });

    it('uses A–K values with the correct suit code per row', () => {
      const spadeRow = rows[0];
      expect(spadeRow[0]).toEqual({ card: { design: 'SPADE', value: 1 }, suit: 1, value: 1 });
      expect(spadeRow[12]).toEqual({ card: { design: 'SPADE', value: 13 }, suit: 1, value: 13 });
      // Rows are ordered SPADE, CLOVER, HEART, DIAMOND.
      expect(rows[1][0].suit).toBe(2);
      expect(rows[2][0].suit).toBe(3);
      expect(rows[3][0].suit).toBe(4);
    });

    it('submits suit 0 / value 1 for the joker option', () => {
      const joker = rows[4][0];
      expect(joker.suit).toBe(0);
      expect(joker.value).toBe(ADJUTANT_JOKER_VALUE);
      expect(joker.card.design).toBe('JOKER');
    });
  });

  describe('isAdjutantCardInHand', () => {
    const hand: Card[] = [
      { design: 'SPADE', value: 1 },
      { design: 'HEART', value: 11 },
      { design: 'JOKER', value: 0 },
    ];
    const rows = buildAdjutantCardRows();
    const optionFor = (design: string, value: number) =>
      rows.flat().find((o) => o.card.design === design && o.value === value) ?? rows[4][0];

    it('matches a suit card held in hand by design and value', () => {
      expect(isAdjutantCardInHand(optionFor('SPADE', 1), hand)).toBe(true);
      expect(isAdjutantCardInHand(optionFor('HEART', 11), hand)).toBe(true);
    });

    it('does not match a suit card absent from hand', () => {
      expect(isAdjutantCardInHand(optionFor('SPADE', 2), hand)).toBe(false);
      expect(isAdjutantCardInHand(optionFor('DIAMOND', 13), hand)).toBe(false);
    });

    it('matches the joker option against any held joker regardless of value', () => {
      expect(isAdjutantCardInHand(rows[4][0], hand)).toBe(true);
      expect(isAdjutantCardInHand(rows[4][0], [{ design: 'SPADE', value: 1 }])).toBe(false);
    });
  });
});
