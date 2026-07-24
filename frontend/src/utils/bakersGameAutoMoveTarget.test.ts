import { describe, expect, it } from 'vitest';
import type { Card, CardDesign } from '../types/card';
import { bakersGameAutoMoveTarget } from './bakersGameAutoMoveTarget';

const card = (design: CardDesign, value: number): Card => ({ design, value });
const emptyFreeCells: (Card | null)[] = [null, null, null, null];
const emptyFoundation: Card[][] = [[], [], [], []];

describe('bakersGameAutoMoveTarget', () => {
  it('targets the matching foundation for an Ace on an empty pile', () => {
    expect(bakersGameAutoMoveTarget(card('SPADE', 1), emptyFoundation, emptyFreeCells)).toEqual({
      zone: 'foundation',
      col: 0,
    });
  });

  it('targets the foundation for the next ascending same-suit rank', () => {
    const foundation: Card[][] = [[], [], [card('HEART', 1)], []];
    expect(bakersGameAutoMoveTarget(card('HEART', 2), foundation, emptyFreeCells)).toEqual({
      zone: 'foundation',
      col: 2,
    });
  });

  it('prefers the foundation over an empty free cell when both are legal', () => {
    // ♠A can go to foundation 0; the fallback free cell must not win.
    expect(bakersGameAutoMoveTarget(card('SPADE', 1), emptyFoundation, emptyFreeCells)).toEqual({
      zone: 'foundation',
      col: 0,
    });
  });

  it('falls back to the first empty free cell when no foundation move exists', () => {
    // ♠K on an empty spade pile is not foundation-playable.
    expect(bakersGameAutoMoveTarget(card('SPADE', 13), emptyFoundation, [card('HEART', 5), null, null, null])).toEqual({
      zone: 'freecell',
      cell: 1,
    });
  });

  it('returns null when no foundation move and no empty free cell exist', () => {
    const fullCells: (Card | null)[] = [card('HEART', 5), card('CLOVER', 6), card('DIAMOND', 7), card('SPADE', 8)];
    expect(bakersGameAutoMoveTarget(card('SPADE', 13), emptyFoundation, fullCells)).toBeNull();
  });
});
