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
  },
  desktop: {
    cardHeight: 84,
    cardOverlap: 22,
    cardWidth: 60,
    cpuCardWidth: 50,
    footerCardWidth: 54,
    sevensCellSize: 26,
    sevensFontSize: '0.75em',
  },
  largeDesktop: {
    cardHeight: 120,
    cardOverlap: 28,
    cardWidth: 80,
    cpuCardWidth: 66,
    footerCardWidth: 72,
    sevensCellSize: 32,
    sevensFontSize: '0.85em',
  },
} as const;

/** Hook that returns responsive card dimensions based on viewport width. */
export function useCardDimensions() {
  const [width, setWidth] = useState(() => window.innerWidth);

  useEffect(() => {
    const handleResize = () => setWidth(window.innerWidth);
    window.addEventListener('resize', handleResize);
    return () => window.removeEventListener('resize', handleResize);
  }, []);

  if (width >= LG_BREAKPOINT) return CARD_DIMENSIONS.largeDesktop;
  if (width >= SM_BREAKPOINT) return CARD_DIMENSIONS.desktop;
  return CARD_DIMENSIONS.mobile;
}

/** Hook that returns true when viewport width is at or above the large-desktop breakpoint. */
export function useIsLargeDesktop(): boolean {
  const [width, setWidth] = useState(() => window.innerWidth);

  useEffect(() => {
    const handleResize = () => setWidth(window.innerWidth);
    window.addEventListener('resize', handleResize);
    return () => window.removeEventListener('resize', handleResize);
  }, []);

  return width >= LG_BREAKPOINT;
}
