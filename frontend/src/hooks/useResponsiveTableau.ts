import { useMemo } from 'react';
import { CARD_DIMENSIONS, useCardDimensions, useWindowWidth } from './useCardDimensions';

/** Responsive card dimensions for a multi-column tableau layout. */
export interface ResponsiveTableauDimensions {
  /** Card width in pixels. */
  cw: number;
  /** Card height in pixels (1.5 × cw). */
  ch: number;
  /** Vertical step (in px) between consecutive stacked cards. */
  co: number;
  /** Horizontal fan offset (in px) used by waste piles next to the tableau. */
  wasteFan: number;
}

/** Minimum visual card width in px. Anything smaller is unreadable on a 375 px portrait phone. */
const MIN_CARD_WIDTH = 24;
/**
 * Vertical-overlap ratio for stacked cards on mobile.
 *
 * Why: with 10-column tableaux on a 375 px viewport each column is ~32 px wide, well below the
 * WCAG 2.5.5 AAA 44×44 px tap target. We can't widen the card without horizontal scroll, so we
 * widen the *vertical* tap strip instead by exposing more of each face-up card. 0.58 keeps the
 * top of the next card visible while giving every card a tap-strip of ~0.42 × ch ≈ 20 px (vs 15
 * at the previous 0.48 ratio) on a 32-px-wide card. See issue #1648.
 */
const MOBILE_VERTICAL_OVERLAP_RATIO = 0.58;
/** Default fan offset in px used when the waste pile is rendered alongside the tableau. */
const DESKTOP_WASTE_FAN = 15;

/**
 * Hook that returns responsive card dimensions for an N-column tableau.
 *
 * On mobile viewports (< 640 px) the card width is computed from the available viewport so all
 * `numCols` columns fit without horizontal scroll. On desktop / large-desktop the standard
 * preset from {@link CARD_DIMENSIONS} is returned unchanged.
 */
export function useResponsiveTableau(numCols: number): ResponsiveTableauDimensions {
  const { cardHeight, cardOverlap, cardWidth, isMobile } = useCardDimensions();
  const windowWidth = useWindowWidth();

  return useMemo<ResponsiveTableauDimensions>(() => {
    if (!isMobile) {
      return { cw: cardWidth, ch: cardHeight, co: cardOverlap, wasteFan: DESKTOP_WASTE_FAN };
    }
    const padX = 16;
    const gapPx = 4;
    const availableWidth = windowWidth - padX - (numCols - 1) * gapPx;
    const colW = Math.floor(availableWidth / numCols);
    const cw = Math.min(Math.max(colW, MIN_CARD_WIDTH), cardWidth);
    const ch = Math.round(cw * 1.5);
    const co = Math.round(cw * MOBILE_VERTICAL_OVERLAP_RATIO);
    const wasteFan = Math.round(cw * 0.3);
    return { cw, ch, co, wasteFan };
  }, [isMobile, windowWidth, numCols, cardWidth, cardHeight, cardOverlap]);
}
