import { act, renderHook } from '@testing-library/react';
import { beforeEach, describe, expect, it } from 'vitest';
import {
  applySpiderResult,
  emptySpiderStat,
  readSpiderStats,
  SPIDER_STATS_KEY,
  spiderWinRate,
  useSpiderStats,
} from './useSpiderStats';

beforeEach(() => {
  localStorage.clear();
});

describe('applySpiderResult (pure reducer)', () => {
  it('a win increments plays and wins and sets initial best score/fewest moves', () => {
    const { stats, update } = applySpiderResult({}, { difficulty: 1, won: true, score: 500, moves: 42 });
    expect(stats['1']).toEqual({ plays: 1, wins: 1, bestScore: 500, fewestMoves: 42 });
    expect(update).toEqual({ newBestScore: true, newFewestMoves: true });
  });

  it('a loss increments plays only and leaves bests null', () => {
    const { stats, update } = applySpiderResult({}, { difficulty: 2, won: false, score: 120, moves: 90 });
    expect(stats['2']).toEqual({ plays: 1, wins: 0, bestScore: null, fewestMoves: null });
    expect(update).toEqual({ newBestScore: false, newFewestMoves: false });
  });

  it('a better win updates best score and fewest moves', () => {
    const first = applySpiderResult({}, { difficulty: 1, won: true, score: 400, moves: 60 }).stats;
    const { stats, update } = applySpiderResult(first, { difficulty: 1, won: true, score: 550, moves: 45 });
    expect(stats['1']).toEqual({ plays: 2, wins: 2, bestScore: 550, fewestMoves: 45 });
    expect(update).toEqual({ newBestScore: true, newFewestMoves: true });
  });

  it('a worse win keeps existing bests', () => {
    const first = applySpiderResult({}, { difficulty: 1, won: true, score: 550, moves: 45 }).stats;
    const { stats, update } = applySpiderResult(first, { difficulty: 1, won: true, score: 300, moves: 80 });
    expect(stats['1']).toEqual({ plays: 2, wins: 2, bestScore: 550, fewestMoves: 45 });
    expect(update).toEqual({ newBestScore: false, newFewestMoves: false });
  });

  it('keeps difficulties separate', () => {
    const s1 = applySpiderResult({}, { difficulty: 1, won: true, score: 500, moves: 40 }).stats;
    const s2 = applySpiderResult(s1, { difficulty: 4, won: false, score: 100, moves: 30 }).stats;
    expect(s2['1']).toEqual({ plays: 1, wins: 1, bestScore: 500, fewestMoves: 40 });
    expect(s2['4']).toEqual({ plays: 1, wins: 0, bestScore: null, fewestMoves: null });
  });
});

describe('spiderWinRate', () => {
  it('is 0 with no plays', () => {
    expect(spiderWinRate(emptySpiderStat())).toBe(0);
  });

  it('rounds to a whole percentage', () => {
    expect(spiderWinRate({ plays: 3, wins: 1, bestScore: null, fewestMoves: null })).toBe(33);
  });
});

describe('readSpiderStats', () => {
  it('returns {} when nothing stored', () => {
    expect(readSpiderStats()).toEqual({});
  });

  it('ignores malformed json', () => {
    localStorage.setItem(SPIDER_STATS_KEY, 'not-json');
    expect(readSpiderStats()).toEqual({});
  });

  it('drops invalid entries but keeps valid ones', () => {
    localStorage.setItem(
      SPIDER_STATS_KEY,
      JSON.stringify({ '1': { plays: 2, wins: 1, bestScore: 500, fewestMoves: 40 }, '2': { bogus: true } }),
    );
    const stats = readSpiderStats();
    expect(stats['1']).toEqual({ plays: 2, wins: 1, bestScore: 500, fewestMoves: 40 });
    expect(stats['2']).toBeUndefined();
  });
});

describe('useSpiderStats', () => {
  it('records a win, persists it, and reads it back', () => {
    const { result } = renderHook(() => useSpiderStats());
    act(() => {
      result.current.recordResult({ difficulty: 1, won: true, score: 480, moves: 55 });
    });
    expect(result.current.getStat(1)).toEqual({ plays: 1, wins: 1, bestScore: 480, fewestMoves: 55 });
    // Persisted to localStorage
    expect(readSpiderStats()['1']).toEqual({ plays: 1, wins: 1, bestScore: 480, fewestMoves: 55 });
  });

  it('records a loss incrementing losses (plays) only', () => {
    const { result } = renderHook(() => useSpiderStats());
    act(() => {
      result.current.recordResult({ difficulty: 2, won: false, score: 90, moves: 70 });
    });
    const stat = result.current.getStat(2);
    expect(stat.plays).toBe(1);
    expect(stat.wins).toBe(0);
  });

  it('returns a best-update flag when a personal best is beaten', () => {
    const { result } = renderHook(() => useSpiderStats());
    let update = { newBestScore: false, newFewestMoves: false };
    act(() => {
      update = result.current.recordResult({ difficulty: 4, won: true, score: 300, moves: 100 });
    });
    expect(update).toEqual({ newBestScore: true, newFewestMoves: true });
  });

  it('getStat returns an empty stat for an untouched difficulty', () => {
    const { result } = renderHook(() => useSpiderStats());
    expect(result.current.getStat(4)).toEqual(emptySpiderStat());
  });
});
