import { describe, expect, it } from 'vitest';
import type { Card } from '../types/card';
import {
  DIPLOMAT_FOUNDATION_SUITS,
  DIPLOMAT_FOUNDATION_TARGET,
  diplomatFoundationRequirement,
} from './diplomatFoundation';

const card = (design: Card['design'], value: number): Card => ({ design, value });

describe('diplomatFoundationRequirement', () => {
  it('maps each foundation index to the expected suit for two decks', () => {
    expect(DIPLOMAT_FOUNDATION_SUITS).toEqual([
      'SPADE',
      'CLOVER',
      'HEART',
      'DIAMOND',
      'SPADE',
      'CLOVER',
      'HEART',
      'DIAMOND',
    ]);
  });

  it('requires Ace (value 1) for empty foundations and only accepts matching suit', () => {
    for (let fIdx = 0; fIdx < 8; fIdx++) {
      const expectedSuit = DIPLOMAT_FOUNDATION_SUITS[fIdx];
      const req = diplomatFoundationRequirement(fIdx, []);
      expect(req.suit).toBe(expectedSuit);
      expect(req.nextRank).toBe(1);

      // Matches suit and value 1
      expect(req.canPlace(card(expectedSuit, 1))).toBe(true);
      // Wrong value
      expect(req.canPlace(card(expectedSuit, 2))).toBe(false);
      // Wrong suit
      const otherSuit = expectedSuit === 'SPADE' ? 'HEART' : 'SPADE';
      expect(req.canPlace(card(otherSuit, 1))).toBe(false);
      // nil card
      expect(req.canPlace(undefined)).toBe(false);
      expect(req.canPlace(null)).toBe(false);
    }
  });

  it('requires ascending ranks as cards are placed', () => {
    const spadePile: Card[] = [card('SPADE', 1), card('SPADE', 2), card('SPADE', 3)];
    const req = diplomatFoundationRequirement(0, spadePile);
    expect(req.suit).toBe('SPADE');
    expect(req.nextRank).toBe(4);
    expect(req.canPlace(card('SPADE', 4))).toBe(true);
    expect(req.canPlace(card('SPADE', 3))).toBe(false);
    expect(req.canPlace(card('HEART', 4))).toBe(false);
  });

  it('returns null nextRank and rejects all cards when foundation is complete (13 cards)', () => {
    const fullPile: Card[] = Array.from({ length: DIPLOMAT_FOUNDATION_TARGET }, (_, i) => card('CLOVER', i + 1));
    const req = diplomatFoundationRequirement(1, fullPile);
    expect(req.suit).toBe('CLOVER');
    expect(req.nextRank).toBeNull();
    expect(req.canPlace(card('CLOVER', 1))).toBe(false);
    expect(req.canPlace(card('CLOVER', 13))).toBe(false);
  });

  it('handles two foundations of the same suit independently', () => {
    const spade0Pile: Card[] = [card('SPADE', 1)]; // fIdx 0 has A
    const spade4Pile: Card[] = []; // fIdx 4 is empty

    const req0 = diplomatFoundationRequirement(0, spade0Pile);
    const req4 = diplomatFoundationRequirement(4, spade4Pile);

    expect(req0.nextRank).toBe(2);
    expect(req4.nextRank).toBe(1);

    const aceSpade = card('SPADE', 1);
    const twoSpade = card('SPADE', 2);

    expect(req0.canPlace(aceSpade)).toBe(false);
    expect(req0.canPlace(twoSpade)).toBe(true);

    expect(req4.canPlace(aceSpade)).toBe(true);
    expect(req4.canPlace(twoSpade)).toBe(false);
  });
});
