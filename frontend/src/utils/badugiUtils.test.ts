import { describe, expect, it } from 'vitest';
import type { Card, CardDesign } from '../types/card';
import { isCompleteBadugiHand } from './badugiUtils';

const card = (design: CardDesign, value: number): Card => ({ design, value });

describe('isCompleteBadugiHand', () => {
  it('returns true when all 4 ranks and all 4 suits differ', () => {
    expect(isCompleteBadugiHand([card('SPADE', 1), card('HEART', 2), card('DIAMOND', 3), card('CLOVER', 4)])).toBe(
      true,
    );
  });

  it('returns false when two cards share a rank', () => {
    expect(isCompleteBadugiHand([card('SPADE', 1), card('HEART', 1), card('DIAMOND', 3), card('CLOVER', 4)])).toBe(
      false,
    );
  });

  it('returns false when two cards share a suit', () => {
    expect(isCompleteBadugiHand([card('SPADE', 1), card('SPADE', 2), card('DIAMOND', 3), card('CLOVER', 4)])).toBe(
      false,
    );
  });

  it('returns false when fewer than 4 cards are supplied', () => {
    expect(isCompleteBadugiHand([card('SPADE', 1), card('HEART', 2), card('DIAMOND', 3)])).toBe(false);
  });

  it('returns false when more than 4 cards are supplied', () => {
    expect(
      isCompleteBadugiHand([
        card('SPADE', 1),
        card('HEART', 2),
        card('DIAMOND', 3),
        card('CLOVER', 4),
        card('SPADE', 5),
      ]),
    ).toBe(false);
  });

  it('returns false when a Joker appears (Badugi is jokerless)', () => {
    expect(isCompleteBadugiHand([card('JOKER', 0), card('HEART', 2), card('DIAMOND', 3), card('CLOVER', 4)])).toBe(
      false,
    );
  });
});
