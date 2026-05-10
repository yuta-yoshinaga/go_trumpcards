import { describe, expect, it } from 'vitest';
import type { Card, CardDesign, MonteCarloBoardCell, MonteCarloResponse } from '../../types/card';
import { MonteCarloPhase } from '../../types/phases';
import { getMonteCarloHint } from './montecarloHint';

function card(design: CardDesign, value: number): Card {
  return { design, value };
}

function emptyBoard(): MonteCarloBoardCell[][] {
  return Array.from({ length: 5 }, () => Array.from({ length: 5 }, () => ({ card: null })));
}

function makeState(overrides: Partial<MonteCarloResponse> = {}): MonteCarloResponse {
  return {
    board: emptyBoard(),
    phase: MonteCarloPhase.PLAYING,
    stockCount: 27,
    removedCount: 0,
    dealCount: 0,
    canUndo: false,
    isStalemate: false,
    message: '',
    ...overrides,
  };
}

describe('getMonteCarloHint', () => {
  it('returns null when game is cleared', () => {
    expect(getMonteCarloHint(makeState({ phase: MonteCarloPhase.GAME_CLEAR }))).toBeNull();
  });

  it('returns null when game is over', () => {
    expect(getMonteCarloHint(makeState({ phase: MonteCarloPhase.GAME_OVER }))).toBeNull();
  });

  it('returns a remove hint for an orthogonally adjacent same-rank pair', () => {
    const board = emptyBoard();
    board[0][0] = { card: card('SPADE', 7) };
    board[0][1] = { card: card('HEART', 7) };
    const hint = getMonteCarloHint(makeState({ board }));
    expect(hint).not.toBeNull();
    expect(hint?.targetAction).toBe('remove-0-0-0-1');
    expect(hint?.reason).toBe('hint.removePair');
    expect(hint?.confidence).toBe('strong');
  });

  it('returns a remove hint for a diagonally adjacent same-rank pair', () => {
    const board = emptyBoard();
    board[1][2] = { card: card('CLOVER', 9) };
    board[2][3] = { card: card('DIAMOND', 9) };
    const hint = getMonteCarloHint(makeState({ board }));
    expect(hint?.targetAction).toBe('remove-1-2-2-3');
  });

  it('does not return a remove hint for a non-adjacent same-rank pair', () => {
    const board = emptyBoard();
    board[0][0] = { card: card('SPADE', 5) };
    board[0][2] = { card: card('HEART', 5) };
    const hint = getMonteCarloHint(makeState({ board, stockCount: 0 }));
    // Cards are >1 cell apart, so the remove-pair scan finds nothing.
    // The compression-gap at (0,1) makes Deal a useful fallback.
    expect(hint?.targetAction).toBe('deal');
  });

  it('falls back to deal when no pair exists but stock has cards', () => {
    const board = emptyBoard();
    board[0][0] = { card: card('SPADE', 1) };
    board[4][4] = { card: card('HEART', 13) };
    const hint = getMonteCarloHint(makeState({ board, stockCount: 5 }));
    expect(hint?.targetAction).toBe('deal');
    expect(hint?.reason).toBe('hint.deal');
    expect(hint?.confidence).toBe('moderate');
  });

  it('falls back to deal when stock is empty but compression would change the layout', () => {
    const board = emptyBoard();
    // Gap before the filled cell triggers hasCompressionGap=true.
    board[0][1] = { card: card('SPADE', 1) };
    board[4][4] = { card: card('HEART', 13) };
    const hint = getMonteCarloHint(makeState({ board, stockCount: 0 }));
    expect(hint?.targetAction).toBe('deal');
  });

  it('returns null when stock is empty, no pair, and no compression gap', () => {
    const board = emptyBoard();
    // Both cards in the leading prefix — no nil-before-filled pattern.
    board[0][0] = { card: card('SPADE', 1) };
    board[0][1] = { card: card('HEART', 13) };
    const hint = getMonteCarloHint(makeState({ board, stockCount: 0 }));
    expect(hint).toBeNull();
  });

  it('prefers the row-major-first pair when several exist', () => {
    const board = emptyBoard();
    board[0][0] = { card: card('SPADE', 7) };
    board[0][1] = { card: card('HEART', 7) };
    board[3][3] = { card: card('CLOVER', 4) };
    board[3][4] = { card: card('DIAMOND', 4) };
    const hint = getMonteCarloHint(makeState({ board }));
    expect(hint?.targetAction).toBe('remove-0-0-0-1');
  });

  it('skips empty cells while scanning', () => {
    const board = emptyBoard();
    board[2][2] = { card: card('SPADE', 5) };
    board[2][3] = { card: card('HEART', 5) };
    const hint = getMonteCarloHint(makeState({ board }));
    expect(hint?.targetAction).toBe('remove-2-2-2-3');
  });
});
