import type { Card, PenguinResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { PenguinPhase } from '../../types/phases';

/** Number of suits. */
const SUIT_COUNT = 4;

/** Threshold for warning about nearly-full free cells (5 of 7). */
const FREE_CELL_WARNING_THRESHOLD = 5;

/** Returns a frontend HintResult for Penguin, or null if no suggestion. */
export function getPenguinHint(state: PenguinResponse): HintResult | null {
  if (state.phase !== PenguinPhase.PLAYING) return null;

  // Priority 1: Move to foundation
  if (hasFoundationMove(state)) {
    return { targetAction: 'move', reason: 'frontendHint.moveToFoundation', confidence: 'strong' };
  }

  // Priority 2: Free cells nearly full (5+ of 7)
  const usedFreeCells = state.freeCells.filter((c) => c !== null).length;
  if (usedFreeCells >= FREE_CELL_WARNING_THRESHOLD) {
    return { targetAction: 'move', reason: 'frontendHint.freeCellsNearlyFull', confidence: 'strong' };
  }

  // Priority 3: Empty column available — Penguin only accepts prevRank(baseRank) on empty columns
  const hasEmptyColumn = state.tableau.some((col) => col.length === 0);
  if (hasEmptyColumn && hasPrevRankToMove(state)) {
    return { targetAction: 'move', reason: 'frontendHint.useEmptyColumn', confidence: 'moderate' };
  }

  // Priority 4: Use free cells while available
  if (usedFreeCells < PenguinPhase.PLAYING + SUIT_COUNT + 3) {
    return { targetAction: 'move', reason: 'frontendHint.useFreeCells', confidence: 'moderate' };
  }

  return null;
}

/** Check if any tableau top card or free cell card can go to a foundation. */
function hasFoundationMove(state: PenguinResponse): boolean {
  for (const col of state.tableau) {
    if (col.length > 0) {
      const top = col[col.length - 1];
      if (top && canMoveToFoundation(top, state.foundation, state.baseRank)) return true;
    }
  }
  for (const cell of state.freeCells) {
    if (cell && canMoveToFoundation(cell, state.foundation, state.baseRank)) return true;
  }
  return false;
}

/** Check if a card can be placed on any foundation pile (starts at baseRank, wraps). */
function canMoveToFoundation(card: Card, foundation: Card[][], baseRank: number): boolean {
  for (let i = 0; i < SUIT_COUNT; i++) {
    const pile = foundation[i];
    if (pile.length === 0) {
      if (card.value === baseRank) return true;
    } else {
      const top = pile[pile.length - 1];
      const nextValue = top.value === 13 ? 1 : top.value + 1;
      if (top.design === card.design && card.value === nextValue) return true;
    }
  }
  return false;
}

/** True if there is a prevRank(baseRank) card available to move into an empty column. */
function hasPrevRankToMove(state: PenguinResponse): boolean {
  const prevRank = state.baseRank === 1 ? 13 : state.baseRank - 1;
  for (const col of state.tableau) {
    if (col.length > 0) {
      const top = col[col.length - 1];
      if (top && top.value === prevRank && col.length > 1) return true;
    }
  }
  for (const cell of state.freeCells) {
    if (cell && cell.value === prevRank) return true;
  }
  return false;
}
