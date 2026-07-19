import { act, renderHook } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { useSpeedTimer } from './useSpeedTimer';

const KEY = (d: number) => `speed_best_time_${d}`;

describe('useSpeedTimer', () => {
  beforeEach(() => {
    localStorage.clear();
    vi.useFakeTimers();
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it('starts at 0 elapsed with no best when nothing persisted', () => {
    const { result } = renderHook(() => useSpeedTimer(false, false, false, 0));
    expect(result.current.elapsedMs).toBe(0);
    expect(result.current.bestMs).toBeNull();
    expect(result.current.isNewBest).toBe(false);
  });

  it('advances elapsed while running (Date.now based)', () => {
    const { result } = renderHook(() => useSpeedTimer(true, false, false, 0));
    act(() => {
      vi.advanceTimersByTime(3000);
    });
    expect(result.current.elapsedMs).toBe(3000);
  });

  it('freezes elapsed at the final time when the game ends', () => {
    const { result, rerender } = renderHook(({ running, ended, won }) => useSpeedTimer(running, ended, won, 0), {
      initialProps: { running: true, ended: false, won: false },
    });
    act(() => {
      vi.advanceTimersByTime(4200);
    });
    rerender({ running: false, ended: true, won: true });
    act(() => {
      vi.advanceTimersByTime(5000);
    });
    expect(result.current.elapsedMs).toBe(4200);
  });

  it('records the best time and flags a new best on a human win', () => {
    const { result, rerender } = renderHook(({ running, ended, won }) => useSpeedTimer(running, ended, won, 1), {
      initialProps: { running: true, ended: false, won: false },
    });
    act(() => {
      vi.advanceTimersByTime(8000);
    });
    rerender({ running: false, ended: true, won: true });
    expect(result.current.isNewBest).toBe(true);
    expect(result.current.bestMs).toBe(8000);
    expect(localStorage.getItem(KEY(1))).toBe('8000');
  });

  it('does not record the best time on a loss', () => {
    const { result, rerender } = renderHook(({ running, ended, won }) => useSpeedTimer(running, ended, won, 1), {
      initialProps: { running: true, ended: false, won: false },
    });
    act(() => {
      vi.advanceTimersByTime(6000);
    });
    rerender({ running: false, ended: true, won: false });
    expect(result.current.isNewBest).toBe(false);
    expect(result.current.bestMs).toBeNull();
    expect(localStorage.getItem(KEY(1))).toBeNull();
  });

  it('reads an existing best from localStorage', () => {
    localStorage.setItem(KEY(2), '5000');
    const { result } = renderHook(() => useSpeedTimer(false, false, false, 2));
    expect(result.current.bestMs).toBe(5000);
  });

  it('updates the best only when the new win is faster', () => {
    localStorage.setItem(KEY(0), '10000');
    // Slower win: 12s > 10s stored -> no update.
    const slow = renderHook(({ running, ended, won }) => useSpeedTimer(running, ended, won, 0), {
      initialProps: { running: true, ended: false, won: false },
    });
    act(() => {
      vi.advanceTimersByTime(12000);
    });
    slow.rerender({ running: false, ended: true, won: true });
    expect(slow.result.current.isNewBest).toBe(false);
    expect(localStorage.getItem(KEY(0))).toBe('10000');

    // Faster win: 7s < 10s stored -> update.
    const fast = renderHook(({ running, ended, won }) => useSpeedTimer(running, ended, won, 0), {
      initialProps: { running: true, ended: false, won: false },
    });
    act(() => {
      vi.advanceTimersByTime(7000);
    });
    fast.rerender({ running: false, ended: true, won: true });
    expect(fast.result.current.isNewBest).toBe(true);
    expect(localStorage.getItem(KEY(0))).toBe('7000');
  });

  it('restarts the timer when a new game begins (running rises again)', () => {
    const { result, rerender } = renderHook(({ running, ended, won }) => useSpeedTimer(running, ended, won, 0), {
      initialProps: { running: true, ended: false, won: false },
    });
    act(() => {
      vi.advanceTimersByTime(9000);
    });
    rerender({ running: false, ended: true, won: true });
    expect(result.current.elapsedMs).toBe(9000);

    // Reset -> fresh PLAY state: elapsed resets to 0 and isNewBest clears.
    rerender({ running: true, ended: false, won: false });
    expect(result.current.elapsedMs).toBe(0);
    expect(result.current.isNewBest).toBe(false);
    act(() => {
      vi.advanceTimersByTime(1500);
    });
    expect(result.current.elapsedMs).toBe(1500);
  });

  it('records only once per game even if ended re-renders', () => {
    const { rerender } = renderHook(({ running, ended, won }) => useSpeedTimer(running, ended, won, 3), {
      initialProps: { running: true, ended: false, won: false },
    });
    act(() => {
      vi.advanceTimersByTime(4000);
    });
    rerender({ running: false, ended: true, won: true });
    expect(localStorage.getItem(KEY(3))).toBe('4000');
    // A faster time is manually stored; a stray re-render must not overwrite it.
    localStorage.setItem(KEY(3), '1000');
    rerender({ running: false, ended: true, won: true });
    expect(localStorage.getItem(KEY(3))).toBe('1000');
  });

  it('reloads the best when the difficulty changes', () => {
    localStorage.setItem(KEY(0), '3000');
    localStorage.setItem(KEY(1), '9000');
    const { result, rerender } = renderHook(({ d }) => useSpeedTimer(false, false, false, d), {
      initialProps: { d: 0 },
    });
    expect(result.current.bestMs).toBe(3000);
    rerender({ d: 1 });
    expect(result.current.bestMs).toBe(9000);
  });
});
