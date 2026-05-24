import type { Card, CardDesign } from '../types/card';

/** Outcome of computing what should go into an empty Gaps cell. */
export type GapsGhostHint =
  | { kind: 'needed'; design: CardDesign; value: number }
  | { kind: 'anySuit'; value: number }
  | { kind: 'blocked' };

/**
 * Compute the ghost-card hint for an empty Gaps cell.
 *
 * Rules:
 * - Column 0: any "2" of any suit may be placed → `anySuit` (value 2).
 * - Left neighbor is a King (13): no card may be placed → `blocked`.
 * - Left neighbor is a `null` (another gap): no determined hint → `null`.
 * - Otherwise: the cell needs (leftSuit, leftValue + 1) → `needed`.
 */
export function computeGapsGhostHint(row: (Card | null)[], col: number): GapsGhostHint | null {
  if (col === 0) return { kind: 'anySuit', value: 2 };
  const leftCell = row[col - 1];
  if (leftCell == null) return null;
  if (leftCell.value === 13) return { kind: 'blocked' };
  return { kind: 'needed', design: leftCell.design, value: leftCell.value + 1 };
}
