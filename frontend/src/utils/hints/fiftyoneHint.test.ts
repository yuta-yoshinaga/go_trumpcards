import { describe, expect, it } from 'vitest';
import type { Card, FiftyOneConfig, FiftyOneResponse } from '../../types/card';
import { getFiftyOneHint } from './fiftyoneHint';

const card = (design: Card['design'], value: number): Card => ({ design, value });

const defaultConfig: FiftyOneConfig = { cpuDifficulty: 1 };

function makeState(overrides: Partial<FiftyOneResponse> = {}): FiftyOneResponse {
  return {
    players: [
      {
        id: 0,
        isHuman: true,
        cardCount: 5,
        cards: [card('HEART', 10), card('HEART', 7), card('SPADE', 3), card('CLOVER', 5), card('DIAMOND', 2)],
        score: 17,
      },
      { id: 1, isHuman: false, cardCount: 5, cards: [], score: 0 },
      { id: 2, isHuman: false, cardCount: 5, cards: [], score: 0 },
      { id: 3, isHuman: false, cardCount: 5, cards: [], score: 0 },
    ],
    tableCards: [card('HEART', 8), card('SPADE', 13), card('CLOVER', 1), card('DIAMOND', 6), card('HEART', 4)],
    phase: 0,
    currentTurn: 0,
    gameEndFlag: false,
    winnerIdx: -1,
    turnNumber: 1,
    stopCallerIdx: -1,
    lastAction: '',
    lastHandIdx: -1,
    lastTableIdx: -1,
    message: '',
    config: defaultConfig,
    ...overrides,
  };
}

describe('getFiftyOneHint', () => {
  it('returns null when game ended', () => {
    expect(getFiftyOneHint(makeState({ gameEndFlag: true }))).toBeNull();
  });

  it('returns null when not human turn', () => {
    expect(getFiftyOneHint(makeState({ currentTurn: 1 }))).toBeNull();
  });

  it('returns null when no human player found', () => {
    expect(getFiftyOneHint(makeState({ players: [] }))).toBeNull();
  });

  it('suggests exchange when table has a card that improves best suit score', () => {
    // Hand: H10, H7, S3, C5, D2 => best suit = HEART (10+7=17)
    // Table: H8, S13, C1, D6, H4
    // Exchanging S3 (non-heart, value 3) for H8 (heart, value 8) would give H10+H7+H8=25
    const hint = getFiftyOneHint(makeState());
    expect(hint).not.toBeNull();
    expect(hint?.targetAction).toBe('exchange');
    expect(hint?.confidence).toBe('strong');
  });

  it('suggests stop when score is high enough', () => {
    const state = makeState({
      players: [
        {
          id: 0,
          isHuman: true,
          cardCount: 5,
          cards: [card('HEART', 10), card('HEART', 13), card('HEART', 12), card('HEART', 11), card('HEART', 1)],
          score: 47,
        },
        { id: 1, isHuman: false, cardCount: 5, cards: [], score: 0 },
        { id: 2, isHuman: false, cardCount: 5, cards: [], score: 0 },
        { id: 3, isHuman: false, cardCount: 5, cards: [], score: 0 },
      ],
      tableCards: [card('SPADE', 2), card('CLOVER', 3), card('DIAMOND', 4), card('SPADE', 5), card('CLOVER', 6)],
    });
    const hint = getFiftyOneHint(state);
    expect(hint?.targetAction).toBe('stop');
    expect(hint?.reason).toBe('hint.stopHigh');
  });

  it('suggests exchange with moderate confidence when improvement is small', () => {
    // Hand: H10, S9, C8, D7, H3 => best suit = HEART (10+3=13)
    // Table: H4, S2, C2, D2, S5
    // H4 replaces D7 => H10+H3+H4=17 (improvement of 4)
    const state = makeState({
      players: [
        {
          id: 0,
          isHuman: true,
          cardCount: 5,
          cards: [card('HEART', 10), card('SPADE', 9), card('CLOVER', 8), card('DIAMOND', 7), card('HEART', 3)],
          score: 13,
        },
        { id: 1, isHuman: false, cardCount: 5, cards: [], score: 0 },
        { id: 2, isHuman: false, cardCount: 5, cards: [], score: 0 },
        { id: 3, isHuman: false, cardCount: 5, cards: [], score: 0 },
      ],
      tableCards: [card('HEART', 4), card('SPADE', 2), card('CLOVER', 2), card('DIAMOND', 2), card('SPADE', 5)],
    });
    const hint = getFiftyOneHint(state);
    expect(hint).not.toBeNull();
    expect(hint?.targetAction).toBe('exchange');
    expect(hint?.confidence).toBe('moderate');
  });
});
