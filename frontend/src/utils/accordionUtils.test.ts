import { describe, expect, it } from 'vitest';
import type { AccordionPile, Card, CardDesign } from '../types/card';
import { accordionLegalOffsets, accordionLegalTargets, accordionNextAutoMove } from './accordionUtils';

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

describe('accordionNextAutoMove', () => {
  it('returns null when no merge is legal', () => {
    const piles = [pile(card('SPADE', 2)), pile(card('HEART', 9))];
    expect(accordionNextAutoMove(piles)).toBeNull();
  });

  it('returns null for an empty board', () => {
    expect(accordionNextAutoMove([])).toBeNull();
  });

  it('prefers an offset-3 merge over an offset-1 merge (matches GetHint priority)', () => {
    // idx3 (SPADE 7) matches idx0 (SPADE 5) by suit at offset 3, and idx2
    // (HEART 8) at offset 1; the offset-3 move must win.
    const piles = [pile(card('SPADE', 5)), pile(card('HEART', 2)), pile(card('SPADE', 8)), pile(card('SPADE', 7))];
    expect(accordionNextAutoMove(piles)).toEqual({ fromIdx: 3, toIdx: 0 });
  });

  it('picks the left-most legal offset-3 source', () => {
    const piles = [
      pile(card('SPADE', 5)),
      pile(card('HEART', 2)),
      pile(card('CLOVER', 8)),
      pile(card('SPADE', 7)),
      pile(card('DIAMOND', 1)),
      pile(card('CLOVER', 9)),
      pile(card('CLOVER', 4)),
    ];
    // idx3→idx0 (suit SPADE) and idx6→idx3? no; idx6 CLOVER matches idx3? SPADE no.
    // Left-most offset-3 match is idx3→idx0.
    expect(accordionNextAutoMove(piles)).toEqual({ fromIdx: 3, toIdx: 0 });
  });

  it('falls back to an offset-1 merge when no offset-3 merge exists', () => {
    const piles = [pile(card('SPADE', 5)), pile(card('SPADE', 9))];
    expect(accordionNextAutoMove(piles)).toEqual({ fromIdx: 1, toIdx: 0 });
  });

  it('matches by rank as well as suit', () => {
    const piles = [pile(card('SPADE', 7)), pile(card('HEART', 7))];
    expect(accordionNextAutoMove(piles)).toEqual({ fromIdx: 1, toIdx: 0 });
  });

  it('skips piles with empty tops', () => {
    const piles = [pile(undefined), pile(card('SPADE', 9))];
    expect(accordionNextAutoMove(piles)).toBeNull();
  });
});
