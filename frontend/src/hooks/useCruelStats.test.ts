import { act, renderHook } from '@testing-library/react';
import { afterEach, describe, expect, it } from 'vitest';
import {
  applyCruelResult,
  CRUEL_STATS_KEY,
  cruelWinRate,
  emptyCruelStats,
  readCruelStats,
  useCruelStats,
} from './useCruelStats';

afterEach(() => localStorage.clear());

describe('Cruel stats', () => {
  it('records wins and losses and calculates win rate', () => {
    const first = applyCruelResult(emptyCruelStats(), { won: true, moves: 42 });
    const second = applyCruelResult(first.stats, { won: false, moves: 20 });
    expect(second.stats).toEqual({ plays: 2, wins: 1, fewestMoves: 42 });
    expect(cruelWinRate(second.stats)).toBe(50);
  });

  it('falls back to empty stats for empty or invalid storage', () => {
    expect(readCruelStats()).toEqual(emptyCruelStats());
    localStorage.setItem(CRUEL_STATS_KEY, 'not-json');
    expect(readCruelStats()).toEqual(emptyCruelStats());
    localStorage.setItem(CRUEL_STATS_KEY, JSON.stringify({ plays: 'bad' }));
    expect(readCruelStats()).toEqual(emptyCruelStats());
  });

  it('flags only a new fewest-moves record', () => {
    expect(applyCruelResult({ plays: 1, wins: 1, fewestMoves: 30 }, { won: true, moves: 25 }).newBest).toBe(true);
    expect(applyCruelResult({ plays: 1, wins: 1, fewestMoves: 30 }, { won: true, moves: 35 }).newBest).toBe(false);
    expect(applyCruelResult({ plays: 1, wins: 1, fewestMoves: 30 }, { won: false, moves: 10 }).newBest).toBe(false);
  });

  it('persists a recorded result', () => {
    const { result } = renderHook(() => useCruelStats());
    act(() => result.current.recordResult({ won: true, moves: 18 }));
    expect(result.current.stats).toEqual({ plays: 1, wins: 1, fewestMoves: 18 });
    expect(readCruelStats()).toEqual(result.current.stats);
  });
});
