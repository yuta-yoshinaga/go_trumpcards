import type { Card, FreeCellResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { FreeCellPhase } from '../../types/phases';

/** Card value for Ace. */
const ACE_VALUE = 1;

/** Number of suits. */
const SUIT_COUNT = 4;

/** Threshold for warning about full free cells. */
const FREE_CELL_WARNING_THRESHOLD = 3;

/** Returns a frontend HintResult for FreeCell, or null if no suggestion. */
export function getFreeCellHint(state: FreeCellResponse): HintResult | null {
  if (state.phase !== FreeCellPhase.PLAYING) return null;

  // Priority 1: Move to foundation
  if (hasFoundationMove(state)) {
    return { targetAction: 'move', reason: 'frontendHint.moveToFoundation', confidence: 'strong' };
  }

  // Priority 2: Free cells nearly full
  const usedFreeCells = state.freeCells.filter((c) => c !== null).length;
  if (usedFreeCells >= FREE_CELL_WARNING_THRESHOLD) {
    return { targetAction: 'move', reason: 'frontendHint.freeCellsFilling', confidence: 'strong' };
  }

  // Priority 3: Empty column available
  const hasEmptyColumn = state.tableau.some((col) => col.length === 0);
  if (hasEmptyColumn) {
    return { targetAction: 'move', reason: 'frontendHint.useEmptyColumn', confidence: 'moderate' };
  }

  // Priority 4: Use free cells
  if (usedFreeCells < SUIT_COUNT) {
    return { targetAction: 'move', reason: 'frontendHint.useFreeCells', confidence: 'moderate' };
  }

  return null;
}

/** Check if any tableau top card or free cell card can go to a foundation. */
function hasFoundationMove(state: FreeCellResponse): boolean {
  const candidates: Card[] = [];

  // Tableau top cards
  for (const col of state.tableau) {
    if (col.length === 0) continue;
    const top = col[col.length - 1];
    if (top) candidates.push(top);
  }

  // Free cell cards
  for (const cell of state.freeCells) {
    if (cell) candidates.push(cell);
  }

  return candidates.some((card) => canMoveToFoundation(card, state.foundation));
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
