import { describe, expect, it } from 'vitest';
import type { Card, YanivResponse } from '../../types/card';
import { YanivPhase } from '../../types/phases';
import { getYanivHint, yanivCardValue } from './yanivHint';

function card(design: Card['design'], value: number): Card {
  return { design, value } as Card;
}

function baseState(overrides: Partial<YanivResponse>): YanivResponse {
  return {
    players: [
      { id: 0, isHuman: true, cardCount: 0, cards: [], score: 0, handTotal: 0, isEliminated: false },
      { id: 1, isHuman: false, cardCount: 5, cards: [], score: 0, handTotal: 0, isEliminated: false },
    ],
    phase: YanivPhase.DISCARD,
    roundNumber: 1,
    currentPlayerIdx: 0,
    pickupCards: [],
    drawPileCount: 30,
    gameEndFlag: false,
    winnerIdx: -1,
    callerIdx: -1,
    asafWinnerIdx: -1,
    isAsaf: false,
    roundScores: [],
    config: { cpuDifficulty: 1, scoreLimit: 200 },
    message: '',
    ...overrides,
  } as YanivResponse;
}

describe('yanivCardValue', () => {
  it('scores jokers as 0, aces as 1, face cards as 10', () => {
    expect(yanivCardValue(card('JOKER', 1))).toBe(0);
    expect(yanivCardValue(card('SPADE', 1))).toBe(1);
    expect(yanivCardValue(card('HEART', 7))).toBe(7);
    expect(yanivCardValue(card('CLOVER', 13))).toBe(10);
  });
});

describe('getYanivHint', () => {
  it('returns null when game has ended', () => {
    expect(getYanivHint(baseState({ gameEndFlag: true }))).toBeNull();
  });

  it('returns null when it is not the human turn', () => {
    expect(getYanivHint(baseState({ currentPlayerIdx: 1 }))).toBeNull();
  });

  it('recommends declaring Yaniv when hand total is low', () => {
    const state = baseState({
      players: [
        {
          id: 0,
          isHuman: true,
          cardCount: 2,
          cards: [card('SPADE', 1), card('HEART', 2)],
          score: 0,
          handTotal: 3,
          isEliminated: false,
        },
      ],
    });
    expect(getYanivHint(state)?.targetAction).toBe('yaniv');
  });

  it('recommends discarding when hand total is high', () => {
    const state = baseState({
      players: [
        {
          id: 0,
          isHuman: true,
          cardCount: 2,
          cards: [card('SPADE', 10), card('HEART', 9)],
          score: 0,
          handTotal: 19,
          isEliminated: false,
        },
      ],
    });
    expect(getYanivHint(state)?.targetAction).toBe('discard');
  });

  it('recommends taking a low pickup card in the draw phase', () => {
    const state = baseState({
      phase: YanivPhase.DRAW,
      pickupCards: [card('SPADE', 2), card('HEART', 9)],
    });
    expect(getYanivHint(state)?.targetAction).toBe('drawpickup');
  });

  it('recommends drawing from stock when pickup ends are high', () => {
    const state = baseState({
      phase: YanivPhase.DRAW,
      pickupCards: [card('SPADE', 9), card('HEART', 8)],
    });
    expect(getYanivHint(state)?.targetAction).toBe('drawstock');
  });

  it('returns null in round-end phase', () => {
    expect(getYanivHint(baseState({ phase: YanivPhase.ROUND_END }))).toBeNull();
  });
});
