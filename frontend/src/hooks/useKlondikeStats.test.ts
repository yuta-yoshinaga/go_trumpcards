import { act, renderHook } from '@testing-library/react';
import { beforeEach, describe, expect, it } from 'vitest';
import {
  applyKlondikeResult,
  emptyKlondikeStat,
  KLONDIKE_STATS_KEY,
  klondikeVariantKey,
  klondikeWinRate,
  readKlondikeStats,
  useKlondikeStats,
} from './useKlondikeStats';

beforeEach(() => {
  localStorage.clear();
});

describe('klondikeVariantKey', () => {
  it('joins draw count and scoring mode', () => {
    expect(klondikeVariantKey(1, 0)).toBe('1:0');
    expect(klondikeVariantKey(3, 1)).toBe('3:1');
  });
});

describe('applyKlondikeResult (pure reducer)', () => {
  it('a win increments plays and wins and sets initial best time/fewest moves', () => {
    const { stats, update } = applyKlondikeResult(
      {},
      { drawCount: 1, scoringMode: 0, won: true, timeSeconds: 120, moves: 90 },
    );
    expect(stats['1:0']).toEqual({ plays: 1, wins: 1, bestTimeSeconds: 120, fewestMoves: 90 });
    expect(update).toEqual({ newBestTime: true, newFewestMoves: true });
  });

  it('a loss increments plays only and leaves bests null', () => {
    const { stats, update } = applyKlondikeResult(
      {},
      { drawCount: 3, scoringMode: 1, won: false, timeSeconds: 200, moves: 40 },
    );
    expect(stats['3:1']).toEqual({ plays: 1, wins: 0, bestTimeSeconds: null, fewestMoves: null });
    expect(update).toEqual({ newBestTime: false, newFewestMoves: false });
  });

  it('a faster win with fewer moves updates both bests', () => {
    const first = applyKlondikeResult(
      {},
      { drawCount: 1, scoringMode: 0, won: true, timeSeconds: 300, moves: 120 },
    ).stats;
    const { stats, update } = applyKlondikeResult(first, {
      drawCount: 1,
      scoringMode: 0,
      won: true,
      timeSeconds: 180,
      moves: 95,
    });
    expect(stats['1:0']).toEqual({ plays: 2, wins: 2, bestTimeSeconds: 180, fewestMoves: 95 });
    expect(update).toEqual({ newBestTime: true, newFewestMoves: true });
  });

  it('a slower win with more moves keeps existing bests', () => {
    const first = applyKlondikeResult(
      {},
      { drawCount: 1, scoringMode: 0, won: true, timeSeconds: 180, moves: 95 },
    ).stats;
    const { stats, update } = applyKlondikeResult(first, {
      drawCount: 1,
      scoringMode: 0,
      won: true,
      timeSeconds: 300,
      moves: 120,
    });
    expect(stats['1:0']).toEqual({ plays: 2, wins: 2, bestTimeSeconds: 180, fewestMoves: 95 });
    expect(update).toEqual({ newBestTime: false, newFewestMoves: false });
  });

  it('ignores a non-positive clear time for the best-time record', () => {
    const { stats, update } = applyKlondikeResult(
      {},
      { drawCount: 1, scoringMode: 0, won: true, timeSeconds: 0, moves: 80 },
    );
    expect(stats['1:0'].bestTimeSeconds).toBeNull();
    expect(stats['1:0'].fewestMoves).toBe(80);
    expect(update).toEqual({ newBestTime: false, newFewestMoves: true });
  });

  it('keeps variants separate by draw count and scoring mode', () => {
    const s1 = applyKlondikeResult({}, { drawCount: 1, scoringMode: 0, won: true, timeSeconds: 100, moves: 80 }).stats;
    const s2 = applyKlondikeResult(s1, {
      drawCount: 3,
      scoringMode: 1,
      won: false,
      timeSeconds: 50,
      moves: 30,
    }).stats;
    expect(s2['1:0']).toEqual({ plays: 1, wins: 1, bestTimeSeconds: 100, fewestMoves: 80 });
    expect(s2['3:1']).toEqual({ plays: 1, wins: 0, bestTimeSeconds: null, fewestMoves: null });
  });
});

describe('klondikeWinRate', () => {
  it('is 0 with no plays', () => {
    expect(klondikeWinRate(emptyKlondikeStat())).toBe(0);
  });

  it('rounds to a whole percentage', () => {
    expect(klondikeWinRate({ plays: 3, wins: 1, bestTimeSeconds: null, fewestMoves: null })).toBe(33);
  });
});

describe('readKlondikeStats', () => {
  it('returns {} when nothing stored', () => {
    expect(readKlondikeStats()).toEqual({});
  });

  it('ignores malformed json', () => {
    localStorage.setItem(KLONDIKE_STATS_KEY, 'not-json');
    expect(readKlondikeStats()).toEqual({});
  });

  it('drops invalid entries but keeps valid ones', () => {
    localStorage.setItem(
      KLONDIKE_STATS_KEY,
      JSON.stringify({
        '1:0': { plays: 2, wins: 1, bestTimeSeconds: 120, fewestMoves: 90 },
        '3:1': { bogus: true },
      }),
    );
    const stats = readKlondikeStats();
    expect(stats['1:0']).toEqual({ plays: 2, wins: 1, bestTimeSeconds: 120, fewestMoves: 90 });
    expect(stats['3:1']).toBeUndefined();
  });
});

describe('useKlondikeStats', () => {
  it('records a win, persists it, and reads it back', () => {
    const { result } = renderHook(() => useKlondikeStats());
    act(() => {
      result.current.recordResult({ drawCount: 1, scoringMode: 0, won: true, timeSeconds: 150, moves: 100 });
    });
    expect(result.current.getStat(1, 0)).toEqual({ plays: 1, wins: 1, bestTimeSeconds: 150, fewestMoves: 100 });
    expect(readKlondikeStats()['1:0']).toEqual({ plays: 1, wins: 1, bestTimeSeconds: 150, fewestMoves: 100 });
  });

  it('records a loss incrementing plays only', () => {
    const { result } = renderHook(() => useKlondikeStats());
    act(() => {
      result.current.recordResult({ drawCount: 3, scoringMode: 1, won: false, timeSeconds: 80, moves: 40 });
    });
    const stat = result.current.getStat(3, 1);
    expect(stat.plays).toBe(1);
    expect(stat.wins).toBe(0);
  });

  it('returns a best-update flag when a personal best is beaten', () => {
    const { result } = renderHook(() => useKlondikeStats());
    let update = { newBestTime: false, newFewestMoves: false };
    act(() => {
      update = result.current.recordResult({ drawCount: 1, scoringMode: 0, won: true, timeSeconds: 90, moves: 70 });
    });
    expect(update).toEqual({ newBestTime: true, newFewestMoves: true });
  });

  it('getStat returns an empty stat for an untouched variant', () => {
    const { result } = renderHook(() => useKlondikeStats());
    expect(result.current.getStat(3, 1)).toEqual(emptyKlondikeStat());
  });
});
