import { renderHook } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import type { GoFishResponse } from '../types/card';
import { useGoFishKnownRanks } from './useGoFishKnownRanks';

function s(partial: Partial<GoFishResponse>): GoFishResponse {
  return {
    players: [
      { id: 0, isHuman: true, cardCount: 5, cards: [], bookCount: 0, books: [] },
      { id: 1, isHuman: false, cardCount: 5, cards: [], bookCount: 0, books: [] },
    ],
    phase: 0,
    currentTurn: 0,
    gameEndFlag: false,
    winnerIdx: -1,
    turnNumber: 1,
    deckRemaining: 30,
    lastAsk: null,
    cpuActions: [],
    humanAction: null,
    message: '',
    config: { cpuDifficulty: 0 },
    ...partial,
  };
}

describe('useGoFishKnownRanks', () => {
  it('returns empty when state is null', () => {
    const { result } = renderHook(() => useGoFishKnownRanks(null));
    expect(result.current).toEqual({});
  });

  it('accumulates ranks across consecutive asks by the same player', () => {
    const initial = s({});
    const { result, rerender } = renderHook(({ st }) => useGoFishKnownRanks(st), {
      initialProps: { st: initial },
    });
    expect(result.current[1]).toEqual([]);

    rerender({
      st: s({
        turnNumber: 2,
        lastAsk: {
          playerIdx: 1,
          targetIdx: 0,
          rank: 7,
          success: false,
          cardsReceived: [],
          drawnCard: null,
          bookFormed: false,
          bookRank: 0,
        },
      }),
    });
    expect(result.current[1]).toEqual([7]);

    rerender({
      st: s({
        turnNumber: 3,
        lastAsk: {
          playerIdx: 1,
          targetIdx: 0,
          rank: 12,
          success: true,
          cardsReceived: [],
          drawnCard: null,
          bookFormed: false,
          bookRank: 0,
        },
      }),
    });
    expect(result.current[1]).toEqual([7, 12]);
  });

  it('drops ranks the player has booked', () => {
    const { result, rerender } = renderHook(({ st }) => useGoFishKnownRanks(st), {
      initialProps: { st: s({}) },
    });
    rerender({
      st: s({
        turnNumber: 2,
        lastAsk: {
          playerIdx: 1,
          targetIdx: 0,
          rank: 7,
          success: false,
          cardsReceived: [],
          drawnCard: null,
          bookFormed: false,
          bookRank: 0,
        },
      }),
    });
    expect(result.current[1]).toEqual([7]);

    rerender({
      st: s({
        turnNumber: 3,
        lastAsk: null,
        players: [
          { id: 0, isHuman: true, cardCount: 5, cards: [], bookCount: 0, books: [] },
          { id: 1, isHuman: false, cardCount: 1, cards: [], bookCount: 1, books: [{ rank: 7, cards: [] }] },
        ],
      }),
    });
    expect(result.current[1]).toEqual([]);
  });

  it('resets when a new game starts (turnNumber back to 1)', () => {
    const { result, rerender } = renderHook(({ st }) => useGoFishKnownRanks(st), {
      initialProps: { st: s({ turnNumber: 2 }) },
    });
    rerender({
      st: s({
        turnNumber: 5,
        lastAsk: {
          playerIdx: 1,
          targetIdx: 0,
          rank: 4,
          success: false,
          cardsReceived: [],
          drawnCard: null,
          bookFormed: false,
          bookRank: 0,
        },
      }),
    });
    expect(result.current[1]).toEqual([4]);
    rerender({ st: s({ turnNumber: 1 }) });
    expect(result.current[1]).toEqual([]);
  });
});
