import type { Card } from '../types/card';

/** Anchor rank for the leftmost column of a Gaps row (must be a 2). */
export const GAPS_ANCHOR_RANK = 2;

/**
 * Compute the locked-prefix length of each row on a Gaps board.
 * A card at `grid[r][c]` is locked when it forms an unbroken run starting
 * with a 2 in column 0 and ascending by one of the same suit thereafter.
 *
 * Mirrors `lockedPrefixLengths` in `internal/domain/Gaps.go`. Locked cards
 * survive a redeal, so the UI uses this to mark them visually.
 *
 * Returns an array whose length matches the input — empty/short rows are
 * handled by returning the count of locked cards (0 when the leftmost card
 * is missing or not a 2).
 */
export function gapsLockedPrefixLengths(grid: readonly (readonly (Card | null)[])[]): number[] {
  return grid.map((row) => lockedPrefixLengthForRow(row));
}

function lockedPrefixLengthForRow(row: readonly (Card | null)[]): number {
  let locked = 0;
  for (let c = 0; c < row.length; c++) {
    const cur = row[c];
    if (cur == null) break;
    if (c === 0) {
      if (cur.value !== GAPS_ANCHOR_RANK) break;
      locked++;
      continue;
    }
    const prev = row[c - 1];
    if (prev == null) break;
    if (cur.design !== prev.design || cur.value !== prev.value + 1) break;
    locked++;
  }
  return locked;
}
