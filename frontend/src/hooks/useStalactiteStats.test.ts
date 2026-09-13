import { act, renderHook } from '@testing-library/react';
import { afterEach, describe, expect, it } from 'vitest';
import {
  applyStalactiteResult,
  emptyStalactiteStats,
  readStalactiteStats,
  STALACTITE_STATS_KEY,
  stalactiteWinRate,
  useStalactiteStats,
} from './useStalactiteStats';

afterEach(() => localStorage.clear());

describe('Stalactite stats', () => {
  it('records wins and losses and calculates win rate', () => {
    const first = applyStalactiteResult(emptyStalactiteStats(), { won: true, moves: 42 });
    const second = applyStalactiteResult(first.stats, { won: false, moves: 20 });
    expect(second.stats).toEqual({ plays: 2, wins: 1, fewestMoves: 42 });
    expect(stalactiteWinRate(second.stats)).toBe(50);
  });

  it('falls back to empty stats for empty or invalid storage', () => {
    expect(readStalactiteStats()).toEqual(emptyStalactiteStats());
    localStorage.setItem(STALACTITE_STATS_KEY, 'not-json');
    expect(readStalactiteStats()).toEqual(emptyStalactiteStats());
    localStorage.setItem(STALACTITE_STATS_KEY, JSON.stringify({ plays: 'bad' }));
    expect(readStalactiteStats()).toEqual(emptyStalactiteStats());
  });

  it('flags only a new fewest-moves record', () => {
    expect(applyStalactiteResult({ plays: 1, wins: 1, fewestMoves: 30 }, { won: true, moves: 25 }).newBest).toBe(true);
    expect(applyStalactiteResult({ plays: 1, wins: 1, fewestMoves: 30 }, { won: true, moves: 35 }).newBest).toBe(false);
    expect(applyStalactiteResult({ plays: 1, wins: 1, fewestMoves: 30 }, { won: false, moves: 10 }).newBest).toBe(
      false,
    );
  });

  it('persists a recorded result', () => {
    const { result } = renderHook(() => useStalactiteStats());
    act(() => result.current.recordResult({ won: true, moves: 18 }));
    expect(result.current.stats).toEqual({ plays: 1, wins: 1, fewestMoves: 18 });
    expect(readStalactiteStats()).toEqual(result.current.stats);
  });
});
