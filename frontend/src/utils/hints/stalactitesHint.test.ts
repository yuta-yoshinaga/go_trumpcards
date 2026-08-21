import { describe, expect, it } from 'vitest';
import type { Card, StalactitesResponse } from '../../types/card';
import { StalactitesPhase } from '../../types/phases';
import { getStalactitesHint } from './stalactitesHint';

const S: Card['design'] = 'SPADE';
const H: Card['design'] = 'HEART';
const D: Card['design'] = 'DIAMOND';
const C: Card['design'] = 'CLOVER';

function card(design: Card['design'], value: number) {
  return { design, value };
}

function makeState(overrides: Partial<StalactitesResponse> = {}): StalactitesResponse {
  return {
    tableau: [[card(S, 7), card(H, 6)], [card(D, 10), card(C, 9)], [card(S, 3)], [card(H, 5)], [], [], [], []],
    baseRank: 1,
    cells: [null, null, null, null],
    foundation: [[], [], [], []],
    maxMovableCards: 80,
    maxMovableCardsToEmptyColumn: 40,
    phase: StalactitesPhase.PLAYING,
    moveCount: 0,
    canUndo: false,
    isStalemate: false,
    message: '',
    ...overrides,
  };
}

describe('getStalactitesHint', () => {
  it('returns null when game is cleared', () => {
    expect(getStalactitesHint(makeState({ phase: StalactitesPhase.GAME_CLEAR }))).toBeNull();
  });

  it('returns null when game is over', () => {
    expect(getStalactitesHint(makeState({ phase: StalactitesPhase.GAME_OVER }))).toBeNull();
  });

  it('suggests moving Ace to foundation', () => {
    const state = makeState({
      tableau: [[card(S, 1)]],
    });
    const hint = getStalactitesHint(state);
    expect(hint?.reason).toBe('frontendHint.moveToFoundation');
    expect(hint?.confidence).toBe('strong');
  });

  it('suggests moving card to foundation from free cell', () => {
    const state = makeState({
      tableau: [[card(H, 7)]],
      baseRank: 1,
      cells: [card(S, 1), null, null, null],
    });
    const hint = getStalactitesHint(state);
    expect(hint?.reason).toBe('frontendHint.moveToFoundation');
  });

  it('suggests moving next card to foundation', () => {
    const state = makeState({
      tableau: [[card(S, 2)]],
      foundation: [[card(S, 1)], [], [], []],
    });
    const hint = getStalactitesHint(state);
    expect(hint?.reason).toBe('frontendHint.moveToFoundation');
  });

  it('warns when free cells are nearly full', () => {
    const state = makeState({
      tableau: [[card(S, 7)], [card(H, 8)]],
      baseRank: 1,
      cells: [card(D, 3), card(C, 4), card(S, 5), null],
    });
    const hint = getStalactitesHint(state);
    expect(hint?.reason).toBe('frontendHint.cellsFilling');
    expect(hint?.confidence).toBe('strong');
  });

  it('suggests using empty column', () => {
    const state = makeState({
      tableau: [[card(S, 7)], []],
      baseRank: 1,
      cells: [null, null, null, null],
    });
    const hint = getStalactitesHint(state);
    expect(hint?.reason).toBe('frontendHint.useEmptyColumn');
    expect(hint?.confidence).toBe('moderate');
  });

  it('suggests using free cells as fallback', () => {
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
      baseRank: 1,
      cells: [card(S, 2), card(H, 3), null, null],
    });
    const hint = getStalactitesHint(state);
    expect(hint?.reason).toBe('frontendHint.useCells');
    expect(hint?.confidence).toBe('moderate');
  });

  it('warns when all free cells full', () => {
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
      baseRank: 1,
      cells: [card(S, 2), card(H, 3), card(D, 6), card(C, 13)],
    });
    const hint = getStalactitesHint(state);
    expect(hint?.reason).toBe('frontendHint.cellsFilling');
  });
});
