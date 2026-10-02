import type { NapoleonsSquareMoveZone, NapoleonsSquareResponse } from '../types/card';

/** A selectable source; tableau sources always identify a column and run head. */
export type NapoleonsSquareMoveSource = { zone: 'waste' } | { zone: 'tableau'; col: number; cardIndex: number };

/**
 * Returns legal destinations for a selected Napoleon's Square source.
 * Rules match `internal/domain/NapoleonsSquare.go` functions `MoveWasteToTableau`,
 * `MoveWasteToFoundation`, `MoveTableauToTableau`, `MoveTableauToFoundation`,
 * `napoleonsSquareIsRun`, `canPlaceOnTableau`, and `findFoundation`.
 */
export function napoleonsSquareLegalTargets(
  state: NapoleonsSquareResponse,
  source: NapoleonsSquareMoveSource,
): NapoleonsSquareMoveZone[] {
  if (state.phase !== 0) return [];

  const card = source.zone === 'waste' ? state.waste.at(-1) : state.tableau[source.col][source.cardIndex]?.card;
  if (!card) return [];

  if (source.zone === 'tableau') {
    const run = state.tableau[source.col].slice(source.cardIndex);
    if (
      run.length === 0 ||
      run.some((entry, index) => {
        if (!entry.card) return true;
        if (index === 0) return false;
        const previous = run[index - 1]?.card;
        return !previous || entry.card.design !== previous.design || entry.card.value !== previous.value - 1;
      })
    )
      return [];
  }

  const targets: NapoleonsSquareMoveZone[] = [];
  state.tableau.forEach((column, col) => {
    if (source.zone === 'tableau' && col === source.col) return;
    const top = column.at(-1)?.card;
    if (!top || (card.design === top.design && card.value === top.value - 1)) {
      targets.push({ zone: 'tableau', col });
    }
  });

  const canMoveToFoundation = source.zone === 'waste' || source.cardIndex === state.tableau[source.col].length - 1;
  if (canMoveToFoundation) {
    const suitOrder = ['SPADE', 'CLOVER', 'HEART', 'DIAMOND', 'SPADE', 'CLOVER', 'HEART', 'DIAMOND'];
    const foundation = state.foundation.findIndex((pile, index) => {
      const top = pile.at(-1);
      return top
        ? card.design === top.design && card.value === top.value + 1
        : card.value === 1 && card.design === suitOrder[index];
    });
    if (foundation >= 0) targets.push({ zone: 'foundation', col: foundation });
  }

  return targets;
}
