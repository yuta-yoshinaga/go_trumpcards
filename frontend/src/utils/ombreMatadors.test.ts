import { describe, expect, it } from 'vitest';
import type { Card } from '../types/card';
import { MATADOR_NAME_KEY, matadorRank } from './ombreMatadors';

const card = (design: Card['design'], value: number): Card => ({ design, value });

describe('matadorRank', () => {
  it('returns null when trump is undecided (unset or none)', () => {
    expect(matadorRank(card('SPADE', 1), -1)).toBeNull();
    expect(matadorRank(card('SPADE', 1), 0)).toBeNull();
  });

  it('identifies Spadille (♠A) as rank 1 for every trump suit', () => {
    for (const trump of [1, 2, 3, 4]) {
      expect(matadorRank(card('SPADE', 1), trump)).toBe(1);
    }
  });

  it('identifies Basto (♣A) as rank 3 for every trump suit', () => {
    for (const trump of [1, 2, 3, 4]) {
      expect(matadorRank(card('CLOVER', 1), trump)).toBe(3);
    }
  });

  it('identifies Manille as the 7 of the trump suit (all four suit branches)', () => {
    expect(matadorRank(card('SPADE', 7), 1)).toBe(2); // spade trump → ♠7
    expect(matadorRank(card('CLOVER', 7), 2)).toBe(2); // club trump → ♣7
    expect(matadorRank(card('HEART', 7), 3)).toBe(2); // heart trump → ♥7
    expect(matadorRank(card('DIAMOND', 7), 4)).toBe(2); // diamond trump → ♦7
  });

  it('uses the 7 (not the 2) as Manille even for black trump suits', () => {
    // Traditional Ombre uses the 2 for black trumps, but this codebase's Go
    // domain always uses the 7 — the badge must match the domain, not tradition.
    expect(matadorRank(card('SPADE', 2), 1)).toBeNull();
    expect(matadorRank(card('CLOVER', 2), 2)).toBeNull();
  });

  it('does not mark a 7 of a non-trump suit as Manille', () => {
    expect(matadorRank(card('HEART', 7), 1)).toBeNull(); // trump is spades
    expect(matadorRank(card('DIAMOND', 7), 2)).toBeNull(); // trump is clubs
  });

  it('returns null for ordinary cards', () => {
    expect(matadorRank(card('HEART', 12), 1)).toBeNull();
    expect(matadorRank(card('DIAMOND', 1), 1)).toBeNull(); // red Ace is not a matador
    expect(matadorRank(card('HEART', 1), 3)).toBeNull(); // trump-suit red Ace (Punto), not a matador
  });

  it('exposes an i18n name key for each rank', () => {
    expect(MATADOR_NAME_KEY[1]).toBe('matador.spadille');
    expect(MATADOR_NAME_KEY[2]).toBe('matador.manille');
    expect(MATADOR_NAME_KEY[3]).toBe('matador.basto');
  });
});
