import { act, renderHook } from '@testing-library/react';
import { beforeEach, describe, expect, it } from 'vitest';
import {
  applyPerseveranceStatsAction,
  emptyPerseveranceStats,
  PERSEVERANCE_STATS_KEY,
  readPerseveranceStats,
  usePerseveranceStats,
} from './usePerseveranceStats';

beforeEach(() => localStorage.clear());

describe('Perseverance stats reducer', () => {
  it('records a clear once and accumulates moves and remaining redeals', () => {
    const first = applyPerseveranceStatsAction(emptyPerseveranceStats(), {
      type: 'clear',
      moveCount: 23,
      redealsLeft: 2,
    }).stats;
    expect(first).toEqual({ games: 1, moves: 23, redeals: 2, clearRecorded: true });
    expect(applyPerseveranceStatsAction(first, { type: 'clear', moveCount: 23, redealsLeft: 2 }).stats).toEqual(first);
  });

  it('allows a new clear after play starts and retains the clear guard when stats are cleared', () => {
    const cleared = { games: 1, moves: 23, redeals: 2, clearRecorded: true };
    const playing = applyPerseveranceStatsAction(cleared, { type: 'playing' }).stats;
    expect(playing.clearRecorded).toBe(false);
    expect(applyPerseveranceStatsAction(playing, { type: 'clear', moveCount: 4, redealsLeft: 1 }).stats).toEqual({
      games: 2,
      moves: 27,
      redeals: 3,
      clearRecorded: true,
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
      JSON.stringify({ games: 1, moves: 10, redeals: 1, clearRecorded: false }),
    );
    const { result } = renderHook(() => usePerseveranceStats());
    expect(result.current.stats.games).toBe(1);
    act(() => result.current.recordClear(15, 2));
    expect(result.current.stats).toEqual({ games: 2, moves: 25, redeals: 3, clearRecorded: true });
    expect(readPerseveranceStats()).toEqual(result.current.stats);
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
});
