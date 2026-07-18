import { describe, expect, it } from 'vitest';
import type { Card, CardDesign, CrescentTableauCard } from '../types/card';
import { crescentCanPlaceOnFoundation, crescentCanPlaceOnTableau } from './crescentTargets';

const card = (design: CardDesign, value: number): Card => ({ design, value });
const tc = (design: CardDesign, value: number): CrescentTableauCard => ({ card: card(design, value), faceUp: true });

/** Foundations pre-seeded like a real deal: 0..3 asc (A), 4..7 desc (K). */
const seededFoundation = (): Card[][] => [
  [card('SPADE', 1)],
  [card('CLOVER', 1)],
  [card('HEART', 1)],
  [card('DIAMOND', 1)],
  [card('SPADE', 13)],
  [card('CLOVER', 13)],
  [card('HEART', 13)],
  [card('DIAMOND', 13)],
];

describe('crescentCanPlaceOnFoundation', () => {
  it('accepts the next rank up on an ascending pile of the matching suit', () => {
    expect(crescentCanPlaceOnFoundation(card('SPADE', 2), seededFoundation(), 0)).toBe(true);
  });

  it('rejects a non-sequential rank on an ascending pile', () => {
    expect(crescentCanPlaceOnFoundation(card('SPADE', 4), seededFoundation(), 0)).toBe(false);
  });

  it('accepts the next rank down on a descending pile of the matching suit', () => {
    expect(crescentCanPlaceOnFoundation(card('SPADE', 12), seededFoundation(), 4)).toBe(true);
  });

  it('rejects a suit mismatch even when the rank fits', () => {
    expect(crescentCanPlaceOnFoundation(card('HEART', 2), seededFoundation(), 0)).toBe(false);
  });

  it('accepts an Ace onto an empty ascending pile and a King onto an empty descending pile', () => {
    const empty: Card[][] = [[], [], [], [], [], [], [], []];
    expect(crescentCanPlaceOnFoundation(card('SPADE', 1), empty, 0)).toBe(true);
    expect(crescentCanPlaceOnFoundation(card('SPADE', 2), empty, 0)).toBe(false);
    expect(crescentCanPlaceOnFoundation(card('SPADE', 13), empty, 4)).toBe(true);
    expect(crescentCanPlaceOnFoundation(card('SPADE', 12), empty, 4)).toBe(false);
  });

  it('rejects out-of-range indices', () => {
    expect(crescentCanPlaceOnFoundation(card('SPADE', 2), seededFoundation(), -1)).toBe(false);
    expect(crescentCanPlaceOnFoundation(card('SPADE', 2), seededFoundation(), 8)).toBe(false);
  });
});

describe('crescentCanPlaceOnTableau', () => {
  const tableau: CrescentTableauCard[][] = [[tc('SPADE', 5)], [tc('SPADE', 3)], [tc('HEART', 6)], []];

  it('accepts a same-suit card one rank higher', () => {
    expect(crescentCanPlaceOnTableau(card('SPADE', 6), tableau, 0)).toBe(true);
  });

  it('accepts a same-suit card one rank lower', () => {
    expect(crescentCanPlaceOnTableau(card('SPADE', 4), tableau, 0)).toBe(true);
  });

  it('rejects a same-suit card two ranks away', () => {
    expect(crescentCanPlaceOnTableau(card('SPADE', 7), tableau, 0)).toBe(false);
  });

  it('rejects a suit mismatch', () => {
    expect(crescentCanPlaceOnTableau(card('HEART', 4), tableau, 0)).toBe(false);
  });

  it('allows the A↔K wrap in both directions', () => {
    const wrap: CrescentTableauCard[][] = [[tc('SPADE', 13)], [tc('SPADE', 1)]];
    expect(crescentCanPlaceOnTableau(card('SPADE', 1), wrap, 0)).toBe(true);
    expect(crescentCanPlaceOnTableau(card('SPADE', 13), wrap, 1)).toBe(true);
  });

  it('rejects an empty column and out-of-range indices', () => {
    expect(crescentCanPlaceOnTableau(card('SPADE', 6), tableau, 3)).toBe(false);
    expect(crescentCanPlaceOnTableau(card('SPADE', 6), tableau, -1)).toBe(false);
    expect(crescentCanPlaceOnTableau(card('SPADE', 6), tableau, 99)).toBe(false);
  });
});
