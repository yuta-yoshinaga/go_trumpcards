import { describe, expect, it } from 'vitest';
import type { AgnesTableauCard, Card, CardDesign } from '../types/card';
import {
  agnesCanPlaceOnFoundation,
  agnesCanPlaceOnTableau,
  agnesHasLegalMove,
  agnesNextFoundationMove,
} from './agnesMoves';

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

describe('agnesCanPlaceOnTableau', () => {
  const tableau: AgnesTableauCard[][] = [[up('SPADE', 8)], [], [up('HEART', 8)]];

  it('accepts a same-color card one rank lower', () => {
    // Clover 7 (black) onto Spade 8 (black).
    expect(agnesCanPlaceOnTableau(card('CLOVER', 7), tableau, 0)).toBe(true);
  });

  it('rejects a different-color card', () => {
    expect(agnesCanPlaceOnTableau(card('HEART', 7), tableau, 0)).toBe(false);
  });

  it('rejects an empty column', () => {
    expect(agnesCanPlaceOnTableau(card('CLOVER', 7), tableau, 1)).toBe(false);
  });

  it('rejects a wrong rank', () => {
    expect(agnesCanPlaceOnTableau(card('CLOVER', 6), tableau, 0)).toBe(false);
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

describe('agnesHasLegalMove', () => {
  it('is true while stock remains', () => {
    expect(agnesHasLegalMove([[]], emptyFoundation(), 5, 23)).toBe(true);
  });

  it('is true when a foundation move exists', () => {
    const foundation: Card[][] = [[card('SPADE', 5)], [], [], []];
    const tableau: AgnesTableauCard[][] = [[up('SPADE', 6)]];
    expect(agnesHasLegalMove(tableau, foundation, 5, 0)).toBe(true);
  });

  it('is true when a tableau-to-tableau move exists', () => {
    const tableau: AgnesTableauCard[][] = [[up('SPADE', 8)], [up('CLOVER', 7)]];
    expect(agnesHasLegalMove(tableau, emptyFoundation(), 5, 0)).toBe(true);
  });

  it('is false when stuck (no deal, no foundation, no tableau move)', () => {
    // Spade 8 and Heart 3: no foundation (base 5), no same-color one-lower stack.
    const tableau: AgnesTableauCard[][] = [[up('SPADE', 8)], [up('HEART', 3)]];
    expect(agnesHasLegalMove(tableau, emptyFoundation(), 5, 0)).toBe(false);
  });
});
