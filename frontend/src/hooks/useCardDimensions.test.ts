import { act, renderHook } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { CARD_DIMENSIONS, useCardDimensions } from './useCardDimensions';

describe('useCardDimensions', () => {
  const originalInnerWidth = window.innerWidth;

  beforeEach(() => {
    vi.spyOn(window, 'addEventListener');
    vi.spyOn(window, 'removeEventListener');
  });

  afterEach(() => {
    Object.defineProperty(window, 'innerWidth', {
      writable: true,
      configurable: true,
      value: originalInnerWidth,
    });
  });

  it('returns mobile dimensions when width is 0 (jsdom default)', () => {
    Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: 0 });
    const { result } = renderHook(() => useCardDimensions());
    expect(result.current).toEqual(CARD_DIMENSIONS.mobile);
  });

  it('returns mobile dimensions when width is below breakpoint (375px)', () => {
    Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: 375 });
    const { result } = renderHook(() => useCardDimensions());
    expect(result.current).toEqual(CARD_DIMENSIONS.mobile);
  });

  it('returns mobile dimensions when width equals 639px (just below breakpoint)', () => {
    Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: 639 });
    const { result } = renderHook(() => useCardDimensions());
    expect(result.current).toEqual(CARD_DIMENSIONS.mobile);
  });

  it('returns desktop dimensions when width equals breakpoint (640px)', () => {
    Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: 640 });
    const { result } = renderHook(() => useCardDimensions());
    expect(result.current).toEqual(CARD_DIMENSIONS.desktop);
  });

  it('returns desktop dimensions when width is above breakpoint (1024px)', () => {
    Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: 1024 });
    const { result } = renderHook(() => useCardDimensions());
    expect(result.current).toEqual(CARD_DIMENSIONS.desktop);
  });

  it('updates dimensions when window is resized to mobile', () => {
    Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: 1024 });
    const { result } = renderHook(() => useCardDimensions());
    expect(result.current).toEqual(CARD_DIMENSIONS.desktop);

    act(() => {
      Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: 375 });
      window.dispatchEvent(new Event('resize'));
    });

    expect(result.current).toEqual(CARD_DIMENSIONS.mobile);
  });

  it('updates dimensions when window is resized to desktop', () => {
    Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: 375 });
    const { result } = renderHook(() => useCardDimensions());
    expect(result.current).toEqual(CARD_DIMENSIONS.mobile);

    act(() => {
      Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: 1024 });
      window.dispatchEvent(new Event('resize'));
    });

    expect(result.current).toEqual(CARD_DIMENSIONS.desktop);
  });

  it('removes resize listener on unmount', () => {
    Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: 1024 });
    const { unmount } = renderHook(() => useCardDimensions());
    unmount();
    expect(window.removeEventListener).toHaveBeenCalledWith('resize', expect.any(Function));
  });

  it('mobile dimensions include sevens cell fields', () => {
    Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: 375 });
    const { result } = renderHook(() => useCardDimensions());
    expect(result.current.sevensCellSize).toBe(20);
    expect(result.current.sevensFontSize).toBe('0.6em');
  });

  it('desktop dimensions include sevens cell fields', () => {
    Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: 1024 });
    const { result } = renderHook(() => useCardDimensions());
    expect(result.current.sevensCellSize).toBe(26);
    expect(result.current.sevensFontSize).toBe('0.75em');
  });
});
