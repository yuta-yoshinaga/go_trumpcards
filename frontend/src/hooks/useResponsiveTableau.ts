import { useMemo } from 'react';
import { useCardDimensions, useWindowHeight, useWindowWidth } from './useCardDimensions';

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

/** Caller-supplied mobile layout dimensions matching the page's actual CSS. */
export interface ResponsiveTableauConfig {
  /** Total horizontal padding (left + right) of the tableau's scroll container, in px. Defaults to 16 (Tailwind `px-2`). */
  padX?: number;
  /** Gap between adjacent tableau columns, in px. Defaults to 4 (Tailwind `gap-1`). */
  gapPx?: number;
  /**
   * Current length of the longest tableau column. When provided on a mobile viewport, the hook
   * shrinks the per-card vertical overlap so the tallest column fits without scrolling.
   * Omit for layouts where columns can never grow tall enough to overflow (#1861).
   */
  maxColCards?: number;
  /**
   * Vertical pixels reserved for non-tableau chrome (header, stock row, footer, settings, message
   * box). Subtracted from `window.innerHeight` to derive the tableau's vertical budget. Defaults
   * to 300 px which matches the standard solo-game shell on a 667 px viewport.
   */
  reservedHeightPx?: number;
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
/**
 * Hard floor for the per-card vertical step when the dynamic Y-offset compressor (#1861) kicks
 * in. Below ~8 px the next card's top edge disappears entirely, leaving no peek/tap strip.
 */
const MIN_VERTICAL_OVERLAP = 8;
/** Default fan offset in px used when the waste pile is rendered alongside the tableau. */
const DESKTOP_WASTE_FAN = 15;
const DEFAULT_PAD_X = 16;
const DEFAULT_GAP_PX = 4;
/**
 * Pixels of vertical viewport consumed by the standard game-page chrome (PhaseIndicator,
 * stock/foundation row, hint+message+actionlog, GameFooter). Calibrated against the solo-game
 * shell at 375×667 — callers can override via `reservedHeightPx` when their layout differs.
 */
const DEFAULT_RESERVED_HEIGHT_PX = 300;

/**
 * Hook that returns responsive card dimensions for an N-column tableau.
 *
 * On mobile viewports (< 640 px) the card width is computed from the available viewport so all
 * `numCols` columns fit without horizontal scroll. On desktop / large-desktop the standard
 * preset from `CARD_DIMENSIONS` is returned unchanged.
 *
 * The optional `config` lets callers override the assumed scroll-container padding and inter-column
 * gap so the calculation matches the page's actual Tailwind classes — without it the math implicitly
 * assumes `px-2` / `gap-1`. When `maxColCards` is supplied, the per-card vertical overlap is also
 * compressed on mobile so the tallest column fits without vertical scrolling (#1861).
 */
export function useResponsiveTableau(
  numCols: number,
  config: ResponsiveTableauConfig = {},
): ResponsiveTableauDimensions {
  const { cardHeight, cardOverlap, cardWidth, isMobile } = useCardDimensions();
  const windowWidth = useWindowWidth();
  const windowHeight = useWindowHeight();
  const padX = config.padX ?? DEFAULT_PAD_X;
  const gapPx = config.gapPx ?? DEFAULT_GAP_PX;
  const maxColCards = config.maxColCards;
  const reservedHeightPx = config.reservedHeightPx ?? DEFAULT_RESERVED_HEIGHT_PX;

  return useMemo<ResponsiveTableauDimensions>(() => {
    if (!isMobile) {
      return { cw: cardWidth, ch: cardHeight, co: cardOverlap, wasteFan: DESKTOP_WASTE_FAN };
    }
    const availableWidth = windowWidth - padX - (numCols - 1) * gapPx;
    const colW = Math.floor(availableWidth / numCols);
    const cw = Math.min(Math.max(colW, MIN_CARD_WIDTH), cardWidth);
    const ch = Math.round(cw * 1.5);
    const naturalCo = Math.round(cw * MOBILE_VERTICAL_OVERLAP_RATIO);
    let co = naturalCo;
    if (maxColCards && maxColCards > 1) {
      const availableHeight = windowHeight - reservedHeightPx;
      const fittingCo = Math.floor((availableHeight - ch) / (maxColCards - 1));
      co = Math.max(MIN_VERTICAL_OVERLAP, Math.min(naturalCo, fittingCo));
    }
    const wasteFan = Math.round(cw * 0.3);
    return { cw, ch, co, wasteFan };
  }, [
    isMobile,
    windowWidth,
    windowHeight,
    numCols,
    padX,
    gapPx,
    maxColCards,
    reservedHeightPx,
    cardWidth,
    cardHeight,
    cardOverlap,
  ]);
}
