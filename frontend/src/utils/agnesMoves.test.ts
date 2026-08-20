import { describe, expect, it } from 'vitest';
import type { AgnesTableauCard, Card, CardDesign } from '../types/card';
import { agnesCanPlaceOnFoundation, agnesNextFoundationMove } from './agnesMoves';

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
