import { act, renderHook } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { useWhiteheadTimer } from './useWhiteheadTimer';

describe('useWhiteheadTimer', () => {
  beforeEach(() => {
    vi.useFakeTimers({ toFake: ['setInterval', 'clearInterval'] });
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it('starts at 0 seconds', () => {
    const { result } = renderHook(() => useWhiteheadTimer(false));
    expect(result.current.elapsedSeconds).toBe(0);
  });

  it('increments when isPlaying is true', () => {
    const { result } = renderHook(() => useWhiteheadTimer(true));
    act(() => {
      vi.advanceTimersByTime(3000);
    });
    expect(result.current.elapsedSeconds).toBe(3);
  });

  it('stops incrementing when isPlaying becomes false', () => {
    const { result, rerender } = renderHook(({ playing }) => useWhiteheadTimer(playing), {
      initialProps: { playing: true },
    });
    act(() => {
      vi.advanceTimersByTime(2000);
    });
    expect(result.current.elapsedSeconds).toBe(2);

    rerender({ playing: false });
    act(() => {
      vi.advanceTimersByTime(3000);
    });
    expect(result.current.elapsedSeconds).toBe(2);
  });

  it('resetTimer resets to 0', () => {
    const { result } = renderHook(() => useWhiteheadTimer(true));
    act(() => {
      vi.advanceTimersByTime(5000);
    });
    expect(result.current.elapsedSeconds).toBe(5);

    act(() => {
      result.current.resetTimer();
    });
    expect(result.current.elapsedSeconds).toBe(0);
  });

  it('timeBonus returns 0 for 0 seconds', () => {
    const { result } = renderHook(() => useWhiteheadTimer(false));
    expect(result.current.timeBonus(0)).toBe(0);
  });

  it('timeBonus calculates correctly for positive seconds', () => {
    const { result } = renderHook(() => useWhiteheadTimer(false));
    expect(result.current.timeBonus(100)).toBe(7000);
    expect(result.current.timeBonus(700)).toBe(1000);
  });

  it('timeBonus for very fast time', () => {
    const { result } = renderHook(() => useWhiteheadTimer(false));
    expect(result.current.timeBonus(1)).toBe(700000);
  });

  it('timeBonus for very slow time', () => {
    const { result } = renderHook(() => useWhiteheadTimer(false));
    expect(result.current.timeBonus(1000000)).toBe(0);
  });

  it('restarts timer when isPlaying toggles off then on', () => {
    const { result, rerender } = renderHook(({ playing }) => useWhiteheadTimer(playing), {
      initialProps: { playing: true },
    });
    act(() => {
      vi.advanceTimersByTime(3000);
    });
    expect(result.current.elapsedSeconds).toBe(3);

    rerender({ playing: false });
    rerender({ playing: true });
    act(() => {
      vi.advanceTimersByTime(2000);
    });
    // Timer continues from where it was (3 + 2 = 5) since we didn't resetTimer
    expect(result.current.elapsedSeconds).toBe(5);
  });
});
