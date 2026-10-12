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
