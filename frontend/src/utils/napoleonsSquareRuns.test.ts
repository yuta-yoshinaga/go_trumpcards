import { describe, expect, it } from 'vitest';
import type { Card, CardDesign, NapoleonsSquareTableauCard } from '../types/card';
import { isNapoleonsSquareRun } from './napoleonsSquareRuns';

const card = (design: CardDesign, value: number): Card => ({ design, value });
const column = (...cards: Card[]): NapoleonsSquareTableauCard[] => cards.map((card) => ({ card, faceUp: true }));

describe('isNapoleonsSquareRun', () => {
  it('accepts a same-suit descending suffix', () => {
    expect(isNapoleonsSquareRun(column(card('SPADE', 9), card('SPADE', 8), card('SPADE', 7)), 0)).toBe(true);
  });

  it('rejects a suffix with a rank gap', () => {
    expect(isNapoleonsSquareRun(column(card('SPADE', 9), card('SPADE', 7)), 0)).toBe(false);
  });

  it('rejects a suffix with a different suit', () => {
    expect(isNapoleonsSquareRun(column(card('SPADE', 9), card('HEART', 8)), 0)).toBe(false);
  });

  it('accepts a single card', () => {
    expect(isNapoleonsSquareRun(column(card('SPADE', 9)), 0)).toBe(true);
  });
});
