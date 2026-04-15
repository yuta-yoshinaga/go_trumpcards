import { describe, expect, it } from 'vitest';
import type { Card, PokerSquaresResponse } from '../../types/card';
import { PokerSquaresPhase } from '../../types/phases';
import { getPokersquaresHint } from './pokersquaresHint';

const c = (design: Card['design'], value: number): Card => ({ design, value });

function emptyBoard(): PokerSquaresResponse['board'] {
  return Array.from({ length: 5 }, () =>
    Array.from({ length: 5 }, () => ({ card: null })),
  ) as PokerSquaresResponse['board'];
}

function buildState(overrides: Partial<PokerSquaresResponse>): PokerSquaresResponse {
  return {
    board: emptyBoard(),
    currentCard: c('SPADE', 5),
    placedCount: 0,
    phase: PokerSquaresPhase.PLAYING,
    canUndo: false,
    rowScores: [0, 0, 0, 0, 0],
    colScores: [0, 0, 0, 0, 0],
    totalScore: 0,
    message: '',
    ...overrides,
  };
}

describe('getPokersquaresHint', () => {
  it('returns null when the game is not playing', () => {
    const state = buildState({ phase: PokerSquaresPhase.COMPLETE });
    expect(getPokersquaresHint(state)).toBeNull();
  });

  it('returns null when there is no current card', () => {
    const state = buildState({ currentCard: null });
    expect(getPokersquaresHint(state)).toBeNull();
  });

  it('returns a placement for any empty cell on an empty board', () => {
    const state = buildState({});
    const hint = getPokersquaresHint(state);
    expect(hint).not.toBeNull();
    expect(hint?.targetAction).toMatch(/^cell-\d-\d$/);
    expect(hint?.reason).toBe('hint.placeAny');
  });

  it('prefers cells that create same-value synergy', () => {
    const board = emptyBoard();
    // Row 2 already has three 5s — placing another 5 there should give strong confidence.
    board[2][0] = { card: c('HEART', 5) };
    board[2][1] = { card: c('DIAMOND', 5) };
    board[2][2] = { card: c('CLOVER', 5) };
    const state = buildState({ board, currentCard: c('SPADE', 5) });
    const hint = getPokersquaresHint(state);
    expect(hint).not.toBeNull();
    expect(hint?.targetAction).toMatch(/^cell-2-(3|4)$/);
    expect(hint?.reason).toBe('hint.placeSynergy');
    expect(hint?.confidence).toBe('strong');
  });

  it('prefers cells that extend same-suit lines', () => {
    const board = emptyBoard();
    board[0][0] = { card: c('SPADE', 2) };
    board[0][1] = { card: c('SPADE', 4) };
    const state = buildState({ board, currentCard: c('SPADE', 9) });
    const hint = getPokersquaresHint(state);
    expect(hint?.targetAction).toMatch(/^cell-0-[234]$/);
    expect(hint?.reason).toBe('hint.placeSynergy');
  });

  it('skips already-filled cells', () => {
    const board = emptyBoard();
    for (let r = 0; r < 5; r++) {
      for (let col = 0; col < 5; col++) {
        if (!(r === 4 && col === 4)) {
          board[r][col] = { card: c('HEART', 2) };
        }
      }
    }
    const state = buildState({ board, currentCard: c('SPADE', 7) });
    const hint = getPokersquaresHint(state);
    expect(hint?.targetAction).toBe('cell-4-4');
  });

  it('returns null when the board is completely full', () => {
    const board = emptyBoard();
    for (let r = 0; r < 5; r++) {
      for (let col = 0; col < 5; col++) {
        board[r][col] = { card: c('HEART', 2) };
      }
    }
    const state = buildState({ board, currentCard: c('SPADE', 7) });
    expect(getPokersquaresHint(state)).toBeNull();
  });
});
