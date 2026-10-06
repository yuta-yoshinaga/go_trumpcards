import { act, renderHook } from '@testing-library/react';
import { beforeEach, describe, expect, it } from 'vitest';
import {
  applyPerseveranceStatsAction,
  emptyPerseveranceStats,
  PERSEVERANCE_STATS_KEY,
  perseveranceClearRate,
  readPerseveranceStats,
  usePerseveranceStats,
} from './usePerseveranceStats';

beforeEach(() => localStorage.clear());

describe('Perseverance stats reducer', () => {
  it('counts each playing game once and calculates a clear rate', () => {
    const started = applyPerseveranceStatsAction(emptyPerseveranceStats(), { type: 'playing' }).stats;
    expect(applyPerseveranceStatsAction(started, { type: 'playing' }).stats).toEqual(started);
    const cleared = applyPerseveranceStatsAction(started, {
      type: 'clear',
      moveCount: 23,
      redealsLeft: 2,
    }).stats;
    expect(cleared).toMatchObject({ starts: 1, games: 1 });
    expect(perseveranceClearRate(cleared)).toBe(100);
    expect(perseveranceClearRate(emptyPerseveranceStats())).toBe(0);
    const next = applyPerseveranceStatsAction(cleared, { type: 'playing' }).stats;
    expect(next.starts).toBe(2);
    expect(perseveranceClearRate(next)).toBe(50);
  });

  it('records a clear once and accumulates moves and remaining redeals', () => {
    const first = applyPerseveranceStatsAction(emptyPerseveranceStats(), {
      type: 'clear',
      moveCount: 23,
      redealsLeft: 2,
    }).stats;
    expect(first).toEqual({
      starts: 1,
      games: 1,
      moves: 23,
      redeals: 2,
      minMoves: 23,
      clearRecorded: true,
      gameStarted: false,
    });
    expect(applyPerseveranceStatsAction(first, { type: 'clear', moveCount: 23, redealsLeft: 2 }).stats).toEqual(first);
  });

  it('allows a new clear after play starts and retains the clear guard when stats are cleared', () => {
    const cleared = {
      starts: 1,
      games: 1,
      moves: 23,
      redeals: 2,
      minMoves: 23,
      clearRecorded: true,
      gameStarted: false,
    };
    const playing = applyPerseveranceStatsAction(cleared, { type: 'playing' }).stats;
    expect(playing.clearRecorded).toBe(false);
    expect(applyPerseveranceStatsAction(playing, { type: 'clear', moveCount: 4, redealsLeft: 1 }).stats).toEqual({
      games: 2,
      starts: 2,
      moves: 27,
      redeals: 3,
      minMoves: 4,
      clearRecorded: true,
      gameStarted: false,
    });
    expect(applyPerseveranceStatsAction(cleared, { type: 'reset' }).stats).toEqual({
      ...emptyPerseveranceStats(),
      clearRecorded: true,
    });
  });
});

describe('usePerseveranceStats', () => {
  it('rehydrates, records a clear, and persists aggregated statistics', () => {
    localStorage.setItem(
      PERSEVERANCE_STATS_KEY,
      JSON.stringify({ starts: 1, games: 1, moves: 10, redeals: 1, clearRecorded: false, gameStarted: true }),
    );
    const { result } = renderHook(() => usePerseveranceStats());
    expect(result.current.stats.games).toBe(1);
    act(() => result.current.recordClear(15, 2));
    expect(result.current.stats).toEqual({
      starts: 1,
      games: 2,
      moves: 25,
      redeals: 3,
      minMoves: 15,
      clearRecorded: true,
      gameStarted: false,
    });
    expect(readPerseveranceStats()).toEqual(result.current.stats);
  });

  it('migrates persisted stats without start tracking while preserving existing totals', () => {
    localStorage.setItem(
      PERSEVERANCE_STATS_KEY,
      JSON.stringify({ games: 1, moves: 10, redeals: 1, clearRecorded: false }),
    );
    expect(readPerseveranceStats()).toEqual({
      starts: 1,
      games: 1,
      moves: 10,
      redeals: 1,
      minMoves: null,
      clearRecorded: false,
      gameStarted: false,
    });
    const { result } = renderHook(() => usePerseveranceStats());
    expect(result.current.stats).toMatchObject({ games: 1, starts: 1, moves: 10, redeals: 1, gameStarted: false });
  });

  it('returns empty stats for malformed persisted values', () => {
    localStorage.setItem(PERSEVERANCE_STATS_KEY, '{broken');
    expect(readPerseveranceStats()).toEqual(emptyPerseveranceStats());
  });

  it('clears totals without allowing the same clear to be recorded again', () => {
    const { result } = renderHook(() => usePerseveranceStats());
    act(() => result.current.recordClear(8, 1));
    act(() => result.current.clearStats());
    act(() => result.current.recordClear(8, 1));
    expect(result.current.stats).toEqual({ ...emptyPerseveranceStats(), clearRecorded: true });
  });

  it('keeps the lowest move count across clears and resets it when results are cleared', () => {
    const { result } = renderHook(() => usePerseveranceStats());
    act(() => result.current.recordClear(20, 1));
    act(() => result.current.markPlaying());
    act(() => result.current.recordClear(24, 1));
    expect(result.current.stats.minMoves).toBe(20);
    act(() => result.current.markPlaying());
    act(() => result.current.recordClear(12, 1));
    expect(result.current.stats.minMoves).toBe(12);
    act(() => result.current.clearStats());
    expect(result.current.stats.minMoves).toBeNull();
  });

  it('persists one start despite repeated playing notifications and clears the start total', () => {
    const { result } = renderHook(() => usePerseveranceStats());
    act(() => {
      result.current.markPlaying();
      result.current.markPlaying();
    });
    expect(result.current.stats.starts).toBe(1);
    act(() => result.current.clearStats());
    expect(result.current.stats.starts).toBe(0);
  });
});
