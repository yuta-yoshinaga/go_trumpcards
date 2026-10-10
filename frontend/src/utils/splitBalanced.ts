/** Split items into the requested number of contiguous rows as evenly as possible. */
export function splitBalanced<T>(items: readonly T[], rowCount: number): Array<{ items: T[]; start: number }> {
  const rows = Math.max(1, Math.min(rowCount, items.length || 1));
  const baseSize = Math.floor(items.length / rows);
  const largerRows = items.length % rows;
  return Array.from({ length: rows }, (_, rowIdx) => {
    const start = rowIdx * baseSize + Math.min(rowIdx, largerRows);
    const length = baseSize + (rowIdx < largerRows ? 1 : 0);
    return { items: items.slice(start, start + length), start };
  });
}
