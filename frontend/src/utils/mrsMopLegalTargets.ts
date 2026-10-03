import type { MrsMopTableauCard } from '../types/games/mrsMop';
import { spiderMovableRun } from './solitaireUtils';

/**
 * Computes legal destination columns for a selected Mrs. Mop tableau run.
 * Mirrors `MrsMop.LegalTargets` and `canPlaceOnTableau` in
 * `internal/domain/MrsMop.go:427-485`: the selected tail must be a movable
 * same-suit descending run, then its bottom card fits an empty column or a
 * card one rank higher; the source column is excluded.
 */
export function mrsMopLegalTargets(
  tableau: MrsMopTableauCard[][],
  fromCol: number | undefined,
  cardIndex: number | undefined,
): number[] {
  if (fromCol === undefined || cardIndex === undefined) return [];
  const source = tableau[fromCol];
  if (!source || spiderMovableRun(source, cardIndex).length === 0) return [];
  const bottom = source[cardIndex]?.card;
  if (!bottom) return [];
  return tableau.flatMap((column, col) => {
    if (col === fromCol) return [];
    const top = column[column.length - 1]?.card;
    return column.length === 0 || (top !== null && top !== undefined && bottom.value === top.value - 1) ? [col] : [];
  });
}
