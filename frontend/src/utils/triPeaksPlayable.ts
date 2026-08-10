import type { TriPeaksCard } from '../types/games/tripeaks';
import { isTriPeaksAdjacent } from './hints/tripeaksHint';

/** Key identifying a tableau cell, as `row-col`. */
export type TriPeaksCellKey = `${number}-${number}`;

/**
 * Exposed tableau cells that can be taken onto the current waste top.
 *
 * Sync: `TriPeaksCuiPresenter.Output`'s `playableCount`.
 *
 * **表示とリングは同じ集合を読む。**枚数を別に数え直すと、リングが付いた札の数と
 * ヘッダーの数字がずれる (#4783)。
 */
export function triPeaksPlayableCells(
  layout: TriPeaksCard[][],
  wasteTopValue: number | undefined,
): Set<TriPeaksCellKey> {
  const cells = new Set<TriPeaksCellKey>();
  if (wasteTopValue === undefined) return cells;

  layout.forEach((row, rowIdx) => {
    row.forEach((tc, colIdx) => {
      if (!tc?.card || tc.removed || !tc.exposed) return;
      if (isTriPeaksAdjacent(tc.card.value, wasteTopValue)) cells.add(`${rowIdx}-${colIdx}`);
    });
  });

  return cells;
}
