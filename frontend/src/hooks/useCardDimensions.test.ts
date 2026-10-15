import { act, renderHook } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import {
  CARD_DIMENSIONS,
  COMPACT_LARGE_DESKTOP_DIMENSIONS,
  useBreakpoint,
  useCardDimensions,
  useIsCompactLargeDesktop,
  useIsLargeDesktop,
  useIsMobile,
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

describe('useBreakpoint', () => {
  const originalInnerWidth = window.innerWidth;

  afterEach(() => setWidth(originalInnerWidth));

  it('returns mobile below 640px', () => {
    setWidth(375);
    const { result } = renderHook(() => useBreakpoint());
    expect(result.current).toBe('mobile');
  });

  it('returns desktop between 640-1024px', () => {
    setWidth(800);
    const { result } = renderHook(() => useBreakpoint());
    expect(result.current).toBe('desktop');
  });

  it('returns largeDesktop above 1024px', () => {
    setWidth(1280);
    const { result } = renderHook(() => useBreakpoint());
    expect(result.current).toBe('largeDesktop');
  });

  it('does not re-render on within-breakpoint resize', () => {
    setWidth(400);
    let renderCount = 0;
    const { result } = renderHook(() => {
      renderCount++;
      return useBreakpoint();
    });
    expect(result.current).toBe('mobile');
    const initialRenderCount = renderCount;

    act(() => {
      setWidth(500);
      window.dispatchEvent(new Event('resize'));
    });

    expect(result.current).toBe('mobile');
    expect(renderCount).toBe(initialRenderCount);
  });

  it('re-renders on cross-breakpoint resize', () => {
    setWidth(600);
    const { result } = renderHook(() => useBreakpoint());
    expect(result.current).toBe('mobile');

    act(() => {
      setWidth(700);
      window.dispatchEvent(new Event('resize'));
    });

    expect(result.current).toBe('desktop');
  });

  it('removes listener on unmount', () => {
    vi.spyOn(window, 'removeEventListener');
    setWidth(800);
    const { unmount } = renderHook(() => useBreakpoint());
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
    expect(result.current).toEqual({ ...COMPACT_LARGE_DESKTOP_DIMENSIONS, isMobile: false });
    expect(result.current.cardWidth).toBe(80);
  });

  it('returns largeDesktop dimensions when width is above lg breakpoint (1280px)', () => {
    setWidth(1280);
    const { result } = renderHook(() => useCardDimensions());
    expect(result.current).toEqual({ ...CARD_DIMENSIONS.largeDesktop, isMobile: false });
  });

  it.each([
    [1023, 60],
    [1024, 80],
    [1279, 80],
    [1280, 100],
  ])('uses expected card width at %i px', (width, cardWidth) => {
    setWidth(width);
    const { result } = renderHook(() => useCardDimensions());
    expect(result.current.cardWidth).toBe(cardWidth);
  });

  it.each([
    [1279, 1280, 100],
    [1280, 1279, 80],
  ])('updates card size crossing %i to %i px', (from, to, cardWidth) => {
    setWidth(from);
    const { result } = renderHook(() => useCardDimensions());
    act(() => {
      setWidth(to);
      window.dispatchEvent(new Event('resize'));
    });
    expect(result.current.cardWidth).toBe(cardWidth);
  });

  it('does not re-render for resizes within compact large desktop', () => {
    setWidth(1100);
    let renderCount = 0;
    const { result } = renderHook(() => {
      renderCount++;
      return useIsCompactLargeDesktop();
    });
    const initialRenderCount = renderCount;
    act(() => {
      setWidth(1200);
      window.dispatchEvent(new Event('resize'));
    });
    expect(result.current).toBe(true);
    expect(renderCount).toBe(initialRenderCount);
  });

  it('updates dimensions when window is resized to mobile', () => {
    setWidth(1024);
    const { result } = renderHook(() => useCardDimensions());
    expect(result.current).toEqual({ ...COMPACT_LARGE_DESKTOP_DIMENSIONS, isMobile: false });

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

describe('useIsMobile', () => {
  const originalInnerWidth = window.innerWidth;

  afterEach(() => setWidth(originalInnerWidth));

  it('returns true when width is below sm breakpoint (375px)', () => {
    setWidth(375);
    const { result } = renderHook(() => useIsMobile());
    expect(result.current).toBe(true);
  });

  it('returns true when width is just below sm breakpoint (639px)', () => {
    setWidth(639);
    const { result } = renderHook(() => useIsMobile());
    expect(result.current).toBe(true);
  });

  it('returns false when width equals sm breakpoint (640px)', () => {
    setWidth(640);
    const { result } = renderHook(() => useIsMobile());
    expect(result.current).toBe(false);
  });

  it('returns false when width is above sm breakpoint (1024px)', () => {
    setWidth(1024);
    const { result } = renderHook(() => useIsMobile());
    expect(result.current).toBe(false);
  });

  it('updates on resize', () => {
    setWidth(375);
    const { result } = renderHook(() => useIsMobile());
    expect(result.current).toBe(true);

    act(() => {
      setWidth(800);
      window.dispatchEvent(new Event('resize'));
    });

    expect(result.current).toBe(false);
  });
});
