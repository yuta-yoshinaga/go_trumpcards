import type { Card, SeahavenTowersResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { SeahavenTowersPhase } from '../../types/phases';

/** Card value for Ace. */
const ACE_VALUE = 1;

/** Number of suits. */
const SUIT_COUNT = 4;

/** Reserved-cell capacity in Seahaven Towers (2). */
const RESERVED_CELL_CAPACITY = 2;

/** Returns a frontend HintResult for Seahaven Towers, or null if no suggestion. */
export function getSeahavenTowersHint(state: SeahavenTowersResponse): HintResult | null {
  if (state.phase !== SeahavenTowersPhase.PLAYING) return null;

  // Priority 1: Move to foundation.
  if (hasFoundationMove(state)) {
    return { targetAction: 'move', reason: 'frontendHint.moveToFoundation', confidence: 'strong' };
  }

  // Priority 2: Reserved cells full. Seahaven only has 2, so "full" is the threshold.
  const usedReserved = state.reservedCells.filter((c) => c !== null).length;
  if (usedReserved >= RESERVED_CELL_CAPACITY) {
    return { targetAction: 'move', reason: 'frontendHint.reservedFilling', confidence: 'strong' };
  }

  // Priority 3: Empty tableau column. Only Kings can land there, so the value is more constrained
  // than FreeCell, but the suggestion is still useful when an empty column exists.
  const hasEmptyColumn = state.tableau.some((col) => col.length === 0);
  if (hasEmptyColumn) {
    return { targetAction: 'move', reason: 'frontendHint.useEmptyColumn', confidence: 'moderate' };
  }

  // Priority 4: Use a reserved cell.
  if (usedReserved < RESERVED_CELL_CAPACITY) {
    return { targetAction: 'move', reason: 'frontendHint.useReserved', confidence: 'moderate' };
  }

  return null;
}

/** Check if any tableau top card or reserved-cell card can go to a foundation. */
function hasFoundationMove(state: SeahavenTowersResponse): boolean {
  for (const col of state.tableau) {
    if (col.length > 0) {
      const top = col[col.length - 1];
      if (top && canMoveToFoundation(top, state.foundation)) return true;
    }
  }
  for (const cell of state.reservedCells) {
    if (cell && canMoveToFoundation(cell, state.foundation)) return true;
  }
  return false;
}

/** Check if a card can be placed on any foundation pile. */
function canMoveToFoundation(card: Card, foundation: Card[][]): boolean {
  for (let i = 0; i < SUIT_COUNT; i++) {
    const pile = foundation[i];
    if (pile.length === 0) {
      if (card.value === ACE_VALUE) return true;
    } else {
      const top = pile[pile.length - 1];
      if (top.design === card.design && card.value === top.value + 1) return true;
    }
  }
  return false;
}
