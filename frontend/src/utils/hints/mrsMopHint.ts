import type { Card, MrsMopResponse, MrsMopTableauCard } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { MrsMopPhase } from '../../types/phases';

/** Threshold for a near-complete sequence (10 of 13 cards K→A in the same suit). */
const NEAR_COMPLETE_THRESHOLD = 10;

/** Returns a frontend HintResult for MrsMop, or null if no suggestion. */
export function getMrsMopHint(state: MrsMopResponse): HintResult | null {
  if (state.phase !== MrsMopPhase.PLAYING) return null;

  // Priority 1: Near-complete same-suit sequence
  if (hasNearCompleteSequence(state.tableau)) {
    return { targetAction: 'move', reason: 'frontendHint.completeSuit', confidence: 'strong' };
  }

  // Priority 2: Same-suit build move available
  if (hasSameSuitBuildMove(state.tableau)) {
    return { targetAction: 'move', reason: 'frontendHint.buildSameSuit', confidence: 'strong' };
  }

  // Priority 3: Face-down cards to reveal
  if (hasFaceDownToReveal(state.tableau)) {
    return { targetAction: 'move', reason: 'frontendHint.revealFaceDown', confidence: 'moderate' };
  }

  // Priority 4: Empty column available
  if (state.tableau.some((col) => col.length === 0)) {
    return { targetAction: 'move', reason: 'frontendHint.useEmptyColumn', confidence: 'moderate' };
  }

  // Priority 5: Deal from stock
  if (state.stockCount > 0) {
    return { targetAction: 'deal', reason: 'frontendHint.dealFromStock', confidence: 'moderate' };
  }

  return null;
}

/** Check if any column has a same-suit descending sequence of NEAR_COMPLETE_THRESHOLD or more. */
function hasNearCompleteSequence(tableau: MrsMopTableauCard[][]): boolean {
  for (const col of tableau) {
    const len = getSameSuitSequenceLength(col);
    if (len >= NEAR_COMPLETE_THRESHOLD) return true;
  }
  return false;
}

/** Check if a face-up card from one column can be moved to another to extend a same-suit sequence. */
function hasSameSuitBuildMove(tableau: MrsMopTableauCard[][]): boolean {
  for (const col of tableau) {
    if (col.length === 0) continue;
    const bottom = getBottomOfSameSuitRun(col);
    if (!bottom) continue;

    for (const target of tableau) {
      if (target === col || target.length === 0) continue;
      const targetTop = target[target.length - 1];
      if (!targetTop.card || !targetTop.faceUp) continue;
      if (targetTop.card.design === bottom.design && targetTop.card.value === bottom.value + 1) {
        return true;
      }
    }
  }
  return false;
}

/** Check if any column has face-down cards that could be revealed by moving face-up cards away. */
function hasFaceDownToReveal(tableau: MrsMopTableauCard[][]): boolean {
  for (const col of tableau) {
    if (col.length < 2) continue;
    const hasFaceDown = col.some((c) => !c.faceUp && c.card);
    const topIsFaceUp = col[col.length - 1].faceUp;
    if (hasFaceDown && topIsFaceUp) return true;
  }
  return false;
}

/** Get the length of the same-suit descending sequence from the bottom of face-up cards. */
function getSameSuitSequenceLength(col: MrsMopTableauCard[]): number {
  if (col.length === 0 || !col[col.length - 1].faceUp) return 0;
  let count = 1;
  for (let i = col.length - 1; i > 0; i--) {
    const current = col[i];
    const above = col[i - 1];
    if (
      !current.faceUp ||
      !above.faceUp ||
      !current.card ||
      !above.card ||
      current.card.design !== above.card.design ||
      above.card.value !== current.card.value + 1
    ) {
      break;
    }
    count++;
  }
  return count;
}

/** Get the card at the bottom of the same-suit run from the top of a column. */
function getBottomOfSameSuitRun(col: MrsMopTableauCard[]): { design: Card['design']; value: number } | null {
  if (col.length === 0) return null;
  const top = col[col.length - 1];
  if (!top.faceUp || !top.card) return null;

  let bottomIdx = col.length - 1;
  for (let i = col.length - 1; i > 0; i--) {
    const current = col[i];
    const above = col[i - 1];
    if (
      !above.faceUp ||
      !above.card ||
      !current.card ||
      current.card.design !== above.card.design ||
      above.card.value !== current.card.value + 1
    ) {
      break;
    }
    bottomIdx = i - 1;
  }
  const bottomCard = col[bottomIdx].card;
  return bottomCard ? { design: bottomCard.design, value: bottomCard.value } : null;
}
