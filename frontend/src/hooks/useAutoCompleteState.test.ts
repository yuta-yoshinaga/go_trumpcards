import { act, renderHook } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import { useAutoCompleteState } from './useAutoCompleteState';

describe('useAutoCompleteState', () => {
  it('starts with isAutoCompleting false', () => {
    const { result } = renderHook(() => useAutoCompleteState());
    expect(result.current.isAutoCompleting).toBe(false);
  });

  it('sets isAutoCompleting to true when startAutoComplete is called', () => {
    const { result } = renderHook(() => useAutoCompleteState());
    act(() => {
      result.current.startAutoComplete();
    });
    expect(result.current.isAutoCompleting).toBe(true);
  });

  it('resets isAutoCompleting to false after the default timeout', async () => {
    vi.useFakeTimers();
    const { result } = renderHook(() => useAutoCompleteState());
    act(() => {
      result.current.startAutoComplete();
    });
    expect(result.current.isAutoCompleting).toBe(true);

    act(() => {
      vi.advanceTimersByTime(3000);
    });
    expect(result.current.isAutoCompleting).toBe(false);
    vi.useRealTimers();
  });

  it('resets after custom duration', () => {
    vi.useFakeTimers();
    const { result } = renderHook(() => useAutoCompleteState(1000));
    act(() => {
      result.current.startAutoComplete();
    });
    expect(result.current.isAutoCompleting).toBe(true);

    act(() => {
      vi.advanceTimersByTime(1000);
    });
    expect(result.current.isAutoCompleting).toBe(false);
    vi.useRealTimers();
  });

  it('clears previous timer when startAutoComplete is called again', () => {
    vi.useFakeTimers();
    const { result } = renderHook(() => useAutoCompleteState(3000));
    act(() => {
      result.current.startAutoComplete();
    });

    // Advance 2s then call again — the old timer should be cleared
    act(() => {
      vi.advanceTimersByTime(2000);
    });
    expect(result.current.isAutoCompleting).toBe(true);

    act(() => {
      result.current.startAutoComplete();
    });

    // After 2s more the first timer would have fired, but it was cleared
    act(() => {
      vi.advanceTimersByTime(2000);
    });
    expect(result.current.isAutoCompleting).toBe(true);

    // Full 3s from second call
    act(() => {
      vi.advanceTimersByTime(1000);
    });
    expect(result.current.isAutoCompleting).toBe(false);
    vi.useRealTimers();
  });

  it('cleans up timer on unmount', () => {
    vi.useFakeTimers();
    const { result, unmount } = renderHook(() => useAutoCompleteState());
    act(() => {
      result.current.startAutoComplete();
    });
    unmount();
    // No error from dangling timer — clearTimeout was called in cleanup
    act(() => {
      vi.advanceTimersByTime(5000);
    });
    vi.useRealTimers();
  });
});
