import type { Card, EightOffResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { EightOffPhase } from '../../types/phases';

/** Card value for Ace. */
const ACE_VALUE = 1;

/** Card value for King. */
const KING_VALUE = 13;

/** Number of suits. */
const SUIT_COUNT = 4;

/** Threshold for warning about nearly-full free cells (out of 8). */
const FREE_CELL_WARNING_THRESHOLD = 6;

/** Returns a frontend HintResult for Eight Off, or null if no suggestion. */
export function getEightOffHint(state: EightOffResponse): HintResult | null {
  if (state.phase !== EightOffPhase.PLAYING) return null;

  // Priority 1: Move to foundation
  if (hasFoundationMove(state)) {
    return { targetAction: 'move', reason: 'frontendHint.moveToFoundation', confidence: 'strong' };
  }

  // Priority 2: Free cells nearly full (out of 8)
  const usedFreeCells = state.freeCells.filter((c) => c !== null).length;
  if (usedFreeCells >= FREE_CELL_WARNING_THRESHOLD) {
    return { targetAction: 'move', reason: 'frontendHint.freeCellsFilling', confidence: 'strong' };
  }

  // Priority 3: Empty column available — but Eight Off only accepts Kings on empty columns
  const hasEmptyColumn = state.tableau.some((col) => col.length === 0);
  if (hasEmptyColumn && hasKingToMove(state)) {
    return { targetAction: 'move', reason: 'frontendHint.useEmptyColumnKing', confidence: 'moderate' };
  }

  // Priority 4: Use free cells while available
  if (usedFreeCells < SUIT_COUNT * 2) {
    return { targetAction: 'move', reason: 'frontendHint.useFreeCells', confidence: 'moderate' };
  }

  return null;
}

/** Check if any tableau top card or free cell card can go to a foundation. */
function hasFoundationMove(state: EightOffResponse): boolean {
  for (const col of state.tableau) {
    if (col.length > 0) {
      const top = col[col.length - 1];
      if (top && canMoveToFoundation(top, state.foundation)) return true;
    }
  }
  for (const cell of state.freeCells) {
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

/** True if there is a King available to move into an empty column. */
function hasKingToMove(state: EightOffResponse): boolean {
  for (const col of state.tableau) {
    if (col.length > 0) {
      const top = col[col.length - 1];
      if (top && top.value === KING_VALUE && col.length > 1) return true;
    }
  }
  for (const cell of state.freeCells) {
    if (cell && cell.value === KING_VALUE) return true;
  }
  return false;
}
