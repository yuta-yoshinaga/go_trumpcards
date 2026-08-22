import { act, renderHook } from '@testing-library/react';
import { beforeEach, describe, expect, it } from 'vitest';
import {
  CARIBBEANDRAW_HISTORY_KEY,
  CARIBBEANDRAW_HISTORY_MAX,
  CaribbeanDrawOutcome,
  type CaribbeanDrawRecord,
  outcomeFromResult,
  readCaribbeanDrawHistory,
  tallyCaribbeanDrawHistory,
  useCaribbeanDrawStats,
} from './useCaribbeanDrawStats';

beforeEach(() => {
  localStorage.clear();
});

describe('outcomeFromResult', () => {
  it('maps a positive result to WIN', () => {
    expect(outcomeFromResult(1)).toBe(CaribbeanDrawOutcome.WIN);
  });
  it('maps a negative result to LOSS', () => {
    expect(outcomeFromResult(-1)).toBe(CaribbeanDrawOutcome.LOSS);
  });
  it('maps a zero result to PUSH', () => {
    expect(outcomeFromResult(0)).toBe(CaribbeanDrawOutcome.PUSH);
  });
});

describe('tallyCaribbeanDrawHistory', () => {
  it('counts wins, losses, pushes, hands, and net chips', () => {
    const tally = tallyCaribbeanDrawHistory([
      { outcome: CaribbeanDrawOutcome.WIN, net: 300 },
      { outcome: CaribbeanDrawOutcome.LOSS, net: -100 },
      { outcome: CaribbeanDrawOutcome.WIN, net: 200 },
      { outcome: CaribbeanDrawOutcome.PUSH, net: 0 },
    ]);
    expect(tally).toEqual({ wins: 2, losses: 1, pushes: 1, hands: 4, net: 400 });
  });
  it('returns zeros for empty history', () => {
    expect(tallyCaribbeanDrawHistory([])).toEqual({ wins: 0, losses: 0, pushes: 0, hands: 0, net: 0 });
  });
});

describe('readCaribbeanDrawHistory', () => {
  it('returns [] when nothing is stored', () => {
    expect(readCaribbeanDrawHistory()).toEqual([]);
  });
  it('returns [] on corrupt (non-JSON) data', () => {
    localStorage.setItem(CARIBBEANDRAW_HISTORY_KEY, 'not json{');
    expect(readCaribbeanDrawHistory()).toEqual([]);
  });
  it('returns [] when the stored value is not an array', () => {
    localStorage.setItem(CARIBBEANDRAW_HISTORY_KEY, JSON.stringify({ wins: 3 }));
    expect(readCaribbeanDrawHistory()).toEqual([]);
  });
  it('filters out records with invalid shape', () => {
    localStorage.setItem(
      CARIBBEANDRAW_HISTORY_KEY,
      JSON.stringify([
        { outcome: CaribbeanDrawOutcome.WIN, net: 100 },
        { outcome: 9, net: 100 }, // invalid outcome code
        { outcome: CaribbeanDrawOutcome.LOSS, net: 'x' }, // invalid net
        { outcome: CaribbeanDrawOutcome.PUSH, net: Number.NaN }, // non-finite net
        42, // not an object
      ]),
    );
    expect(readCaribbeanDrawHistory()).toEqual([{ outcome: CaribbeanDrawOutcome.WIN, net: 100 }]);
  });
});

describe('useCaribbeanDrawStats', () => {
  it('recording a win increments the tally and net', () => {
    const { result } = renderHook(() => useCaribbeanDrawStats());
    expect(result.current.tally).toEqual({ wins: 0, losses: 0, pushes: 0, hands: 0, net: 0 });
    act(() => {
      result.current.recordRound({ outcome: CaribbeanDrawOutcome.WIN, net: 250 });
    });
    expect(result.current.tally).toEqual({ wins: 1, losses: 0, pushes: 0, hands: 1, net: 250 });
  });

  it('accumulates wins, losses, and pushes with signed net', () => {
    const { result } = renderHook(() => useCaribbeanDrawStats());
    act(() => {
      result.current.recordRound({ outcome: CaribbeanDrawOutcome.WIN, net: 200 });
    });
    act(() => {
      result.current.recordRound({ outcome: CaribbeanDrawOutcome.LOSS, net: -300 });
    });
    act(() => {
      result.current.recordRound({ outcome: CaribbeanDrawOutcome.PUSH, net: 0 });
    });
    expect(result.current.tally).toEqual({ wins: 1, losses: 1, pushes: 1, hands: 3, net: -100 });
  });

  it('persists the history to localStorage', () => {
    const { result } = renderHook(() => useCaribbeanDrawStats());
    act(() => {
      result.current.recordRound({ outcome: CaribbeanDrawOutcome.WIN, net: 120 });
    });
    const stored: CaribbeanDrawRecord[] = JSON.parse(localStorage.getItem(CARIBBEANDRAW_HISTORY_KEY) ?? '[]');
    expect(stored).toEqual([{ outcome: CaribbeanDrawOutcome.WIN, net: 120 }]);
  });

  it('rehydrates the history from localStorage on mount', () => {
    localStorage.setItem(CARIBBEANDRAW_HISTORY_KEY, JSON.stringify([{ outcome: CaribbeanDrawOutcome.LOSS, net: -50 }]));
    const { result } = renderHook(() => useCaribbeanDrawStats());
    expect(result.current.tally).toEqual({ wins: 0, losses: 1, pushes: 0, hands: 1, net: -50 });
  });

  it('clearHistory empties the tally and removes the storage key', () => {
    const { result } = renderHook(() => useCaribbeanDrawStats());
    act(() => {
      result.current.recordRound({ outcome: CaribbeanDrawOutcome.WIN, net: 90 });
    });
    act(() => {
      result.current.clearHistory();
    });
    expect(result.current.tally).toEqual({ wins: 0, losses: 0, pushes: 0, hands: 0, net: 0 });
    expect(localStorage.getItem(CARIBBEANDRAW_HISTORY_KEY)).toBeNull();
  });

  it('caps the stored history at CARIBBEANDRAW_HISTORY_MAX', () => {
    const { result } = renderHook(() => useCaribbeanDrawStats());
    act(() => {
      for (let i = 0; i < CARIBBEANDRAW_HISTORY_MAX + 5; i++) {
        result.current.recordRound({ outcome: CaribbeanDrawOutcome.WIN, net: 10 });
      }
    });
    expect(result.current.history).toHaveLength(CARIBBEANDRAW_HISTORY_MAX);
    expect(result.current.tally.hands).toBe(CARIBBEANDRAW_HISTORY_MAX);
  });
});
