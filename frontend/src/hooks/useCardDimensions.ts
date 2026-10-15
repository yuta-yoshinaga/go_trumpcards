import { useEffect, useRef, useState } from 'react';
import { DESKTOP_COMPACT_CARD_WIDTH } from '../styles/gameStyles';

/** Breakpoint between mobile and desktop layouts (px). */
export const SM_BREAKPOINT = 640;
/** Breakpoint between desktop and large-desktop layouts (px). */
export const LG_BREAKPOINT = 1024;
/** Boundary where large-desktop cards return to their full size; 1024–1279px uses compact cards. */
export const XL_BREAKPOINT = 1280;

/** Breakpoint category names. */
export type Breakpoint = 'mobile' | 'desktop' | 'largeDesktop';

/** Card dimension presets for mobile, desktop, and large-desktop viewports. */
export const CARD_DIMENSIONS = {
  mobile: {
    cardHeight: 60,
    cardOverlap: 20,
    cardWidth: 40,
    cpuCardWidth: 34,
    footerCardWidth: 36,
    solitaireMinColWidth: 52,
  },
  desktop: {
    cardHeight: 84,
    cardOverlap: 22,
    cardWidth: 60,
    cpuCardWidth: 50,
    footerCardWidth: 54,
    solitaireMinColWidth: 0,
  },
  largeDesktop: {
    cardHeight: 150,
    cardOverlap: 34,
    cardWidth: 100,
    cpuCardWidth: 82,
    footerCardWidth: 90,
    solitaireMinColWidth: 0,
  },
} as const;

/**
 * Compact dimensions address the roughly 784px board between 1024–1279px, where full-size cards wrap.
 * At 1024px, reducing cards to 80px allowed 48 of the 57 affected visibility checks to pass.
 * `cardWidth` must stay aligned with `DESKTOP_COMPACT_CARD_WIDTH` in gameStyles because seven pages use that constant to size large-desktop cards.
 */
export const COMPACT_LARGE_DESKTOP_DIMENSIONS = {
  cardHeight: 120,
  cardOverlap: 28,
  cardWidth: DESKTOP_COMPACT_CARD_WIDTH,
  cpuCardWidth: 66,
  footerCardWidth: 72,
  solitaireMinColWidth: 0,
} as const;

/** Returns whether a viewport uses compact large-desktop cards. */
function isCompactLargeDesktopWidth(width: number): boolean {
  return width >= LG_BREAKPOINT && width < XL_BREAKPOINT;
}

/** Derive breakpoint category from a pixel width. */
function widthToBreakpoint(width: number): Breakpoint {
  if (width >= LG_BREAKPOINT) return 'largeDesktop';
  if (width >= SM_BREAKPOINT) return 'desktop';
  return 'mobile';
}

/** Shared hook that tracks viewport width with resize listener. */
export function useWindowWidth(): number {
  const [width, setWidth] = useState(() => window.innerWidth);

  useEffect(() => {
    const handleResize = () => setWidth(window.innerWidth);
    window.addEventListener('resize', handleResize);
    return () => window.removeEventListener('resize', handleResize);
  }, []);

  return width;
}

/** Shared hook that tracks viewport height with resize listener. */
export function useWindowHeight(): number {
  const [height, setHeight] = useState(() => window.innerHeight);

  useEffect(() => {
    const handleResize = () => setHeight(window.innerHeight);
    window.addEventListener('resize', handleResize);
    return () => window.removeEventListener('resize', handleResize);
  }, []);

  return height;
}

/**
 * Hook that tracks viewport breakpoint category, only triggering re-renders
 * when the viewport crosses SM_BREAKPOINT (640px) or LG_BREAKPOINT (1024px).
 * Within-breakpoint resizes are ignored to avoid wasted renders.
 */
export function useBreakpoint(): Breakpoint {
  const [bp, setBp] = useState<Breakpoint>(() => widthToBreakpoint(window.innerWidth));
  const bpRef = useRef(bp);

  useEffect(() => {
    const handleResize = () => {
      const next = widthToBreakpoint(window.innerWidth);
      if (next !== bpRef.current) {
        bpRef.current = next;
        setBp(next);
      }
    };
    window.addEventListener('resize', handleResize);
    return () => window.removeEventListener('resize', handleResize);
  }, []);

  return bp;
}

/** Returns true from 1024px up to, but not including, the full-size card breakpoint. */
export function useIsCompactLargeDesktop(): boolean {
  const [isCompact, setIsCompact] = useState(() => isCompactLargeDesktopWidth(window.innerWidth));
  const compactRef = useRef(isCompact);

  useEffect(() => {
    const handleResize = () => {
      const width = window.innerWidth;
      const next = isCompactLargeDesktopWidth(width);
      if (next !== compactRef.current) {
        compactRef.current = next;
        setIsCompact(next);
      }
    };
    window.addEventListener('resize', handleResize);
    return () => window.removeEventListener('resize', handleResize);
  }, []);

  return isCompact;
}

/** Hook that returns responsive card dimensions based on viewport breakpoint. */
export function useCardDimensions() {
  const bp = useBreakpoint();
  const isCompactLargeDesktop = useIsCompactLargeDesktop();
  const isMobile = bp === 'mobile';
  const dimensions = isCompactLargeDesktop ? COMPACT_LARGE_DESKTOP_DIMENSIONS : CARD_DIMENSIONS[bp];
  return { ...dimensions, isMobile };
}

/** Hook that returns true when viewport width is below the sm breakpoint (mobile). */
export function useIsMobile(): boolean {
  return useBreakpoint() === 'mobile';
}

/** Hook that returns true when viewport width is at or above the large-desktop breakpoint. */
export function useIsLargeDesktop(): boolean {
  return useBreakpoint() === 'largeDesktop';
}

/**
 * Provides fixed card dimensions for pages that use smaller-than-shared cards on large desktop:
 * 80×120 from 1280px, and the desktop preset (60×84) at 1024–1279px, where the board is too short.
 * The returned size is meaningful only when `useBreakpoint() === 'largeDesktop'`; below 1024px it
 * returns 80×120, which callers must not use.
 */
export function useDenseLargeDesktopCardSize(): { width: number; height: number } {
  const isCompactLargeDesktop = useIsCompactLargeDesktop();
  return isCompactLargeDesktop
    ? { width: CARD_DIMENSIONS.desktop.cardWidth, height: CARD_DIMENSIONS.desktop.cardHeight }
    : { width: DESKTOP_COMPACT_CARD_WIDTH, height: Math.round(DESKTOP_COMPACT_CARD_WIDTH * 1.5) };
}
