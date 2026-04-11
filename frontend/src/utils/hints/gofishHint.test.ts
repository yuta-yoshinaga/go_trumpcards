import { describe, expect, it } from 'vitest';
import type { Card, GoFishConfig, GoFishResponse } from '../../types/card';
import { GoFishPhase } from '../../types/phases';
import { getGoFishHint } from './gofishHint';

const card = (design: Card['design'], value: number): Card => ({ design, value });

const defaultConfig: GoFishConfig = { cpuDifficulty: 0 };

function makeState(overrides: Partial<GoFishResponse> = {}): GoFishResponse {
  return {
    players: [
      {
        id: 0,
        isHuman: true,
        cardCount: 4,
        cards: [card('HEART', 5), card('SPADE', 5), card('DIAMOND', 8), card('CLOVER', 3)],
        bookCount: 0,
        books: [],
      },
      { id: 1, isHuman: false, cardCount: 5, cards: [], bookCount: 0, books: [] },
      { id: 2, isHuman: false, cardCount: 5, cards: [], bookCount: 0, books: [] },
    ],
    phase: GoFishPhase.PLAY,
    currentTurn: 0,
    gameEndFlag: false,
    winnerIdx: -1,
    turnNumber: 0,
    deckRemaining: 20,
    lastAsk: null,
    cpuActions: [],
    humanAction: null,
    message: '',
    config: defaultConfig,
    ...overrides,
  };
}

describe('getGoFishHint', () => {
  it('returns null when game has ended', () => {
    expect(getGoFishHint(makeState({ gameEndFlag: true }))).toBeNull();
  });

  it('returns null when phase is not PLAY', () => {
    expect(getGoFishHint(makeState({ phase: GoFishPhase.GAME_END, gameEndFlag: false }))).toBeNull();
  });

  it('returns null when it is not the human turn', () => {
    expect(getGoFishHint(makeState({ currentTurn: 1 }))).toBeNull();
  });

  it('returns null when human has no cards', () => {
    const state = makeState();
    state.players[0].cards = [];
    expect(getGoFishHint(state)).toBeNull();
  });

  it('recommends asking with moderate confidence when no opponent info is known', () => {
    const hint = getGoFishHint(makeState());
    expect(hint).not.toBeNull();
    expect(hint?.targetAction).toBe('ask');
    expect(hint?.confidence).toBe('moderate');
    expect(hint?.reason).toBe('hint.askMostCopies');
  });

  it('upgrades confidence when a cpu has recently taken a rank the human holds', () => {
    const state = makeState({
      cpuActions: [
        {
          askPlayerIdx: 1,
          askTargetIdx: 2,
          askRank: 5,
          success: true,
          cardsReceived: 1,
          drawnCard: null,
          bookFormed: false,
          bookRank: 0,
        },
      ],
    });
    const hint = getGoFishHint(state);
    expect(hint?.confidence).toBe('strong');
    expect(hint?.reason).toBe('hint.askKnownRank');
  });

  it('uses lastAsk when available', () => {
    const state = makeState({
      lastAsk: {
        playerIdx: 1,
        targetIdx: 2,
        rank: 5,
        success: true,
        cardsReceived: [card('HEART', 5)],
        drawnCard: null,
        bookFormed: false,
        bookRank: 0,
      },
    });
    const hint = getGoFishHint(state);
    expect(hint?.confidence).toBe('strong');
  });

  it('ignores failed asks', () => {
    const state = makeState({
      cpuActions: [
        {
          askPlayerIdx: 1,
          askTargetIdx: 2,
          askRank: 5,
          success: false,
          cardsReceived: 0,
          drawnCard: null,
          bookFormed: false,
          bookRank: 0,
        },
      ],
    });
    const hint = getGoFishHint(state);
    expect(hint?.confidence).toBe('moderate');
  });

  it('ignores lastAsk where the human was the asker', () => {
    const state = makeState({
      lastAsk: {
        playerIdx: 0, // human was the asker — collectFromLastAsk returns early
        targetIdx: 1,
        rank: 5,
        success: true,
        cardsReceived: [card('HEART', 5)],
        drawnCard: null,
        bookFormed: false,
        bookRank: 0,
      },
    });
    const hint = getGoFishHint(state);
    expect(hint?.confidence).toBe('moderate');
  });

  it('ignores asks targeting the human itself', () => {
    const state = makeState({
      cpuActions: [
        {
          askPlayerIdx: 1,
          askTargetIdx: 0,
          askRank: 5,
          success: true,
          cardsReceived: 1,
          drawnCard: null,
          bookFormed: false,
          bookRank: 0,
        },
      ],
    });
    const hint = getGoFishHint(state);
    expect(hint?.confidence).toBe('moderate');
  });
});
