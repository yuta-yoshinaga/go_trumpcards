import { describe, expect, it } from 'vitest';
import type { Card, ThirtyOneConfig, ThirtyOneResponse } from '../../types/card';
import { getThirtyOneHint } from './thirtyoneHint';

const card = (design: Card['design'], value: number): Card => ({ design, value });

const defaultConfig: ThirtyOneConfig = {
  cpuDifficulty: 1,
  initialLives: 3,
  knockThresholds: { easy: 29, normal: 27, hard: 25 },
};

function makeState(overrides: Partial<ThirtyOneResponse> = {}): ThirtyOneResponse {
  return {
    players: [
      {
        id: 0,
        isHuman: true,
        cardCount: 3,
        cards: [card('HEART', 5), card('SPADE', 3), card('CLOVER', 2)],
        lives: 3,
        score: 5,
        isEliminated: false,
      },
      { id: 1, isHuman: false, cardCount: 3, cards: [], lives: 3, score: 0, isEliminated: false },
      { id: 2, isHuman: false, cardCount: 3, cards: [], lives: 3, score: 0, isEliminated: false },
      { id: 3, isHuman: false, cardCount: 3, cards: [], lives: 3, score: 0, isEliminated: false },
    ],
    phase: 0,
    roundNumber: 1,
    currentPlayerIdx: 0,
    discardTop: card('HEART', 10),
    drawPileCount: 39,
    gameEndFlag: false,
    winnerIdx: -1,
    knockerIdx: -1,
    thirtyOneIdx: -1,
    roundWinnerIdx: -1,
    roundLosers: [],
    message: '',
    config: defaultConfig,
    ...overrides,
  };
}

describe('getThirtyOneHint', () => {
  it('returns null when game ended', () => {
    expect(getThirtyOneHint(makeState({ gameEndFlag: true }))).toBeNull();
  });

  it('returns null when not the human turn', () => {
    expect(getThirtyOneHint(makeState({ currentPlayerIdx: 1 }))).toBeNull();
  });

  it('returns null outside the draw phase', () => {
    expect(getThirtyOneHint(makeState({ phase: 1 }))).toBeNull();
  });

  it('recommends knocking with a strong hand', () => {
    const state = makeState({
      players: [
        {
          id: 0,
          isHuman: true,
          cardCount: 3,
          cards: [card('SPADE', 1), card('SPADE', 13), card('SPADE', 7)], // 11+10+7 = 28
          lives: 3,
          score: 28,
          isEliminated: false,
        },
        { id: 1, isHuman: false, cardCount: 3, cards: [], lives: 3, score: 0, isEliminated: false },
        { id: 2, isHuman: false, cardCount: 3, cards: [], lives: 3, score: 0, isEliminated: false },
        { id: 3, isHuman: false, cardCount: 3, cards: [], lives: 3, score: 0, isEliminated: false },
      ],
    });
    const hint = getThirtyOneHint(state);
    expect(hint?.targetAction).toBe('knock');
    expect(hint?.confidence).toBe('strong');
  });

  it('recommends drawing the discard when it improves the best suit', () => {
    // Hand best suit = HEART 5; discard top HEART 10 lifts heart total.
    const hint = getThirtyOneHint(makeState({ discardTop: card('HEART', 10) }));
    expect(hint?.targetAction).toBe('drawdiscard');
  });

  it('falls back to drawing from stock', () => {
    const hint = getThirtyOneHint(makeState({ discardTop: card('CLOVER', 2) }));
    expect(hint?.targetAction).toBe('drawstock');
  });
});
