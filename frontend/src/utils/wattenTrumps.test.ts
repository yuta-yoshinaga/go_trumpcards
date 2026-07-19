import { describe, expect, it } from 'vitest';
import type { Card } from '../types/card';
import { wattenTrumpCards, wattenTrumpInfo } from './wattenTrumps';

const c = (design: Card['design'], value: number): Card => ({ design, value });

describe('wattenTrumpInfo', () => {
  it('classifies the three permanent trumps regardless of declaration', () => {
    expect(wattenTrumpInfo(c('HEART', 13), 0, 0)).toEqual({ category: 'max', rank: 1000 });
    expect(wattenTrumpInfo(c('DIAMOND', 13), 0, 0)).toEqual({ category: 'belli', rank: 999 });
    expect(wattenTrumpInfo(c('DIAMOND', 7), 0, 0)).toEqual({ category: 'spitz', rank: 998 });
  });

  it('classifies Schlag-rank cards, tie-broken by suit ♥>♦>♠>♣', () => {
    // Schlag = 8; ♥8 must outrank ♣8.
    const heart8 = wattenTrumpInfo(c('HEART', 8), 8, 0);
    const club8 = wattenTrumpInfo(c('CLOVER', 8), 8, 0);
    expect(heart8?.category).toBe('schlag');
    expect(club8?.category).toBe('schlag');
    expect((heart8?.rank ?? 0) > (club8?.rank ?? 0)).toBe(true);
  });

  it('classifies remaining critical-suit cards below Schlag cards', () => {
    // Critical suit = ♠ (code 1); a ♠A is a critical trump but weaker than any Schlag card.
    const spadeA = wattenTrumpInfo(c('SPADE', 1), 8, 1);
    expect(spadeA?.category).toBe('critical');
    expect(spadeA?.rank).toBeLessThan(900);
    expect(spadeA?.rank).toBeGreaterThanOrEqual(800);
  });

  it('gives permanent trumps priority over Schlag/critical matches', () => {
    // ♥K is Max even when Schlag rank is K (13) and critical suit is ♥ (3).
    expect(wattenTrumpInfo(c('HEART', 13), 13, 3)).toEqual({ category: 'max', rank: 1000 });
  });

  it('returns null for plain (non-trump) cards', () => {
    expect(wattenTrumpInfo(c('SPADE', 9), 8, 3)).toBeNull();
  });

  it('treats an unset critical suit (0 or -1) as no critical trumps', () => {
    expect(wattenTrumpInfo(c('SPADE', 9), 0, 0)).toBeNull();
    expect(wattenTrumpInfo(c('SPADE', 9), 0, -1)).toBeNull();
  });
});

describe('wattenTrumpCards', () => {
  const hand: Card[] = [
    c('SPADE', 1), // ♠A
    c('HEART', 13), // ♥K = Max
    c('SPADE', 8), // ♠8
    c('DIAMOND', 7), // ♦7 = Spitz
    c('CLOVER', 9), // ♣9 plain (not the Schlag rank, not the critical suit)
  ];

  it('returns only trumps, sorted strongest first, with indices preserved', () => {
    // Schlag = 8, critical suit = ♠ (code 1).
    const trumps = wattenTrumpCards(hand, 8, 1);
    expect(trumps.map((t) => t.category)).toEqual(['max', 'spitz', 'schlag', 'critical']);
    // ♠A is a critical trump at index 0; ♣9 stays plain and is excluded.
    expect(trumps.map((t) => t.index)).toEqual([1, 3, 2, 0]);
  });

  it('previews only permanent trumps before any declaration', () => {
    const trumps = wattenTrumpCards(hand, 0, 0);
    expect(trumps.map((t) => t.category)).toEqual(['max', 'spitz']);
  });
});
