import { act, renderHook } from '@testing-library/react';
import { afterEach, describe, expect, it } from 'vitest';
import {
  applyTriPeaksResult,
  emptyTriPeaksStats,
  readTriPeaksStats,
  TRIPEAKS_STATS_KEY,
  useTriPeaksStats,
} from './useTriPeaksStats';

describe('applyTriPeaksResult', () => {
  it('records a win and a new best score', () => {
    const { stats, newBest } = applyTriPeaksResult(emptyTriPeaksStats(), { won: true, score: 500 });
    expect(stats).toEqual({ plays: 1, wins: 1, bestScore: 500 });
    expect(newBest).toBe(true);
  });

  it('does not lower an existing best score', () => {
    const { stats, newBest } = applyTriPeaksResult({ plays: 1, wins: 1, bestScore: 900 }, { won: false, score: 400 });
    expect(stats).toEqual({ plays: 2, wins: 1, bestScore: 900 });
    expect(newBest).toBe(false);
  });

  it('counts a loss and still records a best score from it', () => {
    const { stats, newBest } = applyTriPeaksResult(emptyTriPeaksStats(), { won: false, score: 300 });
    expect(stats).toEqual({ plays: 1, wins: 0, bestScore: 300 });
    expect(newBest).toBe(true);
  });

  it('ignores a zero score for the best record', () => {
    const { stats, newBest } = applyTriPeaksResult(emptyTriPeaksStats(), { won: false, score: 0 });
    expect(stats).toEqual({ plays: 1, wins: 0, bestScore: null });
    expect(newBest).toBe(false);
  });
});

describe('readTriPeaksStats', () => {
  afterEach(() => localStorage.clear());

  it('returns an empty record when nothing is stored', () => {
    expect(readTriPeaksStats()).toEqual(emptyTriPeaksStats());
  });

  it('returns an empty record for malformed JSON', () => {
    localStorage.setItem(TRIPEAKS_STATS_KEY, '{not json');
    expect(readTriPeaksStats()).toEqual(emptyTriPeaksStats());
  });

  it('returns an empty record for a structurally invalid value', () => {
    localStorage.setItem(TRIPEAKS_STATS_KEY, JSON.stringify({ plays: 'x' }));
    expect(readTriPeaksStats()).toEqual(emptyTriPeaksStats());
  });
});

describe('useTriPeaksStats', () => {
  afterEach(() => localStorage.clear());

  it('persists a recorded result and reads it back on remount', () => {
    const { result } = renderHook(() => useTriPeaksStats());
    let newBest = false;
    act(() => {
      newBest = result.current.recordResult({ won: true, score: 700 });
    });
    expect(newBest).toBe(true);
    expect(result.current.stats).toEqual({ plays: 1, wins: 1, bestScore: 700 });
    expect(JSON.parse(localStorage.getItem(TRIPEAKS_STATS_KEY) ?? '{}')).toEqual({
      plays: 1,
      wins: 1,
      bestScore: 700,
    });

    const { result: reread } = renderHook(() => useTriPeaksStats());
    expect(reread.current.stats).toEqual({ plays: 1, wins: 1, bestScore: 700 });
  });
});
