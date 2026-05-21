import { act, renderHook } from '@testing-library/react';
import { afterEach, describe, expect, it } from 'vitest';
import { CARD_DIMENSIONS } from './useCardDimensions';
import { useResponsiveTableau } from './useResponsiveTableau';

const setWidth = (w: number) =>
  Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: w });
const setHeight = (h: number) =>
  Object.defineProperty(window, 'innerHeight', { writable: true, configurable: true, value: h });

describe('useResponsiveTableau', () => {
  const originalInnerWidth = window.innerWidth;
  const originalInnerHeight = window.innerHeight;

  afterEach(() => {
    setWidth(originalInnerWidth);
    setHeight(originalInnerHeight);
  });

  it('returns desktop preset when viewport is at desktop breakpoint', () => {
    setWidth(800);
    const { result } = renderHook(() => useResponsiveTableau(10));
    expect(result.current.cw).toBe(CARD_DIMENSIONS.desktop.cardWidth);
    expect(result.current.ch).toBe(CARD_DIMENSIONS.desktop.cardHeight);
    expect(result.current.co).toBe(CARD_DIMENSIONS.desktop.cardOverlap);
  });

  it('returns large-desktop preset above 1024px', () => {
    setWidth(1280);
    const { result } = renderHook(() => useResponsiveTableau(10));
    expect(result.current.cw).toBe(CARD_DIMENSIONS.largeDesktop.cardWidth);
    expect(result.current.ch).toBe(CARD_DIMENSIONS.largeDesktop.cardHeight);
    expect(result.current.co).toBe(CARD_DIMENSIONS.largeDesktop.cardOverlap);
  });

  it('shrinks card width to fit 10 columns on a 375px mobile viewport (default px-2/gap-1)', () => {
    setWidth(375);
    const { result } = renderHook(() => useResponsiveTableau(10));
    // floor((375 - 16 padX - 9 * 4 gap) / 10) = floor(32.3) = 32, clamped to [24, mobile.cardWidth].
    expect(result.current.cw).toBe(32);
    expect(result.current.cw).toBeLessThanOrEqual(CARD_DIMENSIONS.mobile.cardWidth);
    expect(result.current.ch).toBe(Math.round(result.current.cw * 1.5));
  });

  it('honors custom padX/gapPx so SpiderPage (px-4 / gap-0.5) gets accurate sizing', () => {
    setWidth(375);
    const { result } = renderHook(() => useResponsiveTableau(10, { padX: 32, gapPx: 2 }));
    // floor((375 - 32 padX - 9 * 2 gap) / 10) = floor(32.5) = 32 — matches Spider's actual layout.
    expect(result.current.cw).toBe(32);
  });

  it('diverges from defaults on a narrower viewport when Spider-style padding is used', () => {
    setWidth(320);
    const ftLike = renderHook(() => useResponsiveTableau(10)).result.current;
    const spiderLike = renderHook(() => useResponsiveTableau(10, { padX: 32, gapPx: 2 })).result.current;
    // Different layouts → different card widths once the viewport is tight enough.
    expect(spiderLike.cw).not.toBe(ftLike.cw);
  });

  it('exposes a larger tap strip per stacked card on mobile (co >= 0.55 * cw)', () => {
    setWidth(375);
    const { result } = renderHook(() => useResponsiveTableau(10));
    const ratio = result.current.co / result.current.cw;
    expect(ratio).toBeGreaterThanOrEqual(0.55);
  });

  it('clamps to a minimum card width even on very narrow viewports', () => {
    setWidth(280);
    const { result } = renderHook(() => useResponsiveTableau(10));
    expect(result.current.cw).toBeGreaterThanOrEqual(24);
  });

  it('uses fewer columns to keep cards readable when fewer columns are passed', () => {
    setWidth(375);
    const tenCol = renderHook(() => useResponsiveTableau(10)).result.current;
    const eightCol = renderHook(() => useResponsiveTableau(8)).result.current;
    expect(eightCol.cw).toBeGreaterThan(tenCol.cw);
  });

  it('updates on resize', () => {
    setWidth(800);
    const { result } = renderHook(() => useResponsiveTableau(10));
    expect(result.current.cw).toBe(CARD_DIMENSIONS.desktop.cardWidth);
    act(() => {
      setWidth(375);
      window.dispatchEvent(new Event('resize'));
    });
    expect(result.current.cw).toBeLessThanOrEqual(CARD_DIMENSIONS.mobile.cardWidth);
  });

  it('returns wasteFan derived from card width on mobile', () => {
    setWidth(375);
    const { result } = renderHook(() => useResponsiveTableau(10));
    expect(result.current.wasteFan).toBe(Math.round(result.current.cw * 0.3));
  });

  it('returns the desktop wasteFan default on desktop', () => {
    setWidth(800);
    const { result } = renderHook(() => useResponsiveTableau(10));
    expect(result.current.wasteFan).toBe(15);
  });

  describe('dynamic Y-offset (maxColCards)', () => {
    it('leaves co at the natural ratio when tallest column fits in the viewport', () => {
      setWidth(375);
      setHeight(667);
      const natural = renderHook(() => useResponsiveTableau(10)).result.current;
      const fitting = renderHook(() => useResponsiveTableau(10, { maxColCards: 4 })).result.current;
      // 4 short cards × natural overlap (~18px) + ch (~48px) ≈ 102 px — easily fits.
      expect(fitting.co).toBe(natural.co);
    });

    it('compresses co so the tallest column fits the available vertical space', () => {
      setWidth(375);
      setHeight(667);
      // 25 cards stacked at natural overlap (~18 px) would consume 25*18+ch = ~498 px,
      // overflowing the ~367 px tableau budget (667 - 300 reserved chrome). Hook must shrink co.
      const { result } = renderHook(() => useResponsiveTableau(10, { maxColCards: 25 }));
      const ch = result.current.ch;
      const co = result.current.co;
      const tallestCol = (25 - 1) * co + ch;
      expect(tallestCol).toBeLessThanOrEqual(667 - 300);
    });

    it('does not trigger compression when maxColCards is 1 (single-card column)', () => {
      setWidth(375);
      setHeight(667);
      const natural = renderHook(() => useResponsiveTableau(10)).result.current;
      const single = renderHook(() => useResponsiveTableau(10, { maxColCards: 1 })).result.current;
      expect(single.co).toBe(natural.co);
    });

    it('never compresses below the minimum overlap so cards always have a tap strip', () => {
      setWidth(375);
      setHeight(300); // pathologically short viewport
      const { result } = renderHook(() => useResponsiveTableau(10, { maxColCards: 30 }));
      expect(result.current.co).toBeGreaterThanOrEqual(8);
    });

    it('respects a caller-supplied reservedHeightPx', () => {
      setWidth(375);
      setHeight(667);
      // Looser reservation (200 px) → more room → larger co.
      const tight = renderHook(() => useResponsiveTableau(10, { maxColCards: 20, reservedHeightPx: 400 })).result
        .current;
      const loose = renderHook(() => useResponsiveTableau(10, { maxColCards: 20, reservedHeightPx: 200 })).result
        .current;
      expect(loose.co).toBeGreaterThan(tight.co);
    });

    it('does not affect desktop dimensions', () => {
      setWidth(1280);
      setHeight(800);
      const { result } = renderHook(() => useResponsiveTableau(10, { maxColCards: 30 }));
      expect(result.current.co).toBe(CARD_DIMENSIONS.largeDesktop.cardOverlap);
    });
  });
});
