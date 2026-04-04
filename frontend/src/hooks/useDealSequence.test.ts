import { act, renderHook } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { useDealSequence } from './useDealSequence';

describe('useDealSequence', () => {
  beforeEach(() => {
    vi.useFakeTimers();
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it('starts in idle state', () => {
    const { result } = renderHook(() => useDealSequence({ count: 4 }));
    expect(result.current.state).toBe('idle');
  });

  it('transitions to dealing on startDeal', () => {
    const { result } = renderHook(() => useDealSequence({ count: 4 }));
    act(() => result.current.startDeal());
    expect(result.current.state).toBe('dealing');
  });

  it('transitions to dealt after total duration', () => {
    const { result } = renderHook(() => useDealSequence({ count: 4, stagger: 0.12 }));
    act(() => result.current.startDeal());
    expect(result.current.state).toBe('dealing');
    // Total: 4 * 0.12 * 1000 + 300 = 780ms
    act(() => vi.advanceTimersByTime(1000));
    expect(result.current.state).toBe('dealt');
  });

  it('resets to idle', () => {
    const { result } = renderHook(() => useDealSequence({ count: 4 }));
    act(() => result.current.startDeal());
    expect(result.current.state).toBe('dealing');
    act(() => result.current.reset());
    expect(result.current.state).toBe('idle');
  });

  it('getDelay returns index * stagger when dealing', () => {
    const { result } = renderHook(() => useDealSequence({ count: 4, stagger: 0.1 }));
    act(() => result.current.startDeal());
    expect(result.current.getDelay(0)).toBe(0);
    expect(result.current.getDelay(1)).toBeCloseTo(0.1);
    expect(result.current.getDelay(2)).toBeCloseTo(0.2);
    expect(result.current.getDelay(3)).toBeCloseTo(0.3);
  });

  it('getDelay returns 0 when not dealing', () => {
    const { result } = renderHook(() => useDealSequence({ count: 4 }));
    expect(result.current.getDelay(0)).toBe(0);
    expect(result.current.getDelay(2)).toBe(0);
  });

  it('getDelay returns 0 when state is dealt', () => {
    const { result } = renderHook(() => useDealSequence({ count: 2, stagger: 0.1 }));
    act(() => result.current.startDeal());
    act(() => vi.advanceTimersByTime(1000));
    expect(result.current.state).toBe('dealt');
    expect(result.current.getDelay(0)).toBe(0);
  });

  it('handles count = 0 gracefully', () => {
    const { result } = renderHook(() => useDealSequence({ count: 0 }));
    act(() => result.current.startDeal());
    expect(result.current.state).toBe('idle');
  });

  it('uses custom stagger value', () => {
    const { result } = renderHook(() => useDealSequence({ count: 2, stagger: 0.5 }));
    act(() => result.current.startDeal());
    expect(result.current.getDelay(1)).toBeCloseTo(0.5);
  });

  it('uses default stagger of 0.12', () => {
    const { result } = renderHook(() => useDealSequence({ count: 2 }));
    act(() => result.current.startDeal());
    expect(result.current.getDelay(1)).toBeCloseTo(0.12);
  });

  it('clears timer on reset during dealing', () => {
    const { result } = renderHook(() => useDealSequence({ count: 4 }));
    act(() => result.current.startDeal());
    act(() => result.current.reset());
    // Advance past the would-be timeout
    act(() => vi.advanceTimersByTime(5000));
    expect(result.current.state).toBe('idle');
  });

  it('startDeal while already dealing restarts timer', () => {
    const { result } = renderHook(() => useDealSequence({ count: 4, stagger: 0.1 }));
    act(() => result.current.startDeal());
    act(() => vi.advanceTimersByTime(500));
    expect(result.current.state).toBe('dealing');
    // Restart
    act(() => result.current.startDeal());
    act(() => vi.advanceTimersByTime(500));
    expect(result.current.state).toBe('dealing');
    // Should complete after full duration from restart
    act(() => vi.advanceTimersByTime(600));
    expect(result.current.state).toBe('dealt');
  });
});
