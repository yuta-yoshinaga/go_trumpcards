import type { Card } from '../types/card';

/**
 * Whether two vertically adjacent tableau cards continue a same-suit descending
 * run. `upper` is the card nearer the top of the column (lower array index),
 * `lower` the one just below it. Mirrors the Go domain rule in
 * `CurdsAndWhey.curdsAndWheyIsRun`: same design and the upper value is exactly one
 * greater than the lower value.
 */
export function isRunLink(upper: Card, lower: Card): boolean {
  return upper.design === lower.design && upper.value === lower.value + 1;
}

/**
 * Index of the topmost card in a column that begins a valid movable run — a
 * same-suit descending sequence running unbroken to the bottom of the column.
 * Cards at an index `>=` the returned value can be grabbed as a move source;
 * cards above it cannot. Returns 0 for an empty column (no grabbable cards).
 */
export function movableFromIndex(column: Card[]): number {
  if (column.length === 0) return 0;
  let idx = column.length - 1;
  while (idx > 0 && isRunLink(column[idx - 1], column[idx])) {
    idx--;
  }
  return idx;
}

/**
 * Whether the card at `idx` in `column` can be grabbed as the head of a movable
 * run (i.e. `column[idx..]` is a same-suit descending run).
 */
export function isGrabbable(column: Card[], idx: number): boolean {
  return column.length > 0 && idx >= movableFromIndex(column);
}
