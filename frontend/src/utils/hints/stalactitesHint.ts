import type { Card, StalactitesResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { StalactitesPhase } from '../../types/phases';

/** Highest card rank; the foundation sequence wraps from here back to Ace. */
const MAX_RANK = 13;

/** Number of cells. Stalactites deals into all four, so none is free at the start. */
const CELL_COUNT = 4;

/** Threshold for warning about full free cells. */
const FREE_CELL_WARNING_THRESHOLD = 3;

/** Returns a frontend HintResult for Stalactites, or null if no suggestion. */
export function getStalactitesHint(state: StalactitesResponse): HintResult | null {
  if (state.phase !== StalactitesPhase.PLAYING) return null;

  // Priority 1: Move to foundation
  if (hasFoundationMove(state)) {
    return { targetAction: 'move', reason: 'frontendHint.moveToFoundation', confidence: 'strong' };
  }

  // Priority 2: Free cells nearly full
  // This fires before checking for tableau-to-tableau moves because the warning is about resource
  // scarcity — when cells are nearly full, freeing them is more urgent than individual moves.
  // When all 4 cells are full, this also subsumes Priority 4 (usedCells < CELL_COUNT is false).
  const usedCells = state.cells.filter((c) => c !== null).length;
  if (usedCells >= FREE_CELL_WARNING_THRESHOLD) {
    return { targetAction: 'move', reason: 'frontendHint.cellsFilling', confidence: 'strong' };
  }

  // Priority 3: Empty column available
  const hasEmptyColumn = state.tableau.some((col) => col.length === 0);
  if (hasEmptyColumn) {
    return { targetAction: 'move', reason: 'frontendHint.useEmptyColumn', confidence: 'moderate' };
  }

  // Priority 4: Use free cells
  if (usedCells < CELL_COUNT) {
    return { targetAction: 'move', reason: 'frontendHint.useCells', confidence: 'moderate' };
  }

  return null;
}

/** Check if any tableau top card or free cell card can go to a foundation. */
function hasFoundationMove(state: StalactitesResponse): boolean {
  // Tableau top cards
  for (const col of state.tableau) {
    if (col.length > 0) {
      const top = col[col.length - 1];
      if (top && canMoveToFoundation(top, state.foundation, state.baseRank)) return true;
    }
  }

  // Free cell cards
  for (const cell of state.cells) {
    if (cell && canMoveToFoundation(cell, state.foundation, state.baseRank)) return true;
  }

  return false;
}

/**
 * Whether the card can go on any foundation pile.
 *
 * **This duplicates the server rule** (`Stalactites.canPlaceOnFoundation`) --
 * change it whenever the domain changes. The FreeCell version this was cloned
 * from required a matching suit and an Ace on an empty pile, which is wrong for
 * Stalactites in three ways: suit is ignored, an empty pile takes the deal's
 * `baseRank`, and the sequence wraps King -> Ace. Getting it wrong here
 * under-reports legal foundation moves and the hint goes quiet when a move
 * exists.
 */
function canMoveToFoundation(card: Card, foundation: Card[][], baseRank: number): boolean {
  const nextRank = (v: number): number => (v >= MAX_RANK ? 1 : v + 1);
  return foundation.some((pile) =>
    pile.length === 0 ? card.value === baseRank : card.value === nextRank(pile[pile.length - 1].value),
  );
}
