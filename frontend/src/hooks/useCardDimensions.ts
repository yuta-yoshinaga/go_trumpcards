import { useEffect, useState } from 'react';

/** Breakpoint between mobile and desktop layouts (px). */
export const SM_BREAKPOINT = 640;
/** Breakpoint between desktop and large-desktop layouts (px). */
export const LG_BREAKPOINT = 1024;

/** Card dimension presets for mobile, desktop, and large-desktop viewports. */
export const CARD_DIMENSIONS = {
  mobile: {
    cardHeight: 60,
    cardOverlap: 20,
    cardWidth: 40,
    cpuCardWidth: 34,
    footerCardWidth: 36,
    sevensCellSize: 24,
    sevensFontSize: '0.65em',
    solitaireMinColWidth: 48,
  },
  desktop: {
    cardHeight: 84,
    cardOverlap: 22,
    cardWidth: 60,
    cpuCardWidth: 50,
    footerCardWidth: 54,
    sevensCellSize: 26,
    sevensFontSize: '0.75em',
    solitaireMinColWidth: 0,
  },
  largeDesktop: {
    cardHeight: 120,
    cardOverlap: 28,
    cardWidth: 80,
    cpuCardWidth: 66,
    footerCardWidth: 72,
    sevensCellSize: 32,
    sevensFontSize: '0.85em',
    solitaireMinColWidth: 0,
  },
} as const;

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

/** Hook that returns responsive card dimensions based on viewport width. */
export function useCardDimensions() {
  const width = useWindowWidth();
  const isMobile = width < SM_BREAKPOINT;

  if (width >= LG_BREAKPOINT) return { ...CARD_DIMENSIONS.largeDesktop, isMobile };
  if (width >= SM_BREAKPOINT) return { ...CARD_DIMENSIONS.desktop, isMobile };
  return { ...CARD_DIMENSIONS.mobile, isMobile };
}

/** Hook that returns true when viewport width is below the sm breakpoint (mobile). */
export function useIsMobile(): boolean {
  return useWindowWidth() < SM_BREAKPOINT;
}

/** Hook that returns true when viewport width is at or above the large-desktop breakpoint. */
export function useIsLargeDesktop(): boolean {
  return useWindowWidth() >= LG_BREAKPOINT;
}

/** Hook that returns true when viewport is between sm and lg breakpoints (tablet/small desktop). */
export function useIsMediumDesktop(): boolean {
  const width = useWindowWidth();
  return width >= SM_BREAKPOINT && width < LG_BREAKPOINT;
}
