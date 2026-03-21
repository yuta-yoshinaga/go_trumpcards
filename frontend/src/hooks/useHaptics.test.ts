import { afterEach, beforeEach, describe, expect, it, vi } from 'bun:test';
import { renderHook } from '@testing-library/react';
import { useHaptics } from './useHaptics';

describe('useHaptics', () => {
  let vibrateSpy: ReturnType<typeof vi.fn>;

  beforeEach(() => {
    vibrateSpy = vi.fn();
    Object.defineProperty(navigator, 'vibrate', {
      writable: true,
      configurable: true,
      value: vibrateSpy,
    });
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it('tapVibrate calls navigator.vibrate with 10ms', () => {
    const { result } = renderHook(() => useHaptics());
    result.current.tapVibrate();
    expect(vibrateSpy).toHaveBeenCalledWith(10);
  });

  it('selectVibrate calls navigator.vibrate with 20ms', () => {
    const { result } = renderHook(() => useHaptics());
    result.current.selectVibrate();
    expect(vibrateSpy).toHaveBeenCalledWith(20);
  });

  it('winVibrate calls navigator.vibrate with pattern', () => {
    const { result } = renderHook(() => useHaptics());
    result.current.winVibrate();
    expect(vibrateSpy).toHaveBeenCalledWith([50, 30, 50]);
  });

  it('canVibrate is true when navigator.vibrate exists', () => {
    const { result } = renderHook(() => useHaptics());
    expect(result.current.canVibrate).toBe(true);
  });

  it('does not call vibrate when unsupported', () => {
    Object.defineProperty(navigator, 'vibrate', {
      writable: true,
      configurable: true,
      value: undefined,
    });
    const { result } = renderHook(() => useHaptics());
    expect(result.current.canVibrate).toBe(false);
    result.current.tapVibrate();
    result.current.selectVibrate();
    result.current.winVibrate();
    // No error thrown
  });
});
