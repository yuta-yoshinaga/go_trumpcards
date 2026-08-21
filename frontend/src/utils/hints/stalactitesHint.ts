import type { Card, StalactitesResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { StalactitesPhase } from '../../types/phases';

/** Card value for Ace. */
const ACE_VALUE = 1;

/** Number of suits. */
const SUIT_COUNT = 4;

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
  // When all 4 cells are full, this also subsumes Priority 4 (usedCells < SUIT_COUNT is false).
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
  if (usedCells < SUIT_COUNT) {
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
      if (top && canMoveToFoundation(top, state.foundation)) return true;
    }
  }

  // Free cell cards
  for (const cell of state.cells) {
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
