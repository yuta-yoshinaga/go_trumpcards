import { describe, expect, it } from 'vitest';
import type { Card, EightOffResponse } from '../../types/card';
import { EightOffPhase } from '../../types/phases';
import { getEightOffHint } from './eightoffHint';

const S: Card['design'] = 'SPADE';
const H: Card['design'] = 'HEART';
const D: Card['design'] = 'DIAMOND';
const C: Card['design'] = 'CLOVER';

function card(design: Card['design'], value: number) {
  return { design, value };
}

function makeState(overrides: Partial<EightOffResponse> = {}): EightOffResponse {
  return {
    tableau: [
      [card(S, 7), card(S, 6)],
      [card(D, 10), card(D, 9)],
      [card(H, 3)],
      [card(C, 5)],
      [card(S, 4)],
      [card(H, 7)],
      [card(D, 6)],
      [card(C, 11)],
    ],
    freeCells: [card(S, 2), card(H, 8), card(D, 12), card(C, 3), null, null, null, null],
    foundation: [[], [], [], []],
    phase: EightOffPhase.PLAYING,
    moveCount: 0,
    canUndo: false,
    isStalemate: false,
    message: '',
    ...overrides,
  };
}

describe('getEightOffHint', () => {
  it('returns null when game is cleared', () => {
    expect(getEightOffHint(makeState({ phase: EightOffPhase.GAME_CLEAR }))).toBeNull();
  });

  it('returns null when game is over', () => {
    expect(getEightOffHint(makeState({ phase: EightOffPhase.GAME_OVER }))).toBeNull();
  });

  it('suggests moving Ace to foundation from tableau', () => {
    const state = makeState({
      tableau: [
        [card(S, 1)],
        [card(D, 10)],
        [card(H, 3)],
        [card(C, 5)],
        [card(S, 4)],
        [card(H, 7)],
        [card(D, 6)],
        [card(C, 11)],
      ],
      freeCells: [null, null, null, null, null, null, null, null],
    });
    const hint = getEightOffHint(state);
    expect(hint?.reason).toBe('frontendHint.moveToFoundation');
    expect(hint?.confidence).toBe('strong');
  });

  it('suggests moving Ace to foundation from free cell', () => {
    const state = makeState({
      freeCells: [card(S, 1), null, null, null, null, null, null, null],
    });
    const hint = getEightOffHint(state);
    expect(hint?.reason).toBe('frontendHint.moveToFoundation');
  });

  it('suggests moving next card to foundation', () => {
    const state = makeState({
      tableau: [
        [card(S, 2)],
        [card(D, 10)],
        [card(H, 3)],
        [card(C, 5)],
        [card(S, 4)],
        [card(H, 7)],
        [card(D, 6)],
        [card(C, 11)],
      ],
      foundation: [[card(S, 1)], [], [], []],
      freeCells: [null, null, null, null, null, null, null, null],
    });
    const hint = getEightOffHint(state);
    expect(hint?.reason).toBe('frontendHint.moveToFoundation');
  });

  it('warns when free cells are nearly full (>=6)', () => {
    const state = makeState({
      tableau: [
        [card(S, 7)],
        [card(H, 8)],
        [card(D, 9)],
        [card(C, 10)],
        [card(S, 11)],
        [card(H, 12)],
        [card(D, 4)],
        [card(C, 5)],
      ],
      freeCells: [card(S, 2), card(H, 3), card(D, 4), card(C, 6), card(S, 7), card(H, 9), null, null],
      foundation: [[], [], [], []],
    });
    const hint = getEightOffHint(state);
    expect(hint?.reason).toBe('frontendHint.freeCellsFilling');
    expect(hint?.confidence).toBe('strong');
  });

  it('suggests using empty column when a King is available', () => {
    const state = makeState({
      tableau: [
        [card(S, 12), card(S, 13)],
        [],
        [card(H, 3)],
        [card(C, 5)],
        [card(S, 4)],
        [card(H, 7)],
        [card(D, 6)],
        [card(C, 11)],
      ],
      freeCells: [null, null, null, null, null, null, null, null],
      foundation: [[], [], [], []],
    });
    const hint = getEightOffHint(state);
    expect(hint?.reason).toBe('frontendHint.useEmptyColumnKing');
    expect(hint?.confidence).toBe('moderate');
  });

  it('falls through to useFreeCells when no King available for empty column', () => {
    const state = makeState({
      tableau: [[card(S, 7)], [], [card(H, 3)], [card(C, 5)], [card(S, 4)], [card(H, 7)], [card(D, 6)], [card(C, 11)]],
      freeCells: [card(S, 2), null, null, null, null, null, null, null],
      foundation: [[], [], [], []],
    });
    const hint = getEightOffHint(state);
    expect(hint?.reason).toBe('frontendHint.useFreeCells');
  });

  it('returns null when no suggestion fits', () => {
    const state = makeState({
      tableau: [
        [card(S, 7)],
        [card(H, 8)],
        [card(D, 9)],
        [card(C, 10)],
        [card(S, 11)],
        [card(H, 12)],
        [card(D, 4)],
        [card(C, 5)],
      ],
      // 8 free cells all full, no foundation, no empty columns
      freeCells: [card(S, 2), card(H, 3), card(D, 4), card(C, 6), card(S, 7), card(H, 9), card(D, 11), card(C, 12)],
      foundation: [[], [], [], []],
    });
    // freeCells full (8/8) → triggers freeCellsFilling priority
    const hint = getEightOffHint(state);
    expect(hint?.reason).toBe('frontendHint.freeCellsFilling');
  });
});
