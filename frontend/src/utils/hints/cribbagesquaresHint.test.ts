import { describe, expect, it } from 'vitest';
import type { Card, CribbageSquaresResponse, CribbageSquaresScore } from '../../types/card';
import { getCribbageSquaresHint } from './cribbagesquaresHint';

const zero = (): CribbageSquaresScore => ({ fifteens: 0, pairs: 0, runs: 0, flush: 0, nobs: 0, total: 0 });

function makeState(overrides: Partial<CribbageSquaresResponse> = {}): CribbageSquaresResponse {
  return {
    board: Array.from({ length: 4 }, () => Array.from({ length: 4 }, () => ({ card: null as Card | null }))),
    currentCard: { design: 'SPADE', value: 5 },
    starter: null,
    placedCount: 0,
    phase: 0,
    canUndo: false,
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

describe('getCribbageSquaresHint', () => {
  it('returns null once the game is complete', () => {
    expect(
      getCribbageSquaresHint(makeState({ phase: 1, hint: { row: 1, col: 2, score: 4, synergy: true } })),
    ).toBeNull();
  });

  it('returns null with no card in hand', () => {
    expect(
      getCribbageSquaresHint(makeState({ currentCard: null, hint: { row: 1, col: 2, score: 4, synergy: true } })),
    ).toBeNull();
  });

  it('returns null when the server sent no hint', () => {
    expect(getCribbageSquaresHint(makeState())).toBeNull();
  });

  it('forwards the suggested cell', () => {
    expect(getCribbageSquaresHint(makeState({ hint: { row: 2, col: 3, score: 8, synergy: true } }))).toEqual({
      targetAction: 'cell-2-3',
      reason: 'hint.placeSynergy',
      confidence: 'strong',
    });
  });

  // A placement that scores nothing is still legal, so it is offered -- but it
  // must not claim the same confidence as one that actually scores.
  it('drops to moderate and a plainer reason without synergy', () => {
    expect(getCribbageSquaresHint(makeState({ hint: { row: 0, col: 0, score: 0, synergy: false } }))).toEqual({
      targetAction: 'cell-0-0',
      reason: 'hint.placeAny',
      confidence: 'moderate',
    });
  });

  it('calls a small gain moderate', () => {
    const result = getCribbageSquaresHint(makeState({ hint: { row: 1, col: 1, score: 2, synergy: true } }));
    expect(result?.confidence).toBe('moderate');
  });
});
