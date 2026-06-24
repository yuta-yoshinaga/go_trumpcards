import { describe, expect, it } from 'vitest';
import type { AccordionPile, Card, CardDesign } from '../types/card';
import { accordionLegalOffsets, accordionLegalTargets } from './accordionUtils';

const card = (design: CardDesign, value: number): Card => ({ design, value });
const pile = (top: Card | undefined): AccordionPile => ({ cards: top ? [top] : [], size: top ? 1 : 0 });

describe('accordionLegalTargets', () => {
  it('returns no targets when fromIdx is empty', () => {
    expect(accordionLegalTargets([pile(card('SPADE', 5)), pile(undefined)], 1)).toEqual([]);
  });

  it('returns no targets when fromIdx is 0', () => {
    expect(accordionLegalTargets([pile(card('SPADE', 5))], 0)).toEqual([]);
  });

  it('matches by suit at offset 1', () => {
    const piles = [pile(card('SPADE', 2)), pile(card('SPADE', 9))];
    expect(accordionLegalTargets(piles, 1)).toEqual([0]);
  });

  it('matches by rank at offset 3', () => {
    const piles = [pile(card('SPADE', 7)), pile(card('HEART', 2)), pile(card('CLOVER', 3)), pile(card('DIAMOND', 7))];
    expect(accordionLegalTargets(piles, 3)).toEqual([0]);
  });

  it('returns both offsets when both match', () => {
    const piles = [pile(card('SPADE', 7)), pile(card('HEART', 7)), pile(card('CLOVER', 7)), pile(card('DIAMOND', 7))];
    expect(accordionLegalTargets(piles, 3)).toEqual([2, 0]);
  });

  it('skips offset 3 when it would be negative', () => {
    const piles = [pile(card('SPADE', 2)), pile(card('SPADE', 9))];
    expect(accordionLegalTargets(piles, 1)).toEqual([0]);
  });

  it('returns nothing when neither suit nor rank matches', () => {
    const piles = [pile(card('SPADE', 2)), pile(card('HEART', 9))];
    expect(accordionLegalTargets(piles, 1)).toEqual([]);
  });

  it('skips empty target pile at valid offset', () => {
    // fromIdx=1 (SPADE 9), toIdx=0 exists but is empty → no match
    expect(accordionLegalTargets([pile(undefined), pile(card('SPADE', 9))], 1)).toEqual([]);
  });
});

describe('accordionLegalOffsets', () => {
  it('maps legal targets back to their offsets (1 and/or 3)', () => {
    const piles = [pile(card('SPADE', 7)), pile(card('HEART', 7)), pile(card('CLOVER', 7)), pile(card('DIAMOND', 7))];
    // targets [2, 0] from idx 3 → offsets [1, 3]
    expect(accordionLegalOffsets(piles, 3)).toEqual([1, 3]);
  });

  it('returns an empty array when no merge is legal', () => {
    expect(accordionLegalOffsets([pile(card('SPADE', 2)), pile(card('HEART', 9))], 1)).toEqual([]);
  });
});
