import { describe, expect, it } from 'vitest';
import type { Card, CardDesign, CribbageSquaresResponse, CribbageSquaresScore } from '../../../types/card';
import { formatCribbageSquaresState } from './cribbagesquaresFormatter';

const card = (design: CardDesign, value: number): Card => ({ design, value });
const zero = (): CribbageSquaresScore => ({ fifteens: 0, pairs: 0, runs: 0, flush: 0, nobs: 0, total: 0 });

function makeState(overrides: Partial<CribbageSquaresResponse> = {}): CribbageSquaresResponse {
  const board = Array.from({ length: 4 }, () => Array.from({ length: 4 }, () => ({ card: null as Card | null })));
  board[0][0] = { card: card('SPADE', 5) };
  return {
    board,
    currentCard: card('HEART', 10),
    starter: null,
    placedCount: 1,
    phase: 0,
    canUndo: true,
    rowScores: [0, 0, 0, 0],
    colScores: [0, 0, 0, 0],
    rowDetails: [zero(), zero(), zero(), zero()],
    colDetails: [zero(), zero(), zero(), zero()],
    totalScore: 0,
    winScore: 61,
    isWin: false,
    message: '',
    ...overrides,
  };
}

describe('formatCribbageSquaresState', () => {
  it('shows the phase and the placement count out of 16', () => {
    const out = formatCribbageSquaresState(makeState());
    expect(out).toContain('phase: PLAYING');
    expect(out).toContain('placed: 1/16');
  });

  // Saying the starter is face down beats omitting the line, which would read
  // as a rendering bug rather than a rule.
  it('marks the starter face down while it is hidden', () => {
    expect(formatCribbageSquaresState(makeState())).toContain('starter: (face down)');
  });

  it('shows the starter once it is turned', () => {
    const out = formatCribbageSquaresState(makeState({ starter: card('CLOVER', 7) }));
    expect(out).not.toContain('(face down)');
    expect(out).toMatch(/starter: \S+/);
  });

  it('renders empty cells and the card that is down', () => {
    const out = formatCribbageSquaresState(makeState());
    expect(out).toContain('..');
    expect(out).toContain('row0=0');
    expect(out).toContain('col3=0');
  });

  it('shows the total against the target', () => {
    const out = formatCribbageSquaresState(makeState({ totalScore: 44 }));
    expect(out).toContain('total: 44 / 61');
    expect(out).not.toContain('WIN');
  });

  it('marks a winning board', () => {
    const out = formatCribbageSquaresState(makeState({ totalScore: 64, isWin: true }));
    expect(out).toContain('total: 64 / 61');
    expect(out).toContain('WIN');
  });

  // Only the components that scored are listed; a row of zeros would bury them.
  it('lists only the scoring components of a hand', () => {
    const details = [{ fifteens: 4, pairs: 2, runs: 0, flush: 0, nobs: 0, total: 6 }, zero(), zero(), zero()];
    const out = formatCribbageSquaresState(makeState({ rowDetails: details, rowScores: [6, 0, 0, 0] }));
    expect(out).toContain('15s 4');
    expect(out).toContain('pairs 2');
    expect(out).not.toContain('runs 0');
  });

  it('names every scoring component it can show', () => {
    const full = { fifteens: 2, pairs: 2, runs: 3, flush: 5, nobs: 1, total: 13 };
    const out = formatCribbageSquaresState(
      makeState({ colDetails: [full, zero(), zero(), zero()], colScores: [13, 0, 0, 0] }),
    );
    expect(out).toContain('15s 2');
    expect(out).toContain('pairs 2');
    expect(out).toContain('runs 3');
    expect(out).toContain('flush 5');
    expect(out).toContain('nobs 1');
  });

  // An older payload without the detail arrays must still render the board
  // rather than throwing on an undefined lookup.
  it('survives a response with no breakdown arrays', () => {
    const state = makeState();
    const out = formatCribbageSquaresState({
      ...state,
      rowDetails: undefined as unknown as typeof state.rowDetails,
      colDetails: undefined as unknown as typeof state.colDetails,
    });
    expect(out).toContain('row0=0');
    expect(out).toContain('col0=0');
  });

  it('shows the message when there is one', () => {
    expect(formatCribbageSquaresState(makeState({ message: 'boom' }))).toContain('boom');
  });
});
