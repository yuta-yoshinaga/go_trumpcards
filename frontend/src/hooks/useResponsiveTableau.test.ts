import { act, renderHook } from '@testing-library/react';
import { afterEach, describe, expect, it } from 'vitest';
import { CARD_DIMENSIONS } from './useCardDimensions';
import { useResponsiveTableau } from './useResponsiveTableau';

const setWidth = (w: number) =>
  Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: w });

describe('useResponsiveTableau', () => {
  const originalInnerWidth = window.innerWidth;

  afterEach(() => setWidth(originalInnerWidth));

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

  it('shrinks card width to fit 10 columns on a 375px mobile viewport', () => {
    setWidth(375);
    const { result } = renderHook(() => useResponsiveTableau(10));
    // Expected ~ floor((375 - 16 - 9*4) / 10) = 32, clamped to [28, 40].
    expect(result.current.cw).toBeGreaterThanOrEqual(28);
    expect(result.current.cw).toBeLessThanOrEqual(CARD_DIMENSIONS.mobile.cardWidth);
    expect(result.current.ch).toBe(Math.round(result.current.cw * 1.5));
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
});
