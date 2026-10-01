import { describe, expect, it } from 'vitest';
import type { Card, CardDesign, RankAndFileResponse, RankAndFileTableauCard } from '../types/card';
import { rankAndFileLegalTargetCount, rankAndFileLegalTargets } from './rankAndFileLegalTargets';

const card = (design: CardDesign, value: number): Card => ({ design, value });
const tc = (design: CardDesign, value: number, faceUp = true): RankAndFileTableauCard => ({
  card: card(design, value),
  faceUp,
});

function state(tableau: RankAndFileTableauCard[][], overrides: Partial<RankAndFileResponse> = {}): RankAndFileResponse {
  return {
    tableau,
    sequenceStarts: tableau.map((column) => column.map((_, index) => index)),
    stockCount: 0,
    waste: [],
    foundation: [[], [], [], [], [], [], [], []],
    phase: 0,
    moveCount: 0,
    canUndo: false,
    isStalemate: false,
    message: '',
    ...overrides,
  };
}

describe('rankAndFileLegalTargetCount', () => {
  it('counts empty and descending opposite-colour tableau targets for a waste card', () => {
    const current = state([[], [tc('HEART', 8)], [tc('SPADE', 8)], [tc('SPADE', 9)]], {
      waste: [card('SPADE', 7)],
    });
    expect(rankAndFileLegalTargetCount(current, { zone: 'waste' })).toBe(2);
  });

  it('counts the foundation destination for an exposed waste card', () => {
    const current = state([[]], { waste: [card('DIAMOND', 1)] });
    expect(rankAndFileLegalTargetCount(current, { zone: 'waste' })).toBe(2);
  });

  it('rejects a tableau column that cannot accept the selected card', () => {
    const current = state([[tc('SPADE', 8)], [tc('HEART', 6)], [tc('HEART', 8)]], {
      sequenceStarts: [[0], [0], [0]],
    });
    expect(rankAndFileLegalTargetCount(current, { zone: 'tableau', col: 0, cardIndex: 0 })).toBe(0);
  });

  it('excludes the source column while allowing the other empty column', () => {
    const current = state([[tc('SPADE', 8)], []]);
    expect(rankAndFileLegalTargetCount(current, { zone: 'tableau', col: 0, cardIndex: 0 })).toBe(1);
  });

  it('returns zero for an unselected or missing tableau card', () => {
    const current = state([[], []]);
    expect(rankAndFileLegalTargetCount(current, { zone: 'waste' })).toBe(0);
    expect(rankAndFileLegalTargetCount(current, { zone: 'tableau' })).toBe(0);
  });

  it('does not offer a buried card as a foundation move', () => {
    const current = state([[tc('SPADE', 1), tc('HEART', 4)], []], {
      foundation: [[], [], [], [], [], [], [], []],
    });
    expect(rankAndFileLegalTargetCount(current, { zone: 'tableau', col: 0, cardIndex: 0 })).toBe(1);
  });

  it('requires the selected card index to start a movable sequence', () => {
    const current = state([[tc('SPADE', 4), tc('HEART', 5)], []], { sequenceStarts: [[1], []] });
    expect(rankAndFileLegalTargetCount(current, { zone: 'tableau', col: 0, cardIndex: 0 })).toBe(0);
  });

  it('treats a missing sequence-start list as having no movable starts', () => {
    const current = state([[tc('SPADE', 4)], []], { sequenceStarts: [] });
    expect(rankAndFileLegalTargetCount(current, { zone: 'tableau', col: 0, cardIndex: 0 })).toBe(0);
  });

  it('does not count a face-down tableau top as a legal target', () => {
    const current = state([[tc('SPADE', 8)], [tc('HEART', 9, false)]]);
    expect(rankAndFileLegalTargetCount(current, { zone: 'tableau', col: 0, cardIndex: 0 })).toBe(0);
  });

  it('counts a tableau foundation move only for a legal exposed top card', () => {
    const current = state([[tc('SPADE', 1)], []]);
    expect(rankAndFileLegalTargetCount(current, { zone: 'tableau', col: 0, cardIndex: 0 })).toBe(2);
  });

  it('accepts diamond and club as opposite-colour suits', () => {
    const current = state([[tc('DIAMOND', 7)], [tc('CLOVER', 8)]]);
    expect(rankAndFileLegalTargetCount(current, { zone: 'tableau', col: 0, cardIndex: 0 })).toBe(1);
  });

  it('treats a missing tableau column as empty', () => {
    const tableau: RankAndFileTableauCard[][] = [];
    tableau.length = 2;
    tableau[0] = [tc('SPADE', 7)];
    const current = state(tableau, { sequenceStarts: [[0], []] });
    expect(rankAndFileLegalTargetCount(current, { zone: 'tableau', col: 0, cardIndex: 0 })).toBe(1);
  });

  it('does not count an illegal foundation move', () => {
    const current = state([[tc('SPADE', 2)], []]);
    expect(rankAndFileLegalTargetCount(current, { zone: 'tableau', col: 0, cardIndex: 0 })).toBe(1);
  });
});

describe('rankAndFileLegalTargets', () => {
  it('returns legal tableau columns and a foundation target by destination type', () => {
    const current = state([[], [tc('HEART', 8)], [tc('SPADE', 8)]], {
      waste: [card('SPADE', 7)],
      foundation: [[card('SPADE', 6)], [], [], [], [], [], [], []],
    });
    expect(rankAndFileLegalTargets(current, { zone: 'waste' })).toEqual([
      { zone: 'tableau', col: 0 },
      { zone: 'tableau', col: 1 },
      { zone: 'foundation', col: 0 },
    ]);
  });
});
