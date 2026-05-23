import { describe, expect, it } from 'vitest';
import type { Card, CardDesign } from '../types/card';
import { longestFlushSuit } from './highCardFlushUtils';

const card = (design: CardDesign, value: number): Card => ({ design, value });

describe('longestFlushSuit', () => {
  it('returns null for an empty hand', () => {
    expect(longestFlushSuit([])).toBeNull();
  });

  it('returns the only suit when hand is uniform', () => {
    expect(longestFlushSuit([card('SPADE', 1), card('SPADE', 7), card('SPADE', 13)])).toBe('SPADE');
  });

  it('returns the suit with the highest count', () => {
    expect(
      longestFlushSuit([
        card('SPADE', 1),
        card('HEART', 2),
        card('HEART', 5),
        card('HEART', 9),
        card('CLOVER', 4),
        card('DIAMOND', 11),
        card('DIAMOND', 12),
      ]),
    ).toBe('HEART');
  });

  it('uses the highest single card as a tiebreaker when counts tie', () => {
    expect(longestFlushSuit([card('SPADE', 1), card('SPADE', 5), card('HEART', 13), card('HEART', 9)])).toBe('HEART');
  });

  it('compares second card when count and highest card both tie (showdown rule)', () => {
    // SPADE: [K, 2], HEART: [K, 5] → same count, same top, HEART wins on 2nd card.
    expect(longestFlushSuit([card('SPADE', 13), card('SPADE', 2), card('HEART', 13), card('HEART', 5)])).toBe('HEART');
  });

  it('stays deterministic when every position ties — first-inserted suit wins', () => {
    // Identical [K, 5] for both suits: the first inserted (SPADE) wins, not a
    // random Map iteration. Documents the deterministic fall-through.
    expect(longestFlushSuit([card('SPADE', 13), card('SPADE', 5), card('HEART', 13), card('HEART', 5)])).toBe('SPADE');
  });

  it('handles a 7-card hand with a 4-card flush', () => {
    const hand = [
      card('SPADE', 14),
      card('SPADE', 13),
      card('SPADE', 11),
      card('SPADE', 5),
      card('HEART', 7),
      card('CLOVER', 8),
      card('DIAMOND', 9),
    ];
    expect(longestFlushSuit(hand)).toBe('SPADE');
  });
});
