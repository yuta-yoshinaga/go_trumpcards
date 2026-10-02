import { describe, expect, it } from 'vitest';
import type { Card, CardDesign, NapoleonsSquareResponse, NapoleonsSquareTableauCard } from '../types/card';
import { napoleonsSquareLegalTargets } from './napoleonsSquareLegalTargets';

const card = (design: CardDesign, value: number): Card => ({ design, value });
const tableauCard = (design: CardDesign, value: number): NapoleonsSquareTableauCard => ({
  card: card(design, value),
  faceUp: true,
});
const state = (
  cols: NapoleonsSquareTableauCard[][],
  waste: Card[] = [],
  foundation?: Card[][],
): NapoleonsSquareResponse => ({
  tableau: Array.from({ length: 12 }, (_, index) => cols[index] ?? []),
  waste,
  foundation:
    foundation ??
    Array.from({ length: 8 }, (_, index) => [
      card(['SPADE', 'CLOVER', 'HEART', 'DIAMOND'][index % 4] as CardDesign, 1),
    ]),
  stockCount: 0,
  phase: 0,
  moveCount: 0,
  canUndo: false,
  isStalemate: false,
  message: '',
});

describe('napoleonsSquareLegalTargets', () => {
  it('accepts a run onto the same-suit descending tableau card', () => {
    const targets = napoleonsSquareLegalTargets(
      state([[tableauCard('SPADE', 7), tableauCard('SPADE', 6)], [tableauCard('SPADE', 8)]]),
      { zone: 'tableau', col: 0, cardIndex: 0 },
    );
    expect(targets).toContainEqual({ zone: 'tableau', col: 1 });
  });

  it('rejects a non-consecutive tableau run', () => {
    const targets = napoleonsSquareLegalTargets(state([[tableauCard('SPADE', 7), tableauCard('SPADE', 5)]]), {
      zone: 'tableau',
      col: 0,
      cardIndex: 0,
    });
    expect(targets).toEqual([]);
  });

  it('allows a run onto an empty column', () => {
    const targets = napoleonsSquareLegalTargets(state([[tableauCard('HEART', 9), tableauCard('HEART', 8)]]), {
      zone: 'tableau',
      col: 0,
      cardIndex: 0,
    });
    expect(targets).toContainEqual({ zone: 'tableau', col: 1 });
  });

  it('allows the tableau top to its next foundation and rejects completed foundations', () => {
    const foundation = Array.from({ length: 8 }, (_, index) => [
      card(['SPADE', 'CLOVER', 'HEART', 'DIAMOND'][index % 4] as CardDesign, 1),
    ]);
    foundation[0] = [card('SPADE', 1), card('SPADE', 2)];
    const targets = napoleonsSquareLegalTargets(state([[tableauCard('SPADE', 3)]], [], foundation), {
      zone: 'tableau',
      col: 0,
      cardIndex: 0,
    });
    expect(targets).toContainEqual({ zone: 'foundation', col: 0 });

    foundation[0] = Array.from({ length: 13 }, (_, index) => card('SPADE', index + 1));
    expect(
      napoleonsSquareLegalTargets(state([[tableauCard('SPADE', 1)]], [], foundation), {
        zone: 'tableau',
        col: 0,
        cardIndex: 0,
      }),
    ).not.toContainEqual({ zone: 'foundation', col: 0 });
  });

  it('moves the waste top to tableau and foundation targets', () => {
    const targets = napoleonsSquareLegalTargets(
      state(
        [[tableauCard('SPADE', 6)]],
        [card('SPADE', 5)],
        [[card('SPADE', 1), card('SPADE', 4)], ...Array.from({ length: 7 }, () => [card('CLOVER', 1)])],
      ),
      { zone: 'waste' },
    );
    expect(targets).toContainEqual({ zone: 'tableau', col: 0 });
    expect(targets).toContainEqual({ zone: 'foundation', col: 0 });
  });
});
