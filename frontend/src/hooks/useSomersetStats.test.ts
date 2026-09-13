import { act, renderHook } from '@testing-library/react';
import { afterEach, describe, expect, it } from 'vitest';
import {
  applySomersetResult,
  emptySomersetStats,
  readSomersetStats,
  SOMERSET_STATS_KEY,
  somersetWinRate,
  useSomersetStats,
} from './useSomersetStats';

describe('somersetWinRate', () => {
  it('returns the win rate and avoids division by zero', () => {
    expect(somersetWinRate({ plays: 10, wins: 3, fewestMoves: null })).toBe(30);
    expect(somersetWinRate({ plays: 3, wins: 1, fewestMoves: null })).toBe(33);
    expect(somersetWinRate({ plays: 0, wins: 0, fewestMoves: null })).toBe(0);
  });
});

describe('applySomersetResult', () => {
  it('records a clear and a new fewest-moves record', () => {
    const { stats, newBest } = applySomersetResult(emptySomersetStats(), { won: true, moves: 42 });
    expect(stats).toEqual({ plays: 1, wins: 1, fewestMoves: 42 });
    expect(newBest).toBe(true);
  });

  it('updates the record only when the clear used fewer moves', () => {
    const { stats, newBest } = applySomersetResult({ plays: 1, wins: 1, fewestMoves: 30 }, { won: true, moves: 25 });
    expect(stats).toEqual({ plays: 2, wins: 2, fewestMoves: 25 });
    expect(newBest).toBe(true);
  });

  it('does not worsen an existing record for a slower clear', () => {
    const { stats, newBest } = applySomersetResult({ plays: 1, wins: 1, fewestMoves: 30 }, { won: true, moves: 55 });
    expect(stats).toEqual({ plays: 2, wins: 2, fewestMoves: 30 });
    expect(newBest).toBe(false);
  });

  it('counts a loss without touching the fewest-moves record', () => {
    const { stats, newBest } = applySomersetResult({ plays: 1, wins: 1, fewestMoves: 30 }, { won: false, moves: 12 });
    expect(stats).toEqual({ plays: 2, wins: 1, fewestMoves: 30 });
    expect(newBest).toBe(false);
  });

  it('ignores a zero-move clear for the record', () => {
    const { stats, newBest } = applySomersetResult(emptySomersetStats(), { won: true, moves: 0 });
    expect(stats).toEqual({ plays: 1, wins: 1, fewestMoves: null });
    expect(newBest).toBe(false);
  });
});

describe('readSomersetStats', () => {
  afterEach(() => localStorage.clear());

  it('returns an empty record when nothing is stored', () => {
    expect(readSomersetStats()).toEqual(emptySomersetStats());
  });

  it('returns an empty record for malformed JSON', () => {
    localStorage.setItem(SOMERSET_STATS_KEY, '{not json');
    expect(readSomersetStats()).toEqual(emptySomersetStats());
  });

  it('returns an empty record for a structurally invalid value', () => {
    localStorage.setItem(SOMERSET_STATS_KEY, JSON.stringify({ plays: 'x' }));
    expect(readSomersetStats()).toEqual(emptySomersetStats());
  });
});

describe('useSomersetStats', () => {
  afterEach(() => localStorage.clear());

  it('persists a recorded result and reads it back on remount', () => {
    const { result } = renderHook(() => useSomersetStats());
    let newBest = false;
    act(() => {
      newBest = result.current.recordResult({ won: true, moves: 40 });
    });
    expect(newBest).toBe(true);
    expect(result.current.stats).toEqual({ plays: 1, wins: 1, fewestMoves: 40 });
    expect(JSON.parse(localStorage.getItem(SOMERSET_STATS_KEY) ?? '{}')).toEqual({
      plays: 1,
      wins: 1,
      fewestMoves: 40,
    });

    const { result: reread } = renderHook(() => useSomersetStats());
    expect(reread.current.stats).toEqual({ plays: 1, wins: 1, fewestMoves: 40 });
  });
});
