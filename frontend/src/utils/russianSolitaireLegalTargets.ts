import type { Card, RussianSolitaireResponse } from '../types/card';

/** Legal destinations for a selected Russian Solitaire card or block. */
export interface RussianSolitaireLegalTargets {
  tableau: Set<number>;
  foundation: Set<number>;
}

/**
 * Derives legal destinations from the current state.
 * Sync: RussianSolitaire.canPlaceOnTableau / canPlaceOnFoundation in
 * `internal/domain/RussianSolitaire.go` (lines 324-337), and the source
 * restrictions in MoveTableauToTableau / MoveTableauToFoundation (lines 113-177).
 */
export function russianSolitaireLegalTargets(
  state: RussianSolitaireResponse,
  fromCol: number,
  cardIndex: number,
): RussianSolitaireLegalTargets {
  const result: RussianSolitaireLegalTargets = { tableau: new Set(), foundation: new Set() };
  const moving = state.tableau[fromCol]?.[cardIndex];
  const card = moving?.card;
  if (!card || !moving.faceUp) return result;

  state.tableau.forEach((column, index) => {
    if (index === fromCol) return;
    if (column.length === 0) {
      if (card.value === 13) result.tableau.add(index);
      return;
    }
    const top = column[column.length - 1]?.card;
    if (top && card.design === top.design && card.value === top.value - 1) result.tableau.add(index);
  });

  // Foundation moves only take the top card of a tableau column.
  if (cardIndex === state.tableau[fromCol].length - 1) {
    const foundationIndex = ['SPADE', 'CLOVER', 'HEART', 'DIAMOND'].indexOf(card.design);
    const pile = state.foundation[foundationIndex];
    if (pile && canPlaceOnFoundation(pile, card)) result.foundation.add(foundationIndex);
  }
  return result;
}

function canPlaceOnFoundation(pile: Card[], card: Card): boolean {
  const top = pile[pile.length - 1];
  return top ? card.design === top.design && card.value === top.value + 1 : card.value === 1;
}
