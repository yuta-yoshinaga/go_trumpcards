import { describe, expect, it } from 'vitest';
import type { Card, CassinoResponse } from '../../types/card';
import { getCassinoHint } from './cassinoHint';

const card = (design: string, value: number): Card => ({ design, value }) as unknown as Card;

function makeState(overrides: Partial<CassinoResponse> = {}): CassinoResponse {
  return {
    players: [
      {
        id: 0,
        isHuman: true,
        cardCount: 2,
        cards: [card('HEART', 5)],
        capturedCount: 0,
        sweepCount: 0,
        totalScore: 0,
      },
      {
        id: 1,
        isHuman: false,
        cardCount: 2,
        cards: [],
        capturedCount: 0,
        sweepCount: 0,
        totalScore: 0,
      },
    ],
    currentTurn: 0,
    tableCards: [],
    builds: [],
    lastCaptureIdx: -1,
    gameEndFlag: false,
    phase: 'playerTurn',
    config: { targetScore: 21, multiBuildEnabled: true, sweepBonusEnabled: true, cpuDifficulty: 1 },
    cpuActions: [],
    humanAction: null,
    remainingDeck: 0,
    packsDealt: 1,
    roundWinners: [],
    lastRoundDetail: null,
    message: '',
    ...overrides,
  };
}

describe('getCassinoHint', () => {
  it('returns null if game ended', () => {
    expect(getCassinoHint(makeState({ gameEndFlag: true }))).toBeNull();
  });

  it('returns null if no human found', () => {
    expect(getCassinoHint(makeState({ players: [] }))).toBeNull();
  });

  it('returns null if human has no cards', () => {
    const state = makeState({
      players: [{ id: 0, isHuman: true, cardCount: 0, cards: [], capturedCount: 0, sweepCount: 0, totalScore: 0 }],
    });
    expect(getCassinoHint(state)).toBeNull();
  });

  it('recommends take when point cards are on the table', () => {
    const state = makeState({
      tableCards: [card('SPADE', 5)], // ♠5 — point (spade)
      players: [
        {
          id: 0,
          isHuman: true,
          cardCount: 1,
          cards: [card('HEART', 5)],
          capturedCount: 0,
          sweepCount: 0,
          totalScore: 0,
        },
      ],
    });
    const hint = getCassinoHint(state);
    expect(hint?.targetAction).toBe('take');
    expect(hint?.confidence).toBe('strong');
  });

  it('recommends take (moderate) for most-cards race', () => {
    const state = makeState({
      tableCards: [card('HEART', 7), card('HEART', 8)], // both heart, non-point
      players: [
        {
          id: 0,
          isHuman: true,
          cardCount: 1,
          cards: [card('DIAMOND', 9)],
          capturedCount: 0,
          sweepCount: 0,
          totalScore: 0,
        },
      ],
    });
    const hint = getCassinoHint(state);
    expect(hint?.targetAction).toBe('take');
    expect(hint?.confidence).toBe('moderate');
  });

  it('recommends trail when there is nothing useful to take', () => {
    const state = makeState({
      tableCards: [],
      players: [
        {
          id: 0,
          isHuman: true,
          cardCount: 1,
          cards: [card('HEART', 5)],
          capturedCount: 0,
          sweepCount: 0,
          totalScore: 0,
        },
      ],
    });
    const hint = getCassinoHint(state);
    expect(hint?.targetAction).toBe('trail');
  });
});
