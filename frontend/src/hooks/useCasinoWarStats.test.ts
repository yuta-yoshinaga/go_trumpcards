import { act, renderHook } from '@testing-library/react';
import { beforeEach, describe, expect, it } from 'vitest';
import {
  CASINOWAR_HISTORY_KEY,
  CASINOWAR_HISTORY_MAX,
  CasinoWarOutcome,
  outcomeFromResult,
  readCasinoWarHistory,
  tallyCasinoWarHistory,
  useCasinoWarStats,
} from './useCasinoWarStats';

beforeEach(() => {
  localStorage.clear();
});

describe('outcomeFromResult', () => {
  it('maps a positive result to WIN', () => {
    expect(outcomeFromResult(150)).toBe(CasinoWarOutcome.WIN);
  });
  it('maps a negative result to LOSS', () => {
    expect(outcomeFromResult(-100)).toBe(CasinoWarOutcome.LOSS);
  });
  it('maps a zero result to TIE (push)', () => {
    expect(outcomeFromResult(0)).toBe(CasinoWarOutcome.TIE);
  });
});

describe('tallyCasinoWarHistory', () => {
  it('counts wins, losses, and ties', () => {
    const tally = tallyCasinoWarHistory([
      CasinoWarOutcome.WIN,
      CasinoWarOutcome.LOSS,
      CasinoWarOutcome.WIN,
      CasinoWarOutcome.TIE,
    ]);
    expect(tally).toEqual({ wins: 2, losses: 1, ties: 1 });
  });
  it('returns zeros for empty history', () => {
    expect(tallyCasinoWarHistory([])).toEqual({ wins: 0, losses: 0, ties: 0 });
  });
});

describe('readCasinoWarHistory', () => {
  it('returns [] when nothing is stored', () => {
    expect(readCasinoWarHistory()).toEqual([]);
  });
  it('returns [] on corrupt (non-JSON) data', () => {
    localStorage.setItem(CASINOWAR_HISTORY_KEY, 'not-json{');
    expect(readCasinoWarHistory()).toEqual([]);
  });
  it('returns [] when the stored value is not an array', () => {
    localStorage.setItem(CASINOWAR_HISTORY_KEY, JSON.stringify({ foo: 1 }));
    expect(readCasinoWarHistory()).toEqual([]);
  });
  it('drops non-outcome entries from a partially corrupt array', () => {
    localStorage.setItem(CASINOWAR_HISTORY_KEY, JSON.stringify([1, 'x', 2, 9, 0]));
    expect(readCasinoWarHistory()).toEqual([CasinoWarOutcome.WIN, CasinoWarOutcome.LOSS, CasinoWarOutcome.TIE]);
  });
});

describe('useCasinoWarStats', () => {
  it('recording a win increments the win tally and appends a pip', () => {
    const { result } = renderHook(() => useCasinoWarStats());
    expect(result.current.history).toEqual([]);
    act(() => result.current.recordOutcome(CasinoWarOutcome.WIN));
    expect(result.current.history).toEqual([CasinoWarOutcome.WIN]);
    expect(result.current.tally).toEqual({ wins: 1, losses: 0, ties: 0 });
  });

  it('persists recorded outcomes to localStorage', () => {
    const { result } = renderHook(() => useCasinoWarStats());
    act(() => result.current.recordOutcome(CasinoWarOutcome.WIN));
    act(() => result.current.recordOutcome(CasinoWarOutcome.LOSS));
    expect(readCasinoWarHistory()).toEqual([CasinoWarOutcome.WIN, CasinoWarOutcome.LOSS]);
  });

  it('rehydrates history from localStorage on mount', () => {
    localStorage.setItem(CASINOWAR_HISTORY_KEY, JSON.stringify([CasinoWarOutcome.TIE, CasinoWarOutcome.WIN]));
    const { result } = renderHook(() => useCasinoWarStats());
    expect(result.current.history).toEqual([CasinoWarOutcome.TIE, CasinoWarOutcome.WIN]);
    expect(result.current.tally).toEqual({ wins: 1, losses: 0, ties: 1 });
  });

  it('caps stored history at CASINOWAR_HISTORY_MAX (keeping the most recent)', () => {
    const { result } = renderHook(() => useCasinoWarStats());
    act(() => {
      for (let i = 0; i < CASINOWAR_HISTORY_MAX + 5; i += 1) {
        result.current.recordOutcome(CasinoWarOutcome.WIN);
      }
    });
    expect(result.current.history).toHaveLength(CASINOWAR_HISTORY_MAX);
  });

  it('clearHistory empties the history and removes the storage key', () => {
    const { result } = renderHook(() => useCasinoWarStats());
    act(() => result.current.recordOutcome(CasinoWarOutcome.WIN));
    act(() => result.current.clearHistory());
    expect(result.current.history).toEqual([]);
    expect(localStorage.getItem(CASINOWAR_HISTORY_KEY)).toBeNull();
  });
});
