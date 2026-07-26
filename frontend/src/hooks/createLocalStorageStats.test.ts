import { act, renderHook } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it } from 'vitest';
import { createFlatStatsReader, createLocalStorageStats } from './createLocalStorageStats';

interface Stat {
  plays: number;
  best: number | null;
}
const KEY = 'test-stats';
const empty = (): Stat => ({ plays: 0, best: null });
const isStat = (v: unknown): v is Stat => typeof v === 'object' && v !== null && typeof (v as Stat).plays === 'number';

describe('createFlatStatsReader', () => {
  beforeEach(() => localStorage.clear());
  afterEach(() => localStorage.clear());

  const read = createFlatStatsReader(KEY, empty, isStat);

  it('returns empty when the key is absent', () => {
    expect(read()).toEqual({ plays: 0, best: null });
  });

  it('returns the parsed value when valid', () => {
    localStorage.setItem(KEY, JSON.stringify({ plays: 3, best: 90 }));
    expect(read()).toEqual({ plays: 3, best: 90 });
  });

  it('falls back to empty on malformed JSON', () => {
    localStorage.setItem(KEY, '{not json');
    expect(read()).toEqual({ plays: 0, best: null });
  });

  it('falls back to empty when validation fails', () => {
    localStorage.setItem(KEY, JSON.stringify({ nope: true }));
    expect(read()).toEqual({ plays: 0, best: null });
  });
});

describe('createLocalStorageStats', () => {
  beforeEach(() => localStorage.clear());
  afterEach(() => localStorage.clear());

  const read = createFlatStatsReader(KEY, empty, isStat);
  const store = createLocalStorageStats<Stat, { score: number }, boolean>({
    key: KEY,
    read,
    reduce: (prev, result) => {
      const newBest = prev.best === null || result.score > prev.best;
      return { stats: { plays: prev.plays + 1, best: newBest ? result.score : prev.best }, update: newBest };
    },
  });

  it('apply is the provided reduce', () => {
    expect(store.apply({ plays: 0, best: null }, { score: 5 })).toEqual({
      stats: { plays: 1, best: 5 },
      update: true,
    });
  });

  it('useStats records, persists, and returns the update', () => {
    const { result } = renderHook(() => store.useStats());
    let update = false;
    act(() => {
      update = result.current.recordResult({ score: 42 });
    });
    expect(update).toBe(true);
    expect(result.current.stats).toEqual({ plays: 1, best: 42 });
    expect(JSON.parse(localStorage.getItem(KEY) ?? '{}')).toEqual({ plays: 1, best: 42 });

    // A lower score does not beat the best.
    act(() => {
      update = result.current.recordResult({ score: 10 });
    });
    expect(update).toBe(false);
    expect(result.current.stats).toEqual({ plays: 2, best: 42 });
  });
});
