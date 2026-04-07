import { useEffect, useRef, useState } from 'react';

/** Breakpoint between mobile and desktop layouts (px). */
export const SM_BREAKPOINT = 640;
/** Breakpoint between desktop and large-desktop layouts (px). */
export const LG_BREAKPOINT = 1024;

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

/** Hook that returns responsive card dimensions based on viewport breakpoint. */
export function useCardDimensions() {
  const bp = useBreakpoint();
  const isMobile = bp === 'mobile';
  return { ...CARD_DIMENSIONS[bp], isMobile };
}

/** Hook that returns true when viewport width is below the sm breakpoint (mobile). */
export function useIsMobile(): boolean {
  return useBreakpoint() === 'mobile';
}

/** Hook that returns true when viewport width is at or above the large-desktop breakpoint. */
export function useIsLargeDesktop(): boolean {
  return useBreakpoint() === 'largeDesktop';
}

/** Hook that returns true when viewport is between sm and lg breakpoints (tablet/small desktop). */
export function useIsMediumDesktop(): boolean {
  return useBreakpoint() === 'desktop';
}
