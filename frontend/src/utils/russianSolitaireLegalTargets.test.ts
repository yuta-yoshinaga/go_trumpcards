import { describe, expect, it } from 'vitest';
import type { RussianSolitaireResponse } from '../types/card';
import { russianSolitaireLegalTargets } from './russianSolitaireLegalTargets';

const card = (design: 'SPADE' | 'CLOVER' | 'HEART' | 'DIAMOND', value: number) => ({ design, value });
const state: RussianSolitaireResponse = {
  tableau: [
    [{ card: card('SPADE', 7), faceUp: true }],
    [{ card: card('SPADE', 8), faceUp: true }],
    [],
    [{ card: card('HEART', 8), faceUp: true }],
    [{ card: card('SPADE', 13), faceUp: true }],
    [],
    [],
  ],
  foundation: [[card('SPADE', 1), card('SPADE', 6)], [], [], []],
  phase: 0,
  moveCount: 0,
  canUndo: false,
  isStalemate: false,
  message: '',
};

describe('russianSolitaireLegalTargets', () => {
  it('matches same-suit descending tableau and ascending foundation rules', () => {
    const targets = russianSolitaireLegalTargets(state, 0, 0);
    expect(targets.tableau).toEqual(new Set([1]));
    expect(targets.foundation).toEqual(new Set([0]));
  });

  it('allows only kings to empty columns and never targets the source column', () => {
    const targets = russianSolitaireLegalTargets(state, 4, 0);
    expect(targets.tableau).toEqual(new Set([2, 5, 6]));
  });

  it('does not offer foundation movement for a non-top card', () => {
    const stacked: RussianSolitaireResponse = {
      ...state,
      tableau: [
        [
          { card: card('SPADE', 6), faceUp: true },
          { card: card('SPADE', 7), faceUp: true },
        ],
        ...state.tableau.slice(1),
      ],
    };
    expect(russianSolitaireLegalTargets(stacked, 0, 0).foundation.size).toBe(0);
  });
});
