import { describe, expect, it } from 'vitest';
import type { Card, CardDesign, MissMilliganTableauCard } from '../types/card';
import { isMissMilliganRun } from './missMilliganRuns';

const card = (design: CardDesign, value: number): Card => ({ design, value });
const column = (...cards: Card[]): MissMilliganTableauCard[] => cards.map((card) => ({ card, faceUp: true }));

describe('isMissMilliganRun', () => {
  it('accepts a descending alternating-colour suffix', () => {
    expect(isMissMilliganRun(column(card('SPADE', 9), card('HEART', 8), card('CLOVER', 7)), 0)).toBe(true);
  });

  it('rejects a suffix with a rank gap', () => {
    expect(isMissMilliganRun(column(card('SPADE', 9), card('HEART', 7)), 0)).toBe(false);
  });

  it('rejects a suffix with cards of the same colour', () => {
    expect(isMissMilliganRun(column(card('SPADE', 9), card('CLOVER', 8)), 0)).toBe(false);
  });

  it('accepts a single card', () => {
    expect(isMissMilliganRun(column(card('HEART', 9)), 0)).toBe(true);
  });

  it.each([-1, 1])('rejects an out-of-range start index (%s)', (startIndex) => {
    expect(isMissMilliganRun(column(card('HEART', 9)), startIndex)).toBe(false);
  });
});
