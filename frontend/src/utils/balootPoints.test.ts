import { describe, expect, it } from 'vitest';
import type { Card } from '../types/card';
import { balootCardPoints } from './balootPoints';

const card = (design: Card['design'], value: number): Card => ({ design, value });

describe('balootCardPoints', () => {
  it('matches the complete Baloot table for all cards in Sun and Hokom', () => {
    for (const design of ['SPADE', 'CLOVER', 'HEART', 'DIAMOND'] as const) {
      for (const value of [1, 7, 8, 9, 10, 11, 12, 13]) {
        const sun = ({ 1: 11, 10: 10, 11: 2, 12: 3, 13: 4 } as Record<number, number>)[value] ?? 0;
        expect(balootCardPoints({ design, value }, 1, 0)).toBe(sun);
        const hokom =
          design === 'SPADE'
            ? (({ 1: 11, 9: 14, 10: 10, 11: 20, 12: 3, 13: 4 } as Record<number, number>)[value] ?? 0)
            : sun;
        expect(balootCardPoints({ design, value }, 2, 1)).toBe(hokom);
      }
    }
  });
  it('uses the Sun table', () => {
    expect(balootCardPoints(card('SPADE', 11), 1, 0)).toBe(2);
    expect(balootCardPoints(card('SPADE', 9), 1, 0)).toBe(0);
    expect(balootCardPoints(card('SPADE', 1), 1, 0)).toBe(11);
  });

  it('replaces only the trump suit table in Hokom', () => {
    expect(balootCardPoints(card('SPADE', 11), 2, 1)).toBe(20);
    expect(balootCardPoints(card('HEART', 11), 2, 1)).toBe(2);
    expect(balootCardPoints(card('SPADE', 9), 2, 1)).toBe(14);
  });

  it('does not score an undecided mode', () => {
    expect(balootCardPoints(card('SPADE', 1), 0, 1)).toBe(11);
    expect(balootCardPoints(undefined, 2, 1)).toBe(0);
  });
});
