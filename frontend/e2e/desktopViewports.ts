/**
 * Run the board, hand, and action visibility guards at 1280×800 and 1024×800.
 * 1024px is `LG_BREAKPOINT`: the two-column layout starts and `useCardDimensions` reduces cards to
 * 80px (#11811). Specs avoid loading React modules, so 1024 is written here as a literal.
 */
export const PHONE_VIEWPORT = { width: 375, height: 667 } as const;
export const DESKTOP_VIEWPORT_LARGE = { width: 1280, height: 800 } as const;
export const DESKTOP_VIEWPORT_COMPACT = { width: 1024, height: 800 } as const;

export const DESKTOP_VIEWPORTS = [DESKTOP_VIEWPORT_LARGE, DESKTOP_VIEWPORT_COMPACT] as const;

/** Minimum visible horizontal card width, passed into page.evaluate callbacks. */
export const VISIBLE_PX_X = 20;

/** Minimum visible vertical card height, passed into page.evaluate callbacks. */
export const VISIBLE_PX_Y = 52;

/** Returns the largest allowed rendered width for a desktop board card. */
export function getDesktopCardWidthLimit(viewportWidth: number): number {
  // Keep these values aligned with CARD_DIMENSIONS.largeDesktop and
  // COMPACT_LARGE_DESKTOP_DIMENSIONS in src/hooks/useCardDimensions.ts.
  const baseCardWidth = viewportWidth >= 1024 && viewportWidth < 1280 ? 80 : 100;
  return baseCardWidth * 1.5;
}
