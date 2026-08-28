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

  // The server keeps this itself; the CUI has always read it from there. The
  // client-side accumulation lives in component state, so a reload wiped it.
  it('uses the ranks the server sent instead of rebuilding them', () => {
    const state = s({
      players: [
        { id: 0, isHuman: true, cardCount: 5, cards: [], bookCount: 0, books: [], knownRanks: [3, 7] },
        { id: 1, isHuman: false, cardCount: 5, cards: [], bookCount: 0, books: [], knownRanks: [11] },
      ],
    });
    const { result } = renderHook(() => useGoFishKnownRanks(state));
    // No lastAsk history at all: rebuilding from scratch would give {}.
    expect(result.current).toEqual({ 0: [3, 7], 1: [11] });
  });

  it('survives a remount, which is what a page reload is', () => {
    const state = s({
      players: [
        { id: 0, isHuman: true, cardCount: 5, cards: [], bookCount: 0, books: [], knownRanks: [3] },
        { id: 1, isHuman: false, cardCount: 5, cards: [], bookCount: 0, books: [], knownRanks: [] },
      ],
    });
    const first = renderHook(() => useGoFishKnownRanks(state));
    expect(first.result.current[0]).toEqual([3]);
    first.unmount();

    // A fresh mount with the same server state must still know rank 3.
    const second = renderHook(() => useGoFishKnownRanks(state));
    expect(second.result.current[0]).toEqual([3]);
  });

  it('still rebuilds when the server did not send the field', () => {
    // Older responses have no knownRanks; the accumulation must remain the
    // fallback rather than silently reporting an empty table.
    const { result, rerender } = renderHook(({ st }) => useGoFishKnownRanks(st), {
      initialProps: {
        st: s({
          lastAsk: {
            playerIdx: 0,
            targetIdx: 1,
            rank: 5,
            success: true,
            cardsReceived: [],
            drawnCard: null,
            bookFormed: false,
            bookRank: 0,
          },
        }),
      },
    });
    rerender({
      st: s({
        lastAsk: {
          playerIdx: 0,
          targetIdx: 1,
          rank: 5,
          success: true,
          cardsReceived: [],
          drawnCard: null,
          bookFormed: false,
          bookRank: 0,
        },
      }),
    });
    expect(result.current[0]).toEqual([5]);
  });
});
