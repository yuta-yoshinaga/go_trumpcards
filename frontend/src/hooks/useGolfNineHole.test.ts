import { act, renderHook } from '@testing-library/react';
import { beforeEach, describe, expect, it } from 'vitest';
import type { GolfCard } from '../types/card';
import {
  countGolfRemaining,
  emptyGolfNineHoleState,
  GOLF_NINE_HOLE_KEY,
  GOLF_TOTAL_HOLES,
  golfCurrentHole,
  golfNineHoleComplete,
  golfNineHoleTotal,
  readGolfNineHoleState,
  recordGolfHole,
  useGolfNineHole,
} from './useGolfNineHole';

const gc = (removed: boolean): GolfCard => ({ card: { design: 'SPADE', value: 5 }, removed, exposed: true });

beforeEach(() => {
  localStorage.clear();
});

describe('countGolfRemaining', () => {
  it('counts cards that are present and not removed', () => {
    const layout: GolfCard[][] = [
      [gc(false), gc(true), { card: null, removed: true, exposed: false }],
      [gc(false), gc(false)],
    ];
    expect(countGolfRemaining(layout)).toBe(3);
  });

  it('is 0 when every card is removed (a cleared deal)', () => {
    const layout: GolfCard[][] = [[gc(true), gc(true)], [gc(true)]];
    expect(countGolfRemaining(layout)).toBe(0);
  });
});

describe('recordGolfHole (pure accumulation)', () => {
  it('sums two deals correctly into the scorecard', () => {
    let state = emptyGolfNineHoleState();
    state = recordGolfHole(state, 7);
    state = recordGolfHole(state, 4);
    expect(state.scores).toEqual([7, 4]);
    expect(golfNineHoleTotal(state)).toBe(11);
    expect(golfCurrentHole(state)).toBe(3);
    expect(golfNineHoleComplete(state)).toBe(false);
  });

  it('caps at 9 holes and ignores extra records (no double-count past the round)', () => {
    let state = emptyGolfNineHoleState();
    for (let i = 0; i < GOLF_TOTAL_HOLES; i++) state = recordGolfHole(state, 2);
    expect(golfNineHoleComplete(state)).toBe(true);
    expect(golfCurrentHole(state)).toBe(GOLF_TOTAL_HOLES);
    const before = state;
    state = recordGolfHole(state, 99);
    expect(state).toBe(before);
    expect(golfNineHoleTotal(state)).toBe(18);
  });

  it('lower cumulative total is the better (winning) 9-hole result', () => {
    const good = recordGolfHole(recordGolfHole(emptyGolfNineHoleState(), 1), 2);
    const bad = recordGolfHole(recordGolfHole(emptyGolfNineHoleState(), 6), 8);
    expect(golfNineHoleTotal(good)).toBeLessThan(golfNineHoleTotal(bad));
  });
});

describe('readGolfNineHoleState', () => {
  it('returns a zeroed state when nothing is stored', () => {
    expect(readGolfNineHoleState()).toEqual({ enabled: false, scores: [] });
  });

  it('restores a persisted state', () => {
    localStorage.setItem(GOLF_NINE_HOLE_KEY, JSON.stringify({ enabled: true, scores: [3, 5] }));
    expect(readGolfNineHoleState()).toEqual({ enabled: true, scores: [3, 5] });
  });

  it('falls back to zeroed state on malformed JSON', () => {
    localStorage.setItem(GOLF_NINE_HOLE_KEY, '{not json');
    expect(readGolfNineHoleState()).toEqual({ enabled: false, scores: [] });
  });

  it('rejects an over-long scorecard', () => {
    localStorage.setItem(
      GOLF_NINE_HOLE_KEY,
      JSON.stringify({ enabled: true, scores: new Array(GOLF_TOTAL_HOLES + 1).fill(1) }),
    );
    expect(readGolfNineHoleState()).toEqual({ enabled: false, scores: [] });
  });
});

describe('useGolfNineHole', () => {
  it('enabling starts a fresh persisted scorecard', () => {
    const { result } = renderHook(() => useGolfNineHole());
    act(() => result.current.setEnabled(true));
    expect(result.current.nineHole).toEqual({ enabled: true, scores: [] });
    expect(readGolfNineHoleState()).toEqual({ enabled: true, scores: [] });
  });

  it('recordHole accumulates and persists across deals', () => {
    const { result } = renderHook(() => useGolfNineHole());
    act(() => result.current.setEnabled(true));
    act(() => result.current.recordHole(6));
    act(() => result.current.recordHole(2));
    expect(result.current.nineHole.scores).toEqual([6, 2]);
    expect(readGolfNineHoleState().scores).toEqual([6, 2]);
  });

  it('resetCard clears scores but keeps the mode on', () => {
    const { result } = renderHook(() => useGolfNineHole());
    act(() => result.current.setEnabled(true));
    act(() => result.current.recordHole(6));
    act(() => result.current.resetCard());
    expect(result.current.nineHole).toEqual({ enabled: true, scores: [] });
  });
});
