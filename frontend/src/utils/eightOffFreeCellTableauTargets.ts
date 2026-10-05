import type { Card } from '../types/card';

/** Returns legal tableau columns for moving a single Eight Off free-cell card. */
export function eightOffFreeCellTableauTargets(tableau: (Card | null)[][], card: Card): number[] {
  return tableau.flatMap((column, index) => {
    if (column.length === 0) return card.value === 13 ? [index] : [];
    const top = column[column.length - 1];
    return top && top.design === card.design && top.value === card.value + 1 ? [index] : [];
  });
}
