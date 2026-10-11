/** Minimum number of cards before splitting into 2 rows. */
export const TWO_ROW_THRESHOLD = 4;
/** Horizontal padding (px) on the hand container (px-4 = 16px each side). */
export const CONTAINER_PADDING = 32;
/** Base width of each card button including border (cardWidth + 6px border). */
export const BUTTON_EXTRA = 6;
/** Default gap (px) between cards when they fit without overlap. */
export const DEFAULT_CARD_GAP = 2;
/** Maximum overlap ratio — each card stays at least 30% visible. */
export const MAX_OVERLAP_RATIO = 0.7;
/** Minimum visible width (px) per card to meet WCAG 2.5.8 tap-target guidelines. */
export const MIN_CARD_EXPOSURE_PX = 44;
/** Card image height relative to its width (the 250x375 assets, 2:3). */
export const CARD_HEIGHT_RATIO = 1.5;
/** Vertical border around a card button, added to the image height. */
export const CARD_BORDER_Y = 2;
/** Height of an upper row left visible when rows overlap, including its 44px tap target. */
export const STACKED_ROW_VISIBLE_PX = 52;

/** Result of overlap computation including scroll-fallback flag. */
export interface OverlapResult {
  /** Negative margin-left in px (or positive gap). */
  overlap: number;
  /** True when cards cannot fit with minimum tap-target exposure. */
  useScroll: boolean;
}

/** Compute negative margin-left overlap so cards fit within the viewport. */
export function computeOverlap(cardCount: number, buttonWidth: number, viewportWidth: number): OverlapResult {
  if (cardCount <= 1) return { overlap: 0, useScroll: false };
  const availableWidth = viewportWidth - CONTAINER_PADDING;
  const totalNeeded = cardCount * buttonWidth;
  if (totalNeeded <= availableWidth) return { overlap: DEFAULT_CARD_GAP, useScroll: false };
  const rawOverlap = -((totalNeeded - availableWidth) / (cardCount - 1));
  const maxNegative = -buttonWidth * MAX_OVERLAP_RATIO;
  const minExposureOverlap = -(buttonWidth - MIN_CARD_EXPOSURE_PX);
  if (rawOverlap < minExposureOverlap) {
    return { overlap: DEFAULT_CARD_GAP, useScroll: true };
  }
  return { overlap: Math.max(rawOverlap, maxNegative), useScroll: false };
}
