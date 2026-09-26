import type { Card } from '../types/card';

/** Returns legal tableau destinations for a selected Eight Off stack. */
export function eightOffTableauTargets(
  tableau: (Card | null)[][],
  freeCells: (Card | null)[],
  fromCol: number,
  cardIndex: number,
): number[] {
  const source = tableau[fromCol];
  if (!source || cardIndex < 0 || cardIndex >= source.length) return [];
  const moving = source.slice(cardIndex);
  if (
    moving.some(
      (card, index) =>
        !card ||
        (index > 0 &&
          (card.design !== moving[index - 1]?.design || card.value !== (moving[index - 1]?.value ?? 0) - 1)),
    )
  )
    return [];
  const bottom = moving[0];
  if (!bottom) return [];
  const free = freeCells.filter((card) => card === null).length;
  const emptyCols = tableau.filter((col) => col.length === 0).length;
  return tableau.flatMap((col, toCol) => {
    if (toCol === fromCol) return [];
    const emptyDestination = col.length === 0;
    const capacity = (1 + free) * 2 ** (emptyCols - (emptyDestination ? 1 : 0));
    if (moving.length > capacity) return [];
    const top = col[col.length - 1];
    if (
      emptyDestination ? bottom.value === 13 : !!top && top.design === bottom.design && top.value === bottom.value + 1
    )
      return [toCol];
    return [];
  });
}
