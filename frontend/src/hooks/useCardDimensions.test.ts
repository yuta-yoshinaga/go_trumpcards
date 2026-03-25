import { act, renderHook } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import {
  CARD_DIMENSIONS,
  useCardDimensions,
  useIsLargeDesktop,
  useIsMediumDesktop,
  useWindowWidth,
} from './useCardDimensions';

const setWidth = (w: number) =>
  Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: w });

describe('useWindowWidth', () => {
  const originalInnerWidth = window.innerWidth;

  afterEach(() => setWidth(originalInnerWidth));

  it('returns current window.innerWidth', () => {
    setWidth(800);
    const { result } = renderHook(() => useWindowWidth());
    expect(result.current).toBe(800);
  });

  it('updates on resize', () => {
    setWidth(800);
    const { result } = renderHook(() => useWindowWidth());
    act(() => {
      setWidth(400);
      window.dispatchEvent(new Event('resize'));
    });
    expect(result.current).toBe(400);
  });

  it('removes listener on unmount', () => {
    vi.spyOn(window, 'removeEventListener');
    const { unmount } = renderHook(() => useWindowWidth());
    unmount();
    expect(window.removeEventListener).toHaveBeenCalledWith('resize', expect.any(Function));
  });
});

describe('useCardDimensions', () => {
  const originalInnerWidth = window.innerWidth;

  beforeEach(() => {
    vi.spyOn(window, 'addEventListener');
    vi.spyOn(window, 'removeEventListener');
  });

  afterEach(() => setWidth(originalInnerWidth));

  it('returns mobile dimensions when width is 0 (jsdom default)', () => {
    setWidth(0);
    const { result } = renderHook(() => useCardDimensions());
    expect(result.current).toEqual({ ...CARD_DIMENSIONS.mobile, isMobile: true });
  });

  it('returns mobile dimensions when width is below breakpoint (375px)', () => {
    setWidth(375);
    const { result } = renderHook(() => useCardDimensions());
    expect(result.current).toEqual({ ...CARD_DIMENSIONS.mobile, isMobile: true });
  });

  it('returns mobile dimensions when width equals 639px (just below sm breakpoint)', () => {
    setWidth(639);
    const { result } = renderHook(() => useCardDimensions());
    expect(result.current).toEqual({ ...CARD_DIMENSIONS.mobile, isMobile: true });
  });

  it('returns desktop dimensions when width equals sm breakpoint (640px)', () => {
    setWidth(640);
    const { result } = renderHook(() => useCardDimensions());
    expect(result.current).toEqual({ ...CARD_DIMENSIONS.desktop, isMobile: false });
  });

  it('returns desktop dimensions when width is just below lg breakpoint (1023px)', () => {
    setWidth(1023);
    const { result } = renderHook(() => useCardDimensions());
    expect(result.current).toEqual({ ...CARD_DIMENSIONS.desktop, isMobile: false });
  });

  it('returns largeDesktop dimensions when width equals lg breakpoint (1024px)', () => {
    setWidth(1024);
    const { result } = renderHook(() => useCardDimensions());
    expect(result.current).toEqual({ ...CARD_DIMENSIONS.largeDesktop, isMobile: false });
  });

  it('returns largeDesktop dimensions when width is above lg breakpoint (1280px)', () => {
    setWidth(1280);
    const { result } = renderHook(() => useCardDimensions());
    expect(result.current).toEqual({ ...CARD_DIMENSIONS.largeDesktop, isMobile: false });
  });

  it('updates dimensions when window is resized to mobile', () => {
    setWidth(1024);
    const { result } = renderHook(() => useCardDimensions());
    expect(result.current).toEqual({ ...CARD_DIMENSIONS.largeDesktop, isMobile: false });

    act(() => {
      setWidth(375);
      window.dispatchEvent(new Event('resize'));
    });

    expect(result.current).toEqual({ ...CARD_DIMENSIONS.mobile, isMobile: true });
  });

  it('updates dimensions when window is resized to desktop', () => {
    setWidth(375);
    const { result } = renderHook(() => useCardDimensions());
    expect(result.current).toEqual({ ...CARD_DIMENSIONS.mobile, isMobile: true });

    act(() => {
      setWidth(800);
      window.dispatchEvent(new Event('resize'));
    });

    expect(result.current).toEqual({ ...CARD_DIMENSIONS.desktop, isMobile: false });
  });

  it('updates dimensions when window is resized to largeDesktop', () => {
    setWidth(375);
    const { result } = renderHook(() => useCardDimensions());
    expect(result.current).toEqual({ ...CARD_DIMENSIONS.mobile, isMobile: true });

    act(() => {
      setWidth(1280);
      window.dispatchEvent(new Event('resize'));
    });

    expect(result.current).toEqual({ ...CARD_DIMENSIONS.largeDesktop, isMobile: false });
  });

  it('removes resize listener on unmount', () => {
    setWidth(1024);
    const { unmount } = renderHook(() => useCardDimensions());
    unmount();
    expect(window.removeEventListener).toHaveBeenCalledWith('resize', expect.any(Function));
  });
});

describe('useIsLargeDesktop', () => {
  const originalInnerWidth = window.innerWidth;

  afterEach(() => setWidth(originalInnerWidth));

  it('returns false when width is below lg breakpoint', () => {
    setWidth(800);
    const { result } = renderHook(() => useIsLargeDesktop());
    expect(result.current).toBe(false);
  });

  it('returns true when width equals lg breakpoint (1024px)', () => {
    setWidth(1024);
    const { result } = renderHook(() => useIsLargeDesktop());
    expect(result.current).toBe(true);
  });

  it('returns true when width is above lg breakpoint (1280px)', () => {
    setWidth(1280);
    const { result } = renderHook(() => useIsLargeDesktop());
    expect(result.current).toBe(true);
  });

  it('updates on resize', () => {
    setWidth(800);
    const { result } = renderHook(() => useIsLargeDesktop());
    expect(result.current).toBe(false);

    act(() => {
      setWidth(1280);
      window.dispatchEvent(new Event('resize'));
    });

    expect(result.current).toBe(true);
  });

  it('removes listener on unmount', () => {
    vi.spyOn(window, 'removeEventListener');
    setWidth(1024);
    const { unmount } = renderHook(() => useIsLargeDesktop());
    unmount();
    expect(window.removeEventListener).toHaveBeenCalledWith('resize', expect.any(Function));
  });
});

describe('useIsMediumDesktop', () => {
  const originalInnerWidth = window.innerWidth;

  afterEach(() => setWidth(originalInnerWidth));

  it('returns false when width is below sm breakpoint', () => {
    setWidth(375);
    const { result } = renderHook(() => useIsMediumDesktop());
    expect(result.current).toBe(false);
  });

  it('returns true when width equals sm breakpoint (640px)', () => {
    setWidth(640);
    const { result } = renderHook(() => useIsMediumDesktop());
    expect(result.current).toBe(true);
  });

  it('returns true when width is between sm and lg (800px)', () => {
    setWidth(800);
    const { result } = renderHook(() => useIsMediumDesktop());
    expect(result.current).toBe(true);
  });

  it('returns false when width equals lg breakpoint (1024px)', () => {
    setWidth(1024);
    const { result } = renderHook(() => useIsMediumDesktop());
    expect(result.current).toBe(false);
  });

  it('updates on resize', () => {
    setWidth(800);
    const { result } = renderHook(() => useIsMediumDesktop());
    expect(result.current).toBe(true);

    act(() => {
      setWidth(1280);
      window.dispatchEvent(new Event('resize'));
    });

    expect(result.current).toBe(false);
  });
});
