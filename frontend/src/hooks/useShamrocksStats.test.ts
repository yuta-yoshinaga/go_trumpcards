import { act, renderHook } from '@testing-library/react';
import { afterEach, describe, expect, it } from 'vitest';
import {
  applyShamrocksClear,
  emptyShamrocksStats,
  readShamrocksStats,
  SHAMROCKS_STATS_KEY,
  useShamrocksStats,
} from './useShamrocksStats';

afterEach(() => localStorage.clear());

describe('applyShamrocksClear', () => {
  it('adds a clear and its moves', () => {
    expect(applyShamrocksClear(emptyShamrocksStats(), { moves: 24 }).stats).toEqual({
      games: 1,
      moves: 24,
      clearRecorded: true,
    });
  });

  it('does not count a duplicate clear', () => {
    const stats = { games: 1, moves: 24, clearRecorded: true };
    expect(applyShamrocksClear(stats, { moves: 24 }).stats).toBe(stats);
  });

  it('resets totals while retaining the active game guard', () => {
    expect(applyShamrocksClear({ games: 2, moves: 60, clearRecorded: true }, { reset: true }).stats).toEqual({
      games: 0,
      moves: 0,
      clearRecorded: true,
    });
  });
});

describe('readShamrocksStats', () => {
  it('returns empty stats for missing, malformed, or invalid data', () => {
    expect(readShamrocksStats()).toEqual(emptyShamrocksStats());
    localStorage.setItem(SHAMROCKS_STATS_KEY, '{bad');
    expect(readShamrocksStats()).toEqual(emptyShamrocksStats());
    localStorage.setItem(SHAMROCKS_STATS_KEY, JSON.stringify({ games: 1 }));
    expect(readShamrocksStats()).toEqual(emptyShamrocksStats());
  });
});

describe('useShamrocksStats', () => {
  it('persists clear results, suppresses duplicates, and resets totals', () => {
    const { result } = renderHook(() => useShamrocksStats());
    act(() => result.current.recordClear(18));
    expect(result.current.stats).toEqual({ games: 1, moves: 18, clearRecorded: true });

    const remounted = renderHook(() => useShamrocksStats());
    act(() => remounted.result.current.recordClear(18));
    expect(readShamrocksStats()).toEqual({ games: 1, moves: 18, clearRecorded: true });

    act(() => remounted.result.current.resetStats());
    expect(remounted.result.current.stats).toEqual({ games: 0, moves: 0, clearRecorded: true });
    expect(readShamrocksStats()).toEqual({ games: 0, moves: 0, clearRecorded: true });
  });
});
