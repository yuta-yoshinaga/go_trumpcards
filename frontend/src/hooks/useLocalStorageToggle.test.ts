import { act, renderHook } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it } from 'vitest';
import { useLocalStorageToggle } from './useLocalStorageToggle';

const KEY = 'test_toggle_key';

describe('useLocalStorageToggle', () => {
  beforeEach(() => {
    localStorage.clear();
  });

  afterEach(() => {
    localStorage.clear();
  });

  it('returns defaultValue when localStorage has no entry', () => {
    const { result } = renderHook(() => useLocalStorageToggle(KEY, false));
    expect(result.current[0]).toBe(false);
  });

  it('returns true defaultValue when specified', () => {
    const { result } = renderHook(() => useLocalStorageToggle(KEY, true));
    expect(result.current[0]).toBe(true);
  });

  it('reads existing true value from localStorage', () => {
    localStorage.setItem(KEY, 'true');
    const { result } = renderHook(() => useLocalStorageToggle(KEY, false));
    expect(result.current[0]).toBe(true);
  });

  it('reads existing false value from localStorage', () => {
    localStorage.setItem(KEY, 'false');
    const { result } = renderHook(() => useLocalStorageToggle(KEY, true));
    expect(result.current[0]).toBe(false);
  });

  it('updates state and localStorage when setter is called with true', () => {
    const { result } = renderHook(() => useLocalStorageToggle(KEY, false));
    act(() => {
      result.current[1](true);
    });
    expect(result.current[0]).toBe(true);
    expect(localStorage.getItem(KEY)).toBe('true');
  });

  it('updates state and localStorage when setter is called with false', () => {
    const { result } = renderHook(() => useLocalStorageToggle(KEY, true));
    act(() => {
      result.current[1](false);
    });
    expect(result.current[0]).toBe(false);
    expect(localStorage.getItem(KEY)).toBe('false');
  });

  it('returns stable setter reference across re-renders', () => {
    const { result, rerender } = renderHook(() => useLocalStorageToggle(KEY, false));
    const setter1 = result.current[1];
    rerender();
    expect(result.current[1]).toBe(setter1);
  });

  it('ignores non-boolean localStorage values and uses default', () => {
    localStorage.setItem(KEY, 'invalid');
    const { result } = renderHook(() => useLocalStorageToggle(KEY, false));
    expect(result.current[0]).toBe(false);
  });
});
