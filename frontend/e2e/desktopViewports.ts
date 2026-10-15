/**
 * Run the board, hand, and action visibility guards at 1280×800 and 1024×800.
 * 1024px is `LG_BREAKPOINT`: the two-column layout starts and `useCardDimensions` reduces cards to
 * 80px (#11811). Specs avoid loading React modules, so 1024 is written here as a literal.
 */
export const DESKTOP_VIEWPORTS = [
  { width: 1280, height: 800 },
  { width: 1024, height: 800 },
] as const;

/** Minimum visible horizontal card width, passed into page.evaluate callbacks. */
export const VISIBLE_PX_X = 20;

/** Minimum visible vertical card height, passed into page.evaluate callbacks. */
export const VISIBLE_PX_Y = 52;
