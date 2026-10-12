/**
 * Minimum visible width of an overlapped card on desktop. A pointer target, not a touch target:
 * 28px clears WCAG 2.5.8's 24px minimum while letting a 27-card hand fit one 1008px row; touch
 * layouts keep the 44px MIN_CARD_EXPOSURE_PX instead. Same value as PlayerHandSection's desktop overlap.
 */
export const DESKTOP_MIN_CARD_EXPOSURE_PX = 28;

/** Return the smallest balanced row count whose widest row fits at the requested exposure. */
export function fanRows(count: number, itemWidth: number, availableWidth: number, minExposure: number): number {
  if (count <= 1 || availableWidth <= 0) return 1;
  for (let rows = 1; rows < count; rows += 1) {
    const length = Math.ceil(count / rows);
    if (length * itemWidth - (length - 1) * (itemWidth - minExposure) <= availableWidth) return rows;
  }
  return count;
}

/** Return the per-card margin that packs a row to its available width without exceeding minimum exposure. */
export function fanOverlap(
  rowLength: number,
  itemWidth: number,
  availableWidth: number,
  minExposure: number,
  gap: number,
): number {
  if (rowLength <= 1) return 0;
  const naturalWidth = rowLength * itemWidth + (rowLength - 1) * gap;
  if (naturalWidth <= availableWidth) return gap;
  return Math.max(-(itemWidth - minExposure), (availableWidth - rowLength * itemWidth) / (rowLength - 1));
}
