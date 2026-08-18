import { describe, expect, it } from 'vitest';
import type { Card, PenguinResponse } from '../../types/card';
import { PenguinPhase } from '../../types/phases';
import { getPenguinHint } from './penguinHint';

const S: Card['design'] = 'SPADE';
const H: Card['design'] = 'HEART';
const D: Card['design'] = 'DIAMOND';
const C: Card['design'] = 'CLOVER';

function card(design: Card['design'], value: number): Card {
  return { design, value };
}

function makeState(overrides: Partial<PenguinResponse> = {}): PenguinResponse {
  return {
    tableau: [
      [card(S, 7), card(S, 6)],
      [card(D, 10), card(D, 9)],
      [card(H, 3)],
      [card(C, 8)],
      [card(S, 4)],
      [card(H, 7)],
      [card(D, 6)],
    ],
    freeCells: [card(S, 5), card(H, 5), card(D, 5), null, null, null, null],
    foundation: [[], [card(C, 5)], [], []],
    baseRank: 5,
    maxMovableCards: 8,
    maxMovableCardsToEmptyColumn: 4,
    phase: PenguinPhase.PLAYING,
    moveCount: 0,
    canUndo: false,
    isStalemate: false,
    message: '',
    ...overrides,
  };
}

describe('getPenguinHint', () => {
  it('returns null when game is cleared', () => {
    expect(getPenguinHint(makeState({ phase: PenguinPhase.GAME_CLEAR }))).toBeNull();
  });

  it('returns null when game is over', () => {
    expect(getPenguinHint(makeState({ phase: PenguinPhase.GAME_OVER }))).toBeNull();
  });

  it('suggests moving baseRank card to foundation from tableau', () => {
    const state = makeState({
      tableau: [[card(S, 5)], [card(D, 10)], [card(H, 3)], [card(C, 8)], [card(S, 4)], [card(H, 7)], [card(D, 6)]],
      freeCells: [card(H, 5), card(D, 5), null, null, null, null, null],
      foundation: [[], [card(C, 5)], [], []],
      baseRank: 5,
    });
    const hint = getPenguinHint(state);
    expect(hint?.reason).toBe('frontendHint.moveToFoundation');
    expect(hint?.confidence).toBe('strong');
  });

  it('suggests moving baseRank card to foundation from free cell', () => {
    const state = makeState({
      freeCells: [card(S, 5), card(H, 5), card(D, 5), null, null, null, null],
      foundation: [[], [], [], []],
      baseRank: 5,
    });
    const hint = getPenguinHint(state);
    expect(hint?.reason).toBe('frontendHint.moveToFoundation');
  });

  it('suggests moving next card to foundation (wrap K->A)', () => {
    const state = makeState({
      tableau: [[card(C, 1)], [card(D, 10)], [card(H, 3)], [card(S, 8)], [card(S, 4)], [card(H, 7)], [card(D, 6)]],
      foundation: [
        [card(S, 5)],
        [
          card(C, 5),
          card(C, 6),
          card(C, 7),
          card(C, 8),
          card(C, 9),
          card(C, 10),
          card(C, 11),
          card(C, 12),
          card(C, 13),
        ],
        [],
        [],
      ],
      freeCells: [card(H, 5), card(D, 5), null, null, null, null, null],
      baseRank: 5,
    });
    const hint = getPenguinHint(state);
    expect(hint?.reason).toBe('frontendHint.moveToFoundation');
  });

  it('warns when free cells are nearly full (>=5)', () => {
    const state = makeState({
      tableau: [[card(S, 7)], [card(D, 10)], [card(H, 3)], [card(C, 8)], [card(S, 4)], [card(H, 7)], [card(D, 2)]],
      freeCells: [card(S, 8), card(H, 9), card(D, 10), card(C, 11), card(S, 12), null, null],
      foundation: [[card(S, 5)], [card(C, 5)], [card(H, 5)], [card(D, 5)]],
      baseRank: 5,
    });
    const hint = getPenguinHint(state);
    expect(hint?.reason).toBe('frontendHint.freeCellsNearlyFull');
    expect(hint?.confidence).toBe('strong');
  });

  it('suggests using empty column when prevRank card available', () => {
    const state = makeState({
      tableau: [[card(S, 7), card(S, 4)], [], [card(H, 3)], [card(C, 8)], [card(S, 12)], [card(H, 7)], [card(D, 2)]],
      freeCells: [card(S, 9), card(H, 10), null, null, null, null, null],
      foundation: [[card(S, 5)], [card(C, 5)], [card(H, 5)], [card(D, 5)]],
      baseRank: 5,
    });
    const hint = getPenguinHint(state);
    expect(hint?.reason).toBe('frontendHint.useEmptyColumn');
    expect(hint?.confidence).toBe('moderate');
  });

  it('suggests useFreeCells when cells are available', () => {
    const state = makeState({
      tableau: [[card(S, 7)], [card(H, 8)], [card(D, 9)], [card(C, 10)], [card(S, 11)], [card(H, 12)], [card(D, 3)]],
      freeCells: [card(S, 2), card(H, 13), null, null, null, null, null],
      foundation: [[card(S, 5)], [card(C, 5)], [card(H, 5)], [card(D, 5)]],
      baseRank: 5,
    });
    const hint = getPenguinHint(state);
    expect(hint?.reason).toBe('frontendHint.useFreeCells');
  });
});
