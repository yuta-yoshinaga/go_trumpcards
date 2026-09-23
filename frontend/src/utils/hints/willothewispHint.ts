import type { Card } from '../../types/card';
import type { WillOTheWispResponse, WillOTheWispTableauCard } from '../../types/games/willothewisp';
import { WillOTheWispPhase } from '../../types/games/willothewisp';
import type { HintResult } from '../../types/hint';

/** Threshold for a near-complete sequence (10 of 13 cards K→A in the same suit). */
const NEAR_COMPLETE_THRESHOLD = 10;

/**
 * Frontend strategic hint for WillOTheWisp.
 *
 * WillOTheWisp deals seven columns of three face-up cards. Cards build in
 * descending rank regardless of suit, while same-suit runs are the movable
 * units. Returns a {@link HintResult} or null when no suggestion is available.
 */
export function getWillOTheWispHint(state: WillOTheWispResponse | null | undefined): HintResult | null {
  if (!state || state.phase !== WillOTheWispPhase.PLAYING) return null;

  // Priority 1: Near-complete same-suit sequence
  if (hasNearCompleteSequence(state.tableau)) {
    return { targetAction: 'move', reason: 'frontendHint.completeSuit', confidence: 'strong' };
  }

  // Priority 2: Same-suit build move available
  if (hasSameSuitBuildMove(state.tableau)) {
    return { targetAction: 'move', reason: 'frontendHint.buildSameSuit', confidence: 'strong' };
  }

  // Priority 3: Empty column available
  if (state.tableau.some((col) => col.length === 0)) {
    return { targetAction: 'move', reason: 'frontendHint.useEmptyColumn', confidence: 'moderate' };
  }

  // Priority 4: Deal from stock
  if (state.stockCount > 0) {
    return { targetAction: 'deal', reason: 'frontendHint.dealFromStock', confidence: 'moderate' };
  }

  return null;
}

/** Check if any column has a same-suit descending sequence of NEAR_COMPLETE_THRESHOLD or more. */
function hasNearCompleteSequence(tableau: WillOTheWispTableauCard[][]): boolean {
  for (const col of tableau) {
    if (getSameSuitSequenceLength(col) >= NEAR_COMPLETE_THRESHOLD) return true;
  }
  return false;
}

/** Check if a face-up card from one column can be moved to another to extend a same-suit sequence. */
function hasSameSuitBuildMove(tableau: WillOTheWispTableauCard[][]): boolean {
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

/** Get the length of the same-suit descending sequence from the bottom of face-up cards. */
function getSameSuitSequenceLength(col: WillOTheWispTableauCard[]): number {
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
function getBottomOfSameSuitRun(col: WillOTheWispTableauCard[]): { design: Card['design']; value: number } | null {
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
