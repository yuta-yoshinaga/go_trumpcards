import { useEffect, useState } from 'react';

const MOBILE_BREAKPOINT = 640;

export const CARD_DIMENSIONS = {
  mobile: { cardHeight: 60, cardOverlap: 16, cardWidth: 40, sevensCellSize: 20, sevensFontSize: '0.6em' },
  desktop: { cardHeight: 84, cardOverlap: 22, cardWidth: 60, sevensCellSize: 26, sevensFontSize: '0.75em' },
} as const;

export function useCardDimensions() {
  const [width, setWidth] = useState(() => window.innerWidth);

  useEffect(() => {
    const handleResize = () => setWidth(window.innerWidth);
    window.addEventListener('resize', handleResize);
    return () => window.removeEventListener('resize', handleResize);
  }, []);

  return width < MOBILE_BREAKPOINT ? CARD_DIMENSIONS.mobile : CARD_DIMENSIONS.desktop;
}
