import { act, renderHook } from '@testing-library/react';
import { afterEach, describe, expect, it } from 'vitest';
import {
  applyCurdsAndWheyResult,
  CURDS_AND_WHEY_STATS_KEY,
  curdsAndWheyWinRate,
  emptyCurdsAndWheyStats,
  readCurdsAndWheyStats,
  useCurdsAndWheyStats,
} from './useCurdsAndWheyStats';

describe('Curds and Whey stats', () => {
  afterEach(() => localStorage.clear());

  it('counts wins and losses while keeping the fewest winning move count', () => {
    const won = applyCurdsAndWheyResult(emptyCurdsAndWheyStats(), { won: true, moves: 42 });
    expect(won).toEqual({ plays: 1, wins: 1, fewestMoves: 42 });
    expect(applyCurdsAndWheyResult(won, { won: false, moves: 12 })).toEqual({
      plays: 2,
      wins: 1,
      fewestMoves: 42,
    });
    expect(curdsAndWheyWinRate({ plays: 2, wins: 1, fewestMoves: 42 })).toBe(50);
  });

  it('does not update fewest moves for a zero-move win', () => {
    expect(applyCurdsAndWheyResult(emptyCurdsAndWheyStats(), { won: true, moves: 0 })).toEqual({
      plays: 1,
      wins: 1,
      fewestMoves: null,
    });
  });

  it('persists recorded results and returns an empty record for invalid storage', () => {
    const { result } = renderHook(() => useCurdsAndWheyStats());
    act(() => result.current.recordResult({ won: true, moves: 35 }));
    expect(readCurdsAndWheyStats()).toEqual({ plays: 1, wins: 1, fewestMoves: 35 });
    localStorage.setItem(CURDS_AND_WHEY_STATS_KEY, '{invalid');
    expect(readCurdsAndWheyStats()).toEqual(emptyCurdsAndWheyStats());
  });
});
