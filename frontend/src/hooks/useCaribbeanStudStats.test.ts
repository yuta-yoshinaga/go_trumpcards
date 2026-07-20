import { act, renderHook } from '@testing-library/react';
import { beforeEach, describe, expect, it } from 'vitest';
import {
  CARIBBEANSTUD_HISTORY_KEY,
  CARIBBEANSTUD_HISTORY_MAX,
  CaribbeanStudOutcome,
  type CaribbeanStudRecord,
  outcomeFromResult,
  readCaribbeanStudHistory,
  tallyCaribbeanStudHistory,
  useCaribbeanStudStats,
} from './useCaribbeanStudStats';

beforeEach(() => {
  localStorage.clear();
});

describe('outcomeFromResult', () => {
  it('maps a positive result to WIN', () => {
    expect(outcomeFromResult(1)).toBe(CaribbeanStudOutcome.WIN);
  });
  it('maps a negative result to LOSS', () => {
    expect(outcomeFromResult(-1)).toBe(CaribbeanStudOutcome.LOSS);
  });
  it('maps a zero result to PUSH', () => {
    expect(outcomeFromResult(0)).toBe(CaribbeanStudOutcome.PUSH);
  });
});

describe('tallyCaribbeanStudHistory', () => {
  it('counts wins, losses, pushes, hands, and net chips', () => {
    const tally = tallyCaribbeanStudHistory([
      { outcome: CaribbeanStudOutcome.WIN, net: 300 },
      { outcome: CaribbeanStudOutcome.LOSS, net: -100 },
      { outcome: CaribbeanStudOutcome.WIN, net: 200 },
      { outcome: CaribbeanStudOutcome.PUSH, net: 0 },
    ]);
    expect(tally).toEqual({ wins: 2, losses: 1, pushes: 1, hands: 4, net: 400 });
  });
  it('returns zeros for empty history', () => {
    expect(tallyCaribbeanStudHistory([])).toEqual({ wins: 0, losses: 0, pushes: 0, hands: 0, net: 0 });
  });
});

describe('readCaribbeanStudHistory', () => {
  it('returns [] when nothing is stored', () => {
    expect(readCaribbeanStudHistory()).toEqual([]);
  });
  it('returns [] on corrupt (non-JSON) data', () => {
    localStorage.setItem(CARIBBEANSTUD_HISTORY_KEY, 'not json{');
    expect(readCaribbeanStudHistory()).toEqual([]);
  });
  it('returns [] when the stored value is not an array', () => {
    localStorage.setItem(CARIBBEANSTUD_HISTORY_KEY, JSON.stringify({ wins: 3 }));
    expect(readCaribbeanStudHistory()).toEqual([]);
  });
  it('filters out records with invalid shape', () => {
    localStorage.setItem(
      CARIBBEANSTUD_HISTORY_KEY,
      JSON.stringify([
        { outcome: CaribbeanStudOutcome.WIN, net: 100 },
        { outcome: 9, net: 100 }, // invalid outcome code
        { outcome: CaribbeanStudOutcome.LOSS, net: 'x' }, // invalid net
        { outcome: CaribbeanStudOutcome.PUSH, net: Number.NaN }, // non-finite net
        42, // not an object
      ]),
    );
    expect(readCaribbeanStudHistory()).toEqual([{ outcome: CaribbeanStudOutcome.WIN, net: 100 }]);
  });
});

describe('useCaribbeanStudStats', () => {
  it('recording a win increments the tally and net', () => {
    const { result } = renderHook(() => useCaribbeanStudStats());
    expect(result.current.tally).toEqual({ wins: 0, losses: 0, pushes: 0, hands: 0, net: 0 });
    act(() => {
      result.current.recordRound({ outcome: CaribbeanStudOutcome.WIN, net: 250 });
    });
    expect(result.current.tally).toEqual({ wins: 1, losses: 0, pushes: 0, hands: 1, net: 250 });
  });

  it('accumulates wins, losses, and pushes with signed net', () => {
    const { result } = renderHook(() => useCaribbeanStudStats());
    act(() => {
      result.current.recordRound({ outcome: CaribbeanStudOutcome.WIN, net: 200 });
    });
    act(() => {
      result.current.recordRound({ outcome: CaribbeanStudOutcome.LOSS, net: -300 });
    });
    act(() => {
      result.current.recordRound({ outcome: CaribbeanStudOutcome.PUSH, net: 0 });
    });
    expect(result.current.tally).toEqual({ wins: 1, losses: 1, pushes: 1, hands: 3, net: -100 });
  });

  it('persists the history to localStorage', () => {
    const { result } = renderHook(() => useCaribbeanStudStats());
    act(() => {
      result.current.recordRound({ outcome: CaribbeanStudOutcome.WIN, net: 120 });
    });
    const stored: CaribbeanStudRecord[] = JSON.parse(localStorage.getItem(CARIBBEANSTUD_HISTORY_KEY) ?? '[]');
    expect(stored).toEqual([{ outcome: CaribbeanStudOutcome.WIN, net: 120 }]);
  });

  it('rehydrates the history from localStorage on mount', () => {
    localStorage.setItem(CARIBBEANSTUD_HISTORY_KEY, JSON.stringify([{ outcome: CaribbeanStudOutcome.LOSS, net: -50 }]));
    const { result } = renderHook(() => useCaribbeanStudStats());
    expect(result.current.tally).toEqual({ wins: 0, losses: 1, pushes: 0, hands: 1, net: -50 });
  });

  it('clearHistory empties the tally and removes the storage key', () => {
    const { result } = renderHook(() => useCaribbeanStudStats());
    act(() => {
      result.current.recordRound({ outcome: CaribbeanStudOutcome.WIN, net: 90 });
    });
    act(() => {
      result.current.clearHistory();
    });
    expect(result.current.tally).toEqual({ wins: 0, losses: 0, pushes: 0, hands: 0, net: 0 });
    expect(localStorage.getItem(CARIBBEANSTUD_HISTORY_KEY)).toBeNull();
  });

  it('caps the stored history at CARIBBEANSTUD_HISTORY_MAX', () => {
    const { result } = renderHook(() => useCaribbeanStudStats());
    act(() => {
      for (let i = 0; i < CARIBBEANSTUD_HISTORY_MAX + 5; i++) {
        result.current.recordRound({ outcome: CaribbeanStudOutcome.WIN, net: 10 });
      }
    });
    expect(result.current.history).toHaveLength(CARIBBEANSTUD_HISTORY_MAX);
    expect(result.current.tally.hands).toBe(CARIBBEANSTUD_HISTORY_MAX);
  });
});
