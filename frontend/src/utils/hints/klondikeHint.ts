import type { Card, KlondikeResponse, KlondikeTableauCard } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { KlondikePhase } from '../../types/phases';

/** Card value for King. */
const KING_VALUE = 13;

/** Card value for Ace. */
const ACE_VALUE = 1;

/** Number of suits. */
const SUIT_COUNT = 4;

/** Returns a frontend HintResult for Klondike, or null if no suggestion. */
export function getKlondikeHint(state: KlondikeResponse): HintResult | null {
  if (state.phase !== KlondikePhase.PLAYING) return null;

  // Priority 1: Move to foundation
  if (hasFoundationMove(state)) {
    return { targetAction: 'move', reason: 'frontendHint.moveToFoundation', confidence: 'strong' };
  }

  // Priority 2: Reveal face-down card
  if (hasRevealableColumn(state.tableau)) {
    return { targetAction: 'move', reason: 'frontendHint.revealFaceDown', confidence: 'strong' };
  }

  // Priority 3: Move King to empty column
  if (hasKingForEmptyColumn(state.tableau)) {
    return { targetAction: 'move', reason: 'frontendHint.moveKingToEmpty', confidence: 'moderate' };
  }

  // Priority 4: Draw from stock
  if (state.stockCount > 0 || state.waste.length > 0) {
    return { targetAction: 'draw', reason: 'frontendHint.drawFromStock', confidence: 'moderate' };
  }

  return null;
}

/** Check if any face-up tableau top card or waste top can go to a foundation. */
function hasFoundationMove(state: KlondikeResponse): boolean {
  const topCards = getTableauTopCards(state.tableau);
  const wasteTop = state.waste.length > 0 ? state.waste[state.waste.length - 1] : null;
  const candidates = wasteTop ? [...topCards, wasteTop] : topCards;

  for (const card of candidates) {
    if (canMoveToFoundation(card, state.foundation)) return true;
  }
  return false;
}

/** Get top face-up card from each non-empty tableau column. */
function getTableauTopCards(tableau: KlondikeTableauCard[][]): Card[] {
  const cards: Card[] = [];
  for (const col of tableau) {
    if (col.length === 0) continue;
    const top = col[col.length - 1];
    if (top.card && top.faceUp) cards.push(top.card);
  }
  return cards;
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

/** Check if any column has face-down cards that could be revealed by moving face-up cards. */
function hasRevealableColumn(tableau: KlondikeTableauCard[][]): boolean {
  for (const col of tableau) {
    if (col.length < 2) continue;
    const hasFaceDown = col.some((c) => !c.faceUp && c.card);
    const topIsFaceUp = col[col.length - 1].faceUp;
    if (hasFaceDown && topIsFaceUp) return true;
  }
  return false;
}

/** Check if there is an empty column and a King available to move there. */
function hasKingForEmptyColumn(tableau: KlondikeTableauCard[][]): boolean {
  const hasEmpty = tableau.some((col) => col.length === 0);
  if (!hasEmpty) return false;

  for (const col of tableau) {
    if (col.length === 0) continue;
    for (let i = 0; i < col.length; i++) {
      const c = col[i];
      if (c.faceUp && c.card?.value === KING_VALUE && i > 0) {
        // King is face-up but not at the base — moving it to an empty column is useful
        return true;
      }
    }
  }
  return false;
}
