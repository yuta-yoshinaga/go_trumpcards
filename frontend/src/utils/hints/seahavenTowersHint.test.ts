import { describe, expect, it } from 'vitest';
import type { Card, SeahavenTowersResponse } from '../../types/card';
import { SeahavenTowersPhase } from '../../types/phases';
import { getSeahavenTowersHint } from './seahavenTowersHint';

const S: Card['design'] = 'SPADE';
const H: Card['design'] = 'HEART';
const D: Card['design'] = 'DIAMOND';
const C: Card['design'] = 'CLOVER';

function card(design: Card['design'], value: number) {
  return { design, value };
}

function makeState(overrides: Partial<SeahavenTowersResponse> = {}): SeahavenTowersResponse {
  return {
    tableau: [[card(S, 7), card(S, 6)], [card(D, 10), card(D, 9)], [card(C, 3)], [card(H, 5)], [], [], [], [], [], []],
    reservedCells: [null, null],
    foundation: [[], [], [], []],
    phase: SeahavenTowersPhase.PLAYING,
    moveCount: 0,
    canUndo: false,
    isStalemate: false,
    message: '',
    ...overrides,
  };
}

describe('getSeahavenTowersHint', () => {
  it('returns null when game is cleared', () => {
    expect(getSeahavenTowersHint(makeState({ phase: SeahavenTowersPhase.GAME_CLEAR }))).toBeNull();
  });

  it('returns null when game is over', () => {
    expect(getSeahavenTowersHint(makeState({ phase: SeahavenTowersPhase.GAME_OVER }))).toBeNull();
  });

  it('suggests moving Ace to foundation', () => {
    const state = makeState({
      tableau: [[card(S, 1)], [], [], [], [], [], [], [], [], []],
    });
    const hint = getSeahavenTowersHint(state);
    expect(hint?.reason).toBe('frontendHint.moveToFoundation');
    expect(hint?.confidence).toBe('strong');
  });

  it('suggests moving card to foundation from reserved cell', () => {
    const state = makeState({
      tableau: [[card(H, 7)], [], [], [], [], [], [], [], [], []],
      reservedCells: [card(S, 1), null],
    });
    const hint = getSeahavenTowersHint(state);
    expect(hint?.reason).toBe('frontendHint.moveToFoundation');
  });

  it('suggests moving next card to foundation', () => {
    const state = makeState({
      tableau: [[card(S, 2)], [], [], [], [], [], [], [], [], []],
      foundation: [[card(S, 1)], [], [], []],
    });
    const hint = getSeahavenTowersHint(state);
    expect(hint?.reason).toBe('frontendHint.moveToFoundation');
  });

  it('warns when both reserved cells are full', () => {
    const state = makeState({
      tableau: [
        [card(S, 7)],
        [card(H, 8)],
        [card(C, 4)],
        [card(D, 4)],
        [card(S, 4)],
        [card(H, 4)],
        [card(C, 2)],
        [card(D, 2)],
        [card(S, 2)],
        [card(H, 2)],
      ],
      reservedCells: [card(S, 11), card(C, 11)],
    });
    const hint = getSeahavenTowersHint(state);
    expect(hint?.reason).toBe('frontendHint.reservedFilling');
    expect(hint?.confidence).toBe('strong');
  });

  it('suggests using empty column', () => {
    const state = makeState({
      tableau: [
        [card(S, 7)],
        [card(H, 8)],
        [card(C, 4)],
        [card(D, 4)],
        [card(S, 4)],
        [card(H, 4)],
        [card(C, 2)],
        [card(D, 2)],
        [card(S, 2)],
        [],
      ],
      reservedCells: [null, null],
    });
    const hint = getSeahavenTowersHint(state);
    expect(hint?.reason).toBe('frontendHint.useEmptyColumn');
    expect(hint?.confidence).toBe('moderate');
  });

  it('suggests using reserved cell when one is free and no other moves', () => {
    const state = makeState({
      tableau: [
        [card(S, 7)],
        [card(H, 8)],
        [card(C, 4)],
        [card(D, 4)],
        [card(S, 4)],
        [card(H, 4)],
        [card(C, 2)],
        [card(D, 2)],
        [card(S, 2)],
        [card(H, 11)],
      ],
      reservedCells: [card(D, 11), null],
    });
    const hint = getSeahavenTowersHint(state);
    expect(hint?.reason).toBe('frontendHint.useReserved');
    expect(hint?.confidence).toBe('moderate');
  });
});
