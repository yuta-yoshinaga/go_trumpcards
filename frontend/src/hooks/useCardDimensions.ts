import { useEffect, useState } from 'react';

const MOBILE_BREAKPOINT = 640;

/** Card dimension presets for mobile and desktop viewports. */
export const CARD_DIMENSIONS = {
  mobile: {
    cardHeight: 60,
    cardOverlap: 16,
    cardWidth: 40,
    cpuCardWidth: 34,
    footerCardWidth: 36,
    sevensCellSize: 20,
    sevensFontSize: '0.6em',
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
} as const;

/** Hook that returns responsive card dimensions based on viewport width. */
export function useCardDimensions() {
  const [width, setWidth] = useState(() => window.innerWidth);

  useEffect(() => {
    const handleResize = () => setWidth(window.innerWidth);
    window.addEventListener('resize', handleResize);
    return () => window.removeEventListener('resize', handleResize);
  }, []);

  return width < MOBILE_BREAKPOINT ? CARD_DIMENSIONS.mobile : CARD_DIMENSIONS.desktop;
}
