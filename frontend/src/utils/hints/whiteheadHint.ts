import type { Card, WhiteheadResponse, WhiteheadTableauCard } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { WhiteheadPhase } from '../../types/phases';

/** Card value for Ace. */
const ACE_VALUE = 1;

/** Number of suits. */
const SUIT_COUNT = 4;

/** Returns a frontend HintResult for Whitehead, or null if no suggestion. */
export function getWhiteheadHint(state: WhiteheadResponse): HintResult | null {
  if (state.phase !== WhiteheadPhase.PLAYING) return null;

  // Priority 1: Move to foundation
  if (hasFoundationMove(state)) {
    return { targetAction: 'move', reason: 'frontendHint.moveToFoundation', confidence: 'strong' };
  }

  // Klondike's "reveal a face-down card" priority is gone: Whitehead deals every
  // card face up, so that branch could never fire.

  // Priority 2: fill an empty column. **Any card may go there** -- Klondike
  // restricted this to a King, so keeping its check would have stayed silent
  // whenever the only available card was not a King.
  if (hasCardForEmptyColumn(state.tableau)) {
    return { targetAction: 'move', reason: 'frontendHint.moveToEmptyColumn', confidence: 'moderate' };
  }

  // Priority 4: Draw from stock
  if (state.stockCount > 0 || state.waste.length > 0) {
    return { targetAction: 'draw', reason: 'frontendHint.drawFromStock', confidence: 'moderate' };
  }

  return null;
}

/** Check if any face-up tableau top card or waste top can go to a foundation. */
function hasFoundationMove(state: WhiteheadResponse): boolean {
  // Tableau top cards
  for (const col of state.tableau) {
    if (col.length > 0) {
      const top = col[col.length - 1];
      if (top.card && top.faceUp && canMoveToFoundation(top.card, state.foundation)) return true;
    }
  }

  // Waste top
  const wasteTop = state.waste.length > 0 ? state.waste[state.waste.length - 1] : null;
  if (wasteTop && canMoveToFoundation(wasteTop, state.foundation)) return true;

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

/**
 * Whether an empty column exists and some card could move into it.
 *
 * Whitehead lets **any** card take an empty column; Klondike, which this was
 * cloned from, allowed only a King. A card already alone at the base of its own
 * column is excluded -- moving it to another empty column achieves nothing.
 */
function hasCardForEmptyColumn(tableau: WhiteheadTableauCard[][]): boolean {
  if (!tableau.some((col) => col.length === 0)) return false;
  return tableau.some((col) => col.length > 1);
}
