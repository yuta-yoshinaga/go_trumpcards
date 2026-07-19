import { act, renderHook } from '@testing-library/react';
import { afterEach, describe, expect, it } from 'vitest';
import {
  applyPyramidResult,
  emptyPyramidStats,
  PYRAMID_STATS_KEY,
  readPyramidStats,
  usePyramidStats,
} from './usePyramidStats';

describe('applyPyramidResult', () => {
  it('records a clear and a new fewest-moves record', () => {
    const { stats, newBest } = applyPyramidResult(emptyPyramidStats(), { won: true, moves: 42 });
    expect(stats).toEqual({ plays: 1, wins: 1, fewestMoves: 42 });
    expect(newBest).toBe(true);
  });

  it('updates the record only when the clear used fewer moves', () => {
    const { stats, newBest } = applyPyramidResult({ plays: 1, wins: 1, fewestMoves: 30 }, { won: true, moves: 25 });
    expect(stats).toEqual({ plays: 2, wins: 2, fewestMoves: 25 });
    expect(newBest).toBe(true);
  });

  it('does not worsen an existing record for a slower clear', () => {
    const { stats, newBest } = applyPyramidResult({ plays: 1, wins: 1, fewestMoves: 30 }, { won: true, moves: 55 });
    expect(stats).toEqual({ plays: 2, wins: 2, fewestMoves: 30 });
    expect(newBest).toBe(false);
  });

  it('counts a loss without touching the fewest-moves record', () => {
    const { stats, newBest } = applyPyramidResult({ plays: 1, wins: 1, fewestMoves: 30 }, { won: false, moves: 12 });
    expect(stats).toEqual({ plays: 2, wins: 1, fewestMoves: 30 });
    expect(newBest).toBe(false);
  });

  it('ignores a zero-move clear for the record', () => {
    const { stats, newBest } = applyPyramidResult(emptyPyramidStats(), { won: true, moves: 0 });
    expect(stats).toEqual({ plays: 1, wins: 1, fewestMoves: null });
    expect(newBest).toBe(false);
  });
});

describe('readPyramidStats', () => {
  afterEach(() => localStorage.clear());

  it('returns an empty record when nothing is stored', () => {
    expect(readPyramidStats()).toEqual(emptyPyramidStats());
  });

  it('returns an empty record for malformed JSON', () => {
    localStorage.setItem(PYRAMID_STATS_KEY, '{not json');
    expect(readPyramidStats()).toEqual(emptyPyramidStats());
  });

  it('returns an empty record for a structurally invalid value', () => {
    localStorage.setItem(PYRAMID_STATS_KEY, JSON.stringify({ plays: 'x' }));
    expect(readPyramidStats()).toEqual(emptyPyramidStats());
  });
});

describe('usePyramidStats', () => {
  afterEach(() => localStorage.clear());

  it('persists a recorded result and reads it back on remount', () => {
    const { result } = renderHook(() => usePyramidStats());
    let newBest = false;
    act(() => {
      newBest = result.current.recordResult({ won: true, moves: 40 });
    });
    expect(newBest).toBe(true);
    expect(result.current.stats).toEqual({ plays: 1, wins: 1, fewestMoves: 40 });
    expect(JSON.parse(localStorage.getItem(PYRAMID_STATS_KEY) ?? '{}')).toEqual({
      plays: 1,
      wins: 1,
      fewestMoves: 40,
    });

    const { result: reread } = renderHook(() => usePyramidStats());
    expect(reread.current.stats).toEqual({ plays: 1, wins: 1, fewestMoves: 40 });
  });
});
