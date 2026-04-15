import { describe, expect, it } from 'vitest';
import type { Card, CardDesign, PokerSquaresResponse } from '../../../types/card';
import { PokerSquaresPhase } from '../../../types/phases';
import { formatPokerSquaresState } from './pokersquaresFormatter';

const card = (design: CardDesign, value: number): Card => ({ design, value });

const emptyBoard = () =>
  Array.from({ length: 5 }, () => Array.from({ length: 5 }, () => ({ card: null as Card | null })));

const baseState: PokerSquaresResponse = {
  board: emptyBoard(),
  currentCard: card('SPADE', 5),
  placedCount: 0,
  phase: PokerSquaresPhase.PLAYING,
  canUndo: false,
  rowScores: [0, 0, 0, 0, 0],
  colScores: [0, 0, 0, 0, 0],
  totalScore: 0,
  message: '',
};

describe('formatPokerSquaresState', () => {
  it('renders PLAYING phase header, placed count, and current card', () => {
    const result = formatPokerSquaresState(baseState);
    expect(result).toContain('Poker Squares');
    expect(result).toContain('phase: PLAYING');
    expect(result).toContain('placed: 0/25');
    expect(result).toContain('\u26605');
    expect(result).toContain('total: 0');
  });

  it('shows (none) when current card is null', () => {
    const state: PokerSquaresResponse = { ...baseState, currentCard: null };
    expect(formatPokerSquaresState(state)).toContain('current card: (none)');
  });

  it('renders board cells and row/col scores', () => {
    const board = emptyBoard();
    board[0][0] = { card: card('HEART', 13) };
    board[2][3] = { card: card('DIAMOND', 1) };
    const state: PokerSquaresResponse = {
      ...baseState,
      board,
      rowScores: [2, 0, 10, 0, 0],
      colScores: [2, 0, 0, 10, 0],
      totalScore: 24,
    };
    const result = formatPokerSquaresState(state);
    expect(result).toContain('\u2665K');
    expect(result).toContain('\u2666A');
    expect(result).toContain('row0=2');
    expect(result).toContain('row2=10');
    expect(result).toContain('col3=10');
    expect(result).toContain('total: 24');
  });

  it('formats COMPLETE phase', () => {
    const state: PokerSquaresResponse = { ...baseState, phase: PokerSquaresPhase.COMPLETE };
    expect(formatPokerSquaresState(state)).toContain('phase: COMPLETE');
  });

  it('formats unknown phase gracefully', () => {
    const state = { ...baseState, phase: 99 };
    expect(formatPokerSquaresState(state)).toContain('phase: UNKNOWN');
  });

  it('includes message when present', () => {
    const state: PokerSquaresResponse = { ...baseState, message: 'テストメッセージ' };
    expect(formatPokerSquaresState(state)).toContain('テストメッセージ');
  });

  it('omits message when empty', () => {
    const result = formatPokerSquaresState(baseState);
    expect(result).not.toContain('undefined');
    expect(result).not.toContain('null');
  });
});
