import { describe, expect, it } from 'vitest';
import type { AgnesTableauCard, Card, CardDesign } from '../types/card';
import { agnesCanPlaceOnFoundation, agnesCanPlaceOnTableau, agnesNextFoundationMove } from './agnesMoves';

const card = (design: CardDesign, value: number): Card => ({ design, value });
const up = (design: CardDesign, value: number): AgnesTableauCard => ({ card: card(design, value), faceUp: true });
const down = (): AgnesTableauCard => ({ card: null, faceUp: false });

// Foundation piles are indexed by suit: [Spade, Clover, Heart, Diamond].
const emptyFoundation = (): Card[][] => [[], [], [], []];

describe('agnesCanPlaceOnFoundation', () => {
  it('accepts the base rank on an empty pile', () => {
    expect(agnesCanPlaceOnFoundation(card('SPADE', 5), emptyFoundation(), 5)).toBe(true);
  });

  it('rejects a non-base rank on an empty pile', () => {
    expect(agnesCanPlaceOnFoundation(card('SPADE', 6), emptyFoundation(), 5)).toBe(false);
  });

  it('builds up in the same suit', () => {
    const foundation: Card[][] = [[card('SPADE', 5)], [], [], []];
    expect(agnesCanPlaceOnFoundation(card('SPADE', 6), foundation, 5)).toBe(true);
    expect(agnesCanPlaceOnFoundation(card('HEART', 6), foundation, 5)).toBe(false);
  });

  it('wraps King to Ace', () => {
    const foundation: Card[][] = [[card('SPADE', 13)], [], [], []];
    expect(agnesCanPlaceOnFoundation(card('SPADE', 1), foundation, 5)).toBe(true);
  });
});

describe('agnesNextFoundationMove', () => {
  it('returns the first column with a foundation-eligible end card', () => {
    const foundation: Card[][] = [[card('SPADE', 5)], [], [], []];
    const tableau: AgnesTableauCard[][] = [[up('HEART', 2)], [up('SPADE', 6)]];
    expect(agnesNextFoundationMove(tableau, foundation, 5)).toBe(1);
  });

  it('ignores face-down end cards', () => {
    const foundation: Card[][] = [[card('SPADE', 5)], [], [], []];
    const tableau: AgnesTableauCard[][] = [[up('SPADE', 6), down()]];
    expect(agnesNextFoundationMove(tableau, foundation, 5)).toBe(-1);
  });

  it('returns -1 when no move exists', () => {
    const tableau: AgnesTableauCard[][] = [[up('HEART', 9)]];
    expect(agnesNextFoundationMove(tableau, emptyFoundation(), 5)).toBe(-1);
  });
});

describe('agnesCanPlaceOnTableau', () => {
  it('rejects placement on an empty column (Agnes rule: empty columns only fill by deal)', () => {
    expect(agnesCanPlaceOnTableau(card('SPADE', 6), [])).toBe(false);
    expect(agnesCanPlaceOnTableau(card('HEART', 13), [])).toBe(false);
  });

  it('accepts same color with value difference of -1', () => {
    // Spade onto Spade: both black, 6 onto 7
    expect(agnesCanPlaceOnTableau(card('SPADE', 6), [up('SPADE', 7)])).toBe(true);
    // Clover onto Spade: both black, 6 onto 7
    expect(agnesCanPlaceOnTableau(card('CLOVER', 6), [up('SPADE', 7)])).toBe(true);
    // Spade onto Clover: both black, 6 onto 7
    expect(agnesCanPlaceOnTableau(card('SPADE', 6), [up('CLOVER', 7)])).toBe(true);
    // Clover onto Clover: both black, 6 onto 7
    expect(agnesCanPlaceOnTableau(card('CLOVER', 6), [up('CLOVER', 7)])).toBe(true);
    // Heart onto Heart: both red, 4 onto 5
    expect(agnesCanPlaceOnTableau(card('HEART', 4), [up('HEART', 5)])).toBe(true);
    // Diamond onto Heart: both red, 4 onto 5
    expect(agnesCanPlaceOnTableau(card('DIAMOND', 4), [up('HEART', 5)])).toBe(true);
    // Heart onto Diamond: both red, 4 onto 5
    expect(agnesCanPlaceOnTableau(card('HEART', 4), [up('DIAMOND', 5)])).toBe(true);
    // Diamond onto Diamond: both red, 4 onto 5
    expect(agnesCanPlaceOnTableau(card('DIAMOND', 4), [up('DIAMOND', 5)])).toBe(true);
  });

  it('rejects different colors even if value difference is -1', () => {
    // Red onto Black
    expect(agnesCanPlaceOnTableau(card('HEART', 6), [up('SPADE', 7)])).toBe(false);
    expect(agnesCanPlaceOnTableau(card('DIAMOND', 6), [up('CLOVER', 7)])).toBe(false);
    // Black onto Red
    expect(agnesCanPlaceOnTableau(card('SPADE', 4), [up('HEART', 5)])).toBe(false);
    expect(agnesCanPlaceOnTableau(card('CLOVER', 4), [up('DIAMOND', 5)])).toBe(false);
  });

  it('rejects when value difference is not -1', () => {
    // Same value
    expect(agnesCanPlaceOnTableau(card('SPADE', 7), [up('SPADE', 7)])).toBe(false);
    // Value difference is -2
    expect(agnesCanPlaceOnTableau(card('SPADE', 5), [up('SPADE', 7)])).toBe(false);
    // Value difference is +1 (ascending)
    expect(agnesCanPlaceOnTableau(card('SPADE', 8), [up('SPADE', 7)])).toBe(false);
    // King onto Ace (no wrap on tableau)
    expect(agnesCanPlaceOnTableau(card('SPADE', 13), [up('SPADE', 1)])).toBe(false);
  });

  it('rejects placement when the column end card is face-down', () => {
    expect(agnesCanPlaceOnTableau(card('SPADE', 6), [down()])).toBe(false);
    expect(agnesCanPlaceOnTableau(card('SPADE', 6), [up('SPADE', 7), down()])).toBe(false);
  });
});
