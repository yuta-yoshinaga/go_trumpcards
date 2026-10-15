/**
 * Widths used by the three desktop visibility guards (board, hand, and action): 1280×800 and
 * 1024×800. At `LG_BREAKPOINT`, the two-column layout begins and `useCardDimensions` reduces cards
 * to 80px (#11811). The 1024 value is repeated because specs do not load React modules to import
 * the hook's constant.
 */
export const DESKTOP_VIEWPORTS = [
  { width: 1280, height: 800 },
  { width: 1024, height: 800 },
] as const;

/** Minimum visible horizontal card width, passed into page.evaluate callbacks. */
export const VISIBLE_PX_X = 20;

/** Minimum visible vertical card height, passed into page.evaluate callbacks. */
export const VISIBLE_PX_Y = 52;
